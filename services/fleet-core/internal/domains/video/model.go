package video

import (
	"time"

	"github.com/google/uuid"
)

// ─── Enums ───────────────────────────────────────────────────────────────────

type VideoType string

const (
	VideoTypeContinuous  VideoType = "continuous"
	VideoTypeEvent       VideoType = "event"
	VideoTypeManual      VideoType = "manual"
	VideoTypeLiveCapture VideoType = "live_capture"
)

type VideoStatus string

const (
	VideoStatusUploading  VideoStatus = "uploading"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusReady      VideoStatus = "ready"
	VideoStatusFailed     VideoStatus = "failed"
	VideoStatusDeleted    VideoStatus = "deleted"
)

type ProcessingStatus string

const (
	ProcessingQueued      ProcessingStatus = "queued"
	ProcessingTranscoding ProcessingStatus = "transcoding"
	ProcessingThumbnail   ProcessingStatus = "thumbnailing"
	ProcessingAnalyzing   ProcessingStatus = "analyzing"
	ProcessingDone        ProcessingStatus = "done"
	ProcessingFailed      ProcessingStatus = "failed"
)

type CameraPosition string

const (
	CameraFront     CameraPosition = "front"
	CameraDriver    CameraPosition = "driver"
	CameraRear      CameraPosition = "rear"
	CameraCargo     CameraPosition = "cargo"
	CameraSideLeft  CameraPosition = "side_left"
	CameraSideRight CameraPosition = "side_right"
	CameraPanoramic CameraPosition = "panoramic"
)

type EventType string

const (
	EventCollision          EventType = "collision"
	EventHarshBraking       EventType = "harsh_braking"
	EventHarshAcceleration  EventType = "harsh_acceleration"
	EventOverspeed          EventType = "overspeed"
	EventLaneDeparture      EventType = "lane_departure"
	EventTailgating         EventType = "tailgating"
	EventPanicButton        EventType = "panic_button"
	EventImpactSensor       EventType = "impact_sensor"
	EventPhoneUsage         EventType = "phone_usage"
	EventDriverDistraction  EventType = "driver_distraction"
	EventSeatbeltViolation  EventType = "seatbelt_violation"
	EventSmoking            EventType = "smoking"
	EventDrowsiness         EventType = "drowsiness"
	// Camera integrity events
	EventCameraTamper   EventType = "camera_tamper"
	EventCameraOffline  EventType = "camera_offline"
	EventCameraBlocked  EventType = "camera_blocked"  // lens physically blocked
	EventCameraRemoved  EventType = "camera_removed"  // unit physically removed
	EventPowerDisconnect EventType = "power_disconnect" // power cable removed
	EventLensCovered    EventType = "lens_covered"    // lens obscured with material
)

// ─── Domain Structs ───────────────────────────────────────────────────────────

type Camera struct {
	ID              uuid.UUID      `json:"id"`
	TenantID        uuid.UUID      `json:"tenant_id"`
	VehicleID       uuid.UUID      `json:"vehicle_id"`
	Position        CameraPosition `json:"camera_position"`
	SerialNumber    string         `json:"serial_number"`
	FirmwareVersion string         `json:"firmware_version"`
	Status          string         `json:"status"`
	InstalledAt     *time.Time     `json:"installed_at,omitempty"`
	LastHeartbeat   *time.Time     `json:"last_heartbeat,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Video struct {
	ID                  uuid.UUID   `json:"id"`
	TenantID            uuid.UUID   `json:"tenant_id"`
	VehicleID           uuid.UUID   `json:"vehicle_id"`
	CameraID            *uuid.UUID  `json:"camera_id,omitempty"`
	DriverID            *uuid.UUID  `json:"driver_id,omitempty"`
	TripID              *uuid.UUID  `json:"trip_id,omitempty"`
	IncidentID          *uuid.UUID  `json:"incident_id,omitempty"`   // links related clips from same incident
	CameraGroupID       *uuid.UUID  `json:"camera_group_id,omitempty"` // groups multi-camera sync
	Type                VideoType   `json:"video_type"`
	Status              VideoStatus `json:"status"`
	FilePath            string      `json:"file_path"`
	HLSPlaylistPath     string      `json:"hls_playlist_path,omitempty"`
	Codec               string      `json:"codec,omitempty"`
	Resolution          string      `json:"resolution,omitempty"`
	FileSizeBytes       int64       `json:"file_size_bytes,omitempty"`
	DurationSeconds     int         `json:"duration_seconds,omitempty"`
	StartTime           *time.Time  `json:"start_time,omitempty"`
	EndTime             *time.Time  `json:"end_time,omitempty"`
	EvidenceLocked      bool        `json:"evidence_locked"`
	LegalHold           bool        `json:"legal_hold"`
	CaseID              *uuid.UUID  `json:"case_id,omitempty"`
	RetentionOverride   *int        `json:"retention_override_days,omitempty"`
	IntegrityHash       string      `json:"integrity_hash,omitempty"`
	IntegritySignature  string      `json:"integrity_signature,omitempty"`
	DeviceCertSerial    string      `json:"device_cert_serial,omitempty"`
	PrivacyBlurred      bool        `json:"privacy_blurred"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}

