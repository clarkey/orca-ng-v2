package gorm

import (
	"time"
	
	"github.com/orca-ng/orca/pkg/ulid"
	"encoding/json"
	"gorm.io/gorm"
)

type PipelineConfig struct {
	ID          string         `gorm:"primaryKey;size:30" json:"id"`
	Key         string         `gorm:"size:100;not null;uniqueIndex" json:"key"`
	Value       json.RawMessage `gorm:"type:json;not null" json:"value"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (p *PipelineConfig) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = ulid.New(ulid.ConfigPrefix)
	}
	return nil
}

func (PipelineConfig) TableName() string {
	return "pipeline_config"
}