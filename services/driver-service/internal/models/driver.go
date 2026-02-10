package models

import (
	"time"
)

// Driver represents a vehicle driver
type Driver struct {
	ID                   string             `json:"id" db:"id"`
	OrgID                string             `json:"org_id" db:"org_id"`
	FirstName            string             `json:"first_name" db:"first_name"`
	LastName             string             `json:"last_name" db:"last_name"`
	FullName             string             `json:"name" db:"full_name"`
	Email                string             `json:"email" db:"email"`
	Phone                string             `json:"phone" db:"phone"`
	LicenseNumber        string             `json:"license_number" db:"license_number"`
	LicenseType          string             `json:"license_type" db:"license_type"` // LMV, HMV, Transport
	LicenseExpiry        time.Time          `json:"license_expiry" db:"license_expiry"`
	LicenseAuthority     string             `json:"license_authority" db:"license_authority"`
	DateOfBirth          time.Time          `json:"date_of_birth" db:"date_of_birth"`
	Address              Address            `json:"address" db:"address"`
	EmergencyContact     EmergencyContact   `json:"emergency_contact" db:"emergency_contact"`
	Status               string             `json:"status" db:"status"` // active, inactive, suspended, resigned
	CurrentVehicleID     *string            `json:"current_vehicle_id" db:"current_vehicle_id"`
	CurrentTripID        *string            `json:"current_trip_id" db:"current_trip_id"`
	JoiningDate          time.Time          `json:"joining_date" db:"joining_date"`
	ResignationDate      *time.Time         `json:"resignation_date" db:"resignation_date"`
	ExperienceYears      int                `json:"experience_years" db:"experience_years"`
	Rating               float64            `json:"rating" db:"rating"` // 0-5 scale
	TotalTrips           int                `json:"total_trips" db:"total_trips"`
	TotalDistanceKm      float64            `json:"total_distance_km" db:"total_distance_km"`
	TotalHoursDriven     float64            `json:"total_hours_driven" db:"total_hours_driven"`
	BehaviorScore        int                `json:"behavior_score" db:"behavior_score"` // 0-100 scale
	Documents            []DriverDocument   `json:"documents" db:"documents"`
	AttendanceRecords    []AttendanceRecord `json:"attendance_records" db:"attendance_records"`
	Schedules            []WorkSchedule     `json:"schedules" db:"schedules"`
	Metadata             map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt            time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at" db:"updated_at"`
	DeletedAt            *time.Time         `json:"deleted_at" db:"deleted_at"`
}

