package models

import (
	"time"

	"github.com/google/uuid"
)

// VehicleDocument is an immutable document attached to a vehicle (RC, Insurance, etc.).
type VehicleDocument struct {
	ID           uuid.UUID  `json:"id"            db:"id"`
	VehicleID    uuid.UUID  `json:"vehicle_id"    db:"vehicle_id"`
	DocumentType string     `json:"document_type" db:"document_type"` // RC, INSURANCE, PERMIT, FITNESS, PUC, OTHER
	FileURL      string     `json:"file_url"      db:"file_url"`
	Version      int        `json:"version"       db:"version"`
	Hash         string     `json:"hash"          db:"hash"` // SHA-256 hex of file contents
	UploadedAt   time.Time  `json:"uploaded_at"   db:"uploaded_at"`
	UploadedBy   *uuid.UUID `json:"uploaded_by"  db:"uploaded_by"`
}

// UploadDocumentRequest is the request body for uploading a vehicle document.
type UploadDocumentRequest struct {
	DocumentType string `json:"document_type" validate:"required,oneof=RC INSURANCE PERMIT FITNESS PUC OTHER"`
	FileURL      string `json:"file_url"      validate:"required,url"`
	Hash         string `json:"hash"          validate:"required,len=64"` // SHA-256 hex (64 chars)
}
