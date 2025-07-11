package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	
	"github.com/orca-ng/orca/internal/crypto"
	"github.com/orca-ng/orca/internal/cyberark"
	"github.com/orca-ng/orca/internal/database"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
	"github.com/orca-ng/orca/internal/services"
)

// SimpleProcessor processes operations one by one with proper token management
type SimpleProcessor struct {
	db              *database.GormDB
	config          *PipelineConfig
	handlers        map[OperationType]OperationHandler
	logger          *logrus.Logger
	certManager     *services.CertificateManager
	encryptor       *crypto.Encryptor
	events          *services.OperationEventService
	
	// Processing state
	ctx             context.Context
	cancel          context.CancelFunc
	processing      sync.WaitGroup
	
	// CyberArk session management
	sessions        map[string]*cyberArkSession  // instanceID -> session
	sessionsMutex   sync.RWMutex
}

// cyberArkSession represents an authenticated session with a CyberArk instance
type cyberArkSession struct {
	client      *cyberark.Client
	token       string
	instanceID  string
	lastUsed    time.Time
	mutex       sync.Mutex
}

// NewSimpleProcessor creates a new simplified pipeline processor
func NewSimpleProcessor(db *database.GormDB, config *PipelineConfig, logger *logrus.Logger, certManager *services.CertificateManager, encryptionKey string, events *services.OperationEventService) *SimpleProcessor {
	return &SimpleProcessor{
		db:          db,
		config:      config,
		handlers:    make(map[OperationType]OperationHandler),
		logger:      logger,
		certManager: certManager,
		encryptor:   crypto.NewEncryptor(encryptionKey),
		events:      events,
		sessions:    make(map[string]*cyberArkSession),
	}
}

// RegisterHandler registers a handler for a specific operation type
func (p *SimpleProcessor) RegisterHandler(opType OperationType, handler OperationHandler) {
	p.handlers[opType] = handler
}

// Start begins processing operations one by one
func (p *SimpleProcessor) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)
	
	p.logger.Info("Starting simple pipeline processor")
	
	// Start the processor loop
	p.processing.Add(1)
	go p.processLoop()
	
	// Start session cleanup routine
	p.processing.Add(1)
	go p.cleanupSessions()
	
	return nil
}

// Stop gracefully shuts down the processor
func (p *SimpleProcessor) Stop() error {
	p.logger.Info("Stopping simple pipeline processor")
	
	// Signal shutdown
	p.cancel()
	
	// Wait for processing to complete with timeout
	done := make(chan struct{})
	go func() {
		p.processing.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Clean up all sessions
		p.closeAllSessions()
		p.logger.Info("Simple pipeline processor stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for processor to stop")
	}
}

// processLoop is the main processing loop
func (p *SimpleProcessor) processLoop() {
	defer p.processing.Done()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			// Try to fetch and process the next operation
			if err := p.processNext(); err != nil {
				if err != gorm.ErrRecordNotFound {
					p.logger.WithError(err).Error("Error processing operation")
				}
				// Wait before checking for more work
				time.Sleep(time.Second)
			}
		}
	}
}

