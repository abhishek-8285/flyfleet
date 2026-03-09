package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/repository"
	"go.uber.org/zap"
)

type AnalyticsService struct {
	repo   *repository.AnalyticsRepository
	logger *zap.Logger
}

func NewAnalyticsService(repo *repository.AnalyticsRepository, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{
		repo:   repo,
		logger: logger,
	}
}

// GetSystemMetrics returns platform-wide financial statistics
func (s *AnalyticsService) GetSystemMetrics(ctx context.Context) (*models.FinancialMetrics, error) {
	metrics, err := s.repo.GetFinancialMetrics(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch financial metrics", zap.Error(err))
		return nil, err
	}

	churn, err := s.repo.GetChurnMetrics(ctx)
	if err != nil {
		s.logger.Warn("Failed to fetch churn metrics", zap.Error(err))
	} else {
		metrics.ChurnRate = churn
	}

	ltv, err := s.repo.GetLTV(ctx)
	if err != nil {
		s.logger.Warn("Failed to fetch LTV metrics", zap.Error(err))
	} else {
		metrics.LTV = ltv
	}

	return metrics, nil
}
