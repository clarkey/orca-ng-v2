package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"encoding/json"
	"gorm.io/gorm"
)

type Operation struct {
	ID                  string         `gorm:"primaryKey;size:30" json:"id"`
	Type                string         `gorm:"size:50;not null" json:"type"`
	Priority            string         `gorm:"size:10;not null" json:"priority"` // low, normal, medium, high
	Status              string         `gorm:"size:20;not null" json:"status"`   // pending, processing, completed, failed, cancelled
	Payload             json.RawMessage `gorm:"type:json;not null" json:"payload"`
	Result              *json.RawMessage `gorm:"type:json" json:"result,omitempty"`
	ErrorMessage        *string        `gorm:"type:text" json:"error_message,omitempty"`
	RetryCount          int            `gorm:"default:0" json:"retry_count"`
	MaxRetries          int            `gorm:"default:3" json:"max_retries"`
	ScheduledAt         time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"scheduled_at"`
	StartedAt           *time.Time     `json:"started_at,omitempty"`
	CompletedAt         *time.Time     `json:"completed_at,omitempty"`
	CreatedBy           *string        `gorm:"size:30" json:"created_by,omitempty"`
	CyberArkInstanceID  *string        `gorm:"size:30" json:"cyberark_instance_id,omitempty"`
	CorrelationID       *string        `gorm:"size:30" json:"correlation_id,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	Creator          *User             `gorm:"foreignKey:CreatedBy" json:"-"`
	CyberArkInstance *CyberArkInstance `gorm:"foreignKey:CyberArkInstanceID" json:"-"`
}

func (o *Operation) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = ulid.New(ulid.OperationPrefix)
	}
	
	// Set CreatedBy from context
	if userID, ok := tx.Statement.Context.Value("user_id").(string); ok {
		o.CreatedBy = &userID
	}
	
	return nil
}

func (Operation) TableName() string {
	return "operations"
}

// Constants for operation types
const (
	OpTypeSafeProvision = "safe_provision"
	OpTypeUserSync      = "user_sync"
	OpTypeSafeSync      = "safe_sync"
)

// Constants for operation status
const (
	OpStatusPending    = "pending"
	OpStatusProcessing = "processing"
	OpStatusCompleted  = "completed"
	OpStatusFailed     = "failed"
	OpStatusCancelled  = "cancelled"
)

// Constants for operation priority
const (
	OpPriorityLow    = "low"
	OpPriorityNormal = "normal"
	OpPriorityMedium = "medium"
	OpPriorityHigh   = "high"
)