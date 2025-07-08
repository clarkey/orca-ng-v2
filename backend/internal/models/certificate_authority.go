package models

import (
	"time"
)

type CertificateAuthority struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	Certificate string    `json:"certificate" db:"certificate"` // PEM encoded
	Fingerprint string    `json:"fingerprint" db:"fingerprint"`  // SHA256
	Subject     string    `json:"subject" db:"subject"`
	Issuer      string    `json:"issuer" db:"issuer"`
	NotBefore   time.Time `json:"not_before" db:"not_before"`
	NotAfter    time.Time `json:"not_after" db:"not_after"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
}

type CreateCertificateAuthorityRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Certificate string `json:"certificate" binding:"required"` // PEM encoded certificate
	IsActive    *bool  `json:"is_active"`                      // Defaults to true if not specified
}

type UpdateCertificateAuthorityRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	IsActive    *bool  `json:"is_active"`
}

type CertificateAuthorityInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Fingerprint  string    `json:"fingerprint"`
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	IsActive     bool      `json:"is_active"`
	IsExpired    bool      `json:"is_expired"`
	ExpiresInDays int      `json:"expires_in_days"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (ca *CertificateAuthority) ToInfo() *CertificateAuthorityInfo {
	now := time.Now()
	isExpired := now.After(ca.NotAfter)
	expiresInDays := 0
	
	if !isExpired {
		expiresInDays = int(ca.NotAfter.Sub(now).Hours() / 24)
	}
	
	return &CertificateAuthorityInfo{
		ID:            ca.ID,
		Name:          ca.Name,
		Description:   ca.Description,
		Fingerprint:   ca.Fingerprint,
		Subject:       ca.Subject,
		Issuer:        ca.Issuer,
		NotBefore:     ca.NotBefore,
		NotAfter:      ca.NotAfter,
		IsActive:      ca.IsActive,
		IsExpired:     isExpired,
		ExpiresInDays: expiresInDays,
		CreatedAt:     ca.CreatedAt,
		UpdatedAt:     ca.UpdatedAt,
	}
}