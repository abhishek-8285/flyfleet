//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	finService "github.com/flyfleet/fleet-core-service/internal/domains/finance/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	fuelRepo "github.com/flyfleet/fleet-core-service/internal/domains/fuel/repository"
	fuelService "github.com/flyfleet/fleet-core-service/internal/domains/fuel/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/integrations_v1"
)

func TestFuelModuleIntegrationFlow(t *testing.T) {
	ctx := context.Background()
	container, dsn := startPostgres(t, ctx)
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	db, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(db.Close)
	require.NoError(t, createFuelSchema(ctx, db))

	repo := fuelRepo.NewFuelLogRepository(db)
	settingsRepo := fuelRepo.NewSettingsRepository(db)
	svc := fuelService.NewFuelService(
		repo,
		settingsRepo,
		fuelService.NewAnomalyDetectionService(nil, zap.NewNop()),
		finService.NewTaxCalculationService(),
		&noopBus{},
		zap.NewNop(),
	)

	orgID := uuid.New()
	userID := uuid.New()
	vehicleID := uuid.New()
	driverID := uuid.New()

	log, err := svc.CreateLog(ctx, orgID, userID, models.CreateFuelLogRequest{
		VehicleID:       vehicleID,
		DriverID:        driverID,
		StationName:     "Shell",
		FuelType:        "diesel",
		Quantity:        50,
		UnitPrice:       2,
		TotalCost:       100,
		Odometer:        1000,
		TransactionDate: time.Now().UTC(),
	})
	require.NoError(t, err)
	require.Equal(t, models.FuelLogStatusPending, log.Status)

	require.NoError(t, svc.ChangeStatus(ctx, log.ID, orgID, models.ChangeFuelLogStatusRequest{Status: models.FuelLogStatusRejected}))

	_, err = svc.UpdateLog(ctx, log.ID, orgID, models.UpdateFuelLogRequest{Notes: strPtr("nope")})
	require.ErrorIs(t, err, fuelService.ErrLogImmutable)

	_, err = svc.CreateLog(ctx, orgID, userID, models.CreateFuelLogRequest{
		VehicleID:       uuid.New(),
		DriverID:        driverID,
		StationName:     "Shell",
		FuelType:        "diesel",
		Quantity:        20,
		UnitPrice:       2,
		TotalCost:       40,
		Odometer:        1200,
		TransactionDate: time.Now().UTC(),
	})
	require.NoError(t, err)

	anomalyLog, err := svc.CreateLog(ctx, orgID, userID, models.CreateFuelLogRequest{
		VehicleID:       uuid.New(),
		DriverID:        driverID,
		StationName:     "Shell",
		FuelType:        "diesel",
		Quantity:        350,
		UnitPrice:       2,
		TotalCost:       700,
		Odometer:        1500,
		TransactionDate: time.Now().UTC(),
	})
	require.NoError(t, err)
	require.True(t, anomalyLog.AnomalyFlag)

	stats, err := svc.GetStats(ctx, orgID, models.FuelLogFilter{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, stats.TotalCost, 0.0)
}

func startPostgres(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	t.Helper()

	pg, err := pgcontainer.Run(ctx,
		"postgres:16-alpine",
		pgcontainer.WithDatabase("fuel_test"),
		pgcontainer.WithUsername("postgres"),
		pgcontainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(90*time.Second)),
	)
	require.NoError(t, err)

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	return pg, dsn
}

func createFuelSchema(ctx context.Context, db *pgxpool.Pool) error {
	schema := []string{
		`CREATE TABLE fuel_logs (
			id UUID PRIMARY KEY,
			organization_id UUID NOT NULL,
			vehicle_id UUID NOT NULL,
			driver_id UUID NOT NULL,
			station_name VARCHAR(255) NOT NULL,
			fuel_type VARCHAR(50) NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			unit_price DOUBLE PRECISION NOT NULL,
			total_cost DOUBLE PRECISION NOT NULL,
			odometer DOUBLE PRECISION NOT NULL,
			transaction_date TIMESTAMPTZ NOT NULL,
			status VARCHAR(20) NOT NULL,
			gst_number VARCHAR(15),
			gst_rate DOUBLE PRECISION,
			gst_amount DOUBLE PRECISION,
			invoice_number VARCHAR(100),
			vendor_gstin VARCHAR(15),
			receipt_url TEXT,
			notes TEXT,
			anomaly_flag BOOLEAN DEFAULT FALSE,
			anomaly_reason TEXT,
			created_by UUID NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE fuel_audit_logs (
			id UUID PRIMARY KEY,
			fuel_log_id UUID NOT NULL,
			action VARCHAR(50) NOT NULL,
			changed_by UUID NOT NULL,
			changed_at TIMESTAMPTZ DEFAULT NOW(),
			old_values JSONB,
			new_values JSONB,
			diff JSONB
		);`,
		`CREATE TABLE fuel_settings (
			organization_id UUID PRIMARY KEY,
			default_currency VARCHAR(3),
			allowed_fuel_types JSONB,
			validation_rules JSONB,
			auto_sync_enabled BOOLEAN,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("schema setup failed: %w", err)
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }

type noopBus struct{}

func (n *noopBus) Publish(_ context.Context, _ string, _ integrations_v1.EventEnvelope) error {
	return nil
}
func (n *noopBus) Subscribe(_ string, _ integrations_v1.EventHandler) {}
func (n *noopBus) Close() error                                       { return nil }
