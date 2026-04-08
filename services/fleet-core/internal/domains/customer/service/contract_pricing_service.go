package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ContractPricingService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewContractPricingService(repo repository.CustomerV2Repository, logger *zap.Logger) *ContractPricingService {
	return &ContractPricingService{
		repo:   repo,
		logger: logger,
	}
}

// CalculateContractRate checks if a customer has an active custom pricing contract, and calculates shipment cost
func (s *ContractPricingService) CalculateContractRate(ctx context.Context, customerID uuid.UUID, origin, destination []byte, weight float64) (*models.PricingQuote, error) {
	s.logger.Info("Calculating contract rate", zap.String("customer_id", customerID.String()))

	// 1. Fetch active contract from repo
	// contract, err := s.repo.GetActiveContract(ctx, customerID)

	// Stub calculation mimicking a complex enterprise rule engine
	baseRate := 4000.0
	perKmRate := 18.5
	distanceMock := 200.0 // kilometers

	total := baseRate + (perKmRate * distanceMock)

	// If weight > 20000 kg, add heavy cargo surcharge
	if weight > 20000 {
		total += 1500.0
	}

	quote := &models.PricingQuote{
		ID:          uuid.New(),
		BookingID:   uuid.Nil, // Not yet attached to a booking
		TotalAmount: total,
		Currency:    "INR",
		CreatedAt:   time.Now(),
	}

	return quote, nil
}
