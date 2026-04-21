package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Device represents an IoT tracking device assigned to a vehicle
type Device struct {
	ID              uuid.UUID  `json:"id" db:"id" validate:"required"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id" validate:"required"`
	VehicleID       *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	DeviceIMEI      *string    `json:"device_imei,omitempty" db:"device_imei" validate:"omitempty"`
	DeviceSerial    *string    `json:"device_serial,omitempty" db:"device_serial"`
	DeviceModel     *string    `json:"device_model,omitempty" db:"device_model"`
	FirmwareVersion *string    `json:"firmware_version,omitempty" db:"firmware_version"`
	HardwareVersion *string    `json:"hardware_version,omitempty" db:"hardware_version"`
	Status          *string    `json:"status,omitempty" db:"status" validate:"omitempty,oneof=ACTIVE INACTIVE MAINTENANCE ERROR DECOMMISSIONED"`
	LastSeenAt      *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	RegisteredAt    *time.Time `json:"registered_at,omitempty" db:"registered_at"`
}

// DeviceTelemetry represents raw telemetry data ingested from a device
type DeviceTelemetry struct {
	ID             int64           `json:"id" db:"id"`
	DeviceID       *uuid.UUID      `json:"device_id,omitempty" db:"device_id"`
	Timestamp      *time.Time      `json:"timestamp,omitempty" db:"timestamp"`
	GPSLat         *float64        `json:"gps_lat,omitempty" db:"gps_lat"`
	GPSLon         *float64        `json:"gps_lon,omitempty" db:"gps_lon"`
	Speed          *float64        `json:"speed,omitempty" db:"speed"`
	BatteryVoltage *float64        `json:"battery_voltage,omitempty" db:"battery_voltage"`
	SignalStrength *int            `json:"signal_strength,omitempty" db:"signal_strength"`
	EngineRPM      *int            `json:"engine_rpm,omitempty" db:"engine_rpm"`
	Temperature    *float64        `json:"temperature,omitempty" db:"temperature"`
	PayloadJSON    json.RawMessage `json:"payload_json,omitempty" db:"payload_json"`
}

