package models

import (
	"time"

	"github.com/google/uuid"
)

type ComponentWarranty struct {
	ID               uuid.UUID `json:"id" db:"id"`
	OrganizationID   uuid.UUID `json:"organization_id" db:"organization_id"`
	VehicleID        uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	PartNumber       string    `json:"part_number" db:"part_number"`
	WarrantyProvider string    `json:"warranty_provider" db:"warranty_provider"`
	WarrantyStart    time.Time `json:"warranty_start" db:"warranty_start"`
	WarrantyEnd      time.Time `json:"warranty_end" db:"warranty_end"`
	WarrantyTerms    string    `json:"warranty_terms" db:"warranty_terms"`
	ClaimContact     string    `json:"claim_contact" db:"claim_contact"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type CreateWarrantyRequest struct {
	VehicleID        uuid.UUID `json:"vehicle_id" validate:"required"`
	PartNumber       string    `json:"part_number" validate:"required"`
	WarrantyProvider string    `json:"warranty_provider"`
	WarrantyStart    time.Time `json:"warranty_start" validate:"required"`
	WarrantyEnd      time.Time `json:"warranty_end" validate:"required"`
	WarrantyTerms    string    `json:"warranty_terms"`
	ClaimContact     string    `json:"claim_contact"`
}
