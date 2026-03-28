//go:build integration
// +build integration

package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAnalyticsService is a mock implementation of service.AnalyticsService
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stubs for other methods to satisfy the interface
func (m *MockAnalyticsService) CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error {
	return nil
}
func (m *MockAnalyticsService) GetPredictiveModel(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveModel, error) {
	return nil, nil
}
func (m *MockAnalyticsService) GetPredictiveModelByVersion(ctx context.Context, name, version string, orgID uuid.UUID) (*models.PredictiveModel, error) {
	return nil, nil
}
func (m *MockAnalyticsService) UpdatePredictiveModel(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *MockAnalyticsService) DeletePredictiveModel(ctx context.Context, id string, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) ListPredictiveModels(ctx context.Context, orgID uuid.UUID, modelType *string, status *string, cursor string, limit, offset int) ([]*models.PredictiveModel, int, error) {
	return nil, 0, nil
}

func (m *MockAnalyticsService) CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}
func (m *MockAnalyticsService) GetPredictionResult(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictionResult, error) {
	return nil, nil
}
func (m *MockAnalyticsService) GetPredictionResultsByModel(ctx context.Context, modelID string, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	return nil, 0, nil
}
func (m *MockAnalyticsService) GetPredictionResultsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	return nil, 0, nil
}
func (m *MockAnalyticsService) UpdatePredictionResult(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *MockAnalyticsService) DeletePredictionResult(ctx context.Context, id string, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, cursor string, limit, offset int) ([]*models.PredictionResult, int, error) {
	return nil, 0, nil
}

func (m *MockAnalyticsService) CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error {
	return nil
}
func (m *MockAnalyticsService) GetAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) (*models.AnalyticsJob, error) {
	return nil, nil
}
func (m *MockAnalyticsService) UpdateAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *MockAnalyticsService) DeleteAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, cursor string, limit, offset int) ([]*models.AnalyticsJob, int, error) {
	return nil, 0, nil
}

func (m *MockAnalyticsService) CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error {
	return nil
}
func (m *MockAnalyticsService) GetFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.FleetAnalytics, error) {
	return nil, nil
}
func (m *MockAnalyticsService) GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
	return nil, nil
}
func (m *MockAnalyticsService) UpdateFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *MockAnalyticsService) DeleteFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, cursor string, limit, offset int) ([]*models.FleetAnalytics, int, error) {
	return nil, 0, nil
}

func (m *MockAnalyticsService) CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error {
	return nil
}
func (m *MockAnalyticsService) GetDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.DriverBehaviorAnalytics, error) {
	return nil, nil
}
func (m *MockAnalyticsService) GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	return nil, 0, nil
}
func (m *MockAnalyticsService) GetDriverBehaviorAnalyticsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	return nil, 0, nil
}
func (m *MockAnalyticsService) UpdateDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *MockAnalyticsService) DeleteDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, cursor string, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	return nil, 0, nil
}

func (m *MockAnalyticsService) CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error {
	return nil
}
func (m *MockAnalyticsService) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveMaintenanceAnalytics, error) {
	return nil, nil
}
func (m *MockAnalyticsService) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	return nil, 0, nil
}
func (m *MockAnalyticsService) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *MockAnalyticsService) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, cursor string, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	return nil, 0, nil
}

func (m *MockAnalyticsService) GetVehicleDailyStats(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.VehicleDailyStats, error) {
	args := m.Called(ctx, vehicleID, orgID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.VehicleDailyStats), args.Error(1)
}
func (m *MockAnalyticsService) GetDriverDailyStats(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.DriverDailyStats, error) {
	return nil, nil
}
func (m *MockAnalyticsService) GetFleetHourlyStats(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.FleetHourlyStats, error) {
	return nil, nil
}

func (m *MockAnalyticsService) GetVehicleAnalytics(w http.ResponseWriter, r *http.Request)  {}
func (m *MockAnalyticsService) GetDriverAnalytics(w http.ResponseWriter, r *http.Request)   {}
func (m *MockAnalyticsService) GetTripAnalytics(w http.ResponseWriter, r *http.Request)     {}
func (m *MockAnalyticsService) GetForecastAnalytics(w http.ResponseWriter, r *http.Request) {}

func (m *MockAnalyticsService) InvalidateByPattern(ctx context.Context, pattern string) error {
	return nil
}
func (m *MockAnalyticsService) InvalidateDriverCache(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) InvalidateVehicleCache(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) InvalidateFleetCache(ctx context.Context, orgID uuid.UUID) error {
	return nil
}
func (m *MockAnalyticsService) SetRaw(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return nil
}
func (m *MockAnalyticsService) IncrementFleetVersion(ctx context.Context, orgID uuid.UUID) error {
	return nil
}
func TestHealthCheck(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService, nil, logger)

	mockService.On("HealthCheck", mock.Anything).Return(nil)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/health", handler.HealthCheck)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"healthy"}`, w.Body.String())
	mockService.AssertExpectations(t)
}
