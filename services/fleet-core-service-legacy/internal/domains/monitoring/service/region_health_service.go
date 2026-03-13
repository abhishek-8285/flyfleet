package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/repository"
	"github.com/flyfleet/fleet-core-service/internal/platform/observability"
)

type MultiRegionHealthService struct {
	repo   repository.MonitoringRepository
	logger *zap.Logger
}

func NewMultiRegionHealthService(repo repository.MonitoringRepository, logger *zap.Logger) *MultiRegionHealthService {
	return &MultiRegionHealthService{repo: repo, logger: logger}
}

func (s *MultiRegionHealthService) GetGlobalHealth(ctx context.Context) ([]models.RegionHealth, error) {
	ctx, span := observability.StartSpan(ctx, "MultiRegionHealth.GetGlobalHealth")
	defer span.End()

	regions := []models.RegionHealth{
		{
			RegionName:    "aws-us-east-1",
			Status:        "HEALTHY",
			LatencyMS:     45,
			FailoverReady: true,
			LastChecked:   time.Now(),
		},
		{
			RegionName:    "aws-eu-central-1",
			Status:        "HEALTHY",
			LatencyMS:     120,
			FailoverReady: true,
			LastChecked:   time.Now(),
		},
		{
			RegionName:    "aws-ap-south-1",
			Status:        "DEGRADED",
			LatencyMS:     450,
			FailoverReady: false,
			LastChecked:   time.Now(),
		},
	}

	return regions, nil
}
