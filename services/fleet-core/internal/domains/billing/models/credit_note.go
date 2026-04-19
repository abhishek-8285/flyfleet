package models

import (
	"time"

	"github.com/google/uuid"
)

// CreditNote represents a credit note for refunds/adjustments
type CreditNote struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	CreditNoteNo       string     `json:"credit_note_no" db:"credit_note_no"`
	OrgID              uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OriginalInvoiceID  uuid.UUID  `json:"original_invoice_id" db:"original_invoice_id"`
	CustomerID         uuid.UUID  `json:"customer_id" db:"customer_id"`
	CustomerName       string     `json:"customer_name" db:"customer_name"`
	CustomerGSTIN      string     `json:"customer_gstin,omitempty" db:"customer_gstin"`
	CustomerPAN        string     `json:"customer_pan,omitempty" db:"customer_pan"`
	TaxableValue       float64    `json:"taxable_value" db:"taxable_value"`
	CGST               float64    `json:"cgst" db:"cgst"`
	SGST               float64    `json:"sgst" db:"sgst"`
	IGST               float64    `json:"igst" db:"igst"`
	TotalAmount        float64    `json:"total_amount" db:"total_amount"`
	Status             string     `json:"status" db:"status"` // draft, issued, applied, refunded, cancelled
	Reason             string     `json:"reason" db:"reason"`
	Remarks            string     `json:"remarks,omitempty" db:"remarks"`
	AppliedToInvoiceID *uuid.UUID `json:"applied_to_invoice_id,omitempty" db:"applied_to_invoice_id"`
	AppliedAmount      float64    `json:"applied_amount" db:"applied_amount"`
	AppliedAt          *time.Time `json:"applied_at,omitempty" db:"applied_at"`
	IssueDate          time.Time  `json:"issue_date" db:"issue_date"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	CreatedBy          uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy          uuid.UUID  `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreditNoteItem represents a line item in a credit note
type CreditNoteItem struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	CreditNoteID          uuid.UUID  `json:"credit_note_id" db:"credit_note_id"`
	OriginalInvoiceItemID *uuid.UUID `json:"original_invoice_item_id,omitempty" db:"original_invoice_item_id"`
	Description           string     `json:"description" db:"description"`
	HSNCode               string     `json:"hsn_code,omitempty" db:"hsn_code"`
	Quantity              int        `json:"quantity" db:"quantity"`
	UnitPrice             float64    `json:"unit_price" db:"unit_price"`
	TaxPercent            float64    `json:"tax_percent" db:"tax_percent"`
	TaxAmount             float64    `json:"tax_amount" db:"tax_amount"`
	TotalAmount           float64    `json:"total_amount" db:"total_amount"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
}

// CreditNoteStatus defines valid credit note statuses
type CreditNoteStatus string

const (
	CreditNoteDraft     CreditNoteStatus = "draft"
	CreditNoteIssued    CreditNoteStatus = "issued"
	CreditNoteApplied   CreditNoteStatus = "applied"
	CreditNoteRefunded  CreditNoteStatus = "refunded"
	CreditNoteCancelled CreditNoteStatus = "cancelled"
)

// CreditNoteReason defines standard reasons for credit notes
type CreditNoteReason string

const (
	ReasonGoodsReturned    CreditNoteReason = "goods_returned"
	ReasonServicesReturned CreditNoteReason = "services_returned"
	ReasonInvoiceError     CreditNoteReason = "invoice_error"
	ReasonPriceRevision    CreditNoteReason = "price_revision"
	ReasonDiscountOmitted  CreditNoteReason = "discount_omitted"
	ReasonOther            CreditNoteReason = "other"
)

// IsValidReason checks if a credit note reason is valid
func IsValidReason(reason string) bool {
	validReasons := map[string]bool{
		string(ReasonGoodsReturned):    true,
		string(ReasonServicesReturned): true,
		string(ReasonInvoiceError):     true,
		string(ReasonPriceRevision):    true,
		string(ReasonDiscountOmitted):  true,
		string(ReasonOther):            true,
	}
	return validReasons[reason]
}

// IsValidStatus checks if a credit note status is valid
func IsValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"draft": true, "issued": true, "applied": true, "refunded": true, "cancelled": true,
	}
	return validStatuses[status]
}

// CanEdit checks if credit note can be edited
func (cn *CreditNote) CanEdit() bool {
	return cn.Status == string(CreditNoteDraft)
}

// CanApply checks if credit note can be applied to invoice
func (cn *CreditNote) CanApply() bool {
	return cn.Status == string(CreditNoteIssued) && cn.AppliedToInvoiceID == nil
}

// CanCancel checks if credit note can be cancelled
func (cn *CreditNote) CanCancel() bool {
	return cn.Status == string(CreditNoteDraft) || cn.Status == string(CreditNoteIssued)
}
