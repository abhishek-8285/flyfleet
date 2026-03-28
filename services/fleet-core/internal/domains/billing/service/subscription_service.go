package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionService struct {
	repo      *repository.SubscriptionRepository
	auditRepo *repository.AuditLogRepository
	logger    *zap.Logger
}

func NewSubscriptionService(repo *repository.SubscriptionRepository, auditRepo *repository.AuditLogRepository, logger *zap.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:      repo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// Subscribe starts a new subscription for an organization
func (s *SubscriptionService) Subscribe(ctx context.Context, orgID uuid.UUID, planID uuid.UUID) (*models.Subscription, error) {
	// Check if already subscribed
	existing, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.Status != models.SubscriptionCancelled {
		return nil, fmt.Errorf("organization already has an active subscription")
	}

	now := time.Now().UTC()
	trialDays := 14
	trialEnd := now.AddDate(0, 0, trialDays)

	sub := &models.Subscription{
		ID:                 uuid.New(),
		OrgID:              orgID,
		PlanID:             planID,
		Status:             models.SubscriptionTrial,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0), // Default to monthly
		TrialEnd:           &trialEnd,
		CancelAtPeriodEnd:  false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err = s.repo.CreateSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}

	s.logger.Info("organization subscribed", zap.String("org_id", orgID.String()), zap.String("plan_id", planID.String()))
	return sub, nil
}

// ChangePlan handles upgrading or downgrading a subscription
func (s *SubscriptionService) ChangePlan(ctx context.Context, orgID uuid.UUID, newPlanID uuid.UUID) (*models.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("subscription not found for organization")
	}

	if sub.PlanID == newPlanID {
		return sub, nil
	}

	// Simple plan switch logic: update plan and reset period
	// PRORATION TODO: Calculate credit for remaining time on old plan
	sub.PlanID = newPlanID
	sub.UpdatedAt = time.Now().UTC()

	err = s.repo.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}

	s.logger.Info("subscription plan changed", zap.String("org_id", orgID.String()), zap.String("new_plan_id", newPlanID.String()))
	return sub, nil
}

// CancelSubscription marks a subscription for cancellation at the end of the period
func (s *SubscriptionService) CancelSubscription(ctx context.Context, orgID uuid.UUID, immediate bool) error {
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("subscription not found")
	}

	if immediate {
		sub.Status = models.SubscriptionCancelled
	} else {
		sub.CancelAtPeriodEnd = true
	}

	sub.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateSubscription(ctx, sub)
}

// GetSubscription retrieves the current subscription for an organization
func (s *SubscriptionService) GetSubscription(ctx context.Context, orgID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetSubscription(ctx, orgID)
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	err := s.repo.UpdateSubscription(ctx, sub)
	if err != nil {
		s.logger.Error("Failed to update subscription", zap.Error(err), zap.String("sub_id", sub.ID.String()))
		return err
	}

	// Audit logging for update
	_ = s.auditRepo.CreateAuditLogEntry(ctx, sub.OrgID, models.AuditActionUpdated, nil, sub, uuid.Nil, "subscription", sub.ID.String(), "", "")

	return nil
}

// ListPlans returns all available billing plans
func (s *SubscriptionService) ListPlans(ctx context.Context) ([]models.Plan, error) {
	return s.repo.ListPlans(ctx)
}

// IdentifyOrganizationsForInvoicing identifies orgs that reached their billing period end
func (s *SubscriptionService) IdentifyOrganizationsForInvoicing(ctx context.Context) ([]uuid.UUID, error) {
	return s.repo.IdentifyOrganizationsForInvoicing(ctx)
}
