package models

import (
	"time"

	"github.com/google/uuid"
)

// VerificationStatus defines document or identity approval state.
type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "PENDING"
	VerificationApproved VerificationStatus = "APPROVED"
	VerificationRejected VerificationStatus = "REJECTED"
)

// DocumentType defines categorized regulatory documents.
type DocumentType string

const (
	DocInsurance DocumentType = "INSURANCE"
	DocPUC       DocumentType = "PUC"
	DocPermit    DocumentType = "PERMIT"
	DocFitness   DocumentType = "FITNESS"
	DocTax       DocumentType = "TAX"
)

// VehicleDocument represents a regulatory compliance record.
type VehicleDocument struct {
	ID             uuid.UUID          `json:"id"               db:"id"`
	VehicleID      uuid.UUID          `json:"vehicle_id"       db:"vehicle_id"`
	Type           DocumentType       `json:"type"             db:"type"`
	DocumentNumber string             `json:"document_number"  db:"document_number"`
	ExpiryDate     time.Time          `json:"expiry_date"      db:"expiry_date"`
	Status         VerificationStatus `json:"status"           db:"status"`
	Version        int                `json:"version"          db:"version"` // Support for document versioning
	Hash           string             `json:"hash"             db:"hash"`    // SHA-256 for integrity
	DocumentURL    *string            `json:"document_url,omitempty" db:"document_url"`
	UploadedBy     *uuid.UUID         `json:"uploaded_by,omitempty" db:"uploaded_by"`
	VerifiedAt     *time.Time         `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy     *uuid.UUID         `json:"verified_by,omitempty" db:"verified_by"`
	CreatedAt      time.Time          `json:"created_at"       db:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"       db:"updated_at"`
	DeletedAt      *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UploadDocumentRequest is the request body for uploading a vehicle document.
type UploadDocumentRequest struct {
	Type           DocumentType `json:"type"            validate:"required"`
	DocumentNumber string       `json:"document_number" validate:"required"`
	ExpiryDate     time.Time    `json:"expiry_date"     validate:"required"`
	DocumentURL    string       `json:"document_url"    validate:"required,url"`
	Hash           string       `json:"hash"            validate:"required,len=64"`
}
