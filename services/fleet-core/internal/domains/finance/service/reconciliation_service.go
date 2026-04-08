package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"go.uber.org/zap"
)

type ReconciliationService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewReconciliationService(repo *repository.FinanceRepository, logger *zap.Logger) *ReconciliationService {
	return &ReconciliationService{repo: repo, logger: logger}
}

// ReconcileFuelCardTransactions matches external fuel card transactions with internal fuel logs.
func (s *ReconciliationService) ReconcileFuelCardTransactions(ctx context.Context, orgID string) error {
	s.logger.Info("Starting fuel card reconciliation", zap.String("org_id", orgID))

	// 1. Get unreconciled card transactions
	cardTxs, err := s.repo.GetUnreconciledFuelCardTransactions(ctx, orgID)
	if err != nil {
		return err
	}

	for _, tx := range cardTxs {
		// 2. Search for matching fuel transaction/log
		// Match criteria: Vehicle (if mapped), Amount (+/- 1%), Date (+/- 1 day)
		s.logger.Debug("Attempting to reconcile card transaction",
			zap.String("card_id", tx.CardID),
			zap.Float64("amount", tx.Amount),
		)
		// Stub: full fuzzy matching logic would go here
	}

	return nil
}
