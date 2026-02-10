package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Domain    string    `json:"domain,omitempty" db:"domain"`
	DefaultTheme string    `json:"default_theme" db:"default_theme"`
	Status       string    `json:"status" db:"status"`
	IsDeleted    bool      `json:"is_deleted" db:"is_deleted"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	Name           string     `json:"name" db:"name"`
	Role           string     `json:"role" db:"-"`
	OrgID          uuid.UUID  `json:"org_id" db:"-"`
	Phone          string     `json:"phone,omitempty" db:"phone"`
	MFAEnabled     bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret      string     `json:"-" db:"mfa_secret"`
	Status         string     `json:"status" db:"status"`
	RoleID         *uuid.UUID `json:"role_id,omitempty" db:"role_id"`
	FailedLogins   int        `json:"failed_logins" db:"failed_logins"`
	LockedUntil    *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	Verified       bool       `json:"verified" db:"verified"`
	Theme          string     `json:"theme" db:"theme"`
	ThemeUpdatedAt *time.Time `json:"theme_updated_at,omitempty" db:"theme_updated_at"`
	IsDeleted      bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type Role struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description,omitempty" db:"description"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	IsSystem       bool      `json:"is_system" db:"is_system"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type PermissionDefinition struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Module      string    `json:"module" db:"module"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description,omitempty" db:"description"`
}

type RolePermissionMapping struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
}

