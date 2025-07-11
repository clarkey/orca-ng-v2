package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"gorm.io/gorm"
)

// CyberArkUser represents a user synchronized from a CyberArk instance
type CyberArkUser struct {
	ID                    string     `gorm:"primaryKey;size:30" json:"id"`
	CyberArkInstanceID    string     `gorm:"column:cyberark_instance_id;size:30;not null;index" json:"cyberark_instance_id"`
	Username              string     `gorm:"size:255;not null" json:"username"`
	UserID                string     `gorm:"size:255;not null" json:"user_id"` // CyberArk's internal ID
	FirstName             *string    `gorm:"size:255" json:"first_name,omitempty"`
	LastName              *string    `gorm:"size:255" json:"last_name,omitempty"`
	Email                 *string    `gorm:"size:255" json:"email,omitempty"`
	UserType              string     `gorm:"size:50;not null" json:"user_type"` // EPVUser, ExternalUser, etc.
	Location              *string    `gorm:"size:255" json:"location,omitempty"`
	ComponentUser         bool       `gorm:"default:false" json:"component_user"`
	Suspended             bool       `gorm:"default:false" json:"suspended"`
	EnableUser            bool       `gorm:"default:true" json:"enable_user"`
	ChangePassOnNextLogon bool       `gorm:"default:false" json:"change_pass_on_next_logon"`
	ExpiryDate            *time.Time `json:"expiry_date,omitempty"`
	LastSuccessfulLoginAt *time.Time `json:"last_successful_login_at,omitempty"`
	
	// Sync metadata
	LastSyncedAt         time.Time  `gorm:"not null" json:"last_synced_at"`
	CyberArkLastModified *time.Time `gorm:"column:cyberark_last_modified" json:"cyberark_last_modified,omitempty"`
	IsDeleted            bool       `gorm:"default:false" json:"is_deleted"` // soft delete for removed users
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	CyberArkInstance *CyberArkInstance `gorm:"foreignKey:CyberArkInstanceID" json:"-"`
}

// BeforeCreate generates ULID for new users
func (u *CyberArkUser) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = ulid.New(ulid.CyberArkUserPrefix)
	}
	return nil
}

// TableName specifies the table name for GORM
func (CyberArkUser) TableName() string {
	return "cyberark_users"
}