package service

import (
	"context"

	"github.com/flyfleet/billing-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	ListInvoices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error)
	CreateInvoice(ctx context.Context, inv *models.Invoice) error
	GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	UpdateInvoice(ctx context.Context, inv *models.Invoice) error
	DeleteInvoice(ctx context.Context, id uuid.UUID) error
	GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error)
	UpdateSettings(ctx context.Context, s *models.BillingSettings) error
	GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error)
	RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error
}

type BillingService struct {
	repo   Repository
	logger *zap.Logger
}

func NewBillingService(repo Repository, logger *zap.Logger) *BillingService {
	return &BillingService{
		repo:   repo,
		logger: logger,
	}
}

func (s *BillingService) ListInvoices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error) {
	return s.repo.ListInvoices(ctx, orgID, page, size)
}

func (s *BillingService) CreateInvoice(ctx context.Context, inv *models.Invoice) error {
	// Calculate totals if not provided or to ensure correctness
	total := 0.0
	for i := range inv.Items {
		item := &inv.Items[i]
		itemTotal := float64(item.Quantity) * item.UnitPrice
		item.Total = itemTotal + (itemTotal * (item.TaxPercent / 100))
		total += item.Total
	}
	inv.TotalAmount = total
	inv.BalanceDue = total - inv.AmountPaid
	
	if inv.Status == "" {
		inv.Status = models.StatusPending
	}
	
	return s.repo.CreateInvoice(ctx, inv)
}

func (s *BillingService) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	return s.repo.GetInvoice(ctx, id)
}

func (s *BillingService) UpdateInvoice(ctx context.Context, inv *models.Invoice) error {
	// Recalculate totals
	total := 0.0
	for i := range inv.Items {
		item := &inv.Items[i]
		itemTotal := float64(item.Quantity) * item.UnitPrice
		item.Total = itemTotal + (itemTotal * (item.TaxPercent / 100))
		total += item.Total
	}
	inv.TotalAmount = total
	inv.BalanceDue = total - inv.AmountPaid

	return s.repo.UpdateInvoice(ctx, inv)
}

func (s *BillingService) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteInvoice(ctx, id)
}

func (s *BillingService) RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	return s.repo.RecordPayment(ctx, id, amount)
}

func (s *BillingService) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error) {
	settings, err := s.repo.GetSettings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		// Return default settings
		return &models.BillingSettings{
			OrgID:                orgID,
			DefaultTaxRate:       18.0,
			BaseCurrency:         "USD",
			GracePeriodDays:      14,
			AutoInvoicingEnabled: false,
		}, nil
	}
	return settings, nil
}

func (s *BillingService) UpdateSettings(ctx context.Context, settings *models.BillingSettings) error {
	return s.repo.UpdateSettings(ctx, settings)
}

func (s *BillingService) GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error) {
	return s.repo.GetMetrics(ctx, orgID)
}
