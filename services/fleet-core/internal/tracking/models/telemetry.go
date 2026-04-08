package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// VehicleStatus constants derived from speed + ignition + age.
const (
	StatusMoving  = "moving"
	StatusIdle    = "idle"
	StatusStopped = "stopped"
	StatusOffline = "offline"
)

// TelemetryEvent represents a single GPS position report from a vehicle device.
// Aligned with AIS-140 telemetry format.
type TelemetryEvent struct {
	EventID     uuid.UUID  `json:"event_id"`
	DeviceID    uuid.UUID  `json:"device_id"`
	VehicleID   uuid.UUID  `json:"vehicle_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Lat         float64    `json:"latitude"`
	Lng         float64    `json:"longitude"`
	Speed       float64    `json:"speed_kmh"`
	Heading     float64    `json:"heading"`
	Ignition    bool       `json:"ignition"`
	FuelLevel   float64    `json:"fuel_level"`
	Odometer    float64    `json:"odometer"`
	SequenceNo  int64      `json:"sequence_no"`
	RecordedAt  time.Time  `json:"recorded_at"`
	ReceivedAt  time.Time  `json:"received_at"`
	RawData     string     `json:"raw_data,omitempty"`
}

// LiveLocation is the cached "last known position" for a vehicle,
// stored in Redis with a 5-minute TTL.
type LiveLocation struct {
	VehicleID  uuid.UUID `json:"vehicle_id"`
	Lat        float64   `json:"latitude"`
	Lng        float64   `json:"longitude"`
	Speed      float64   `json:"speed_kmh"`
	Heading    float64   `json:"heading"`
	Ignition   bool      `json:"ignition"`
	FuelLevel  float64   `json:"fuel_level"`
	Odometer   float64   `json:"odometer"`
	LastSeen   time.Time `json:"last_seen"`
	Status     string    `json:"status"` // moving/idle/stopped/offline
}

// VehicleLocationResponse is the enriched response for the bulk
// GET /api/v1/vehicles/locations endpoint, joining vehicle +
// driver + live location data.
type VehicleLocationResponse struct {
	VehicleID       string       `json:"vehicle_id"`
	Registration    string       `json:"registration"`
	Make            string       `json:"make"`
	Model           string       `json:"model"`
	Latitude        float64      `json:"latitude"`
	Longitude       float64      `json:"longitude"`
	SpeedKmh        float64      `json:"speed_kmh"`
	Heading         float64      `json:"heading"`
	Ignition        bool         `json:"ignition"`
	FuelLevel       float64      `json:"fuel_level"`
	Odometer        float64      `json:"odometer"`
	Status          string       `json:"status"`
	Driver          *DriverInfo  `json:"driver,omitempty"`
	LastUpdate      string       `json:"last_update"`
	LastSeenSeconds int64        `json:"last_seen_seconds_ago"`
}

// DriverInfo carries minimal driver data for display on the map.
type DriverInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// CalculateStatus derives a vehicle's operational status from
// speed, ignition state, and how recently we heard from it.
func CalculateStatus(speed float64, ignition bool, lastSeen time.Time) string {
	age := time.Since(lastSeen)
	if age > 5*time.Minute {
		return StatusOffline
	}
	if !ignition {
		return StatusStopped
	}
	if speed < 2.0 {
		return StatusIdle
	}
	return StatusMoving
}
