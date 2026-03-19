//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pool           *pgxpool.Pool
	repo           *repository.DiagnosticsRepository
	telemetryService *service.TelemetryService
	commandService   *service.DeviceCommandService
	rolloutService   *service.FirmwareRolloutService
	healthService    *service.DeviceHealthService
	alertService     *service.DeviceAlertService
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres Container
	dbName := "testdb"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start container: %s", err))
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %s", err))
	}

	// 2. Connect to Pool
	pool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %s", err))
	}
	defer pool.Close()

	// 3. Run Migrations
	migrationFiles := []string{
		"062_diagnostics_devices.sql",
		"063_diagnostics_audit_logging.sql",
		"064_diagnostics_device_alerts.sql",
		"065_rollout_device_status.sql",
		"066_device_health_view.sql",
	}

	migrationDir := "../../../../../migrations"

	for _, file := range migrationFiles {
		content, err := os.ReadFile(filepath.Join(migrationDir, file))
		if err != nil {
			panic(fmt.Sprintf("failed to read migration file %s: %s", file, err))
		}
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			panic(fmt.Sprintf("failed to execute migration %s: %s", file, err))
		}
	}

	// 4. Init Repo & Services
	repo = repository.NewDiagnosticsRepository(pool)
	healthService = service.NewDeviceHealthService(repo)
	telemetryService = service.NewTelemetryService(repo, healthService)
	commandService = service.NewDeviceCommandService(repo)
	rolloutService = service.NewFirmwareRolloutService(repo)
	alertService = service.NewDeviceAlertService(repo)

	// 5. Run Tests
	code := m.Run()

	// 6. Cleanup
	if err := postgresContainer.Terminate(ctx); err != nil {
		panic(fmt.Sprintf("failed to terminate container: %s", err))
	}

	os.Exit(code)
}

func ptr[T any](v T) *T {
	return &v
}
