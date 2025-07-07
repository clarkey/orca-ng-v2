package pipeline

import (
	"context"
	"encoding/json"
	"time"
)

// Priority represents operation priority levels
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// Status represents operation status
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

// OperationType represents different types of operations
type OperationType string

const (
	OpTypeSafeProvision   OperationType = "safe_provision"
	OpTypeSafeModify      OperationType = "safe_modify"
	OpTypeSafeDelete      OperationType = "safe_delete"
	OpTypeAccessGrant     OperationType = "access_grant"
	OpTypeAccessRevoke    OperationType = "access_revoke"
	OpTypeUserSync        OperationType = "user_sync"
	OpTypeSafeSync        OperationType = "safe_sync"
	OpTypeGroupSync       OperationType = "group_sync"
)

// Operation represents a queued operation in the pipeline
type Operation struct {
	ID                 string           `json:"id" db:"id"`
	Type               OperationType    `json:"type" db:"type"`
	Priority           Priority         `json:"priority" db:"priority"`
	Status             Status           `json:"status" db:"status"`
	Payload            json.RawMessage  `json:"payload" db:"payload"`
	Result             *json.RawMessage `json:"result,omitempty" db:"result"`
	ErrorMessage       *string          `json:"error_message,omitempty" db:"error_message"`
	RetryCount         int              `json:"retry_count" db:"retry_count"`
	MaxRetries         int              `json:"max_retries" db:"max_retries"`
	ScheduledAt        time.Time        `json:"scheduled_at" db:"scheduled_at"`
	StartedAt          *time.Time       `json:"started_at,omitempty" db:"started_at"`
	CompletedAt        *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	CreatedBy          *string          `json:"created_by,omitempty" db:"created_by"`
	CyberArkInstanceID *string          `json:"cyberark_instance_id,omitempty" db:"cyberark_instance_id"`
	CorrelationID      *string          `json:"correlation_id,omitempty" db:"correlation_id"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}

// OperationHandler defines the interface for handling specific operation types
type OperationHandler interface {
	// Handle processes the operation
	Handle(ctx context.Context, op *Operation) error
	
	// CanRetry determines if an error is retryable
	CanRetry(err error) bool
	
	// ValidatePayload validates the operation payload
	ValidatePayload(payload json.RawMessage) error
}

// PipelineConfig represents the pipeline configuration
type PipelineConfig struct {
	// Total processing capacity
	TotalCapacity int `json:"total_capacity"`
	
	// Priority allocation (percentages)
	PriorityAllocation map[Priority]float64 `json:"priority_allocation"`
	
	// Retry policy
	RetryPolicy RetryPolicy `json:"retry_policy"`
	
	// Operation timeouts in seconds
	OperationTimeouts map[OperationType]int `json:"operation_timeouts"`
	
	// Default timeout for operations not in the map
	DefaultTimeout int `json:"default_timeout"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts        int  `json:"max_attempts"`
	BackoffBaseSeconds int  `json:"backoff_base_seconds"`
	BackoffMultiplier  int  `json:"backoff_multiplier"`
	BackoffJitter      bool `json:"backoff_jitter"`
}

// ProcessingMetrics holds pipeline performance metrics
type ProcessingMetrics struct {
	QueueDepth         map[Priority]int       `json:"queue_depth"`
	ProcessingCount    map[Priority]int       `json:"processing_count"`
	CompletedCount     map[OperationType]int64 `json:"completed_count"`
	FailedCount        map[OperationType]int64 `json:"failed_count"`
	AvgProcessingTime  map[OperationType]float64 `json:"avg_processing_time"`
	WorkerUtilization  float64                `json:"worker_utilization"`
}

// CreateOperationRequest represents an API request to create an operation
type CreateOperationRequest struct {
	Type               OperationType   `json:"type" binding:"required"`
	Priority           Priority        `json:"priority"`
	Payload            json.RawMessage `json:"payload" binding:"required"`
	Wait               bool            `json:"wait"`
	WaitTimeoutSeconds int             `json:"wait_timeout_seconds"`
	ScheduledAt        *time.Time      `json:"scheduled_at,omitempty"`
	CorrelationID      *string         `json:"correlation_id,omitempty"`
}

// OperationResponse represents an API response for an operation
type OperationResponse struct {
	ID           string          `json:"id"`
	Type         OperationType   `json:"type"`
	Priority     Priority        `json:"priority"`
	Status       Status          `json:"status"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	ScheduledAt  time.Time       `json:"scheduled_at"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CreatedBy    *string         `json:"created_by,omitempty"`
	CreatedByUser *UserInfo      `json:"created_by_user,omitempty"`
}

// UserInfo represents basic user information for API responses
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}