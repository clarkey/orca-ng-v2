package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/database"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/pkg/ulid"
)

// Store handles database operations for the pipeline using GORM
type Store struct {
	db *database.GormDB
}

// NewStore creates a new pipeline store using GORM
func NewStore(db *database.GormDB) *Store {
	return &Store{db: db}
}

// CreateOperation creates a new operation in the database
func (s *Store) CreateOperation(ctx context.Context, req *CreateOperationRequest, createdBy *string) (*Operation, error) {
	gormOp := &gormmodels.Operation{
		ID:         ulid.New(ulid.OperationPrefix),
		Type:       string(req.Type),
		Priority:   string(req.Priority),
		Status:     gormmodels.OpStatusPending,
		Payload:    req.Payload,
		RetryCount: 0,
		MaxRetries: 3, // Default, can be overridden
		CreatedBy:  createdBy,
	}
	
	// Set scheduled time
	if req.ScheduledAt != nil {
		gormOp.ScheduledAt = *req.ScheduledAt
	} else {
		gormOp.ScheduledAt = time.Now()
	}
	
	// Set correlation ID if provided
	gormOp.CorrelationID = req.CorrelationID
	
	// Default priority to normal if not specified
	if gormOp.Priority == "" {
		gormOp.Priority = gormmodels.OpPriorityNormal
	}
	
	// Create with user context if available
	createCtx := ctx
	if createdBy != nil {
		createCtx = context.WithValue(ctx, "user_id", *createdBy)
	}
	
	// Insert into database
	if err := s.db.WithContext(createCtx).Create(gormOp).Error; err != nil {
		return nil, fmt.Errorf("create operation: %w", err)
	}
	
	// Convert to pipeline Operation
	return s.convertToOperation(gormOp), nil
}

// GetOperation retrieves an operation by ID
func (s *Store) GetOperation(ctx context.Context, id string) (*Operation, error) {
	var gormOp gormmodels.Operation
	
	result := s.db.WithContext(ctx).First(&gormOp, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("operation not found")
		}
		return nil, fmt.Errorf("get operation: %w", result.Error)
	}
	
	return s.convertToOperation(&gormOp), nil
}

// ListOperations retrieves operations with filtering
func (s *Store) ListOperations(ctx context.Context, filters ListOperationsFilters) ([]*Operation, error) {
	query := s.db.WithContext(ctx).Model(&gormmodels.Operation{})
	
	// Apply filters
	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}
	
	if filters.Type != nil {
		query = query.Where("type = ?", string(*filters.Type))
	}
	
	if filters.Priority != nil {
		query = query.Where("priority = ?", string(*filters.Priority))
	}
	
	if filters.CreatedBy != nil {
		query = query.Where("created_by = ?", *filters.CreatedBy)
	}
	
	if filters.CorrelationID != nil {
		query = query.Where("correlation_id = ?", *filters.CorrelationID)
	}
	
	// Search filter - uses full-text search with fallback to fuzzy matching
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where(
			"search_vector @@ plainto_tsquery('english', ?) OR id ILIKE ? OR type::text ILIKE ?",
			*filters.Search, searchPattern, searchPattern,
		)
	}
	
	// Date range filters
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}
	
	// Legacy date filters
	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}
	
	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}
	
	// Add ordering
	if filters.SortBy != nil && filters.SortOrder != nil {
		// Validate sort column to prevent SQL injection
		validColumns := map[string]bool{
			"id":         true,
			"type":       true,
			"priority":   true,
			"status":     true,
			"created_at": true,
		}
		
		if validColumns[*filters.SortBy] {
			order := "ASC"
			if *filters.SortOrder == "desc" {
				order = "DESC"
			}
			query = query.Order(fmt.Sprintf("%s %s", *filters.SortBy, order))
		} else {
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}
	
	// Add limit and offset for pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	
	// Execute query
	var gormOps []gormmodels.Operation
	if err := query.Find(&gormOps).Error; err != nil {
		return nil, fmt.Errorf("list operations: %w", err)
	}
	
	// Convert to pipeline operations
	operations := make([]*Operation, len(gormOps))
	for i, gormOp := range gormOps {
		operations[i] = s.convertToOperation(&gormOp)
	}
	
	return operations, nil
}

