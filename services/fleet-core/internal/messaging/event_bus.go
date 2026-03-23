package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Event represents a standardized platform event
type Event struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Data      any       `json:"data"`
	OrgID     string    `json:"org_id"`
	Timestamp time.Time `json:"timestamp"`
}

// EventBus defines the interface for standardized platform messaging
type EventBus interface {
	Publish(ctx context.Context, subject string, data any, orgID string) error
	Subscribe(subject string, handler func(event Event)) (*nats.Subscription, error)
}

// NatsEventBus implements EventBus using NATS
type NatsEventBus struct {
	nc     *nats.Conn
	logger *zap.Logger
}

// NewNatsEventBus creates a new NATS-backed event bus
func NewNatsEventBus(nc *nats.Conn, logger *zap.Logger) EventBus {
	return &NatsEventBus{
		nc:     nc,
		logger: logger,
	}
}

// Publish publishes an event to a subject with multi-tenant context
func (b *NatsEventBus) Publish(ctx context.Context, subject string, data any, orgID string) error {
	event := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Subject:   subject,
		Data:      data,
		OrgID:     orgID,
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := b.nc.Publish(subject, payload); err != nil {
		return fmt.Errorf("failed to publish %s to NATS: %w", subject, err)
	}

	b.logger.Debug("event published", zap.String("subject", subject), zap.String("org_id", orgID))
	return nil
}

// Subscribe subscribes to a subject and handles events
func (b *NatsEventBus) Subscribe(subject string, handler func(event Event)) (*nats.Subscription, error) {
	return b.nc.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			b.logger.Error("failed to unmarshal event on subject "+subject, zap.Error(err))
			return
		}
		handler(event)
	})
}
