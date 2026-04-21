package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VendorPaymentService manages the vendor payment ledger.
type VendorPaymentService struct {
	repo       *repository.FinanceRepository
	tdsService *TDSTaxService
	logger     *zap.Logger
}

func NewVendorPaymentService(repo *repository.FinanceRepository, tds *TDSTaxService, logger *zap.Logger) *VendorPaymentService {
	return &VendorPaymentService{repo: repo, tdsService: tds, logger: logger}
}

// CreatePayment records a vendor payment and applies TDS.
func (s *VendorPaymentService) CreatePayment(ctx context.Context, tenantID uuid.UUID, createdBy uuid.UUID, req models.CreateVendorPaymentRequest) (*models.VendorPayment, error) {
	// Deduct TDS (194C rate for contractors by default)
	tdsReq := models.DeductTDSRequest{
		EntityType:  "vendor_payment",
		EntityID:    uuid.New(), // placeholder, will be updated after creation
		VendorID:    &req.VendorID,
		TenantID:       tenantID,
		GrossAmount: req.Amount,
		TDSSection:  "194C",
	}
	tdsRecord, err := s.tdsService.DeductTDS(ctx, tdsReq)
	if err != nil {
		s.logger.Warn("TDS deduction failed, proceeding without", zap.Error(err))
	}

	payment := &models.VendorPayment{
		ID:            uuid.New(),
		VendorID:      req.VendorID,
		TenantID:         tenantID,
		InvoiceNumber: req.InvoiceNumber,
		Amount:        req.Amount,
		Currency:      "INR",
		PaymentStatus: models.VendorPaymentStatusPending,
		PaymentMode:   &req.PaymentMode,
		BankAccount:   req.BankAccount,
		IFSCCode:      req.IFSCCode,
		Notes:         req.Notes,
		CreatedBy:     &createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if tdsRecord != nil {
		payment.TDSDeducted = tdsRecord.TDSAmount
		payment.NetPayable = tdsRecord.NetAmount
		payment.TDSRecordID = &tdsRecord.ID
	} else {
		payment.NetPayable = req.Amount
	}

	if err := s.repo.CreateVendorPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create vendor payment: %w", err)
	}

	return payment, nil
}

// ListPayments returns payments for an org with optional vendor filter.
func (s *VendorPaymentService) ListPayments(ctx context.Context, tenantID uuid.UUID, vendorID *uuid.UUID) ([]models.VendorPayment, error) {
	return s.repo.ListVendorPayments(ctx, tenantID, vendorID)
}
