//go:build integration
// +build integration

package maintenance_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/handlers"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/repository"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/service"
)

type MaintenanceIntegrationTestSuite struct {
	suite.Suite
	ctx               context.Context
	pool              *pgxpool.Pool
	postgresContainer *postgres.PostgresContainer
	natsContainer     testcontainers.Container
	natsConn          *nats.Conn

	// Services
	woService          *service.WorkOrderService
	maintenanceService *service.MaintenanceService
	procurementService *service.ProcurementService
	analyticsService   *service.AnalyticsService
	downtimeService    *service.DowntimeService

	// Handlers
	analyticsHandler *handlers.AnalyticsHandler

	// IDs
	tenantID uuid.UUID
}

func (s *MaintenanceIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.tenantID = uuid.New()

	// 1. Start Postgres
	var err error
	s.postgresContainer, err = postgres.Run(s.ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	s.NoError(err)

	connStr, err := s.postgresContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.NoError(err)

	s.pool, err = pgxpool.New(s.ctx, connStr)
	s.NoError(err)

	// 2. Bootstrap Missing Tables (that migrations expect to exist)
	bootstrapSQL := `
		CREATE TABLE IF NOT EXISTS organizations (
			id UUID PRIMARY KEY,
			name TEXT,
			tax_id TEXT,
			status TEXT
		);
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			organization_id UUID,
			email TEXT UNIQUE,
			full_name TEXT
		);
		CREATE TABLE IF NOT EXISTS vehicles (
			id UUID PRIMARY KEY,
			tenant_id UUID,
			license_plate TEXT,
			vin TEXT,
			vehicle_type TEXT,
			fuel_type TEXT,
			status TEXT,
			make TEXT,
			model TEXT,
			year INT,
			color TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMPTZ
		);
		CREATE TABLE IF NOT EXISTS maintenance_logs (
			id UUID PRIMARY KEY,
			organization_id UUID,
			vehicle_id UUID,
			schedule_id TEXT,
			type TEXT,
			name TEXT,
			description TEXT,
			start_date TIMESTAMPTZ,
			end_date TIMESTAMPTZ,
			duration_minutes INT,
			performed_by UUID,
			cost DECIMAL(15, 2),
			total_cost DECIMAL(15, 2),
			actual_cost DECIMAL(15, 2),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			completion_date TIMESTAMPTZ
		);
		-- Placeholder tables for FK constraints in migrations
		CREATE TABLE IF NOT EXISTS subscriptions (id UUID PRIMARY KEY, organization_id UUID);
		CREATE TABLE IF NOT EXISTS roles (id UUID PRIMARY KEY, organization_id UUID);
		CREATE TABLE IF NOT EXISTS maintenance_vendors (
			id UUID PRIMARY KEY,
			organization_id UUID,
			name TEXT,
			status TEXT
		);
	`
	_, err = s.pool.Exec(s.ctx, bootstrapSQL)
	s.NoError(err)

	// 2.5 Seed Test Organization
	_, err = s.pool.Exec(s.ctx, "INSERT INTO organizations (id, name, status) VALUES ($1, $2, $3)", s.tenantID, "Test Org", "active")
	s.NoError(err)

	migrationFiles := []string{
		"034_maintenance_system.sql",
		"071_vehicle_enterprise.sql",
		"083_organizations_table.sql",
		"255_maintenance_enterprise_upgrade.sql",
	}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsDir := filepath.Join(basepath, "../../../migrations")
	for _, file := range migrationFiles {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		s.NoError(err, "Failed to read migration: %s", file)

		// Strip CONCURRENTLY from indexes as it cannot run in transactions/certain test setups
		sqlContent := strings.ReplaceAll(string(content), "CONCURRENTLY", "")

		_, err = s.pool.Exec(s.ctx, sqlContent)
		s.NoError(err, "Failed migration: %s", file)
	}

	// 3. Start NATS
	s.natsContainer, err = testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nats:2.9-alpine",
			ExposedPorts: []string{"4222/tcp"},
			WaitingFor:   wait.ForListeningPort("4222/tcp"),
		},
		Started: true,
	})
	s.NoError(err)

	endpoint, err := s.natsContainer.Endpoint(s.ctx, "")
	s.NoError(err)
	s.natsConn, err = nats.Connect("nats://" + endpoint)
	s.NoError(err)

	// 4. Initialize Repos & Services
	logger, _ := zap.NewDevelopment()
	repo := repository.NewMaintenanceRepository(s.pool)
	inventoryRepo := repository.NewInventoryRepository(s.pool)
	woRepo := repository.NewWorkOrderRepository(s.pool)
	partsRepo := repository.NewWorkOrderPartsRepository(s.pool)
	settingsRepo := repository.NewSettingsRepository(s.pool)
	vendorRepo := repository.NewVendorRepository(s.pool)
	inspectionRepo := repository.NewInspectionRepository(s.pool)

	s.woService = service.NewWorkOrderService(woRepo, partsRepo, inventoryRepo, settingsRepo, vendorRepo, inspectionRepo, repo, s.natsConn, logger)
	s.maintenanceService = service.NewMaintenanceService(repo, logger)
	s.procurementService = service.NewProcurementService(repo, inventoryRepo, logger)
	s.analyticsService = service.NewAnalyticsService(repo, logger)
	s.downtimeService = service.NewDowntimeService(repo, logger)

	s.analyticsHandler = handlers.NewAnalyticsHandler(s.analyticsService, s.downtimeService, logger)
}

