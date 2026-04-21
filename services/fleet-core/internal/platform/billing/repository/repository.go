package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/platform/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository interface {
	GetSubscriptionByOrganization(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error)
	GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error)
	RecordUsage(ctx context.Context, record *models.UsageRecord) error
	GetUsage(ctx context.Context, tenantID uuid.UUID, metric string) (int, error)
}

type billingRepository struct {
	db *pgxpool.Pool
}

func NewBillingRepository(db *pgxpool.Pool) BillingRepository {
	return &billingRepository{db: db}
}

func (r *billingRepository) GetSubscriptionByOrganization(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error) {
	s := &models.Subscription{}
	query := `SELECT id, organization_id, plan_id, status, start_date, end_date FROM subscriptions WHERE organization_id = $1 AND status = 'active'`
	err := r.db.QueryRow(ctx, query, tenantID).Scan(&s.ID, &s.OrganizationID, &s.PlanID, &s.Status, &s.StartDate, &s.EndDate)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *billingRepository) GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error) {
	p := &models.Plan{}
	query := `SELECT id, name, monthly_price, max_vehicles, max_drivers, features FROM plans WHERE id = $1`
	err := r.db.QueryRow(ctx, query, planID).Scan(&p.ID, &p.Name, &p.MonthlyPrice, &p.MaxVehicles, &p.MaxDrivers, &p.Features)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *billingRepository) RecordUsage(ctx context.Context, record *models.UsageRecord) error {
	query := `INSERT INTO usage_records (organization_id, metric, value, recorded_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, record.OrganizationID, record.Metric, record.Value, record.RecordedAt)
	return err
}

func (r *billingRepository) GetUsage(ctx context.Context, tenantID uuid.UUID, metric string) (int, error) {
	var total int
	query := `SELECT COALESCE(SUM(value), 0) FROM usage_records WHERE organization_id = $1 AND metric = $2`
	err := r.db.QueryRow(ctx, query, tenantID, metric).Scan(&total)
	return total, err
}