// Address represents driver's address information
type Address struct {
	Street     string `json:"street" db:"street"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	Pincode    string `json:"pincode" db:"pincode"`
	Country    string `json:"country" db:"country"`
	Latitude   *float64 `json:"latitude" db:"latitude"`
	Longitude  *float64 `json:"longitude" db:"longitude"`
}

// EmergencyContact represents emergency contact information
type EmergencyContact struct {
	Name     string `json:"name" db:"name"`
	Phone    string `json:"phone" db:"phone"`
	Relation string `json:"relation" db:"relation"`
}

// DriverDocument represents driver-related documents
type DriverDocument struct {
	ID          string    `json:"id" db:"id"`
	DocumentType string   `json:"document_type" db:"document_type"` // license, aadhaar, pan, medical, training
	DocumentNumber string `json:"document_number" db:"document_number"`
	IssuingAuthority string `json:"issuing_authority" db:"issuing_authority"`
	IssueDate   time.Time `json:"issue_date" db:"issue_date"`
	ExpiryDate  time.Time `json:"expiry_date" db:"expiry_date"`
	FileURL     string    `json:"file_url" db:"file_url"`
	Verified    bool      `json:"verified" db:"verified"`
	VerifiedBy  *string   `json:"verified_by" db:"verified_by"`
	VerifiedAt  *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AttendanceRecord represents driver attendance records
type AttendanceRecord struct {
	ID          string    `json:"id" db:"id"`
	Date        time.Time `json:"date" db:"date"`
	Status      string    `json:"status" db:"status"` // present, absent, leave, holiday
	CheckInTime *time.Time `json:"check_in_time" db:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time" db:"check_out_time"`
	Remarks     string    `json:"remarks" db:"remarks"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// WorkSchedule represents driver work schedules
type WorkSchedule struct {
	ID          string    `json:"id" db:"id"`
	DayOfWeek   int       `json:"day_of_week" db:"day_of_week"` // 0-6 (Sunday-Saturday)
	StartTime   string    `json:"start_time" db:"start_time"`   // HH:MM format
	EndTime     string    `json:"end_time" db:"end_time"`       // HH:MM format
	IsWorking   bool      `json:"is_working" db:"is_working"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDriverRequest represents the request body for creating a driver
type CreateDriverRequest struct {
	FirstName        string           `json:"first_name" validate:"required,max=50"`
	LastName         string           `json:"last_name" validate:"required,max=50"`
	Email            string           `json:"email" validate:"required,email"`
	Phone            string           `json:"phone" validate:"required,e164"`
	LicenseNumber    string           `json:"license_number" validate:"required"`
	LicenseType      string           `json:"license_type" validate:"required,oneof=LMV HMV Transport"`
	LicenseExpiry    time.Time        `json:"license_expiry" validate:"required"`
	LicenseAuthority string           `json:"license_authority" validate:"required"`
	DateOfBirth      time.Time        `json:"date_of_birth" validate:"required"`
	Address          Address          `json:"address" validate:"required"`
	EmergencyContact EmergencyContact `json:"emergency_contact" validate:"required"`
	ExperienceYears  int              `json:"experience_years" validate:"gte=0"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// UpdateDriverRequest represents the request body for updating a driver
type UpdateDriverRequest struct {
	FirstName        *string           `json:"first_name,omitempty" validate:"omitempty,max=50"`
	LastName         *string           `json:"last_name,omitempty" validate:"omitempty,max=50"`
	Email            *string           `json:"email,omitempty" validate:"omitempty,email"`
	Phone            *string           `json:"phone,omitempty" validate:"omitempty,e164"`
	LicenseNumber    *string           `json:"license_number,omitempty"`
	LicenseType      *string           `json:"license_type,omitempty" validate:"omitempty,oneof=LMV HMV Transport"`
	LicenseExpiry    *time.Time        `json:"license_expiry,omitempty"`
	LicenseAuthority *string           `json:"license_authority,omitempty"`
	DateOfBirth      *time.Time        `json:"date_of_birth,omitempty"`
	Address          *Address          `json:"address,omitempty"`
	EmergencyContact *EmergencyContact `json:"emergency_contact,omitempty"`
	ExperienceYears  *int              `json:"experience_years,omitempty" validate:"omitempty,gte=0"`
	Metadata         *map[string]interface{} `json:"metadata,omitempty"`
}

// AssignVehicleRequest represents the request body for assigning a vehicle to a driver
type AssignVehicleRequest struct {
	VehicleID string `json:"vehicle_id" validate:"required,uuid4"`
}

// UpdateStatusRequest represents the request body for updating driver status
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended resigned"`
}

// UpdateBehaviorScoreRequest represents the request body for updating behavior score
type UpdateBehaviorScoreRequest struct {
	ScoreChange int    `json:"score_change" validate:"required"`
	Reason      string `json:"reason" validate:"required,max=200"`
}

// UploadDocumentRequest represents the request body for uploading driver documents
type UploadDocumentRequest struct {
	DocumentType     string    `json:"document_type" validate:"required,oneof=license aadhaar pan medical training other"`
	DocumentNumber   string    `json:"document_number" validate:"required"`
	IssuingAuthority string    `json:"issuing_authority" validate:"required"`
	IssueDate        time.Time `json:"issue_date" validate:"required"`
	ExpiryDate       time.Time `json:"expiry_date" validate:"required"`
	FileURL          string    `json:"file_url" validate:"required,url"`
}

// MarkAttendanceRequest represents the request body for marking attendance
type MarkAttendanceRequest struct {
	Date        time.Time `json:"date" validate:"required"`
	Status      string    `json:"status" validate:"required,oneof=present absent leave holiday"`
	CheckInTime *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	Remarks     string    `json:"remarks,omitempty" validate:"max=200"`
}

// SetScheduleRequest represents the request body for setting work schedule
type SetScheduleRequest struct {
	DayOfWeek int    `json:"day_of_week" validate:"required,min=0,max=6"`
	StartTime string `json:"start_time" validate:"required,datetime=15:04"`
	EndTime   string `json:"end_time" validate:"required,datetime=15:04"`
	IsWorking bool   `json:"is_working"`
}

// DriverListFilter represents filters for listing drivers
type DriverListFilter struct {
	Status    string `json:"status" form:"status"`
	VehicleID string `json:"vehicle_id" form:"vehicle_id"`
	Name      string `json:"name" form:"name"`
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"page_size" form:"page_size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDir  string `json:"order_dir" form:"order_dir"`
}

// DriverAnalytics represents driver performance analytics
type DriverAnalytics struct {
	TotalTrips           int     `json:"total_trips"`
	CompletedTrips       int     `json:"completed_trips"`
	CancelledTrips       int     `json:"cancelled_trips"`
	TotalDistanceKm      float64 `json:"total_distance_km"`
	TotalHoursDriven     float64 `json:"total_hours_driven"`
	AverageRating        float64 `json:"average_rating"`
	CurrentBehaviorScore int     `json:"current_behavior_score"`
	AttendanceRate       float64 `json:"attendance_rate"`
	OnTimeRate           float64 `json:"on_time_rate"`
	FuelEfficiency       float64 `json:"fuel_efficiency"` // km/liter
	SpeedingIncidents    int     `json:"speeding_incidents"`
	HarshBrakingCount    int     `json:"harsh_braking_count"`
	IdlingTimeHours      float64 `json:"idling_time_hours"`
	CompletionRate       float64 `json:"completion_rate"`
}

// BehaviorEvent represents driver behavior events
type BehaviorEvent struct {
	ID          string    `json:"id" db:"id"`
	DriverID    string    `json:"driver_id" db:"driver_id"`
	EventType   string    `json:"event_type" db:"event_type"` // speeding, harsh_braking, harsh_acceleration, phone_usage, seatbelt_violation
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Severity    string    `json:"severity" db:"severity"` // low, medium, high, critical
	Value       float64   `json:"value" db:"value"`       // speed, g-force, etc.
	Location    *Location `json:"location" db:"location"` // optional GPS location
	TripID      *string   `json:"trip_id" db:"trip_id"`   // optional trip association
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	Address   string  `json:"address" db:"address"`
}

// DriverPerformanceSummary represents driver performance summary
type DriverPerformanceSummary struct {
	DriverID             string    `json:"driver_id"`
	Name                 string    `json:"name"`
	CurrentRating        float64   `json:"current_rating"`
	BehaviorScore        int       `json:"behavior_score"`
	TotalTrips           int       `json:"total_trips"`
	CompletionRate       float64   `json:"completion_rate"`
	AverageTripDistance  float64   `json:"average_trip_distance"`
	AverageTripDuration  float64   `json:"average_trip_duration"`
	LastActive           time.Time `json:"last_active"`
	Status               string    `json:"status"`
	VehicleAssigned      bool      `json:"vehicle_assigned"`
	LicenseValid         bool      `json:"license_valid"`
}

// DriverBehavior represents comprehensive driver behavior analytics
type DriverBehavior struct {
	ID             string                 `json:"id" db:"id"`
	DriverID       string                 `json:"driver_id" db:"driver_id"`
	OrgID          string                 `json:"org_id" db:"org_id"`
	PeriodStart    time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd      time.Time              `json:"period_end" db:"period_end"`
	Score          float64                `json:"score" db:"score"` // 0-100
	SafetyScore    float64                `json:"safety_score" db:"safety_score"`
	EfficiencyScore float64               `json:"efficiency_score" db:"efficiency_score"`
	ComplianceScore float64               `json:"compliance_score" db:"compliance_score"`
	
	// Safety Metrics
	HarshAccelerations int                 `json:"harsh_accelerations" db:"harsh_accelerations"`
	HarshBrakings      int                 `json:"harsh_brakings" db:"harsh_brakings"`
	HarshCornerings    int                 `json:"harsh_cornerings" db:"harsh_cornerings"`
	SpeedingIncidents  int                 `json:"speeding_incidents" db:"speeding_incidents"`
	PhoneUsage         int                 `json:"phone_usage" db:"phone_usage"`
	SeatbeltViolations int                 `json:"seatbelt_violations" db:"seatbelt_violations"`
	
	// Efficiency Metrics
	FuelConsumption    float64             `json:"fuel_consumption" db:"fuel_consumption"` // liters
	DistanceDriven     float64             `json:"distance_driven" db:"distance_driven"`   // kilometers
	IdleTime           int                 `json:"idle_time" db:"idle_time"`               // seconds
	EcoDrivingScore    float64             `json:"eco_driving_score" db:"eco_driving_score"`
	
	// Compliance Metrics
	HoursOfService     float64             `json:"hours_of_service" db:"hours_of_service"` // hours
	RestPeriods        int                 `json:"rest_periods" db:"rest_periods"`
	DocumentCompliance bool                `json:"document_compliance" db:"document_compliance"`
	MaintenanceAdherence bool              `json:"maintenance_adherence" db:"maintenance_adherence"`
	
	// Behavioral Patterns
	NightDrivingHours  float64             `json:"night_driving_hours" db:"night_driving_hours"`
	HighwayPercentage  float64             `json:"highway_percentage" db:"highway_percentage"`
	CityDrivingPercentage float64          `json:"city_driving_percentage" db:"city_driving_percentage"`
	
	// Risk Assessment
	RiskLevel          string              `json:"risk_level" db:"risk_level"` // low, medium, high, critical
	RiskFactors        []string            `json:"risk_factors" db:"risk_factors"`
	ImprovementAreas   []string            `json:"improvement_areas" db:"improvement_areas"`
	
	// Metadata
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt          time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at" db:"updated_at"`
}

// DriverIncident represents a specific driving incident
type DriverIncident struct {
	ID          string                 `json:"id" db:"id"`
	DriverID    string                 `json:"driver_id" db:"driver_id"`
	OrgID       string                 `json:"org_id" db:"org_id"`
	Type        string                 `json:"type" db:"type"` // speeding, harsh_braking, harsh_acceleration, etc.
	Severity    string                 `json:"severity" db:"severity"` // low, medium, high, critical
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Location    Location               `json:"location" db:"location"`
	Speed       float64                `json:"speed" db:"speed"`
	Details     map[string]interface{} `json:"details" db:"details"`
	Resolved    bool                   `json:"resolved" db:"resolved"`
	ResolutionNotes string             `json:"resolution_notes" db:"resolution_notes"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// BehaviorReportRequest represents request for behavior reports
type BehaviorReportRequest struct {
	DriverID   *string    `json:"driver_id,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	PeriodType string     `json:"period_type"` // daily, weekly, monthly
}

// BehaviorScoreUpdate represents behavior score adjustment
type BehaviorScoreUpdate struct {
	ID          string                 `json:"id" db:"id"`
	DriverID    string                 `json:"driver_id" db:"driver_id"`
	ScoreChange int                    `json:"score_change" db:"score_change"`
	Reason      string                 `json:"reason" db:"reason"`
	EventType   string                 `json:"event_type" db:"event_type"`
	Details     map[string]interface{} `json:"details" db:"details"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}