type UserRoleAssignment struct {
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	RoleID     uuid.UUID `json:"role_id" db:"role_id"`
	AssignedBy uuid.UUID `json:"assigned_by" db:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

type Vehicle struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrgID        uuid.UUID  `json:"org_id" db:"org_id"`
	CurrentDriverID *uuid.UUID `json:"current_driver_id,omitempty" db:"current_driver_id"`
	Name         string     `json:"name,omitempty" db:"name"` // Optional, microservice uses Make/Model often
	LicensePlate string     `json:"registration_number" db:"license_plate"`
	Make         string     `json:"make,omitempty" db:"make"`
	Model        string     `json:"model,omitempty" db:"model"`
	Year         int        `json:"year,omitempty" db:"year"`
	FuelType     string     `json:"fuel_type,omitempty" db:"fuel_type"`
	VehicleNo    string     `json:"vehicle_no,omitempty" db:"vehicle_no"`
	VIN          string     `json:"vin,omitempty" db:"vin"`
	Odometer     int64      `json:"odometer,omitempty" db:"odometer"`
	ServiceDue   *time.Time `json:"service_due,omitempty" db:"service_due"`
	Notes        string     `json:"notes,omitempty" db:"notes"`
	Status       string     `json:"status" db:"status"`
	IsDeleted    bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type Driver struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrgID           uuid.UUID  `json:"org_id" db:"org_id"`
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	FullName        string     `json:"name" db:"full_name"`
	Email           string     `json:"email" db:"email"`
	Phone           string     `json:"phone" db:"phone"`
	LicenseNumber   string     `json:"license_number,omitempty" db:"license_number"`
	Status          string     `json:"status" db:"status"`
	Rating          float64    `json:"rating,omitempty" db:"rating"`
	TotalTrips      int        `json:"total_trips,omitempty" db:"total_trips"`
	IsDeleted       bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type Device struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrgID        uuid.UUID  `json:"org_id" db:"org_id"`
	VehicleID    *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	SerialNumber string     `json:"serial_number" db:"serial_number"`
	Type         string     `json:"type" db:"type"`
	Status       string     `json:"status" db:"status"`
	LastPingAt   *time.Time `json:"last_ping_at,omitempty" db:"last_ping_at"`
	IsDeleted    bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Document represents a compliance document associated with a vehicle
type Document struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	VehicleID   uuid.UUID       `json:"vehicle_id" db:"vehicle_id"`
	Type        string          `json:"type" db:"type"`
	DocNumber   string          `json:"doc_number" db:"doc_number"`
	IssueDate   time.Time       `json:"issue_date" db:"issue_date"`
	ExpiryDate  time.Time       `json:"expiry_date" db:"expiry_date"`
	FileURL     string          `json:"file_url,omitempty" db:"file_url"`
	Status      string          `json:"status" db:"status"`
	Metadata    interface{} 	`json:"metadata,omitempty" db:"metadata"` // Changed from json.RawMessage to interface{} to avoid import issues if any, or verify json import.
	// Wait, models.go does NOT import encoding/json based on my previous read (lines 1-8). 
	// I need to check if I can add the import.
	// Actually, line 116 uses `Details string` for JSON, and line 139 `Metadata string`.
	// Line 186 uses `map[string]interface{}`.
	// I should probably use `map[string]interface{}` for Metadata to be consistent with `Trip` struct (line 186) or `string` if it's raw JSON text.
	// `documents` table definition said JSONB. 
	// `Trip` uses `map[string]interface{}`. Use that.
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// DocumentMetadata defines structure for JSONB metadata
type DocumentMetadata struct {
	IssuingAuthority string  `json:"issuing_authority,omitempty"`
	CoverageAmount   float64 `json:"coverage_amount,omitempty"`
	Notes            string  `json:"notes,omitempty"`
}

type AuditLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Module      string    `json:"module" db:"module"`
	Action      string    `json:"action" db:"action"`
	Resource    string    `json:"resource" db:"resource"`
	EntityID    string    `json:"entity_id,omitempty" db:"entity_id"`
	Severity    string    `json:"severity" db:"severity"`
	Status      string    `json:"status" db:"status"`
	Details     string    `json:"details,omitempty" db:"details"`
	BeforeState string    `json:"before_state,omitempty" db:"before_state"`
	AfterState  string    `json:"after_state,omitempty" db:"after_state"`
	Metadata    string    `json:"metadata,omitempty" db:"metadata"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type EntityAuditHistory struct {
	ID         uuid.UUID `json:"id" db:"id"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   string    `json:"entity_id" db:"entity_id"`
	ChangeData string    `json:"change_data" db:"change_data"` // JSON string
	Actor      uuid.UUID `json:"actor" db:"actor"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
}

type EvidenceExport struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrgID           uuid.UUID  `json:"org_id" db:"org_id"`
	Module          string     `json:"module" db:"module"`
	DateRangeStart  time.Time  `json:"date_range_start" db:"date_range_start"`
	DateRangeEnd    time.Time  `json:"date_range_end" db:"date_range_end"`
	Format          string     `json:"format" db:"format"`
	Status          string     `json:"status" db:"status"`
	RequestedBy     uuid.UUID  `json:"requested_by" db:"requested_by"`
	FileURL         string     `json:"file_url,omitempty" db:"file_url"`
	ErrorMessage    string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type AuditPolicy struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrgID           uuid.UUID `json:"org_id" db:"org_id"`
	RetentionYears  int       `json:"retention_years" db:"retention_years"`
	ArchivePolicy   string    `json:"archive_policy" db:"archive_policy"`
	MaskingRules    []string  `json:"masking_rules" db:"masking_rules"`
	UpdatedBy       uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}


type AuditSettings struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	OrgID                 uuid.UUID `json:"org_id" db:"org_id"`
	RetentionDaysOnline   int       `json:"retention_days_online" db:"retention_days_online"`
	RetentionDaysArchival int       `json:"retention_days_archival" db:"retention_days_archival"`
	MaskingRules          []string  `json:"masking_rules" db:"masking_rules"`
	TrackedEventTypes     []string  `json:"tracked_event_types" db:"tracked_event_types"`
	AlertOnCritical       bool      `json:"alert_on_critical" db:"alert_on_critical"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

type SecurityEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	EventType string    `json:"event_type" db:"event_type"` // e.g., "LOGIN_FAILED", "LOGIN_SUCCESS", "MFA_VERIFIED", "ACCOUNT_LOCKED"
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Metadata  string    `json:"metadata,omitempty" db:"metadata"` // JSON string for extra info
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserOrganization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	OrgID     uuid.UUID `json:"org_id" db:"org_id"`
	Role      string    `json:"role" db:"role"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AuthToken struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	DeviceID     string    `json:"device_id,omitempty" db:"device_id"`
	IPAddress    string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string    `json:"user_agent,omitempty" db:"user_agent"`
	LastActive   time.Time `json:"last_active" db:"last_active"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type MFACode struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Code      string    `json:"code" db:"code"`
	Method    string    `json:"method" db:"method"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Monitoring & Operations