// CountOperations counts operations matching the given filters
func (s *Store) CountOperations(ctx context.Context, filters ListOperationsFilters) (int, error) {
	query := s.db.WithContext(ctx).Model(&gormmodels.Operation{})
	
	// Apply same filters as ListOperations (without LIMIT/OFFSET)
	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
	}
	
	if filters.Type != nil {
		query = query.Where("type = ?", string(*filters.Type))
	}
	
	if filters.Priority != nil {
		query = query.Where("priority = ?", string(*filters.Priority))
	}
	
	if filters.CreatedBy != nil {
		query = query.Where("created_by = ?", *filters.CreatedBy)
	}
	
	if filters.CorrelationID != nil {
		query = query.Where("correlation_id = ?", *filters.CorrelationID)
	}
	
	// Search filter
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where(
			"search_vector @@ plainto_tsquery('english', ?) OR id ILIKE ? OR type::text ILIKE ?",
			*filters.Search, searchPattern, searchPattern,
		)
	}
	
	// Date range filters
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}
	
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count operations: %w", err)
	}
	
	return int(count), nil
}

// GetOperationStats returns aggregated statistics for operations
func (s *Store) GetOperationStats(ctx context.Context, startDate, endDate *time.Time) (*OperationStats, error) {
	stats := &OperationStats{
		ByStatus:   make(map[Status]int),
		ByType:     make(map[OperationType]int),
		ByPriority: make(map[Priority]int),
		ByHour:     []HourlyStats{},
	}
	
	// Base query
	baseQuery := s.db.WithContext(ctx).Model(&gormmodels.Operation{})
	
	if startDate != nil {
		baseQuery = baseQuery.Where("created_at >= ?", *startDate)
	}
	
	if endDate != nil {
		baseQuery = baseQuery.Where("created_at <= ?", *endDate)
	}
	
	// Get counts by status
	type statusCount struct {
		Status string
		Count  int
	}
	var statusCounts []statusCount
	
	if err := baseQuery.
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("query status stats: %w", err)
	}
	
	for _, sc := range statusCounts {
		status := Status(sc.Status)
		stats.ByStatus[status] = sc.Count
		stats.TotalCount += sc.Count
	}
	
	// Get counts by type
	type typeCount struct {
		Type  string
		Count int
	}
	var typeCounts []typeCount
	
	if err := baseQuery.
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, fmt.Errorf("query type stats: %w", err)
	}
	
	for _, tc := range typeCounts {
		opType := OperationType(tc.Type)
		stats.ByType[opType] = tc.Count
	}
	
	// Get counts by priority
	type priorityCount struct {
		Priority string
		Count    int
	}
	var priorityCounts []priorityCount
	
	if err := baseQuery.
		Select("priority, COUNT(*) as count").
		Group("priority").
		Scan(&priorityCounts).Error; err != nil {
		return nil, fmt.Errorf("query priority stats: %w", err)
	}
	
	for _, pc := range priorityCounts {
		priority := Priority(pc.Priority)
		stats.ByPriority[priority] = pc.Count
	}
	
	// Get hourly distribution for the last 24 hours
	type hourlyCount struct {
		Hour  time.Time
		Count int
	}
	var hourlyCounts []hourlyCount
	
	if err := baseQuery.
		Select("date_trunc('hour', created_at) as hour, COUNT(*) as count").
		Group("hour").
		Order("hour DESC").
		Limit(24).
		Scan(&hourlyCounts).Error; err != nil {
		return nil, fmt.Errorf("query hourly stats: %w", err)
	}
	
	for _, hc := range hourlyCounts {
		stats.ByHour = append(stats.ByHour, HourlyStats{
			Hour:  hc.Hour,
			Count: hc.Count,
		})
	}
	
	// Get average times
	type avgTimes struct {
		AvgWaitTime    *float64
		AvgProcessTime *float64
	}
	var times avgTimes
	
	if err := baseQuery.
		Select(`
			AVG(EXTRACT(EPOCH FROM (COALESCE(started_at, NOW()) - created_at))) as avg_wait_time,
			AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_process_time
		`).
		Where("started_at IS NOT NULL").
		Scan(&times).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query avg times: %w", err)
	}
	
	if times.AvgWaitTime != nil {
		stats.AvgWaitTime = *times.AvgWaitTime
	}
	if times.AvgProcessTime != nil {
		stats.AvgProcessTime = *times.AvgProcessTime
	}
	
	return stats, nil
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
	result := s.db.WithContext(ctx).
		Model(&gormmodels.Operation{}).
		Where("id = ? AND status IN (?, ?)", id, gormmodels.OpStatusPending, gormmodels.OpStatusProcessing).
		Update("status", gormmodels.OpStatusCancelled)
	
	if result.Error != nil {
		return fmt.Errorf("cancel operation: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
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
	
	// Define a model for pipeline_config table
	type pipelineConfigRow struct {
		Key   string          `gorm:"primaryKey"`
		Value json.RawMessage `gorm:"type:json"`
	}
	
	// Query all config values
	var rows []pipelineConfigRow
	if err := s.db.WithContext(ctx).
		Table("pipeline_config").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}
	
	for _, row := range rows {
		switch row.Key {
		case "processing_capacity":
			var capacityConfig struct {
				Total              int                  `json:"total"`
				PriorityAllocation map[Priority]float64 `json:"priority_allocation"`
			}
			if err := json.Unmarshal(row.Value, &capacityConfig); err != nil {
				return nil, fmt.Errorf("unmarshal capacity config: %w", err)
			}
			config.TotalCapacity = capacityConfig.Total
			config.PriorityAllocation = capacityConfig.PriorityAllocation
			
		case "retry_policy":
			if err := json.Unmarshal(row.Value, &config.RetryPolicy); err != nil {
				return nil, fmt.Errorf("unmarshal retry policy: %w", err)
			}
			
		case "operation_timeouts":
			var timeouts struct {
				Default           int                       `json:"default"`
				OperationTimeouts map[OperationType]int     `json:"operation_timeouts"`
			}
			if err := json.Unmarshal(row.Value, &timeouts); err != nil {
				// Try unmarshaling directly as map
				if err := json.Unmarshal(row.Value, &config.OperationTimeouts); err != nil {
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
	
	// Use map for updates to avoid struct tags
	updates := map[string]interface{}{
		"value":      jsonValue,
		"updated_at": time.Now(),
	}
	
	result := s.db.WithContext(ctx).
		Table("pipeline_config").
		Where("key = ?", key).
		Updates(updates)
	
	if result.Error != nil {
		return fmt.Errorf("update config: %w", result.Error)
	}
	
	return nil
}

// convertToOperation converts GORM operation to pipeline Operation
func (s *Store) convertToOperation(gormOp *gormmodels.Operation) *Operation {
	return &Operation{
		ID:                 gormOp.ID,
		Type:               OperationType(gormOp.Type),
		Priority:           Priority(gormOp.Priority),
		Status:             Status(gormOp.Status),
		Payload:            gormOp.Payload,
		Result:             gormOp.Result,
		ErrorMessage:       gormOp.ErrorMessage,
		RetryCount:         gormOp.RetryCount,
		MaxRetries:         gormOp.MaxRetries,
		ScheduledAt:        gormOp.ScheduledAt,
		StartedAt:          gormOp.StartedAt,
		CompletedAt:        gormOp.CompletedAt,
		CreatedBy:          gormOp.CreatedBy,
		CyberArkInstanceID: gormOp.CyberArkInstanceID,
		CorrelationID:      gormOp.CorrelationID,
		CreatedAt:          gormOp.CreatedAt,
		UpdatedAt:          gormOp.UpdatedAt,
	}
}