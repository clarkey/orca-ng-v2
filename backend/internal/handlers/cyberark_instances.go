package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/orca-ng/orca/internal/crypto"
	"github.com/orca-ng/orca/internal/cyberark"
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/middleware"
	"github.com/orca-ng/orca/internal/models"
)

// CyberArkInstancesHandler handles CyberArk instance-related API endpoints
type CyberArkInstancesHandler struct {
	db        *database.DB
	logger    *logrus.Logger
	encryptor *crypto.Encryptor
}

// NewCyberArkInstancesHandler creates a new CyberArk instances handler
func NewCyberArkInstancesHandler(db *database.DB, logger *logrus.Logger, encryptionKey string) *CyberArkInstancesHandler {
	return &CyberArkInstancesHandler{
		db:        db,
		logger:    logger,
		encryptor: crypto.NewEncryptor(encryptionKey),
	}
}

// ListInstances returns all CyberArk instances
func (h *CyberArkInstancesHandler) ListInstances(c *gin.Context) {
	// Check if user wants only active instances
	onlyActive := c.Query("active") == "true"

	instances, err := h.db.GetCyberArkInstances(c.Request.Context(), onlyActive)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get CyberArk instances")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// GetInstance returns a single CyberArk instance
func (h *CyberArkInstancesHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")

	instance, err := h.db.GetCyberArkInstance(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "instance not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// CreateInstance creates a new CyberArk instance
func (h *CyberArkInstancesHandler) CreateInstance(c *gin.Context) {
	var req models.CreateCyberArkInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate URL format
	if err := cyberark.ValidateURL(req.BaseURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if name already exists
	exists, err := h.db.CheckCyberArkInstanceNameExists(c.Request.Context(), req.Name, "")
	if err != nil {
		h.logger.WithError(err).Error("Failed to check instance name")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate instance name"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Instance name already exists"})
		return
	}

	// Test the connection first
	testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	client := cyberark.NewClient(req.BaseURL, req.Username, req.Password)
	success, message, err := client.TestConnection(testCtx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to test CyberArk connection")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Connection test failed: " + err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Connection test failed: " + message})
		return
	}

	// Encrypt the password
	encryptedPassword, err := h.encryptor.Encrypt(req.Password)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure credentials"})
		return
	}

	// Create the instance
	user := middleware.GetUser(c)
	instance := &models.CyberArkInstance{
		Name:              req.Name,
		BaseURL:           req.BaseURL,
		Username:          req.Username,
		PasswordEncrypted: encryptedPassword,
		ConcurrentSessions: true, // Default to true if not specified
		IsActive:          true,
	}
	
	// Override with request value if provided
	if req.ConcurrentSessions != nil {
		instance.ConcurrentSessions = *req.ConcurrentSessions
	}

	if err := h.db.CreateCyberArkInstance(c.Request.Context(), instance, user.ID); err != nil {
		h.logger.WithError(err).Error("Failed to create CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance"})
		return
	}

	// Update test result
	h.db.UpdateCyberArkInstanceTestResult(c.Request.Context(), instance.ID, true, "")

	h.logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"name":        instance.Name,
		"user_id":     user.ID,
	}).Info("CyberArk instance created")

	c.JSON(http.StatusCreated, instance)
}