type Trip struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	OrgID                uuid.UUID              `json:"org_id" db:"org_id"`
	VehicleID            uuid.UUID              `json:"vehicle_id" db:"vehicle_id"`
	DriverID             uuid.UUID              `json:"driver_id" db:"driver_id"`
	Status               string                 `json:"status" db:"status"`
	Title                string                 `json:"title" db:"title"`
	Description          string                 `json:"description" db:"description"`
	EstimatedDistanceKm  float64                `json:"estimated_distance_km" db:"estimated_distance_km"`
	DistanceKm           float64                `json:"distance_km" db:"actual_distance_km"`
	FuelConsumedLiters   float64                `json:"fuel_consumed_liters" db:"fuel_consumed_liters"`
	ScheduledStart       time.Time              `json:"scheduled_start" db:"scheduled_start"`
	StartTime            *time.Time             `json:"start_time" db:"started_at"`
	EndTime              *time.Time             `json:"end_time" db:"ended_at"`
	StartLocation        map[string]interface{} `json:"start_location" db:"start_location"`
	EndLocation          map[string]interface{} `json:"end_location" db:"end_location"`
	Waypoints            []interface{}          `json:"waypoints" db:"waypoints"`
	Route                []interface{}          `json:"route,omitempty" db:"actual_route"`
	Events               []interface{}          `json:"events" db:"events"`
	IsDeleted            bool                   `json:"is_deleted" db:"is_deleted"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
}

type Alert struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	OrgID     uuid.UUID              `json:"org_id" db:"org_id"`
	Type      string                 `json:"type" db:"type"`
	Severity  string                 `json:"severity" db:"severity"`
	VehicleID *uuid.UUID             `json:"vehicle_id,omitempty" db:"vehicle_id"`
	DriverID  *uuid.UUID             `json:"driver_id,omitempty" db:"driver_id"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
	Location  map[string]interface{} `json:"location" db:"location"`
	Status    string                 `json:"status" db:"status"`
	Details   map[string]interface{} `json:"details" db:"details"`
}

type AlertRule struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	OrgID      uuid.UUID              `json:"org_id" db:"org_id"`
	Name       string                 `json:"name" db:"name"`
	Type       string                 `json:"type" db:"type"`
	Severity   string                 `json:"severity" db:"severity"`
	Conditions map[string]interface{} `json:"conditions" db:"conditions"`
	Enabled    bool                   `json:"enabled" db:"enabled"`
}

type Geofence struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	OrgID     uuid.UUID              `json:"org_id" db:"org_id"`
	Name      string                 `json:"name" db:"name"`
	Type      string                 `json:"type" db:"type"`
	Geometry  map[string]interface{} `json:"geometry" db:"geometry"` // Generic bucket for coordinates/radius
	Enabled   bool                   `json:"enabled" db:"enabled"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
	MFAWait      bool   `json:"mfa_wait,omitempty"`
}

type RegisterRequest struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	OrgID    uuid.UUID `json:"org_id"`
	OrgName  string    `json:"org_name"`
	OrgType  string    `json:"org_type"`
	Role     string    `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
// MFA Security Extensions

