package models

import (
	"time"
)

type User struct {
	ID          string     `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	IsAdmin     bool       `json:"is_admin" db:"is_admin"`
}

type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UserAgent *string   `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress *string   `json:"ip_address,omitempty" db:"ip_address"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}