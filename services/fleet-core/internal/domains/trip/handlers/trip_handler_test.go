package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/trip/models"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTripService is a mock implementation for testing
type MockTripService struct {
	mock.Mock
}

func (m *MockTripService) Create(ctx context.Context, tenantID string, req *models.CreateTripRequest) (*models.Trip, error) {
	args := m.Called(ctx, tenantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) GetTripByID(ctx context.Context, id string) (*models.Trip, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Trip), args.Int(1), args.Error(2)
}

func (m *MockTripService) Update(ctx context.Context, id string, req *models.UpdateTripRequest) (*models.Trip, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTripService) Start(ctx context.Context, tripID, driverID string, req *models.StartTripRequest) (*models.Trip, error) {
	args := m.Called(ctx, tripID, driverID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) Pause(ctx context.Context, tripID, driverID string) (*models.Trip, error) {
	args := m.Called(ctx, tripID, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) Resume(ctx context.Context, tripID, driverID string) (*models.Trip, error) {
	args := m.Called(ctx, tripID, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) Complete(ctx context.Context, tripID, driverID string, req *models.CompleteTripRequest) (*models.Trip, error) {
	args := m.Called(ctx, tripID, driverID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripService) Cancel(ctx context.Context, tripID, reason string) error {
	args := m.Called(ctx, tripID, reason)
	return args.Error(0)
}

func (m *MockTripService) Fail(ctx context.Context, id, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockTripService) Approve(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTripService) AddEvent(ctx context.Context, tripID, driverID string, req *models.AddEventRequest) error {
	args := m.Called(ctx, tripID, driverID, req)
	return args.Error(0)
}

func (m *MockTripService) SubmitPOD(ctx context.Context, tripID, driverID string, req *models.SubmitPODRequest) error {
	args := m.Called(ctx, tripID, driverID, req)
	return args.Error(0)
}

func (m *MockTripService) OfferTrip(ctx context.Context, tripID, driverID string) (*models.TripOffer, error) {
	args := m.Called(ctx, tripID, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripOffer), args.Error(1)
}

func (m *MockTripService) RespondToOffer(ctx context.Context, offerID, driverID, response, reason string) error {
	args := m.Called(ctx, offerID, driverID, response, reason)
	return args.Error(0)
}

func (m *MockTripService) GetDriverOffers(ctx context.Context, driverID string) ([]*models.TripOffer, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TripOffer), args.Error(1)
}

func (m *MockTripService) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripService) GetTripSummary(ctx context.Context, tenantID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripSummary), args.Error(1)
}

func (m *MockTripService) GetDriverStats(ctx context.Context, tenantID string) ([]models.DriverPerformance, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DriverPerformance), args.Error(1)
}

func (m *MockTripService) GetTripStats(ctx context.Context, tenantID string) ([]models.DailyTripTrend, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DailyTripTrend), args.Error(1)
}

func (m *MockTripService) GetStats(ctx context.Context) (*models.TripStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripStats), args.Error(1)
}

func (m *MockTripService) GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error) {
	args := m.Called(ctx, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TelemetryPoint), args.Error(1)
}

func (m *MockTripService) OptimizeRoute(ctx context.Context, tripID string) (*models.RouteOptimization, error) {
	args := m.Called(ctx, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RouteOptimization), args.Error(1)
}

