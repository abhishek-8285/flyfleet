package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverCertification represents the driver certifying their daily logs
type DriverCertification struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	DriverID    uuid.UUID  `json:"driver_id" db:"driver_id"`
	LogDate     time.Time  `json:"log_date" db:"log_date"`
	Certified   bool       `json:"certified" db:"certified"`
	Signature   string     `json:"signature" db:"signature"`
	CertifiedAt *time.Time `json:"certified_at,omitempty" db:"certified_at"`
}

// ELDEvent represents low level FMCSA ELD events like power loops and tracking telemetry sequences
type ELDEvent struct {
	EventID     uuid.UUID  `json:"event_id" db:"event_id"`
	SequenceID  int64      `json:"sequence_id" db:"sequence_id"`
	DriverID    *uuid.UUID `json:"driver_id" db:"driver_id"`
	VehicleID   uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	EventType   string     `json:"event_type" db:"event_type"` // ENGINE_POWER_ON, VEHICLE_MOTION, etc
	Timestamp   time.Time  `json:"timestamp" db:"timestamp"`
	Location    string     `json:"location" db:"location"`
	Odometer    float64    `json:"odometer" db:"odometer"`
	EngineHours float64    `json:"engine_hours" db:"engine_hours"`
	Origin      string     `json:"origin" db:"origin"`
	Checksum    string     `json:"checksum" db:"checksum"`
}

// UnidentifiedDriverEvent is generated when vehicle moves without driver
type UnidentifiedDriverEvent struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	VehicleID        uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	EventTime        time.Time  `json:"event_time" db:"event_time"`
	Location         string     `json:"location" db:"location"`
	Odometer         float64    `json:"odometer" db:"odometer"`
	EngineHours      float64    `json:"engine_hours" db:"engine_hours"`
	AssignedDriverID *uuid.UUID `json:"assigned_driver_id,omitempty" db:"assigned_driver_id"`
	Resolved         bool       `json:"resolved" db:"resolved"`
}

// EventEditHistory records the audit trail for log edits as per ELD mandate
type EventEditHistory struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OriginalEventID uuid.UUID `json:"original_event_id" db:"original_event_id"`
	EditedEventID   uuid.UUID `json:"edited_event_id" db:"edited_event_id"`
	EditReason      string    `json:"edit_reason" db:"edit_reason"`
	EditedBy        uuid.UUID `json:"edited_by" db:"edited_by"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
}
