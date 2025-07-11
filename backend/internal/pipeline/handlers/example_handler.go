package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/orca-ng/orca/internal/cyberark"
	"github.com/orca-ng/orca/internal/pipeline"
)

// ExampleHandler demonstrates how to implement an operation handler
type ExampleHandler struct {
	// Add any dependencies here
}

// Handle processes the operation
func (h *ExampleHandler) Handle(ctx context.Context, op *pipeline.Operation) error {
	// Extract payload
	var payload struct {
		SafeName    string `json:"safe_name"`
		Description string `json:"description"`
	}
	
	if err := json.Unmarshal(op.Payload, &payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}
	
	// Get CyberArk client from context (injected by processor)
	client, ok := ctx.Value("cyberark_client").(*cyberark.Client)
	if !ok {
		return fmt.Errorf("no CyberArk client in context")
	}
	
	// Check if authenticated
	if !client.IsAuthenticated() {
		return fmt.Errorf("CyberArk client not authenticated")
	}
	
	// TODO: Implement actual CyberArk API calls here
	// For example:
	// err := client.CreateSafe(payload.SafeName, payload.Description)
	// if err != nil {
	//     return fmt.Errorf("failed to create safe: %w", err)
	// }
	
	// Set operation result
	result := map[string]interface{}{
		"safe_name": payload.SafeName,
		"created":   true,
		"message":   "Safe created successfully",
	}
	
	resultJSON, _ := json.Marshal(result)
	resultRaw := json.RawMessage(resultJSON)
	op.Result = &resultRaw
	
	return nil
}

// CanRetry determines if the error is retryable
func (h *ExampleHandler) CanRetry(err error) bool {
	// Define which errors are retryable
	// For example: network errors, temporary failures, etc.
	
	// Check for specific error types
	if err == nil {
		return false
	}
	
	errMsg := err.Error()
	
	// Retry on network/connection errors
	if contains(errMsg, []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"service unavailable",
	}) {
		return true
	}
	
	// Don't retry on authentication or authorization errors
	if contains(errMsg, []string{
		"unauthorized",
		"forbidden",
		"invalid credentials",
		"authentication failed",
	}) {
		return false
	}
	
	// Default to not retrying
	return false
}

// contains checks if the string contains any of the substrings
func contains(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) && s[len(s)-len(substr):] == substr {
			return true
		}
	}
	return false
}