// processNext fetches and processes the next operation
func (p *SimpleProcessor) processNext() error {
	var op gormmodels.Operation
	
	// Use a transaction to fetch and lock the next operation
	err := p.db.Transaction(func(tx *gorm.DB) error {
		// Fetch next pending operation, ordered by priority and scheduled time
		// Process high priority first, then normal, then low
		result := tx.Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
			Where("status = ? AND scheduled_at <= ?", 
				gormmodels.OpStatusPending, time.Now()).
			Order("CASE priority WHEN 'high' THEN 1 WHEN 'normal' THEN 2 WHEN 'low' THEN 3 END, scheduled_at").
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
		
		// Update the operation object with new values
		op.Status = gormmodels.OpStatusProcessing
		op.StartedAt = &now
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// Process the operation
	p.logger.WithFields(logrus.Fields{
		"operation_id": op.ID,
		"type":        op.Type,
		"priority":    op.Priority,
		"instance_id": op.CyberArkInstanceID,
	}).Info("Processing operation")
	
	// Publish processing event
	if p.events != nil {
		p.events.PublishOperationUpdated(&op)
	}
	
	// Execute the operation with the appropriate CyberArk session
	err = p.executeOperation(&op)
	
	if err != nil {
		// Check if retryable
		handler, exists := p.handlers[OperationType(op.Type)]
		if exists && handler.CanRetry(err) && op.RetryCount < op.MaxRetries {
			p.retryOperation(&op, err)
		} else {
			p.completeOperation(&op, nil, err)
		}
	} else {
		// Success
		p.completeOperation(&op, nil, nil)
	}
	
	return nil
}

// executeOperation executes an operation with proper session management
func (p *SimpleProcessor) executeOperation(op *gormmodels.Operation) error {
	// Get handler
	handler, exists := p.handlers[OperationType(op.Type)]
	if !exists {
		return fmt.Errorf("no handler registered for operation type: %s", op.Type)
	}
	
	// Get or create CyberArk session if instance ID is set
	var session *cyberArkSession
	if op.CyberArkInstanceID != nil {
		var err error
		session, err = p.getOrCreateSession(*op.CyberArkInstanceID)
		if err != nil {
			return fmt.Errorf("get CyberArk session: %w", err)
		}
	}
	
	// Create timeout context
	timeout := p.config.DefaultTimeout
	if t, ok := p.config.OperationTimeouts[OperationType(op.Type)]; ok {
		timeout = t
	}
	
	ctx, cancel := context.WithTimeout(p.ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	
	// Convert to pipeline operation
	pipelineOp := &Operation{
		ID:                 op.ID,
		Type:               OperationType(op.Type),
		Priority:           Priority(op.Priority),
		Status:             Status(op.Status),
		Payload:            op.Payload,
		Result:             op.Result,
		ErrorMessage:       op.ErrorMessage,
		RetryCount:         op.RetryCount,
		MaxRetries:         op.MaxRetries,
		ScheduledAt:        op.ScheduledAt,
		StartedAt:          op.StartedAt,
		CompletedAt:        op.CompletedAt,
		CreatedBy:          op.CreatedBy,
		CyberArkInstanceID: op.CyberArkInstanceID,
		CorrelationID:      op.CorrelationID,
		CreatedAt:          op.CreatedAt,
		UpdatedAt:          op.UpdatedAt,
	}
	
	// Inject CyberArk client into context if we have a session
	if session != nil {
		ctx = context.WithValue(ctx, "cyberark_client", session.client)
	}
	
	// Execute handler
	err := handler.Handle(ctx, pipelineOp)
	
	// Update operation result if handler modified it
	if pipelineOp.Result != nil {
		op.Result = pipelineOp.Result
	}
	
	return err
}

// getOrCreateSession gets an existing session or creates a new one
func (p *SimpleProcessor) getOrCreateSession(instanceID string) (*cyberArkSession, error) {
	// Check for existing session
	p.sessionsMutex.RLock()
	session, exists := p.sessions[instanceID]
	p.sessionsMutex.RUnlock()
	
	if exists {
		// Validate session is still good
		session.mutex.Lock()
		defer session.mutex.Unlock()
		
		// Check if token is still valid (simple time-based check)
		// CyberArk tokens typically expire after 20 minutes of inactivity
		if time.Since(session.lastUsed) < 15*time.Minute {
			session.lastUsed = time.Now()
			return session, nil
		}
		
		// Token might be expired, try to use it and recreate if needed
		// For now, we'll recreate to be safe
		p.logger.WithField("instance_id", instanceID).Debug("Session expired, creating new one")
	}
	
	// Create new session
	return p.createSession(instanceID)
}

// createSession creates a new authenticated session
func (p *SimpleProcessor) createSession(instanceID string) (*cyberArkSession, error) {
	// Load instance configuration
	var instance gormmodels.CyberArkInstance
	if err := p.db.Where("id = ?", instanceID).First(&instance).Error; err != nil {
		return nil, fmt.Errorf("load instance: %w", err)
	}
	
	// Decrypt password
	password, err := p.encryptor.Decrypt(instance.PasswordEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt password: %w", err)
	}
	
	// Create client
	client, err := cyberark.NewClient(cyberark.Config{
		BaseURL:        instance.BaseURL,
		Username:       instance.Username,
		Password:       password,
		SkipTLSVerify:  instance.SkipTLSVerify,
		RequestTimeout: 30 * time.Second,
		CertManager:    p.certManager,
	})
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	
	// Authenticate
	token, err := client.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	
	// Create session
	session := &cyberArkSession{
		client:     client,
		token:      token,
		instanceID: instanceID,
		lastUsed:   time.Now(),
	}
	
	// Store session
	p.sessionsMutex.Lock()
	oldSession := p.sessions[instanceID]
	p.sessions[instanceID] = session
	p.sessionsMutex.Unlock()
	
	// Clean up old session if exists
	if oldSession != nil {
		go func() {
			if err := oldSession.client.Logoff(); err != nil {
				p.logger.WithError(err).Warn("Failed to logoff old session")
			}
		}()
	}
	
	p.logger.WithField("instance_id", instanceID).Info("Created new CyberArk session")
	
	return session, nil
}

// cleanupSessions periodically cleans up expired sessions
func (p *SimpleProcessor) cleanupSessions() {
	defer p.processing.Done()
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions removes sessions that haven't been used recently
func (p *SimpleProcessor) cleanupExpiredSessions() {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()
	
	for instanceID, session := range p.sessions {
		session.mutex.Lock()
		if time.Since(session.lastUsed) > 20*time.Minute {
			// Session expired, remove it
			delete(p.sessions, instanceID)
			
			// Logoff in background
			go func(s *cyberArkSession) {
				if err := s.client.Logoff(); err != nil {
					p.logger.WithError(err).Warn("Failed to logoff expired session")
				}
			}(session)
			
			p.logger.WithField("instance_id", instanceID).Debug("Cleaned up expired session")
		}
		session.mutex.Unlock()
	}
}

// closeAllSessions closes all active sessions
func (p *SimpleProcessor) closeAllSessions() {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()
	
	for instanceID, session := range p.sessions {
		if err := session.client.Logoff(); err != nil {
			p.logger.WithError(err).Warn("Failed to logoff session")
		}
		delete(p.sessions, instanceID)
	}
}

// completeOperation marks an operation as completed or failed
func (p *SimpleProcessor) completeOperation(op *gormmodels.Operation, result json.RawMessage, err error) {
	now := time.Now()
	updates := map[string]interface{}{
		"completed_at": now,
	}
	
	var errMsg string
	if err != nil {
		updates["status"] = gormmodels.OpStatusFailed
		errMsg = err.Error()
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
		return
	}
	
	// Apply updates to the operation object
	if err != nil {
		op.Status = gormmodels.OpStatusFailed
		op.ErrorMessage = &errMsg
	} else {
		op.Status = gormmodels.OpStatusCompleted
		if result != nil {
			op.Result = &result
		}
	}
	op.CompletedAt = &now
	
	// Publish event
	if p.events != nil {
		p.events.PublishOperationUpdated(op)
	}
}

// retryOperation schedules an operation for retry
func (p *SimpleProcessor) retryOperation(op *gormmodels.Operation, err error) {
	op.RetryCount++
	
	// Simple exponential backoff: 10s, 20s, 40s, etc.
	backoffSeconds := 10 * (1 << (op.RetryCount - 1))
	if backoffSeconds > 300 {
		backoffSeconds = 300 // Max 5 minutes
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
		return
	}
	
	// Update the operation object
	op.Status = gormmodels.OpStatusPending
	op.ScheduledAt = scheduledAt
	op.ErrorMessage = &errMsg
	op.StartedAt = nil
	
	// Publish retry event
	if p.events != nil {
		p.events.PublishOperationUpdated(op)
	}
}

// GetMetrics returns simplified metrics
func (p *SimpleProcessor) GetMetrics() ProcessingMetrics {
	metrics := ProcessingMetrics{
		QueueDepth:        make(map[Priority]int),
		ProcessingCount:   make(map[Priority]int),
		CompletedCount:    make(map[OperationType]int64),
		FailedCount:       make(map[OperationType]int64),
		AvgProcessingTime: make(map[OperationType]float64),
		WorkerUtilization: 0.0, // Will be 0 or 1 for single processor
	}
	
	// Count pending operations by priority
	var pendingCounts []struct {
		Priority string
		Count    int64
	}
	
	p.db.Model(&gormmodels.Operation{}).
		Select("priority, COUNT(*) as count").
		Where("status = ?", gormmodels.OpStatusPending).
		Group("priority").
		Scan(&pendingCounts)
		
	for _, pc := range pendingCounts {
		metrics.QueueDepth[Priority(pc.Priority)] = int(pc.Count)
	}
	
	// Check if we're processing
	var processingCount int64
	p.db.Model(&gormmodels.Operation{}).
		Where("status = ?", gormmodels.OpStatusProcessing).
		Count(&processingCount)
		
	if processingCount > 0 {
		metrics.WorkerUtilization = 1.0
	}
	
	// Active sessions count
	p.sessionsMutex.RLock()
	sessionCount := len(p.sessions)
	p.sessionsMutex.RUnlock()
	
	p.logger.WithFields(logrus.Fields{
		"queue_depth":      metrics.QueueDepth,
		"processing":       processingCount > 0,
		"active_sessions":  sessionCount,
	}).Debug("Pipeline metrics")
	
	return metrics
}