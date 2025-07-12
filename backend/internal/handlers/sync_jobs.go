package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/middleware"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/services"
	"github.com/orca-ng/orca/pkg/ulid"
)

// SyncJobsHandler handles sync job operations
type SyncJobsHandler struct {
	db          *database.GormDB
	logger      *logrus.Logger
	syncService *services.SyncJobService
	events      *services.OperationEventService
}

// NewSyncJobsHandler creates a new sync jobs handler
func NewSyncJobsHandler(db *database.GormDB, logger *logrus.Logger, syncService *services.SyncJobService, events *services.OperationEventService) *SyncJobsHandler {
	return &SyncJobsHandler{
		db:          db,
		logger:      logger,
		syncService: syncService,
		events:      events,
	}
}

// ListSyncJobs lists sync jobs with optional filtering
func (h *SyncJobsHandler) ListSyncJobs(c *gin.Context) {
	// Parse query parameters
	instanceID := c.Query("instance_id")
	syncType := c.Query("sync_type")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Build query
	query := h.db.Model(&gormmodels.SyncJob{}).
		Preload("CyberArkInstance").
		Preload("CreatedByUser").
		Order("created_at DESC")

	if instanceID != "" {
		query = query.Where("cyberark_instance_id = ?", instanceID)
	}
	if syncType != "" {
		query = query.Where("sync_type = ?", syncType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get results
	var jobs []gormmodels.SyncJob
	if err := query.Limit(limit).Offset(offset).Find(&jobs).Error; err != nil {
		h.logger.WithError(err).Error("Failed to list sync jobs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sync jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sync_jobs": jobs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetSyncJob gets a specific sync job
func (h *SyncJobsHandler) GetSyncJob(c *gin.Context) {
	id := c.Param("id")

	var job gormmodels.SyncJob
	err := h.db.Preload("CyberArkInstance").
		Preload("CreatedByUser").
		First(&job, "id = ?", id).Error

	if err != nil {
		h.logger.WithError(err).Error("Failed to get sync job")
		c.JSON(http.StatusNotFound, gin.H{"error": "Sync job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// TriggerSyncRequest represents a request to trigger a sync
type TriggerSyncRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	SyncType   string `json:"sync_type" binding:"required,oneof=users safes groups"`
}

// TriggerSync manually triggers a sync job
func (h *SyncJobsHandler) TriggerSync(c *gin.Context) {
	var req TriggerSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Check if instance exists and is active
	var instance gormmodels.CyberArkInstance
	if err := h.db.First(&instance, "id = ? AND is_active = ?", req.InstanceID, true).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Active instance not found"})
		return
	}

	// Create sync job
	job, err := h.syncService.CreateSyncJob(
		req.InstanceID,
		req.SyncType,
		gormmodels.TriggeredByManual,
		&user.ID,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create sync job")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync job"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sync job created",
		"job_id":  job.ID,
		"job":     job,
	})
}

// GetSyncConfigs gets sync configurations for an instance
func (h *SyncJobsHandler) GetSyncConfigs(c *gin.Context) {
	instanceID := c.Param("instance_id")

	// Check if instance exists
	var instance gormmodels.CyberArkInstance
	if err := h.db.First(&instance, "id = ?", instanceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	// Get or create configs for all sync types
	syncTypes := []string{
		gormmodels.SyncTypeUsers,
		gormmodels.SyncTypeSafes,
		gormmodels.SyncTypeGroups,
	}

	configs := make(map[string]*gormmodels.InstanceSyncConfig)
	for _, syncType := range syncTypes {
		config, err := h.syncService.GetSyncConfig(instanceID, syncType)
		if err != nil {
			h.logger.WithError(err).WithFields(logrus.Fields{
				"instance_id": instanceID,
				"sync_type":   syncType,
			}).Error("Failed to get sync config")
			continue
		}
		configs[syncType] = config
	}

	c.JSON(http.StatusOK, gin.H{
		"instance_id": instanceID,
		"configs":     configs,
	})
}

// UpdateSyncConfigRequest represents a request to update sync configuration
type UpdateSyncConfigRequest struct {
	Enabled         *bool `json:"enabled"`
	IntervalMinutes *int  `json:"interval_minutes" binding:"omitempty,min=5"`
	PageSize        *int  `json:"page_size" binding:"omitempty,min=1,max=1000"`
	RetryAttempts   *int  `json:"retry_attempts" binding:"omitempty,min=0,max=10"`
	TimeoutMinutes  *int  `json:"timeout_minutes" binding:"omitempty,min=1,max=120"`
}

// UpdateSyncConfig updates sync configuration for a specific sync type
func (h *SyncJobsHandler) UpdateSyncConfig(c *gin.Context) {
	instanceID := c.Param("instance_id")
	syncType := c.Param("sync_type")

	// Validate sync type
	if syncType != gormmodels.SyncTypeUsers && 
	   syncType != gormmodels.SyncTypeSafes && 
	   syncType != gormmodels.SyncTypeGroups {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sync type"})
		return
	}

	var req UpdateSyncConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.IntervalMinutes != nil {
		updates["interval_minutes"] = *req.IntervalMinutes
	}
	if req.PageSize != nil {
		updates["page_size"] = *req.PageSize
	}
	if req.RetryAttempts != nil {
		updates["retry_attempts"] = *req.RetryAttempts
	}
	if req.TimeoutMinutes != nil {
		updates["timeout_minutes"] = *req.TimeoutMinutes
	}

	// Update config
	if err := h.syncService.UpdateSyncConfig(instanceID, syncType, updates); err != nil {
		h.logger.WithError(err).Error("Failed to update sync config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sync config"})
		return
	}

	// Get updated config
	config, err := h.syncService.GetSyncConfig(instanceID, syncType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get updated sync config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// StreamSyncJobs streams sync job updates via SSE
func (h *SyncJobsHandler) StreamSyncJobs(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create client ID
	clientID := ulid.New("sseconn")
	
	// Subscribe to events
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	
	eventChan := h.events.Subscribe(ctx, clientID)
	
	// Create ticker for heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Send initial connected event
	c.SSEvent("connected", gin.H{"client_id": clientID})
	c.Writer.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			h.logger.Debug("SSE client disconnected")
			return
			
		case event := <-eventChan:
			// Only send sync job events
			if event.SyncJob != nil {
				eventData, err := services.MarshalEventToJSON(event)
				if err != nil {
					h.logger.WithError(err).Error("Failed to marshal sync job event")
					continue
				}
				
				c.SSEvent(event.Type, eventData)
				c.Writer.Flush()
			}
			
		case <-ticker.C:
			// Send heartbeat
			c.SSEvent("heartbeat", gin.H{"timestamp": time.Now().Unix()})
			c.Writer.Flush()
		}
	}
}