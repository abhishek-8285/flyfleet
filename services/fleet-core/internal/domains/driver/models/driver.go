package models

import (
	"time"
)

// Driver represents a vehicle driver
type Driver struct {
	ID                string                 `json:"id" db:"id"`
	OrgID             string                 `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	FirstName         string                 `json:"first_name" db:"first_name" validate:"required,max=50"`
	LastName          string                 `json:"last_name" db:"last_name" validate:"required,max=50"`
	FullName          string                 `json:"name" db:"full_name"`
	Email             string                 `json:"email" db:"email" validate:"required,email"`
	Phone             string                 `json:"phone" db:"phone" validate:"required,e164"`
	EmployeeCode      *string                `json:"employee_code" db:"employee_code" validate:"omitempty,max=50"`
	AadhaarNumber     *string                `json:"aadhaar_number" db:"aadhaar_number" validate:"omitempty,len=12"`
	PanNumber         *string                `json:"pan_number" db:"pan_number" validate:"omitempty,len=10"`
	LicenseNumber     string                 `json:"license_number" db:"license_number" validate:"required,max=50"`
	LicenseType       *string                `json:"license_type" db:"license_type" validate:"omitempty,oneof=LMV HMV Transport"`
	LicenseExpiry     *time.Time             `json:"license_expiry" db:"license_expiry" validate:"omitempty"`
	LicenseIssueDate  *time.Time             `json:"license_issue_date" db:"license_issue_date" validate:"omitempty"`
	LicenseAuthority  *string                `json:"license_authority" db:"license_authority" validate:"omitempty,max=100"`
	DateOfBirth       *time.Time             `json:"date_of_birth" db:"date_of_birth" validate:"omitempty"`
	Address           Address                `json:"address" db:"address" validate:"required"`
	EmergencyContact  EmergencyContact       `json:"emergency_contact" db:"emergency_contact" validate:"required"`
	Status            string                 `json:"status" db:"status" validate:"required,oneof=active inactive suspended resigned"`
	CurrentVehicleID  *string                `json:"current_vehicle_id" db:"current_vehicle_id"`
	CurrentTripID     *string                `json:"current_trip_id" db:"current_trip_id"`
	JoiningDate       time.Time              `json:"joining_date" db:"joining_date"`
	ResignationDate   *time.Time             `json:"resignation_date" db:"resignation_date"`
	ExperienceYears   int                    `json:"experience_years" db:"experience_years" validate:"gte=0,lte=50"`
	EmploymentType    *string                `json:"employment_type" db:"employment_type" validate:"omitempty,oneof=FULL_TIME CONTRACT"`
	Rating            float64                `json:"rating" db:"rating" validate:"gte=0,lte=5"`
	TotalTrips        int                    `json:"total_trips" db:"total_trips" validate:"gte=0"`
	TotalDistanceKm   float64                `json:"total_distance_km" db:"total_distance_km" validate:"gte=0"`
	TotalHoursDriven  float64                `json:"total_hours_driven" db:"total_hours_driven" validate:"gte=0"`
	BehaviorScore     int                    `json:"behavior_score" db:"behavior_score" validate:"gte=0,lte=100"`
	Documents         []DriverDocument       `json:"documents" db:"documents"`
	AttendanceRecords []AttendanceRecord     `json:"attendance_records" db:"attendance_records"`
	Schedules         []WorkSchedule         `json:"schedules" db:"schedules"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time             `json:"deleted_at" db:"deleted_at"`
}

