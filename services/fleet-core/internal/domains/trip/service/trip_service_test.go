//go:build integration
// +build integration

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/trip/models"
	"github.com/flyfleet/fleet-core/internal/domains/trip/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTripRepository is a mock implementation for testing
type MockTripRepository struct {
	mock.Mock
}

func (m *MockTripRepository) Create(ctx context.Context, trip *models.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockTripRepository) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripRepository) List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]*models.Trip), args.Int(1), args.Error(2)
}

func (m *MockTripRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockTripRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTripRepository) AddEvent(ctx context.Context, tripID string, event models.TripEvent) error {
	args := m.Called(ctx, tripID, event)
	return args.Error(0)
}

func (m *MockTripRepository) AddPODDocument(ctx context.Context, tripID string, pod models.PODDocument) error {
	args := m.Called(ctx, tripID, pod)
	return args.Error(0)
}

func (m *MockTripRepository) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetTripSummary(ctx context.Context, tenantID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripSummary), args.Error(1)
}

func (m *MockTripRepository) GetStats(ctx context.Context) (*models.TripStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripStats), args.Error(1)
}

func (m *MockTripRepository) GetDriverStats(ctx context.Context, tenantID string) ([]models.DriverPerformance, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DriverPerformance), args.Error(1)
}

func (m *MockTripRepository) GetTripStats(ctx context.Context, tenantID string) ([]models.DailyTripTrend, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DailyTripTrend), args.Error(1)
}

func (m *MockTripRepository) Approve(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTripRepository) GetLastEvent(ctx context.Context, tripID string) (*models.TripEvent, error) {
	args := m.Called(ctx, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripEvent), args.Error(1)
}

func (m *MockTripRepository) GetLastCompletedTripForVehicle(ctx context.Context, vehicleID string) (*models.Trip, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetActiveTripsForDriver(ctx context.Context, driverID string) ([]*models.Trip, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) CheckOverlappingTrips(ctx context.Context, vehicleID, driverID string, start, end time.Time) (bool, error) {
	args := m.Called(ctx, vehicleID, driverID, start, end)
	return args.Bool(0), args.Error(1)
}

func (m *MockTripRepository) AddTelemetryPoint(ctx context.Context, point models.TelemetryPoint) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockTripRepository) GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error) {
	args := m.Called(ctx, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TelemetryPoint), args.Error(1)
}

func (m *MockTripRepository) CreateOffer(ctx context.Context, offer *models.TripOffer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

func (m *MockTripRepository) GetOffer(ctx context.Context, offerID string) (*models.TripOffer, error) {
	args := m.Called(ctx, offerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TripOffer), args.Error(1)
}

func (m *MockTripRepository) RespondToOffer(ctx context.Context, offerID string, status, reason string) error {
	args := m.Called(ctx, offerID, status, reason)
	return args.Error(0)
}

func (m *MockTripRepository) GetOffersForDriver(ctx context.Context, driverID string) ([]*models.TripOffer, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TripOffer), args.Error(1)
}

func (m *MockTripRepository) SaveProofOfDelivery(ctx context.Context, pod *models.ProofOfDelivery) error {
	args := m.Called(ctx, pod)
	return args.Error(0)
}

func (m *MockTripRepository) GetPODsForTrip(ctx context.Context, tripID string) ([]*models.ProofOfDelivery, error) {
	args := m.Called(ctx, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProofOfDelivery), args.Error(1)
}