func (s *MaintenanceIntegrationTestSuite) TearDownSuite() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
	if s.natsContainer != nil {
		s.natsContainer.Terminate(s.ctx)
	}
	if s.pool != nil {
		s.pool.Close()
	}
	if s.postgresContainer != nil {
		s.postgresContainer.Terminate(s.ctx)
	}
}

func TestMaintenanceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceIntegrationTestSuite))
}

// Phase 1: Technician Management Tests
func (s *MaintenanceIntegrationTestSuite) TestPhase1_TechnicianManagement() {
	var err error
	techs := []struct {
		name  string
		skill string
	}{
		{"John Engine", "engine"},
		{"Mike Trans", "transmission"},
		{"Sarah Spark", "electrical"},
	}

	techIDs := make(map[string]uuid.UUID)
	for _, t := range techs {
		id := uuid.New()
		err := s.maintenanceService.CreateTechnician(s.ctx, &models.Technician{
			ID:                 id,
			OrganizationID:     s.tenantID,
			Name:               t.name,
			Specialization:     t.skill,
			AvailabilityStatus: models.TechnicianStatusAvailable,
		})
		s.NoError(err)

		err = s.maintenanceService.AddTechnicianSkill(s.ctx, id, t.skill, 5)
		s.NoError(err)
		techIDs[t.skill] = id
	}

	vehicleID := uuid.New()
	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "TECH-001")
	s.NoError(err)

	req := &models.CreateWorkOrderRequest{
		VehicleID:       vehicleID,
		Title:           "Engine Oil Leak Repair",
		Description:     "Technician needs engine skills",
		MaintenanceType: models.WOMaintenanceTypeCorrective,
		Priority:        models.WorkOrderPriorityHigh,
	}

	wo, err := s.woService.CreateWorkOrder(s.ctx, s.tenantID, uuid.New(), req)
	s.NoError(err)
	s.NotNil(wo)

	err = s.woService.AutoAssignTechnician(s.ctx, wo)
	s.NoError(err)

	updatedWO, _ := s.woService.GetWorkOrder(s.ctx, wo.ID)
	s.Equal(techIDs["engine"], *updatedWO.AssignedTo)

	tech, _ := s.maintenanceService.GetTechnician(s.ctx, techIDs["engine"])
	s.Equal(1, tech.CurrentWorkload)
}

// Phase 2: Warranty Management Tests
func (s *MaintenanceIntegrationTestSuite) TestPhase2_WarrantyManagement() {
	vehicleID := uuid.New()
	partNumber := "ENG-001"

	err := s.maintenanceService.CreateWarranty(s.ctx, &models.ComponentWarranty{
		OrganizationID:   s.tenantID,
		VehicleID:        vehicleID,
		PartNumber:       partNumber,
		WarrantyProvider: "FleetParts Corp",
		WarrantyStart:    time.Now().AddDate(-1, 0, 0),
		WarrantyEnd:      time.Now().AddDate(1, 0, 0),
	})
	s.NoError(err)

	covered, err := s.maintenanceService.CheckWarrantyCoverage(s.ctx, vehicleID, []string{partNumber})
	s.NoError(err)
	s.Len(covered, 1)
	s.Equal(partNumber, covered[0].PartNumber)

	alerts, err := s.maintenanceService.GetWarrantyExpiryAlerts(s.ctx, s.tenantID, 30)
	s.NoError(err)
	s.Len(alerts, 0)

	err = s.maintenanceService.CreateWarranty(s.ctx, &models.ComponentWarranty{
		OrganizationID: s.tenantID,
		VehicleID:      vehicleID,
		PartNumber:     "EXP-001",
		WarrantyEnd:    time.Now().AddDate(0, 0, 5),
	})
	s.NoError(err)

	alerts, err = s.maintenanceService.GetWarrantyExpiryAlerts(s.ctx, s.tenantID, 30)
	s.NoError(err)
	s.Len(alerts, 1)
}

