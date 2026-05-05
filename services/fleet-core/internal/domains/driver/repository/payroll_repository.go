package repository

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PayrollRepository struct {
	db *pgxpool.Pool
}

func NewPayrollRepository(db *pgxpool.Pool) *PayrollRepository {
	return &PayrollRepository{db: db}
}

func (r *PayrollRepository) CreateSalaryConfig(ctx context.Context, config *models.SalaryConfig) error {
	query := `
		INSERT INTO driver_salary (
			id, driver_id, base_salary, per_km_rate, daily_allowance, night_allowance, effective_from
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		config.ID, config.DriverID, config.BaseSalary, config.PerKmRate,
		config.DailyAllowance, config.NightAllowance, config.EffectiveFrom,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *PayrollRepository) CreateTripEarning(ctx context.Context, earning *models.TripEarning) error {
	query := `
		INSERT INTO driver_trip_earnings (
			id, driver_id, trip_id, distance_km, earnings, bonus, penalty
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if earning.ID == uuid.Nil {
		earning.ID = uuid.New()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		earning.ID, earning.DriverID, earning.TripID, earning.DistanceKm,
		earning.Earnings, earning.Bonus, earning.Penalty,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *PayrollRepository) GetSalaryConfig(ctx context.Context, driverID uuid.UUID) (*models.SalaryConfig, error) {
	query := `
		SELECT id, driver_id, base_salary, per_km_rate, daily_allowance, night_allowance, effective_from
		FROM driver_salary
		WHERE driver_id = $1
		ORDER BY effective_from DESC
		LIMIT 1`

	var c models.SalaryConfig
	start := time.Now()
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&c.ID, &c.DriverID, &c.BaseSalary, &c.PerKmRate,
		&c.DailyAllowance, &c.NightAllowance, &c.EffectiveFrom,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *PayrollRepository) GetTripEarnings(ctx context.Context, driverID uuid.UUID, start, end time.Time) ([]*models.TripEarning, error) {
	query := `
		SELECT id, driver_id, trip_id, distance_km, earnings, bonus, penalty, created_at
		FROM driver_trip_earnings
		WHERE driver_id = $1 AND created_at >= $2 AND created_at <= $3`

	queryStart := time.Now()
	rows, err := r.db.Query(ctx, query, driverID, start, end)
	metrics.RecordDBQuery(query, time.Since(queryStart))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var earnings []*models.TripEarning
	for rows.Next() {
		var e models.TripEarning
		err := rows.Scan(&e.ID, &e.DriverID, &e.TripID, &e.DistanceKm, &e.Earnings, &e.Bonus, &e.Penalty, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		earnings = append(earnings, &e)
	}
	return earnings, nil
}

func (r *PayrollRepository) CreatePayout(ctx context.Context, payout *models.Payout) error {
	query := `
		INSERT INTO driver_payouts (
			id, driver_id, period_start, period_end, total_amount, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if payout.ID == uuid.Nil {
		payout.ID = uuid.New()
	}
	if payout.CreatedAt.IsZero() {
		payout.CreatedAt = time.Now()
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		payout.ID, payout.DriverID, payout.PeriodStart, payout.PeriodEnd,
		payout.TotalAmount, payout.Status, payout.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}
