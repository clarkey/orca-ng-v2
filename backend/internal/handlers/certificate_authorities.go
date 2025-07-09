package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/models"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/services"
)

type CertificateAuthoritiesHandler struct {
	db          *database.GormDB
	logger      *logrus.Logger
	certService *services.CertificateService
	certManager *services.CertificateManager
}

func NewCertificateAuthoritiesHandler(db *database.GormDB, logger *logrus.Logger, certManager *services.CertificateManager) *CertificateAuthoritiesHandler {
	return &CertificateAuthoritiesHandler{
		db:          db,
		logger:      logger,
		certService: services.NewCertificateService(),
		certManager: certManager,
	}
}

func (h *CertificateAuthoritiesHandler) List(c *gin.Context) {
	var authorities []gormmodels.CertificateAuthority
	
	if err := h.db.Order("name ASC").Find(&authorities).Error; err != nil {
		h.logger.WithError(err).Error("Failed to fetch certificate authorities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch certificate authorities"})
		return
	}
	
	// Convert to response format
	var response []models.CertificateAuthority
	for _, ca := range authorities {
		response = append(response, models.CertificateAuthority{
			ID:               ca.ID,
			Name:             ca.Name,
			Description:      ca.Description,
			Certificate:      ca.Certificate,
			CertificateCount: ca.CertificateCount,
			Fingerprint:      ca.Fingerprint,
			Subject:          ca.Subject,
			Issuer:           ca.Issuer,
			IsRootCA:         ca.IsRootCA,
			IsIntermediate:   ca.IsIntermediate,
			ChainInfo:        ca.ChainInfo,
			NotBefore:        ca.NotBefore,
			NotAfter:         ca.NotAfter,
			IsActive:         ca.IsActive,
			CreatedAt:        ca.CreatedAt,
			UpdatedAt:        ca.UpdatedAt,
			CreatedBy:        ca.CreatedBy,
			UpdatedBy:        ca.UpdatedBy,
		})
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *CertificateAuthoritiesHandler) Get(c *gin.Context) {
	id := c.Param("id")
	
	var ca gormmodels.CertificateAuthority
	if err := h.db.First(&ca, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Certificate authority not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get certificate authority"})
		return
	}
	
	// Convert to response format
	response := models.CertificateAuthority{
		ID:               ca.ID,
		Name:             ca.Name,
		Description:      ca.Description,
		Certificate:      ca.Certificate,
		CertificateCount: ca.CertificateCount,
		Fingerprint:      ca.Fingerprint,
		Subject:          ca.Subject,
		Issuer:           ca.Issuer,
		IsRootCA:         ca.IsRootCA,
		IsIntermediate:   ca.IsIntermediate,
		ChainInfo:        ca.ChainInfo,
		NotBefore:        ca.NotBefore,
		NotAfter:         ca.NotAfter,
		IsActive:         ca.IsActive,
		CreatedAt:        ca.CreatedAt,
		UpdatedAt:        ca.UpdatedAt,
		CreatedBy:        ca.CreatedBy,
		UpdatedBy:        ca.UpdatedBy,
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *CertificateAuthoritiesHandler) Create(c *gin.Context) {
	var req models.CreateCertificateAuthorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parse and validate the certificate chain
	chainParsed, err := h.certService.ParseCertificateChain(req.Certificate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate: " + err.Error()})
		return
	}
	
	// Validate the certificate chain
	if err := h.certService.ValidateCertificateChain(req.Certificate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get the current user
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)
	userID := user.ID
	
	// Set default for IsActive if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	
	// Build chain info JSON
	var chainInfo []models.CertificateChainInfo
	for _, cert := range chainParsed.Certificates {
		chainInfo = append(chainInfo, models.CertificateChainInfo{
			Subject:      cert.Subject,
			Issuer:       cert.Issuer,
			Fingerprint:  cert.Fingerprint,
			NotBefore:    cert.NotBefore,
			NotAfter:     cert.NotAfter,
			IsCA:         cert.IsCA,
			IsSelfSigned: cert.IsSelfSigned,
		})
	}
	
	chainInfoJSON, _ := json.Marshal(chainInfo)
	
	// The primary certificate is the first one in the chain
	primaryCert := chainParsed.PrimaryCert
	
	ca := gormmodels.CertificateAuthority{
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		Certificate:      req.Certificate, // Store the full chain
		CertificateCount: len(chainParsed.Certificates),
		Fingerprint:      primaryCert.Fingerprint,
		Subject:          primaryCert.Subject,
		Issuer:           primaryCert.Issuer,
		IsRootCA:         primaryCert.IsSelfSigned && primaryCert.IsCA,
		IsIntermediate:   !primaryCert.IsSelfSigned && primaryCert.IsCA,
		ChainInfo:        string(chainInfoJSON),
		NotBefore:        primaryCert.NotBefore,
		NotAfter:         primaryCert.NotAfter,
		IsActive:         isActive,
	}
	
	// Create with user context
	ctx := context.WithValue(c.Request.Context(), "user_id", userID)
	if err := h.db.WithContext(ctx).Create(&ca).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "UNIQUE constraint") {
			if strings.Contains(err.Error(), "name") {
				c.JSON(http.StatusConflict, gin.H{"error": "A certificate authority with this name already exists"})
			} else if strings.Contains(err.Error(), "fingerprint") {
				c.JSON(http.StatusConflict, gin.H{"error": "This certificate is already registered"})
			} else {
				c.JSON(http.StatusConflict, gin.H{"error": "Certificate authority already exists"})
			}
			return
		}
		h.logger.WithError(err).Error("Failed to create certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create certificate authority"})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"ca_id": ca.ID,
		"name":  ca.Name,
		"fingerprint": ca.Fingerprint,
		"user_id": userID,
	}).Info("Certificate authority created")
	
	// Force refresh the certificate pool to immediately use the new certificate
	if err := h.certManager.ForceRefresh(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to refresh certificate pool after create")
		// Don't fail the request, the certificate was created successfully
	}
	
	// Convert to response format
	response := models.CertificateAuthority{
		ID:               ca.ID,
		Name:             ca.Name,
		Description:      ca.Description,
		Certificate:      ca.Certificate,
		CertificateCount: ca.CertificateCount,
		Fingerprint:      ca.Fingerprint,
		Subject:          ca.Subject,
		Issuer:           ca.Issuer,
		IsRootCA:         ca.IsRootCA,
		IsIntermediate:   ca.IsIntermediate,
		ChainInfo:        ca.ChainInfo,
		NotBefore:        ca.NotBefore,
		NotAfter:         ca.NotAfter,
		IsActive:         ca.IsActive,
		CreatedAt:        ca.CreatedAt,
		UpdatedAt:        ca.UpdatedAt,
		CreatedBy:        ca.CreatedBy,
		UpdatedBy:        ca.UpdatedBy,
	}
	
	c.JSON(http.StatusCreated, response)
}

