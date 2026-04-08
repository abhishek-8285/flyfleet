package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// ListPlans retrieves all active billing plans with their pricing components
func (r *SubscriptionRepository) ListPlans(ctx context.Context) ([]models.Plan, error) {
	query := `SELECT id, name, description, is_active, created_at, updated_at 
	          FROM billing_plans WHERE is_active = true ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		// Fetch components for each plan
		p.Components, err = r.getPlanComponents(ctx, p.ID)
		if err != nil {
			return nil, err
		}

		plans = append(plans, p)
	}

	return plans, nil
}

// IdentifyOrganizationsForInvoicing returns org IDs that need a new invoice
func (r *SubscriptionRepository) IdentifyOrganizationsForInvoicing(ctx context.Context) ([]uuid.UUID, error) {
	query := `SELECT org_id FROM billing_subscriptions 
	          WHERE current_period_end <= NOW() AND status = 'ACTIVE' AND cancel_at_period_end = false`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions for invoicing: %w", err)
	}
	defer rows.Close()

	var orgIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan org_id: %w", err)
		}
		orgIDs = append(orgIDs, id)
	}
	return orgIDs, nil
}

func (r *SubscriptionRepository) getPlanComponents(ctx context.Context, planID uuid.UUID) ([]models.PricingComponent, error) {
	query := `SELECT id, plan_id, name, type, unit_price, currency, meter_event_name, frequency, created_at, updated_at 
	          FROM billing_pricing_components WHERE plan_id = $1`

	rows, err := r.db.Query(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query components: %w", err)
	}
	defer rows.Close()

	var components []models.PricingComponent
	for rows.Next() {
		var c models.PricingComponent
		err := rows.Scan(&c.ID, &c.PlanID, &c.Name, &c.Type, &c.UnitPrice, &c.Currency, &c.MeterEventName, &c.Frequency, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan component: %w", err)
		}
		components = append(components, c)
	}
	return components, nil
}

// GetSubscription retrieves the current subscription for an organization
func (r *SubscriptionRepository) GetSubscription(ctx context.Context, orgID uuid.UUID) (*models.Subscription, error) {
	var s models.Subscription
	query := `SELECT id, org_id, plan_id, status, current_period_start, current_period_end, trial_end, cancel_at_period_end, metadata, created_at, updated_at 
	          FROM billing_subscriptions WHERE org_id = $1`

	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&s.ID, &s.OrgID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEnd,
		&s.CancelAtPeriodEnd, &s.Metadata, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &s, nil
}

// GetPlan retrieves a billing plan with its pricing components by ID
func (r *SubscriptionRepository) GetPlan(ctx context.Context, planID uuid.UUID) (*models.Plan, error) {
	var p models.Plan
	query := `SELECT id, name, description, is_active, created_at, updated_at
	          FROM billing_plans WHERE id = $1`

	err := r.db.QueryRow(ctx, query, planID).Scan(
		&p.ID, &p.Name, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Fetch components
	components, err := r.getPlanComponents(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Components = components

	return &p, nil
}

// CreateSubscription initializes a new subscription for an organization
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	now := time.Now().UTC()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	query := `INSERT INTO billing_subscriptions (id, org_id, plan_id, status, current_period_start, current_period_end, trial_end, cancel_at_period_end, metadata, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(ctx, query,
		sub.ID, sub.OrgID, sub.PlanID, sub.Status,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.TrialEnd,
		sub.CancelAtPeriodEnd, sub.Metadata, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

// UpdateSubscription updates an existing subscription
func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	sub.UpdatedAt = time.Now().UTC()

	query := `UPDATE billing_subscriptions SET 
	          plan_id = $1, status = $2, current_period_start = $3, current_period_end = $4, 
	          trial_end = $5, cancel_at_period_end = $6, metadata = $7, updated_at = $8 
	          WHERE id = $9`

	_, err := r.db.Exec(ctx, query,
		sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.TrialEnd, sub.CancelAtPeriodEnd, sub.Metadata, sub.UpdatedAt, sub.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}
