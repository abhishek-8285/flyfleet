package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MeteringRepository struct {
	db *pgxpool.Pool
}

func NewMeteringRepository(db *pgxpool.Pool) *MeteringRepository {
	return &MeteringRepository{db: db}
}

// RecordConsumption logs a new usage event
func (r *MeteringRepository) RecordConsumption(ctx context.Context, event *models.ConsumptionEvent) error {
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}

	query := `INSERT INTO consumption_events (time, org_id, vehicle_id, meter_code, amount, metadata) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query, event.Time, event.OrgID, event.VehicleID, event.MeterCode, event.Amount, event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to record consumption: %w", err)
	}
	return nil
}

// GetUnbilledUsage aggregates usage events that haven't been invoiced yet
func (r *MeteringRepository) GetUnbilledUsage(ctx context.Context, orgID uuid.UUID) ([]models.UsageSummary, error) {
	query := `SELECT ce.meter_code, um.name, um.unit, SUM(ce.amount) as total, SUM(ce.amount * um.unit_price) as amount
	          FROM consumption_events ce
	          JOIN usage_meters um ON ce.meter_code = um.code
	          WHERE ce.org_id = $1 AND ce.processed_invoice_id IS NULL
	          GROUP BY ce.meter_code, um.name, um.unit`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unbilled usage: %w", err)
	}
	defer rows.Close()

	var summaries []models.UsageSummary
	for rows.Next() {
		var s models.UsageSummary
		err := rows.Scan(&s.MeterCode, &s.Name, &s.Unit, &s.Total, &s.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// MarkUsageAsBilled links consumption events to a generated invoice
func (r *MeteringRepository) MarkUsageAsBilled(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID) error {
	query := `UPDATE consumption_events SET processed_invoice_id = $1 
	          WHERE org_id = $2 AND processed_invoice_id IS NULL`

	_, err := r.db.Exec(ctx, query, invoiceID, orgID)
	if err != nil {
		return fmt.Errorf("failed to mark usage as billed: %w", err)
	}
	return nil
}

// ListMeters retrieves all available usage meters
func (r *MeteringRepository) ListMeters(ctx context.Context) ([]models.UsageMeter, error) {
	query := `SELECT id, code, name, description, unit, unit_price, currency, status, created_at FROM usage_meters`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query meters: %w", err)
	}
	defer rows.Close()

	var meters []models.UsageMeter
	for rows.Next() {
		var m models.UsageMeter
		err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Description, &m.Unit, &m.UnitPrice, &m.Currency, &m.Status, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meter: %w", err)
		}
		meters = append(meters, m)
	}
	return meters, nil
}
