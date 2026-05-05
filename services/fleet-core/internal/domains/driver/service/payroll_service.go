package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PayrollRepo interface {
	CreateSalaryConfig(ctx context.Context, config *models.SalaryConfig) error
	CreateTripEarning(ctx context.Context, earning *models.TripEarning) error
	GetSalaryConfig(ctx context.Context, driverID uuid.UUID) (*models.SalaryConfig, error)
	GetTripEarnings(ctx context.Context, driverID uuid.UUID, start, end time.Time) ([]*models.TripEarning, error)
	CreatePayout(ctx context.Context, payout *models.Payout) error
}

type PayrollService struct {
	repo   PayrollRepo
	logger *zap.Logger
}

func NewPayrollService(repo PayrollRepo, logger *zap.Logger) *PayrollService {
	return &PayrollService{
		repo:   repo,
		logger: logger,
	}
}

func (s *PayrollService) CalculateTripEarnings(ctx context.Context, driverID uuid.UUID, tripID uuid.UUID, distanceKm float64) (*models.TripEarning, error) {
	config, err := s.repo.GetSalaryConfig(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get salary config: %w", err)
	}

	earning := &models.TripEarning{
		DriverID:   driverID,
		TripID:     tripID,
		DistanceKm: distanceKm,
		Earnings:   distanceKm * config.PerKmRate,
	}

	if err := s.repo.CreateTripEarning(ctx, earning); err != nil {
		return nil, err
	}

	return earning, nil
}

func (s *PayrollService) GenerateMonthlyPayout(ctx context.Context, driverID uuid.UUID, year int, month time.Month) (*models.Payout, error) {
	config, err := s.repo.GetSalaryConfig(ctx, driverID)
	if err != nil {
		return nil, err
	}

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Second)

	earnings, err := s.repo.GetTripEarnings(ctx, driverID, start, end)
	if err != nil {
		return nil, err
	}

	totalEarnings := config.BaseSalary
	for _, e := range earnings {
		totalEarnings += e.Earnings + e.Bonus - e.Penalty
	}

	payout := &models.Payout{
		DriverID:    driverID,
		PeriodStart: start,
		PeriodEnd:   end,
		TotalAmount: totalEarnings,
		Status:      "PENDING",
	}

	if err := s.repo.CreatePayout(ctx, payout); err != nil {
		return nil, err
	}

	return payout, nil
}

func (s *PayrollService) GetTripEarnings(ctx context.Context, driverID uuid.UUID, start, end time.Time) ([]*models.TripEarning, error) {
	return s.repo.GetTripEarnings(ctx, driverID, start, end)
}
