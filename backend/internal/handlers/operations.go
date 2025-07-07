package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/orca-ng/orca/internal/middleware"
	"github.com/orca-ng/orca/internal/pipeline"
)

// OperationsHandler handles operation-related API endpoints
type OperationsHandler struct {
	store  *pipeline.Store
	db     *sql.DB
	logger *logrus.Logger
}

// NewOperationsHandler creates a new operations handler
func NewOperationsHandler(store *pipeline.Store, db *sql.DB, logger *logrus.Logger) *OperationsHandler {
	return &OperationsHandler{
		store:  store,
		db:     db,
		logger: logger,
	}
}

// GetOperation retrieves an operation by ID
func (h *OperationsHandler) GetOperation(c *gin.Context) {
	id := c.Param("id")
	
	op, err := h.store.GetOperation(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "operation not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get operation"})
		return
	}
	
	response := h.operationToResponse(op)
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
	
	filters := pipeline.ListOperationsFilters{
		Limit:  pageSize,
		Offset: offset,
	}
	
	// Parse query parameters
	if status := c.Query("status"); status != "" {
		s := pipeline.Status(status)
		filters.Status = &s
	}
	
	if opType := c.Query("type"); opType != "" {
		t := pipeline.OperationType(opType)
		filters.Type = &t
	}
	
	if priority := c.Query("priority"); priority != "" {
		p := pipeline.Priority(priority)
		filters.Priority = &p
	}
	
	if correlationID := c.Query("correlation_id"); correlationID != "" {
		filters.CorrelationID = &correlationID
	}
	
	// Parse search query
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}
	
	// Parse date range
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters.StartDate = &t
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters.EndDate = &t
		}
	}
	
	// Parse sorting
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = &sortBy
	}
	
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = &sortOrder
	}
	
	// Get operations
	operations, err := h.store.ListOperations(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list operations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list operations"})
		return
	}
	
	// Get total count
	countFilters := filters
	countFilters.Limit = 0  // Remove limit for count
	countFilters.Offset = 0
	totalCount, err := h.store.CountOperations(c.Request.Context(), countFilters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count operations")
		totalCount = len(operations) // Fallback
	}
	
	// Convert to responses
	responses := make([]pipeline.OperationResponse, len(operations))
	for i, op := range operations {
		responses[i] = h.operationToResponse(op)
	}
	
	// Calculate pagination metadata
	totalPages := (totalCount + pageSize - 1) / pageSize
	
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
	
	err := h.store.CancelOperation(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "operation cannot be cancelled (not found or already completed)" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.logger.WithError(err).Error("Failed to cancel operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel operation"})
		return
	}
	
	h.logger.WithField("operation_id", id).Info("Operation cancelled")
	c.JSON(http.StatusOK, gin.H{"message": "Operation cancelled"})
}

// GetPipelineMetrics returns pipeline processing metrics
func (h *OperationsHandler) GetPipelineMetrics(c *gin.Context) {
	// This would be connected to the processor's GetMetrics method
	// For now, return sample metrics
	metrics := pipeline.ProcessingMetrics{
		QueueDepth: map[pipeline.Priority]int{
			pipeline.PriorityHigh:   5,
			pipeline.PriorityMedium: 12,
			pipeline.PriorityNormal: 23,
			pipeline.PriorityLow:    8,
		},
		ProcessingCount: map[pipeline.Priority]int{
			pipeline.PriorityHigh:   2,
			pipeline.PriorityMedium: 3,
			pipeline.PriorityNormal: 1,
			pipeline.PriorityLow:    0,
		},
		CompletedCount: map[pipeline.OperationType]int64{
			pipeline.OpTypeSafeProvision: 145,
			pipeline.OpTypeAccessGrant:   892,
			pipeline.OpTypeUserSync:      24,
		},
		FailedCount: map[pipeline.OperationType]int64{
			pipeline.OpTypeSafeProvision: 3,
			pipeline.OpTypeAccessGrant:   12,
			pipeline.OpTypeUserSync:      1,
		},
		AvgProcessingTime: map[pipeline.OperationType]float64{
			pipeline.OpTypeSafeProvision: 2.5,
			pipeline.OpTypeAccessGrant:   0.8,
			pipeline.OpTypeUserSync:      5.2,
		},
		WorkerUtilization: 0.65,
	}
	
	c.JSON(http.StatusOK, metrics)
}

