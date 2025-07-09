package database

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
	
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/models"
	"github.com/orca-ng/orca/pkg/ulid"
)

// GetCyberArkInstances returns all CyberArk instances using GORM
func (db *GormDB) GetCyberArkInstances(ctx context.Context, onlyActive bool) ([]models.CyberArkInstance, error) {
	query := db.WithContext(ctx).Model(&gormmodels.CyberArkInstance{})
	
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	
	var gormInstances []gormmodels.CyberArkInstance
	if err := query.Order("name ASC").Find(&gormInstances).Error; err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	
	// Convert GORM models to regular models
	instances := make([]models.CyberArkInstance, len(gormInstances))
	for i, gi := range gormInstances {
		instances[i] = convertToInstance(&gi)
	}
	
	return instances, nil
}

// GetCyberArkInstance returns a single CyberArk instance by ID using GORM
func (db *GormDB) GetCyberArkInstance(ctx context.Context, id string) (*models.CyberArkInstance, error) {
	var gormInstance gormmodels.CyberArkInstance
	
	if err := db.WithContext(ctx).First(&gormInstance, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("instance not found")
		}
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	
	instance := convertToInstance(&gormInstance)
	return &instance, nil
}

// CreateCyberArkInstance creates a new CyberArk instance using GORM
func (db *GormDB) CreateCyberArkInstance(ctx context.Context, instance *models.CyberArkInstance, userID string) error {
	// Convert to GORM model
	gormInstance := &gormmodels.CyberArkInstance{
		ID:                 ulid.New(ulid.CyberArkInstancePrefix),
		Name:               instance.Name,
		BaseURL:            instance.BaseURL,
		Username:           instance.Username,
		PasswordEncrypted:  instance.PasswordEncrypted,
		ConcurrentSessions: instance.ConcurrentSessions,
		SkipTLSVerify:      instance.SkipTLSVerify,
		IsActive:           instance.IsActive,
	}
	
	// Create with user context
	createCtx := context.WithValue(ctx, "user_id", userID)
	if err := db.WithContext(createCtx).Create(gormInstance).Error; err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}
	
	// Update the original instance with generated values
	*instance = convertToInstance(gormInstance)
	
	return nil
}

// UpdateCyberArkInstance updates an existing CyberArk instance using GORM
func (db *GormDB) UpdateCyberArkInstance(ctx context.Context, id string, updates map[string]interface{}, userID string) error {
	// Add updated_by to updates
	updates["updated_by"] = userID
	
	// Update with user context for audit
	updateCtx := context.WithValue(ctx, "user_id", userID)
	result := db.WithContext(updateCtx).
		Model(&gormmodels.CyberArkInstance{}).
		Where("id = ?", id).
		Updates(updates)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update instance: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found")
	}
	
	return nil
}

// UpdateCyberArkInstanceTestResult updates the test result for an instance using GORM
func (db *GormDB) UpdateCyberArkInstanceTestResult(ctx context.Context, id string, success bool, errorMsg string) error {
	updates := map[string]interface{}{
		"last_test_at":      gorm.Expr("CURRENT_TIMESTAMP"),
		"last_test_success": success,
	}
	
	if errorMsg != "" {
		updates["last_test_error"] = errorMsg
	} else {
		updates["last_test_error"] = nil
	}
	
	result := db.WithContext(ctx).
		Model(&gormmodels.CyberArkInstance{}).
		Where("id = ?", id).
		Updates(updates)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update test result: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found")
	}
	
	return nil
}

// DeleteCyberArkInstance deletes a CyberArk instance using GORM
func (db *GormDB) DeleteCyberArkInstance(ctx context.Context, id string) error {
	result := db.WithContext(ctx).Delete(&gormmodels.CyberArkInstance{}, "id = ?", id)
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete instance: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found")
	}
	
	return nil
}

// CheckCyberArkInstanceNameExists checks if an instance name already exists using GORM
func (db *GormDB) CheckCyberArkInstanceNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := db.WithContext(ctx).Model(&gormmodels.CyberArkInstance{}).
		Where("name = ?", name)
	
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check name existence: %w", err)
	}
	
	return count > 0, nil
}

// convertToInstance converts GORM model to regular model
func convertToInstance(gi *gormmodels.CyberArkInstance) models.CyberArkInstance {
	instance := models.CyberArkInstance{
		ID:                 gi.ID,
		Name:               gi.Name,
		BaseURL:            gi.BaseURL,
		Username:           gi.Username,
		PasswordEncrypted:  gi.PasswordEncrypted,
		ConcurrentSessions: gi.ConcurrentSessions,
		SkipTLSVerify:      gi.SkipTLSVerify,
		IsActive:           gi.IsActive,
		LastTestAt:         gi.LastTestAt,
		CreatedAt:          gi.CreatedAt,
		UpdatedAt:          gi.UpdatedAt,
	}
	
	// Convert pointer types to sql.Null types
	if gi.LastTestSuccess != nil {
		instance.LastTestSuccess = sql.NullBool{Bool: *gi.LastTestSuccess, Valid: true}
	}
	
	if gi.LastTestError != nil {
		instance.LastTestError = sql.NullString{String: *gi.LastTestError, Valid: true}
	}
	
	// Convert string to sql.NullString for CreatedBy and UpdatedBy
	if gi.CreatedBy != "" {
		instance.CreatedBy = sql.NullString{String: gi.CreatedBy, Valid: true}
	}
	
	if gi.UpdatedBy != "" {
		instance.UpdatedBy = sql.NullString{String: gi.UpdatedBy, Valid: true}
	}
	
	return instance
}