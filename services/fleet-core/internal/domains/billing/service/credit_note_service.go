package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreditNoteService struct {
	repo       *repository.CreditNoteRepository
	ledgerServ *LedgerService
	logger     *zap.Logger
}

func NewCreditNoteService(repo *repository.CreditNoteRepository, ledgerServ *LedgerService, logger *zap.Logger) *CreditNoteService {
	return &CreditNoteService{
		repo:       repo,
		ledgerServ: ledgerServ,
		logger:     logger,
	}
}

func (s *CreditNoteService) IssueCreditNote(ctx context.Context, cn *models.CreditNote, items []models.CreditNoteItem) error {
	if cn.ID == uuid.Nil {
		cn.ID = uuid.New()
	}
	if cn.CreditNoteNo == "" {
		cn.CreditNoteNo = fmt.Sprintf("CN-%d", time.Now().UnixNano())
	}
	if cn.Status == "" {
		cn.Status = string(models.CreditNoteIssued)
	}

	// The existing repo expects items as a separate argument
	err := s.repo.Create(ctx, cn, items)
	if err != nil {
		s.logger.Error("Failed to issue credit note", zap.Error(err), zap.String("tenant_id", cn.OrgID.String()))
		return err
	}

	// Double-entry accounting: Record credit mapped to the invoice
	if s.ledgerServ != nil && cn.OriginalInvoiceID != uuid.Nil {
		// Calculate total amount from items
		totalAmount := 0.0
		for _, item := range items {
			totalAmount += item.TotalAmount
		}
		_ = s.ledgerServ.RecordCreditApplied(ctx, cn.OrgID, cn.OriginalInvoiceID, cn.ID, totalAmount, "USD")
	}

	return nil
}

func (s *CreditNoteService) ListCreditNotes(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.CreditNote, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	return s.repo.ListByOrg(ctx, orgID, page, size)
}