type MFAPolicy struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	OrgID                uuid.UUID `json:"org_id" db:"org_id"`
	RequireMFA           bool      `json:"require_mfa" db:"require_mfa"`
	AllowedMethods       []string  `json:"allowed_methods" db:"allowed_methods"`
	MaxAttempts          int       `json:"max_attempts" db:"max_attempts"`
	LockoutMinutes       int       `json:"lockout_minutes" db:"lockout_minutes"`
	GracePeriodDays      int       `json:"grace_period_days" db:"grace_period_days"`
	RememberDeviceDays   int       `json:"remember_device_days" db:"remember_device_days"`
	EnforceAtLogin       bool      `json:"enforce_at_login" db:"enforce_at_login"`
	AdminOnlyMode        bool      `json:"admin_only_mode" db:"admin_only_mode"`
	RememberDeviceAllowed bool      `json:"remember_device_allowed" db:"remember_device_allowed"`
	MFAChallengeTimeout  int       `json:"mfa_challenge_timeout" db:"mfa_challenge_timeout"`
	Version              int       `json:"version" db:"version"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type MFAOrgOverride struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	OrgID              uuid.UUID `json:"org_id" db:"org_id"`
	Enabled            bool      `json:"enabled" db:"enabled"`
	AllowedMethods     []string  `json:"allowed_methods" db:"allowed_methods"`
	MaxAttempts        int       `json:"max_attempts" db:"max_attempts"`
	LockoutMinutes     int       `json:"lockout_minutes" db:"lockout_minutes"`
	RememberDeviceDays int       `json:"remember_device_days" db:"remember_device_days"`
	GracePeriodDays    int       `json:"grace_period_days" db:"grace_period_days"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type MFAPolicyHistory struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PolicyID     uuid.UUID `json:"policy_id" db:"policy_id"`
	OrgID        uuid.UUID `json:"org_id" db:"org_id"`
	Snapshot     string    `json:"snapshot" db:"snapshot"` // JSON string
	ChangedBy    uuid.UUID `json:"changed_by" db:"changed_by"`
	ChangeReason string    `json:"change_reason" db:"change_reason"`
	Version      int       `json:"version" db:"version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type MFAEvent struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	UserID    uuid.UUID      `json:"user_id" db:"user_id"`
	OrgID     uuid.UUID      `json:"org_id" db:"org_id"`
	EventType string         `json:"event_type" db:"event_type"`
	Status    string         `json:"status" db:"status"`
	Method    string         `json:"method" db:"method"`
	IPAddress string         `json:"ip_address" db:"ip_address"`
	UserAgent string         `json:"user_agent" db:"user_agent"`
	Location  string         `json:"location" db:"location"`
	Details   map[string]any `json:"details" db:"details"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

type MFAFailure struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	OrgID         uuid.UUID `json:"org_id" db:"org_id"`
	Method        string    `json:"method" db:"method"`
	FailureReason string    `json:"failure_reason" db:"failure_reason"`
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	Location      string    `json:"location" db:"location"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type MFAException struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	OrgID     uuid.UUID  `json:"org_id" db:"org_id"`
	Reason    string     `json:"reason" db:"reason"`
	GrantedBy uuid.UUID  `json:"granted_by" db:"granted_by"`
	StartDate time.Time  `json:"start_date" db:"start_date"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type MFALockout struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	OrgID      uuid.UUID  `json:"org_id" db:"org_id"`
	Reason     string     `json:"reason" db:"reason"`
	LockedAt   time.Time  `json:"locked_at" db:"locked_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	UnlockedAt *time.Time `json:"unlocked_at" db:"unlocked_at"`
	UnlockedBy *uuid.UUID `json:"unlocked_by,omitempty" db:"unlocked_by"`
}

type MFADashboardSummary struct {
	TotalUsers        int `json:"total_users"`
	MFAEnabled        int `json:"mfa_enabled"`
	MFAPending        int `json:"mfa_pending"`
	MFADisabled       int `json:"mfa_disabled"`
	FailedAttempts24h int `json:"failed_attempts_24h"`
	LockedAccounts    int `json:"locked_accounts"`
}
