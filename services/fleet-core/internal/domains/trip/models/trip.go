package models

import (
	"time"
)

// Trip State Machine Constants
const (
	TripStatusCreated   = "CREATED"
	TripStatusScheduled = "SCHEDULED"
	TripStatusStarted   = "STARTED"
	TripStatusPaused    = "PAUSED"
	TripStatusResumed   = "RESUMED"
	TripStatusCompleted = "COMPLETED"
	TripStatusCancelled = "CANCELLED"
	TripStatusFailed    = "FAILED"
)

// Validation constants
const (
	MaxTitleLength        = 100
	MaxDescriptionLength  = 500
	MaxNotesLength        = 500
	MinRouteWaypoints     = 2
	MaxPhotoCount         = 10
	MinFuelLevel          = 0.0
	MaxFuelLevel          = 100.0
	MinDistanceKm         = 0.0
	MinDurationMinutes    = 0
	MinOdometerReading    = 0.0
	MinFuelConsumedLiters = 0.0
	MinIdlingTimeMinutes  = 0
	MinPODDocuments       = 1
)

// Trip represents a vehicle journey
type Trip struct {
	ID                   string       `json:"id" db:"id"`
	OrgID                string       `json:"org_id" db:"org_id"`
	VehicleID            string       `json:"vehicle_id" db:"vehicle_id"`
	DriverID             *string      `json:"driver_id" db:"driver_id"`
	Status               string       `json:"status" db:"status"` // planned, in_progress, paused, completed, cancelled
	TripType             string       `json:"trip_type" db:"trip_type"` // delivery, service, long_haul, standard, local, return
	Title                string       `json:"title" db:"title"`
	Description          string       `json:"description" db:"description"`
	PickupGeo            *Location    `json:"pickup_geo" db:"pickup_geo"`
	DropGeo              *Location    `json:"drop_geo" db:"drop_geo"`
	PlannedRoute         []Waypoint   `json:"planned_route" db:"planned_route"`
	ActualRoute          []Waypoint   `json:"actual_route" db:"actual_route"`
	Waypoints            []Waypoint   `json:"waypoints" db:"waypoints"`
	PlannedDistance      float64      `json:"planned_distance" db:"planned_distance"`
	TotalDistance        *float64     `json:"total_distance" db:"total_distance"`
	EstimatedDurationMin int          `json:"estimated_duration_min" db:"estimated_duration_min"`
	ActualDurationMin    *int         `json:"actual_duration_min" db:"actual_duration_min"`
	CargoDetails         CargoDetails `json:"cargo_details" db:"cargo_details"`
	ScheduledStart       time.Time    `json:"scheduled_start" db:"scheduled_start"`
	PlannedETA           *time.Time   `json:"planned_eta" db:"planned_eta"`
	StartTime            *time.Time   `json:"start_time" db:"start_time"`
	EndTime              *time.Time   `json:"end_time" db:"end_time"`
	StartLocation        *Location    `json:"start_location" db:"start_location"`
	EndLocation          *Location    `json:"end_location" db:"end_location"`
	StartOdometer        *float64     `json:"start_odometer" db:"start_odometer"`
	EndOdometer          *float64     `json:"end_odometer" db:"end_odometer"`
	FuelLevelSnapshot    *float64     `json:"fuel_level_snapshot" db:"fuel_level_snapshot"`
	FuelUsed             *float64     `json:"fuel_used" db:"fuel_used"`
	IdlingTimeMinutes    *int         `json:"idling_time_minutes" db:"idling_time_minutes"`

	// India-Specific Compliance
	EWayBillNo        string `json:"eway_bill_no" db:"eway_bill_no"`
	GSTClassification string `json:"gst_classification" db:"gst_classification"`
	PermitVerified    bool   `json:"permit_verified" db:"permit_verified"`
	AIS140Compliant   bool   `json:"ais140_compliant" db:"ais140_compliant"`

	// Compliance & Audit (Section 9)
	LegalHold     bool   `json:"legal_hold" db:"legal_hold"`         // AIS-140/CERT-In 7-year archival
	CorrelationID string `json:"correlation_id" db:"correlation_id"` // Per-trip distributed tracing ID

	// India-ready Documents & Metadata
	PODDocuments       []PODDocument `json:"pod_documents" db:"pod_documents"`
	CustomerName       string        `json:"customer_name" db:"customer_name"`
	CustomerPhone      string        `json:"customer_phone" db:"customer_phone"`
	CustomerEmail      string        `json:"customer_email" db:"customer_email"`
	CancellationReason string        `json:"cancellation_reason" db:"cancellation_reason"`
	IsDeleted          bool          `json:"is_deleted" db:"is_deleted"`
	Version            int           `json:"version" db:"version"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" db:"updated_at"`
}

