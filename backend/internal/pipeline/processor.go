package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/database"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
)

// Processor manages the processing pipeline with priority lanes using GORM
type Processor struct {
	db              *database.GormDB
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

// NewProcessor creates a new pipeline processor using GORM
func NewProcessor(db *database.GormDB, config *PipelineConfig, logger *logrus.Logger) *Processor {
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
				if !errors.Is(err, gorm.ErrRecordNotFound) {
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
	var op gormmodels.Operation
	
	// Use a transaction to fetch and lock an operation
	err := w.proc.db.Transaction(func(tx *gorm.DB) error {
		// Fetch next pending operation for this priority
		// Using raw SQL for FOR UPDATE SKIP LOCKED as GORM doesn't have direct support
		result := tx.Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
			Where("status = ? AND priority = ? AND scheduled_at <= ?", 
				gormmodels.OpStatusPending, w.priority, time.Now()).
			Order("scheduled_at").
			Limit(1).
			First(&op)
		
		if result.Error != nil {
			return result.Error
		}
		
		// Mark as processing
		now := time.Now()
		updates := map[string]interface{}{
			"status":     gormmodels.OpStatusProcessing,
			"started_at": now,
		}
		
		if err := tx.Model(&op).Updates(updates).Error; err != nil {
			return fmt.Errorf("update operation status: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		return err
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
	
	// Convert GORM operation to pipeline Operation
	pipelineOp := w.convertToOperation(&op)
	
	// Get handler
	handler, exists := w.proc.handlers[OperationType(op.Type)]
	if !exists {
		w.proc.completeOperation(&op, nil, fmt.Errorf("no handler registered for operation type: %s", op.Type))
		return nil
	}
	
	// Create timeout context
	timeout := w.proc.config.DefaultTimeout
	if t, ok := w.proc.config.OperationTimeouts[OperationType(op.Type)]; ok {
		timeout = t
	}
	
	opCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	
	// Execute handler
	err = handler.Handle(opCtx, pipelineOp)
	
	if err != nil {
		// Check if retryable
		if handler.CanRetry(err) && op.RetryCount < op.MaxRetries {
			w.proc.retryOperation(&op, err)
		} else {
			w.proc.completeOperation(&op, nil, err)
		}
	} else {
		// Success - use Result from the pipeline operation
		var result json.RawMessage
		if pipelineOp.Result != nil {
			result = *pipelineOp.Result
		}
		w.proc.completeOperation(&op, result, nil)
		
		// Update metrics
		atomic.AddInt64(w.proc.processedOps[OperationType(op.Type)], 1)
	}
	
	return nil
}

// convertToOperation converts GORM operation to pipeline Operation
func (w *worker) convertToOperation(gormOp *gormmodels.Operation) *Operation {
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

// completeOperation marks an operation as completed or failed
func (p *Processor) completeOperation(op *gormmodels.Operation, result json.RawMessage, err error) {
	now := time.Now()
	updates := map[string]interface{}{
		"completed_at": now,
	}
	
	if err != nil {
		updates["status"] = gormmodels.OpStatusFailed
		errMsg := err.Error()
		updates["error_message"] = errMsg
		
		p.logger.WithFields(logrus.Fields{
			"operation_id": op.ID,
			"error":        err,
		}).Error("Operation failed")
	} else {
		updates["status"] = gormmodels.OpStatusCompleted
		if result != nil {
			updates["result"] = result
		}
		
		duration := now.Sub(*op.StartedAt).Seconds()
		p.logger.WithFields(logrus.Fields{
			"operation_id": op.ID,
			"duration":     duration,
		}).Info("Operation completed")
	}
	
	// Update in database
	if dbErr := p.db.Model(op).Updates(updates).Error; dbErr != nil {
		p.logger.WithError(dbErr).Error("Failed to update operation status")
	}
}

// retryOperation schedules an operation for retry
func (p *Processor) retryOperation(op *gormmodels.Operation, err error) {
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
	updates := map[string]interface{}{
		"status":        gormmodels.OpStatusPending,
		"retry_count":   op.RetryCount,
		"scheduled_at":  scheduledAt,
		"error_message": errMsg,
		"started_at":    nil,
	}
	
	if dbErr := p.db.Model(op).Updates(updates).Error; dbErr != nil {
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
	type metricResult struct {
		Priority string
		Status   string
		Count    int64
	}
	
	var results []metricResult
	err := p.db.Model(&gormmodels.Operation{}).
		Select("priority, status, COUNT(*) as count").
		Where("status IN ?", []string{gormmodels.OpStatusPending, gormmodels.OpStatusProcessing}).
		Group("priority, status").
		Scan(&results).Error
		
	if err == nil {
		for _, r := range results {
			priority := Priority(r.Priority)
			if r.Status == gormmodels.OpStatusPending {
				metrics.QueueDepth[priority] = int(r.Count)
			} else if r.Status == gormmodels.OpStatusProcessing {
				metrics.ProcessingCount[priority] = int(r.Count)
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