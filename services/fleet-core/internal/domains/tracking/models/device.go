package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a registered vehicle device
type Device struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	DeviceID            string     `json:"device_id" db:"device_id"` // Hardware device identifier
	TenantID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	VehicleID           uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	HardwareAttestation string     `json:"hardware_attestation" db:"hardware_attestation"`
	AndroidVersion      string     `json:"android_version" db:"android_version"`
	DeviceModel         string     `json:"device_model" db:"device_model"`
	AppVersion          string     `json:"app_version" db:"app_version"`
	Status              string     `json:"status" db:"status"` // active, inactive, suspended
	JWTToken            string     `json:"-" db:"jwt_token"`   // Not returned in responses
	RefreshToken        string     `json:"-" db:"refresh_token"`
	TokenExpiresAt      time.Time  `json:"token_expires_at" db:"token_expires_at"`
	LastHeartbeat       *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	CreatedBy           uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy           uuid.UUID  `json:"updated_by" db:"updated_by"`
}

// DeviceRegisterRequest is the request payload for device registration
type DeviceRegisterRequest struct {
	DeviceID            string `json:"device_id" validate:"required"`
	HardwareAttestation string `json:"hardware_attestation" validate:"required"`
	AndroidVersion      string `json:"android_version" validate:"required"`
	DeviceModel         string `json:"device_model" validate:"required"`
	AppVersion          string `json:"app_version"`
	TenantID            string `json:"tenant_id"` // Optional, can be assigned later
	VehicleID           string `json:"vehicle_id"`      // Optional, can be assigned later
}

