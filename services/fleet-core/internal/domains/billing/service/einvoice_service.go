package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EInvoiceService handles e-invoicing (IRN) operations
type EInvoiceService struct {
	logger *zap.Logger
}

// NewEInvoiceService creates a new e-invoice service
func NewEInvoiceService(logger *zap.Logger) *EInvoiceService {
	return &EInvoiceService{
		logger: logger,
	}
}

// GenerateIRNHash generates IRN hash for an invoice
// IRN = SHA-256(InvoiceNo|FinYear|GSTIN|InvType|InvValue)
func (s *EInvoiceService) GenerateIRNHash(invoiceNo, finYear, gstin, invType string, invValue float64) string {
	input := fmt.Sprintf("%s|%s|%s|%s|%.2f", invoiceNo, finYear, gstin, invType, invValue)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// CheckEInvoiceMandatory checks if e-invoice is mandatory for an amount
func (s *EInvoiceService) CheckEInvoiceMandatory(amount float64, config *models.EInvoiceConfig) bool {
	if !config.IsEnabled || !config.IsActive {
		return false
	}
	return amount > config.MandatoryAbove
}

// GenerateIRN generates IRN for an invoice (calls NIC API in production)
func (s *EInvoiceService) GenerateIRN(ctx context.Context, invoiceID uuid.UUID, invoiceNo string, amount float64, gstin string) (*models.EInvoiceIRN, error) {
	// Generate financial year
	finYear := s.getFinancialYear(time.Now())

	// Generate IRN hash
	irn := s.GenerateIRNHash(invoiceNo, finYear, gstin, "INV", amount)

	// Create IRN record
	irnRecord := &models.EInvoiceIRN{
		ID:        uuid.New(),
		InvoiceID: invoiceID,
		IRN:       irn,
		IRNStatus: string(models.IRNStatusGenerated),
		CreatedAt: time.Now(),
	}

	s.logger.Info("IRN generated",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("irn", irn),
		zap.String("invoice_no", invoiceNo),
	)

	// In production, this would call NIC API:
	// 1. Prepare JSON as per NIC schema
	// 2. Sign with app key
	// 3. Call NIC API endpoint
	// 4. Store signed response and QR code

	return irnRecord, nil
}

// CancelIRN cancels an IRN (calls NIC API in production)
func (s *EInvoiceService) CancelIRN(ctx context.Context, irnID uuid.UUID, reason string) error {
	// In production, this would call NIC API to cancel IRN
	// Cancellation is only allowed within 24 hours of generation

	s.logger.Info("IRN cancelled",
		zap.String("irn_id", irnID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// GetFinancialYear returns financial year for a date
func (s *EInvoiceService) getFinancialYear(date time.Time) string {
	year := date.Year()
	month := date.Month()

	// Financial year starts April 1 in India
	if month >= 4 {
		return fmt.Sprintf("%d-%d", year, year+1)
	}
	return fmt.Sprintf("%d-%d", year-1, year)
}

// ValidateHSNCode validates HSN code format
func (s *EInvoiceService) ValidateHSNCode(hsnCode string) error {
	if hsnCode == "" {
		return errors.New("HSN code is required")
	}

	// HSN code should be 4, 6, or 8 digits
	length := len(hsnCode)
	if length != 4 && length != 6 && length != 8 {
		return errors.New("HSN code must be 4, 6, or 8 digits")
	}

	return nil
}

// PrepareEInvoiceJSON prepares JSON payload for NIC API
func (s *EInvoiceService) PrepareEInvoiceJSON(invoice interface{}) (map[string]interface{}, error) {
	// In production, this would prepare JSON as per NIC schema
	// Schema includes:
	// - Invoice details (number, date, value)
	// - Party details (GSTIN, name, address)
	// - Item details (HSN, quantity, rate, tax)
	// - Tax breakdown (CGST, SGST, IGST)

	return map[string]interface{}{
		"Version": "1.0",
		"TranDtls": map[string]interface{}{
			"TranId": uuid.New().String(),
		},
	}, nil
}

// GetEInvoiceComplianceStatus gets compliance status for an invoice
func (s *EInvoiceService) GetEInvoiceComplianceStatus(amount float64, hasIRN bool, config *models.EInvoiceConfig) string {
	if !config.IsEnabled || !config.IsActive {
		return "Not Required"
	}

	if amount <= config.MandatoryAbove {
		return "Not Required"
	}

	if hasIRN {
		return "Compliant"
	}

	return "Non-Compliant"
}
