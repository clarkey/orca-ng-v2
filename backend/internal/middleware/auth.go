package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/models"
	"github.com/sirupsen/logrus"
)

func AuthRequired(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Authorization header first
		authHeader := c.GetHeader("Authorization")
		var token string
		
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
		
		// If no Authorization header, check session cookie
		if token == "" {
			var err error
			token, err = c.Cookie("session_token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
				c.Abort()
				return
			}
		}
		
		// Validate session
		session, err := db.GetSessionByToken(c.Request.Context(), token)
		if err != nil {
			logrus.WithError(err).Debug("Invalid session token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			c.Abort()
			return
		}
		
		// Get user
		user, err := db.GetUserByID(c.Request.Context(), session.UserID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get user for session")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}
		
		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
			c.Abort()
			return
		}
		
		// Store user and session in context
		c.Set("user", user)
		c.Set("session", session)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		
		user, ok := userInterface.(*models.User)
		if !ok || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}