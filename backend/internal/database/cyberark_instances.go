package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/orca-ng/orca/internal/models"
)

// GetCyberArkInstances returns all CyberArk instances
func (db *DB) GetCyberArkInstances(ctx context.Context, onlyActive bool) ([]models.CyberArkInstance, error) {
	query := `
		SELECT id, name, base_url, username, password_encrypted, concurrent_sessions, is_active,
			   last_test_at, last_test_success, last_test_error,
			   created_at, updated_at, created_by, updated_by
		FROM cyberark_instances
		WHERE ($1 = false OR is_active = true)
		ORDER BY name ASC`

	rows, err := db.pool.Query(ctx, query, onlyActive)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	instances := make([]models.CyberArkInstance, 0)
	for rows.Next() {
		var instance models.CyberArkInstance
		err := rows.Scan(
			&instance.ID,
			&instance.Name,
			&instance.BaseURL,
			&instance.Username,
			&instance.PasswordEncrypted,
			&instance.ConcurrentSessions,
			&instance.IsActive,
			&instance.LastTestAt,
			&instance.LastTestSuccess,
			&instance.LastTestError,
			&instance.CreatedAt,
			&instance.UpdatedAt,
			&instance.CreatedBy,
			&instance.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance: %w", err)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// GetCyberArkInstance returns a single CyberArk instance by ID
func (db *DB) GetCyberArkInstance(ctx context.Context, id string) (*models.CyberArkInstance, error) {
	query := `
		SELECT id, name, base_url, username, password_encrypted, concurrent_sessions, is_active,
			   last_test_at, last_test_success, last_test_error,
			   created_at, updated_at, created_by, updated_by
		FROM cyberark_instances
		WHERE id = $1`

	var instance models.CyberArkInstance
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&instance.ID,
		&instance.Name,
		&instance.BaseURL,
		&instance.Username,
		&instance.PasswordEncrypted,
		&instance.ConcurrentSessions,
		&instance.IsActive,
		&instance.LastTestAt,
		&instance.LastTestSuccess,
		&instance.LastTestError,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&instance.CreatedBy,
		&instance.UpdatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("instance not found")
		}
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return &instance, nil
}

// CreateCyberArkInstance creates a new CyberArk instance
func (db *DB) CreateCyberArkInstance(ctx context.Context, instance *models.CyberArkInstance, userID string) error {
	// Generate ULID with prefix
	instance.ID = "cai_" + ulid.Make().String()

	query := `
		INSERT INTO cyberark_instances (id, name, base_url, username, password_encrypted, concurrent_sessions, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING created_at, updated_at`

	err := db.pool.QueryRow(ctx, query,
		instance.ID,
		instance.Name,
		instance.BaseURL,
		instance.Username,
		instance.PasswordEncrypted,
		instance.ConcurrentSessions,
		userID,
	).Scan(&instance.CreatedAt, &instance.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	instance.CreatedBy = sql.NullString{String: userID, Valid: true}
	instance.UpdatedBy = sql.NullString{String: userID, Valid: true}

	return nil
}

// UpdateCyberArkInstance updates an existing CyberArk instance
func (db *DB) UpdateCyberArkInstance(ctx context.Context, id string, updates map[string]interface{}, userID string) error {
	// Always update the updated_by field
	updates["updated_by"] = userID

	// Build the update query dynamically
	setClause := ""
	args := []interface{}{id}
	argIndex := 2

	for field, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE cyberark_instances
		SET %s
		WHERE id = $1`, setClause)

	result, err := db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found")
	}

	return nil
}

// UpdateCyberArkInstanceTestResult updates the test result for an instance
func (db *DB) UpdateCyberArkInstanceTestResult(ctx context.Context, id string, success bool, errorMsg string) error {
	query := `
		UPDATE cyberark_instances
		SET last_test_at = CURRENT_TIMESTAMP,
			last_test_success = $2,
			last_test_error = $3
		WHERE id = $1`

	var testError sql.NullString
	if errorMsg != "" {
		testError = sql.NullString{String: errorMsg, Valid: true}
	}

	result, err := db.pool.Exec(ctx, query, id, success, testError)
	if err != nil {
		return fmt.Errorf("failed to update test result: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found")
	}

	return nil
}

// DeleteCyberArkInstance deletes a CyberArk instance
func (db *DB) DeleteCyberArkInstance(ctx context.Context, id string) error {
	query := `DELETE FROM cyberark_instances WHERE id = $1`

	result, err := db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found")
	}

	return nil
}

// CheckCyberArkInstanceNameExists checks if an instance name already exists
func (db *DB) CheckCyberArkInstanceNameExists(ctx context.Context, name string, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM cyberark_instances 
			WHERE name = $1 AND ($2 = '' OR id != $2)
		)`

	var exists bool
	err := db.pool.QueryRow(ctx, query, name, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check name existence: %w", err)
	}

	return exists, nil
}