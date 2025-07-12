package services

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/orca-ng/orca/internal/database"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/pkg/ulid"
)

// SyncJobService handles sync job operations
type SyncJobService struct {
	db     *database.GormDB
	logger *logrus.Logger
	events *OperationEventService
}

// NewSyncJobService creates a new sync job service
func NewSyncJobService(db *database.GormDB, logger *logrus.Logger, events *OperationEventService) *SyncJobService {
	return &SyncJobService{
		db:     db,
		logger: logger,
		events: events,
	}
}

// CreateSyncJob creates a new sync job
func (s *SyncJobService) CreateSyncJob(instanceID, syncType, triggeredBy string, userID *string) (*gormmodels.SyncJob, error) {
	job := &gormmodels.SyncJob{
		ID:                 ulid.New(ulid.SyncJobPrefix),
		CyberArkInstanceID: instanceID,
		SyncType:           syncType,
		Status:             gormmodels.SyncJobStatusPending,
		TriggeredBy:        triggeredBy,
		CreatedBy:          userID,
	}

	if err := s.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("create sync job: %w", err)
	}

	// Load relations for event
	s.db.Preload("CyberArkInstance").Preload("CreatedByUser").First(job, "id = ?", job.ID)

	// Publish creation event
	if s.events != nil {
		s.events.PublishSyncJobCreated(job)
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":      job.ID,
		"instance_id": instanceID,
		"sync_type":   syncType,
		"triggered_by": triggeredBy,
	}).Info("Sync job created")

	return job, nil
}

// StartSyncJob marks a sync job as running
func (s *SyncJobService) StartSyncJob(jobID string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     gormmodels.SyncJobStatusRunning,
		"started_at": now,
	}

	if err := s.db.Model(&gormmodels.SyncJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return fmt.Errorf("update sync job: %w", err)
	}

	// Load and publish event
	var job gormmodels.SyncJob
	if err := s.db.Preload("CyberArkInstance").First(&job, "id = ?", jobID).Error; err == nil {
		if s.events != nil {
			s.events.PublishSyncJobUpdated(&job)
		}
	}

	return nil
}

// CompleteSyncJob marks a sync job as completed
func (s *SyncJobService) CompleteSyncJob(jobID string, stats SyncStats) error {
	now := time.Now()
	
	// Get job to calculate duration
	var job gormmodels.SyncJob
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return fmt.Errorf("get sync job: %w", err)
	}

	var duration float64
	if job.StartedAt != nil {
		duration = now.Sub(*job.StartedAt).Seconds()
	}

	updates := map[string]interface{}{
		"status":           gormmodels.SyncJobStatusCompleted,
		"completed_at":     now,
		"duration_seconds": duration,
		"records_synced":   stats.RecordsSynced,
		"records_created":  stats.RecordsCreated,
		"records_updated":  stats.RecordsUpdated,
		"records_deleted":  stats.RecordsDeleted,
		"records_failed":   stats.RecordsFailed,
	}

	if err := s.db.Model(&gormmodels.SyncJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return fmt.Errorf("update sync job: %w", err)
	}

	// Update sync config last run info
	s.updateSyncConfigLastRun(job.CyberArkInstanceID, job.SyncType, gormmodels.SyncJobStatusCompleted, "Sync completed successfully")

	// Load and publish event
	if err := s.db.Preload("CyberArkInstance").First(&job, "id = ?", jobID).Error; err == nil {
		if s.events != nil {
			s.events.PublishSyncJobUpdated(&job)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"job_id":   jobID,
		"duration": duration,
		"stats":    stats,
	}).Info("Sync job completed")

	return nil
}

