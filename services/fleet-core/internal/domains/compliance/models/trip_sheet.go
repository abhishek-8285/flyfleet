package models

import (
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	TripStatusCreated    TripStatus = "CREATED"
	TripStatusAssigned   TripStatus = "ASSIGNED"
	TripStatusInProgress TripStatus = "IN_PROGRESS"
	TripStatusCompleted  TripStatus = "COMPLETED"
	TripStatusClosed     TripStatus = "CLOSED"
)

type TripSheet struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	VehicleID  uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	DriverID   uuid.UUID  `json:"driver_id" db:"driver_id"`
	TripStart  *time.Time `json:"trip_start" db:"trip_start"`
	TripEnd    *time.Time `json:"trip_end" db:"trip_end"`
	Origin     string     `json:"origin" db:"origin"`
	Destination string     `json:"destination" db:"destination"`
	Status     TripStatus `json:"status" db:"status"`
	DistanceKM float64    `json:"distance_km" db:"distance_km"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

type TripEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TripID    uuid.UUID `json:"trip_id" db:"trip_id"`
	EventType string    `json:"event_type" db:"event_type"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Location  string    `json:"location" db:"location"`
	Metadata  []byte    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
