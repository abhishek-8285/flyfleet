package service_test

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ------------------------------------------------------------------
// Mocks
// ------------------------------------------------------------------

type mockVehicleRepo struct{ mock.Mock }

func (m *mockVehicleRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}
func (m *mockVehicleRepo) Update(ctx context.Context, v *models.Vehicle) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}
func (m *mockVehicleRepo) Create(ctx context.Context, v *models.Vehicle) error {
	return m.Called(ctx, v).Error(0)
}
func (m *mockVehicleRepo) GetByOrgID(ctx context.Context, orgID uuid.UUID, l, o int) ([]*models.Vehicle, int64, error) {
	return nil, 0, nil
}
func (m *mockVehicleRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockVehicleRepo) AssignDriver(ctx context.Context, v, d uuid.UUID) error { return nil }
func (m *mockVehicleRepo) GetStats(ctx context.Context) (*models.VehicleStats, error) {
	return nil, nil
}
func (m *mockVehicleRepo) GetByRegistrationNumber(ctx context.Context, t uuid.UUID, r string) (*models.Vehicle, error) {
	return nil, nil
}
func (m *mockVehicleRepo) Search(ctx context.Context, p models.VehicleSearchParams) ([]*models.Vehicle, int64, error) {
	return nil, 0, nil
}

type mockComplianceRepo struct{ mock.Mock }

func (m *mockComplianceRepo) Get(ctx context.Context, id uuid.UUID) (*models.VehicleCompliance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VehicleCompliance), args.Error(1)
}
func (m *mockComplianceRepo) Upsert(ctx context.Context, c *models.VehicleCompliance) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockComplianceRepo) GetExpiringVehicles(ctx context.Context, t uuid.UUID, d int) ([]models.ComplianceExpiryAlert, error) {
	return nil, nil
}

type mockEventsRepo struct{ mock.Mock }

func (m *mockEventsRepo) Append(ctx context.Context, e *models.VehicleEvent) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockEventsRepo) List(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.VehicleEvent, int64, error) {
	return nil, 0, nil
}

// ------------------------------------------------------------------
// LifecycleService Tests
// ------------------------------------------------------------------

func setupLifecycleSvc() (*service.LifecycleService, *mockVehicleRepo, *mockComplianceRepo, *mockEventsRepo) {
	vr := &mockVehicleRepo{}
	cr := &mockComplianceRepo{}
	er := &mockEventsRepo{}
	svc := service.NewLifecycleService(vr, cr, er, zap.NewNop())
	return svc, vr, cr, er
}

// TestSetMaintenance_Idempotent — calling SetMaintenance on an already-MAINTENANCE vehicle
// should return the vehicle without calling Update again.
func TestSetMaintenance_Idempotent(t *testing.T) {
	svc, vr, _, er := setupLifecycleSvc()
	ctx := context.Background()
	vid := uuid.New()

	vehicle := &models.Vehicle{ID: vid, Status: models.StatusMaintenance, Version: 1}
	vr.On("GetByID", ctx, vid).Return(vehicle, nil)
	// Update should NOT be called (idempotent)
	er.On("Append", ctx, mock.Anything).Return(nil).Maybe()

	result, err := svc.SetMaintenance(ctx, vid)
	require.NoError(t, err)
	assert.Equal(t, models.StatusMaintenance, result.Status)
	vr.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestSetMaintenance_FromActive emits a status-changed audit event.
func TestSetMaintenance_FromActive(t *testing.T) {
	svc, vr, _, er := setupLifecycleSvc()
	ctx := context.Background()
	vid := uuid.New()

	vehicle := &models.Vehicle{ID: vid, Status: models.StatusActive, Version: 1}
	vr.On("GetByID", ctx, vid).Return(vehicle, nil)
	vr.On("Update", ctx, mock.Anything).Return(nil)
	er.On("Append", ctx, mock.MatchedBy(func(e *models.VehicleEvent) bool {
		return e.EventType == models.EventVehicleStatusChanged
	})).Return(nil)

	result, err := svc.SetMaintenance(ctx, vid)
	require.NoError(t, err)
	assert.Equal(t, models.StatusMaintenance, result.Status)
	er.AssertExpectations(t)
}

// TestSetMaintenance_DecommissionedGuard confirms DECOMMISSIONED vehicles cannot be set to MAINTENANCE.
func TestSetMaintenance_DecommissionedGuard(t *testing.T) {
	svc, vr, _, _ := setupLifecycleSvc()
	ctx := context.Background()
	vid := uuid.New()

	vehicle := &models.Vehicle{ID: vid, Status: models.StatusDecommissioned, Version: 1}
	vr.On("GetByID", ctx, vid).Return(vehicle, nil)

	_, err := svc.SetMaintenance(ctx, vid)
	assert.ErrorIs(t, err, service.ErrDecommissionedIrreversible)
}

// TestDecommission_Idempotent — already DECOMMISSIONED vehicle is returned without write.
func TestDecommission_Idempotent(t *testing.T) {
	svc, vr, _, _ := setupLifecycleSvc()
	ctx := context.Background()
	vid := uuid.New()

	vehicle := &models.Vehicle{ID: vid, Status: models.StatusDecommissioned, Version: 1}
	vr.On("GetByID", ctx, vid).Return(vehicle, nil)

	result, err := svc.Decommission(ctx, vid, "already done")
	require.NoError(t, err)
	assert.Equal(t, models.StatusDecommissioned, result.Status)
	vr.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestActivate_BlockedByLegalHold verifies compliance gate prevents activation.
func TestActivate_BlockedByLegalHold(t *testing.T) {
	svc, vr, cr, _ := setupLifecycleSvc()
	ctx := context.Background()
	vid := uuid.New()

	vehicle := &models.Vehicle{ID: vid, Status: models.StatusSuspended, Version: 1}
	vr.On("GetByID", ctx, vid).Return(vehicle, nil)
	cr.On("Get", ctx, vid).Return(&models.VehicleCompliance{LegalHold: true}, nil)

	_, err := svc.Activate(ctx, vid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "legal hold")
}