// GetPipelineConfig returns the current pipeline configuration
func (h *OperationsHandler) GetPipelineConfig(c *gin.Context) {
	config, err := h.store.GetPipelineConfig(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pipeline config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get configuration"})
		return
	}
	
	c.JSON(http.StatusOK, config)
}

// GetOperationStats returns aggregated statistics about operations
func (h *OperationsHandler) GetOperationStats(c *gin.Context) {
	// Parse query parameters for time range
	var startDate, endDate *time.Time
	if start := c.Query("start_date"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endDate = &t
		}
	}
	
	// Get operation statistics
	stats, err := h.store.GetOperationStats(c.Request.Context(), startDate, endDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get operation stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get operation statistics"})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

// UpdatePipelineConfig updates pipeline configuration
func (h *OperationsHandler) UpdatePipelineConfig(c *gin.Context) {
	var updates map[string]json.RawMessage
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Update each configuration key
	for key, value := range updates {
		switch key {
		case "processing_capacity":
			var capacityConfig struct {
				Total              int                         `json:"total"`
				PriorityAllocation map[pipeline.Priority]float64 `json:"priority_allocation"`
			}
			if err := json.Unmarshal(value, &capacityConfig); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid processing_capacity format"})
				return
			}
			
			// Validate allocation adds up to 1.0
			total := 0.0
			for _, v := range capacityConfig.PriorityAllocation {
				total += v
			}
			if total < 0.99 || total > 1.01 { // Allow small floating point error
				c.JSON(http.StatusBadRequest, gin.H{"error": "Priority allocation must sum to 1.0"})
				return
			}
			
			if err := h.store.UpdatePipelineConfig(c.Request.Context(), key, capacityConfig); err != nil {
				h.logger.WithError(err).Error("Failed to update config")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
				return
			}
			
		case "retry_policy":
			var retryPolicy pipeline.RetryPolicy
			if err := json.Unmarshal(value, &retryPolicy); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid retry_policy format"})
				return
			}
			
			if err := h.store.UpdatePipelineConfig(c.Request.Context(), key, retryPolicy); err != nil {
				h.logger.WithError(err).Error("Failed to update config")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
				return
			}
			
		case "operation_timeouts":
			var timeouts map[string]interface{}
			if err := json.Unmarshal(value, &timeouts); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation_timeouts format"})
				return
			}
			
			if err := h.store.UpdatePipelineConfig(c.Request.Context(), key, timeouts); err != nil {
				h.logger.WithError(err).Error("Failed to update config")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
				return
			}
			
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown configuration key: " + key})
			return
		}
	}
	
	h.logger.Info("Pipeline configuration updated")
	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated"})
}

// operationToResponse converts an operation to API response format
func (h *OperationsHandler) operationToResponse(op *pipeline.Operation) pipeline.OperationResponse {
	resp := pipeline.OperationResponse{
		ID:           op.ID,
		Type:         op.Type,
		Priority:     op.Priority,
		Status:       op.Status,
		Payload:      op.Payload,
		ErrorMessage: op.ErrorMessage,
		ScheduledAt:  op.ScheduledAt,
		StartedAt:    op.StartedAt,
		CreatedAt:    op.CreatedAt,
		CompletedAt:        op.CompletedAt,
		CreatedBy:          op.CreatedBy,
		CyberArkInstanceID: op.CyberArkInstanceID,
	}
	
	// Dereference Result if not nil
	if op.Result != nil {
		resp.Result = *op.Result
	}
	
	// Fetch user info if created_by is set
	if op.CreatedBy != nil && *op.CreatedBy != "" {
		var username string
		err := h.db.QueryRow("SELECT username FROM users WHERE id = $1", *op.CreatedBy).Scan(&username)
		if err == nil {
			resp.CreatedByUser = &pipeline.UserInfo{
				ID:       *op.CreatedBy,
				Username: username,
			}
		} else {
			h.logger.WithError(err).Warn("Failed to fetch user info for operation")
		}
	}
	
	// Fetch CyberArk instance info if cyberark_instance_id is set
	if op.CyberArkInstanceID != nil && *op.CyberArkInstanceID != "" {
		var instanceName string
		err := h.db.QueryRow("SELECT name FROM cyberark_instances WHERE id = $1", *op.CyberArkInstanceID).Scan(&instanceName)
		if err == nil {
			resp.CyberArkInstanceInfo = &pipeline.CyberArkInstanceInfo{
				ID:   *op.CyberArkInstanceID,
				Name: instanceName,
			}
		} else {
			h.logger.WithError(err).Debug("Failed to fetch CyberArk instance info for operation")
		}
	}
	
	return resp
}