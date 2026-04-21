package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
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
func (s *SubscriptionService) Subscribe(ctx context.Context, tenantID uuid.UUID, planID uuid.UUID) (*models.Subscription, error) {
	// Check if already subscribed
	existing, err := s.repo.GetSubscription(ctx, tenantID)
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
		TenantID:              tenantID,
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

	s.logger.Info("organization subscribed", zap.String("tenant_id", tenantID.String()), zap.String("plan_id", planID.String()))
	return sub, nil
}

// ChangePlan handles upgrading or downgrading a subscription with proration
func (s *SubscriptionService) ChangePlan(ctx context.Context, tenantID uuid.UUID, newPlanID uuid.UUID) (*models.Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("subscription not found for organization")
	}

	if sub.PlanID == newPlanID {
		return sub, nil
	}

	// Fetch old plan details for proration calculation
	oldPlan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}
	if oldPlan == nil {
		return nil, fmt.Errorf("old plan not found")
	}

	// Calculate proration credit for unused time on old plan
	now := time.Now().UTC()
	totalPeriod := sub.CurrentPeriodEnd.Sub(sub.CurrentPeriodStart)
	remaining := sub.CurrentPeriodEnd.Sub(now)

	var prorationCredit float64 = 0
	if remaining > 0 && totalPeriod > 0 {
		fraction := remaining.Hours() / totalPeriod.Hours()

		// Sum monthly price of all FLAT components in old plan
		var monthlyTotal float64
		for _, comp := range oldPlan.Components {
			if comp.Type == models.ComponentFlat {
				monthlyTotal += comp.UnitPrice
			}
		}

		prorationCredit = monthlyTotal * fraction
		s.logger.Info("Proration credit calculated",
			zap.String("tenant_id", tenantID.String()),
			zap.String("old_plan_id", sub.PlanID.String()),
			zap.Float64("credit_amount", prorationCredit),
			zap.Float64("fraction", fraction))
	}

	// Update subscription to new plan
	sub.PlanID = newPlanID
	sub.UpdatedAt = now

	// Store proration credit in metadata for application to next invoice
	if sub.Metadata == nil {
		sub.Metadata = make(map[string]interface{})
	}
	sub.Metadata["proration_credit"] = prorationCredit
	sub.Metadata["proration_calculated_at"] = now.Format(time.RFC3339)

	err = s.repo.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}

	s.logger.Info("subscription plan changed with proration",
		zap.String("tenant_id", tenantID.String()),
		zap.String("old_plan_id", sub.PlanID.String()),
		zap.String("new_plan_id", newPlanID.String()),
		zap.Float64("proration_credit", prorationCredit))

	return sub, nil
}

// CancelSubscription marks a subscription for cancellation at the end of the period
func (s *SubscriptionService) CancelSubscription(ctx context.Context, tenantID uuid.UUID, immediate bool) error {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
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
func (s *SubscriptionService) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetSubscription(ctx, tenantID)
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	err := s.repo.UpdateSubscription(ctx, sub)
	if err != nil {
		s.logger.Error("Failed to update subscription", zap.Error(err), zap.String("sub_id", sub.ID.String()))
		return err
	}

	// Audit logging for update
	_ = s.auditRepo.CreateAuditLogEntry(ctx, sub.TenantID, models.AuditActionUpdated, nil, sub, uuid.Nil, "subscription", sub.ID.String(), "", "")

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
