package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Processor manages the processing pipeline with priority lanes
type Processor struct {
	db              *sql.DB
	config          *PipelineConfig
	handlers        map[OperationType]OperationHandler
	logger          *logrus.Logger
	
	// Worker management
	workers         []*worker
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	
	// Metrics
	activeWorkers   int32
	processedOps    map[OperationType]*int64
	metricsLock     sync.RWMutex
}

// worker represents a processing unit assigned to a priority lane
type worker struct {
	id       int
	priority Priority
	proc     *Processor
}

// NewProcessor creates a new pipeline processor
func NewProcessor(db *sql.DB, config *PipelineConfig, logger *logrus.Logger) *Processor {
	processedOps := make(map[OperationType]*int64)
	for _, opType := range []OperationType{
		OpTypeSafeProvision, OpTypeSafeModify, OpTypeSafeDelete,
		OpTypeAccessGrant, OpTypeAccessRevoke,
		OpTypeUserSync, OpTypeSafeSync, OpTypeGroupSync,
	} {
		var count int64
		processedOps[opType] = &count
	}
	
	return &Processor{
		db:           db,
		config:       config,
		handlers:     make(map[OperationType]OperationHandler),
		logger:       logger,
		processedOps: processedOps,
	}
}

// RegisterHandler registers a handler for a specific operation type
func (p *Processor) RegisterHandler(opType OperationType, handler OperationHandler) {
	p.handlers[opType] = handler
}

// Start begins processing operations
func (p *Processor) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)
	
	// Calculate worker allocation based on priority percentages
	workerAllocation := p.calculateWorkerAllocation()
	
	p.logger.WithFields(logrus.Fields{
		"total_capacity":    p.config.TotalCapacity,
		"worker_allocation": workerAllocation,
	}).Info("Starting processing pipeline")
	
	// Create workers for each priority lane
	workerID := 0
	for priority, count := range workerAllocation {
		for i := 0; i < count; i++ {
			w := &worker{
				id:       workerID,
				priority: priority,
				proc:     p,
			}
			p.workers = append(p.workers, w)
			
			p.wg.Add(1)
			go w.run(p.ctx)
			
			workerID++
		}
	}
	
	// Start metrics collector
	p.wg.Add(1)
	go p.collectMetrics(p.ctx)
	
	return nil
}

// Stop gracefully shuts down the processor
func (p *Processor) Stop() error {
	p.logger.Info("Stopping processing pipeline")
	
	// Signal all workers to stop
	p.cancel()
	
	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		p.logger.Info("Processing pipeline stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for workers to stop")
	}
}

// calculateWorkerAllocation determines how many workers per priority
func (p *Processor) calculateWorkerAllocation() map[Priority]int {
	allocation := make(map[Priority]int)
	
	for priority, percentage := range p.config.PriorityAllocation {
		count := int(math.Round(float64(p.config.TotalCapacity) * percentage))
		if count < 1 && percentage > 0 {
			count = 1 // Ensure at least 1 worker if percentage > 0
		}
		allocation[priority] = count
	}
	
	// Adjust for rounding errors
	total := 0
	for _, count := range allocation {
		total += count
	}
	
	if total < p.config.TotalCapacity && len(allocation) > 0 {
		// Add remaining capacity to highest priority
		allocation[PriorityHigh] += p.config.TotalCapacity - total
	}
	
	return allocation
}