type VideoEvent struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	VehicleID       uuid.UUID `json:"vehicle_id"`
	VideoID         *uuid.UUID `json:"video_id,omitempty"`
	EventType       EventType `json:"event_type"`
	OccurredAt      time.Time `json:"occurred_at"`
	Severity        int       `json:"severity"`
	ConfidenceScore float64   `json:"confidence_score"`
	OffsetSeconds   int       `json:"offset_seconds"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type VideoThumbnail struct {
	ID            uuid.UUID `json:"id"`
	VideoID       uuid.UUID `json:"video_id"`
	OffsetSeconds int       `json:"offset_seconds"`
	ThumbnailPath string    `json:"thumbnail_path"`
	CreatedAt     time.Time `json:"created_at"`
}

type VideoProcessingJob struct {
	ID            uuid.UUID        `json:"id"`
	VideoID       uuid.UUID        `json:"video_id"`
	Status        ProcessingStatus `json:"status"`
	WorkerID      string           `json:"worker_id,omitempty"`
	Attempts      int              `json:"attempts"`
	RetryCount    int              `json:"retry_count"`
	MaxRetries    int              `json:"max_retries"`
	RetryAfter    *time.Time       `json:"retry_after,omitempty"`
	LastError     string           `json:"last_error,omitempty"`
	LastHeartbeat *time.Time       `json:"last_heartbeat,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// MultipartSession tracks a chunked upload flow for large video files.
type MultipartSession struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	VehicleID    uuid.UUID              `json:"vehicle_id"`
	StorageKey   string                 `json:"storage_key"`
	UploadID     string                 `json:"upload_id"`
	TotalParts   *int                   `json:"total_parts,omitempty"`
	PartsReceived int                   `json:"parts_received"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ExpiresAt    time.Time              `json:"expires_at"`
	CreatedAt    time.Time              `json:"created_at"`
}

// MultipartPart is a single received chunk.
type MultipartPart struct {
	ID          uuid.UUID `json:"id"`
	SessionID   uuid.UUID `json:"session_id"`
	PartNumber  int       `json:"part_number"`
	SizeBytes   int64     `json:"size_bytes"`
	ETag        string    `json:"etag,omitempty"`
	ReceivedAt  time.Time `json:"received_at"`
}

// ─── Request/Response Types ───────────────────────────────────────────────────

type UploadVideoRequest struct {
	VehicleID       uuid.UUID  `json:"vehicle_id"`
	CameraID        *uuid.UUID `json:"camera_id,omitempty"`
	DriverID        *uuid.UUID `json:"driver_id,omitempty"`
	TripID          *uuid.UUID `json:"trip_id,omitempty"`
	Type            VideoType  `json:"video_type"`
	Codec           string     `json:"codec"`
	Resolution      string     `json:"resolution"`
	DurationSeconds int        `json:"duration_seconds"`
	StartTime       *time.Time `json:"start_time,omitempty"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	DeviceCertSerial string    `json:"device_cert_serial,omitempty"`
}

type LockEvidenceRequest struct {
	CaseID    uuid.UUID `json:"case_id"`
	LegalHold bool      `json:"legal_hold"`
}

type RequestClipRequest struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	CameraID  uuid.UUID `json:"camera_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

type ListVideosFilter struct {
	VehicleID *uuid.UUID
	DriverID  *uuid.UUID
	CameraID  *uuid.UUID
	TripID    *uuid.UUID
	EventType *EventType
	VideoType *VideoType
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}

type StreamResponse struct {
	// Direct MP4 signed URL
	SignedURL string `json:"signed_url,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
	// HLS adaptive streaming
	HLSPlaylistURL string `json:"hls_url,omitempty"`
	// Multi-camera sync metadata
	SyncedStreams []SyncedStream `json:"synced_streams,omitempty"`
}

type SyncedStream struct {
	CameraID       uuid.UUID `json:"camera_id"`
	Position       string    `json:"position"`
	HLSPlaylistURL string    `json:"hls_url"`
	TimestampOffset float64  `json:"timestamp_offset_seconds"`
}
