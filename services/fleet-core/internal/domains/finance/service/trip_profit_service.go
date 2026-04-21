package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	finTracing "github.com/flyfleet/fleet-core/internal/domains/finance/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// TripProfitService calculates trip-level profitability.
type TripProfitService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewTripProfitService(repo *repository.FinanceRepository, logger *zap.Logger) *TripProfitService {
	return &TripProfitService{repo: repo, logger: logger}
}

// CalculateTripProfit gathers all costs for a trip and persists profitability metrics.
func (s *TripProfitService) CalculateTripProfit(ctx context.Context, tripID uuid.UUID, tenantID uuid.UUID) (*models.TripProfitability, error) {
	ctx, span := finTracing.StartSpan(ctx, "finance.calculate_trip_profit",
		attribute.String("trip_id", tripID.String()),
		attribute.String("tenant_id", tenantID.String()),
	)
	defer span.End()

	// Fetch component costs from repository
	fuelCost, err := s.repo.GetTripFuelCost(ctx, tripID.String())
	if err != nil {
		s.logger.Warn("failed to get trip fuel cost", zap.Error(err))
	}

	tollCost, err := s.repo.GetTripTollCost(ctx, tripID.String())
	if err != nil {
		s.logger.Warn("failed to get trip toll cost", zap.Error(err))
	}

	distanceKM, err := s.repo.GetTripDistance(ctx, tripID.String())
	if err != nil {
		s.logger.Warn("failed to get trip distance", zap.Error(err))
	}

	// Revenue — would come from billing/booking module in production
	tripRevenue, err := s.repo.GetTripRevenue(ctx, tripID.String())
	if err != nil {
		s.logger.Warn("failed to get trip revenue", zap.Error(err))
	}

	totalCost := fuelCost + tollCost
	profit := tripRevenue - totalCost
	profitMargin := 0.0
	if tripRevenue > 0 {
		profitMargin = (profit / tripRevenue) * 100
	}
	costPerKM := 0.0
	if distanceKM > 0 {
		costPerKM = totalCost / distanceKM
	}

	tp := &models.TripProfitability{
		ID:           uuid.New(),
		TripID:       tripID,
		TenantID:        tenantID,
		TripRevenue:  tripRevenue,
		FuelCost:     fuelCost,
		TollCost:     tollCost,
		DriverCost:   0,
		TotalCost:    totalCost,
		Profit:       profit,
		ProfitMargin: profitMargin,
		DistanceKM:   distanceKM,
		CostPerKM:    costPerKM,
		CalculatedAt: time.Now(),
	}

	if err := s.repo.UpsertTripProfitability(ctx, tp); err != nil {
		return nil, fmt.Errorf("failed to save trip profitability: %w", err)
	}

	return tp, nil
}

// GetTripProfitability retrieves pre-computed profit metrics for a trip.
func (s *TripProfitService) GetTripProfitability(ctx context.Context, tripID uuid.UUID) (*models.TripProfitability, error) {
	return s.repo.GetTripProfitability(ctx, tripID)
}

// GetDashboardMetrics retrieves the materialized finance dashboard metrics for an org.
func (s *TripProfitService) GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*models.FinanceDashboardMetrics, error) {
	return s.repo.GetFinanceDashboardMetrics(ctx, tenantID)
}
