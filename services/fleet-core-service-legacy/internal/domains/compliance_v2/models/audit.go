package models

import (
	"time"

	"github.com/google/uuid"
)

type ComplianceAuditLog struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ActorID    uuid.UUID `json:"actor_id" db:"actor_id"`
	Action     string    `json:"action" db:"action"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id" db:"entity_id"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Metadata   []byte    `json:"metadata" db:"metadata"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
