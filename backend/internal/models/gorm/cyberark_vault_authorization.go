package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"gorm.io/gorm"
)

// CyberArkVaultAuthorization represents a user's vault authorization
type CyberArkVaultAuthorization struct {
	ID                 string     `gorm:"primaryKey;size:30" json:"id"`
	CyberArkInstanceID string     `gorm:"column:cyberark_instance_id;size:30;not null;index" json:"cyberark_instance_id"`
	UserID             string     `gorm:"size:255;not null;index" json:"user_id"` // CyberArk's internal user ID
	Username           string     `gorm:"size:255;not null" json:"username"`
	Authorization      string     `gorm:"size:255;not null" json:"authorization"`
	
	// Sync metadata
	LastSyncedAt time.Time `gorm:"not null" json:"last_synced_at"`
	IsDeleted    bool      `gorm:"default:false" json:"is_deleted"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	CyberArkInstance *CyberArkInstance `gorm:"foreignKey:CyberArkInstanceID" json:"-"`
}

// BeforeCreate generates ULID for new authorization
func (va *CyberArkVaultAuthorization) BeforeCreate(tx *gorm.DB) error {
	if va.ID == "" {
		va.ID = ulid.New(ulid.VaultAuthPrefix)
	}
	return nil
}

// TableName specifies the table name for GORM
func (CyberArkVaultAuthorization) TableName() string {
	return "cyberark_vault_authorizations"
}