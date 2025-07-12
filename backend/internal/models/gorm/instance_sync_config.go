package gorm

import (
	"time"

	"gorm.io/gorm"
)

// InstanceSyncConfig represents sync configuration for a specific sync type on an instance
type InstanceSyncConfig struct {
	ID                 string     `gorm:"primaryKey;size:30" json:"id"`
	CyberArkInstanceID string     `gorm:"column:cyberark_instance_id;size:30;not null;uniqueIndex:idx_instance_sync_type" json:"cyberark_instance_id"`
	SyncType           string     `gorm:"size:50;not null;uniqueIndex:idx_instance_sync_type" json:"sync_type"` // users, safes, groups
	Enabled            bool       `gorm:"default:true" json:"enabled"`
	IntervalMinutes    int        `gorm:"not null;default:60" json:"interval_minutes"`
	PageSize           int        `gorm:"default:100" json:"page_size"`
	RetryAttempts      int        `gorm:"default:3" json:"retry_attempts"`
	TimeoutMinutes     int        `gorm:"default:30" json:"timeout_minutes"`
	LastRunAt          *time.Time `json:"last_run_at"`
	LastRunStatus      *string    `gorm:"size:20" json:"last_run_status"`
	LastRunMessage     *string    `gorm:"type:text" json:"last_run_message"`
	NextRunAt          *time.Time `json:"next_run_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relations
	CyberArkInstance *CyberArkInstance `gorm:"foreignKey:CyberArkInstanceID" json:"cyberark_instance,omitempty"`
}

func (InstanceSyncConfig) TableName() string {
	return "instance_sync_configs"
}

// CalculateNextRunAt calculates the next run time based on the interval
func (c *InstanceSyncConfig) CalculateNextRunAt() time.Time {
	if c.LastRunAt == nil {
		// If never run, schedule for now
		return time.Now()
	}
	return c.LastRunAt.Add(time.Duration(c.IntervalMinutes) * time.Minute)
}

// IsOverdue checks if the sync is overdue based on the next run time
func (c *InstanceSyncConfig) IsOverdue() bool {
	if c.NextRunAt == nil {
		return true
	}
	return time.Now().After(*c.NextRunAt)
}