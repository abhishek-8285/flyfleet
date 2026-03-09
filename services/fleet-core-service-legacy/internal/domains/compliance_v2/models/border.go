package models

import (
	"time"

	"github.com/google/uuid"
)

type StateBorderCrossing struct {
	ID          uuid.UUID `json:"id" db:"id"`
	VehicleID   uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	FromState   string    `json:"from_state" db:"from_state"`
	ToState     string    `json:"to_state" db:"to_state"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	GPSLocation string    `json:"gps_location" db:"gps_location"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
