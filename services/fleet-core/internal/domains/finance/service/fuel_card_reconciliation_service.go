package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FuelCardReconciliationService matches external card charges with internal fuel logs.
type FuelCardReconciliationService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewFuelCardReconciliationService(repo *repository.FinanceRepository, logger *zap.Logger) *FuelCardReconciliationService {
	return &FuelCardReconciliationService{repo: repo, logger: logger}
}

// RunReconciliation matches unreconciled card transactions with internal fuel logs.
func (s *FuelCardReconciliationService) RunReconciliation(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	s.logger.Info("Starting fuel card reconciliation", zap.String("tenant_id", tenantID.String()))

	// Get unreconciled card transactions (external)
	cardTxs, err := s.repo.GetUnreconciledFuelCardTransactions(ctx, tenantID.String())
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get card transactions: %w", err)
	}

	matched, unmatched := 0, 0
	for _, cardTx := range cardTxs {
		recon := &models.FuelCardReconciliation{
			ID:                uuid.New(),
			CardTransactionID: &cardTx.ID,
			TenantID:             tenantID,
			Status:            models.ReconciliationStatusUnmatched,
			CardAmount:        &cardTx.Amount,
			CreatedAt:         time.Now(),
		}

		// Attempt amount+time window matching
		internalLog, err := s.repo.GetFuelLogForMatching(ctx, tenantID.String(), cardTx.Amount, cardTx.TransactionTime)
		if err == nil && internalLog != nil {
			recon.InternalTransactionID = &internalLog.ID
			recon.Status = models.ReconciliationStatusMatched
			recon.InternalAmount = &internalLog.TotalCost
			recon.Variance = math.Abs(cardTx.Amount - internalLog.TotalCost)
			now := time.Now()
			recon.ReconciledAt = &now
			matched++
		} else {
			unmatched++
		}

		if err := s.repo.CreateFuelCardReconciliation(ctx, recon); err != nil {
			s.logger.Error("Failed to save reconciliation record", zap.Error(err))
		}
	}

	s.logger.Info("Reconciliation complete",
		zap.Int("matched", matched),
		zap.Int("unmatched", unmatched),
	)
	return matched, unmatched, nil
}

// ManualMatch allows a finance officer to manually reconcile a record.
func (s *FuelCardReconciliationService) ManualMatch(ctx context.Context, reconID, internalTxID, userID uuid.UUID) error {
	return s.repo.UpdateReconciliationStatus(ctx, reconID, internalTxID, userID, models.ReconciliationStatusMatched)
}
