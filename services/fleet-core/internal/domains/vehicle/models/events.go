package models

import (
	"time"

	"github.com/google/uuid"
)

// Event type constants per spec §3 / §5.
const (
	EventVehicleCreated          = "vehicle.created"
	EventVehicleUpdated          = "vehicle.updated"
	EventVehicleStatusChanged    = "vehicle.status.changed"
	EventVehicleComplianceExpired = "vehicle.compliance.expired"
	EventVehicleDecommissioned   = "vehicle.decommissioned"
	EventVehicleDocumentUploaded = "vehicle.document.uploaded"
)

// VehicleEvent is an append-only, hash-chained audit event.
type VehicleEvent struct {
	ID        uuid.UUID              `json:"id"          db:"id"`
	VehicleID uuid.UUID              `json:"vehicle_id"  db:"vehicle_id"`
	EventType string                 `json:"event_type"  db:"event_type"`
	Payload   map[string]interface{} `json:"payload"     db:"payload"`
	Hash      string                 `json:"hash"        db:"hash"`     // SHA-256 of (vehicle_id || event_type || payload || prev_hash)
	PrevHash  *string                `json:"prev_hash"   db:"prev_hash"` // nil for first event
	ActorID   *uuid.UUID             `json:"actor_id"    db:"actor_id"`
	CreatedAt time.Time              `json:"created_at"  db:"created_at"`
}

// VehicleEventList is the response for listing events.
type VehicleEventList struct {
	Events []VehicleEvent `json:"events"`
	Total  int64          `json:"total"`
}