// Address represents driver's address information
type Address struct {
	Street    string   `json:"street" db:"street" validate:"required,max=200"`
	City      string   `json:"city" db:"city" validate:"required,max=100"`
	State     string   `json:"state" db:"state" validate:"required,max=100"`
	Pincode   string   `json:"pincode" db:"pincode" validate:"required,max=20"`
	Country   string   `json:"country" db:"country" validate:"required,max=100"`
	Latitude  *float64 `json:"latitude" db:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude *float64 `json:"longitude" db:"longitude" validate:"omitempty,gte=-180,lte=180"`
}

// EmergencyContact represents emergency contact information
type EmergencyContact struct {
	Name     string `json:"name" db:"name" validate:"required,max=100"`
	Phone    string `json:"phone" db:"phone" validate:"required,e164"`
	Relation string `json:"relation" db:"relation" validate:"required,max=50"`
}

// DriverDocument represents driver-related documents
type DriverDocument struct {
	ID               string     `json:"id" db:"id" validate:"required,uuid"`
	DocumentType     string     `json:"document_type" db:"document_type" validate:"required,oneof=license aadhaar pan medical training other"`
	DocumentNumber   string     `json:"document_number" db:"document_number" validate:"required,max=100"`
	IssuingAuthority string     `json:"issuing_authority" db:"issuing_authority" validate:"required,max=100"`
	IssueDate        time.Time  `json:"issue_date" db:"issue_date" validate:"required"`
	ExpiryDate       time.Time  `json:"expiry_date" db:"expiry_date" validate:"required"`
	FileURL          string     `json:"file_url" db:"file_url" validate:"required,url"`
	Verified         bool       `json:"verified" db:"verified"`
	VerifiedBy       *string    `json:"verified_by" db:"verified_by"`
	VerifiedAt       *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// AttendanceRecord represents driver attendance records
type AttendanceRecord struct {
	ID           string     `json:"id" db:"id" validate:"required,uuid"`
	Date         time.Time  `json:"date" db:"date" validate:"required"`
	Status       string     `json:"status" db:"status" validate:"required,oneof=present absent leave holiday"`
	CheckInTime  *time.Time `json:"check_in_time" db:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time" db:"check_out_time"`
	Remarks      string     `json:"remarks" db:"remarks" validate:"max=200"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// WorkSchedule represents driver work schedules
type WorkSchedule struct {
	ID        string    `json:"id" db:"id" validate:"required,uuid"`
	DayOfWeek int       `json:"day_of_week" db:"day_of_week" validate:"required,min=0,max=6"`
	StartTime string    `json:"start_time" db:"start_time" validate:"required"`
	EndTime   string    `json:"end_time" db:"end_time" validate:"required"`
	IsWorking bool      `json:"is_working" db:"is_working"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDriverRequest represents the request body for creating a driver
type CreateDriverRequest struct {
	FirstName        string                 `json:"first_name" validate:"required,max=50"`
	LastName         string                 `json:"last_name" validate:"required,max=50"`
	Email            string                 `json:"email" validate:"required,email"`
	Phone            string                 `json:"phone" validate:"required,e164"`
	LicenseNumber    string                 `json:"license_number" validate:"required,max=50"`
	LicenseType      string                 `json:"license_type" validate:"required,oneof=LMV HMV Transport"`
	LicenseExpiry    time.Time              `json:"license_expiry" validate:"required"`
	LicenseAuthority string                 `json:"license_authority" validate:"required,max=100"`
	LicenseIssueDate *time.Time             `json:"license_issue_date" validate:"omitempty"`
	DateOfBirth      time.Time              `json:"date_of_birth" validate:"required"`
	EmployeeCode     string                 `json:"employee_code" validate:"omitempty,max=50"`
	AadhaarNumber    string                 `json:"aadhaar_number" validate:"omitempty,len=12"`
	PanNumber        string                 `json:"pan_number" validate:"omitempty,len=10"`
	EmploymentType   string                 `json:"employment_type" validate:"required,oneof=FULL_TIME CONTRACT"`
	Address          Address                `json:"address" validate:"required"`
	EmergencyContact EmergencyContact       `json:"emergency_contact" validate:"required"`
	ExperienceYears  int                    `json:"experience_years" validate:"gte=0,lte=50"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// UpdateDriverRequest represents the request body for updating a driver
type UpdateDriverRequest struct {
	FirstName        *string                 `json:"first_name,omitempty" validate:"omitempty,max=50"`
	LastName         *string                 `json:"last_name,omitempty" validate:"omitempty,max=50"`
	Email            *string                 `json:"email,omitempty" validate:"omitempty,email"`
	Phone            *string                 `json:"phone,omitempty" validate:"omitempty,e164"`
	LicenseNumber    *string                 `json:"license_number,omitempty" validate:"omitempty,max=50"`
	LicenseType      *string                 `json:"license_type,omitempty" validate:"omitempty,oneof=LMV HMV Transport"`
	LicenseExpiry    *time.Time              `json:"license_expiry,omitempty" validate:"omitempty"`
	LicenseAuthority *string                 `json:"license_authority,omitempty" validate:"omitempty,max=100"`
	LicenseIssueDate *time.Time              `json:"license_issue_date,omitempty" validate:"omitempty"`
	DateOfBirth      *time.Time              `json:"date_of_birth,omitempty" validate:"omitempty"`
	EmployeeCode     *string                 `json:"employee_code,omitempty" validate:"omitempty,max=50"`
	AadhaarNumber    *string                 `json:"aadhaar_number,omitempty" validate:"omitempty,len=12"`
	PanNumber        *string                 `json:"pan_number,omitempty" validate:"omitempty,len=10"`
	EmploymentType   *string                 `json:"employment_type,omitempty" validate:"omitempty,oneof=FULL_TIME CONTRACT"`
	Address          *Address                `json:"address,omitempty" validate:"omitempty"`
	EmergencyContact *EmergencyContact       `json:"emergency_contact,omitempty" validate:"omitempty"`
	ExperienceYears  *int                    `json:"experience_years,omitempty" validate:"omitempty,gte=0,lte=50"`
	Metadata         *map[string]interface{} `json:"metadata,omitempty"`
}

// AssignVehicleRequest represents the request body for assigning a vehicle to a driver
type AssignVehicleRequest struct {
	VehicleID string `json:"vehicle_id" validate:"required,uuid"`
}

// UpdateStatusRequest represents the request body for updating driver status
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended resigned"`
}

// UpdateBehaviorScoreRequest represents the request body for updating behavior score
type UpdateBehaviorScoreRequest struct {
	ScoreChange int    `json:"score_change" validate:"required,gte=-20,lte=20"`
	Reason      string `json:"reason" validate:"required,max=200"`
}

// UploadDocumentRequest represents the request body for uploading driver documents
type UploadDocumentRequest struct {
	DocumentType     string    `json:"document_type" validate:"required,oneof=license aadhaar pan medical training other"`
	DocumentNumber   string    `json:"document_number" validate:"required,max=100"`
	IssuingAuthority string    `json:"issuing_authority" validate:"required,max=100"`
	IssueDate        time.Time `json:"issue_date" validate:"required"`
	ExpiryDate       time.Time `json:"expiry_date" validate:"required"`
	FileURL          string    `json:"file_url" validate:"required,url"`
}

// MarkAttendanceRequest represents the request body for marking attendance
type MarkAttendanceRequest struct {
	Date         time.Time  `json:"date" validate:"required"`
	Status       string     `json:"status" validate:"required,oneof=present absent leave holiday"`
	CheckInTime  *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	Remarks      string     `json:"remarks,omitempty" validate:"max=200"`
}

// SetScheduleRequest represents the request body for setting work schedule
type SetScheduleRequest struct {
	DayOfWeek int    `json:"day_of_week" validate:"required,min=0,max=6"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
	IsWorking bool   `json:"is_working"`
}

// DriverListFilter represents filters for listing drivers
type DriverListFilter struct {
	OrgID     string `json:"tenant_id" form:"tenant_id" validate:"omitempty,uuid"`
	Status    string `json:"status" form:"status" validate:"omitempty,oneof=active inactive suspended resigned"`
	VehicleID string `json:"vehicle_id" form:"vehicle_id" validate:"omitempty,uuid"`
	Name      string `json:"name" form:"name" validate:"omitempty,max=100"`
	Page      int    `json:"page" form:"page" validate:"gte=1"`
	PageSize  int    `json:"page_size" form:"page_size" validate:"gte=1,lte=100"`
	OrderBy   string `json:"order_by" form:"order_by" validate:"omitempty,oneof=created_at full_name rating total_trips behavior_score status"`
	OrderDir  string `json:"order_dir" form:"order_dir" validate:"omitempty,oneof=ASC DESC"`
}

// DriverAnalytics represents driver performance analytics
type DriverAnalytics struct {
	TotalTrips           int     `json:"total_trips" validate:"gte=0"`
	CompletedTrips       int     `json:"completed_trips" validate:"gte=0"`
	CancelledTrips       int     `json:"cancelled_trips" validate:"gte=0"`
	TotalDistanceKm      float64 `json:"total_distance_km" validate:"gte=0"`
	TotalHoursDriven     float64 `json:"total_hours_driven" validate:"gte=0"`
	AverageRating        float64 `json:"average_rating" validate:"gte=0,lte=5"`
	CurrentBehaviorScore int     `json:"current_behavior_score" validate:"gte=0,lte=100"`
	AttendanceRate       float64 `json:"attendance_rate" validate:"gte=0,lte=100"`
	OnTimeRate           float64 `json:"on_time_rate" validate:"gte=0,lte=100"`
	FuelEfficiency       float64 `json:"fuel_efficiency" validate:"gte=0"`
	SpeedingIncidents    int     `json:"speeding_incidents" validate:"gte=0"`
	HarshBrakingCount    int     `json:"harsh_braking_count" validate:"gte=0"`
	IdlingTimeHours      float64 `json:"idling_time_hours" validate:"gte=0"`
	CompletionRate       float64 `json:"completion_rate" validate:"gte=0,lte=100"`
}

// BehaviorEvent represents driver behavior events
type BehaviorEvent struct {
	ID        string    `json:"id" db:"id" validate:"required,uuid"`
	DriverID  string    `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	EventType string    `json:"event_type" db:"event_type" validate:"required,oneof=speeding harsh_braking harsh_acceleration phone_usage seatbelt_violation"`
	Timestamp time.Time `json:"timestamp" db:"timestamp" validate:"required"`
	Severity  string    `json:"severity" db:"severity" validate:"required,oneof=low medium high critical"`
	Value     float64   `json:"value" db:"value"`
	Location  *Location `json:"location" db:"location"`
	TripID    *string   `json:"trip_id" db:"trip_id" validate:"omitempty,uuid"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Behavior Penalty Constants
const (
	PenaltySpeeding          = 5
	PenaltyHarshBrake        = 5
	PenaltyHarshAcceleration = 3
	PenaltyPhoneUsage        = 15
	PenaltySeatbelt          = 10
)

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude" validate:"gte=-90,lte=90"`
	Longitude float64 `json:"longitude" db:"longitude" validate:"gte=-180,lte=180"`
	Address   string  `json:"address" db:"address" validate:"max=200"`
}

// DriverPerformanceSummary represents driver performance summary
type DriverPerformanceSummary struct {
	DriverID            string    `json:"driver_id" validate:"required,uuid"`
	Name                string    `json:"name" validate:"required,max=100"`
	CurrentRating       float64   `json:"current_rating" validate:"gte=0,lte=5"`
	BehaviorScore       int       `json:"behavior_score" validate:"gte=0,lte=100"`
	TotalTrips          int       `json:"total_trips" validate:"gte=0"`
	CompletionRate      float64   `json:"completion_rate" validate:"gte=0,lte=100"`
	AverageTripDistance float64   `json:"average_trip_distance" validate:"gte=0"`
	AverageTripDuration float64   `json:"average_trip_duration" validate:"gte=0"`
	LastActive          time.Time `json:"last_active"`
	Status              string    `json:"status" validate:"required,oneof=active inactive suspended resigned"`
	VehicleAssigned     bool      `json:"vehicle_assigned"`
	LicenseValid        bool      `json:"license_valid"`
}

// DriverBehavior represents comprehensive driver behavior analytics
type DriverBehavior struct {
	ID              string    `json:"id" db:"id" validate:"required,uuid"`
	DriverID        string    `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	OrgID           string    `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	PeriodStart     time.Time `json:"period_start" db:"period_start" validate:"required"`
	PeriodEnd       time.Time `json:"period_end" db:"period_end" validate:"required"`
	Score           float64   `json:"score" db:"score" validate:"gte=0,lte=100"`
	SafetyScore     float64   `json:"safety_score" db:"safety_score" validate:"gte=0,lte=100"`
	EfficiencyScore float64   `json:"efficiency_score" db:"efficiency_score" validate:"gte=0,lte=100"`
	ComplianceScore float64   `json:"compliance_score" db:"compliance_score" validate:"gte=0,lte=100"`

	// Safety Metrics
	HarshAccelerations int `json:"harsh_accelerations" db:"harsh_accelerations" validate:"gte=0"`
	HarshBrakings      int `json:"harsh_brakings" db:"harsh_brakings" validate:"gte=0"`
	HarshCornerings    int `json:"harsh_cornerings" db:"harsh_cornerings" validate:"gte=0"`
	SpeedingIncidents  int `json:"speeding_incidents" db:"speeding_incidents" validate:"gte=0"`
	PhoneUsage         int `json:"phone_usage" db:"phone_usage" validate:"gte=0"`
	SeatbeltViolations int `json:"seatbelt_violations" db:"seatbelt_violations" validate:"gte=0"`

	// Efficiency Metrics
	FuelConsumption float64 `json:"fuel_consumption" db:"fuel_consumption" validate:"gte=0"`
	DistanceDriven  float64 `json:"distance_driven" db:"distance_driven" validate:"gte=0"`
	IdleTime        int     `json:"idle_time" db:"idle_time" validate:"gte=0"`
	EcoDrivingScore float64 `json:"eco_driving_score" db:"eco_driving_score" validate:"gte=0,lte=100"`

	// Compliance Metrics
	HoursOfService       float64 `json:"hours_of_service" db:"hours_of_service" validate:"gte=0"`
	RestPeriods          int     `json:"rest_periods" db:"rest_periods" validate:"gte=0"`
	DocumentCompliance   bool    `json:"document_compliance" db:"document_compliance"`
	MaintenanceAdherence bool    `json:"maintenance_adherence" db:"maintenance_adherence"`

	// Behavioral Patterns
	NightDrivingHours     float64 `json:"night_driving_hours" db:"night_driving_hours" validate:"gte=0"`
	HighwayPercentage     float64 `json:"highway_percentage" db:"highway_percentage" validate:"gte=0,lte=100"`
	CityDrivingPercentage float64 `json:"city_driving_percentage" db:"city_driving_percentage" validate:"gte=0,lte=100"`

	// Risk Assessment
	RiskLevel        string   `json:"risk_level" db:"risk_level" validate:"required,oneof=low medium high critical"`
	RiskFactors      []string `json:"risk_factors" db:"risk_factors"`
	ImprovementAreas []string `json:"improvement_areas" db:"improvement_areas"`

	// Metadata
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// DriverIncident represents a specific driving incident
type DriverIncident struct {
	ID              string                 `json:"id" db:"id" validate:"required,uuid"`
	DriverID        string                 `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	OrgID           string                 `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	Type            string                 `json:"type" db:"type" validate:"required"`
	Severity        string                 `json:"severity" db:"severity" validate:"required,oneof=low medium high critical"`
	Timestamp       time.Time              `json:"timestamp" db:"timestamp" validate:"required"`
	Location        Location               `json:"location" db:"location" validate:"required"`
	Speed           float64                `json:"speed" db:"speed" validate:"gte=0"`
	Details         map[string]interface{} `json:"details" db:"details"`
	Resolved        bool                   `json:"resolved" db:"resolved"`
	ResolutionNotes string                 `json:"resolution_notes" db:"resolution_notes" validate:"max=500"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// BehaviorReportRequest represents request for behavior reports
type BehaviorReportRequest struct {
	DriverID   *string    `json:"driver_id,omitempty" validate:"omitempty,uuid"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	PeriodType string     `json:"period_type" validate:"required,oneof=daily weekly monthly"`
}

// BehaviorScoreUpdate represents behavior score adjustment
type BehaviorScoreUpdate struct {
	ID          string                 `json:"id" db:"id" validate:"required,uuid"`
	DriverID    string                 `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	ScoreChange int                    `json:"score_change" db:"score_change" validate:"gte=-20,lte=20"`
	Reason      string                 `json:"reason" db:"reason" validate:"required,max=200"`
	EventType   string                 `json:"event_type" db:"event_type" validate:"required"`
	Details     map[string]interface{} `json:"details" db:"details"`
	CreatedBy   string                 `json:"created_by" db:"created_by" validate:"required"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// DriverMessage represents messages between driver and dispatch
type DriverMessage struct {
	ID            string    `json:"id" db:"id" validate:"required,uuid"`
	DriverID      string    `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	TripID        *string   `json:"trip_id" db:"trip_id" validate:"omitempty,uuid"`
	Direction     string    `json:"direction" db:"direction" validate:"required,oneof=INBOUND OUTBOUND"`
	Type          string    `json:"type" db:"type" validate:"required,oneof=TEXT AUDIO IMAGE"`
	Content       string    `json:"content" db:"content" validate:"max=1000"`
	AttachmentURL string    `json:"attachment_url" db:"attachment_url" validate:"omitempty,url"`
	IsRead        bool      `json:"is_read" db:"is_read"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// DriverExpense represents expenses incurred by the driver
type DriverExpense struct {
	ID              string    `json:"id" db:"id" validate:"required,uuid"`
	DriverID        string    `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	TripID          *string   `json:"trip_id" db:"trip_id" validate:"omitempty,uuid"`
	Type            string    `json:"type" db:"type" validate:"required,oneof=FUEL TOLL MAINTENANCE FOOD LODGING OTHER"`
	Amount          float64   `json:"amount" db:"amount" validate:"required,gt=0"`
	Currency        string    `json:"currency" db:"currency" validate:"required,len=3"`
	ReceiptURL      string    `json:"receipt_url" db:"receipt_url" validate:"omitempty,url"`
	Notes           string    `json:"notes" db:"notes" validate:"max=500"`
	Status          string    `json:"status" db:"status" validate:"required,oneof=PENDING APPROVED REJECTED"`
	TransactionDate time.Time `json:"transaction_date" db:"transaction_date" validate:"required"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// DutyStatusLog represents HOOS logs
type DutyStatusLog struct {
	ID              string     `json:"id" db:"id" validate:"required,uuid"`
	DriverID        string     `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	Status          string     `json:"status" db:"status" validate:"required,oneof=OFF_DUTY SLEEPER_BERTH DRIVING ON_DUTY REST BREAK"`
	Location        *Location  `json:"location" db:"location" validate:"required"`
	StartedAt       time.Time  `json:"started_at" db:"started_at" validate:"required"`
	EndedAt         *time.Time `json:"ended_at" db:"ended_at"`
	DurationMinutes *int       `json:"duration_minutes" db:"duration_minutes" validate:"omitempty,gte=0"`
	Remarks         string     `json:"remarks" db:"remarks" validate:"max=200"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// OnboardingStep represents a step in the driver onboarding workflow
type OnboardingStep struct {
	ID          string     `json:"id" db:"id"`
	DriverID    string     `json:"driver_id" db:"driver_id"`
	Step        string     `json:"step" db:"step"`     // APPLICATION, DOCUMENT_VERIFICATION, etc.
	Status      string     `json:"status" db:"status"` // PENDING, COMPLETED, REJECTED
	Remarks     string     `json:"remarks" db:"remarks"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Shift represents a driver's scheduled shift
type Shift struct {
	ID         string    `json:"id" db:"id"`
	DriverID   string    `json:"driver_id" db:"driver_id"`
	ShiftStart time.Time `json:"shift_start" db:"shift_start"`
	ShiftEnd   time.Time `json:"shift_end" db:"shift_end"`
	ShiftType  string    `json:"shift_type" db:"shift_type"` // DAY, NIGHT, ROTATIONAL
	Status     string    `json:"status" db:"status"`         // SCHEDULED, ACTIVE, COMPLETED, CANCELLED
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// TrainingModule represents a training course available for drivers
type TrainingModule struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// TrainingRecord represents a driver's completion of a training module
type TrainingRecord struct {
	ID             string     `json:"id" db:"id"`
	DriverID       string     `json:"driver_id" db:"driver_id"`
	ModuleID       string     `json:"module_id" db:"module_id"`
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`
	CertificateURL string     `json:"certificate_url" db:"certificate_url"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// SalaryConfig represents a driver's salary configuration
type SalaryConfig struct {
	ID             string    `json:"id" db:"id"`
	DriverID       string    `json:"driver_id" db:"driver_id"`
	BaseSalary     float64   `json:"base_salary" db:"base_salary"`
	PerKmRate      float64   `json:"per_km_rate" db:"per_km_rate"`
	DailyAllowance float64   `json:"daily_allowance" db:"daily_allowance"`
	NightAllowance float64   `json:"night_allowance" db:"night_allowance"`
	EffectiveFrom  time.Time `json:"effective_from" db:"effective_from"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// TripEarning represents earnings for a specific trip
type TripEarning struct {
	ID         string    `json:"id" db:"id"`
	DriverID   string    `json:"driver_id" db:"driver_id"`
	TripID     string    `json:"trip_id" db:"trip_id"`
	DistanceKm float64   `json:"distance_km" db:"distance_km"`
	Earnings   float64   `json:"earnings" db:"earnings"`
	Bonus      float64   `json:"bonus" db:"bonus"`
	Penalty    float64   `json:"penalty" db:"penalty"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Payout represents a salary payout to a driver
type Payout struct {
	ID          string     `json:"id" db:"id"`
	DriverID    string     `json:"driver_id" db:"driver_id"`
	PeriodStart time.Time  `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time  `json:"period_end" db:"period_end"`
	TotalAmount float64    `json:"total_amount" db:"total_amount"`
	Status      string     `json:"status" db:"status"` // PENDING, PAID, CANCELLED
	PaidAt      *time.Time `json:"paid_at" db:"paid_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Rating represents a rating given to a driver for a trip
type Rating struct {
	ID               string    `json:"id" db:"id"`
	DriverID         string    `json:"driver_id" db:"driver_id"`
	TripID           string    `json:"trip_id" db:"trip_id"`
	PunctualityScore int       `json:"punctuality_score" db:"punctuality_score"`
	BehaviorScore    int       `json:"behavior_score" db:"behavior_score"`
	DrivingScore     int       `json:"driving_score" db:"driving_score"`
	OverallRating    float64   `json:"overall_rating" db:"overall_rating"`
	Comments         string    `json:"comments" db:"comments"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// SafetyEvent represents a safety-related event captured from telemetry
type SafetyEvent struct {
	ID        string    `json:"id" db:"id"`
	DriverID  string    `json:"driver_id" db:"driver_id"`
	TripID    *string   `json:"trip_id" db:"trip_id"`
	EventType string    `json:"event_type" db:"event_type"` // OVERSPEED, HARSH_BRAKE, PANIC_ALERT, etc.
	Severity  string    `json:"severity" db:"severity"`     // LOW, MEDIUM, HIGH, CRITICAL
	GPSLat    float64   `json:"gps_lat" db:"gps_lat"`
	GPSLng    float64   `json:"gps_lng" db:"gps_lng"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MobileSyncItem represents an item in the mobile sync queue
type MobileSyncItem struct {
	ID         string                 `json:"id" db:"id"`
	DriverID   string                 `json:"driver_id" db:"driver_id"`
	ActionType string                 `json:"action_type" db:"action_type"`
	Payload    map[string]interface{} `json:"payload" db:"payload"`
	Status     string                 `json:"status" db:"status"` // PENDING, SYNCED, FAILED
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// DriverDutyHours represents aggregated duty hours for a driver
type DriverDutyHours struct {
	DriverID               string  `json:"driver_id"`
	ContinuousDrivingHours float64 `json:"continuous_driving_hours"`
	DailyDutyHours         float64 `json:"daily_duty_hours"`
	WeeklyDrivingHours     float64 `json:"weekly_driving_hours"`
}