// UpdateInstance updates an existing CyberArk instance
func (h *CyberArkInstancesHandler) UpdateInstance(c *gin.Context) {
	id := c.Param("id")
	
	var req models.UpdateCyberArkInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the existing instance
	existing, err := h.db.GetCyberArkInstance(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "instance not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	testConnection := false
	newBaseURL := existing.BaseURL
	newUsername := existing.Username
	newPassword := ""

	if req.Name != "" && req.Name != existing.Name {
		// Check if new name already exists
		exists, err := h.db.CheckCyberArkInstanceNameExists(c.Request.Context(), req.Name, id)
		if err != nil {
			h.logger.WithError(err).Error("Failed to check instance name")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate instance name"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Instance name already exists"})
			return
		}
		updates["name"] = req.Name
	}

	if req.BaseURL != "" && req.BaseURL != existing.BaseURL {
		if err := cyberark.ValidateURL(req.BaseURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["base_url"] = req.BaseURL
		newBaseURL = req.BaseURL
		testConnection = true
	}

	if req.Username != "" && req.Username != existing.Username {
		updates["username"] = req.Username
		newUsername = req.Username
		testConnection = true
	}

	if req.Password != "" {
		// Encrypt the new password
		encryptedPassword, err := h.encryptor.Encrypt(req.Password)
		if err != nil {
			h.logger.WithError(err).Error("Failed to encrypt password")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure credentials"})
			return
		}
		updates["password_encrypted"] = encryptedPassword
		newPassword = req.Password
		testConnection = true
	}

	if req.ConcurrentSessions != nil {
		updates["concurrent_sessions"] = *req.ConcurrentSessions
	}
	
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// If connection details changed, test the new connection
	if testConnection {
		// For testing, use the new password if provided, otherwise decrypt the existing one
		testPassword := newPassword
		if testPassword == "" {
			decrypted, err := h.encryptor.Decrypt(existing.PasswordEncrypted)
			if err != nil {
				h.logger.WithError(err).Error("Failed to decrypt password")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve credentials"})
				return
			}
			testPassword = decrypted
		}

		testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		client := cyberark.NewClient(newBaseURL, newUsername, testPassword)
		success, message, err := client.TestConnection(testCtx)
		if err != nil {
			h.logger.WithError(err).Error("Failed to test CyberArk connection")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Connection test failed: " + err.Error()})
			return
		}

		if !success {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Connection test failed: " + message})
			return
		}
	}

	// Update the instance
	user := middleware.GetUser(c)
	if err := h.db.UpdateCyberArkInstance(c.Request.Context(), id, updates, user.ID); err != nil {
		h.logger.WithError(err).Error("Failed to update CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance"})
		return
	}

	// Update test result if connection was tested
	if testConnection {
		h.db.UpdateCyberArkInstanceTestResult(c.Request.Context(), id, true, "")
	}

	h.logger.WithFields(logrus.Fields{
		"instance_id": id,
		"user_id":     user.ID,
		"updates":     len(updates),
	}).Info("CyberArk instance updated")

	c.JSON(http.StatusOK, gin.H{"message": "Instance updated successfully"})
}

// DeleteInstance deletes a CyberArk instance
func (h *CyberArkInstancesHandler) DeleteInstance(c *gin.Context) {
	id := c.Param("id")

	// Check if instance exists
	_, err := h.db.GetCyberArkInstance(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "instance not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	// Delete the instance
	if err := h.db.DeleteCyberArkInstance(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("Failed to delete CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance"})
		return
	}

	user := middleware.GetUser(c)
	h.logger.WithFields(logrus.Fields{
		"instance_id": id,
		"user_id":     user.ID,
	}).Info("CyberArk instance deleted")

	c.JSON(http.StatusOK, gin.H{"message": "Instance deleted successfully"})
}

// TestConnection tests a CyberArk connection
func (h *CyberArkInstancesHandler) TestConnection(c *gin.Context) {
	var req models.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate URL format
	if err := cyberark.ValidateURL(req.BaseURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Test the connection
	testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	client := cyberark.NewClient(req.BaseURL, req.Username, req.Password)
	success, message, err := client.TestConnection(testCtx)
	responseTime := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.WithError(err).Error("Failed to test CyberArk connection")
		c.JSON(http.StatusOK, models.TestConnectionResponse{
			Success:      false,
			Message:      "Connection test failed: " + err.Error(),
			ResponseTime: responseTime,
		})
		return
	}

	c.JSON(http.StatusOK, models.TestConnectionResponse{
		Success:      success,
		Message:      message,
		ResponseTime: responseTime,
	})
}

// TestInstanceConnection tests an existing instance's connection
func (h *CyberArkInstancesHandler) TestInstanceConnection(c *gin.Context) {
	id := c.Param("id")

	// Get the instance
	instance, err := h.db.GetCyberArkInstance(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "instance not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	// Decrypt the password
	password, err := h.encryptor.Decrypt(instance.PasswordEncrypted)
	if err != nil {
		h.logger.WithError(err).Error("Failed to decrypt password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve credentials"})
		return
	}

	// Test the connection
	testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	client := cyberark.NewClient(instance.BaseURL, instance.Username, password)
	success, message, err := client.TestConnection(testCtx)
	responseTime := time.Since(startTime).Milliseconds()

	// Update test result
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else if !success {
		errorMsg = message
	}
	h.db.UpdateCyberArkInstanceTestResult(c.Request.Context(), id, success, errorMsg)

	if err != nil {
		h.logger.WithError(err).Error("Failed to test CyberArk connection")
		c.JSON(http.StatusOK, models.TestConnectionResponse{
			Success:      false,
			Message:      "Connection test failed: " + err.Error(),
			ResponseTime: responseTime,
		})
		return
	}

	c.JSON(http.StatusOK, models.TestConnectionResponse{
		Success:      success,
		Message:      message,
		ResponseTime: responseTime,
	})
}