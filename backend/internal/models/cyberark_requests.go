package models

import "time"

// CyberArkInstanceInfo is the response model for CyberArk instance (without password)
type CyberArkInstanceInfo struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	BaseURL            string    `json:"base_url"`
	Username           string    `json:"username"`
	ConcurrentSessions bool      `json:"concurrent_sessions"`
	SkipTLSVerify      bool      `json:"skip_tls_verify"`
	IsActive           bool      `json:"is_active"`
	LastTestAt         *time.Time `json:"last_test_at,omitempty"`
	LastTestSuccess    *bool      `json:"last_test_success,omitempty"`
	LastTestError      *string    `json:"last_test_error,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// TestCyberArkConnectionRequest is the request model for testing a CyberArk connection
type TestCyberArkConnectionRequest struct {
	BaseURL       string  `json:"base_url" binding:"required"`
	Username      string  `json:"username" binding:"required"`
	Password      string  `json:"password" binding:"required"`
	SkipTLSVerify *bool   `json:"skip_tls_verify,omitempty"`
}