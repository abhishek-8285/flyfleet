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
	OrgID     uuid.UUID `json:"org_id"`
	UserID    uuid.UUID `json:"user_id"`
	RequestID uuid.UUID `json:"request_id"`
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
}

