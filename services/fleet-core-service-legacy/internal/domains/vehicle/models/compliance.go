package models

import (
	"time"

	"github.com/google/uuid"
)

// VehicleCompliance holds India-regulatory compliance data for a vehicle.
type VehicleCompliance struct {
	VehicleID           uuid.UUID  `json:"vehicle_id"          db:"vehicle_id"`
	RCExpiryDate        *time.Time `json:"rc_expiry_date"       db:"rc_expiry_date"`
	InsuranceExpiryDate *time.Time `json:"insurance_expiry_date" db:"insurance_expiry_date"`
	PUCExpiryDate       *time.Time `json:"puc_expiry_date"      db:"puc_expiry_date"`
	FitnessExpiryDate   *time.Time `json:"fitness_expiry_date"  db:"fitness_expiry_date"`
	PermitExpiryDate    *time.Time `json:"permit_expiry_date"   db:"permit_expiry_date"`
	PermitType          *string    `json:"permit_type"          db:"permit_type"`
	InsurancePolicyNo   *string    `json:"insurance_policy_no"  db:"insurance_policy_no"`
	IsBlacklisted       bool       `json:"is_blacklisted"       db:"is_blacklisted"`
	LegalHold           bool       `json:"legal_hold"           db:"legal_hold"`
	LastVerifiedAt      *time.Time `json:"last_verified_at"     db:"last_verified_at"`
	VerifiedBy          *uuid.UUID `json:"verified_by"          db:"verified_by"`
	CreatedAt           time.Time  `json:"created_at"           db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"           db:"updated_at"`
}

// VerifyComplianceRequest is the request body for POST /vehicles/{id}/verify-compliance.
type VerifyComplianceRequest struct {
	RCExpiryDate        *time.Time `json:"rc_expiry_date"`
	InsuranceExpiryDate *time.Time `json:"insurance_expiry_date"`
	PUCExpiryDate       *time.Time `json:"puc_expiry_date"`
	FitnessExpiryDate   *time.Time `json:"fitness_expiry_date"`
	PermitExpiryDate    *time.Time `json:"permit_expiry_date"`
	PermitType          *string    `json:"permit_type"`
	InsurancePolicyNo   *string    `json:"insurance_policy_no"`
}

// ComplianceExpiryAlert is returned by GET /vehicles/expiring.
type ComplianceExpiryAlert struct {
	VehicleID          uuid.UUID `json:"vehicle_id"`
	TenantID           uuid.UUID `json:"tenant_id"`
	RegistrationNumber string    `json:"registration_number"`
	Status             string    `json:"status"`
	EarliestExpiry     *string   `json:"earliest_expiry"`
	DaysToExpiry       *int      `json:"days_to_expiry"`
}

// ComplianceStatus describes the derived compliance health of a vehicle.
type ComplianceStatus struct {
	IsCompliant         bool     `json:"is_compliant"`
	ExpiredDocuments    []string `json:"expired_documents"`
	ExpiringDocuments   []string `json:"expiring_documents"` // within 30 days
	IsBlacklisted       bool     `json:"is_blacklisted"`
	LegalHold           bool     `json:"legal_hold"`
}
