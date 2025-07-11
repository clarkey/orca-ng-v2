package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"gorm.io/gorm"
)

// CyberArkGroupMembership represents a user's membership in a group
type CyberArkGroupMembership struct {
	ID                 string     `gorm:"primaryKey;size:30" json:"id"`
	CyberArkInstanceID string     `gorm:"column:cyberark_instance_id;size:30;not null;index" json:"cyberark_instance_id"`
	UserID             string     `gorm:"size:255;not null;index" json:"user_id"` // CyberArk's internal user ID
	Username           string     `gorm:"size:255;not null" json:"username"`
	GroupID            int        `gorm:"not null;index" json:"group_id"` // CyberArk's internal group ID
	GroupName          string     `gorm:"size:255;not null" json:"group_name"`
	GroupType          string     `gorm:"size:50;not null" json:"group_type"` // Vault, Directory, etc.
	
	// Sync metadata
	LastSyncedAt time.Time `gorm:"not null" json:"last_synced_at"`
	IsDeleted    bool      `gorm:"default:false" json:"is_deleted"` // soft delete for removed memberships
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	CyberArkInstance *CyberArkInstance `gorm:"foreignKey:CyberArkInstanceID" json:"-"`
}

// BeforeCreate generates ULID for new group membership
func (gm *CyberArkGroupMembership) BeforeCreate(tx *gorm.DB) error {
	if gm.ID == "" {
		gm.ID = ulid.New(ulid.GroupMembershipPrefix)
	}
	return nil
}

// TableName specifies the table name for GORM
func (CyberArkGroupMembership) TableName() string {
	return "cyberark_group_memberships"
}