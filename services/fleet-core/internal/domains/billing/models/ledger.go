package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction types
const (
	TxTypeRefund       = "REFUND"
	TxTypeManualCredit = "MANUAL_CREDIT"
	TxTypeDebitInvoice = "DEBIT_FOR_INVOICE"
	TxTypePromotional  = "PROMOTIONAL"
	TxTypePrepaidTopup = "PREPAID_TOPUP"
)

// Ledger entry types
const (
	LedgerTypeInvoiceIssued     = "INVOICE_ISSUED"
	LedgerTypePaymentReceived   = "PAYMENT_RECEIVED"
	LedgerTypeCreditApplied     = "CREDIT_APPLIED"
	LedgerTypeRefundProcessed   = "REFUND_PROCESSED"
	LedgerTypeRevenueRecognized = "REVENUE_RECOGNIZED"
)

// Ledger Accounts
const (
	AccountAccountsReceivable = "ACCOUNTS_RECEIVABLE"
	AccountDeferredRevenue    = "DEFERRED_REVENUE"
	AccountCash               = "CASH"
	AccountRevenue            = "REVENUE"
	AccountCustomerCredits    = "CUSTOMER_CREDITS"
)

// CustomerBalance represents a prepaid wallet or promotional credit balance for an Organization.
type CustomerBalance struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrgID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Currency    string    `json:"currency" db:"currency"`
	Balance     float64   `json:"balance" db:"balance"`
	CreditLimit float64   `json:"credit_limit" db:"credit_limit"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreditTransaction represents a single credit or debit operation against a CustomerBalance.
type CreditTransaction struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	BalanceID       uuid.UUID  `json:"balance_id" db:"balance_id"`
	TransactionType string     `json:"transaction_type" db:"transaction_type"`
	Amount          float64    `json:"amount" db:"amount"` // Positive adds to balance, negative deducts
	Currency        string     `json:"currency" db:"currency"`
	ReferenceType   *string    `json:"reference_type" db:"reference_type"`
	ReferenceID     *uuid.UUID `json:"reference_id" db:"reference_id"`
	Description     *string    `json:"description" db:"description"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CreatedBy       *uuid.UUID `json:"created_by" db:"created_by"`
}

// LedgerEntry represents a single, immutable journal entry for Double-Entry Accounting tracking.
type LedgerEntry struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrgID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	EntryType     string     `json:"entry_type" db:"entry_type"`
	Amount        float64    `json:"amount" db:"amount"`
	Currency      string     `json:"currency" db:"currency"`
	InvoiceID     *uuid.UUID `json:"invoice_id" db:"invoice_id"`
	PaymentID     *uuid.UUID `json:"payment_id" db:"payment_id"`
	CreditNoteID  *uuid.UUID `json:"credit_note_id" db:"credit_note_id"`
	LedgerAccount string     `json:"ledger_account" db:"ledger_account"`
	Description   *string    `json:"description" db:"description"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// RevenueSchedule tracks the amortization of upfront billed items (like Annual CRM plans).
type RevenueSchedule struct {
	ID                uuid.UUID `json:"id" db:"id"`
	OrgID             uuid.UUID `json:"tenant_id" db:"tenant_id"`
	InvoiceLineItemID uuid.UUID `json:"invoice_line_item_id" db:"invoice_line_item_id"`
	TotalAmount       float64   `json:"total_amount" db:"total_amount"`
	Currency          string    `json:"currency" db:"currency"`
	StartDate         time.Time `json:"start_date" db:"start_date"`
	EndDate           time.Time `json:"end_date" db:"end_date"`
	Status            string    `json:"status" db:"status"` // PENDING, IN_PROGRESS, COMPLETED
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// RecognizedRevenue is a single installment of recognized revenue against a RevenueSchedule.
type RecognizedRevenue struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ScheduleID   uuid.UUID `json:"schedule_id" db:"schedule_id"`
	Amount       float64   `json:"amount" db:"amount"`
	RecognizedAt time.Time `json:"recognized_at" db:"recognized_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
