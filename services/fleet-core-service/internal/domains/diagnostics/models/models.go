package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a tracking device
type Device struct {
	ID               uuid.UUID  `json:"id"`
	OrgID            uuid.UUID  `json:"org_id"`
	VehicleID        *uuid.UUID `json:"vehicle_id,omitempty"`
	SerialNumber     string     `json:"serial_number"`
	Type             string     `json:"type"`
	Model            *string    `json:"model,omitempty"`
	Manufacturer     *string    `json:"manufacturer,omitempty"`
	FirmwareVersion  *string    `json:"firmware_version,omitempty"`
	HardwareRevision *string    `json:"hardware_revision,omitempty"`
	BatteryLevel     *int       `json:"battery_level,omitempty"`
	SignalStrength   *int       `json:"signal_strength,omitempty"`
	Status           string     `json:"status"`
	LastPingAt       *time.Time `json:"last_ping_at,omitempty"`
	LastHeartbeat    *time.Time `json:"last_heartbeat,omitempty"`
	IsDeleted        bool       `json:"is_deleted"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// DeviceSummary for dashboard stats
type DeviceSummary struct {
	TotalDevices     int `json:"total_devices"`
	OnlineDevices    int `json:"online_devices"`
	OfflineDevices   int `json:"offline_devices"`
	LowBattery       int `json:"low_battery"`
	WeakSignal       int `json:"weak_signal"`
	SystemErrors     int `json:"system_errors"`
	NewDevicesWeek   int `json:"new_devices_week"`
}

// DeviceTelemetry represents raw packet data
type DeviceTelemetry struct {
	ID         uuid.UUID              `json:"id"`
	DeviceID   uuid.UUID              `json:"device_id"`
	OrgID      uuid.UUID              `json:"org_id"`
	Payload    map[string]interface{} `json:"payload"`
	PacketType string                 `json:"packet_type"`
	RawHex     *string                `json:"raw_hex,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// DeviceCommand represents a remote command
type DeviceCommand struct {
	ID           uuid.UUID              `json:"id"`
	DeviceID     uuid.UUID              `json:"device_id"`
	OrgID        uuid.UUID              `json:"org_id"`
	Command      string                 `json:"command"`
	Category     string                 `json:"category"`
	Params       map[string]interface{} `json:"params,omitempty"`
	Status       string                 `json:"status"`
	Result       map[string]interface{} `json:"result,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	CreatedBy    *uuid.UUID             `json:"created_by,omitempty"`
	SentAt       *time.Time             `json:"sent_at,omitempty"`
	ExecutedAt   *time.Time             `json:"executed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// FirmwarePackage represents a firmware version
type FirmwarePackage struct {
	ID             uuid.UUID  `json:"id"`
	Version        string     `json:"version"`
	HardwareTarget string     `json:"hardware_target"`
	FileURL        string     `json:"file_url"`
	FileSize       *int       `json:"file_size,omitempty"`
	Checksum       *string    `json:"checksum,omitempty"`
	ReleaseNotes   *string    `json:"release_notes,omitempty"`
	Status         string     `json:"status"`
	IsCritical     bool       `json:"is_critical"`
	ReleasedAt     *time.Time `json:"released_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// FirmwareRollout represents a deployment campaign
type FirmwareRollout struct {
	ID               uuid.UUID              `json:"id"`
	FirmwareID       uuid.UUID              `json:"firmware_id"`
	OrgID            uuid.UUID              `json:"org_id"`
	Name             *string                `json:"name,omitempty"`
	DeviceGroup      *string                `json:"device_group,omitempty"`
	TargetFilter     map[string]interface{} `json:"target_filter,omitempty"`
	TotalDevices     int                    `json:"total_devices"`
	CompletedDevices int                    `json:"completed_devices"`
	FailedDevices    int                    `json:"failed_devices"`
	Status           string                 `json:"status"`
	RolloutStrategy  string                 `json:"rollout_strategy"`
	MaxConcurrent    int                    `json:"max_concurrent"`
	CreatedBy        *uuid.UUID             `json:"created_by,omitempty"`
	StartedAt        *time.Time             `json:"started_at,omitempty"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// DiagnosticsLog represents a device log entry
type DiagnosticsLog struct {
	ID        uuid.UUID              `json:"id"`
	DeviceID  uuid.UUID              `json:"device_id"`
	OrgID     uuid.UUID              `json:"org_id"`
	LogType   string                 `json:"log_type"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SensorReading represents current sensor values
type SensorReading struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value"`
	Unit   string      `json:"unit,omitempty"`
	Status string      `json:"status"` // normal, warning, critical
}
