package models

import (
	"time"

	"github.com/google/uuid"
)

type DocumentType string

const (
	DocRC        DocumentType = "RC"
	DocFitness   DocumentType = "FITNESS"
	DocPUC       DocumentType = "PUC"
	DocInsurance DocumentType = "INSURANCE"
	DocPermit    DocumentType = "PERMIT"
	DocLicense   DocumentType = "DRIVING_LICENSE"
	DocAadhaar   DocumentType = "AADHAAR"
)

type DocumentStatus string

const (
	DocStatusActive  DocumentStatus = "ACTIVE"
	DocStatusExpired DocumentStatus = "EXPIRED"
	DocStatusRevoked DocumentStatus = "REVOKED"
)

type VehicleDocument struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	VehicleID      uuid.UUID      `json:"vehicle_id" db:"vehicle_id"`
	DocumentType   DocumentType   `json:"document_type" db:"document_type"`
	DocumentNumber string         `json:"document_number" db:"document_number"`
	IssueDate      *time.Time     `json:"issue_date" db:"issue_date"`
	ExpiryDate     time.Time      `json:"expiry_date" db:"expiry_date"`
	FileURL        string         `json:"file_url,omitempty" db:"file_url"`
	Status         DocumentStatus `json:"status" db:"status"`
	Verified       bool           `json:"verified" db:"verified"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type DriverDocument struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	DriverID       uuid.UUID      `json:"driver_id" db:"driver_id"`
	DocumentType   DocumentType   `json:"document_type" db:"document_type"`
	DocumentNumber string         `json:"document_number" db:"document_number"`
	IssueDate      *time.Time     `json:"issue_date" db:"issue_date"`
	ExpiryDate     time.Time      `json:"expiry_date" db:"expiry_date"`
	FileURL        string         `json:"file_url,omitempty" db:"file_url"`
	Status         DocumentStatus `json:"status" db:"status"`
	Verified       bool           `json:"verified" db:"verified"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type DocumentExpiryAlert struct {
	DocumentID   uuid.UUID    `json:"document_id"`
	VehicleID    uuid.UUID    `json:"vehicle_id"`
	DocumentType DocumentType `json:"document_type"`
	DaysToExpiry int          `json:"days_to_expiry"`
	IsExpired    bool         `json:"is_expired"`
}
