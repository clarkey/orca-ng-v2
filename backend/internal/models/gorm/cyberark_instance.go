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
	
	// Note: All sync configuration has been moved to the instance_sync_configs table
	// This provides per-sync-type configuration (users, safes, groups)
	
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