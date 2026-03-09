package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a tracking device
type Device struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrgID            uuid.UUID  `json:"org_id" db:"org_id" validate:"required"`
	VehicleID        *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	SerialNumber     string     `json:"serial_number" db:"serial_number" validate:"required,max=100"`
	Type             string     `json:"type" db:"type" validate:"required,oneof=GPS_TRACKER OBDII HARDWIRED"`
	Model            *string    `json:"model,omitempty" db:"model" validate:"omitempty,max=100"`
	Manufacturer     *string    `json:"manufacturer,omitempty" db:"manufacturer" validate:"omitempty,max=100"`
	FirmwareVersion  *string    `json:"firmware_version,omitempty" db:"firmware_version" validate:"omitempty,max=50"`
	HardwareRevision *string    `json:"hardware_revision,omitempty" db:"hardware_revision" validate:"omitempty,max=50"`
	BatteryLevel     *int       `json:"battery_level,omitempty" db:"battery_level" validate:"omitempty,gte=0,lte=100"`
	SignalStrength   *int       `json:"signal_strength,omitempty" db:"signal_strength"`
	Status           string     `json:"status" db:"status" validate:"required,oneof=active inactive error maintenance"`
	LastPingAt       *time.Time `json:"last_ping_at,omitempty" db:"last_ping_at"`
	LastHeartbeat    *time.Time `json:"last_heartbeat,omitempty" db:"last_heartbeat"`
	IsDeleted        bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// DeviceSummary for dashboard stats
type DeviceSummary struct {
	TotalDevices   int `json:"total_devices"`
	OnlineDevices  int `json:"online_devices"`
	OfflineDevices int `json:"offline_devices"`
	LowBattery     int `json:"low_battery"`
	WeakSignal     int `json:"weak_signal"`
	SystemErrors   int `json:"system_errors"`
	NewDevicesWeek int `json:"new_devices_week"`
}

// DeviceTelemetry represents raw packet data
type DeviceTelemetry struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	DeviceID   uuid.UUID              `json:"device_id" db:"device_id" validate:"required"`
	OrgID      uuid.UUID              `json:"org_id" db:"org_id" validate:"required"`
	Payload    map[string]interface{} `json:"payload" db:"payload" validate:"required"`
	PacketType string                 `json:"packet_type" db:"packet_type" validate:"required,max=50"`
	RawHex     *string                `json:"raw_hex,omitempty" db:"raw_hex"`
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
}

// DeviceCommand represents a remote command
type DeviceCommand struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	DeviceID     uuid.UUID              `json:"device_id" db:"device_id" validate:"required"`
	OrgID        uuid.UUID              `json:"org_id" db:"org_id" validate:"required"`
	Command      string                 `json:"command" db:"command" validate:"required,max=100"`
	Category     string                 `json:"category" db:"category" validate:"required,oneof=CONFIG DIAGNOSTIC CONTROL"`
	Params       map[string]interface{} `json:"params,omitempty" db:"params"`
	Status       string                 `json:"status" db:"status" validate:"required,oneof=created queued delivered acked applied failed timed_out"`
	Result       map[string]interface{} `json:"result,omitempty" db:"result"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	CreatedBy    *uuid.UUID             `json:"created_by,omitempty" db:"created_by"`
	SentAt       *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	ExecutedAt   *time.Time             `json:"executed_at,omitempty" db:"executed_at"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// DeviceTwin represents desired vs reported state for reliability tracking.
