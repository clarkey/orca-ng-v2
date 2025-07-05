package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/orca-ng/orca/pkg/ulid"
)

// Store handles database operations for the pipeline
type Store struct {
	db *sql.DB
}

// NewStore creates a new pipeline store
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateOperation creates a new operation in the database
func (s *Store) CreateOperation(ctx context.Context, req *CreateOperationRequest, createdBy *string) (*Operation, error) {
	op := &Operation{
		ID:         ulid.New(ulid.OperationPrefix),
		Type:       req.Type,
		Priority:   req.Priority,
		Status:     StatusPending,
		Payload:    req.Payload,
		RetryCount: 0,
		MaxRetries: 3, // Default, can be overridden
		CreatedBy:  createdBy,
	}
	
	// Set scheduled time
	if req.ScheduledAt != nil {
		op.ScheduledAt = *req.ScheduledAt
	} else {
		op.ScheduledAt = time.Now()
	}
	
	// Set correlation ID if provided
	op.CorrelationID = req.CorrelationID
	
	// Default priority to normal if not specified
	if op.Priority == "" {
		op.Priority = PriorityNormal
	}
	
	// Insert into database
	query := `
		INSERT INTO operations (
			id, type, priority, status, payload, retry_count, max_retries,
			scheduled_at, created_by, correlation_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING created_at, updated_at`
	
	err := s.db.QueryRowContext(
		ctx, query,
		op.ID, op.Type, op.Priority, op.Status, op.Payload,
		op.RetryCount, op.MaxRetries, op.ScheduledAt,
		op.CreatedBy, op.CorrelationID, time.Now(), time.Now(),
	).Scan(&op.CreatedAt, &op.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("create operation: %w", err)
	}
	
	return op, nil
}

// GetOperation retrieves an operation by ID
func (s *Store) GetOperation(ctx context.Context, id string) (*Operation, error) {
	var op Operation
	
	query := `
		SELECT id, type, priority, status, payload, result, error_message,
		       retry_count, max_retries, scheduled_at, started_at, completed_at,
		       created_by, cyberark_instance_id, correlation_id, created_at, updated_at
		FROM operations
		WHERE id = $1`
	
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&op.ID, &op.Type, &op.Priority, &op.Status, &op.Payload, &op.Result,
		&op.ErrorMessage, &op.RetryCount, &op.MaxRetries, &op.ScheduledAt,
		&op.StartedAt, &op.CompletedAt, &op.CreatedBy, &op.CyberArkInstanceID,
		&op.CorrelationID, &op.CreatedAt, &op.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("operation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get operation: %w", err)
	}
	
	return &op, nil
}

// ListOperations retrieves operations with filtering
func (s *Store) ListOperations(ctx context.Context, filters ListOperationsFilters) ([]*Operation, error) {
	query := `
		SELECT id, type, priority, status, payload, result, error_message,
		       retry_count, max_retries, scheduled_at, started_at, completed_at,
		       created_by, cyberark_instance_id, correlation_id, created_at, updated_at
		FROM operations
		WHERE 1=1`
	
	args := []interface{}{}
	argCount := 0
	
	// Build dynamic query based on filters
	if filters.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
	}
	
	if filters.Type != nil {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filters.Type)
	}
	
	if filters.Priority != nil {
		argCount++
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, *filters.Priority)
	}
	
	if filters.CreatedBy != nil {
		argCount++
		query += fmt.Sprintf(" AND created_by = $%d", argCount)
		args = append(args, *filters.CreatedBy)
	}
	
	if filters.CorrelationID != nil {
		argCount++
		query += fmt.Sprintf(" AND correlation_id = $%d", argCount)
		args = append(args, *filters.CorrelationID)
	}
	
	if filters.CreatedAfter != nil {
		argCount++
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.CreatedAfter)
	}
	
	if filters.CreatedBefore != nil {
		argCount++
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.CreatedBefore)
	}
	
	// Add ordering
	query += " ORDER BY created_at DESC"
	
	// Add limit
	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}
	
	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list operations: %w", err)
	}
	defer rows.Close()
	
	operations := []*Operation{}
	for rows.Next() {
		var op Operation
		err := rows.Scan(
			&op.ID, &op.Type, &op.Priority, &op.Status, &op.Payload, &op.Result,
			&op.ErrorMessage, &op.RetryCount, &op.MaxRetries, &op.ScheduledAt,
			&op.StartedAt, &op.CompletedAt, &op.CreatedBy, &op.CyberArkInstanceID,
			&op.CorrelationID, &op.CreatedAt, &op.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan operation: %w", err)
		}
		operations = append(operations, &op)
	}
	
	return operations, nil
}

