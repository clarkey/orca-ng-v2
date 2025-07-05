package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orca-ng/orca/internal/models"
	"github.com/orca-ng/orca/pkg/crypto"
	"github.com/orca-ng/orca/pkg/session"
	"github.com/orca-ng/orca/pkg/ulid"
)

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, created_at, updated_at, last_login_at, is_active, is_admin
		FROM users
		WHERE username = $1
	`
	
	row := db.pool.QueryRow(ctx, query, username)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.IsActive,
		&user.IsAdmin,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

func (db *DB) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password_hash, created_at, updated_at, last_login_at, is_active, is_admin
		FROM users
		WHERE id = $1
	`
	
	row := db.pool.QueryRow(ctx, query, userID)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
		&user.IsActive,
		&user.IsAdmin,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

func (db *DB) CreateUser(ctx context.Context, username, password string, isAdmin bool) (*models.User, error) {
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	user := &models.User{
		ID:           ulid.New(ulid.UserPrefix),
		Username:     username,
		PasswordHash: hashedPassword,
		IsActive:     true,
		IsAdmin:      isAdmin,
	}
	
	query := `
		INSERT INTO users (id, username, password_hash, is_active, is_admin)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	
	row := db.pool.QueryRow(ctx, query, user.ID, user.Username, user.PasswordHash, user.IsActive, user.IsAdmin)
	err = row.Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

func (db *DB) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := db.pool.Exec(ctx, query, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (db *DB) CreateSession(ctx context.Context, userID, userAgent, ipAddress string, duration time.Duration) (*models.Session, error) {
	token, err := session.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	
	sess := &models.Session{
		ID:        ulid.New(ulid.SessionPrefix),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(duration),
	}
	
	if userAgent != "" {
		sess.UserAgent = &userAgent
	}
	if ipAddress != "" {
		sess.IPAddress = &ipAddress
	}
	
	query := `
		INSERT INTO sessions (id, user_id, token, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	
	row := db.pool.QueryRow(ctx, query, sess.ID, sess.UserID, sess.Token, sess.ExpiresAt, sess.UserAgent, sess.IPAddress)
	err = row.Scan(&sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	return sess, nil
}

func (db *DB) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	var sess models.Session
	query := `
		SELECT id, user_id, token, expires_at, created_at, updated_at, user_agent, ip_address::text
		FROM sessions
		WHERE token = $1 AND expires_at > $2
	`
	
	row := db.pool.QueryRow(ctx, query, token, time.Now().UTC())
	err := row.Scan(
		&sess.ID,
		&sess.UserID,
		&sess.Token,
		&sess.ExpiresAt,
		&sess.CreatedAt,
		&sess.UpdatedAt,
		&sess.UserAgent,
		&sess.IPAddress,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return &sess, nil
}

func (db *DB) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := db.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (db *DB) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	_, err := db.pool.Exec(ctx, query, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

func (db *DB) UpdateUserPassword(ctx context.Context, username, newPassword string) error {
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	query := `
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE username = $2
	`
	
	result, err := db.pool.Exec(ctx, query, hashedPassword, username)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}