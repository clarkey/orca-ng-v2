package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/middleware"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
)

// OperationsHandlerGorm handles operation-related API endpoints
type OperationsHandler struct {
	db     *database.GormDB
	logger *logrus.Logger
}

// NewOperationsHandlerGorm creates a new operations handler
func NewOperationsHandler(db *database.GormDB, logger *logrus.Logger) *OperationsHandler {
	return &OperationsHandler{
		db:     db,
		logger: logger,
	}
}

// GetOperation retrieves an operation by ID
func (h *OperationsHandler) GetOperation(c *gin.Context) {
	id := c.Param("id")
	
	var op gormmodels.Operation
	if err := h.db.
		Preload("Creator").
		Preload("CyberArkInstance").
		First(&op, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get operation"})
		return
	}
	
	response := h.operationToResponse(&op)
	c.JSON(http.StatusOK, response)
}

// ListOperations lists operations with filtering and pagination
func (h *OperationsHandler) ListOperations(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize
	
	// Build query
	query := h.db.Model(&gormmodels.Operation{}).
		Preload("Creator").
		Preload("CyberArkInstance")
	
	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if opType := c.Query("type"); opType != "" {
		query = query.Where("type = ?", opType)
	}
	
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}
	
	if correlationID := c.Query("correlation_id"); correlationID != "" {
		query = query.Where("correlation_id = ?", correlationID)
	}
	
	// Parse date range
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}
	
	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count operations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count operations"})
		return
	}
	
	// Apply sorting
	sortBy := "created_at"
	sortOrder := "desc"
	if sb := c.Query("sort_by"); sb != "" {
		// Validate sort field
		validSortFields := map[string]bool{
			"created_at": true,
			"scheduled_at": true,
			"started_at": true,
			"completed_at": true,
			"type": true,
			"status": true,
			"priority": true,
		}
		if validSortFields[sb] {
			sortBy = sb
		}
	}
	if so := c.Query("sort_order"); so == "asc" || so == "desc" {
		sortOrder = so
	}
	query = query.Order(sortBy + " " + sortOrder)
	
	// Apply pagination
	var operations []gormmodels.Operation
	if err := query.Offset(offset).Limit(pageSize).Find(&operations).Error; err != nil {
		h.logger.WithError(err).Error("Failed to list operations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list operations"})
		return
	}
	
	// Convert to responses
	responses := make([]interface{}, len(operations))
	for i, op := range operations {
		responses[i] = h.operationToResponse(&op)
	}
	
	// Calculate pagination metadata
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	
	c.JSON(http.StatusOK, gin.H{
		"operations": responses,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// CancelOperation cancels a pending or processing operation
func (h *OperationsHandler) CancelOperation(c *gin.Context) {
	id := c.Param("id")
	
	// Update operation status to cancelled only if it's pending or processing
	result := h.db.Model(&gormmodels.Operation{}).
		Where("id = ? AND status IN (?, ?)", id, gormmodels.OpStatusPending, gormmodels.OpStatusProcessing).
		Updates(map[string]interface{}{
			"status": gormmodels.OpStatusCancelled,
			"completed_at": time.Now(),
		})
	
	if result.Error != nil {
		h.logger.WithError(result.Error).Error("Failed to cancel operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel operation"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation cannot be cancelled (not found or already completed)"})
		return
	}
	
	h.logger.WithField("operation_id", id).Info("Operation cancelled")
	c.JSON(http.StatusOK, gin.H{"message": "Operation cancelled"})
}

// CreateOperation creates a new operation
func (h *OperationsHandler) CreateOperation(c *gin.Context) {
	var req struct {
		Type               string                 `json:"type" binding:"required"`
		Priority           string                 `json:"priority" binding:"required,oneof=low normal medium high"`
		Payload            map[string]interface{} `json:"payload" binding:"required"`
		CyberArkInstanceID *string                `json:"cyberark_instance_id"`
		CorrelationID      *string                `json:"correlation_id"`
		ScheduledAt        *time.Time             `json:"scheduled_at"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get current user
	user := middleware.GetUser(c)
	
	// Convert payload to JSON
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}
	
	operation := &gormmodels.Operation{
		Type:               req.Type,
		Priority:           req.Priority,
		Status:             gormmodels.OpStatusPending,
		Payload:            payloadJSON,
		CyberArkInstanceID: req.CyberArkInstanceID,
		CorrelationID:      req.CorrelationID,
	}
	
	if req.ScheduledAt != nil {
		operation.ScheduledAt = *req.ScheduledAt
	} else {
		operation.ScheduledAt = time.Now()
	}
	
	// Create with user context
	ctx := context.WithValue(c.Request.Context(), "user_id", user.ID)
	if err := h.db.WithContext(ctx).Create(operation).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation"})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"operation_id": operation.ID,
		"type":         operation.Type,
		"priority":     operation.Priority,
		"user_id":      user.ID,
	}).Info("Operation created")
	
	c.JSON(http.StatusCreated, h.operationToResponse(operation))
}

// operationToResponse converts an operation to API response format
func (h *OperationsHandler) operationToResponse(op *gormmodels.Operation) interface{} {
	resp := map[string]interface{}{
		"id":           op.ID,
		"type":         op.Type,
		"priority":     op.Priority,
		"status":       op.Status,
		"payload":      op.Payload,
		"result":       op.Result,
		"error_message": op.ErrorMessage,
		"retry_count":   op.RetryCount,
		"max_retries":   op.MaxRetries,
		"scheduled_at":  op.ScheduledAt,
		"started_at":    op.StartedAt,
		"completed_at":  op.CompletedAt,
		"created_at":    op.CreatedAt,
		"updated_at":    op.UpdatedAt,
		"created_by":    op.CreatedBy,
		"cyberark_instance_id": op.CyberArkInstanceID,
		"correlation_id": op.CorrelationID,
	}
	
	// Add user info if available
	if op.Creator != nil {
		resp["created_by_user"] = map[string]interface{}{
			"id":       op.Creator.ID,
			"username": op.Creator.Username,
		}
	}
	
	// Add CyberArk instance info if available
	if op.CyberArkInstance != nil {
		resp["cyberark_instance_info"] = map[string]interface{}{
			"id":   op.CyberArkInstance.ID,
			"name": op.CyberArkInstance.Name,
		}
	}
	
	return resp
}

// UpdatePriority updates the priority of an operation
func (h *OperationsHandler) UpdatePriority(c *gin.Context) {
	id := c.Param("id")
	
	var req struct {
		Priority string `json:"priority" binding:"required,oneof=low normal high"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get the operation
	var operation gormmodels.Operation
	if err := h.db.First(&operation, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get operation"})
		return
	}
	
	// Check if operation can be updated (only pending or processing)
	if operation.Status != gormmodels.OpStatusPending && operation.Status != gormmodels.OpStatusProcessing {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only update priority for pending or processing operations"})
		return
	}
	
	// Update priority
	if err := h.db.Model(&operation).Update("priority", req.Priority).Error; err != nil {
		h.logger.WithError(err).Error("Failed to update operation priority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update priority"})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"operation_id": operation.ID,
		"old_priority": operation.Priority,
		"new_priority": req.Priority,
	}).Info("Operation priority updated")
	
	c.JSON(http.StatusOK, gin.H{"message": "Priority updated successfully"})
}