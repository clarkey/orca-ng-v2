package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	
	"github.com/orca-ng/orca/internal/database"
	"github.com/orca-ng/orca/internal/models"
	"github.com/orca-ng/orca/internal/services"
	"github.com/orca-ng/orca/pkg/ulid"
)

type CertificateAuthoritiesHandler struct {
	db          *database.DB
	logger      *logrus.Logger
	certService *services.CertificateService
}

func NewCertificateAuthoritiesHandler(db *database.DB, logger *logrus.Logger) *CertificateAuthoritiesHandler {
	return &CertificateAuthoritiesHandler{
		db:          db,
		logger:      logger,
		certService: services.NewCertificateService(),
	}
}

func (h *CertificateAuthoritiesHandler) List(c *gin.Context) {
	query := `
		SELECT id, name, description, certificate, fingerprint, subject, issuer,
		       not_before, not_after, is_active, created_at, updated_at, created_by, updated_by
		FROM certificate_authorities
		ORDER BY name ASC
	`
	
	rows, err := h.db.Pool().Query(c.Request.Context(), query)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch certificate authorities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch certificate authorities"})
		return
	}
	defer rows.Close()
	
	var authorities []models.CertificateAuthority
	for rows.Next() {
		var ca models.CertificateAuthority
		err := rows.Scan(
			&ca.ID, &ca.Name, &ca.Description, &ca.Certificate, &ca.Fingerprint,
			&ca.Subject, &ca.Issuer, &ca.NotBefore, &ca.NotAfter, &ca.IsActive,
			&ca.CreatedAt, &ca.UpdatedAt, &ca.CreatedBy, &ca.UpdatedBy,
		)
		if err != nil {
			h.logger.WithError(err).Error("Failed to scan certificate authority")
			continue
		}
		authorities = append(authorities, ca)
	}
	
	// Convert to info objects (without the actual certificate data)
	infos := make([]*models.CertificateAuthorityInfo, len(authorities))
	for i, ca := range authorities {
		infos[i] = ca.ToInfo()
	}
	
	c.JSON(http.StatusOK, gin.H{"certificate_authorities": infos})
}

func (h *CertificateAuthoritiesHandler) Get(c *gin.Context) {
	id := c.Param("id")
	
	query := `
		SELECT id, name, description, certificate, fingerprint, subject, issuer,
		       not_before, not_after, is_active, created_at, updated_at, created_by, updated_by
		FROM certificate_authorities
		WHERE id = $1
	`
	
	var ca models.CertificateAuthority
	err := h.db.Pool().QueryRow(c.Request.Context(), query, id).Scan(
		&ca.ID, &ca.Name, &ca.Description, &ca.Certificate, &ca.Fingerprint,
		&ca.Subject, &ca.Issuer, &ca.NotBefore, &ca.NotAfter, &ca.IsActive,
		&ca.CreatedAt, &ca.UpdatedAt, &ca.CreatedBy, &ca.UpdatedBy,
	)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate authority not found"})
		return
	}
	
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch certificate authority"})
		return
	}
	
	c.JSON(http.StatusOK, ca)
}

func (h *CertificateAuthoritiesHandler) Create(c *gin.Context) {
	var req models.CreateCertificateAuthorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parse and validate the certificate
	parsed, err := h.certService.ParseCertificate(req.Certificate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate: " + err.Error()})
		return
	}
	
	// Validate it's a CA certificate
	if err := h.certService.ValidateCertificate(req.Certificate); err != nil {
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
	
	ca := models.CertificateAuthority{
		ID:          ulid.New(ulid.CAPrefix),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Certificate: parsed.PEMData,
		Fingerprint: parsed.Fingerprint,
		Subject:     parsed.Subject,
		Issuer:      parsed.Issuer,
		NotBefore:   parsed.NotBefore,
		NotAfter:    parsed.NotAfter,
		IsActive:    isActive,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	
	query := `
		INSERT INTO certificate_authorities (
			id, name, description, certificate, fingerprint, subject, issuer,
			not_before, not_after, is_active, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`
	
	err = h.db.Pool().QueryRow(
		c.Request.Context(), query,
		ca.ID, ca.Name, ca.Description, ca.Certificate, ca.Fingerprint,
		ca.Subject, ca.Issuer, ca.NotBefore, ca.NotAfter, ca.IsActive,
		ca.CreatedBy, ca.UpdatedBy,
	).Scan(&ca.CreatedAt, &ca.UpdatedAt)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			if strings.Contains(err.Error(), "certificate_authorities_name_unique") {
				c.JSON(http.StatusConflict, gin.H{"error": "A certificate authority with this name already exists"})
			} else if strings.Contains(err.Error(), "certificate_authorities_fingerprint_unique") {
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
	
	c.JSON(http.StatusCreated, ca)
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
	
	// Build dynamic update query
	var updates []string
	var args []interface{}
	argCount := 1
	
	if req.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, strings.TrimSpace(req.Name))
		argCount++
	}
	
	if req.Description != "" {
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, strings.TrimSpace(req.Description))
		argCount++
	}
	
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *req.IsActive)
		argCount++
	}
	
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Add updated_by
	updates = append(updates, fmt.Sprintf("updated_by = $%d", argCount))
	args = append(args, userID)
	argCount++
	
	// Add WHERE clause
	args = append(args, id)
	
	query := fmt.Sprintf(`
		UPDATE certificate_authorities
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, certificate, fingerprint, subject, issuer,
		          not_before, not_after, is_active, created_at, updated_at, created_by, updated_by
	`, strings.Join(updates, ", "), argCount)
	
	var ca models.CertificateAuthority
	err := h.db.Pool().QueryRow(c.Request.Context(), query, args...).Scan(
		&ca.ID, &ca.Name, &ca.Description, &ca.Certificate, &ca.Fingerprint,
		&ca.Subject, &ca.Issuer, &ca.NotBefore, &ca.NotAfter, &ca.IsActive,
		&ca.CreatedAt, &ca.UpdatedAt, &ca.CreatedBy, &ca.UpdatedBy,
	)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate authority not found"})
		return
	}
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "A certificate authority with this name already exists"})
			return
		}
		h.logger.WithError(err).Error("Failed to update certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update certificate authority"})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"ca_id": ca.ID,
		"user_id": userID,
	}).Info("Certificate authority updated")
	
	c.JSON(http.StatusOK, ca)
}

func (h *CertificateAuthoritiesHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	
	query := `DELETE FROM certificate_authorities WHERE id = $1`
	
	result, err := h.db.Pool().Exec(c.Request.Context(), query, id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete certificate authority")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete certificate authority"})
		return
	}
	
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate authority not found"})
		return
	}
	
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)
	h.logger.WithFields(logrus.Fields{
		"ca_id": id,
		"user_id": user.ID,
	}).Info("Certificate authority deleted")
	
	c.JSON(http.StatusOK, gin.H{"message": "Certificate authority deleted successfully"})
}