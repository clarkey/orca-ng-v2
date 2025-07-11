package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/middleware"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/pipeline"
	"github.com/orca-ng/orca/internal/services"
	"github.com/orca-ng/orca/pkg/ulid"
)

type SyncSchedulesHandler struct {
	db     *database.GormDB
	logger *logrus.Logger
	events *services.OperationEventService
}

// EntitySchedule represents a sync schedule for a specific entity type
type EntitySchedule struct {
	EntityType   string     `json:"entityType"`
	Enabled      bool       `json:"enabled"`
	Interval     int        `json:"interval"` // minutes
	PageSize     *int       `json:"pageSize,omitempty"` // pagination size for users
	LastSyncAt   *time.Time `json:"lastSyncAt,omitempty"`
	LastStatus   *string    `json:"lastStatus,omitempty"`
	NextSyncAt   time.Time  `json:"nextSyncAt"`
	RecordCount  *int       `json:"recordCount,omitempty"`
	LastError    *string    `json:"lastError,omitempty"`
}

// SyncScheduleResponse represents the response for sync schedules
type SyncScheduleResponse struct {
	InstanceID       string           `json:"instanceId"`
	InstanceName     string           `json:"instanceName"`
	Enabled          bool             `json:"enabled"`
	Schedules        []EntitySchedule `json:"schedules"`
	UserSyncPageSize *int             `json:"userSyncPageSize,omitempty"`
}

// UpdateScheduleRequest represents a request to update sync schedules
type UpdateScheduleRequest struct {
	Enabled   *bool                     `json:"enabled,omitempty"`
	Schedules []map[string]interface{} `json:"schedules,omitempty"`
}

// UpdateEntityScheduleRequest represents a request to update a specific entity schedule
type UpdateEntityScheduleRequest struct {
	Enabled  *bool `json:"enabled,omitempty"`
	Interval *int  `json:"interval,omitempty"`
	PageSize *int  `json:"pageSize,omitempty"` // For user sync only
}

func NewSyncSchedulesHandler(db *database.GormDB, logger *logrus.Logger, events *services.OperationEventService) *SyncSchedulesHandler {
	return &SyncSchedulesHandler{
		db:     db,
		logger: logger,
		events: events,
	}
}

// GetSchedules returns all sync schedules
func (h *SyncSchedulesHandler) GetSchedules(c *gin.Context) {
	var instances []gormmodels.CyberArkInstance
	
	// Get all instances
	if err := h.db.Find(&instances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch instances"})
		return
	}

	// Build response
	response := make([]SyncScheduleResponse, 0)
	for _, instance := range instances {
		// Get last sync operations for this instance
		schedules := h.buildSchedulesForInstance(&instance)
		
		response = append(response, SyncScheduleResponse{
			InstanceID:       instance.ID,
			InstanceName:     instance.Name,
			Enabled:          instance.SyncEnabled,
			Schedules:        schedules,
			UserSyncPageSize: instance.UserSyncPageSize,
		})
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSchedule updates sync settings for an instance
func (h *SyncSchedulesHandler) UpdateSchedule(c *gin.Context) {
	instanceID := c.Param("instanceId")
	
	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update instance sync settings
	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["sync_enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := h.db.Model(&gormmodels.CyberArkInstance{}).
			Where("id = ?", instanceID).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule updated successfully"})
}

// UpdateEntitySchedule updates a specific entity sync schedule
func (h *SyncSchedulesHandler) UpdateEntitySchedule(c *gin.Context) {
	instanceID := c.Param("instanceId")
	entityType := c.Param("entityType")
	
	var req UpdateEntityScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate entity type
	validTypes := map[string]string{
		"users":  "user_sync_interval",
		"groups": "group_sync_interval",
		"safes":  "safe_sync_interval",
	}
	
	intervalField, valid := validTypes[entityType]
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
		return
	}

	// Update the specific interval
	updates := make(map[string]interface{})
	if req.Interval != nil {
		updates[intervalField] = *req.Interval
	}
	
	// Update page size for user sync
	if entityType == "users" && req.PageSize != nil {
		updates["user_sync_page_size"] = *req.PageSize
	}

	if len(updates) > 0 {
		if err := h.db.Model(&gormmodels.CyberArkInstance{}).
			Where("id = ?", instanceID).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entity schedule"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entity schedule updated successfully"})
}

// TriggerSync creates a new sync operation for immediate execution
func (h *SyncSchedulesHandler) TriggerSync(c *gin.Context) {
	instanceID := c.Param("instanceId")
	entityType := c.Param("entityType")
	
	// Map entity type to operation type
	opTypeMap := map[string]pipeline.OperationType{
		"users":  pipeline.OpTypeUserSync,
		"groups": pipeline.OpTypeGroupSync,
		"safes":  pipeline.OpTypeSafeSync,
	}
	
	opType, valid := opTypeMap[entityType]
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
		return
	}

	// Create sync operation
	payload := map[string]interface{}{
		"instance_id": instanceID,
		"sync_mode":   "manual",
	}
	
	payloadBytes, _ := json.Marshal(payload)
	
	operation := &gormmodels.Operation{
		ID:                 ulid.New(ulid.OperationPrefix),
		Type:               string(opType),
		Priority:           gormmodels.OpPriorityHigh, // Manual syncs get high priority
		Status:             gormmodels.OpStatusPending,
		Payload:            payloadBytes,
		ScheduledAt:        time.Now(),
		CyberArkInstanceID: &instanceID,
		RetryCount:         0,
		MaxRetries:         3,
	}
	
	// Get user from context
	if user := middleware.GetUser(c); user != nil {
		operation.CreatedBy = &user.ID
	}
	
	if err := h.db.Create(operation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync operation"})
		return
	}
	
	// Publish creation event
	if h.events != nil {
		// Load related data for the event
		h.db.Preload("Creator").Preload("CyberArkInstance").First(operation, "id = ?", operation.ID)
		h.events.PublishOperationCreated(operation)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Sync operation created",
		"operation_id": operation.ID,
	})
}

