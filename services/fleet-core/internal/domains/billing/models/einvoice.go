package models

import (
	"time"

	"github.com/google/uuid"
)

// EInvoiceIRN represents e-invoice IRN details
type EInvoiceIRN struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	InvoiceID          uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	TenantID              uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	IRN                string     `json:"irn,omitempty" db:"irn"`
	IRNStatus          string     `json:"irn_status" db:"irn_status"` // pending, generated, cancelled, failed
	AckNo              string     `json:"ack_no,omitempty" db:"ack_no"`
	AckDt              *time.Time `json:"ack_dt,omitempty" db:"ack_dt"`
	SignedResponse     string     `json:"signed_response,omitempty" db:"signed_response"`
	QRCodeStr          string     `json:"qr_code_str,omitempty" db:"qr_code_str"`
	QRCodeImage        []byte     `json:"-" db:"qr_code_image"` // Binary image data
	Cancelled          bool       `json:"cancelled" db:"cancelled"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancellationReason string     `json:"cancellation_reason,omitempty" db:"cancellation_reason"`
	RetryCount         int        `json:"retry_count" db:"retry_count"`
	LastRetryAt        *time.Time `json:"last_retry_at,omitempty" db:"last_retry_at"`
	ErrorMessage       string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	CreatedBy          uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy          uuid.UUID  `json:"updated_by,omitempty" db:"updated_by"`
}

// EInvoiceConfig represents e-invoice configuration
type EInvoiceConfig struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TenantID          uuid.UUID `json:"tenant_id" db:"tenant_id"`
	AppKey         string    `json:"-" db:"app_key"`    // Never expose
	SecretKey      string    `json:"-" db:"secret_key"` // Never expose
	Username       string    `json:"-" db:"username"`   // Never expose
	Password       string    `json:"-" db:"password"`   // Never expose
	IsEnabled      bool      `json:"is_enabled" db:"is_enabled"`
	Environment    string    `json:"environment" db:"environment"` // sandbox, production
	GSTIN          string    `json:"gstin" db:"gstin"`
	MandatoryAbove float64   `json:"mandatory_above" db:"mandatory_above"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// EInvoiceItemMapping represents HSN/SAC code mapping
type EInvoiceItemMapping struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TenantID           uuid.UUID `json:"tenant_id" db:"tenant_id"`
	ProductID       uuid.UUID `json:"product_id,omitempty" db:"product_id"`
	HSNCode         string    `json:"hsn_code" db:"hsn_code"`
	SACCode         string    `json:"sac_code,omitempty" db:"sac_code"`
	NICCode         string    `json:"nic_code,omitempty" db:"nic_code"`
	GSTRate         float64   `json:"gst_rate" db:"gst_rate"`
	ExemptionReason string    `json:"exemption_reason,omitempty" db:"exemption_reason"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// IRNStatus defines valid IRN statuses
type IRNStatus string

const (
	IRNStatusPending   IRNStatus = "pending"
	IRNStatusGenerated IRNStatus = "generated"
	IRNStatusCancelled IRNStatus = "cancelled"
	IRNStatusFailed    IRNStatus = "failed"
)

// EInvoiceEnvironment defines e-invoice environments
type EInvoiceEnvironment string

const (
	EInvoiceSandbox    EInvoiceEnvironment = "sandbox"
	EInvoiceProduction EInvoiceEnvironment = "production"
)

// IsValidIRNStatus checks if IRN status is valid
func IsValidIRNStatus(status string) bool {
	switch IRNStatus(status) {
	case IRNStatusPending, IRNStatusGenerated, IRNStatusCancelled, IRNStatusFailed:
		return true
	}
	return false
}

// IsValidEnvironment checks if environment is valid
func IsValidEnvironment(env string) bool {
	switch EInvoiceEnvironment(env) {
	case EInvoiceSandbox, EInvoiceProduction:
		return true
	}
	return false
}

// DefaultEInvoiceConfig returns default e-invoice configuration
func DefaultEInvoiceConfig(tenantID uuid.UUID) *EInvoiceConfig {
	return &EInvoiceConfig{
		ID:             uuid.New(),
		TenantID:          tenantID,
		IsEnabled:      false,
		Environment:    string(EInvoiceSandbox),
		MandatoryAbove: 50000, // ₹50,000
		IsActive:       true,
	}
}
