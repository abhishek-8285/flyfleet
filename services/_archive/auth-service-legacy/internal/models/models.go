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
	OrgID          uuid.UUID  `json:"org_id" db:"org_id"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	FirstName      *string    `json:"first_name,omitempty" db:"first_name"`
	LastName       *string    `json:"last_name,omitempty" db:"last_name"`
	Name           *string    `json:"name,omitempty" db:"name"` // Keep for compat
	Role           string     `json:"role" db:"role"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	MFAEnabled     bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret      string     `json:"-" db:"mfa_secret"`
	Status         string     `json:"status" db:"status"`
	RoleID         *uuid.UUID `json:"role_id,omitempty" db:"role_id"`
	FailedLogins   int        `json:"failed_logins" db:"failed_logins"`
	LockedUntil    *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	Verified       bool       `json:"verified" db:"verified"`
	Theme          string     `json:"theme" db:"theme"`
	ThemeUpdatedAt *time.Time `json:"theme_updated_at,omitempty" db:"theme_updated_at"`
	IsDeleted      bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UserFilter struct {
	OrgID    uuid.UUID
	Role     string
	Status   string
	Search   string
	Page     int
	Limit    int
	SortBy   string
	SortOrder string
}

type UserListResponse struct {
	Data []User    `json:"data"`
	Meta UserMeta  `json:"meta"`
}

type UserMeta struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
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
	ID            uuid.UUID   `json:"id" db:"id"`
	VehicleID     *uuid.UUID  `json:"vehicle_id,omitempty" db:"vehicle_id"`
	OrgID         uuid.UUID   `json:"org_id" db:"org_id"`
	Type          string      `json:"type" db:"type"`
	Title         string      `json:"title" db:"title"`
	DocNumber     string      `json:"doc_number" db:"doc_number"`
	FileName      string      `json:"file_name" db:"file_name"`
	FileSizeBytes int         `json:"file_size_bytes" db:"file_size_bytes"`
	MimeType      string      `json:"mime_type" db:"mime_type"`
	StoragePath   string      `json:"-" db:"storage_path"` // Never exposed in API
	Status        string      `json:"status" db:"status"`
	IssueDate     *time.Time  `json:"issued_date,omitempty" db:"issue_date"`
	ExpiryDate    *time.Time  `json:"expiry_date,omitempty" db:"expiry_date"`
	ReminderDays  int         `json:"reminder_days" db:"reminder_days"`
	UploadedBy    string      `json:"uploaded_by" db:"uploaded_by"` // email from join
	UploadedAt    time.Time   `json:"uploaded_at" db:"created_at"`
	Metadata      interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

// DocumentFilter represents query parameters for listing documents
type DocumentFilter struct {
	VehicleID          *uuid.UUID
	Type               string
	Status             string
	ExpiringWithinDays int
	Page               int
	Limit              int
}

// DocumentListResponse represents the paginated response for listing documents
type DocumentListResponse struct {
	Data []Document         `json:"data"`
	Meta DocumentListMeta   `json:"meta"`
}

// DocumentListMeta holds pagination metadata
type DocumentListMeta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
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
	Code      string    `json:"-" db:"-"`           // Used internally for verification, not stored in DB
	CodeHash  string    `json:"-" db:"code_hash"`  // Store hashed code, don't expose in JSON
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

type SystemStats struct {
	TotalUsers     int `json:"total_users"`
	ActiveUsers    int `json:"active_users"`
	TotalDrivers   int `json:"total_drivers"`
	TotalOrgs      int `json:"total_orgs"`
	ActiveOrgs     int `json:"active_orgs"`
	PendingOrgs    int `json:"pending_orgs"`
	InactiveOrgs   int `json:"inactive_orgs"`
}

