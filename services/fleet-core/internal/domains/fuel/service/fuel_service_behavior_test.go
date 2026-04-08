package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	finService "github.com/flyfleet/fleet-core/internal/domains/finance/service"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/repository"
)

type fakeFuelRepo struct {
	logs          map[uuid.UUID]*models.FuelLog
	lastOdometers map[string]float64
	dup           bool
}

var _ repository.FuelLogRepository = (*fakeFuelRepo)(nil)

func newFakeFuelRepo() *fakeFuelRepo {
	return &fakeFuelRepo{
		logs:          make(map[uuid.UUID]*models.FuelLog),
		lastOdometers: make(map[string]float64),
	}
}

func (f *fakeFuelRepo) Create(_ context.Context, log *models.FuelLog) error {
	cpy := *log
	f.logs[log.ID] = &cpy
	f.lastOdometers[log.VehicleID.String()+":"+log.OrganizationID.String()] = log.Odometer
	return nil
}

func (f *fakeFuelRepo) GetByID(_ context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error) {
	log, ok := f.logs[id]
	if !ok || log.OrganizationID != orgID {
		return nil, nil
	}
	cpy := *log
	return &cpy, nil
}

func (f *fakeFuelRepo) List(_ context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error) {
	logs := make([]models.FuelLog, 0)
	for _, l := range f.logs {
		if l.OrganizationID == filter.OrganizationID {
			logs = append(logs, *l)
		}
	}
	return &models.FuelLogListResponse{Total: len(logs), Page: filter.Page, PageSize: filter.PageSize, Logs: logs}, nil
}

func (f *fakeFuelRepo) Update(_ context.Context, log *models.FuelLog) error {
	cpy := *log
	f.logs[log.ID] = &cpy
	return nil
}

func (f *fakeFuelRepo) Delete(_ context.Context, id uuid.UUID, _ uuid.UUID) error {
	delete(f.logs, id)
	return nil
}

func (f *fakeFuelRepo) UpdateStatus(_ context.Context, id uuid.UUID, _ uuid.UUID, status models.FuelLogStatus) error {
	if l, ok := f.logs[id]; ok {
		l.Status = status
	}
	return nil
}

func (f *fakeFuelRepo) GetStats(_ context.Context, _ uuid.UUID, _ models.FuelLogFilter) (*models.FuelLogStats, error) {
	return &models.FuelLogStats{AvgCostPerKm: 3.4, AvgConsumption: 4.5}, nil
}

func (f *fakeFuelRepo) GetLastOdometer(_ context.Context, vehicleID uuid.UUID, orgID uuid.UUID) (float64, error) {
	return f.lastOdometers[vehicleID.String()+":"+orgID.String()], nil
}

func (f *fakeFuelRepo) CheckDuplicate(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time, _ int) (bool, error) {
	return f.dup, nil
}

func (f *fakeFuelRepo) GetVehicleSpecs(_ context.Context, _ uuid.UUID) (*models.VehicleSpecs, error) {
	return nil, nil
}

func (f *fakeFuelRepo) GetDriverFuelLimit(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.DriverFuelLimit, error) {
	return nil, nil
}

func (f *fakeFuelRepo) GetOrgFuelSettings(_ context.Context, _ uuid.UUID) (*models.OrgFuelSettings, error) {
	return nil, nil
}

func (f *fakeFuelRepo) GetFuelUsage(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ time.Time) (float64, float64, error) {
	return 0, 0, nil
}

func (f *fakeFuelRepo) GetLogsForAnomalyDetection(_ context.Context, _ time.Time) ([]models.FuelLog, error) {
	return nil, nil
}

func (f *fakeFuelRepo) MarkAsAnomalous(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}

func (f *fakeFuelRepo) CreateVendor(_ context.Context, _ *models.FuelVendor) error {
	return nil
}

func (f *fakeFuelRepo) ListVendors(_ context.Context, _ uuid.UUID) ([]models.FuelVendor, error) {
	return nil, nil
}

func (f *fakeFuelRepo) CreateAuditLog(_ context.Context, _ *models.FuelAuditLog) error {
	return nil
}

// CreateWithLimitValidation is a test double that calls Create
// It implements theFuelLogRepository interface extension for atomic limit validation
func (f *fakeFuelRepo) CreateWithLimitValidation(_ context.Context, log *models.FuelLog, _quantity, _amount float64) error {
	// For testing, just create the log without actual limit enforcement
	// Tests that need to validate limit behavior should use a specialized mock
	return f.Create(context.Background(), log)
}

type fakeSettingsRepo struct{}

var _ repository.SettingsRepository = (*fakeSettingsRepo)(nil)

func (f *fakeSettingsRepo) Get(_ context.Context, _ uuid.UUID) (*models.FuelSettings, error) {
	return nil, nil
}

func (f *fakeSettingsRepo) Upsert(_ context.Context, _ *models.FuelSettings) error {
	return nil
}

