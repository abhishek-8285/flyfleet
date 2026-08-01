package queue

import (
	"context"
	"encoding/json"
	"time"
)

type Message struct {
	ID         string            `json:"id"`
	TenantID   string            `json:"tenant_id"`
	Type       string            `json:"type"`
	OccurredAt time.Time         `json:"occurred_at"`
	Payload    json.RawMessage   `json:"payload"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type Handler func(ctx context.Context, msg Message) error

type Bus interface {
	Publish(ctx context.Context, subject string, msg Message) error
	Subscribe(subject string, queueGroup string, handler Handler) error
	Close() error
}