func (m *MockTripService) GetTripCustomers(ctx context.Context, tenantID string) ([]string, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func setupHandlerTest(t *testing.T) (*TripHandler, *MockTripService) {
	logger := zap.NewNop()
	mockService := &MockTripService{}
	handler := NewTripHandler(mockService, logger)
	return handler, mockService
}

func TestTripHandler_CreateTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.CreateTripRequest{
		VehicleID:            "vehicle-123",
		DriverID:             "driver-123",
		Title:                "Test Trip",
		Description:          "Test Description",
		PlannedRoute:         []models.Waypoint{{Name: "Start"}, {Name: "End"}},
		Waypoints:            []models.Waypoint{{Name: "Start"}, {Name: "End"}},
		PlannedDistance:      100.5,
		EstimatedDurationMin: 120,
		CargoDetails:         models.CargoDetails{Type: "general", WeightKg: 500},
		ScheduledStart:       time.Now().Add(24 * time.Hour),
		CustomerName:         "Test Customer",
		CustomerPhone:        "+1234567890",
		CustomerEmail:        "test@example.com",
	}

	expectedTrip := &models.Trip{
		ID:     "trip-123",
		Status: "planned",
		Title:  "Test Trip",
	}

	mockService.On("Create", mock.Anything, "tenant-123", mock.Anything).Return(expectedTrip, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.TenantIDKey, "tenant-123"))
	rr := httptest.NewRecorder()

	handler.CreateTrip(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "trip-123")
	mockService.AssertExpectations(t)
}

func TestTripHandler_CreateTrip_InvalidJSON(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader([]byte("invalid-json")))
	// Inject tenant context so handler can proceed past auth check to body decode
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.TenantIDKey, "tenant-123"))
	rr := httptest.NewRecorder()

	handler.CreateTrip(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request body")
}

func TestTripHandler_CreateTrip_MissingOrgContext(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.CreateTripRequest{
		VehicleID: "vehicle-123",
		DriverID:  "driver-123",
		Title:     "Test Trip",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader(body))
	// Intentionally don't set tenant_id in context to verify auth enforcement
	rr := httptest.NewRecorder()

	handler.CreateTrip(rr, req)

	// Handler correctly returns 401 when tenant context is missing (security-first design)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Tenant context missing")
	mockService.AssertNotCalled(t, "Create")
}

func TestTripHandler_GetTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	expectedTrip := &models.Trip{
		ID:     "trip-123",
		Status: "planned",
		Title:  "Test Trip",
	}

	mockService.On("GetTripByID", mock.Anything, "trip-123").Return(expectedTrip, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips/trip-123", nil)
	rr := httptest.NewRecorder()

	// Setup chi URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetTrip(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "trip-123")
	mockService.AssertExpectations(t)
}

func TestTripHandler_GetTrip_NotFound(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	mockService.On("GetTripByID", mock.Anything, "trip-notfound").Return(nil, fmt.Errorf("trip not found"))

	req := httptest.NewRequest(http.MethodGet, "/trips/trip-notfound", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-notfound")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetTrip(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestTripHandler_ListTrips_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	expectedTrips := []*models.Trip{
		{ID: "trip-1", Status: models.TripStatusScheduled},
		{ID: "trip-2", Status: models.TripStatusStarted},
	}

	mockService.On("List", mock.Anything, mock.Anything).Return(expectedTrips, 2, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips?status=SCHEDULED&page=1&page_size=20", nil)
	// Inject tenant context required by handler
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.TenantIDKey, "tenant-123"))
	rr := httptest.NewRecorder()

	handler.ListTrips(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "trip-1")
	assert.Contains(t, rr.Body.String(), `"total":2`)
	mockService.AssertExpectations(t)
}

func TestTripHandler_StartTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.StartTripRequest{
		OdometerReading: 10000,
		FuelLevel:       80.0,
	}

	expectedTrip := &models.Trip{
		ID:     "trip-123",
		Status: "in_progress",
	}

	mockService.On("Start", mock.Anything, "trip-123", "driver-123", mock.Anything).Return(expectedTrip, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/start", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserIDKey, "driver-123"))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.StartTrip(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "in_progress")
	mockService.AssertExpectations(t)
}

func TestTripHandler_StartTrip_MissingDriverContext(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	reqBody := &models.StartTripRequest{
		OdometerReading: 10000,
		FuelLevel:       80.0,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/start", bytes.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.StartTrip(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Driver context missing")
}

