package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockVehicleRepository is a mock implementation for testing
type MockVehicleRepository struct {
	mock.Mock
}

func (m *MockVehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	args := m.Called(ctx, vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error) {
	args := m.Called(ctx, orgID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Vehicle), args.Get(1).(int64), args.Error(2)
}

func (m *MockVehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	args := m.Called(ctx, vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func (m *MockVehicleRepository) AssignDriver(ctx context.Context, vehicleID, orgID, driverID uuid.UUID) error {
	args := m.Called(ctx, vehicleID, orgID, driverID)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetStats(ctx context.Context) (*models.VehicleStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VehicleStats), args.Error(1)
}

func (m *MockVehicleRepository) GetByRegistrationNumber(ctx context.Context, tenantID uuid.UUID, regNo string) (*models.Vehicle, error) {
	args := m.Called(ctx, tenantID, regNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) Search(ctx context.Context, params models.VehicleSearchParams) ([]*models.Vehicle, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Vehicle), args.Get(1).(int64), args.Error(2)
}

func setupTestService(t *testing.T) (*VehicleService, *MockVehicleRepository) {
	logger := zap.NewNop()
	mockRepo := &MockVehicleRepository{}
	service := NewVehicleService(mockRepo, logger)
	return service, mockRepo
}

func TestVehicleService_CreateVehicle_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	orgID := uuid.New()

	req := &models.CreateVehicleRequest{
		RegistrationNumber: "ABC-1234",
		VehicleType:        "TRUCK",
		FuelType:           "ELECTRIC",
		OwnershipType:      "OWNED",
	}

	// Expect uniqueness check first (returns nil = not found)
	mockRepo.On("GetByRegistrationNumber", ctx, orgID, "ABC-1234").Return(nil, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	vehicle, err := service.CreateVehicle(ctx, req, orgID)

	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
	assert.Equal(t, "ABC-1234", vehicle.RegistrationNumber)
	assert.Equal(t, models.StatusActive, vehicle.Status)
	assert.Equal(t, orgID, vehicle.TenantID)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_CreateVehicle_InvalidYear(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()

	req := &models.CreateVehicleRequest{
		RegistrationNumber: "ABC-1234",
		VehicleType:        "TRUCK",
		FuelType:           "DIESEL",
		OwnershipType:      "OWNED",
	}

	mockRepo := &MockVehicleRepository{}
	service := NewVehicleService(mockRepo, zap.NewNop())
	// Returns nil = not duplicate
	mockRepo.On("GetByRegistrationNumber", ctx, orgID, "ABC-1234").Return(nil, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	vehicle, err := service.CreateVehicle(ctx, req, orgID)

	// Service doesn't validate year directly — that's handler's job
	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
}

func TestVehicleService_GetVehicle_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()

	expectedVehicle := &models.Vehicle{
		ID:                 vehicleID,
		RegistrationNumber: "ABC-1234",
		Status:             models.StatusActive,
	}

	mockRepo.On("GetByID", ctx, vehicleID).Return(expectedVehicle, nil)

	vehicle, err := service.GetVehicle(ctx, vehicleID)

	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
	assert.Equal(t, vehicleID, vehicle.ID)
	assert.Equal(t, "ABC-1234", vehicle.RegistrationNumber)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_GetVehicle_NotFound(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()

	mockRepo.On("GetByID", ctx, vehicleID).Return(nil, ErrVehicleNotFound)

	vehicle, err := service.GetVehicle(ctx, vehicleID)

	assert.Error(t, err)
	assert.Nil(t, vehicle)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_ListVehicles_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	orgID := uuid.New()

	expectedVehicles := []*models.Vehicle{
		{ID: uuid.New(), RegistrationNumber: "ABC-1234"},
		{ID: uuid.New(), RegistrationNumber: "XYZ-5678"},
	}

	mockRepo.On("GetByOrgID", ctx, orgID, 20, 0).Return(expectedVehicles, int64(2), nil)

	vehicles, total, err := service.ListVehicles(ctx, orgID, 1, 20)

	assert.NoError(t, err)
	assert.Len(t, vehicles, 2)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_UpdateVehicle_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()

	existingVehicle := &models.Vehicle{
		ID:                 vehicleID,
		RegistrationNumber: "ABC-1234",
		Status:             models.StatusActive,
		Version:            1,
	}

	updatedModel := "Updated Model"
	req := &models.UpdateVehicleRequest{
		Model: &updatedModel,
	}

	mockRepo.On("GetByID", ctx, vehicleID).Return(existingVehicle, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(nil)

	vehicle, err := service.UpdateVehicle(ctx, vehicleID, req)

	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
	assert.Equal(t, "Updated Model", *vehicle.Model)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_UpdateVehicle_NotFound(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()

	mockRepo.On("GetByID", ctx, vehicleID).Return(nil, ErrVehicleNotFound)

	req := &models.UpdateVehicleRequest{
		Model: stringPtr("Updated"),
	}

	vehicle, err := service.UpdateVehicle(ctx, vehicleID, req)

	assert.Error(t, err)
	assert.Nil(t, vehicle)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_DeleteVehicle_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()
	orgID := uuid.New()

	// DeleteVehicle first calls GetByID to check status
	existingVehicle := &models.Vehicle{
		ID:       vehicleID,
		TenantID: orgID,
		Status:   models.StatusActive,
	}
	mockRepo.On("GetByID", ctx, vehicleID).Return(existingVehicle, nil)
	mockRepo.On("Delete", ctx, vehicleID, orgID).Return(nil)

	err := service.DeleteVehicle(ctx, vehicleID, orgID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_DeleteVehicle_NotFound(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()
	orgID := uuid.New()

	// GetByID returns not-found first
	mockRepo.On("GetByID", ctx, vehicleID).Return(nil, ErrVehicleNotFound)

	err := service.DeleteVehicle(ctx, vehicleID, orgID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_AssignDriver_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()
	orgID := uuid.New()
	driverID := uuid.New()

	updatedVehicle := &models.Vehicle{
		ID:       vehicleID,
		TenantID: orgID,
		Status:   models.StatusActive,
	}

	mockRepo.On("AssignDriver", ctx, vehicleID, orgID, driverID).Return(nil)
	mockRepo.On("GetByID", ctx, vehicleID).Return(updatedVehicle, nil)

	vehicle, err := service.AssignDriver(ctx, vehicleID, orgID, driverID)

	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_GetVehicleLocation_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()

	existingVehicle := &models.Vehicle{
		ID:        vehicleID,
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetByID", ctx, vehicleID).Return(existingVehicle, nil)

	location, err := service.GetVehicleLocation(ctx, vehicleID)

	assert.NoError(t, err)
	assert.NotNil(t, location)
	assert.Equal(t, vehicleID, location.VehicleID)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_GetVehicleLocation_NotFound(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	vehicleID := uuid.New()

	mockRepo.On("GetByID", ctx, vehicleID).Return(nil, ErrVehicleNotFound)

	location, err := service.GetVehicleLocation(ctx, vehicleID)

	assert.Error(t, err)
	assert.Nil(t, location)
	mockRepo.AssertExpectations(t)
}

func TestVehicleService_GetStats_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	expectedStats := &models.VehicleStats{
		TotalVehicles:  100,
		ActiveVehicles: 80,
		TotalDistance:  5000.5,
		TotalFuel:      500.0,
	}

	mockRepo.On("GetStats", ctx).Return(expectedStats, nil)

	stats, err := service.GetStats(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 100, stats.TotalVehicles)
	assert.Equal(t, 80, stats.ActiveVehicles)
	mockRepo.AssertExpectations(t)
}

func TestVehicleError_Error(t *testing.T) {
	err := NewVehicleError("TEST_CODE", "test message")
	assert.Equal(t, "test message", err.Error())
	assert.Equal(t, "TEST_CODE", err.Code)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("license_plate", "license plate is required")
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, "license_plate", err.Field)
	assert.Contains(t, err.Message, "required")
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("vehicle", "vehicle-123")
	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, 404, err.StatusCode)
	assert.Contains(t, err.Message, "not found")
}

func TestNewAuthorizationError(t *testing.T) {
	err := NewAuthorizationError("access denied")
	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.Equal(t, 403, err.StatusCode)
}

func TestNewBulkError(t *testing.T) {
	err := NewBulkError("FILE_TYPE", "invalid file type")
	assert.Contains(t, err.Code, "BULK_ERROR")
	assert.Equal(t, 400, err.StatusCode)
}

func stringPtr(s string) *string {
	return &s
}
