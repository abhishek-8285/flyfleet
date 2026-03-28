package models

import (
	"time"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────────────────────
// Legacy telemetry types (kept for backward-compat with existing handlers)
// ─────────────────────────────────────────────────────────────────────────────

type GPSTelemetry struct {
	Time       time.Time              `json:"time" validate:"required"`
	VehicleID  uuid.UUID              `json:"vehicle_id" validate:"required"`
	OrgID      uuid.UUID              `json:"org_id" validate:"required"`
	Latitude   float64                `json:"latitude" validate:"required"`
	Longitude  float64                `json:"longitude" validate:"required"`
	Altitude   *float64               `json:"altitude,omitempty"`
	Speed      *float64               `json:"speed,omitempty"`
	Heading    *float64               `json:"heading,omitempty"`
	Accuracy   *float64               `json:"accuracy,omitempty"`
	Ignition   *bool                  `json:"ignition,omitempty"`
	Odometer   *float64               `json:"odometer,omitempty"`
	Battery    *int                   `json:"battery_level,omitempty"`
	Satellites *int                   `json:"satellites,omitempty"`
	RawData    map[string]interface{} `json:"raw_data,omitempty"`
}

type OBDTelemetry struct {
	Time        time.Time              `json:"time" validate:"required"`
	VehicleID   uuid.UUID              `json:"vehicle_id" validate:"required"`
	OrgID       uuid.UUID              `json:"org_id" validate:"required"`
	RPM         *int                   `json:"rpm,omitempty"`
	CoolantTemp *float64               `json:"coolant_temp,omitempty"`
	EngineLoad  *float64               `json:"engine_load,omitempty"`
	ThrottlePos *float64               `json:"throttle_pos,omitempty"`
	FuelLevel   *float64               `json:"fuel_level,omitempty"`
	DTCs        []string               `json:"dtcs,omitempty"`
	VIN         *string                `json:"vin,omitempty"`
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
}

type FuelTelemetry struct {
	Time         time.Time              `json:"time" validate:"required"`
	VehicleID    uuid.UUID              `json:"vehicle_id" validate:"required"`
	OrgID        uuid.UUID              `json:"org_id" validate:"required"`
	LevelPercent *float64               `json:"level_percent,omitempty"`
	LevelLiters  *float64               `json:"level_liters,omitempty"`
	RefuelEvent  *bool                  `json:"refuel_event,omitempty"`
	TheftEvent   *bool                  `json:"theft_event,omitempty"`
	Latitude     *float64               `json:"latitude,omitempty"`
	Longitude    *float64               `json:"longitude,omitempty"`
	RawData      map[string]interface{} `json:"raw_data,omitempty"`
}

type TelemetryBatch struct {
	GPSData   []GPSTelemetry  `json:"gps_data,omitempty"`
	OBData    []OBDTelemetry  `json:"obd_data,omitempty"`
	FuelData  []FuelTelemetry `json:"fuel_data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	DeviceID  string          `json:"device_id"`
}

type GeofenceEvent struct {
	ID         uuid.UUID `json:"id"`
	VehicleID  uuid.UUID `json:"vehicle_id"`
	OrgID      uuid.UUID `json:"org_id"`
	GeofenceID uuid.UUID `json:"geofence_id"`
	EventType  string    `json:"event_type"` // ENTRY, EXIT, DWELL
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timestamp  time.Time `json:"timestamp"`
	CreatedAt  time.Time `json:"created_at"`
}

type DeviceHeartbeat struct {
	DeviceID       string    `json:"device_id"`
	VehicleID      uuid.UUID `json:"vehicle_id"`
	OrgID          uuid.UUID `json:"org_id"`
	Status         string    `json:"status"` // ONLINE, OFFLINE, ERROR
	LastSeen       time.Time `json:"last_seen"`
	Battery        *int      `json:"battery_level,omitempty"`
	SignalStrength *int      `json:"signal_strength,omitempty"`
	ErrorCodes     []string  `json:"error_codes,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Enterprise AIS-140 / GPS event schema
// ─────────────────────────────────────────────────────────────────────────────

// EventType constants
const (
	EventTypeGPS    = "GPS"
	EventTypePanic  = "PANIC"
	EventTypeHarsh  = "HARSH"
	EventTypeIdle   = "IDLE"
	EventTypeEngine = "ENGINE"
)

// TelemetryEvent is the canonical AIS-140-aligned ingest payload.
// Partition key: VehicleID.  Kafka topic: telemetry.raw.
type TelemetryEvent struct {
	EventID    string    `json:"event_id" validate:"required"` // UUID from device
	DeviceID   string    `json:"device_id" validate:"required"`
	VehicleID  uuid.UUID `json:"vehicle_id" validate:"required"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Lat        float64   `json:"lat" validate:"required,min=-90,max=90"`
	Lng        float64   `json:"lng" validate:"required,min=-180,max=180"`
	Speed      float64   `json:"speed" validate:"min=0,max=300"`   // km/h
	Heading    int       `json:"heading" validate:"min=0,max=360"` // degrees
	Ignition   bool      `json:"ignition"`
	FuelLevel  *float64  `json:"fuel_level,omitempty"` // %
	Odometer   *float64  `json:"odometer,omitempty"`   // km
	EventType  string    `json:"event_type" validate:"required"`
	SequenceNo int64     `json:"sequence_no"`
	RecordedAt time.Time `json:"recorded_at" validate:"required"`
	ReceivedAt time.Time `json:"received_at"`
	// HMAC signature from device — validated at ingestion layer
	Signature string `json:"signature,omitempty"`
}

// LiveLocation is the Redis-cached live state of a vehicle.
type LiveLocation struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Speed     float64   `json:"speed"`
	Heading   int       `json:"heading"`
	Ignition  bool      `json:"ignition"`
	LastSeen  time.Time `json:"last_seen"`
	Status    string    `json:"status"` // online | offline
}

// ─────────────────────────────────────────────────────────────────────────────
// Anomaly detection
// ─────────────────────────────────────────────────────────────────────────────

// AnomalyType constants match the Kafka topics emitted by the anomaly engine.
const (
	AnomalyOverspeed        = "tracking.overspeed"
	AnomalyRouteDeviation   = "tracking.route_deviation"
	AnomalyTamper           = "tracking.tamper_detected"
	AnomalyDeviceOffline    = "tracking.device_offline"
	AnomalyPanic            = "tracking.panic"
	AnomalyHarshEvent       = "tracking.harsh_event"
	AnomalyFuelDrop         = "tracking.fuel_drop"
	AnomalyOdometerRollback = "tracking.odometer_rollback"
	AnomalyTeleport         = "tracking.teleport"
)

type AnomalyEvent struct {
	Type      string         `json:"type"` // one of Anomaly* constants
	VehicleID uuid.UUID      `json:"vehicle_id"`
	DeviceID  string         `json:"device_id"`
	Severity  string         `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Device command channel
// ─────────────────────────────────────────────────────────────────────────────

// DeviceCommandType constants
const (
	CmdPing         = "ping"
	CmdReboot       = "reboot"
	CmdImmobilize   = "immobilize"
	CmdConfigUpdate = "config_update"
)

type DeviceCommand struct {
	CommandID string         `json:"command_id"`
	DeviceID  string         `json:"device_id"`
	Type      string         `json:"type" validate:"required,oneof=ping reboot immobilize config_update"`
	Params    map[string]any `json:"params,omitempty"`
	Status    string         `json:"status"` // pending | executed | expired
	IssuedAt  time.Time      `json:"issued_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	Result    *string        `json:"result,omitempty"` // ack payload from device
}

// SendCommandRequest is the HTTP request body for the command endpoint.
type SendCommandRequest struct {
	Type   string         `json:"type" validate:"required,oneof=ping reboot immobilize config_update"`
	Params map[string]any `json:"params,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Query helpers
// ─────────────────────────────────────────────────────────────────────────────

type HistoryQuery struct {
	VehicleID uuid.UUID
	From      time.Time
	To        time.Time
	Page      int
	PageSize  int
}

type NearbyQuery struct {
	Lat     float64
	Lng     float64
	RadiusM float64 // metres
}