// FailSyncJob marks a sync job as failed
func (s *SyncJobService) FailSyncJob(jobID string, err error) error {
	now := time.Now()
	errorMsg := err.Error()
	
	// Get job to calculate duration
	var job gormmodels.SyncJob
	if dbErr := s.db.First(&job, "id = ?", jobID).Error; dbErr != nil {
		return fmt.Errorf("get sync job: %w", dbErr)
	}

	var duration float64
	if job.StartedAt != nil {
		duration = now.Sub(*job.StartedAt).Seconds()
	}

	updates := map[string]interface{}{
		"status":           gormmodels.SyncJobStatusFailed,
		"completed_at":     now,
		"duration_seconds": duration,
		"error_message":    errorMsg,
	}

	if dbErr := s.db.Model(&gormmodels.SyncJob{}).Where("id = ?", jobID).Updates(updates).Error; dbErr != nil {
		return fmt.Errorf("update sync job: %w", dbErr)
	}

	// Update sync config last run info
	s.updateSyncConfigLastRun(job.CyberArkInstanceID, job.SyncType, gormmodels.SyncJobStatusFailed, errorMsg)

	// Load and publish event
	if dbErr := s.db.Preload("CyberArkInstance").First(&job, "id = ?", jobID).Error; dbErr == nil {
		if s.events != nil {
			s.events.PublishSyncJobUpdated(&job)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"job_id": jobID,
		"error":  errorMsg,
	}).Error("Sync job failed")

	return nil
}

// GetSyncConfig gets or creates sync configuration for an instance and sync type
func (s *SyncJobService) GetSyncConfig(instanceID, syncType string) (*gormmodels.InstanceSyncConfig, error) {
	var config gormmodels.InstanceSyncConfig
	
	err := s.db.Where("cyberark_instance_id = ? AND sync_type = ?", instanceID, syncType).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		// Create default config
		config = gormmodels.InstanceSyncConfig{
			ID:                 ulid.New(ulid.SyncConfigPrefix),
			CyberArkInstanceID: instanceID,
			SyncType:           syncType,
			Enabled:            true,
			IntervalMinutes:    60,
			PageSize:           100,
			RetryAttempts:      3,
			TimeoutMinutes:     30,
		}
		
		if err := s.db.Create(&config).Error; err != nil {
			return nil, fmt.Errorf("create sync config: %w", err)
		}
		
		s.logger.WithFields(logrus.Fields{
			"instance_id": instanceID,
			"sync_type":   syncType,
		}).Info("Created default sync config")
	} else if err != nil {
		return nil, fmt.Errorf("get sync config: %w", err)
	}
	
	return &config, nil
}

// UpdateSyncConfig updates sync configuration
func (s *SyncJobService) UpdateSyncConfig(instanceID, syncType string, updates map[string]interface{}) error {
	config, err := s.GetSyncConfig(instanceID, syncType)
	if err != nil {
		return err
	}

	if err := s.db.Model(config).Updates(updates).Error; err != nil {
		return fmt.Errorf("update sync config: %w", err)
	}

	// Calculate next run time if interval changed
	if _, ok := updates["interval_minutes"]; ok {
		nextRun := config.CalculateNextRunAt()
		s.db.Model(config).Update("next_run_at", nextRun)
	}

	return nil
}

// GetDueSyncJobs gets all sync jobs that are due to run
func (s *SyncJobService) GetDueSyncJobs() ([]*gormmodels.InstanceSyncConfig, error) {
	var configs []*gormmodels.InstanceSyncConfig
	
	err := s.db.Preload("CyberArkInstance").
		Where("enabled = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, time.Now()).
		Find(&configs).Error
		
	if err != nil {
		return nil, fmt.Errorf("get due sync jobs: %w", err)
	}
	
	return configs, nil
}

// updateSyncConfigLastRun updates the last run information for a sync config
func (s *SyncJobService) updateSyncConfigLastRun(instanceID, syncType, status, message string) {
	config, err := s.GetSyncConfig(instanceID, syncType)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get sync config for update")
		return
	}

	now := time.Now()
	nextRun := config.CalculateNextRunAt()
	
	updates := map[string]interface{}{
		"last_run_at":      now,
		"last_run_status":  status,
		"last_run_message": message,
		"next_run_at":      nextRun,
	}

	if err := s.db.Model(config).Updates(updates).Error; err != nil {
		s.logger.WithError(err).Error("Failed to update sync config last run")
	}
}

// SyncStats holds statistics for a sync job
type SyncStats struct {
	RecordsSynced  int `json:"records_synced"`
	RecordsCreated int `json:"records_created"`
	RecordsUpdated int `json:"records_updated"`
	RecordsDeleted int `json:"records_deleted"`
	RecordsFailed  int `json:"records_failed"`
}