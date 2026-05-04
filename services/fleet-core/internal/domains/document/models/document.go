package models

import (
	"time"

	"github.com/google/uuid"
)

// Document Types
const (
	DocTypeInsurance = "INSURANCE"
	DocTypePermit    = "PERMIT"
	DocTypeLicense   = "LICENSE"
	DocTypeAadhaar   = "AADHAAR"
	DocTypePAN       = "PAN"
	DocTypeRC        = "RC_BOOK"
	DocTypeFitness   = "FITNESS"
	DocTypePollution = "POLLUTION"
	DocTypeMedical   = "MEDICAL"
	DocTypePolice    = "POLICE_VERIFICATION"
)

// Document Status
const (
	DocStatusPending  = "PENDING"
	DocStatusVerified = "VERIFIED"
	DocStatusRejected = "REJECTED"
)

// Document represents a file associated with an entity in the system.
type Document struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	EntityType   string    `json:"entity_type" db:"entity_type"` // 'VEHICLE', 'DRIVER'
	EntityID     uuid.UUID `json:"entity_id" db:"entity_id"`
	
	DocumentType string    `json:"document_type" db:"document_type"`
	FileName     string    `json:"file_name" db:"file_name"`
	FilePath     string    `json:"file_path" db:"file_path"`
	FileSize     int64     `json:"file_size" db:"file_size_bytes"`
	MimeType     string    `json:"mime_type" db:"mime_type"`
	
	ExpiryDate   *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
	Status       string     `json:"status" db:"status"`
	Remarks      *string    `json:"remarks,omitempty" db:"remarks"`
	
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	UploadedBy   *uuid.UUID             `json:"uploaded_by,omitempty" db:"uploaded_by"`
	
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UploadDocumentRequest is the payload for starting an upload
type UploadDocumentRequest struct {
	EntityType   string     `json:"entity_type" validate:"required"`
	EntityID     uuid.UUID  `json:"entity_id" validate:"required"`
	DocumentType string     `json:"document_type" validate:"required"`
	ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
}
