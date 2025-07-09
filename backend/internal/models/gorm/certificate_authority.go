package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"gorm.io/gorm"
)

type CertificateAuthority struct {
	ID               string    `gorm:"primaryKey;type:text" json:"id"`
	Name             string    `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Description      string    `gorm:"type:text" json:"description,omitempty"`
	Certificate      string    `gorm:"type:text;not null" json:"certificate"` // PEM encoded certificate chain
	CertificateCount int       `gorm:"not null;default:1" json:"certificate_count"` // Number of certificates in chain
	Fingerprint      string    `gorm:"type:text;not null;uniqueIndex" json:"fingerprint"` // Fingerprint of the primary (first) certificate
	Subject          string    `gorm:"type:text;not null" json:"subject"` // Subject of the primary certificate
	Issuer           string    `gorm:"type:text;not null" json:"issuer"` // Issuer of the primary certificate
	IsRootCA         bool      `gorm:"default:false" json:"is_root_ca"` // True if primary cert is self-signed root
	IsIntermediate   bool      `gorm:"default:false" json:"is_intermediate"` // True if primary cert is intermediate CA
	ChainInfo        string    `gorm:"type:text" json:"chain_info"` // JSON array of all certs in chain
	NotBefore        time.Time `gorm:"not null" json:"not_before"` // NotBefore of the primary certificate
	NotAfter         time.Time `gorm:"not null" json:"not_after"` // NotAfter of the primary certificate
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy        string    `gorm:"type:text;not null" json:"created_by"`
	UpdatedBy        string    `gorm:"type:text;not null" json:"updated_by"`
}

func (ca *CertificateAuthority) BeforeCreate(tx *gorm.DB) error {
	if ca.ID == "" {
		ca.ID = ulid.New(ulid.CAPrefix)
	}
	
	// Set CreatedBy and UpdatedBy from context
	if userID, ok := tx.Statement.Context.Value("user_id").(string); ok {
		ca.CreatedBy = userID
		ca.UpdatedBy = userID
	}
	
	return nil
}

func (ca *CertificateAuthority) BeforeUpdate(tx *gorm.DB) error {
	// Set UpdatedBy from context
	if userID, ok := tx.Statement.Context.Value("user_id").(string); ok {
		ca.UpdatedBy = userID
	}
	return nil
}

func (CertificateAuthority) TableName() string {
	return "certificate_authorities"
}