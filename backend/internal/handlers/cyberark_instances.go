package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/orca-ng/orca/internal/crypto"
	"github.com/orca-ng/orca/internal/cyberark"
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/middleware"
	"github.com/orca-ng/orca/internal/models"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/services"
)

// CyberArkInstancesHandlerGorm handles CyberArk instance-related API endpoints
type CyberArkInstancesHandler struct {
	db        *database.GormDB
	logger    *logrus.Logger
	encryptor *crypto.Encryptor
	certManager *services.CertificateManager
}

// NewCyberArkInstancesHandlerGorm creates a new CyberArk instances handler
func NewCyberArkInstancesHandler(db *database.GormDB, logger *logrus.Logger, encryptionKey string, certManager *services.CertificateManager) *CyberArkInstancesHandler {
	return &CyberArkInstancesHandler{
		db:        db,
		logger:    logger,
		encryptor: crypto.NewEncryptor(encryptionKey),
		certManager: certManager,
	}
}

// ListInstances returns all CyberArk instances
func (h *CyberArkInstancesHandler) ListInstances(c *gin.Context) {
	// Check if user wants only active instances
	onlyActive := c.Query("active") == "true"

	query := h.db.Model(&gormmodels.CyberArkInstance{})
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	var instances []gormmodels.CyberArkInstance
	if err := query.Order("name ASC").Find(&instances).Error; err != nil {
		h.logger.WithError(err).Error("Failed to get CyberArk instances")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instances"})
		return
	}

	// Convert to response format (decrypt passwords)
	var response []models.CyberArkInstanceInfo
	for _, inst := range instances {
		info := models.CyberArkInstanceInfo{
			ID:                 inst.ID,
			Name:               inst.Name,
			BaseURL:            inst.BaseURL,
			Username:           inst.Username,
			ConcurrentSessions: inst.ConcurrentSessions,
			SkipTLSVerify:      inst.SkipTLSVerify,
			IsActive:           inst.IsActive,
			UserSyncPageSize:   inst.UserSyncPageSize,
			LastTestAt:         inst.LastTestAt,
			LastTestSuccess:    inst.LastTestSuccess,
			LastTestError:      inst.LastTestError,
			CreatedAt:          inst.CreatedAt,
			UpdatedAt:          inst.UpdatedAt,
		}
		response = append(response, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": response,
		"count":     len(response),
	})
}

