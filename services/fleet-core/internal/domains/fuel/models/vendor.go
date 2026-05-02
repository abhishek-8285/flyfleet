package models

import (
	"time"

	"github.com/google/uuid"
)

type FuelVendorStatus string

const (
	FuelVendorStatusActive      FuelVendorStatus = "ACTIVE"
	FuelVendorStatusInactive    FuelVendorStatus = "INACTIVE"
	FuelVendorStatusBlacklisted FuelVendorStatus = "BLACKLISTED"
)

type FuelVendor struct {
	ID             uuid.UUID        `json:"id" db:"id"`
	TenantID       uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	Name           string           `json:"name" db:"name"`
	GSTIN          *string          `json:"gstin,omitempty" db:"gstin"`
	Address        *string          `json:"address,omitempty" db:"address"`
	ContactPhone   *string          `json:"contact_phone,omitempty" db:"contact_phone"`
	Email          *string          `json:"email,omitempty" db:"email"`
	Status         FuelVendorStatus `json:"status" db:"status"`
	Rating         float64          `json:"rating" db:"rating"`
	Metadata       map[string]any   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

type CreateFuelVendorRequest struct {
	Name         string  `json:"name" validate:"required"`
	GSTIN        *string `json:"gstin,omitempty"`
	Address      *string `json:"address,omitempty"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	Email        *string `json:"email,omitempty" validate:"omitempty,email"`
}
