package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/orca-ng/orca/internal/middleware"
	"github.com/orca-ng/orca/internal/pipeline"
)

// OperationsHandler handles operation-related API endpoints
type OperationsHandler struct {
	store  *pipeline.Store
	logger *logrus.Logger
}

// NewOperationsHandler creates a new operations handler
func NewOperationsHandler(store *pipeline.Store, logger *logrus.Logger) *OperationsHandler {
	return &OperationsHandler{
		store:  store,
		logger: logger,
	}
}

// CreateOperation creates a new operation
func (h *OperationsHandler) CreateOperation(c *gin.Context) {
	var req pipeline.CreateOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get user from context
	user := middleware.GetUser(c)
	
	// Create operation
	op, err := h.store.CreateOperation(c.Request.Context(), &req, &user.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create operation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create operation"})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"operation_id": op.ID,
		"type":        op.Type,
		"priority":    op.Priority,
		"user_id":     user.ID,
	}).Info("Operation created")
	
	// Check if client wants to wait for completion
	if req.Wait {
		timeout := 30 * time.Second // Default timeout
		if req.WaitTimeoutSeconds > 0 {
			timeout = time.Duration(req.WaitTimeoutSeconds) * time.Second
		}
		
		// Wait for operation to complete
		completedOp, err := h.store.WaitForOperation(c.Request.Context(), op.ID, timeout)
		if err != nil {
			// Return current state even if timeout
			h.logger.WithError(err).Warn("Wait for operation failed")
		} else {
			op = completedOp
		}
	}
	
	// Convert to response
	response := h.operationToResponse(op)
	c.JSON(http.StatusCreated, response)
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

// ListOperations lists operations with filtering
func (h *OperationsHandler) ListOperations(c *gin.Context) {
	filters := pipeline.ListOperationsFilters{
		Limit: 100, // Default limit
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
	
	// List operations
	operations, err := h.store.ListOperations(c.Request.Context(), filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list operations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list operations"})
		return
	}
	
	// Convert to responses
	responses := make([]pipeline.OperationResponse, len(operations))
	for i, op := range operations {
		responses[i] = h.operationToResponse(op)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"operations": responses,
		"count":      len(responses),
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
		ErrorMessage: op.ErrorMessage,
		CreatedAt:    op.CreatedAt,
		CompletedAt:  op.CompletedAt,
	}
	
	// Dereference Result if not nil
	if op.Result != nil {
		resp.Result = *op.Result
	}
	
	return resp
}