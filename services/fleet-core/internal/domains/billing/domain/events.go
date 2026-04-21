package domain

import (
	"context"
	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID   `json:"id"`
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
	Metadata  Metadata    `json:"metadata"`
}

type Metadata struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
}