func TestTripHandler_CompleteTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.CompleteTripRequest{
		OdometerReading:   10100,
		FuelLevel:         50.0,
		TotalDistance:     105.5,
		FuelUsed:          12.5,
		IdlingTimeMinutes: 30,
		PODDocuments:      []models.PODDocument{{Type: "signature", URL: "http://example.com/pod.png"}},
	}

	expectedTrip := &models.Trip{
		ID:     "trip-123",
		Status: "completed",
	}

	mockService.On("Complete", mock.Anything, "trip-123", "driver-123", mock.Anything).Return(expectedTrip, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/complete", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserIDKey, "driver-123"))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CompleteTrip(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "completed")
	mockService.AssertExpectations(t)
}

func TestTripHandler_CancelTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	mockService.On("Cancel", mock.Anything, "trip-123", "Customer requested").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/cancel?reason=Customer+requested", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.CancelTrip(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockService.AssertExpectations(t)
}

func TestTripHandler_CancelTrip_MissingReason(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/cancel", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.CancelTrip(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "reason is required")
}

func TestTripHandler_AddEvent_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.AddEventRequest{
		Type:        "arrival",
		Description: "Arrived at waypoint",
		Location:    models.Location{Latitude: 37.7749, Longitude: -122.4194},
	}

	mockService.On("AddEvent", mock.Anything, "trip-123", "driver-123", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/events", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserIDKey, "driver-123"))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.AddEvent(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockService.AssertExpectations(t)
}

func TestTripHandler_SubmitPOD_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.SubmitPODRequest{
		Type:        "signature",
		URL:         "http://example.com/pod.png",
		Recipient:   models.Recipient{Name: "Test Recipient", Phone: "+1234567890"},
		OTPVerified: true,
	}

	mockService.On("SubmitPOD", mock.Anything, "trip-123", "driver-123", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/trips/trip-123/pod", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserIDKey, "driver-123"))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.SubmitPOD(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockService.AssertExpectations(t)
}

func TestTripHandler_OptimizeRoute_Success(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/trips/trip-123/optimize-route", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.OptimizeRoute(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "routes")
	assert.Contains(t, rr.Body.String(), "polyline")
}

func TestTripHandler_GetDriverStats_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	expectedStats := []models.DriverPerformance{
		{DriverID: "driver-1", TotalDistance: 1000, TripsCompleted: 50},
	}

	mockService.On("GetDriverStats", mock.Anything, "tenant-123").Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips/stats/drivers", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.TenantIDKey, "tenant-123"))
	rr := httptest.NewRecorder()

	handler.GetDriverStats(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "driver-1")
	mockService.AssertExpectations(t)
}

func TestTripHandler_GetTripStats_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	expectedStats := []models.DailyTripTrend{
		{Date: "2024-01-15", TotalTrips: 10, TotalDistance: 500},
	}

	mockService.On("GetTripStats", mock.Anything, "tenant-123").Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips/stats/trips", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.TenantIDKey, "tenant-123"))
	rr := httptest.NewRecorder()

	handler.GetTripStats(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "2024-01-15")
	mockService.AssertExpectations(t)
}

func TestTripHandler_GetStats_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	expectedStats := &models.TripStats{
		ActiveTrips: 25,
	}

	mockService.On("GetStats", mock.Anything).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips/stats", nil)
	// Set required role for admin-only endpoint
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.RoleKey, "admin"))
	rr := httptest.NewRecorder()

	handler.GetStats(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "25")
	mockService.AssertExpectations(t)
}

func TestTripHandler_UpdateTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.UpdateTripRequest{
		Title:       stringPtr("Updated Title"),
		Description: stringPtr("Updated Description"),
	}

	expectedTrip := &models.Trip{
		ID:          "trip-123",
		Title:       "Updated Title",
		Description: "Updated Description",
	}

	mockService.On("Update", mock.Anything, "trip-123", mock.Anything).Return(expectedTrip, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/trips/trip-123", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.UpdateTrip(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Updated Title")
	mockService.AssertExpectations(t)
}

func TestTripHandler_DeleteTrip_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	mockService.On("Delete", mock.Anything, "trip-123").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/trips/trip-123", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "trip-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.DeleteTrip(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockService.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

func TestTripHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockTripService{}
	handler := NewTripHandler(mockService, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.service)
}