// DeviceCommand represents a command dispatched to a device
type DeviceCommand struct {
	ID             uuid.UUID       `json:"id" db:"id" validate:"required"`
	DeviceID       *uuid.UUID      `json:"device_id,omitempty" db:"device_id"`
	CommandType    *string         `json:"command_type,omitempty" db:"command_type"`
	CommandPayload json.RawMessage `json:"command_payload,omitempty" db:"command_payload"`
	Status         *string         `json:"status,omitempty" db:"status" validate:"omitempty,oneof=CREATED QUEUED DELIVERED ACKED APPLIED FAILED"`
	CreatedAt      *time.Time      `json:"created_at,omitempty" db:"created_at"`
	DeliveredAt    *time.Time      `json:"delivered_at,omitempty" db:"delivered_at"`
	AcknowledgedAt *time.Time      `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	AppliedAt      *time.Time      `json:"applied_at,omitempty" db:"applied_at"`
}

// DeviceTwin represents the desired vs reported configuration state of a device
type DeviceTwin struct {
	DeviceID      uuid.UUID       `json:"device_id" db:"device_id" validate:"required"`
	DesiredState  json.RawMessage `json:"desired_state,omitempty" db:"desired_state"`
	ReportedState json.RawMessage `json:"reported_state,omitempty" db:"reported_state"`
	UpdatedAt     *time.Time      `json:"updated_at,omitempty" db:"updated_at"`
}

// DeviceDiagnosticsLog represents logs emitted by a device
type DeviceDiagnosticsLog struct {
	ID         int64      `json:"id" db:"id"`
	DeviceID   *uuid.UUID `json:"device_id,omitempty" db:"device_id"`
	LogLevel   *string    `json:"log_level,omitempty" db:"log_level"`
	LogMessage *string    `json:"log_message,omitempty" db:"log_message"`
	CreatedAt  *time.Time `json:"created_at,omitempty" db:"created_at"`
}

// FirmwarePackage represents an OTA firmware binary
type FirmwarePackage struct {
	ID              uuid.UUID  `json:"id" db:"id" validate:"required"`
	FirmwareVersion *string    `json:"firmware_version,omitempty" db:"firmware_version"`
	DeviceModel     *string    `json:"device_model,omitempty" db:"device_model"`
	Checksum        *string    `json:"checksum,omitempty" db:"checksum"`
	FileURL         *string    `json:"file_url,omitempty" db:"file_url"`
	SizeBytes       *int       `json:"size_bytes,omitempty" db:"size_bytes"`
	Status          *string    `json:"status,omitempty" db:"status" validate:"omitempty,oneof=DRAFT PUBLISHED DEPRECATED"`
	CreatedAt       *time.Time `json:"created_at,omitempty" db:"created_at"`
}

// FirmwareRollout represents a rollout campaign applying firmware to devices
type FirmwareRollout struct {
	ID                uuid.UUID  `json:"id" db:"id" validate:"required"`
	FirmwareID        *uuid.UUID `json:"firmware_id,omitempty" db:"firmware_id"`
	TenantID             *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	RolloutStrategy   *string    `json:"rollout_strategy,omitempty" db:"rollout_strategy"`
	TargetDeviceCount *int       `json:"target_device_count,omitempty" db:"target_device_count"`
	Status            *string    `json:"status,omitempty" db:"status" validate:"omitempty,oneof=SCHEDULED IN_PROGRESS COMPLETED FAILED CANCELLED"`

	// Safety Controls added in Phase 2
	FailureThreshold *int  `json:"failure_threshold,omitempty" db:"failure_threshold"`
	BatchSize        *int  `json:"batch_size,omitempty" db:"batch_size"`
	RollbackEnabled  *bool `json:"rollback_enabled,omitempty" db:"rollback_enabled"`

	ScheduledAt *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// RolloutDeviceStatus tracks firmware rollout per device
type RolloutDeviceStatus struct {
	ID        int64      `json:"id" db:"id"`
	RolloutID *uuid.UUID `json:"rollout_id,omitempty" db:"rollout_id"`
	DeviceID  *uuid.UUID `json:"device_id,omitempty" db:"device_id"`
	Status    *string    `json:"status,omitempty" db:"status" validate:"omitempty,oneof=PENDING UPDATING SUCCESS FAILED"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// FirmwareRolloutEvent represents the detailed progress events for a rollout
type FirmwareRolloutEvent struct {
	ID        int64           `json:"id" db:"id"`
	RolloutID *uuid.UUID      `json:"rollout_id,omitempty" db:"rollout_id"`
	DeviceID  *uuid.UUID      `json:"device_id,omitempty" db:"device_id"`
	EventType *string         `json:"event_type,omitempty" db:"event_type"`
	Metadata  json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt *time.Time      `json:"created_at,omitempty" db:"created_at"`
}

// DiagnosticsAuditLog represents the audit trail structure for the module
type DiagnosticsAuditLog struct {
	ID         int64           `json:"id" db:"id"`
	EntityType *string         `json:"entity_type,omitempty" db:"entity_type"`
	EntityID   *uuid.UUID      `json:"entity_id,omitempty" db:"entity_id"`
	Action     *string         `json:"action,omitempty" db:"action"`
	ActorID    *uuid.UUID      `json:"actor_id,omitempty" db:"actor_id"`
	Metadata   json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  *time.Time      `json:"created_at,omitempty" db:"created_at"`
}

// DeviceAlert represents an automated alert for a device problem
type DeviceAlert struct {
	ID           int64      `json:"id" db:"id"`
	DeviceID     *uuid.UUID `json:"device_id,omitempty" db:"device_id" validate:"required"`
	AlertType    *string    `json:"alert_type,omitempty" db:"alert_type" validate:"required"`
	Severity     *string    `json:"severity,omitempty" db:"severity" validate:"required"`
	AlertMessage *string    `json:"alert_message,omitempty" db:"alert_message"`
	CreatedAt    *time.Time `json:"created_at,omitempty" db:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}
