package models

import (
	"time"

	"github.com/google/uuid"
)

type ExternalSystem string

const (
	SystemQuickBooks ExternalSystem = "QUICKBOOKS"
	SystemXero       ExternalSystem = "XERO"
	SystemSAP        ExternalSystem = "SAP"
)

type ExportType string

const (
	ExportLedger   ExportType = "LEDGER"
	ExportInvoices ExportType = "INVOICES"
	ExportPayments ExportType = "PAYMENTS"
)

type ExportStatus string

const (
	ExportStatusPending   ExportStatus = "PENDING"
	ExportStatusCompleted ExportStatus = "COMPLETED"
	ExportStatusFailed    ExportStatus = "FAILED"
)

type AccountingExport struct {
	ID             uuid.UUID      `json:"id"`
	OrgID          uuid.UUID      `json:"org_id"`
	ExternalSystem ExternalSystem `json:"external_system"`
	ExportType     ExportType     `json:"export_type"`
	Status         ExportStatus   `json:"status"`
	FileKey        string         `json:"file_key,omitempty"`
	RecordsCount   int            `json:"records_count"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	CreatedBy      uuid.UUID      `json:"created_by"`
}

// AccountingExporter defines the interface for specific platform exporters
type AccountingExporter interface {
	GetName() ExternalSystem
	ExportLedger(orgID uuid.UUID, entries []LedgerEntry) ([]byte, string, error)
	ExportInvoices(orgID uuid.UUID, invoices []Invoice) ([]byte, string, error)
}
