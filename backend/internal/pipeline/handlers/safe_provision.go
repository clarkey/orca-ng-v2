package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/orca-ng/orca/internal/pipeline"
)

// SafeProvisionRequest represents the payload for safe provisioning
type SafeProvisionRequest struct {
	SafeName            string                 `json:"safe_name" binding:"required"`
	Description         string                 `json:"description"`
	CyberArkInstanceID  string                 `json:"cyberark_instance_id" binding:"required"`
	ManagingCPM         string                 `json:"managing_cpm"`
	NumberOfDaysRetention int                  `json:"number_of_days_retention"`
	Permissions         []SafePermission       `json:"permissions"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// SafePermission represents permissions to set on a safe
type SafePermission struct {
	UserOrGroup string            `json:"user_or_group" binding:"required"`
	IsGroup     bool              `json:"is_group"`
	Permissions map[string]bool   `json:"permissions" binding:"required"`
}

// SafeProvisionResult represents the result of safe provisioning
type SafeProvisionResult struct {
	SafeID      string    `json:"safe_id"`
	SafeName    string    `json:"safe_name"`
	CreatedAt   time.Time `json:"created_at"`
	Permissions int       `json:"permissions_set"`
}

// SafeProvisionHandler handles safe provisioning operations
type SafeProvisionHandler struct {
	// In a real implementation, this would have CyberArk client
	// cyberarkClient *cyberark.Client
}

// NewSafeProvisionHandler creates a new safe provision handler
func NewSafeProvisionHandler() *SafeProvisionHandler {
	return &SafeProvisionHandler{}
}

// Handle processes the safe provisioning operation
func (h *SafeProvisionHandler) Handle(ctx context.Context, op *pipeline.Operation) error {
	// Parse payload
	var req SafeProvisionRequest
	if err := json.Unmarshal(op.Payload, &req); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	
	// Validate safe name
	if len(req.SafeName) < 3 || len(req.SafeName) > 28 {
		return fmt.Errorf("safe name must be between 3 and 28 characters")
	}
	
	// TODO: In real implementation, this would:
	// 1. Get CyberArk client for the specified instance
	// 2. Create the safe using CyberArk REST API
	// 3. Set permissions for each user/group
	// 4. Handle any CyberArk-specific errors
	
	// Simulate processing time
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
		// Simulate work
	}
	
	// Create result
	result := SafeProvisionResult{
		SafeID:      fmt.Sprintf("SAFE_%s_%d", req.SafeName, time.Now().Unix()),
		SafeName:    req.SafeName,
		CreatedAt:   time.Now(),
		Permissions: len(req.Permissions),
	}
	
	// Marshal result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	
	// Update operation result
	op.Result = (*json.RawMessage)(&resultJSON)
	
	return nil
}

// CanRetry determines if an error is retryable
func (h *SafeProvisionHandler) CanRetry(err error) bool {
	// In real implementation, check for:
	// - Network errors
	// - Timeout errors
	// - HTTP 429 (rate limit)
	// - HTTP 503 (service unavailable)
	// - Specific CyberArk temporary errors
	
	// For now, just check if it's a timeout
	if err == context.DeadlineExceeded {
		return true
	}
	
	// Check for specific HTTP status codes
	if httpErr, ok := err.(interface{ StatusCode() int }); ok {
		switch httpErr.StatusCode() {
		case http.StatusTooManyRequests,
		     http.StatusServiceUnavailable,
		     http.StatusGatewayTimeout:
			return true
		}
	}
	
	return false
}

// ValidatePayload validates the operation payload
func (h *SafeProvisionHandler) ValidatePayload(payload json.RawMessage) error {
	var req SafeProvisionRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	
	// Validate required fields
	if req.SafeName == "" {
		return fmt.Errorf("safe_name is required")
	}
	
	if req.CyberArkInstanceID == "" {
		return fmt.Errorf("cyberark_instance_id is required")
	}
	
	// Validate safe name constraints
	if len(req.SafeName) < 3 || len(req.SafeName) > 28 {
		return fmt.Errorf("safe name must be between 3 and 28 characters")
	}
	
	// Validate permissions
	for i, perm := range req.Permissions {
		if perm.UserOrGroup == "" {
			return fmt.Errorf("permissions[%d].user_or_group is required", i)
		}
		if len(perm.Permissions) == 0 {
			return fmt.Errorf("permissions[%d].permissions cannot be empty", i)
		}
	}
	
	return nil
}