// SecuritySettings represents security configuration for an organization
type SecuritySettings struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	OrgID                   uuid.UUID `json:"org_id" db:"org_id"`
	PasswordComplexity      PasswordComplexity `json:"password_complexity" db:"password_complexity"`
	SessionTimeoutMinutes   int                `json:"session_timeout_minutes" db:"session_timeout_minutes"`
	MaxConcurrentSessions   int                `json:"max_concurrent_sessions" db:"max_concurrent_sessions"`
	EnableMFA               bool               `json:"enable_mfa" db:"enable_mfa"`
	MFARequiredForAdmins    bool               `json:"mfa_required_for_admins" db:"mfa_required_for_admins"`
	EnableIPWhitelist       bool               `json:"enable_ip_whitelist" db:"enable_ip_whitelist"`
	EnableBruteForceProtection bool           `json:"enable_brute_force_protection" db:"enable_brute_force_protection"`
	BruteForceThreshold     int                `json:"brute_force_threshold" db:"brute_force_threshold"`
	BruteForceLockoutMins   int                `json:"brute_force_lockout_mins" db:"brute_force_lockout_mins"`
	EnableSecurityQuestions bool               `json:"enable_security_questions" db:"enable_security_questions"`
	EnableDeviceTrust       bool               `json:"enable_device_trust" db:"enable_device_trust"`
	EnableRiskBasedAuth     bool               `json:"enable_risk_based_auth" db:"enable_risk_based_auth"`
	RequireEncryptedTraffic bool               `json:"require_encrypted_traffic" db:"require_encrypted_traffic"`
	EnableAuditLogging      bool               `json:"enable_audit_logging" db:"enable_audit_logging"`
	EnableAnomalyDetection  bool               `json:"enable_anomaly_detection" db:"enable_anomaly_detection"`
	CreatedAt               time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time          `json:"updated_at" db:"updated_at"`
}

type PasswordComplexity struct {
	MinLength            int  `json:"min_length"`
	RequireUppercase     bool `json:"require_uppercase"`
	RequireLowercase     bool `json:"require_lowercase"`
	RequireNumbers       bool `json:"require_numbers"`
	RequireSpecialChars  bool `json:"require_special_chars"`
	MaxAgeDays           int  `json:"max_age_days"`
	ReuseHistoryLimit    int  `json:"reuse_history_limit"`
}

// SecurityMetrics represents security metrics for an organization
type SecurityMetrics struct {
	OrgID                    uuid.UUID `json:"org_id"`
	MFARate                  float64   `json:"mfa_rate"`
	PasswordStrengthAvg      float64   `json:"password_strength_avg"`
	FailedLoginRate          float64   `json:"failed_login_rate"`
	AccountLockoutEvents     int       `json:"account_lockout_events"`
	SuspiciousActivityEvents int       `json:"suspicious_activity_events"`
	ComplianceScore          float64   `json:"compliance_score"`
	SecurityPostureScore     float64   `json:"security_posture_score"`
	LastUpdated              time.Time `json:"last_updated"`
}

// VulnerabilityReport represents vulnerability assessment results
type VulnerabilityReport struct {
	OrgID         uuid.UUID         `json:"org_id"`
	ReportDate    time.Time         `json:"report_date"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Summary       VulnerabilitySummary `json:"summary"`
	Recommendations []string        `json:"recommendations"`
}

type Vulnerability struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // critical, high, medium, low
	CVSSScore   float64   `json:"cvss_score"`
	Status      string    `json:"status"` // open, in_progress, resolved, false_positive
	FoundDate   time.Time `json:"found_date"`
	Target      string    `json:"target"`
	RemediatedDate *time.Time `json:"remediated_date,omitempty"`
}

type VulnerabilitySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
	Resolved int `json:"resolved"`
	Open     int `json:"open"`
}

// ComplianceReport represents compliance assessment results
type ComplianceReport struct {
	OrgID       uuid.UUID        `json:"org_id"`
	ReportType  string           `json:"report_type"` // soc2, gdpr, iso27001, etc.
	ReportDate  time.Time        `json:"report_date"`
	Requirements []ComplianceRequirement `json:"requirements"`
	Summary     ComplianceSummary `json:"summary"`
	Findings    []ComplianceFinding `json:"findings"`
	Recommendations []string      `json:"recommendations"`
}

type ComplianceRequirement struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"` // compliant, non_compliant, not_applicable
	Evidence    string `json:"evidence"`
	LastChecked time.Time `json:"last_checked"`
}

type ComplianceSummary struct {
	TotalRequirements int `json:"total_requirements"`
	Compliant       int `json:"compliant"`
	NonCompliant    int `json:"non_compliant"`
	NotApplicable   int `json:"not_applicable"`
	CompliancePercentage float64 `json:"compliance_percentage"`
}

type ComplianceFinding struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // high, medium, low
	Status      string    `json:"status"` // open, in_progress, resolved
	Remediation string    `json:"remediation"`
	FoundDate   time.Time `json:"found_date"`
}
