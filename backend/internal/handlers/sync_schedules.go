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
	db          *database.GormDB
	logger      *logrus.Logger
	events      *services.OperationEventService
	syncService *services.SyncJobService
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
	// Create sync service if not injected (for backward compatibility)
	syncService := services.NewSyncJobService(db, logger, events)
	
	return &SyncSchedulesHandler{
		db:          db,
		logger:      logger,
		events:      events,
		syncService: syncService,
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
		
		// Get user sync config for page size (if needed)
		var userPageSize *int
		if h.syncService != nil {
			if config, err := h.syncService.GetSyncConfig(instance.ID, "users"); err == nil {
				userPageSize = &config.PageSize
			}
		}
		
		response = append(response, SyncScheduleResponse{
			InstanceID:       instance.ID,
			InstanceName:     instance.Name,
			Enabled:          true, // Always true since we control per sync type now
			Schedules:        schedules,
			UserSyncPageSize: userPageSize,
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

	// Update sync settings using sync service
	if req.Enabled != nil && h.syncService != nil {
		// Update all sync types
		for _, syncType := range []string{"users", "safes", "groups"} {
			h.syncService.UpdateSyncConfig(instanceID, syncType, map[string]interface{}{
				"enabled": *req.Enabled,
			})
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
	validTypes := []string{"users", "groups", "safes"}
	valid := false
	for _, vt := range validTypes {
		if entityType == vt {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
		return
	}

	// Update the sync config
	if h.syncService != nil {
		updates := make(map[string]interface{})
		if req.Interval != nil {
			updates["interval_minutes"] = *req.Interval
		}
		if req.Enabled != nil {
			updates["enabled"] = *req.Enabled
		}
		if len(updates) > 0 {
			if err := h.syncService.UpdateSyncConfig(instanceID, entityType, updates); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entity schedule"})
				return
			}
		}
	}
	
	// Update page size for user sync
	if entityType == "users" && req.PageSize != nil && h.syncService != nil {
		h.syncService.UpdateSyncConfig(instanceID, "users", map[string]interface{}{
			"page_size": *req.PageSize,
		})
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
	
	// Pause all sync types for this instance
	if h.syncService != nil {
		for _, syncType := range []string{"users", "safes", "groups"} {
			if err := h.syncService.UpdateSyncConfig(instanceID, syncType, map[string]interface{}{
				"enabled": false,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause instance sync"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance sync paused"})
}

// ResumeInstance resumes all sync for an instance
func (h *SyncSchedulesHandler) ResumeInstance(c *gin.Context) {
	instanceID := c.Param("instanceId")
	
	// Resume all sync types for this instance
	if h.syncService != nil {
		for _, syncType := range []string{"users", "safes", "groups"} {
			if err := h.syncService.UpdateSyncConfig(instanceID, syncType, map[string]interface{}{
				"enabled": true,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume instance sync"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance sync resumed"})
}

// PauseAll pauses all instance syncs
func (h *SyncSchedulesHandler) PauseAll(c *gin.Context) {
	// Get all instances
	var instances []gormmodels.CyberArkInstance
	if err := h.db.Find(&instances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get instances"})
		return
	}
	
	// Pause all sync types for all instances
	if h.syncService != nil {
		for _, instance := range instances {
			for _, syncType := range []string{"users", "safes", "groups"} {
				h.syncService.UpdateSyncConfig(instance.ID, syncType, map[string]interface{}{
					"enabled": false,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "All syncs paused"})
}

// ResumeAll resumes all instance syncs
func (h *SyncSchedulesHandler) ResumeAll(c *gin.Context) {
	// Get all instances
	var instances []gormmodels.CyberArkInstance
	if err := h.db.Find(&instances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get instances"})
		return
	}
	
	// Resume all sync types for all instances
	if h.syncService != nil {
		for _, instance := range instances {
			for _, syncType := range []string{"users", "safes", "groups"} {
				h.syncService.UpdateSyncConfig(instance.ID, syncType, map[string]interface{}{
					"enabled": true,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "All syncs resumed"})
}

// buildSchedulesForInstance builds entity schedules for an instance
func (h *SyncSchedulesHandler) buildSchedulesForInstance(instance *gormmodels.CyberArkInstance) []EntitySchedule {
	schedules := []EntitySchedule{}
	
	// Users schedule
	userSchedule := h.buildEntitySchedule(instance, "users", nil)
	schedules = append(schedules, userSchedule)
	
	// Groups schedule
	groupSchedule := h.buildEntitySchedule(instance, "groups", nil)
	schedules = append(schedules, groupSchedule)
	
	// Safes schedule
	safeSchedule := h.buildEntitySchedule(instance, "safes", nil)
	schedules = append(schedules, safeSchedule)
	
	return schedules
}

// buildEntitySchedule builds a schedule for a specific entity type
func (h *SyncSchedulesHandler) buildEntitySchedule(instance *gormmodels.CyberArkInstance, entityType string, interval *int) EntitySchedule {
	schedule := EntitySchedule{
		EntityType: entityType,
		Enabled:    false,
		Interval:   60, // Default
	}
	
	// Get config from sync service if available
	if h.syncService != nil {
		if config, err := h.syncService.GetSyncConfig(instance.ID, entityType); err == nil {
			schedule.Enabled = config.Enabled
			schedule.Interval = config.IntervalMinutes
			if entityType == "users" {
				schedule.PageSize = &config.PageSize
			}
		}
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

// UpdateInstanceSchedule updates sync schedule for an instance using instance_id from path
func (h *SyncSchedulesHandler) UpdateInstanceSchedule(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "instanceId", Value: c.Param("instance_id")})
	h.UpdateSchedule(c)
}

// UpdateInstanceEntitySchedule updates entity-specific schedule using instance_id from path
func (h *SyncSchedulesHandler) UpdateInstanceEntitySchedule(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "instanceId", Value: c.Param("instance_id")})
	c.Params = append(c.Params, gin.Param{Key: "entityType", Value: c.Param("entity_type")})
	h.UpdateEntitySchedule(c)
}

// TriggerInstanceSync triggers sync for an entity using instance_id from path
func (h *SyncSchedulesHandler) TriggerInstanceSync(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "instanceId", Value: c.Param("instance_id")})
	c.Params = append(c.Params, gin.Param{Key: "entityType", Value: c.Param("entity_type")})
	h.TriggerSync(c)
}