// WaitForOperation waits for an operation to complete or timeout
func (s *Store) WaitForOperation(ctx context.Context, id string, timeout time.Duration) (*Operation, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			op, err := s.GetOperation(ctx, id)
			if err != nil {
				return nil, err
			}
			
			// Check if operation is in terminal state
			if op.Status == StatusCompleted || op.Status == StatusFailed || op.Status == StatusCancelled {
				return op, nil
			}
			
			// Check timeout
			if time.Now().After(deadline) {
				return op, fmt.Errorf("timeout waiting for operation")
			}
		}
	}
}

// CancelOperation cancels a pending or processing operation
func (s *Store) CancelOperation(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE operations 
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status IN ($4, $5)`,
		StatusCancelled, time.Now(), id, StatusPending, StatusProcessing,
	)
	
	if err != nil {
		return fmt.Errorf("cancel operation: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("operation cannot be cancelled (not found or already completed)")
	}
	
	return nil
}

// GetPipelineConfig retrieves pipeline configuration from database
func (s *Store) GetPipelineConfig(ctx context.Context) (*PipelineConfig, error) {
	config := &PipelineConfig{
		PriorityAllocation: make(map[Priority]float64),
		OperationTimeouts:  make(map[OperationType]int),
	}
	
	// Query all config values
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM pipeline_config`)
	if err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var key string
		var value json.RawMessage
		
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		
		switch key {
		case "processing_capacity":
			var capacityConfig struct {
				Total              int                  `json:"total"`
				PriorityAllocation map[Priority]float64 `json:"priority_allocation"`
			}
			if err := json.Unmarshal(value, &capacityConfig); err != nil {
				return nil, fmt.Errorf("unmarshal capacity config: %w", err)
			}
			config.TotalCapacity = capacityConfig.Total
			config.PriorityAllocation = capacityConfig.PriorityAllocation
			
		case "retry_policy":
			if err := json.Unmarshal(value, &config.RetryPolicy); err != nil {
				return nil, fmt.Errorf("unmarshal retry policy: %w", err)
			}
			
		case "operation_timeouts":
			var timeouts struct {
				Default           int                       `json:"default"`
				OperationTimeouts map[OperationType]int     `json:"operation_timeouts"`
			}
			if err := json.Unmarshal(value, &timeouts); err != nil {
				// Try unmarshaling directly as map
				if err := json.Unmarshal(value, &config.OperationTimeouts); err != nil {
					return nil, fmt.Errorf("unmarshal timeouts: %w", err)
				}
				config.DefaultTimeout = 300 // Default 5 minutes
			} else {
				config.DefaultTimeout = timeouts.Default
				// Copy specific timeouts
				for k, v := range timeouts.OperationTimeouts {
					config.OperationTimeouts[k] = v
				}
			}
		}
	}
	
	// Set defaults if not configured
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 300
	}
	
	return config, nil
}

// UpdatePipelineConfig updates pipeline configuration
func (s *Store) UpdatePipelineConfig(ctx context.Context, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal config value: %w", err)
	}
	
	_, err = s.db.ExecContext(ctx, `
		UPDATE pipeline_config 
		SET value = $1, updated_at = $2
		WHERE key = $3`,
		jsonValue, time.Now(), key,
	)
	
	if err != nil {
		return fmt.Errorf("update config: %w", err)
	}
	
	return nil
}

// ListOperationsFilters defines filters for listing operations
type ListOperationsFilters struct {
	Status        *Status
	Type          *OperationType
	Priority      *Priority
	CreatedBy     *string
	CorrelationID *string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Limit         int
}