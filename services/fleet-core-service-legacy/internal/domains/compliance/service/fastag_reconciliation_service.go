package service

import (
	"context"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FASTagReconciliationService interface {
	ReconcileTransaction(ctx context.Context, txn *models.TollTransaction) error
}

type fastagReconciliationServiceImpl struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewFASTagReconciliationService(repo repository.ComplianceRepository, logger logger.Logger) FASTagReconciliationService {
	return &fastagReconciliationServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *fastagReconciliationServiceImpl) ReconcileTransaction(ctx context.Context, txn *models.TollTransaction) error {
	// 1. Match toll transaction against trip sheets or expected costs
	// 2. Validate toll amount (NPCI/GST validation placeholder)
	// 3. Persist reconciliation record

	recon := &models.FASTagReconciliation{
		ID:             uuid.New(),
		TransactionID:  txn.ID,
		VehicleID:      txn.VehicleID,
		ExpectedAmount: txn.Amount, // Assume expected = actual for simple reconciler, or fetch from Master Toll table
		ActualAmount:   txn.Amount,
		Status:         models.ReconMatched,
	}

	if err := s.repo.RecordFASTagReconciliation(ctx, recon); err != nil {
		s.logger.Error("failed to record FASTag reconciliation", zap.Error(err), zap.String("id", txn.ID.String()))
		return err
	}

	s.logger.Info("FASTag reconciliation completed", zap.String("recon_status", string(recon.Status)), zap.String("id", recon.ID.String()))
	return nil
}
