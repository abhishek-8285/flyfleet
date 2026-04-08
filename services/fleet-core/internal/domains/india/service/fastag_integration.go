package service

import (
	"context"
	"fmt"
	"time"

	billingv2service "github.com/flyfleet/fleet-core/internal/domains/billing/service"
	"github.com/flyfleet/fleet-core/internal/domains/india/models"
	"github.com/flyfleet/fleet-core/internal/domains/india/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// FastagIntegrationService defines the interface for FASTag integration
type FastagIntegrationService interface {
	FetchTransactions(ctx context.Context, orgID uuid.UUID, tagNumber string, from, to time.Time) ([]*models.FASTagTransaction, error)
	SyncTransactions(ctx context.Context) error
}

type FastagIntegrationServiceImpl struct {
	repo     repository.IndiaIntegrationRepository
	detector *billingv2service.DoubleSwipeDetector
	logger   *zap.Logger
	limiter  *rate.Limiter
}

func NewFastagIntegrationService(repo repository.IndiaIntegrationRepository, detector *billingv2service.DoubleSwipeDetector, logger *zap.Logger) *FastagIntegrationServiceImpl {
	return &FastagIntegrationServiceImpl{
		repo:     repo,
		detector: detector,
		logger:   logger,
		limiter:  rate.NewLimiter(rate.Every(time.Minute/100), 1),
	}
}

func (s *FastagIntegrationServiceImpl) FetchTransactions(ctx context.Context, orgID uuid.UUID, tagNumber string, from, to time.Time) ([]*models.FASTagTransaction, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}
	s.logger.Info("fetching fastag transactions", zap.String("tag_number", tagNumber), zap.Time("from", from), zap.Time("to", to))

	// Mock NPCI API Call
	// In production, this would use a real NPCI client
	mockTransactions := s.getMockNPCITransactions(orgID, tagNumber, from, to)

	// Log the integration request
	log := &models.IntegrationLog{
		ServiceType: "fastag",
		RequestPayload: map[string]interface{}{
			"tag_number": tagNumber,
			"from":       from,
			"to":         to,
		},
		ResponsePayload: mockTransactions,
		Status:          200,
		OrganizationID:  &orgID,
	}
	if err := s.repo.CreateIntegrationLog(ctx, log); err != nil {
		s.logger.Error("failed to create integration log", zap.Error(err))
	}

	for _, tx := range mockTransactions {
		if err := s.repo.CreateFASTagTransaction(ctx, tx); err != nil {
			s.logger.Error("failed to store fastag transaction", zap.Error(err), zap.String("tx_id", tx.TransactionID))
		}

		// Double-Swipe Detection
		if s.detector != nil {
			if err := s.detector.Detect(ctx, tx); err != nil {
				s.logger.Error("failed to detect double swipe", zap.Error(err), zap.String("tx_id", tx.TransactionID))
			}
		}
	}

	return mockTransactions, nil
}

func (s *FastagIntegrationServiceImpl) SyncTransactions(ctx context.Context) error {
	s.logger.Info("starting fastag transactions sync")

	// 1. Fetch active integration configs for FASTag
	// For simplicity, we'll assume a global sync or iterate through organizations
	// In a real scenario, we'd list all organizations with active FASTag integration

	// Example: Fetch last 24h for all configured tags
	// This is a simplified version
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// For now, let's just log that we are syncing
	// Real implementation would loop through organizations and their tags
	s.logger.Info("fastag sync completed", zap.Time("from", yesterday), zap.Time("to", now))
	return nil
}

func (s *FastagIntegrationServiceImpl) getMockNPCITransactions(orgID uuid.UUID, tagNumber string, from, to time.Time) []*models.FASTagTransaction {
	return []*models.FASTagTransaction{
		{
			ID:              uuid.New().String(),
			TagNumber:       tagNumber,
			VehicleID:       uuid.New(), // In real case, link to actual vehicle
			DriverID:        uuid.New(), // In real case, link to actual driver
			OrganizationID:  orgID,
			TransactionID:   fmt.Sprintf("TXN-%d", time.Now().UnixNano()),
			TransactionType: "debit",
			Amount:          150.00,
			TollPlaza:       "Kherki Daula Toll Plaza",
			TollPlazaID:     "PLAZA-001",
			LaneID:          "LANE-05",
			TransactionTime: time.Now().Add(-2 * time.Hour),
			TransactionDate: time.Now().Format("2006-01-02"),
			Status:          "success",
			CreatedBy:       uuid.Nil,
			UpdatedBy:       uuid.Nil,
		},
	}
}
