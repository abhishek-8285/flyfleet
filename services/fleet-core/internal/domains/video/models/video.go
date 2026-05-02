package models

import (
	"time"

	"github.com/google/uuid"
)

// VideoStream represents a live video stream from a vehicle
type VideoStream struct {
	ID               string     `json:"id"`
	CameraID         string     `json:"camera_id"`
	VehicleID        uuid.UUID  `json:"vehicle_id"`
	DriverID         uuid.UUID  `json:"driver_id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	URL              string     `json:"url"`
	Status           string     `json:"status"`     // active, inactive, error
	Resolution       string     `json:"resolution"` // e.g., "1080p", "720p"
	Framerate        int        `json:"framerate"`  // frames per second
	Bitrate          int        `json:"bitrate"`    // kbps
	Protocol         string     `json:"protocol"`   // RTMP, HLS, WebRTC
	Location         Location   `json:"location"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	AnalyticsEnabled bool       `json:"analytics_enabled"`
}

// VideoRecording represents a recorded video clip
type VideoRecording struct {
	ID             string    `json:"id"`
	StreamID       string    `json:"stream_id"`
	CameraID       string    `json:"camera_id"`
	VehicleID      uuid.UUID `json:"vehicle_id"`
	DriverID       uuid.UUID `json:"driver_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	Filename       string    `json:"filename"`
	Filepath       string    `json:"filepath"`
	FileSize       int64     `json:"file_size"` // in bytes
	Duration       int       `json:"duration"`  // in seconds
	Resolution     string    `json:"resolution"`
	Framerate      int       `json:"framerate"`
	Bitrate        int       `json:"bitrate"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	LocationStart  Location  `json:"location_start"`
	LocationEnd    Location  `json:"location_end"`
	Status         string    `json:"status"` // recorded, processing, available, archived, deleted
	ThumbnailPath  *string   `json:"thumbnail_path,omitempty"`
	Tags           []string  `json:"tags"`
	EventType      *string   `json:"event_type,omitempty"` // sudden_stop, harsh_acceleration, accident, etc.
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	PlaybackURL    string    `json:"playback_url,omitempty"`
}

// VideoAnalytics represents AI-powered insights from video analysis
type VideoAnalytics struct {
	ID              string            `json:"id"`
	RecordingID     string            `json:"recording_id"`
	StreamID        string            `json:"stream_id"`
	ObjectsDetected []ObjectDetection `json:"objects_detected"`
	BehaviorEvents  []BehaviorEvent   `json:"behavior_events"`
	TrafficSigns    []TrafficSign     `json:"traffic_signs"`
	Incidents       []Incident        `json:"incidents"`
	Timestamp       time.Time         `json:"timestamp"`
	Confidence      float64           `json:"confidence"` // 0.0 to 1.0
	ProcessedAt     time.Time         `json:"processed_at"`
	CreatedAt       time.Time         `json:"created_at"`
}

// ObjectDetection represents detected objects in the video
type ObjectDetection struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`       // person, vehicle, pedestrian, animal, etc.
	Label       string                 `json:"label"`      // specific classification
	Confidence  float64                `json:"confidence"` // 0.0 to 1.0
	BoundingBox BoundingBox            `json:"bounding_box"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// BoundingBox represents coordinates of detected objects
type BoundingBox struct {
	X      float64 `json:"x"`      // x coordinate of top-left corner
	Y      float64 `json:"y"`      // y coordinate of top-left corner
	Width  float64 `json:"width"`  // width of bounding box
	Height float64 `json:"height"` // height of bounding box
}

// BehaviorEvent represents behavioral analysis from video
type BehaviorEvent struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // sudden_stop, harsh_acceleration, lane_departure, drowsiness, etc.
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`   // low, medium, high, critical
	Confidence  float64   `json:"confidence"` // 0.0 to 1.0
	Description string    `json:"description"`
}

// TrafficSign represents detected traffic signs
type TrafficSign struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`  // speed_limit, stop, yield, no_entry, etc.
	Value      string    `json:"value"` // specific value like "30" for speed limit
	Timestamp  time.Time `json:"timestamp"`
	Confidence float64   `json:"confidence"` // 0.0 to 1.0
	Location   Location  `json:"location"`
}

// Incident represents safety incidents detected in video
type Incident struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"` // accident, near_miss, collision, etc.
	Timestamp        time.Time `json:"timestamp"`
	Severity         string    `json:"severity"`   // minor, moderate, severe, critical
	Confidence       float64   `json:"confidence"` // 0.0 to 1.0
	Description      string    `json:"description"`
	InvolvedVehicles []string  `json:"involved_vehicles,omitempty"`
	InvolvedPeople   []string  `json:"involved_people,omitempty"`
}

// Location represents geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// VideoRetentionPolicy represents retention settings
type VideoRetentionPolicy struct {
	ID              string    `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	EventType       *string   `json:"event_type,omitempty"` // Apply only to specific events
	RetentionPeriod int       `json:"retention_period"`     // in days
	AutoArchive     bool      `json:"auto_archive"`
	AutoDelete      bool      `json:"auto_delete"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// VideoThumbnail represents thumbnail data
type VideoThumbnail struct {
	ID          string    `json:"id"`
	RecordingID string    `json:"recording_id"`
	ImagePath   string    `json:"image_path"`
	ImageURL    string    `json:"image_url"`
	Timestamp   time.Time `json:"timestamp"` // Time offset in the video
	CreatedAt   time.Time `json:"created_at"`
}

// RequestClipRequest represents a request to capture a specific video clip from a device
type RequestClipRequest struct {
	TenantID  uuid.UUID `json:"tenant_id"`   // Organization making the request
	VehicleID uuid.UUID `json:"vehicle_id"`  // Target vehicle
	CameraID  uuid.UUID `json:"camera_id"`   // Specific camera on the device
	StartTime time.Time `json:"start_time"`  // Clip start timestamp
	EndTime   time.Time `json:"end_time"`    // Clip end timestamp
	Reason    string    `json:"reason"`      // Justification for clip request
}
