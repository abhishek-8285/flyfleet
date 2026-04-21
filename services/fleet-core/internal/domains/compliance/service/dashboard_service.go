package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DashboardService interface {
	GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*models.ComplianceDashboardMetrics, error)
	RefreshMetrics(ctx context.Context, tenantID uuid.UUID) error
}

type dashboardServiceImpl struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewDashboardService(repo repository.ComplianceRepository, logger logger.Logger) DashboardService {
	return &dashboardServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *dashboardServiceImpl) GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*models.ComplianceDashboardMetrics, error) {
	metrics, err := s.repo.GetDashboardMetrics(ctx, tenantID)
	if err != nil {
		// If not found, trigger a refresh and return fresh/empty
		s.logger.Info("dashboard metrics not found, refreshing", zap.String("tenant_id", tenantID.String()))
		_ = s.RefreshMetrics(ctx, tenantID)
		return s.repo.GetDashboardMetrics(ctx, tenantID)
	}
	return metrics, nil
}

func (s *dashboardServiceImpl) RefreshMetrics(ctx context.Context, tenantID uuid.UUID) error {
	// Aggregate metrics from various tables
	// (In real implementation, these would be separate repository counts)

	metrics := &models.ComplianceDashboardMetrics{
		OrganizationID: tenantID,
		UpdatedAt:      time.Now(),
	}

	// Placeholder:
	// s.repo.GetActiveDriverCount(ctx, tenantID)
	// s.repo.GetExpiredDocCount(ctx, tenantID)
	// etc.

	metrics.DriversActive = 15
	metrics.DriversOverLimit = 2
	metrics.VehiclesWithExpiredDocs = 5
	metrics.PanicAlertsToday = 1
	metrics.OfflineDevices = 3
	metrics.WeeklyViolations = 8

	return s.repo.UpdateDashboardMetrics(ctx, metrics)
}
