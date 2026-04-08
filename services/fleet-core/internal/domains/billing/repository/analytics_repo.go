package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetFinancialMetrics calculates system-wide SaaS metrics
func (r *AnalyticsRepository) GetFinancialMetrics(ctx context.Context) (*models.FinancialMetrics, error) {
	metrics := &models.FinancialMetrics{
		Timestamp:     time.Now(),
		RevenueByPlan: make(map[string]float64),
	}

	// 1. Calculate MRR, ARR, and Revenue by Plan
	// We join active subscriptions with their plan's FLAT pricing components
	query := `
		SELECT 
			p.name,
			SUM(pc.unit_price) as total_flat_revenue,
			pc.frequency
		FROM billing_subscriptions s
		JOIN billing_plans p ON s.plan_id = p.id
		JOIN billing_pricing_components pc ON p.id = pc.plan_id
		WHERE s.status = 'ACTIVE' AND pc.type = 'FLAT'
		GROUP BY p.name, pc.frequency
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate mrr/arr: %w", err)
	}
	defer rows.Close()

	var totalMRR, totalARR float64
	for rows.Next() {
		var planName string
		var revenue float64
		var freq string
		if err := rows.Scan(&planName, &revenue, &freq); err != nil {
			return nil, err
		}

		planMRR := 0.0
		switch freq {
		case "MONTHLY":
			planMRR = revenue
		case "YEARLY":
			planMRR = revenue / 12
		}

		totalMRR += planMRR
		totalARR += planMRR * 12
		metrics.RevenueByPlan[planName] += planMRR
	}
	metrics.MRR = totalMRR
	metrics.ARR = totalARR

	// 2. Calculate Active Subscription Count
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM billing_subscriptions WHERE status = 'ACTIVE'").Scan(&metrics.ActiveSubs)
	if err != nil {
		return nil, fmt.Errorf("failed to count active subs: %w", err)
	}

	// 3. Calculate Net Revenue (from ledger last 30 days)
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM billing_ledger_entries
		WHERE account_type = 'CASH' AND entry_type = 'DEBIT' 
		AND created_at >= CURRENT_DATE - INTERVAL '30 days'
	`).Scan(&metrics.NetRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate net revenue: %w", err)
	}

	return metrics, nil
}

// GetChurnMetrics calculates the churn rate over the last 30 days
func (r *AnalyticsRepository) GetChurnMetrics(ctx context.Context) (float64, error) {
	var churnCount int
	var totalActiveAtStart int

	// Count number of cancellations in last 30 days
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM billing_audit_logs 
		WHERE table_name = 'subscription' AND action = 'UPDATED'
		AND (new_data->>'status') = 'CANCELLED'
		AND created_at >= CURRENT_DATE - INTERVAL '30 days'
	`).Scan(&churnCount)
	if err != nil {
		return 0, err
	}

	// Total active 30 days ago
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM billing_subscriptions 
		WHERE status = 'ACTIVE' OR (status = 'CANCELLED' AND updated_at >= CURRENT_DATE - INTERVAL '30 days')
	`).Scan(&totalActiveAtStart)

	if totalActiveAtStart == 0 {
		return 0, nil
	}

	return float64(churnCount) / float64(totalActiveAtStart) * 100, nil
}

// GetLTV calculates Average Customer Lifetime Value
func (r *AnalyticsRepository) GetLTV(ctx context.Context) (float64, error) {
	var averageRevenuePerUser float64
	var churnRate float64

	// ARPU (Monthly)
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(mrr), 0) FROM (
			SELECT SUM(pc.unit_price * CASE WHEN pc.frequency = 'YEARLY' THEN 1.0/12.0 ELSE 1.0 END) as mrr
			FROM billing_subscriptions s
			JOIN billing_pricing_components pc ON s.plan_id = pc.plan_id
			WHERE s.status = 'ACTIVE' AND pc.type = 'FLAT'
			GROUP BY s.org_id
		) sub
	`).Scan(&averageRevenuePerUser)
	if err != nil {
		return 0, err
	}

	churnRate, err = r.GetChurnMetrics(ctx)
	if err != nil {
		return 0, err
	}

	if churnRate <= 0 {
		return averageRevenuePerUser * 12 * 3, nil // Fallback to 3-year projection if churn is 0
	}

	// LTV = ARPU / ChurnRate
	return averageRevenuePerUser / (churnRate / 100), nil
}
