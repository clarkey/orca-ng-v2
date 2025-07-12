package handlers

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/orca-ng/orca/internal/database"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/services"
	"github.com/orca-ng/orca/pkg/ulid"
)

// ActivityHandler handles unified activity view
type ActivityHandler struct {
	db     *database.GormDB
	logger *logrus.Logger
	events *services.OperationEventService
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(db *database.GormDB, logger *logrus.Logger, events *services.OperationEventService) *ActivityHandler {
	return &ActivityHandler{
		db:     db,
		logger: logger,
		events: events,
	}
}

// ActivityItem represents a unified activity item
type ActivityItem struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"` // "operation" or "sync"
	Status     string                    `json:"status"`
	Title      string                    `json:"title"`
	Subtitle   string                    `json:"subtitle"`
	Instance   *gormmodels.CyberArkInstance `json:"instance,omitempty"`
	CreatedBy  *gormmodels.User          `json:"created_by,omitempty"`
	CreatedAt  time.Time                 `json:"created_at"`
	StartedAt  *time.Time                `json:"started_at,omitempty"`
	CompletedAt *time.Time               `json:"completed_at,omitempty"`
	Duration   *float64                  `json:"duration_seconds,omitempty"`
	Error      *string                   `json:"error,omitempty"`
	Operation  *gormmodels.Operation     `json:"operation,omitempty"`
	SyncJob    *gormmodels.SyncJob       `json:"sync_job,omitempty"`
}

// ListActivity returns a unified view of operations and sync jobs
func (h *ActivityHandler) ListActivity(c *gin.Context) {
	// Parse query parameters
	instanceID := c.Query("instance_id")
	activityType := c.Query("type") // "operation", "sync", or empty for all
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	var items []ActivityItem
	
	// Get operations if not filtered to sync only
	if activityType != "sync" {
		query := h.db.Model(&gormmodels.Operation{}).
			Preload("CyberArkInstance").
			Preload("Creator")
			
		if instanceID != "" {
			query = query.Where("cyberark_instance_id = ?", instanceID)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		
		var operations []gormmodels.Operation
		if err := query.Find(&operations).Error; err != nil {
			h.logger.WithError(err).Error("Failed to get operations")
		} else {
			for _, op := range operations {
				item := ActivityItem{
					ID:          op.ID,
					Type:        "operation",
					Status:      op.Status,
					Title:       getOperationTitle(op.Type),
					Subtitle:    getOperationSubtitle(&op),
					Instance:    op.CyberArkInstance,
					CreatedBy:   op.Creator,
					CreatedAt:   op.CreatedAt,
					StartedAt:   op.StartedAt,
					CompletedAt: op.CompletedAt,
					Error:       op.ErrorMessage,
					Operation:   &op,
				}
				
				// Calculate duration
				if op.StartedAt != nil && op.CompletedAt != nil {
					duration := op.CompletedAt.Sub(*op.StartedAt).Seconds()
					item.Duration = &duration
				}
				
				items = append(items, item)
			}
		}
	}
	
	// Get sync jobs if not filtered to operations only
	if activityType != "operation" {
		query := h.db.Model(&gormmodels.SyncJob{}).
			Preload("CyberArkInstance").
			Preload("CreatedByUser")
			
		if instanceID != "" {
			query = query.Where("cyberark_instance_id = ?", instanceID)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		
		var syncJobs []gormmodels.SyncJob
		if err := query.Find(&syncJobs).Error; err != nil {
			h.logger.WithError(err).Error("Failed to get sync jobs")
		} else {
			for _, job := range syncJobs {
				item := ActivityItem{
					ID:          job.ID,
					Type:        "sync",
					Status:      job.Status,
					Title:       getSyncJobTitle(job.SyncType),
					Subtitle:    getSyncJobSubtitle(&job),
					Instance:    job.CyberArkInstance,
					CreatedBy:   job.CreatedByUser,
					CreatedAt:   job.CreatedAt,
					StartedAt:   job.StartedAt,
					CompletedAt: job.CompletedAt,
					Duration:    job.DurationSeconds,
					Error:       job.ErrorMessage,
					SyncJob:     &job,
				}
				
				items = append(items, item)
			}
		}
	}
	
	// Sort by created_at descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	
	// Apply pagination
	total := len(items)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	
	paginatedItems := items[start:end]
	
	c.JSON(http.StatusOK, gin.H{
		"activities": paginatedItems,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// StreamActivity streams unified activity updates via SSE
func (h *ActivityHandler) StreamActivity(c *gin.Context) {
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
			// Send all events (both operations and sync jobs)
			eventData, err := services.MarshalEventToJSON(event)
			if err != nil {
				h.logger.WithError(err).Error("Failed to marshal activity event")
				continue
			}
			
			c.SSEvent(event.Type, eventData)
			c.Writer.Flush()
			
		case <-ticker.C:
			// Send heartbeat
			c.SSEvent("heartbeat", gin.H{"timestamp": time.Now().Unix()})
			c.Writer.Flush()
		}
	}
}

// Helper functions
func getOperationTitle(opType string) string {
	switch opType {
	case "sync_users":
		return "User Synchronization"
	case "sync_safes":
		return "Safe Synchronization"
	case "sync_groups":
		return "Group Synchronization"
	case "create_safe":
		return "Create Safe"
	case "modify_safe":
		return "Modify Safe"
	case "delete_safe":
		return "Delete Safe"
	case "grant_access":
		return "Grant Access"
	case "revoke_access":
		return "Revoke Access"
	default:
		return opType
	}
}

func getOperationSubtitle(op *gormmodels.Operation) string {
	if op.CyberArkInstance != nil {
		return op.CyberArkInstance.Name
	}
	return ""
}

func getSyncJobTitle(syncType string) string {
	switch syncType {
	case gormmodels.SyncTypeUsers:
		return "User Synchronization"
	case gormmodels.SyncTypeSafes:
		return "Safe Synchronization"
	case gormmodels.SyncTypeGroups:
		return "Group Synchronization"
	default:
		return syncType
	}
}

func getSyncJobSubtitle(job *gormmodels.SyncJob) string {
	subtitle := ""
	if job.CyberArkInstance != nil {
		subtitle = job.CyberArkInstance.Name
	}
	
	if job.RecordsSynced > 0 {
		if subtitle != "" {
			subtitle += " - "
		}
		subtitle += strconv.Itoa(job.RecordsSynced) + " records"
	}
	
	return subtitle
}