func (m *MockTripRepository) WithTx(ctx context.Context, fn func(repository.TripRepository) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func setupTestService(t *testing.T) (*TripService, *MockTripRepository) {
	logger := zap.NewNop()
	mockRepo := &MockTripRepository{}
	service := NewTripService(mockRepo, logger)
	return service, mockRepo
}

func TestTripService_Create_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tenantID := "tenant-123"

	req := &models.CreateTripRequest{
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

	mockRepo.On("Create", ctx, mock.Anything).Return(nil)
	mockRepo.On("CheckOverlappingTrips", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	trip, err := service.Create(ctx, tenantID, req)

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, models.TripStatusCreated, trip.Status)
	assert.Equal(t, tenantID, trip.TenantID)
	assert.Equal(t, req.Title, trip.Title)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Get_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"

	expectedTrip := &models.Trip{
		ID:     tripID,
		Status: models.TripStatusScheduled,
		Title:  "Test Trip",
	}

	mockRepo.On("GetByID", ctx, tripID).Return(expectedTrip, nil)

	trip, err := service.GetTripByID(ctx, tripID)

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, tripID, trip.ID)
	assert.Equal(t, models.TripStatusScheduled, trip.Status)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Get_NotFound(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-notfound"

	mockRepo.On("GetByID", ctx, tripID).Return(nil, fmt.Errorf("trip not found"))

	trip, err := service.GetTripByID(ctx, tripID)

	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestTripService_Start_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"

	existingTrip := &models.Trip{
		ID:             tripID,
		DriverID:       &driverID,
		Status:         models.TripStatusScheduled,
		PermitVerified: true, // Required for starting trip
		PlannedRoute: []models.Waypoint{
			{Location: models.Location{Latitude: 37.7749, Longitude: -122.4194}},
		},
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("GetActiveTripsForVehicle", ctx, existingTrip.VehicleID).Return([]*models.Trip{}, nil)
	mockRepo.On("GetActiveTripsForDriver", ctx, driverID).Return([]*models.Trip{}, nil)
	mockRepo.On("GetLastCompletedTripForVehicle", ctx, existingTrip.VehicleID).Return((*models.Trip)(nil), fmt.Errorf("no trips found"))
	mockRepo.On("Update", ctx, tripID, mock.Anything).Return(nil)
	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	req := &models.StartTripRequest{
		OdometerReading: 10000,
		FuelLevel:       80.0,
	}

	trip, err := service.Start(ctx, tripID, driverID, req)

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Start_Unauthorized(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"
	wrongDriverID := "driver-456"

	existingTrip := &models.Trip{
		ID:       tripID,
		DriverID: &driverID,
		Status:   "planned",
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	req := &models.StartTripRequest{
		OdometerReading: 10000,
		FuelLevel:       80.0,
	}

	trip, err := service.Start(ctx, tripID, wrongDriverID, req)

	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "unauthorized")
	mockRepo.AssertExpectations(t)
}

func TestTripService_Start_InvalidStatus(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"

	existingTrip := &models.Trip{
		ID:       tripID,
		DriverID: &driverID,
		Status:   "completed",
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	req := &models.StartTripRequest{
		OdometerReading: 10000,
		FuelLevel:       80.0,
	}

	trip, err := service.Start(ctx, tripID, driverID, req)

	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "cannot be started")
	mockRepo.AssertExpectations(t)
}

func TestTripService_Pause_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"

	existingTrip := &models.Trip{
		ID:       tripID,
		DriverID: &driverID,
		Status:   models.TripStatusStarted,
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("Update", ctx, tripID, mock.Anything).Return(nil)
	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	trip, err := service.Pause(ctx, tripID, driverID)

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Resume_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"

	existingTrip := &models.Trip{
		ID:       tripID,
		DriverID: &driverID,
		Status:   models.TripStatusPaused,
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("Update", ctx, tripID, mock.Anything).Return(nil)
	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	trip, err := service.Resume(ctx, tripID, driverID)

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Complete_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"
	startedAt := time.Now().Add(-2 * time.Hour)

	existingTrip := &models.Trip{
		ID:        tripID,
		DriverID:  &driverID,
		Status:    models.TripStatusStarted,
		StartTime: &startedAt,
		PlannedRoute: []models.Waypoint{
			{},
			{Location: models.Location{Latitude: 37.7749, Longitude: -122.4194}},
		},
		PermitVerified: true,
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("Update", ctx, tripID, mock.Anything).Return(nil)
	mockRepo.On("AddPODDocument", ctx, tripID, mock.Anything).Return(nil)
	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	req := &models.CompleteTripRequest{
		OdometerReading:   10100,
		FuelLevel:         50.0,
		TotalDistance:     105.5,
		FuelUsed:          12.5,
		IdlingTimeMinutes: 30,
		PODDocuments:      []models.PODDocument{{Type: "signature", URL: "http://example.com/pod.png"}},
	}

	trip, err := service.Complete(ctx, tripID, driverID, req)

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Cancel_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"

	existingTrip := &models.Trip{
		ID:     tripID,
		Status: models.TripStatusScheduled,
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("Update", ctx, tripID, mock.Anything).Return(nil)

	err := service.Cancel(ctx, tripID, "Customer requested cancellation")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Cancel_AlreadyCompleted(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"

	existingTrip := &models.Trip{
		ID:     tripID,
		Status: models.TripStatusCompleted,
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)

	err := service.Cancel(ctx, tripID, "Test reason")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel")
	mockRepo.AssertExpectations(t)
}

func TestTripService_AddEvent_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"

	existingTrip := &models.Trip{
		ID:       tripID,
		DriverID: &driverID,
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("GetLastEvent", ctx, tripID).Return((*models.TripEvent)(nil), assert.AnError)
	mockRepo.On("AddEvent", ctx, tripID, mock.MatchedBy(func(e models.TripEvent) bool {
		return e.Type == "arrival" && e.CreatedBy == driverID
	})).Return(nil)

	req := &models.AddEventRequest{
		Type:        "arrival",
		Description: "Arrived at waypoint",
		Location:    models.Location{Latitude: 37.7749, Longitude: -122.4194},
	}

	err := service.AddEvent(ctx, tripID, driverID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTripService_OfferTrip_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tripID := "trip-123"
	driverID := "driver-123"

	existingTrip := &models.Trip{
		ID:     tripID,
		Status: "planned", // Must be "planned" or "offered" for OfferTrip to succeed
	}

	mockRepo.On("GetByID", ctx, tripID).Return(existingTrip, nil)
	mockRepo.On("CreateOffer", ctx, mock.Anything).Return(nil)

	offer, err := service.OfferTrip(ctx, tripID, driverID)

	assert.NoError(t, err)
	assert.NotNil(t, offer)
	assert.Equal(t, "OFFERED", offer.Status)
	mockRepo.AssertExpectations(t)
}

func TestTripService_RespondToOffer_Accepted(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	offerID := "offer-123"
	driverID := "driver-123"

	existingOffer := &models.TripOffer{
		ID:        offerID,
		TripID:    "trip-123",
		DriverID:  driverID,
		Status:    "OFFERED",
		ExpiresAt: nil,
	}

	mockRepo.On("GetOffer", ctx, offerID).Return(existingOffer, nil)
	mockRepo.On("RespondToOffer", ctx, offerID, "ACCEPTED", "").Return(nil)
	mockRepo.On("Update", ctx, "trip-123", mock.Anything).Return(nil)

	err := service.RespondToOffer(ctx, offerID, driverID, "ACCEPTED", "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTripService_RespondToOffer_Expired(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	offerID := "offer-123"
	driverID := "driver-123"

	expiredTime := time.Now().Add(-1 * time.Hour)
	existingOffer := &models.TripOffer{
		ID:        offerID,
		TripID:    "trip-123",
		DriverID:  driverID,
		Status:    "OFFERED",
		ExpiresAt: &expiredTime,
	}

	mockRepo.On("GetOffer", ctx, offerID).Return(existingOffer, nil)
	mockRepo.On("RespondToOffer", ctx, offerID, "EXPIRED", "Auto-expired").Return(nil)

	err := service.RespondToOffer(ctx, offerID, driverID, "ACCEPTED", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
	mockRepo.AssertExpectations(t)
}

func TestTripService_GetDriverOffers_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	driverID := "driver-123"

	expectedOffers := []*models.TripOffer{
		{ID: "offer-1", DriverID: driverID, Status: "OFFERED"},
		{ID: "offer-2", DriverID: driverID, Status: "OFFERED"},
	}

	mockRepo.On("GetOffersForDriver", ctx, driverID).Return(expectedOffers, nil)

	offers, err := service.GetDriverOffers(ctx, driverID)

	assert.NoError(t, err)
	assert.Len(t, offers, 2)
	mockRepo.AssertExpectations(t)
}

func TestTripService_GetTripSummary_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tenantID := "tenant-123"
	startDate := time.Now().Add(-30 * 24 * time.Hour)
	endDate := time.Now()

	expectedSummary := &models.TripSummary{
		TotalTrips:     100,
		CompletedTrips: 90,
	}

	mockRepo.On("GetTripSummary", ctx, tenantID, startDate, endDate).Return(expectedSummary, nil)

	summary, err := service.GetTripSummary(ctx, tenantID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 100, summary.TotalTrips)
	mockRepo.AssertExpectations(t)
}

func TestTripService_GetStats_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	expectedStats := &models.TripStats{
		ActiveTrips: 25,
	}

	mockRepo.On("GetStats", ctx).Return(expectedStats, nil)

	stats, err := service.GetStats(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 25, stats.ActiveTrips)
	mockRepo.AssertExpectations(t)
}

func TestTripError_Error(t *testing.T) {
	err := NewTripError("TEST_CODE", "test message")
	assert.Equal(t, "test message", err.Error())
	assert.Equal(t, "TEST_CODE", err.Code)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("title", "title is required")
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, "title", err.Field)
	assert.Contains(t, err.Message, "required")
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("trip", "trip-123")
	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, 404, err.StatusCode)
	assert.Contains(t, err.Message, "not found")
}

func TestNewAuthorizationError(t *testing.T) {
	err := NewAuthorizationError("access denied")
	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.Equal(t, 403, err.StatusCode)
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("trip already completed")
	assert.Equal(t, "CONFLICT", err.Code)
	assert.Equal(t, 409, err.StatusCode)
}
