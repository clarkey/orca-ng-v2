package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	gormmodels "github.com/orca-ng/orca/internal/models/gorm"
)

// OperationEvent represents an operation state change event
type OperationEvent struct {
	Type      string                 `json:"type"` // "created", "updated", "completed", "failed"
	Operation *gormmodels.Operation  `json:"operation"`
	Timestamp time.Time              `json:"timestamp"`
}

// OperationEventService manages operation event subscriptions
type OperationEventService struct {
	mu          sync.RWMutex
	subscribers map[string]chan *OperationEvent
	logger      *logrus.Logger
}

// NewOperationEventService creates a new operation event service
func NewOperationEventService(logger *logrus.Logger) *OperationEventService {
	return &OperationEventService{
		subscribers: make(map[string]chan *OperationEvent),
		logger:      logger,
	}
}

// Subscribe creates a new subscription for operation events
func (s *OperationEventService) Subscribe(ctx context.Context, clientID string) <-chan *OperationEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create buffered channel to prevent blocking
	ch := make(chan *OperationEvent, 100)
	s.subscribers[clientID] = ch

	// Remove subscription when context is cancelled
	go func() {
		<-ctx.Done()
		s.Unsubscribe(clientID)
	}()

	s.logger.WithField("client_id", clientID).Debug("Client subscribed to operation events")
	return ch
}

// Unsubscribe removes a subscription
func (s *OperationEventService) Unsubscribe(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, exists := s.subscribers[clientID]; exists {
		close(ch)
		delete(s.subscribers, clientID)
		s.logger.WithField("client_id", clientID).Debug("Client unsubscribed from operation events")
	}
}

// PublishOperationCreated publishes an operation created event
func (s *OperationEventService) PublishOperationCreated(operation *gormmodels.Operation) {
	s.publish(&OperationEvent{
		Type:      "created",
		Operation: operation,
		Timestamp: time.Now(),
	})
}

// PublishOperationUpdated publishes an operation updated event
func (s *OperationEventService) PublishOperationUpdated(operation *gormmodels.Operation) {
	eventType := "updated"
	
	// Determine specific event type based on status
	switch operation.Status {
	case gormmodels.OpStatusProcessing:
		eventType = "started"
	case gormmodels.OpStatusCompleted:
		eventType = "completed"
	case gormmodels.OpStatusFailed:
		eventType = "failed"
	case gormmodels.OpStatusCancelled:
		eventType = "cancelled"
	}

	s.publish(&OperationEvent{
		Type:      eventType,
		Operation: operation,
		Timestamp: time.Now(),
	})
}

// publish sends an event to all subscribers
func (s *OperationEventService) publish(event *OperationEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Log the event
	s.logger.WithFields(logrus.Fields{
		"type":         event.Type,
		"operation_id": event.Operation.ID,
		"status":       event.Operation.Status,
	}).Debug("Publishing operation event")

	// Send to all subscribers
	for clientID, ch := range s.subscribers {
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Channel is full, log and skip
			s.logger.WithField("client_id", clientID).Warn("Operation event channel full, skipping event")
		}
	}
}

// GetActiveSubscriberCount returns the number of active subscribers
func (s *OperationEventService) GetActiveSubscriberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers)
}

// MarshalEventToJSON marshals an operation event to JSON for SSE
func MarshalEventToJSON(event *OperationEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(data), nil
}