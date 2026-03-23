package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TollService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewTollService(repo *repository.FinanceRepository, logger *zap.Logger) *TollService {
	return &TollService{repo: repo, logger: logger}
}

func (s *TollService) IngestToll(ctx context.Context, req models.IngestTollRequest) error {
	tx := &models.TollTransaction{
		VehicleID:             req.VehicleID,
		FASTagID:              req.FASTagID,
		TollPlaza:             req.TollPlaza,
		Amount:                req.Amount,
		Currency:              "INR", // Default for FASTag
		TransactionTime:       req.TransactionTime,
		ExternalTransactionID: req.ExternalTransactionID,
	}

	if err := s.repo.CreateTollTransaction(ctx, tx); err != nil {
		s.logger.Error("failed to create toll transaction", zap.Error(err))
		return fmt.Errorf("failed to save toll transaction: %w", err)
	}

	return nil
}

func (s *TollService) GetVehicleTolls(ctx context.Context, vehicleID uuid.UUID) ([]models.TollTransaction, error) {
	// For now returns empty, can implement retrieval in repository later if needed
	return []models.TollTransaction{}, nil
}
