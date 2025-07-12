package gorm

import (
	"time"

	"gorm.io/gorm"
)

// SyncJob represents a sync job execution record
type SyncJob struct {
	ID                 string     `gorm:"primaryKey;size:30" json:"id"`
	CyberArkInstanceID string     `gorm:"column:cyberark_instance_id;size:30;not null;index" json:"cyberark_instance_id"`
	SyncType           string     `gorm:"size:50;not null;index" json:"sync_type"` // users, safes, groups
	Status             string     `gorm:"size:20;not null;index" json:"status"`    // pending, running, completed, failed
	TriggeredBy        string     `gorm:"size:50;not null" json:"triggered_by"`    // manual, scheduled
	StartedAt          *time.Time `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	NextRunAt          *time.Time `json:"next_run_at"`
	RecordsSynced      int        `json:"records_synced"`
	RecordsCreated     int        `json:"records_created"`
	RecordsUpdated     int        `json:"records_updated"`
	RecordsDeleted     int        `json:"records_deleted"`
	RecordsFailed      int        `json:"records_failed"`
	ErrorMessage       *string    `json:"error_message"`
	ErrorDetails       *string    `gorm:"type:text" json:"error_details"`
	DurationSeconds    *float64   `json:"duration_seconds"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relations
	CyberArkInstance *CyberArkInstance `gorm:"foreignKey:CyberArkInstanceID" json:"cyberark_instance,omitempty"`
	CreatedByUser    *User             `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
	CreatedBy        *string           `gorm:"size:30" json:"created_by,omitempty"`
}

// SyncJobStatus constants
const (
	SyncJobStatusPending   = "pending"
	SyncJobStatusRunning   = "running"
	SyncJobStatusCompleted = "completed"
	SyncJobStatusFailed    = "failed"
	SyncJobStatusCancelled = "cancelled"
)

// SyncType constants
const (
	SyncTypeUsers  = "users"
	SyncTypeSafes  = "safes"
	SyncTypeGroups = "groups"
)

// TriggeredBy constants
const (
	TriggeredByManual    = "manual"
	TriggeredByScheduled = "scheduled"
	TriggeredBySystem    = "system"
)

func (SyncJob) TableName() string {
	return "sync_jobs"
}