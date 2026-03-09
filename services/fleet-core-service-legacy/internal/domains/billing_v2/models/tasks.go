package models

import (
	"github.com/google/uuid"
)

const (
	// TypeGenerateInvoice defines the task type for monthly invoice generation
	TypeGenerateInvoice = "billing:generate_invoice"
	
	// TypeProcessDunning defines the task type for payment retry and notifications
	TypeProcessDunning = "billing:process_dunning"
	
	// TypeUsageSnapshot defines the task type for periodic usage aggregation
	TypeUsageSnapshot = "billing:usage_snapshot"

	// TypeRenderPDF defines the task type for asynchronous PDF generation
	TypeRenderPDF = "billing:render_pdf"
)

// GenerateInvoicePayload represents the payload for TypeGenerateInvoice
type GenerateInvoicePayload struct {
	OrgID uuid.UUID `json:"org_id"`
}

// ProcessDunningPayload represents the payload for TypeProcessDunning
type ProcessDunningPayload struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
	RetryCount int      `json:"retry_count"`
}

// RenderPDFPayload represents the payload for TypeRenderPDF
type RenderPDFPayload struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
}