func (f *fakeSettingsRepo) GetOrDefault(_ context.Context, orgID uuid.UUID) (*models.FuelSettings, error) {
	return &models.FuelSettings{
		OrganizationID:   orgID,
		DefaultCurrency:  "USD",
		AllowedFuelTypes: json.RawMessage(`["diesel"]`),
		ValidationRules:  json.RawMessage(`{"odometer_check":true}`),
		AutoSyncEnabled:  true,
	}, nil
}

func newTestFuelService(repo *fakeFuelRepo) FuelService {
	logger := zap.NewNop()
	return NewFuelService(
		repo,
		&fakeSettingsRepo{},
		NewAnomalyDetectionService(nil, logger),
		finService.NewTaxCalculationService(),
		logger,
	)
}

func validCreateRequest(vehicleID, driverID uuid.UUID) models.CreateFuelLogRequest {
	return models.CreateFuelLogRequest{
		VehicleID:       vehicleID,
		DriverID:        driverID,
		StationName:     "Shell",
		FuelType:        "diesel",
		Quantity:        40,
		UnitPrice:       2,
		TotalCost:       80,
		Odometer:        1000,
		TransactionDate: time.Now().UTC(),
	}
}

func TestCreateLogRejectsCostMismatch(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)
	orgID := uuid.New()
	req := validCreateRequest(uuid.New(), uuid.New())
	req.TotalCost = 81

	_, err := svc.CreateLog(context.Background(), orgID, uuid.New(), req)
	require.ErrorIs(t, err, ErrInvalidCost)
}

func TestCreateLogRejectsOdometerRollback(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)
	orgID := uuid.New()
	vehicleID := uuid.New()
	repo.lastOdometers[vehicleID.String()+":"+orgID.String()] = 2000

	req := validCreateRequest(vehicleID, uuid.New())
	req.Odometer = 1999

	_, err := svc.CreateLog(context.Background(), orgID, uuid.New(), req)
	require.ErrorIs(t, err, ErrInvalidOdometer)
}

func TestCreateLogRejectsDuplicate(t *testing.T) {
	repo := newFakeFuelRepo()
	repo.dup = true
	svc := newTestFuelService(repo)

	_, err := svc.CreateLog(context.Background(), uuid.New(), uuid.New(), validCreateRequest(uuid.New(), uuid.New()))
	require.ErrorIs(t, err, ErrDuplicateFuelLog)
}

func TestRejectedLogsAreImmutable(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)
	orgID := uuid.New()
	logID := uuid.New()
	repo.logs[logID] = &models.FuelLog{
		ID:             logID,
		OrganizationID: orgID,
		Status:         models.FuelLogStatusRejected,
		CreatedBy:      uuid.New(),
	}

	_, err := svc.UpdateLog(context.Background(), logID, orgID, models.UpdateFuelLogRequest{Notes: strPtr("x")})
	require.ErrorIs(t, err, ErrLogImmutable)

	err = svc.DeleteLog(context.Background(), logID, orgID)
	require.ErrorIs(t, err, ErrLogImmutable)
}

func TestStatusWorkflowOnlyPendingAllowed(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)
	orgID := uuid.New()
	logID := uuid.New()
	repo.logs[logID] = &models.FuelLog{ID: logID, OrganizationID: orgID, Status: models.FuelLogStatusPending, CreatedBy: uuid.New()}

	err := svc.ChangeStatus(context.Background(), logID, orgID, models.ChangeFuelLogStatusRequest{Status: models.FuelLogStatusApproved})
	require.NoError(t, err)

	err = svc.ChangeStatus(context.Background(), logID, orgID, models.ChangeFuelLogStatusRequest{Status: models.FuelLogStatusRejected})
	require.ErrorIs(t, err, ErrStatusNotPending)
}

func TestCreateLogFlagsAnomaly(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)
	orgID := uuid.New()
	req := validCreateRequest(uuid.New(), uuid.New())
	req.Quantity = 301
	req.TotalCost = 602

	log, err := svc.CreateLog(context.Background(), orgID, uuid.New(), req)
	require.NoError(t, err)
	require.True(t, log.AnomalyFlag)
	require.NotNil(t, log.AnomalyReason)
}

func TestGetSettingsReturnsDefaults(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)
	orgID := uuid.New()

	settings, err := svc.GetSettings(context.Background(), orgID)
	require.NoError(t, err)
	require.Equal(t, "USD", settings.DefaultCurrency)
}

func TestGetStatsReturnsComputedFields(t *testing.T) {
	repo := newFakeFuelRepo()
	svc := newTestFuelService(repo)

	stats, err := svc.GetStats(context.Background(), uuid.New(), models.FuelLogFilter{})
	require.NoError(t, err)
	require.Equal(t, 3.4, stats.AvgCostPerKm)
	require.Equal(t, 4.5, stats.AvgConsumption)
}

func strPtr(s string) *string { return &s }
