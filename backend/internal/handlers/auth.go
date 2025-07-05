package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/models"
	"github.com/orca-ng/orca/pkg/crypto"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	db             *database.DB
	sessionTimeout time.Duration
}

func NewAuthHandler(db *database.DB, sessionTimeout time.Duration) *AuthHandler {
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
	user, err := h.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		logrus.WithError(err).Debug("User not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
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
	session, err := h.db.CreateSession(c.Request.Context(), user.ID, userAgent, clientIP, h.sessionTimeout)
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Update last login
	if err := h.db.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		logrus.WithError(err).Error("Failed to update last login")
		// Don't fail the login for this
	}

	// Set session cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"session_token",
		session.Token,
		int(h.sessionTimeout.Seconds()),
		"/",
		"",
		false, // secure - set to true in production with HTTPS
		true,  // httpOnly
	)

	// Return response - DO NOT include token in response body for security
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"message": "login successful",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from header or cookie
	token := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Extract bearer token
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}
	
	if token == "" {
		token, _ = c.Cookie("session_token")
	}

	if token != "" {
		// Delete session from database
		if err := h.db.DeleteSession(c.Request.Context(), token); err != nil {
			logrus.WithError(err).Error("Failed to delete session")
		}
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

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// LoginCLI handles login for CLI clients that need the token in the response
func (h *AuthHandler) LoginCLI(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get user by username
	user, err := h.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		logrus.WithError(err).Debug("User not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
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
	session, err := h.db.CreateSession(c.Request.Context(), user.ID, userAgent, clientIP, h.sessionTimeout)
	if err != nil {
		logrus.WithError(err).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Update last login
	if err := h.db.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		logrus.WithError(err).Error("Failed to update last login")
		// Don't fail the login for this
	}

	// For CLI, return the token in the response
	c.JSON(http.StatusOK, gin.H{
		"token":      session.Token,
		"expires_at": session.ExpiresAt,
		"user":       user,
	})
}