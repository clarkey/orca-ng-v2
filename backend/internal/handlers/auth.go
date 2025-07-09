package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/models"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/pkg/crypto"
	"github.com/orca-ng/orca/pkg/session"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	db             *database.GormDB
	sessionTimeout time.Duration
}

func NewAuthHandler(db *database.GormDB, sessionTimeout time.Duration) *AuthHandler {
	return &AuthHandler{
		db:             db,
		sessionTimeout: sessionTimeout,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get user by username
	var user gormmodels.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithField("username", req.Username).Debug("User not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		logrus.WithError(err).Error("Failed to query user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Verify password
	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		logrus.WithError(err).Debug("Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	// Create session
	userAgent := c.Request.UserAgent()
	clientIP := c.ClientIP()
	
	token, err := session.GenerateToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate session token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	
	sess := &gormmodels.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(h.sessionTimeout),
	}
	
	if userAgent != "" {
		sess.UserAgent = &userAgent
	}
	if clientIP != "" {
		sess.IPAddress = &clientIP
	}
	
	if err := h.db.Create(sess).Error; err != nil {
		logrus.WithError(err).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Update last login
	now := time.Now().UTC()
	if err := h.db.Model(&user).Update("last_login_at", now).Error; err != nil {
		logrus.WithError(err).Error("Failed to update last login")
		// Don't fail the login for this
	}

	// Log session creation for debugging
	logrus.WithFields(logrus.Fields{
		"session_token": sess.Token,
		"user_id": user.ID,
		"expires_at": sess.ExpiresAt,
	}).Debug("Session created successfully")
	
	// Set session cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"session_token",
		sess.Token,
		int(h.sessionTimeout.Seconds()),
		"/",
		"",
		false, // secure - set to true in production with HTTPS
		true,  // httpOnly
	)

	// Convert to old model format for API compatibility
	userResponse := models.User{
		ID:          user.ID,
		Username:    user.Username,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
		IsActive:    user.IsActive,
		IsAdmin:     user.IsAdmin,
	}

	// Return response - DO NOT include token in response body for security
	c.JSON(http.StatusOK, gin.H{
		"user": userResponse,
		"message": "login successful",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from header or cookie
	token := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else if cookie, err := c.Cookie("session_token"); err == nil {
		token = cookie
	}

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no session token provided"})
		return
	}

	// Delete session
	if err := h.db.Where("token = ?", token).Delete(&gormmodels.Session{}).Error; err != nil {
		logrus.WithError(err).Error("Failed to delete session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	// Clear cookie
	c.SetCookie(
		"session_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *AuthHandler) LoginCLI(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get user by username
	var user gormmodels.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		logrus.WithError(err).Error("Failed to query user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Verify password
	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	// Create session
	userAgent := "CLI"
	clientIP := c.ClientIP()
	
	token, err := session.GenerateToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate session token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	
	sess := &gormmodels.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(h.sessionTimeout),
		UserAgent: &userAgent,
	}
	
	if clientIP != "" {
		sess.IPAddress = &clientIP
	}
	
	if err := h.db.Create(sess).Error; err != nil {
		logrus.WithError(err).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Update last login
	now := time.Now().UTC()
	if err := h.db.Model(&user).Update("last_login_at", now).Error; err != nil {
		logrus.WithError(err).Error("Failed to update last login")
		// Don't fail the login for this
	}

	// For CLI, return the token in the response
	c.JSON(http.StatusOK, gin.H{
		"token": sess.Token,
		"expires_at": sess.ExpiresAt,
		"user": gin.H{
			"id": user.ID,
			"username": user.Username,
			"is_admin": user.IsAdmin,
		},
	})
}

func (h *AuthHandler) GetSessionByToken(ctx context.Context, token string) (*gormmodels.Session, *gormmodels.User, error) {
	var sess gormmodels.Session
	
	// Get session with user preloaded
	if err := h.db.WithContext(ctx).
		Preload("User").
		Where("token = ? AND expires_at > ?", token, time.Now().UTC()).
		First(&sess).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("session not found or expired")
		}
		return nil, nil, err
	}
	
	return &sess, &sess.User, nil
}

func (h *AuthHandler) DeleteExpiredSessions(ctx context.Context) error {
	return h.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&gormmodels.Session{}).Error
}