func (h *CertificateAuthoritiesHandler) Update(c *gin.Context) {
	id := c.Param("id")
	
	var req models.UpdateCertificateAuthorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get the current user
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)
	userID := user.ID
	
	// Build update map
	updates := make(map[string]interface{})
	
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Update with user context
	ctx := context.WithValue(c.Request.Context(), "user_id", userID)
	
	var ca gormmodels.CertificateAuthority
	result := h.db.WithContext(ctx).Model(&ca).Where("id = ?", id).Updates(updates)
	
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value") || strings.Contains(result.Error.Error(), "UNIQUE constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "A certificate authority with this name already exists"})
			return
		}
		h.logger.WithError(result.Error).Error("Failed to update certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update certificate authority"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate authority not found"})
		return
	}
	
	// Fetch the updated record
	if err := h.db.First(&ca, "id = ?", id).Error; err != nil {
		h.logger.WithError(err).Error("Failed to fetch updated certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated certificate authority"})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"ca_id": ca.ID,
		"user_id": userID,
	}).Info("Certificate authority updated")
	
	// Force refresh the certificate pool if the active status changed
	if req.IsActive != nil {
		if err := h.certManager.ForceRefresh(c.Request.Context()); err != nil {
			h.logger.WithError(err).Error("Failed to refresh certificate pool after update")
			// Don't fail the request, the certificate was updated successfully
		}
	}
	
	// Convert to response format
	response := models.CertificateAuthority{
		ID:               ca.ID,
		Name:             ca.Name,
		Description:      ca.Description,
		Certificate:      ca.Certificate,
		CertificateCount: ca.CertificateCount,
		Fingerprint:      ca.Fingerprint,
		Subject:          ca.Subject,
		Issuer:           ca.Issuer,
		IsRootCA:         ca.IsRootCA,
		IsIntermediate:   ca.IsIntermediate,
		ChainInfo:        ca.ChainInfo,
		NotBefore:        ca.NotBefore,
		NotAfter:         ca.NotAfter,
		IsActive:         ca.IsActive,
		CreatedAt:        ca.CreatedAt,
		UpdatedAt:        ca.UpdatedAt,
		CreatedBy:        ca.CreatedBy,
		UpdatedBy:        ca.UpdatedBy,
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *CertificateAuthoritiesHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	
	result := h.db.Delete(&gormmodels.CertificateAuthority{}, "id = ?", id)
	if result.Error != nil {
		h.logger.WithError(result.Error).Error("Failed to delete certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete certificate authority"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate authority not found"})
		return
	}
	
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)
	h.logger.WithFields(logrus.Fields{
		"ca_id": id,
		"user_id": user.ID,
	}).Info("Certificate authority deleted")
	
	// Force refresh the certificate pool to remove the deleted certificate
	if err := h.certManager.ForceRefresh(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to refresh certificate pool after delete")
		// Don't fail the request, the certificate was deleted successfully
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Certificate authority deleted successfully"})
}

// RefreshPool forces a refresh of the certificate pool
// This is useful when certificates have been modified directly in the database
func (h *CertificateAuthoritiesHandler) RefreshPool(c *gin.Context) {
	// Get the current user for logging
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)
	
	h.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"action": "manual_cert_pool_refresh",
	}).Info("Manual certificate pool refresh requested")
	
	// Force refresh the certificate pool
	if err := h.certManager.ForceRefresh(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to refresh certificate pool")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh certificate pool"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Certificate pool refreshed successfully"})
}