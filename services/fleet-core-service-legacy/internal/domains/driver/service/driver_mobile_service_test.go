package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockDriverRepository is a mock implementation of Repository
type MockDriverRepository struct {
	mock.Mock
}

func (m *MockDriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	args := m.Called(ctx, driver)
	return args.Error(0)
}
func (m *MockDriverRepository) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Driver), args.Error(1)
}
func (m *MockDriverRepository) GetByEmail(ctx context.Context, email string) (*models.Driver, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Driver), args.Error(1)
}
func (m *MockDriverRepository) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error) {
	args := m.Called(ctx, licenseNumber)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Driver), args.Error(1)
}
func (m *MockDriverRepository) List(ctx context.Context, filter models.DriverListFilter) ([]*models.Driver, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Driver), args.Int(1), args.Error(2)
}
func (m *MockDriverRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}
func (m *MockDriverRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockDriverRepository) AssignVehicle(ctx context.Context, driverID, vehicleID string) error {
	args := m.Called(ctx, driverID, vehicleID)
	return args.Error(0)
}
func (m *MockDriverRepository) UnassignVehicle(ctx context.Context, driverID string) error {
	args := m.Called(ctx, driverID)
	return args.Error(0)
}
func (m *MockDriverRepository) UpdateBehaviorScore(ctx context.Context, driverID string, scoreChange int) error {
	args := m.Called(ctx, driverID, scoreChange)
	return args.Error(0)
}
func (m *MockDriverRepository) IncrementStats(ctx context.Context, driverID string, distanceKm, hoursDriven float64) error {
	args := m.Called(ctx, driverID, distanceKm, hoursDriven)
	return args.Error(0)
}
func (m *MockDriverRepository) GetDriverAnalytics(ctx context.Context, driverID string) (*models.DriverAnalytics, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.DriverAnalytics), args.Error(1)
}
func (m *MockDriverRepository) GetActiveDriversForVehicle(ctx context.Context, vehicleID string) ([]*models.Driver, error) {
	args := m.Called(ctx, vehicleID)
	return args.Get(0).([]*models.Driver), args.Error(1)
}
func (m *MockDriverRepository) CreateDutyStatusLog(ctx context.Context, log *models.DutyStatusLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}
func (m *MockDriverRepository) GetLatestDutyStatus(ctx context.Context, driverID string) (*models.DutyStatusLog, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.DutyStatusLog), args.Error(1)
}
func (m *MockDriverRepository) CreateMessage(ctx context.Context, msg *models.DriverMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}
func (m *MockDriverRepository) GetMessages(ctx context.Context, driverID string, limit, offset int) ([]*models.DriverMessage, error) {
	args := m.Called(ctx, driverID, limit, offset)
	return args.Get(0).([]*models.DriverMessage), args.Error(1)
}
func (m *MockDriverRepository) CreateExpense(ctx context.Context, expense *models.DriverExpense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}
func (m *MockDriverRepository) ListExpenses(ctx context.Context, driverID string) ([]*models.DriverExpense, error) {
	args := m.Called(ctx, driverID)
	return args.Get(0).([]*models.DriverExpense), args.Error(1)
}

func TestUpdateDutyStatus(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewDriverService(mockRepo, logger)

	ctx := context.Background()
	driverID := "driver-123"
	req := &models.DutyStatusLog{Status: "ON_DUTY"}

	mockRepo.On("GetLatestDutyStatus", ctx, driverID).Return(nil, nil)
	mockRepo.On("CreateDutyStatusLog", ctx, mock.AnythingOfType("*models.DutyStatusLog")).Return(nil)

	err := service.UpdateDutyStatus(ctx, driverID, req)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSendMessage(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewDriverService(mockRepo, logger)

	ctx := context.Background()
	driverID := "driver-123"
	
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.DriverMessage")).Return(nil)

	msg, err := service.SendMessage(ctx, driverID, "Hello Dispatch", "CHAT", "")
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "Hello Dispatch", msg.Content)
	assert.Equal(t, "INBOUND", msg.Direction)
	mockRepo.AssertExpectations(t)
}

func TestSubmitExpense(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewDriverService(mockRepo, logger)

	ctx := context.Background()
	driverID := "driver-123"
	req := &models.DriverExpense{Amount: 50.0, Currency: "USD"}

	mockRepo.On("CreateExpense", ctx, req).Return(nil)

	res, err := service.SubmitExpense(ctx, driverID, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "PENDING", res.Status)
	mockRepo.AssertExpectations(t)
}