// DeviceRegisterResponse is the response payload for device registration
type DeviceRegisterResponse struct {
	DeviceID     string `json:"device_id"`
	JWTToken     string `json:"jwt_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp
}

// HeartbeatRequest is the payload for device heartbeat
type HeartbeatRequest struct {
	DeviceID string `json:"device_id" validate:"required"`
	Status   string `json:"status" validate:"required"` // ok, degraded, critical
	Uptime   int64  `json:"uptime"`                     // Seconds since boot

	// Module health status
	ModuleHealth ModuleHealth `json:"module_health" validate:"required"`

	// Telemetry summary since last heartbeat
	TelemetrySummary TelemetrySummary `json:"telemetry_summary"`
}

// ModuleHealth represents the health status of device modules
type ModuleHealth struct {
	Camera  ModuleStatus `json:"camera"`
	AI      ModuleStatus `json:"ai"`
	GPS     ModuleStatus `json:"gps"`
	CANBus  ModuleStatus `json:"can_bus"`
	Storage ModuleStatus `json:"storage"`
	Network ModuleStatus `json:"network"`
}

// ModuleStatus represents the status of a single module
type ModuleStatus struct {
	Status      string  `json:"status"`       // ok, degraded, failed
	Temperature float64 `json:"temperature"`  // Celsius
	MemoryUsage float64 `json:"memory_usage"` // Percentage 0-100
	ErrorCount  int     `json:"error_count"`
}

// TelemetrySummary summarizes telemetry data since last heartbeat
type TelemetrySummary struct {
	GPSPointsCount      int     `json:"gps_points_count"`
	CANBusMessagesCount int     `json:"can_bus_messages_count"`
	AverageSpeedKph     float64 `json:"average_speed_kph"`
	MaxSpeedKph         float64 `json:"max_speed_kph"`
	DistanceKm          float64 `json:"distance_km"`
}

// HeartbeatResponse is the response payload for device heartbeat
type HeartbeatResponse struct {
	Status          string    `json:"status"` // ok, config_update_available
	ConfigVersion   string    `json:"config_version"`
	PendingCommands []Command `json:"pending_commands"`
}

// ConfigResponse is the response payload for device config fetch
type ConfigResponse struct {
	Config    DeviceRuntimeConfig `json:"config"`
	Signature string              `json:"signature"` // Cryptographic signature for verification
	Version   string              `json:"version"`
	ExpiresAt int64               `json:"expires_at"` // Unix timestamp
}

// DeviceRuntimeConfig is the runtime configuration for device
type DeviceRuntimeConfig struct {
	SchemaVersion     string          `json:"schema_version"`
	HeartbeatSeconds  int             `json:"heartbeat_seconds"`
	RingBufferMinutes int             `json:"ring_buffer_minutes"`
	GPS               GPSConfig       `json:"gps"`
	AI                AIConfig        `json:"ai"`
	Retention         RetentionConfig `json:"retention"`
	FeatureFlags      FeatureFlags    `json:"feature_flags"`
}

// GPSConfig represents GPS configuration
type GPSConfig struct {
	ActiveIntervalSeconds int `json:"active_interval_seconds"`
	IdleIntervalSeconds   int `json:"idle_interval_seconds"`
}

// AIConfig represents AI configuration
type AIConfig struct {
	Enabled            bool               `json:"enabled"`
	FrameSampleStride  int                `json:"frame_sample_stride"`
	SeverityThresholds SeverityThresholds `json:"severity_thresholds"`
}

// SeverityThresholds represents AI severity thresholds
type SeverityThresholds struct {
	WarnMin     float64 `json:"warn_min"`
	CriticalMin float64 `json:"critical_min"`
}

// RetentionConfig represents retention configuration
type RetentionConfig struct {
	CriticalClipDays    int  `json:"critical_clip_days"`
	MinStorageReserveMB int  `json:"min_storage_reserve_mb"`
	LegalHold           bool `json:"legal_hold"`
}

// FeatureFlags represents feature flags
type FeatureFlags struct {
	CameraEnabled           bool `json:"camera_enabled"`
	AudioEventChecksEnabled bool `json:"audio_event_checks_enabled"`
	AIEnabled               bool `json:"ai_enabled"`
	CANBusEnabled           bool `json:"can_bus_enabled"`
	DriverLockoutEnabled    bool `json:"driver_lockout_enabled"`
}

// TelemetryBatchRequest is the payload for batch telemetry upload
type TelemetryBatchRequest struct {
	DeviceID  string           `json:"device_id" validate:"required"`
	BatchID   string           `json:"batch_id" validate:"required"`
	Telemetry []TelemetryPoint `json:"telemetry" validate:"required,dive"`
	Timestamp int64            `json:"timestamp"` // Batch timestamp
}

// TelemetryPoint represents a single telemetry data point
type TelemetryPoint struct {
	Timestamp int64       `json:"timestamp"` // Unix milliseconds
	GPS       GPSData     `json:"gps"`
	CANBus    *CANBusData `json:"can_bus,omitempty"`
	SpeedKph  float64     `json:"speed_kph"`
	Heading   float64     `json:"heading"`
	Altitude  float64     `json:"altitude"`
	Accuracy  float64     `json:"accuracy"`
}

// GPSData represents GPS coordinates
type GPSData struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Accuracy   float64 `json:"accuracy"`
	Satellites int     `json:"satellites"`
}

// CANBusData represents CAN bus telemetry
type CANBusData struct {
	EngineRPM    int     `json:"engine_rpm"`
	SpeedKph     float64 `json:"speed_kph"`
	BrakePressed bool    `json:"brake_pressed"`
	TurnSignalOn bool    `json:"turn_signal_on"`
	FuelLevel    float64 `json:"fuel_level"`
	CoolantTemp  float64 `json:"coolant_temp"`
	DTCCount     int     `json:"dtc_count"`
	OdometerKm   float64 `json:"odometer_km"`
}

// TelemetryResponse is the response payload for telemetry batch upload
type TelemetryResponse struct {
	Status        string `json:"status"` // accepted, rejected
	ReceivedCount int    `json:"received_count"`
	NextBatchID   string `json:"next_batch_id"`
}

// UploadUrlRequest is the request for getting a pre-signed upload URL
// For local-first storage, this endpoint registers the clip metadata
type UploadUrlRequest struct {
	DeviceID        string `json:"device_id" validate:"required"`
	ClipID          string `json:"clip_id" validate:"required"`
	SizeBytes       int64  `json:"size_bytes" validate:"required"`
	ContentType     string `json:"content_type" validate:"required"`
	DurationSeconds int    `json:"duration_seconds"`
	EventType       string `json:"event_type"`   // Event that triggered clip (optional)
	StorageMode     string `json:"storage_mode"` // "local" (default) or "cloud"
}

// UploadUrlResponse is the response with storage location
// For local storage, returns local path instead of cloud URL
type UploadUrlResponse struct {
	UploadURL      string `json:"upload_url,omitempty"` // Cloud URL (if cloud storage)
	LocalPath      string `json:"local_path,omitempty"` // Local storage path (if local storage)
	ExpiresAt      int64  `json:"expires_at,omitempty"` // Unix timestamp
	ObjectKey      string `json:"object_key"`
	StorageMode    string `json:"storage_mode"`      // "local" or "cloud"
	MaxLocalSizeMB int64  `json:"max_local_size_mb"` // Max local storage quota
}

// ConfirmUploadRequest is the request to confirm upload completion
type ConfirmUploadRequest struct {
	ClipID          string `json:"clip_id" validate:"required"`
	ChecksumSHA256  string `json:"checksum_sha256" validate:"required"`
	DurationSeconds int    `json:"duration_seconds"`
	UploadedBytes   int64  `json:"uploaded_bytes"`
	StorageMode     string `json:"storage_mode"`         // "local" or "cloud"
	LocalPath       string `json:"local_path,omitempty"` // Path on device (if local storage)
}

// ConfirmUploadResponse is the response for upload confirmation
type ConfirmUploadResponse struct {
	Status          string `json:"status"`           // confirmed, failed
	StorageLocation string `json:"storage_location"` // Local path or cloud URL
	StorageMode     string `json:"storage_mode"`     // "local" or "cloud"
	SyncStatus      string `json:"sync_status"`      // "pending_sync", "synced", "cloud_only"
}

// AttestationRequest is the request for evidence integrity attestation
type AttestationRequest struct {
	DeviceID     string `json:"device_id" validate:"required"`
	EvidenceHash string `json:"evidence_hash" validate:"required"`
	ChainHash    string `json:"chain_hash" validate:"required"`
	Timestamp    int64  `json:"timestamp"`
}

// AttestationResponse is the response with attestation
type AttestationResponse struct {
	AttestationID string `json:"attestation_id"`
	Signature     string `json:"signature"`
	Timestamp     int64  `json:"timestamp"`
	BlockchainTx  string `json:"blockchain_tx,omitempty"` // Optional blockchain transaction
}

// Command represents a remote command for the device
type Command struct {
	ID        string                 `json:"id"`
	Type      CommandType            `json:"type"`
	Params    map[string]interface{} `json:"params"`
	ExpiresAt int64                  `json:"expires_at"` // Unix timestamp
	Priority  int                    `json:"priority"`   // 1 = highest, 5 = lowest
	CreatedAt int64                  `json:"created_at"`
}

// CommandType represents the type of remote command
type CommandType string

const (
	CommandReboot         CommandType = "reboot"
	CommandFactoryReset   CommandType = "factory_reset"
	CommandUpdateConfig   CommandType = "update_config"
	CommandStartRecording CommandType = "start_recording"
	CommandStopRecording  CommandType = "stop_recording"
	CommandExtractClip    CommandType = "extract_clip"
	CommandDiagnostics    CommandType = "diagnostics"
	CommandWipeData       CommandType = "wipe_data"
)

// DriverLoginRequest is the request for driver login
type DriverLoginRequest struct {
	DeviceID   string `json:"device_id" validate:"required"`
	DriverID   string `json:"driver_id" validate:"required"`
	RFIDToken  string `json:"rfid_token"`
	PhotoHash  string `json:"photo_hash"`
	DutyStatus string `json:"duty_status"` // on_duty, driving, sleeper, off_duty
}

// DriverLoginResponse is the response for driver login
type DriverLoginResponse struct {
	Status        string         `json:"status"` // success, failed
	DriverProfile *DriverProfile `json:"driver_profile"`
	HOSStatus     *HOSStatus     `json:"hos_status"`
	ShiftStart    int64          `json:"shift_start"` // Unix timestamp
}

// DriverProfile represents driver profile information
type DriverProfile struct {
	DriverID      string    `json:"driver_id"`
	Name          string    `json:"name"`
	LicenseNumber string    `json:"license_number"`
	HOSLimit      HOSLimits `json:"hos_limits"`
}

// HOSLimits represents HOS regulatory limits
type HOSLimits struct {
	MaxDrivingHours    float64 `json:"max_driving_hours"`     // 11 hours (US)
	MaxOnDutyHours     float64 `json:"max_on_duty_hours"`     // 14 hours (US)
	MinBreakAfterHours float64 `json:"min_break_after_hours"` // 8 hours driving
	MinBreakDuration   float64 `json:"min_break_duration"`    // 30 minutes
	MaxOnDutyDays      int     `json:"max_on_duty_days"`      // 60/70 hours in 7/8 days
}

// HOSStatus represents current Hours of Service status
type HOSStatus struct {
	DutyStatus           string   `json:"duty_status"`            // on_duty, driving, sleeper, off_duty
	DrivingTimeRemaining float64  `json:"driving_time_remaining"` // Hours remaining
	OnDutyTimeRemaining  float64  `json:"on_duty_time_remaining"` // Hours remaining
	BreakRequired        bool     `json:"break_required"`
	Violations           []string `json:"violations"`
}

// DriverLogoutRequest is the request for driver logout
type DriverLogoutRequest struct {
	DeviceID      string  `json:"device_id" validate:"required"`
	DriverID      string  `json:"driver_id" validate:"required"`
	ShiftDuration float64 `json:"shift_duration"` // Hours
	EndDutyStatus string  `json:"end_duty_status"`
}

// DriverLogoutResponse is the response for driver logout
type DriverLogoutResponse struct {
	Status     string     `json:"status"` // success, failed
	HOSSummary HOSSummary `json:"hos_summary"`
	Violations []string   `json:"violations"`
}

// HOSSummary summarizes the driver's HOS for the shift
type HOSSummary struct {
	ShiftStart   int64   `json:"shift_start"`
	ShiftEnd     int64   `json:"shift_end"`
	TotalDriving float64 `json:"total_driving"` // Hours
	TotalOnDuty  float64 `json:"total_on_duty"` // Hours
	TotalBreaks  float64 `json:"total_breaks"`  // Hours
	DistanceKm   float64 `json:"distance_km"`
}

// HOSUpdateRequest is the request for HOS duty status update
type HOSUpdateRequest struct {
	DeviceID   string    `json:"device_id" validate:"required"`
	DriverID   string    `json:"driver_id" validate:"required"`
	DutyStatus string    `json:"duty_status" validate:"required"`
	Timestamp  int64     `json:"timestamp"`
	Location   *Location `json:"location,omitempty"`
}

// Location represents geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// HOSUpdateResponse is the response for HOS update
type HOSUpdateResponse struct {
	Status              string   `json:"status"` // accepted, rejected
	Violations          []string `json:"violations"`
	RemainingDriveTime  float64  `json:"remaining_drive_time"`   // Hours
	RemainingOnDutyTime float64  `json:"remaining_on_duty_time"` // Hours
}
