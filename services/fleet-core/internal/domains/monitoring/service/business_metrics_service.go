package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/repository"
	"github.com/flyfleet/fleet-core-service/internal/platform/observability"
	"go.uber.org/zap"
)

type BusinessMetricsService struct {
	repo   repository.MonitoringRepository
	logger *zap.Logger
}

func NewBusinessMetricsService(repo repository.MonitoringRepository, logger *zap.Logger) *BusinessMetricsService {
	return &BusinessMetricsService{repo: repo, logger: logger}
}

func (s *BusinessMetricsService) GetKPIs(ctx context.Context) (*models.BusinessKPIs, error) {
	ctx, span := observability.StartSpan(ctx, "BusinessMetrics.GetKPIs")
	defer span.End()

	// Simulate data retrieval from across domain aggregates
	kpis := &models.BusinessKPIs{
		ActiveVehicles:      1250,
		ActiveDrivers:       840,
		TelemetryRatePerSec: 450.5,
		WorkOrdersPerHour:   12.4,
		ActiveTenants:       85,
		Timestamp:           time.Now(),
	}

	// Record to Prometheus metrics for dashboarding
	// (Assumes internal/platform/metrics has these gauges)
	// metrics.ActiveVehicles.Set(float64(kpis.ActiveVehicles))
	// metrics.ActiveTenants.Set(float64(kpis.ActiveTenants))

	return kpis, nil
}
