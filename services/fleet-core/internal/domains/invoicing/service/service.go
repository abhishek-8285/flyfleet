package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/flyfleet/fleet-core/internal/domains/invoicing/repository"
	"github.com/google/uuid"
)

type InvoicingService struct {
	repo repository.InvoiceRepository
}

func NewInvoicingService(repo repository.InvoiceRepository) *InvoicingService {
	return &InvoicingService{repo: repo}
}

func (s *InvoicingService) CreateInvoice(ctx context.Context, orgID uuid.UUID, inv *invoicing.Invoice) (*invoicing.Invoice, error) {
	// 1. Validation
	if inv.CustomerName == "" {
		return nil, fmt.Errorf("customer name is required")
	}
	if inv.CustomerPhone == "" {
		return nil, fmt.Errorf("customer phone number is mandatory")
	}

	// 2. Enrich invoice
	inv.OrgID = orgID
	inv.InvoiceNumber = s.generateInvoiceNumber()
	inv.Status = invoicing.StatusPending
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = time.Now()

	// 2. Calculate totals
	var subtotal float64
	for i := range inv.Items {
		inv.Items[i].OrgID = orgID
		inv.Items[i].Amount = inv.Items[i].Quantity * inv.Items[i].UnitPrice
		subtotal += inv.Items[i].Amount
	}
	inv.Subtotal = subtotal
	inv.TaxAmount = subtotal * 0.18 // 18% standard GST fallback
	inv.TotalAmount = inv.Subtotal + inv.TaxAmount
	inv.AmountPaid = 0
	inv.AmountRemaining = inv.TotalAmount

	// 3. Persist
	return s.repo.Create(ctx, inv)
}

func (s *InvoicingService) GetInvoice(ctx context.Context, orgID, id uuid.UUID) (*invoicing.Invoice, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

func (s *InvoicingService) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error) {
	return s.repo.List(ctx, orgID, limit, offset)
}

func (s *InvoicingService) VoidInvoice(ctx context.Context, orgID, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, orgID, id, invoicing.StatusVoid, uuid.Nil)
}

func (s *InvoicingService) generateInvoiceNumber() string {
	return fmt.Sprintf("INV-%d", time.Now().UnixNano()/1e6)
}