// run is the main worker loop
func (w *worker) run(ctx context.Context) {
	defer w.proc.wg.Done()
	
	w.proc.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"priority":  w.priority,
	}).Debug("Worker started")
	
	for {
		select {
		case <-ctx.Done():
			w.proc.logger.WithField("worker_id", w.id).Debug("Worker stopped")
			return
		default:
			// Try to fetch and process an operation
			if err := w.processNext(ctx); err != nil {
				if err != sql.ErrNoRows {
					w.proc.logger.WithError(err).Error("Error processing operation")
				}
				// Back off a bit if no work or error
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// processNext fetches and processes the next operation for this worker's priority
func (w *worker) processNext(ctx context.Context) error {
	// Start a transaction to fetch and lock an operation
	tx, err := w.proc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Fetch next pending operation for this priority
	var op Operation
	query := `
		SELECT id, type, priority, status, payload, result, error_message,
		       retry_count, max_retries, scheduled_at, started_at, completed_at,
		       created_by, cyberark_instance_id, correlation_id, created_at, updated_at
		FROM operations
		WHERE status = 'pending' 
		  AND priority = $1
		  AND scheduled_at <= NOW()
		ORDER BY scheduled_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`
	
	err = tx.QueryRowContext(ctx, query, w.priority).Scan(
		&op.ID, &op.Type, &op.Priority, &op.Status, &op.Payload, &op.Result,
		&op.ErrorMessage, &op.RetryCount, &op.MaxRetries, &op.ScheduledAt,
		&op.StartedAt, &op.CompletedAt, &op.CreatedBy, &op.CyberArkInstanceID,
		&op.CorrelationID, &op.CreatedAt, &op.UpdatedAt,
	)
	if err != nil {
		return err
	}
	
	// Mark as processing
	now := time.Now()
	op.Status = StatusProcessing
	op.StartedAt = &now
	
	_, err = tx.ExecContext(ctx, `
		UPDATE operations 
		SET status = $1, started_at = $2, updated_at = $3
		WHERE id = $4
	`, op.Status, op.StartedAt, now, op.ID)
	if err != nil {
		return fmt.Errorf("update operation status: %w", err)
	}
	
	// Commit transaction to release the lock
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	
	// Increment active workers count
	atomic.AddInt32(&w.proc.activeWorkers, 1)
	defer atomic.AddInt32(&w.proc.activeWorkers, -1)
	
	// Process the operation
	w.proc.logger.WithFields(logrus.Fields{
		"operation_id": op.ID,
		"type":        op.Type,
		"priority":    op.Priority,
		"worker_id":   w.id,
	}).Info("Processing operation")
	
	// Get handler
	handler, exists := w.proc.handlers[op.Type]
	if !exists {
		w.proc.completeOperation(&op, nil, fmt.Errorf("no handler registered for operation type: %s", op.Type))
		return nil
	}
	
	// Create timeout context
	timeout := w.proc.config.DefaultTimeout
	if t, ok := w.proc.config.OperationTimeouts[op.Type]; ok {
		timeout = t
	}
	
	opCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	
	// Execute handler
	err = handler.Handle(opCtx, &op)
	
	if err != nil {
		// Check if retryable
		if handler.CanRetry(err) && op.RetryCount < op.MaxRetries {
			w.proc.retryOperation(&op, err)
		} else {
			w.proc.completeOperation(&op, nil, err)
		}
	} else {
		// Success - dereference Result if it's not nil
		var result json.RawMessage
		if op.Result != nil {
			result = *op.Result
		}
		w.proc.completeOperation(&op, result, nil)
		
		// Update metrics
		atomic.AddInt64(w.proc.processedOps[op.Type], 1)
	}
	
	return nil
}

// completeOperation marks an operation as completed or failed
func (p *Processor) completeOperation(op *Operation, result json.RawMessage, err error) {
	now := time.Now()
	op.CompletedAt = &now
	
	if err != nil {
		op.Status = StatusFailed
		errMsg := err.Error()
		op.ErrorMessage = &errMsg
		
		p.logger.WithFields(logrus.Fields{
			"operation_id": op.ID,
			"error":        err,
		}).Error("Operation failed")
	} else {
		op.Status = StatusCompleted
		if result != nil {
			op.Result = &result
		}
		
		p.logger.WithFields(logrus.Fields{
			"operation_id": op.ID,
			"duration":     op.CompletedAt.Sub(*op.StartedAt).Seconds(),
		}).Info("Operation completed")
	}
	
	// Update in database
	_, dbErr := p.db.Exec(`
		UPDATE operations 
		SET status = $1, result = $2, error_message = $3, completed_at = $4, updated_at = $5
		WHERE id = $6
	`, op.Status, op.Result, op.ErrorMessage, op.CompletedAt, now, op.ID)
	
	if dbErr != nil {
		p.logger.WithError(dbErr).Error("Failed to update operation status")
	}
}

// retryOperation schedules an operation for retry
func (p *Processor) retryOperation(op *Operation, err error) {
	op.RetryCount++
	
	// Calculate next retry time with exponential backoff
	backoffSeconds := p.config.RetryPolicy.BackoffBaseSeconds * 
		int(math.Pow(float64(p.config.RetryPolicy.BackoffMultiplier), float64(op.RetryCount-1)))
	
	// Add jitter if enabled
	if p.config.RetryPolicy.BackoffJitter {
		jitter := rand.Intn(backoffSeconds / 2)
		backoffSeconds += jitter
	}
	
	scheduledAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)
	
	p.logger.WithFields(logrus.Fields{
		"operation_id":  op.ID,
		"retry_count":   op.RetryCount,
		"next_retry_in": backoffSeconds,
		"error":         err,
	}).Warn("Scheduling operation for retry")
	
	// Update in database
	errMsg := err.Error()
	_, dbErr := p.db.Exec(`
		UPDATE operations 
		SET status = $1, retry_count = $2, scheduled_at = $3, error_message = $4, 
		    started_at = NULL, updated_at = $5
		WHERE id = $6
	`, StatusPending, op.RetryCount, scheduledAt, errMsg, time.Now(), op.ID)
	
	if dbErr != nil {
		p.logger.WithError(dbErr).Error("Failed to schedule retry")
	}
}

// GetMetrics returns current processing metrics
func (p *Processor) GetMetrics() ProcessingMetrics {
	p.metricsLock.RLock()
	defer p.metricsLock.RUnlock()
	
	metrics := ProcessingMetrics{
		QueueDepth:        make(map[Priority]int),
		ProcessingCount:   make(map[Priority]int),
		CompletedCount:    make(map[OperationType]int64),
		FailedCount:       make(map[OperationType]int64),
		AvgProcessingTime: make(map[OperationType]float64),
		WorkerUtilization: float64(atomic.LoadInt32(&p.activeWorkers)) / float64(p.config.TotalCapacity),
	}
	
	// Copy completed counts
	for opType, count := range p.processedOps {
		metrics.CompletedCount[opType] = atomic.LoadInt64(count)
	}
	
	// Query queue depths and processing counts from database
	rows, err := p.db.Query(`
		SELECT priority, status, COUNT(*) 
		FROM operations 
		WHERE status IN ('pending', 'processing')
		GROUP BY priority, status
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var priority Priority
			var status Status
			var count int
			if err := rows.Scan(&priority, &status, &count); err == nil {
				if status == StatusPending {
					metrics.QueueDepth[priority] = count
				} else if status == StatusProcessing {
					metrics.ProcessingCount[priority] = count
				}
			}
		}
	}
	
	return metrics
}

// collectMetrics periodically collects metrics
func (p *Processor) collectMetrics(ctx context.Context) {
	defer p.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := p.GetMetrics()
			p.logger.WithFields(logrus.Fields{
				"queue_depth":        metrics.QueueDepth,
				"processing_count":   metrics.ProcessingCount,
				"worker_utilization": metrics.WorkerUtilization,
			}).Debug("Pipeline metrics")
		}
	}
}