package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"gorm.io/gorm"
)

type User struct {
	ID           string     `gorm:"primaryKey;size:30" json:"id"`
	Username     string     `gorm:"size:255;not null;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsAdmin      bool       `gorm:"default:false" json:"is_admin"`
	
	// Relationships
	Sessions            []Session            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedOperations   []Operation          `gorm:"foreignKey:CreatedBy" json:"-"`
	CreatedInstances    []CyberArkInstance   `gorm:"foreignKey:CreatedBy" json:"-"`
	UpdatedInstances    []CyberArkInstance   `gorm:"foreignKey:UpdatedBy" json:"-"`
	CreatedAuthorities  []CertificateAuthority `gorm:"foreignKey:CreatedBy" json:"-"`
	UpdatedAuthorities  []CertificateAuthority `gorm:"foreignKey:UpdatedBy" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = ulid.New(ulid.UserPrefix)
	}
	return nil
}

func (User) TableName() string {
	return "users"
}

type Session struct {
	ID        string    `gorm:"primaryKey;size:30" json:"id"`
	UserID    string    `gorm:"size:30;not null;index" json:"user_id"`
	Token     string    `gorm:"size:255;not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UserAgent *string   `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress *string   `gorm:"size:45" json:"ip_address,omitempty"` // Size 45 for IPv6
	
	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = ulid.New(ulid.SessionPrefix)
	}
	return nil
}

func (Session) TableName() string {
	return "sessions"
}