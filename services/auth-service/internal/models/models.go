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
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Permission struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	Resource  string    `json:"resource" db:"resource"`
	Action    string    `json:"action" db:"action"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Vehicle struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrgID        uuid.UUID  `json:"org_id" db:"org_id"`
	Name         string     `json:"name,omitempty" db:"name"` // Optional, microservice uses Make/Model often
	LicensePlate string     `json:"license_plate" db:"license_plate"`
	Make         string     `json:"make,omitempty" db:"make"`
	Model        string     `json:"model,omitempty" db:"model"`
	Year         int        `json:"year,omitempty" db:"year"`
	FuelType     string     `json:"fuel_type,omitempty" db:"fuel_type"`
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
	FullName        string     `json:"full_name" db:"full_name"`
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
	SerialNumber string     `json:"serial_number" db:"serial_number"`
	Type         string     `json:"type" db:"type"`
	Status       string     `json:"status" db:"status"`
	LastPingAt   *time.Time `json:"last_ping_at,omitempty" db:"last_ping_at"`
	IsDeleted    bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type AuditLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	Details   string    `json:"details,omitempty" db:"details"` // JSON string
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
	ActualDistanceKm     float64                `json:"actual_distance_km" db:"actual_distance_km"`
	ScheduledStart       time.Time              `json:"scheduled_start" db:"scheduled_start"`
	StartedAt            *time.Time             `json:"started_at" db:"started_at"`
	EndedAt              *time.Time             `json:"ended_at" db:"ended_at"`
	StartLocation        map[string]interface{} `json:"start_location" db:"start_location"`
	EndLocation          map[string]interface{} `json:"end_location" db:"end_location"`
	Waypoints            []interface{}          `json:"waypoints" db:"waypoints"`
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
// MFA Security Extensions

type MFAPolicy struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrgID          uuid.UUID `json:"org_id" db:"org_id"`
	RequireMFA     bool      `json:"require_mfa" db:"require_mfa"`
	AllowedMethods []string  `json:"allowed_methods" db:"allowed_methods"`
	MaxAttempts    int       `json:"max_attempts" db:"max_attempts"`
	LockoutMinutes int       `json:"lockout_minutes" db:"lockout_minutes"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
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