// PauseInstance pauses all sync for an instance
func (h *SyncSchedulesHandler) PauseInstance(c *gin.Context) {
	instanceID := c.Param("instanceId")
	
	if err := h.db.Model(&gormmodels.CyberArkInstance{}).
		Where("id = ?", instanceID).
		Update("sync_enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause instance sync"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance sync paused"})
}

// ResumeInstance resumes all sync for an instance
func (h *SyncSchedulesHandler) ResumeInstance(c *gin.Context) {
	instanceID := c.Param("instanceId")
	
	if err := h.db.Model(&gormmodels.CyberArkInstance{}).
		Where("id = ?", instanceID).
		Update("sync_enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume instance sync"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance sync resumed"})
}

// PauseAll pauses all instance syncs
func (h *SyncSchedulesHandler) PauseAll(c *gin.Context) {
	if err := h.db.Model(&gormmodels.CyberArkInstance{}).
		Where("sync_enabled = ?", true).
		Update("sync_enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause all syncs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All syncs paused"})
}

// ResumeAll resumes all instance syncs
func (h *SyncSchedulesHandler) ResumeAll(c *gin.Context) {
	if err := h.db.Model(&gormmodels.CyberArkInstance{}).
		Where("sync_enabled = ?", false).
		Update("sync_enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume all syncs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All syncs resumed"})
}

// buildSchedulesForInstance builds entity schedules for an instance
func (h *SyncSchedulesHandler) buildSchedulesForInstance(instance *gormmodels.CyberArkInstance) []EntitySchedule {
	schedules := []EntitySchedule{}
	
	// Users schedule (with page size)
	userSchedule := h.buildEntitySchedule(instance, "users", instance.UserSyncInterval)
	userSchedule.PageSize = instance.UserSyncPageSize
	schedules = append(schedules, userSchedule)
	
	// Groups schedule
	groupSchedule := h.buildEntitySchedule(instance, "groups", instance.GroupSyncInterval)
	schedules = append(schedules, groupSchedule)
	
	// Safes schedule
	safeSchedule := h.buildEntitySchedule(instance, "safes", instance.SafeSyncInterval)
	schedules = append(schedules, safeSchedule)
	
	return schedules
}

// buildEntitySchedule builds a schedule for a specific entity type
func (h *SyncSchedulesHandler) buildEntitySchedule(instance *gormmodels.CyberArkInstance, entityType string, interval *int) EntitySchedule {
	schedule := EntitySchedule{
		EntityType: entityType,
		Enabled:    interval != nil && *interval > 0,
		Interval:   30, // Default
	}
	
	if interval != nil {
		schedule.Interval = *interval
	}
	
	// Get last sync info from operations
	var lastOp gormmodels.Operation
	opType := h.getOperationType(entityType)
	
	err := h.db.Where("cyberark_instance_id = ? AND type = ?", instance.ID, opType).
		Order("created_at DESC").
		First(&lastOp).Error
		
	if err == nil {
		if lastOp.StartedAt != nil {
			schedule.LastSyncAt = lastOp.StartedAt
		}
		
		// Determine status
		status := string(lastOp.Status)
		if lastOp.Status == gormmodels.OpStatusCompleted {
			status = "success"
		} else if lastOp.Status == gormmodels.OpStatusFailed {
			status = "failed"
		} else if lastOp.Status == gormmodels.OpStatusProcessing {
			status = "running"
		}
		schedule.LastStatus = &status
		
		if lastOp.ErrorMessage != nil {
			schedule.LastError = lastOp.ErrorMessage
		}
		
		// Try to get record count from result
		if lastOp.Result != nil {
			var result map[string]interface{}
			if err := json.Unmarshal(*lastOp.Result, &result); err == nil {
				if count, ok := result["total_records"].(float64); ok {
					intCount := int(count)
					schedule.RecordCount = &intCount
				}
			}
		}
	}
	
	// Calculate next sync time
	if schedule.LastSyncAt != nil && schedule.Enabled {
		schedule.NextSyncAt = schedule.LastSyncAt.Add(time.Duration(schedule.Interval) * time.Minute)
	} else if schedule.Enabled {
		// If never synced, next sync is now
		schedule.NextSyncAt = time.Now()
	} else {
		// If disabled, set far future
		schedule.NextSyncAt = time.Now().Add(365 * 24 * time.Hour)
	}
	
	return schedule
}

// getOperationType maps entity type to pipeline operation type
func (h *SyncSchedulesHandler) getOperationType(entityType string) string {
	switch entityType {
	case "users":
		return string(pipeline.OpTypeUserSync)
	case "groups":
		return string(pipeline.OpTypeGroupSync)
	case "safes":
		return string(pipeline.OpTypeSafeSync)
	default:
		return ""
	}
}