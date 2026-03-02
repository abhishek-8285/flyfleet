package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/domain/invoice"
	ms_events "github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/events" // Note: reuse same pattern for events
	"github.com/google/uuid"
)

type InvoiceService struct {
	repo      invoice.Repository
	publisher ms_events.Publisher // Interface to be implemented in Phase 4
}

func NewInvoiceService(repo invoice.Repository, publisher ms_events.Publisher) *InvoiceService {
	return &InvoiceService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, orgID uuid.UUID, inv *invoice.Invoice) (*invoice.Invoice, error) {
	inv.OrgID = orgID
	inv.Status = invoice.StatusDraft
	
	userID := s.getUserID(ctx)
	inv.CreatedBy = userID
	inv.UpdatedBy = userID

	// Basic calculation
	var subtotal float64
	for i := range inv.Items {
		inv.Items[i].OrgID = orgID
		inv.Items[i].CreatedBy = userID
		inv.Items[i].UpdatedBy = userID
		inv.Items[i].Amount = float64(inv.Items[i].Quantity) * inv.Items[i].UnitPrice
		subtotal += inv.Items[i].Amount
	}
	inv.Subtotal = subtotal
	inv.TotalAmount = subtotal + inv.TaxAmount
	inv.AmountRemaining = inv.TotalAmount

	createdInv, err := s.repo.Create(ctx, inv)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice service: %w", err)
	}

	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, "invoice.created", createdInv)
	}

	return createdInv, nil
}

func (s *InvoiceService) GetInvoice(ctx context.Context, orgID, id uuid.UUID) (*invoice.Invoice, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

func (s *InvoiceService) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]invoice.Invoice, int64, error) {
	return s.repo.List(ctx, orgID, limit, offset)
}

func (s *InvoiceService) VoidInvoice(ctx context.Context, orgID, id uuid.UUID) error {
	err := s.repo.UpdateStatus(ctx, orgID, id, invoice.StatusVoid, s.getUserID(ctx))
	if err != nil {
		return err
	}

	if s.publisher != nil {
		_ = s.publisher.Publish(ctx, "invoice.updated", map[string]interface{}{
			"id":     id,
			"status": invoice.StatusVoid,
		})
	}
	return nil
}

func (s *InvoiceService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}