// Waypoint represents a location along a route
type Waypoint struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"` // pickup, dropoff, checkpoint
	Address     string    `json:"address" db:"address"`
	Location    Location  `json:"location" db:"location"`
	Sequence    int       `json:"sequence" db:"sequence"`
	ArrivalTime time.Time `json:"arrival_time" db:"arrival_time"`
	Status      string    `json:"status" db:"status"` // pending, arrived, departed, skipped
	Notes       string    `json:"notes" db:"notes"`
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	Address   string  `json:"address" db:"address"`
}

// CargoDetails represents shipment information
type CargoDetails struct {
	Type        string  `json:"type" db:"type"`
	WeightKg    float64 `json:"weight_kg" db:"weight_kg"`
	VolumeM3    float64 `json:"volume_m3" db:"volume_m3"`
	ValueINR    float64 `json:"value_inr" db:"value_inr"`
	Description string  `json:"description" db:"description"`
	Fragile     bool    `json:"fragile" db:"fragile"`
	Temperature string  `json:"temperature" db:"temperature"` // ambient, cold, frozen
}

// TripEvent represents significant events during a trip
type TripEvent struct {
	ID          string                 `json:"id" db:"id"`
	Type        string                 `json:"type" db:"type"` // arrival, departure, delivery_attempt, delivery_success, delivery_failed, delay, breakdown, accident
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Location    Location               `json:"location" db:"location"`
	Description string                 `json:"description" db:"description"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Photos      []string               `json:"photos" db:"photos"`
	CreatedBy   string                 `json:"created_by" db:"created_by"` // driver_id or system
	Hash        string                 `json:"hash" db:"hash"`
	PrevHash    string                 `json:"prev_hash" db:"prev_hash"`
}

// TelemetryPoint represents a single GPS/IoT data point
type TelemetryPoint struct {
	TripID         string    `json:"trip_id" db:"trip_id"`
	Latitude       float64   `json:"lat" db:"lat"`
	Longitude      float64   `json:"lng" db:"lng"`
	Speed          float64   `json:"speed" db:"speed"`
	Heading        float64   `json:"heading" db:"heading"`
	FuelLevel      float64   `json:"fuel_level" db:"fuel_level"`
	IgnitionStatus bool      `json:"ignition_status" db:"ignition_status"`
	RecordedAt     time.Time `json:"recorded_at" db:"recorded_at"`
}

// PODDocument represents Proof of Delivery documents
type PODDocument struct {
	ID          string    `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"` // signature, photo, invoice, receipt
	URL         string    `json:"url" db:"url"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	UploadedBy  string    `json:"uploaded_by" db:"uploaded_by"`
	Recipient   Recipient `json:"recipient" db:"recipient"`
	OTPVerified bool      `json:"otp_verified" db:"otp_verified"`
	Signature   string    `json:"signature" db:"signature"` // base64 encoded
}

// Recipient represents delivery recipient information
type Recipient struct {
	Name  string `json:"name" db:"name"`
	Phone string `json:"phone" db:"phone"`
	Email string `json:"email" db:"email"`
}

// CreateTripRequest represents the request body for creating a trip
type CreateTripRequest struct {
	VehicleID            string       `json:"vehicle_id" validate:"required,uuid4"`
	DriverID             string       `json:"driver_id" validate:"required,uuid4"`
	Title                string       `json:"title" validate:"required,max=100"`
	Description          string       `json:"description" validate:"max=500"`
	PlannedRoute         []Waypoint   `json:"planned_route" validate:"required,min=2"`
	Waypoints            []Waypoint   `json:"waypoints" validate:"required,min=2"`
	PlannedDistance      float64      `json:"planned_distance" validate:"required,gt=0"`
	EstimatedDurationMin int          `json:"estimated_duration_min" validate:"required,gt=0"`
	CargoDetails         CargoDetails `json:"cargo_details" validate:"required"`
	ScheduledStart       time.Time    `json:"scheduled_start" validate:"required"`
	PlannedETA           *time.Time   `json:"planned_eta" validate:"omitempty"`
	StartOdometer        *float64     `json:"start_odometer" validate:"omitempty,gte=0"`
	FuelLevelSnapshot    *float64     `json:"fuel_level_snapshot" validate:"omitempty,gte=0,lte=100"`
	CustomerName         string       `json:"customer_name" validate:"required"`
	CustomerPhone        string       `json:"customer_phone" validate:"required,e164"`
	CustomerEmail        string       `json:"customer_email" validate:"omitempty,email"`

	// Advanced Fields
	PickupGeo         *Location `json:"pickup_geo" validate:"omitempty"`
	DropGeo           *Location `json:"drop_geo" validate:"omitempty"`
	EWayBillNo        string    `json:"eway_bill_no" validate:"omitempty"`
	GSTClassification string    `json:"gst_classification" validate:"omitempty"`
}

// UpdateTripRequest represents the request body for updating a trip
type UpdateTripRequest struct {
	Title                *string       `json:"title,omitempty" validate:"omitempty,max=100"`
	Description          *string       `json:"description,omitempty" validate:"omitempty,max=500"`
	PlannedRoute         *[]Waypoint   `json:"planned_route,omitempty" validate:"omitempty,min=2"`
	Waypoints            *[]Waypoint   `json:"waypoints,omitempty" validate:"omitempty,min=2"`
	PlannedDistance      *float64      `json:"planned_distance,omitempty" validate:"omitempty,gt=0"`
	EstimatedDurationMin *int          `json:"estimated_duration_min,omitempty" validate:"omitempty,gt=0"`
	CargoDetails         *CargoDetails `json:"cargo_details,omitempty"`
	ScheduledStart       *time.Time    `json:"scheduled_start,omitempty"`
	CustomerName         *string       `json:"customer_name,omitempty"`
	CustomerPhone        *string       `json:"customer_phone,omitempty" validate:"omitempty,e164"`
	CustomerEmail        *string       `json:"customer_email,omitempty" validate:"omitempty,email"`
}

// StartTripRequest represents the request body for starting a trip
type StartTripRequest struct {
	OdometerReading float64 `json:"odometer_reading" validate:"required,gt=0"`
	FuelLevel       float64 `json:"fuel_level" validate:"required,gte=0,lte=100"`
	Notes           string  `json:"notes" validate:"max=500"`
}

// CompleteTripRequest represents the request body for completing a trip
type CompleteTripRequest struct {
	OdometerReading   float64       `json:"odometer_reading" validate:"required,gt=0"`
	FuelLevel         float64       `json:"fuel_level" validate:"required,gte=0,lte=100"`
	TotalDistance     float64       `json:"total_distance" validate:"required,gt=0"`
	FuelUsed          float64       `json:"fuel_used" validate:"required,gte=0"`
	IdlingTimeMinutes int           `json:"idling_time_minutes" validate:"gte=0"`
	PODDocuments      []PODDocument `json:"pod_documents" validate:"required,min=1"`
	Notes             string        `json:"notes" validate:"max=500"`
}

// AddEventRequest represents the request body for adding trip events
type AddEventRequest struct {
	Type        string   `json:"type" validate:"required,oneof=arrival departure delivery_attempt delivery_success delivery_failed delay breakdown accident"`
	Location    Location `json:"location" validate:"required"`
	Description string   `json:"description" validate:"required,max=500"`
	Photos      []string `json:"photos" validate:"max=10"`
}

// SubmitPODRequest represents the request body for submitting proof of delivery
type SubmitPODRequest struct {
	Type        string    `json:"type" validate:"required,oneof=signature photo invoice receipt"`
	URL         string    `json:"url" validate:"required,url"`
	Recipient   Recipient `json:"recipient" validate:"required"`
	OTPVerified bool      `json:"otp_verified"`
	Signature   string    `json:"signature"` // base64 encoded
}

// OptimizeRouteRequest represents the request body for route optimization
type OptimizeRouteRequest struct {
	AvoidTolls   bool `json:"avoid_tolls"`
	TrafficAware bool `json:"traffic_aware"`
	MinimizeCost bool `json:"minimize_cost"`
	MinimizeTime bool `json:"minimize_time"`
}

// RouteOptimization represents the result of a route optimization request
type RouteOptimization struct {
	TripID    string     `json:"trip_id"`
	Polyline  string     `json:"polyline"`
	Waypoints []Waypoint `json:"waypoints"`
	Distance  float64    `json:"distance"`
	Duration  int        `json:"duration"`
	Optimized bool       `json:"optimized"`
}

// TripListFilter represents filters for listing trips
type TripListFilter struct {
	OrgID        string `json:"org_id"`
	VehicleID    string `json:"vehicle_id"`
	DriverID     string `json:"driver_id"`
	CustomerName string `json:"customer_name"` // Party/organization that gave the job
	Status       string `json:"status"`
	TripType     string `json:"trip_type"` // delivery, service, long_haul, standard, local, return
	From         string `json:"from"`      // Date range start
	To           string `json:"to"`        // Date range end
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	OrderBy      string `json:"order_by"`
	OrderDir     string `json:"order_dir"`
}

// TripSummary represents trip analytics and summary information
type TripSummary struct {
	TotalTrips            int     `json:"total_trips"`
	CompletedTrips        int     `json:"completed_trips"`
	CancelledTrips        int     `json:"cancelled_trips"`
	TotalDistance         float64 `json:"total_distance"`
	TotalDurationHours    float64 `json:"total_duration_hours"`
	AverageFuelEfficiency float64 `json:"average_fuel_efficiency"`
	FuelUsed              float64 `json:"fuel_used"`
	TotalIdleTimeMinutes  int     `json:"total_idle_time_minutes"`
	OnTimeRate            float64 `json:"on_time_rate"`
	AverageTripDuration   float64 `json:"average_trip_duration"`
}

// TollInfo represents toll information along a route
type TollInfo struct {
	Name     string   `json:"name"`
	CostINR  float64  `json:"cost_inr"`
	Location Location `json:"location"`
}

// RouteOptimizationResult represents the result of route optimization
type RouteOptimizationResult struct {
	Routes []OptimizedRoute `json:"routes"`
}

// OptimizedRoute represents an optimized route option
type OptimizedRoute struct {
	Polyline         string     `json:"polyline"`
	DistanceKm       float64    `json:"distance_km"`
	DurationMinutes  int        `json:"duration_minutes"`
	Tolls            []TollInfo `json:"tolls"`
	FuelCostEstimate float64    `json:"fuel_cost_estimate"`
	WeatherAlerts    []string   `json:"weather_alerts"`
	Instructions     []string   `json:"instructions"`
}

type TripStats struct {
	ActiveTrips int `json:"active_trips"`
}

// ProofOfDelivery represents a POD document stored in the separate proof_of_delivery table
type ProofOfDelivery struct {
	ID         string                 `json:"id" db:"id"`
	TripID     string                 `json:"trip_id" db:"trip_id"`
	Type       string                 `json:"type" db:"type"`
	URL        string                 `json:"url" db:"url"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
	UploadedBy string                 `json:"uploaded_by" db:"uploaded_by"`
	Verified   bool                   `json:"verified" db:"verified"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// TripOffer represents a trip assignment offer to a driver
type TripOffer struct {
	ID              string     `json:"id" db:"id"`
	TripID          string     `json:"trip_id" db:"trip_id"`
	DriverID        string     `json:"driver_id" db:"driver_id"`
	Status          string     `json:"status" db:"status"` // OFFERED, ACCEPTED, REJECTED, EXPIRED
	RejectionReason string     `json:"rejection_reason" db:"rejection_reason"`
	ExpiresAt       *time.Time `json:"expires_at" db:"expires_at"`
	RespondedAt     *time.Time `json:"responded_at" db:"responded_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// DriverPerformance represents performance metrics for a driver
type DriverPerformance struct {
	DriverID       string  `json:"driver_id"`
	Name           string  `json:"name"`
	TotalDistance  float64 `json:"total_distance"`
	SafetyScore    int     `json:"safety_score"`
	TripsCompleted int     `json:"trips_completed"`
	TotalTrips     int     `json:"total_trips"`
	OnTimeRate     float64 `json:"on_time_rate"`
}

// DailyTripTrend represents trip metrics for a single day
type DailyTripTrend struct {
	Date            string  `json:"date"`
	TotalTrips      int     `json:"total_trips"`
	TotalDistance   float64 `json:"total_distance"`
	ActiveTimeHours float64 `json:"active_time_hours"`
}