// GetInstance returns a single CyberArk instance
func (h *CyberArkInstancesHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")

	var instance gormmodels.CyberArkInstance
	if err := h.db.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	// Convert to response format
	response := models.CyberArkInstanceInfo{
		ID:                 instance.ID,
		Name:               instance.Name,
		BaseURL:            instance.BaseURL,
		Username:           instance.Username,
		ConcurrentSessions: instance.ConcurrentSessions,
		SkipTLSVerify:      instance.SkipTLSVerify,
		IsActive:           instance.IsActive,
		UserSyncPageSize:   instance.UserSyncPageSize,
		LastTestAt:         instance.LastTestAt,
		LastTestSuccess:    instance.LastTestSuccess,
		LastTestError:      instance.LastTestError,
		CreatedAt:          instance.CreatedAt,
		UpdatedAt:          instance.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
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
	var count int64
	if err := h.db.Model(&gormmodels.CyberArkInstance{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
		h.logger.WithError(err).Error("Failed to check instance name")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate instance name"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Instance name already exists"})
		return
	}

	// Test the connection first
	testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	skipTLS := false
	if req.SkipTLSVerify != nil {
		skipTLS = *req.SkipTLSVerify
	}
	// Create client with certificate manager
	clientFactory := func() (*http.Client, error) {
		return h.certManager.GetHTTPClient(testCtx, skipTLS, 30*time.Second)
	}
	client := cyberark.NewClientWithHTTPClientFactory(req.BaseURL, req.Username, req.Password, clientFactory)
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
	instance := &gormmodels.CyberArkInstance{
		Name:              req.Name,
		BaseURL:           req.BaseURL,
		Username:          req.Username,
		PasswordEncrypted: encryptedPassword,
		ConcurrentSessions: true, // Default to true if not specified
		SkipTLSVerify:     false, // Default to false if not specified
		IsActive:          true,
	}
	
	// Override with request values if provided
	if req.ConcurrentSessions != nil {
		instance.ConcurrentSessions = *req.ConcurrentSessions
	}
	if req.SkipTLSVerify != nil {
		instance.SkipTLSVerify = *req.SkipTLSVerify
	}
	if req.UserSyncPageSize != nil {
		instance.UserSyncPageSize = req.UserSyncPageSize
	}

	// Create with user context
	ctx := context.WithValue(c.Request.Context(), "user_id", user.ID)
	if err := h.db.WithContext(ctx).Create(instance).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create instance"})
		return
	}

	// Update test result
	testSuccess := true
	h.db.Model(instance).Updates(map[string]interface{}{
		"last_test_at":      time.Now(),
		"last_test_success": &testSuccess,
		"last_test_error":   nil,
	})

	h.logger.WithFields(logrus.Fields{
		"instance_id": instance.ID,
		"name":        instance.Name,
		"user_id":     user.ID,
	}).Info("CyberArk instance created")

	// Convert to response format
	response := models.CyberArkInstanceInfo{
		ID:                 instance.ID,
		Name:               instance.Name,
		BaseURL:            instance.BaseURL,
		Username:           instance.Username,
		ConcurrentSessions: instance.ConcurrentSessions,
		SkipTLSVerify:      instance.SkipTLSVerify,
		IsActive:           instance.IsActive,
		UserSyncPageSize:   instance.UserSyncPageSize,
		LastTestAt:         instance.LastTestAt,
		LastTestSuccess:    instance.LastTestSuccess,
		LastTestError:      instance.LastTestError,
		CreatedAt:          instance.CreatedAt,
		UpdatedAt:          instance.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
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
	var existing gormmodels.CyberArkInstance
	if err := h.db.First(&existing, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
		var count int64
		if err := h.db.Model(&gormmodels.CyberArkInstance{}).
			Where("name = ? AND id != ?", req.Name, id).
			Count(&count).Error; err != nil {
			h.logger.WithError(err).Error("Failed to check instance name")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate instance name"})
			return
		}
		if count > 0 {
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
	
	if req.SkipTLSVerify != nil {
		updates["skip_tls_verify"] = *req.SkipTLSVerify
	}
	
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	
	if req.UserSyncPageSize != nil {
		updates["user_sync_page_size"] = *req.UserSyncPageSize
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

		// Use existing skip_tls_verify setting if not being updated
		skipTLS := existing.SkipTLSVerify
		if req.SkipTLSVerify != nil {
			skipTLS = *req.SkipTLSVerify
		}
		// Create client with certificate manager
		clientFactory := func() (*http.Client, error) {
			return h.certManager.GetHTTPClient(testCtx, skipTLS, 30*time.Second)
		}
		client := cyberark.NewClientWithHTTPClientFactory(newBaseURL, newUsername, testPassword, clientFactory)
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
	ctx := context.WithValue(c.Request.Context(), "user_id", user.ID)
	
	if err := h.db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
		h.logger.WithError(err).Error("Failed to update CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance"})
		return
	}

	// Update test result if connection was tested
	if testConnection {
		testSuccess := true
		h.db.Model(&existing).Updates(map[string]interface{}{
			"last_test_at":      time.Now(),
			"last_test_success": &testSuccess,
			"last_test_error":   nil,
		})
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
	var instance gormmodels.CyberArkInstance
	if err := h.db.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	// Delete the instance
	if err := h.db.Delete(&instance).Error; err != nil {
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
	var req models.TestCyberArkConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate URL
	if err := cyberark.ValidateURL(req.BaseURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	skipTLS := false
	if req.SkipTLSVerify != nil {
		skipTLS = *req.SkipTLSVerify
	}

	// Create client with certificate manager
	clientFactory := func() (*http.Client, error) {
		return h.certManager.GetHTTPClient(testCtx, skipTLS, 30*time.Second)
	}
	client := cyberark.NewClientWithHTTPClientFactory(req.BaseURL, req.Username, req.Password, clientFactory)
	
	success, message, err := client.TestConnection(testCtx)
	if err != nil {
		h.logger.WithError(err).Debug("Connection test failed")
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": message,
	})
}

// TestInstanceConnection tests an existing instance's connection
func (h *CyberArkInstancesHandler) TestInstanceConnection(c *gin.Context) {
	id := c.Param("id")

	// Get the instance
	var instance gormmodels.CyberArkInstance
	if err := h.db.First(&instance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get CyberArk instance")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instance"})
		return
	}

	// Decrypt password
	password, err := h.encryptor.Decrypt(instance.PasswordEncrypted)
	if err != nil {
		h.logger.WithError(err).Error("Failed to decrypt password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve credentials"})
		return
	}

	testCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Create client with certificate manager
	clientFactory := func() (*http.Client, error) {
		return h.certManager.GetHTTPClient(testCtx, instance.SkipTLSVerify, 30*time.Second)
	}
	client := cyberark.NewClientWithHTTPClientFactory(instance.BaseURL, instance.Username, password, clientFactory)
	
	success, message, err := client.TestConnection(testCtx)
	testResult := map[string]interface{}{
		"last_test_at": time.Now(),
	}
	
	if err != nil {
		h.logger.WithError(err).Debug("Instance connection test failed")
		testSuccess := false
		errMsg := err.Error()
		testResult["last_test_success"] = &testSuccess
		testResult["last_test_error"] = &errMsg
	} else {
		testResult["last_test_success"] = &success
		if !success {
			testResult["last_test_error"] = &message
		} else {
			testResult["last_test_error"] = nil
		}
	}

	// Update test result
	h.db.Model(&instance).Updates(testResult)

	if err != nil || !success {
		responseMsg := message
		if err != nil {
			responseMsg = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": responseMsg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}