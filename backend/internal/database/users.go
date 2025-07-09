package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/models"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/pkg/crypto"
	"github.com/orca-ng/orca/pkg/session"
	"github.com/orca-ng/orca/pkg/ulid"
)

// GetUserByUsername retrieves a user by username using GORM
func (db *GormDB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var gormUser gormmodels.User
	
	if err := db.WithContext(ctx).Where("username = ?", username).First(&gormUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return convertToUser(&gormUser), nil
}

// GetUserByID retrieves a user by ID using GORM
func (db *GormDB) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var gormUser gormmodels.User
	
	if err := db.WithContext(ctx).First(&gormUser, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return convertToUser(&gormUser), nil
}

// CreateUser creates a new user using GORM
func (db *GormDB) CreateUser(ctx context.Context, username, password string, isAdmin bool) (*models.User, error) {
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	gormUser := &gormmodels.User{
		ID:           ulid.New(ulid.UserPrefix),
		Username:     username,
		PasswordHash: hashedPassword,
		IsActive:     true,
		IsAdmin:      isAdmin,
	}
	
	if err := db.WithContext(ctx).Create(gormUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return convertToUser(gormUser), nil
}

// UpdateLastLogin updates the last login timestamp for a user using GORM
func (db *GormDB) UpdateLastLogin(ctx context.Context, userID string) error {
	result := db.WithContext(ctx).
		Model(&gormmodels.User{}).
		Where("id = ?", userID).
		Update("last_login_at", time.Now().UTC())
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}
	
	return nil
}

// CreateSession creates a new session using GORM
func (db *GormDB) CreateSession(ctx context.Context, userID, userAgent, ipAddress string, duration time.Duration) (*models.Session, error) {
	token, err := session.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	
	gormSession := &gormmodels.Session{
		ID:        ulid.New(ulid.SessionPrefix),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(duration),
	}
	
	if userAgent != "" {
		gormSession.UserAgent = &userAgent
	}
	if ipAddress != "" {
		gormSession.IPAddress = &ipAddress
	}
	
	if err := db.WithContext(ctx).Create(gormSession).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	return convertToSession(gormSession), nil
}

// GetSessionByToken retrieves a session by token using GORM
func (db *GormDB) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	var gormSession gormmodels.Session
	
	err := db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now().UTC()).
		First(&gormSession).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return convertToSession(&gormSession), nil
}

// DeleteSession deletes a session by token using GORM
func (db *GormDB) DeleteSession(ctx context.Context, token string) error {
	result := db.WithContext(ctx).Where("token = ?", token).Delete(&gormmodels.Session{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}
	
	return nil
}

// DeleteExpiredSessions deletes expired sessions using GORM
func (db *GormDB) DeleteExpiredSessions(ctx context.Context) error {
	result := db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&gormmodels.Session{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", result.Error)
	}
	
	return nil
}

// UpdateUserPassword updates a user's password using GORM
func (db *GormDB) UpdateUserPassword(ctx context.Context, username, newPassword string) error {
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	result := db.WithContext(ctx).
		Model(&gormmodels.User{}).
		Where("username = ?", username).
		Update("password_hash", hashedPassword)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// convertToUser converts GORM user model to regular user model
func convertToUser(gu *gormmodels.User) *models.User {
	return &models.User{
		ID:           gu.ID,
		Username:     gu.Username,
		PasswordHash: gu.PasswordHash,
		CreatedAt:    gu.CreatedAt,
		UpdatedAt:    gu.UpdatedAt,
		LastLoginAt:  gu.LastLoginAt,
		IsActive:     gu.IsActive,
		IsAdmin:      gu.IsAdmin,
	}
}

// convertToSession converts GORM session model to regular session model
func convertToSession(gs *gormmodels.Session) *models.Session {
	return &models.Session{
		ID:        gs.ID,
		UserID:    gs.UserID,
		Token:     gs.Token,
		ExpiresAt: gs.ExpiresAt,
		CreatedAt: gs.CreatedAt,
		UpdatedAt: gs.UpdatedAt,
		UserAgent: gs.UserAgent,
		IPAddress: gs.IPAddress,
	}
}