// Phase 3: Procurement Flow Tests
func (s *MaintenanceIntegrationTestSuite) TestPhase3_ProcurementFlow() {
	partID := uuid.New()

	err := s.pool.QueryRow(s.ctx, `
		INSERT INTO inventory_parts (id, organization_id, part_number, name, quantity_available, reorder_level, unit_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		partID, s.tenantID, "PART-LOW-001", "Low Stock Part", 3, 5, 100.0).Scan(&partID)
	s.NoError(err)

	err = s.procurementService.AutoCreatePOIfLowStock(s.ctx, s.tenantID, partID)
	s.NoError(err)

	var poStatus string
	var poID uuid.UUID
	err = s.pool.QueryRow(s.ctx, "SELECT id, status FROM purchase_orders WHERE organization_id = $1 ORDER BY created_at DESC LIMIT 1", s.tenantID).Scan(&poID, &poStatus)
	s.NoError(err)
	s.Equal("DRAFT", poStatus)

	err = s.procurementService.ReceivePO(s.ctx, poID)
	s.NoError(err)

	var qty int
	err = s.pool.QueryRow(s.ctx, "SELECT quantity_available FROM inventory_parts WHERE id = $1", partID).Scan(&qty)
	s.NoError(err)
	s.Equal(13, qty)
}

// Phase 4: Predictive Maintenance Engine
func (s *MaintenanceIntegrationTestSuite) TestPhase4_PredictiveMaintenance() {
	var err error
	vehicleID := uuid.New()
	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "PREDICT-001")
	s.NoError(err)

	for i := 0; i < 5; i++ {
		_, err := s.pool.Exec(s.ctx, `
			INSERT INTO maintenance_logs (id, organization_id, vehicle_id, type, start_date, completion_date, cost, total_cost, created_at)
			VALUES ($1, $2, $3, 'corrective', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days', 100.0, 100.0, NOW())`,
			uuid.New(), s.tenantID, vehicleID)
		s.NoError(err)
	}

	worker := service.NewPredictiveMaintenanceWorker(
		repository.NewMaintenanceRepository(s.pool),
		s.woService,
		zap.NewNop(),
	)

	worker.RunEvaluation(s.ctx)

	var count int
	err = s.pool.QueryRow(s.ctx, "SELECT COUNT(*) FROM work_orders WHERE vehicle_id = $1 AND title = 'PREDICTIVE PREVENTIVE MAINTENANCE'", vehicleID).Scan(&count)
	s.NoError(err)
	s.True(count > 0)
}

// Phase 5: Downtime Tracking
func (s *MaintenanceIntegrationTestSuite) TestPhase5_DowntimeTracking() {
	var err error
	vehicleID := uuid.New()
	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "DOWNTIME-001")
	s.NoError(err)

	req := &models.CreateWorkOrderRequest{
		VehicleID:       vehicleID,
		Title:           "Downtime Test WO",
		MaintenanceType: models.WOMaintenanceTypeCorrective,
	}
	wo, err := s.woService.CreateWorkOrder(s.ctx, s.tenantID, uuid.New(), req)
	s.NoError(err)

	err = s.woService.UpdateWorkOrderStatus(s.ctx, wo.ID, models.WorkOrderStatusInProgress)
	s.NoError(err)

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()
	err = s.downtimeService.RecordDowntime(s.ctx, &models.VehicleDowntime{
		OrganizationID: s.tenantID,
		VehicleID:      vehicleID,
		DowntimeStart:  start,
		DowntimeEnd:    &end,
		Reason:         "Work Order " + wo.ID.String(),
	})
	s.NoError(err)

	metrics, err := s.downtimeService.CalculateVehicleMetrics(s.ctx, vehicleID)
	s.NoError(err)
	s.Equal(120.0, metrics.MTTR)
}

// Phase 6: Analytics Endpoints
func (s *MaintenanceIntegrationTestSuite) TestPhase6_AnalyticsEndpoints() {
	var err error
	vehicleID := uuid.New()
	r := chi.NewRouter()
	s.analyticsHandler.RegisterRoutes(r)

	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "ANALYTIC-001")
	s.NoError(err)
	_, _ = s.pool.Exec(s.ctx, `INSERT INTO maintenance_vendors (id, organization_id, name, status) VALUES ($1, $2, $3, $4)`, uuid.New(), s.tenantID, "Main Vendor", "active")

	_, _ = s.pool.Exec(s.ctx, `INSERT INTO maintenance_logs (id, organization_id, vehicle_id, cost, total_cost, type, start_date) VALUES ($1, $2, $3, 500.0, 500.0, 'routine', NOW())`, uuid.New(), s.tenantID, vehicleID)

	req := httptest.NewRequest("GET", "/analytics/costs", nil)
	// Inject org ID into context
	ctx := context.WithValue(req.Context(), middleware.TenantIDKey, s.tenantID.String())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
}

