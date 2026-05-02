package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/platform/billing/models"
	"github.com/flyfleet/fleet-core/internal/platform/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BillingService interface {
	CheckLimit(ctx context.Context, tenantID uuid.UUID, metric string) (bool, error)
	GetTenantPlan(ctx context.Context, tenantID uuid.UUID) (*models.Plan, error)
}

type billingService struct {
	repo   repository.BillingRepository
	logger *zap.Logger
}

func NewBillingService(repo repository.BillingRepository, logger *zap.Logger) BillingService {
	return &billingService{
		repo:   repo,
		logger: logger,
	}
}

func (s *billingService) GetTenantPlan(ctx context.Context, tenantID uuid.UUID) (*models.Plan, error) {
	sub, err := s.repo.GetSubscriptionByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	plan, err := s.repo.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return plan, nil
}

func (s *billingService) CheckLimit(ctx context.Context, tenantID uuid.UUID, metric string) (bool, error) {
	plan, err := s.GetTenantPlan(ctx, tenantID)
	if err != nil {
		return false, err
	}

	usage, err := s.repo.GetUsage(ctx, tenantID, metric)
	if err != nil {
		return false, err
	}

	switch metric {
	case "vehicles":
		return usage < plan.MaxVehicles, nil
	case "drivers":
		return usage < plan.MaxDrivers, nil
	default:
		// Check JSONB features if needed
		return true, nil
	}
}