type DeviceTwin struct {
	DeviceID          uuid.UUID              `json:"device_id" db:"device_id" validate:"required"`
	OrgID             uuid.UUID              `json:"org_id" db:"org_id" validate:"required"`
	DesiredState      map[string]interface{} `json:"desired_state,omitempty" db:"desired_state"`
	ReportedState     map[string]interface{} `json:"reported_state,omitempty" db:"reported_state"`
	LastSeen          *time.Time             `json:"last_seen,omitempty" db:"last_seen"`
	DriftStatus       string                 `json:"drift_status" db:"drift_status" validate:"required,oneof=unknown in_sync drifted"`
	DriftReason       *string                `json:"drift_reason,omitempty" db:"drift_reason"`
	LastCommandID     *uuid.UUID             `json:"last_command_id,omitempty" db:"last_command_id"`
	LastCommandStatus *string                `json:"last_command_status,omitempty" db:"last_command_status"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// FirmwarePackage represents a firmware version
type FirmwarePackage struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Version        string     `json:"version" db:"version" validate:"required,max=50"`
	HardwareTarget string     `json:"hardware_target" db:"hardware_target" validate:"required,max=100"`
	FileURL        string     `json:"file_url" db:"file_url" validate:"required,url"`
	FileSize       *int       `json:"file_size,omitempty" db:"file_size" validate:"omitempty,gt=0"`
	Checksum       *string    `json:"checksum,omitempty" db:"checksum" validate:"omitempty,max=100"`
	ReleaseNotes   *string    `json:"release_notes,omitempty" db:"release_notes"`
	Status         string     `json:"status" db:"status" validate:"required,oneof=draft published deprecated"`
	IsCritical     bool       `json:"is_critical" db:"is_critical"`
	ReleasedAt     *time.Time `json:"released_at,omitempty" db:"released_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// FirmwareRollout represents a deployment campaign
type FirmwareRollout struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	FirmwareID       uuid.UUID              `json:"firmware_id" db:"firmware_id" validate:"required"`
	OrgID            uuid.UUID              `json:"org_id" db:"org_id" validate:"required"`
	Name             *string                `json:"name,omitempty" db:"name" validate:"required,max=255"`
	DeviceGroup      *string                `json:"device_group,omitempty" db:"device_group" validate:"omitempty,max=100"`
	TargetFilter     map[string]interface{} `json:"target_filter,omitempty" db:"target_filter"`
	TotalDevices     int                    `json:"total_devices" db:"total_devices" validate:"required,gt=0"`
	CompletedDevices int                    `json:"completed_devices" db:"completed_devices" validate:"gte=0"`
	FailedDevices    int                    `json:"failed_devices" db:"failed_devices" validate:"gte=0"`
	Status           string                 `json:"status" db:"status" validate:"required,oneof=scheduled in_progress paused completed cancelled"`
	RolloutStrategy  string                 `json:"rollout_strategy" db:"rollout_strategy" validate:"required,oneof=gradual canary full"`
	MaxConcurrent    int                    `json:"max_concurrent" db:"max_concurrent" validate:"required,gt=0"`
	CreatedBy        *uuid.UUID             `json:"created_by,omitempty" db:"created_by"`
	StartedAt        *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// FirmwareRolloutEvent captures immutable rollout lifecycle/audit events.
type FirmwareRolloutEvent struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	RolloutID   uuid.UUID              `json:"rollout_id" db:"rollout_id" validate:"required"`
	OrgID       uuid.UUID              `json:"org_id" db:"org_id" validate:"required"`
	EventType   string                 `json:"event_type" db:"event_type" validate:"required,oneof=rollout.note rollout.transition rollout.created"`
	FromStatus  *string                `json:"from_status,omitempty" db:"from_status"`
	ToStatus    *string                `json:"to_status,omitempty" db:"to_status"`
	ActorUserID *uuid.UUID             `json:"actor_user_id,omitempty" db:"actor_user_id"`
	Message     string                 `json:"message" db:"message" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// DiagnosticsLog represents a device log entry
type DiagnosticsLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	DeviceID  uuid.UUID              `json:"device_id" db:"device_id" validate:"required"`
	OrgID     uuid.UUID              `json:"org_id" db:"org_id" validate:"required"`
	LogType   string                 `json:"log_type" db:"log_type" validate:"required,oneof=info warning error critical"`
	Source    string                 `json:"source" db:"source" validate:"required,max=100"`
	Message   string                 `json:"message" db:"message" validate:"required"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// SensorReading represents current sensor values
type SensorReading struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value"`
	Unit   string      `json:"unit,omitempty"`
	Status string      `json:"status"` // normal, warning, critical
}

// DeviceStatus defines valid device statuses
type DeviceStatus string

const (
	StatusActive      DeviceStatus = "active"
	StatusInactive    DeviceStatus = "inactive"
	StatusError       DeviceStatus = "error"
	StatusMaintenance DeviceStatus = "maintenance"
)

// CommandStatus defines valid command statuses
type CommandStatus string

const (
	CommandCreated   CommandStatus = "created"
	CommandQueued    CommandStatus = "queued"
	CommandDelivered CommandStatus = "delivered"
	CommandAcked     CommandStatus = "acked"
	CommandApplied   CommandStatus = "applied"
	CommandFailed    CommandStatus = "failed"
	CommandTimedOut  CommandStatus = "timed_out"
)

// FirmwareStatus defines valid firmware statuses
type FirmwareStatus string

const (
	FirmwareDraft     FirmwareStatus = "draft"
	FirmwarePublished FirmwareStatus = "published"
	FirmwareDeprecated FirmwareStatus = "deprecated"
)

// RolloutStatus defines valid rollout statuses
type RolloutStatus string

const (
	RolloutScheduled   RolloutStatus = "scheduled"
	RolloutInProgress  RolloutStatus = "in_progress"
	RolloutPaused      RolloutStatus = "paused"
	RolloutCompleted   RolloutStatus = "completed"
	RolloutCancelled   RolloutStatus = "cancelled"
)

// IsValidCommandTransition checks if a command status transition is valid
func IsValidCommandTransition(from, to CommandStatus) bool {
	validTransitions := map[CommandStatus][]CommandStatus{
		CommandCreated:   {CommandQueued},
		CommandQueued:    {CommandDelivered, CommandFailed},
		CommandDelivered: {CommandAcked, CommandTimedOut},
		CommandAcked:     {CommandApplied, CommandFailed},
		CommandApplied:   {}, // Terminal state
		CommandFailed:    {}, // Terminal state
		CommandTimedOut:  {}, // Terminal state
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// IsValidRolloutTransition checks if a rollout status transition is valid
func IsValidRolloutTransition(from, to RolloutStatus) bool {
	validTransitions := map[RolloutStatus][]RolloutStatus{
		RolloutScheduled:  {RolloutInProgress, RolloutCancelled},
		RolloutInProgress: {RolloutPaused, RolloutCompleted, RolloutCancelled},
		RolloutPaused:     {RolloutInProgress, RolloutCancelled},
		RolloutCompleted:  {}, // Terminal state
		RolloutCancelled:  {}, // Terminal state
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}