// Phase 7: Asset Lifecycle Recommendation
func (s *MaintenanceIntegrationTestSuite) TestPhase7_AssetLifecycle() {
	var err error
	vehicleID := uuid.New()
	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "LIFE-001")
	s.NoError(err)

	for i := 0; i < 10; i++ {
		_, err = s.pool.Exec(s.ctx, `INSERT INTO maintenance_logs (id, organization_id, vehicle_id, cost, type, start_date) VALUES ($1, $2, $3, 15000.0, 'corrective', NOW())`, uuid.New(), s.tenantID, vehicleID)
		s.NoError(err)
	}

	err = s.analyticsService.UpdateVehicleLifecycleMetrics(s.ctx, vehicleID)
	s.NoError(err)

	var recommendation string
	err = s.pool.QueryRow(s.ctx, "SELECT replacement_recommendation FROM vehicle_lifecycle_metrics WHERE vehicle_id = $1", vehicleID).Scan(&recommendation)
	s.NoError(err)
	s.Contains(recommendation, "REPLACE")
}

// Phase 8: Inventory + Work Order Integration
func (s *MaintenanceIntegrationTestSuite) TestPhase8_InventoryWOIntegration() {
	var err error
	partID := uuid.New()
	woID := uuid.New()
	vehicleID := uuid.New()
	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "INV-001")
	s.NoError(err)

	_, _ = s.pool.Exec(s.ctx, "INSERT INTO work_orders (id, organization_id, vehicle_id, title, maintenance_type, priority, created_by) VALUES ($1, $2, $3, 'INV TEST', 'routine', 'LOW', $4)", woID, s.tenantID, vehicleID, uuid.New())
	_, _ = s.pool.Exec(s.ctx, "INSERT INTO inventory_parts (id, organization_id, part_number, name, quantity_available, unit_cost) VALUES ($1, $2, 'PART-INV-001', 'Inv Test Part', 10, 50.0)", partID, s.tenantID)

	_, err = s.woService.AssignPart(s.ctx, woID, &models.AssignPartRequest{PartID: partID, Quantity: 2})
	s.NoError(err)

	var qty int
	_ = s.pool.QueryRow(s.ctx, "SELECT quantity_available FROM inventory_parts WHERE id = $1", partID).Scan(&qty)
	s.Equal(8, qty)
}

// Phase 9: Event Bus Validation
func (s *MaintenanceIntegrationTestSuite) TestPhase9_EventBusValidation() {
	var err error
	sub, _ := s.natsConn.SubscribeSync("fleet.maintenance.work_order.created")

	vehicleID := uuid.New()
	_, err = s.pool.Exec(s.ctx, `INSERT INTO vehicles (id, tenant_id, registration_number) VALUES ($1, $2, $3)`, vehicleID, s.tenantID, "EVENT-001")
	s.NoError(err)

	req := &models.CreateWorkOrderRequest{
		VehicleID:       vehicleID,
		Title:           "Event Test WO",
		MaintenanceType: models.WOMaintenanceTypeRoutine,
	}
	_, err = s.woService.CreateWorkOrder(s.ctx, s.tenantID, uuid.New(), req)
	s.NoError(err)

	msg, err := sub.NextMsg(2 * time.Second)
	s.NoError(err)
	s.NotNil(msg)
}

// Phase 10: Load Test
func (s *MaintenanceIntegrationTestSuite) TestPhase10_LoadTest() {
	count := 20
	techCount := 5

	for i := 0; i < techCount; i++ {
		_ = s.maintenanceService.CreateTechnician(s.ctx, &models.Technician{
			OrganizationID:     s.tenantID,
			Name:               fmt.Sprintf("Tech %d", i),
			Specialization:     "engine",
			AvailabilityStatus: models.TechnicianStatusAvailable,
		})
	}

	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		go func(idx int) {
			req := &models.CreateWorkOrderRequest{
				VehicleID:       uuid.New(),
				Title:           fmt.Sprintf("Load Test WO %d", idx),
				Description:     "Requires engine skills",
				MaintenanceType: models.WOMaintenanceTypeRoutine,
			}
			wo, err := s.woService.CreateWorkOrder(s.ctx, s.tenantID, uuid.New(), req)
			if err != nil {
				errs <- err
				return
			}
			errs <- s.woService.AutoAssignTechnician(s.ctx, wo)
		}(i)
	}

	for i := 0; i < count; i++ {
		s.NoError(<-errs)
	}
}
