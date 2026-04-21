package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedmw "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/service"
)

type mockFuelService struct {
	lastFilter models.FuelLogFilter
	lastTenantID  uuid.UUID
	lastUserID uuid.UUID
}

var _ service.FuelService = (*mockFuelService)(nil)

func (m *mockFuelService) CreateLog(_ context.Context, tenantID uuid.UUID, userID uuid.UUID, req models.CreateFuelLogRequest) (*models.FuelLog, error) {
	m.lastTenantID = tenantID
	m.lastUserID = userID
	return &models.FuelLog{ID: uuid.New(), OrganizationID: tenantID, VehicleID: req.VehicleID, DriverID: req.DriverID, Status: models.FuelLogStatusPending}, nil
}
func (m *mockFuelService) GetLog(_ context.Context, _ uuid.UUID, tenantID uuid.UUID) (*models.FuelLog, error) {
	return &models.FuelLog{ID: uuid.New(), OrganizationID: tenantID}, nil
}
func (m *mockFuelService) ListLogs(_ context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error) {
	m.lastFilter = filter
	return &models.FuelLogListResponse{Total: 0, Page: 1, PageSize: 20, Logs: []models.FuelLog{}}, nil
}
func (m *mockFuelService) UpdateLog(_ context.Context, _ uuid.UUID, tenantID uuid.UUID, _ models.UpdateFuelLogRequest) (*models.FuelLog, error) {
	return &models.FuelLog{ID: uuid.New(), OrganizationID: tenantID}, nil
}
func (m *mockFuelService) DeleteLog(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *mockFuelService) ChangeStatus(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ models.ChangeFuelLogStatusRequest) error {
	return nil
}
func (m *mockFuelService) GetStats(_ context.Context, _ uuid.UUID, _ models.FuelLogFilter) (*models.FuelLogStats, error) {
	return &models.FuelLogStats{}, nil
}
func (m *mockFuelService) GetSettings(_ context.Context, tenantID uuid.UUID) (*models.FuelSettings, error) {
	return &models.FuelSettings{OrganizationID: tenantID, DefaultCurrency: "USD"}, nil
}
func (m *mockFuelService) UpdateSettings(_ context.Context, tenantID uuid.UUID, _ models.UpdateFuelSettingsRequest) (*models.FuelSettings, error) {
	return &models.FuelSettings{OrganizationID: tenantID, DefaultCurrency: "USD"}, nil
}

func setupHandlerRouter(svc service.FuelService) chi.Router {
	r := chi.NewRouter()
	h := NewFuelHandler(svc, zap.NewNop())
	h.RegisterRoutes(r)
	return r
}

func TestListLogsRequiresAuthHeaders(t *testing.T) {
	r := setupHandlerRouter(&mockFuelService{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestContextUUIDExtraction(t *testing.T) {
	m := &mockFuelService{}
	r := setupHandlerRouter(m)
	tenantID := uuid.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), sharedmw.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, sharedmw.UserIDKey, userID)
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, tenantID, m.lastFilter.OrganizationID)
}

func TestCreateLogValidationErrorReturnsBadRequest(t *testing.T) {
	m := &mockFuelService{}
	r := setupHandlerRouter(m)
	tenantID := uuid.New()
	userID := uuid.New()

	body := map[string]interface{}{
		"vehicle_id":       uuid.NewString(),
		"driver_id":        uuid.NewString(),
		"station_name":     "Station",
		"fuel_type":        "diesel",
		"quantity":         0,
		"unit_price":       2,
		"total_cost":       0,
		"odometer":         100,
		"transaction_date": "2026-03-08T00:00:00Z",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	ctx := context.WithValue(req.Context(), sharedmw.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, sharedmw.UserIDKey, userID)
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
}
