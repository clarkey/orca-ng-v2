package models

import (
	"encoding/json"
	"time"
)

type CertificateAuthority struct {
	ID               string    `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Description      string    `json:"description,omitempty" db:"description"`
	Certificate      string    `json:"certificate" db:"certificate"` // PEM encoded certificate chain
	CertificateCount int       `json:"certificate_count" db:"certificate_count"` // Number of certificates in chain
	Fingerprint      string    `json:"fingerprint" db:"fingerprint"`  // SHA256 of primary certificate
	Subject          string    `json:"subject" db:"subject"` // Subject of primary certificate
	Issuer           string    `json:"issuer" db:"issuer"` // Issuer of primary certificate
	IsRootCA         bool      `json:"is_root_ca" db:"is_root_ca"` // True if primary cert is self-signed
	IsIntermediate   bool      `json:"is_intermediate" db:"is_intermediate"` // True if primary cert is intermediate
	ChainInfo        string    `json:"chain_info" db:"chain_info"` // JSON array with info about all certs
	NotBefore        time.Time `json:"not_before" db:"not_before"` // NotBefore of primary certificate
	NotAfter         time.Time `json:"not_after" db:"not_after"` // NotAfter of primary certificate
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy        string    `json:"created_by" db:"created_by"`
	UpdatedBy        string    `json:"updated_by" db:"updated_by"`
}

type CertificateChainInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	Fingerprint  string    `json:"fingerprint"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	IsCA         bool      `json:"is_ca"`
	IsSelfSigned bool      `json:"is_self_signed"`
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
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	CertificateCount int                    `json:"certificate_count"`
	Fingerprint      string                 `json:"fingerprint"`
	Subject          string                 `json:"subject"`
	Issuer           string                 `json:"issuer"`
	IsRootCA         bool                   `json:"is_root_ca"`
	IsIntermediate   bool                   `json:"is_intermediate"`
	ChainInfo        []CertificateChainInfo `json:"chain_info,omitempty"`
	NotBefore        time.Time              `json:"not_before"`
	NotAfter         time.Time              `json:"not_after"`
	IsActive         bool                   `json:"is_active"`
	IsExpired        bool                   `json:"is_expired"`
	ExpiresInDays    int                    `json:"expires_in_days"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

func (ca *CertificateAuthority) ToInfo() *CertificateAuthorityInfo {
	now := time.Now()
	isExpired := now.After(ca.NotAfter)
	expiresInDays := 0
	
	if !isExpired {
		expiresInDays = int(ca.NotAfter.Sub(now).Hours() / 24)
	}
	
	// Parse chain info if available
	var chainInfo []CertificateChainInfo
	if ca.ChainInfo != "" {
		_ = json.Unmarshal([]byte(ca.ChainInfo), &chainInfo)
	}
	
	return &CertificateAuthorityInfo{
		ID:               ca.ID,
		Name:             ca.Name,
		Description:      ca.Description,
		CertificateCount: ca.CertificateCount,
		Fingerprint:      ca.Fingerprint,
		Subject:          ca.Subject,
		Issuer:           ca.Issuer,
		IsRootCA:         ca.IsRootCA,
		IsIntermediate:   ca.IsIntermediate,
		ChainInfo:        chainInfo,
		NotBefore:        ca.NotBefore,
		NotAfter:         ca.NotAfter,
		IsActive:         ca.IsActive,
		IsExpired:        isExpired,
		ExpiresInDays:    expiresInDays,
		CreatedAt:        ca.CreatedAt,
		UpdatedAt:        ca.UpdatedAt,
	}
}