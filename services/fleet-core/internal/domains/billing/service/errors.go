package service

import "errors"

// Invoice service errors
var (
	ErrInvoiceNotFound                = errors.New("invoice not found")
	ErrInvoiceImmutable               = errors.New("invoice cannot be modified after being issued")
	ErrInvoiceAlreadyPaid             = errors.New("invoice is already fully paid")
	ErrPaymentExceedsOutstanding      = errors.New("payment amount exceeds outstanding balance")
	ErrDueDateBeforeInvoiceDate       = errors.New("due date cannot be before invoice date")
	ErrInvalidInvoiceStatus           = errors.New("invalid invoice status")
	ErrInvalidStatusTransition        = errors.New("invalid status transition")
	ErrCreditNoteMustReferenceInvoice = errors.New("credit note must reference an original invoice")
	ErrInvalidOriginalInvoice         = errors.New("original invoice not found or not in issued status")
	ErrInvoiceNumberNotUnique         = errors.New("invoice number must be unique per organization and financial year")
	ErrInvalidGSTINFormat             = errors.New("invalid GSTIN format (must be 15 characters)")
	ErrInvalidPANFormat               = errors.New("invalid PAN format (must be 10 characters)")
	ErrNegativeAmount                 = errors.New("amount cannot be negative")
	ErrInvalidCurrencyPrecision       = errors.New("currency amount must have exactly 2 decimal places")
)

// InvoiceStatusTransition defines valid status transitions
var InvoiceStatusTransition = map[string][]string{
	"draft":     {"issued", "cancelled"},
	"issued":    {"paid", "partial", "overdue", "cancelled"},
	"partial":   {"paid", "overdue", "cancelled"},
	"overdue":   {"paid", "partial", "cancelled"},
	"paid":      {}, // Terminal state - no transitions allowed
	"cancelled": {}, // Terminal state - no transitions allowed
}

// IsValidStatusTransition checks if a status transition is valid
func IsValidStatusTransition(from, to string) bool {
	allowedTransitions, exists := InvoiceStatusTransition[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}
	return false
}

// IsInvoiceEditable checks if an invoice can be edited
func IsInvoiceEditable(status string) bool {
	return status == "draft"
}

// IsInvoiceTerminal checks if invoice is in terminal state
func IsInvoiceTerminal(status string) bool {
	return status == "paid" || status == "cancelled"
}
