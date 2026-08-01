package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gonotify/gonotify/internal/queue"
)

type Event struct {
	Type       string         `json:"type"`
	TenantID   string         `json:"tenant_id"`
	UserID     string         `json:"user_id"`
	ItemID     string         `json:"item_id,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type Publisher struct {
	bus    *queue.NATSBus
	logger *slog.Logger
}

func NewPublisher(bus *queue.NATSBus, logger *slog.Logger) *Publisher {
	return &Publisher{
		bus:    bus,
		logger: logger,
	}
}

func Subject(tenantID string, userID string) string {
	return fmt.Sprintf("gn.inbox.%s.%s", tenantID, userID)
}

func (p *Publisher) Publish(
	ctx context.Context,
	tenantID string,
	userID string,
	event Event,
) error {
	if p == nil || p.bus == nil {
		return nil
	}

	event.TenantID = tenantID
	event.UserID = userID

	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal realtime event: %w", err)
	}

	subject := Subject(tenantID, userID)

	if err := p.bus.PublishTransient(subject, data); err != nil {
		return fmt.Errorf("publish realtime event: %w", err)
	}

	p.logger.DebugContext(
		ctx,
		"published realtime event",
		"subject", subject,
		"type", event.Type,
		"item_id", event.ItemID,
	)

	return nil
}
