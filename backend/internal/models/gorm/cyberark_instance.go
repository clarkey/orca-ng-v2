package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"gorm.io/gorm"
)

type CyberArkInstance struct {
	ID                  string     `gorm:"primaryKey;size:30" json:"id"`
	Name                string     `gorm:"size:255;not null;uniqueIndex" json:"name"`
	BaseURL             string     `gorm:"type:text;not null" json:"base_url"`
	Username            string     `gorm:"size:255;not null" json:"username"`
	PasswordEncrypted   string     `gorm:"type:text;not null" json:"-"`
	Password            string     `gorm:"-" json:"-"` // Not stored in DB
	ConcurrentSessions  bool       `gorm:"default:true;not null" json:"concurrent_sessions"`
	SkipTLSVerify       bool       `gorm:"default:false" json:"skip_tls_verify"`
	IsActive            bool       `gorm:"default:true" json:"is_active"`
	LastTestAt          *time.Time `json:"last_test_at,omitempty"`
	LastTestSuccess     *bool      `json:"last_test_success,omitempty"`
	LastTestError       *string    `gorm:"type:text" json:"last_test_error,omitempty"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy           string     `gorm:"size:30" json:"created_by"`
	UpdatedBy           string     `gorm:"size:30" json:"updated_by"`
	
	// Sync configuration
	SyncEnabled         bool       `gorm:"default:true" json:"sync_enabled"`
	UserSyncInterval    *int       `gorm:"default:30" json:"user_sync_interval"`    // minutes
	GroupSyncInterval   *int       `gorm:"default:60" json:"group_sync_interval"`   // minutes
	SafeSyncInterval    *int       `gorm:"default:120" json:"safe_sync_interval"`   // minutes
	UserSyncPageSize    *int       `gorm:"default:100" json:"user_sync_page_size"`  // pagination size
	
	// Sync status tracking
	LastUserSyncAt      *time.Time `json:"last_user_sync_at,omitempty"`
	LastUserSyncStatus  *string    `gorm:"size:20" json:"last_user_sync_status,omitempty"`
	LastUserSyncError   *string    `gorm:"type:text" json:"last_user_sync_error,omitempty"`
	
	// Relationships
	Operations []Operation `gorm:"foreignKey:CyberArkInstanceID" json:"-"`
}

func (c *CyberArkInstance) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = ulid.New(ulid.CyberArkInstancePrefix)
	}
	
	// Set CreatedBy and UpdatedBy from context
	if userID, ok := tx.Statement.Context.Value("user_id").(string); ok {
		c.CreatedBy = userID
		c.UpdatedBy = userID
	}
	
	return nil
}

func (c *CyberArkInstance) BeforeUpdate(tx *gorm.DB) error {
	// Set UpdatedBy from context
	if userID, ok := tx.Statement.Context.Value("user_id").(string); ok {
		c.UpdatedBy = userID
	}
	return nil
}

func (CyberArkInstance) TableName() string {
	return "cyberark_instances"
}