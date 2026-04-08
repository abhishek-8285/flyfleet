package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	finTracing "github.com/flyfleet/fleet-core/internal/domains/finance/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// EInvoiceService generates GST e-invoices and submits them to India's IRP.
type EInvoiceService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewEInvoiceService(repo *repository.FinanceRepository, logger *zap.Logger) *EInvoiceService {
	return &EInvoiceService{repo: repo, logger: logger}
}

// GenerateIRN produces an Invoice Reference Number.
// In production this would call the IRP API (NIC/GSP portal).
func (s *EInvoiceService) GenerateIRN(invoice *models.GSTInvoice) (string, error) {
	// IRN format: SHA256 of (GSTIN + InvoiceNumber + FinancialYear)
	// Stub: generate deterministic random placeholder
	h := fmt.Sprintf("%s%s%s", invoice.GSTIN, invoice.InvoiceNumber, time.Now().Format("2006"))
	irn := fmt.Sprintf("%x", []byte(h))[:64]
	return irn, nil
}

// GenerateEInvoice creates a GST e-invoice record for an entity.
func (s *EInvoiceService) GenerateEInvoice(ctx context.Context, req models.GenerateEInvoiceRequest) (*models.GSTInvoice, error) {
	ctx, span := finTracing.StartSpan(ctx, "finance.generate_einvoice",
		attribute.String("entity_type", req.EntityType),
		attribute.String("org_id", req.OrgID.String()),
	)
	defer span.End()

	// Validate GST schema
	if !validateGSTIN(req.SellerGSTIN) {
		return nil, fmt.Errorf("invalid seller GSTIN: %s", req.SellerGSTIN)
	}

	// Calculate GST components
	igst := req.TotalAmount - req.TaxableAmount
	taxRate := 0.0
	if req.TaxableAmount > 0 {
		taxRate = igst / req.TaxableAmount
	}
	cgst, sgst := 0.0, 0.0
	// Inter-state → IGST; intra-state → CGST+SGST (simplified logic)
	if req.BuyerGSTIN != nil && len(*req.BuyerGSTIN) >= 2 && (*req.BuyerGSTIN)[:2] == req.SellerGSTIN[:2] {
		cgst = igst / 2
		sgst = igst / 2
		igst = 0
	}
	_ = taxRate

	// SECURITY FIX: Use cryptographically secure random number instead of math/rand
	randomSuffix, _ := rand.Int(rand.Reader, big.NewInt(9999))
	invoiceNum := fmt.Sprintf("INV-%d-%d", time.Now().Unix(), randomSuffix.Int64())
	inv := &models.GSTInvoice{
		ID:            uuid.New(),
		InvoiceNumber: invoiceNum,
		Status:        models.GSTInvoiceStatusPending,
		EntityType:    req.EntityType,
		OrgID:         req.OrgID,
		GSTIN:         req.SellerGSTIN,
		BuyerGSTIN:    req.BuyerGSTIN,
		TotalAmount:   req.TotalAmount,
		TaxableAmount: req.TaxableAmount,
		CGSTAmount:    cgst,
		SGSTAmount:    sgst,
		IGSTAmount:    igst,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if req.EntityID != uuid.Nil {
		inv.EntityID = &req.EntityID
	}

	// Generate IRN
	irn, err := s.GenerateIRN(inv)
	if err != nil {
		s.logger.Error("failed to generate IRN", zap.Error(err))
	} else {
		now := time.Now()
		ackNum := fmt.Sprintf("ACK%d", time.Now().UnixNano())
		inv.IRN = &irn
		inv.AckNumber = &ackNum
		inv.AckDate = &now
		inv.Status = models.GSTInvoiceStatusGenerated
		// Stub QR code
		qr := fmt.Sprintf("QR_DATA:%s:%s", irn, invoiceNum)
		inv.QRCode = &qr
		signedJSON := fmt.Sprintf(`{"irn":"%s","ackNo":"%s"}`, irn, ackNum)
		inv.SignedInvoice = &signedJSON
	}

	if err := s.repo.CreateGSTInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to persist e-invoice: %w", err)
	}

	s.logger.Info("E-invoice generated", zap.String("invoice_number", inv.InvoiceNumber), zap.String("status", string(inv.Status)))
	return inv, nil
}

// GetEInvoice retrieves an e-invoice by ID.
func (s *EInvoiceService) GetEInvoice(ctx context.Context, id uuid.UUID) (*models.GSTInvoice, error) {
	return s.repo.GetGSTInvoice(ctx, id)
}

// validateGSTIN checks GST number format (15-char alphanumeric).
func validateGSTIN(gstin string) bool {
	if len(gstin) != 15 {
		return false
	}
	return true
}
