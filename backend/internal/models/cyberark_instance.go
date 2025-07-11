package models

import (
	"database/sql"
	"time"
)

// CyberArkInstance represents a CyberArk PVWA instance configuration
type CyberArkInstance struct {
	ID                string         `db:"id" json:"id"`
	Name              string         `db:"name" json:"name"`
	BaseURL           string         `db:"base_url" json:"base_url"`
	Username          string         `db:"username" json:"username"`
	PasswordEncrypted string         `db:"password_encrypted" json:"-"` // Never expose in JSON
	ConcurrentSessions bool          `db:"concurrent_sessions" json:"concurrent_sessions"`
	SkipTLSVerify     bool           `db:"skip_tls_verify" json:"skip_tls_verify"`
	IsActive          bool           `db:"is_active" json:"is_active"`
	LastTestAt        *time.Time     `db:"last_test_at" json:"last_test_at,omitempty"`
	LastTestSuccess   sql.NullBool   `db:"last_test_success" json:"last_test_success,omitempty"`
	LastTestError     sql.NullString `db:"last_test_error" json:"last_test_error,omitempty"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
	CreatedBy         sql.NullString `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy         sql.NullString `db:"updated_by" json:"updated_by,omitempty"`
}

// CreateCyberArkInstanceRequest represents the request to create a new instance
type CreateCyberArkInstanceRequest struct {
	Name     string `json:"name" binding:"required"`
	BaseURL  string `json:"base_url" binding:"required,url"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	ConcurrentSessions *bool `json:"concurrent_sessions"`
	SkipTLSVerify *bool `json:"skip_tls_verify"`
	UserSyncPageSize *int `json:"user_sync_page_size" binding:"omitempty,min=1,max=1000"`
}

// UpdateCyberArkInstanceRequest represents the request to update an instance
type UpdateCyberArkInstanceRequest struct {
	Name     string `json:"name,omitempty"`
	BaseURL  string `json:"base_url,omitempty" binding:"omitempty,url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	ConcurrentSessions *bool `json:"concurrent_sessions,omitempty"`
	SkipTLSVerify *bool `json:"skip_tls_verify,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	UserSyncPageSize *int `json:"user_sync_page_size,omitempty" binding:"omitempty,min=1,max=1000"`
}

// TestConnectionRequest represents the request to test a CyberArk connection
type TestConnectionRequest struct {
	BaseURL  string `json:"base_url" binding:"required,url"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	SkipTLSVerify bool `json:"skip_tls_verify"`
}

// TestConnectionResponse represents the response from testing a connection
type TestConnectionResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ResponseTime int64  `json:"response_time_ms"`
	Version      string `json:"version,omitempty"`
}