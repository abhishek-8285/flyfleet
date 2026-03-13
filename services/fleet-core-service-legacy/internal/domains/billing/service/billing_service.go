package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Subscription represents a tenant's billing state
type Subscription struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	PlanID         uuid.UUID `json:"plan_id"`
	Status         string    `json:"status"`
	BillingCycle   string    `json:"billing_cycle"`
	EndDate        time.Time `json:"current_period_end"`
}

// Plan represents a billing tier
type Plan struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Price float64   `json:"price"`
}

// BillingService handles subscriptions and usage tracking
type BillingService interface {
	GetSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error)
	CreateSubscription(ctx context.Context, orgID, planID uuid.UUID, cycle string) (*Subscription, error)
	GetUsage(ctx context.Context, orgID uuid.UUID, usageType string) (int, error)
	RecordUsage(ctx context.Context, orgID uuid.UUID, usageType string, quantity int) error
	CheckFeatureAccess(ctx context.Context, orgID uuid.UUID, feature string) (bool, error)
}

type BillingServiceImpl struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewBillingService(db *pgxpool.Pool, logger *zap.Logger) BillingService {
	return &BillingServiceImpl{
		db:     db,
		logger: logger,
	}
}

func (s *BillingServiceImpl) GetSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	query := `SELECT id, organization_id, plan_id, status, billing_cycle, current_period_end 
			  FROM billing_subscriptions WHERE organization_id = $1`
	
	sub := &Subscription{}
	err := s.db.QueryRow(ctx, query, orgID).Scan(
		&sub.ID, &sub.OrganizationID, &sub.PlanID, &sub.Status, &sub.BillingCycle, &sub.EndDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

func (s *BillingServiceImpl) CreateSubscription(ctx context.Context, orgID, planID uuid.UUID, cycle string) (*Subscription, error) {
	sub := &Subscription{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PlanID:         planID,
		Status:         "active",
		BillingCycle:   cycle,
		EndDate:        time.Now().AddDate(0, 1, 0),
	}

	if cycle == "yearly" {
		sub.EndDate = time.Now().AddDate(1, 0, 0)
	}

	query := `INSERT INTO billing_subscriptions (id, organization_id, plan_id, status, billing_cycle, current_period_start, current_period_end)
			  VALUES ($1, $2, $3, $4, $5, NOW(), $6)`
	
	_, err := s.db.Exec(ctx, query, sub.ID, sub.OrganizationID, sub.PlanID, sub.Status, sub.BillingCycle, sub.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	
	s.logger.Info("subscription created", zap.String("org_id", orgID.String()), zap.String("plan_id", planID.String()))
	return sub, nil
}

func (s *BillingServiceImpl) GetUsage(ctx context.Context, orgID uuid.UUID, usageType string) (int, error) {
	query := `SELECT COALESCE(SUM(quantity), 0) FROM billing_usage_records 
			  WHERE organization_id = $1 AND usage_type = $2`
	var total int
	err := s.db.QueryRow(ctx, query, orgID, usageType).Scan(&total)
	return total, err
}

func (s *BillingServiceImpl) RecordUsage(ctx context.Context, orgID uuid.UUID, usageType string, quantity int) error {
	query := `INSERT INTO billing_usage_records (id, organization_id, usage_type, quantity, recorded_at)
			  VALUES ($1, $2, $3, $4, NOW())`
	_, err := s.db.Exec(ctx, query, uuid.New(), orgID, usageType, quantity)
	return err
}

func (s *BillingServiceImpl) CheckFeatureAccess(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
	query := `SELECT p.features FROM billing_plans p 
			  JOIN billing_subscriptions s ON p.id = s.plan_id 
			  WHERE s.organization_id = $1 AND s.status = 'active'`
	
	var features map[string]interface{}
	err := s.db.QueryRow(ctx, query, orgID).Scan(&features)
	if err != nil {
		return false, nil // Assume no access if no subscription
	}

	val, ok := features[feature]
	if !ok {
		return false, nil
	}

	switch v := val.(type) {
	case bool:
		return v, nil
	case int, float64:
		return true, nil // Numeric limits are handled separately
	default:
		return false, nil
	}
}
