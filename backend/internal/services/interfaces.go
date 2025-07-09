package services

import (
	"context"
	
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
)

// SessionService defines the interface for session operations
type SessionService interface {
	GetSessionByToken(ctx context.Context, token string) (*gormmodels.Session, *gormmodels.User, error)
}