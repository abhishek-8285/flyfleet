package models

import (
	"time"

	"github.com/google/uuid"
)

// IP Access Control

type IPRule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	CIDR        string    `json:"cidr" db:"cidr"`
	Action      string    `json:"action" db:"action"` // ALLOW, BLOCK
	Description string    `json:"description" db:"description"`
	AddedBy     uuid.UUID `json:"added_by" db:"added_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Access Review Workflow

type AccessReview struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	OrgID          uuid.UUID   `json:"org_id" db:"org_id"`
	Title          string      `json:"title" db:"title"`
	Status         string      `json:"status" db:"status"` // PENDING, IN_PROGRESS, COMPLETED
	DueDate        time.Time   `json:"due_date" db:"due_date"`
	TotalUsers     int         `json:"total_users" db:"total_users"`
	ReviewedUsers  int         `json:"reviewed_users" db:"reviewed_users"`
	CreatedBy      uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	ReviewItems    []ReviewItem `json:"review_items,omitempty" db:"-"`
}

type ReviewItem struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ReviewID       uuid.UUID `json:"review_id" db:"review_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	RoleID         uuid.UUID `json:"role_id" db:"role_id"`
	CurrentStatus  string    `json:"current_status" db:"current_status"` // APPROVED, REVOKED, PENDING
	ReviewerID     *uuid.UUID `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	Notes          string    `json:"notes" db:"notes"`
	User           *User     `json:"user,omitempty" db:"-"`
	Role           *Role     `json:"role,omitempty" db:"-"`
}

// Security Events & Monitoring

type LoginEvent struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     uuid.UUID  `json:"org_id" db:"org_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Email     string     `json:"email" db:"email"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	Device    string     `json:"device" db:"device"`
	Browser   string     `json:"browser" db:"browser"`
	Location  string     `json:"location" db:"location"`
	Status    string     `json:"status" db:"status"` // SUCCESS, FAILED, CHALLENGED
	RiskLevel string     `json:"risk_level" db:"risk_level"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type ThreatAlert struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	OrgID       uuid.UUID  `json:"org_id" db:"org_id"`
	Type        string     `json:"type" db:"type"`         // BRUTE_FORCE, GEO_ANOMALY, etc.
	Severity    string     `json:"severity" db:"severity"` // critical, warning, info
	IP          string     `json:"ip" db:"ip"`
	UserID      *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Description string     `json:"description" db:"description"`
	Status      string     `json:"status" db:"status"` // open, acknowledged, resolved, blocked
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type MFAUser struct {
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	OrgID               uuid.UUID  `json:"org_id" db:"org_id"`
	Status              string     `json:"status" db:"status"` // ENABLED, DISABLED, ENFORCED
	Method              string     `json:"method" db:"method"`
	TrustedDevicesCount int        `json:"trusted_devices_count" db:"trusted_devices_count"`
	LastReset           *time.Time `json:"last_reset,omitempty" db:"last_reset"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}
