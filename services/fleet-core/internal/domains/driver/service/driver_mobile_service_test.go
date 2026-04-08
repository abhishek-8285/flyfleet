package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Driver), args.Error(1)
}
func (m *MockDriverRepository) GetByEmail(ctx context.Context, email string) (*models.Driver, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Driver), args.Error(1)
}
func (m *MockDriverRepository) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error) {
	args := m.Called(ctx, licenseNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

func (m *MockDriverRepository) LogBehaviorScoreChange(ctx context.Context, driverID, orgID uuid.UUID, oldScore, newScore int, reason, eventType string, userID uuid.UUID, userEmail string) error {
	args := m.Called(ctx, driverID, orgID, oldScore, newScore, reason, eventType, userID, userEmail)
	return args.Error(0)
}

func (m *MockDriverRepository) CreateBehaviorScoreUpdate(ctx context.Context, update *models.BehaviorScoreUpdate) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockDriverRepository) GetBehaviorRecords(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DriverBehavior, error) {
	args := m.Called(ctx, driverID, startDate, endDate)
	return args.Get(0).([]*models.DriverBehavior), args.Error(1)
}

func (m *MockDriverRepository) GetBehaviorSummary(ctx context.Context, driverID string) (*models.DriverBehavior, error) {
	args := m.Called(ctx, driverID)
	return args.Get(0).(*models.DriverBehavior), args.Error(1)
}

func (m *MockDriverRepository) GetIncidents(ctx context.Context, driverID string, filter models.BehaviorReportRequest) ([]*models.DriverIncident, error) {
	args := m.Called(ctx, driverID, filter)
	return args.Get(0).([]*models.DriverIncident), args.Error(1)
}

func (m *MockDriverRepository) EndDutyStatusLog(ctx context.Context, logID string, durationMinutes int) error {
	args := m.Called(ctx, logID, durationMinutes)
	return args.Error(0)
}

func (m *MockDriverRepository) GetDutyStatusLogs(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DutyStatusLog, error) {
	args := m.Called(ctx, driverID, startDate, endDate)
	return args.Get(0).([]*models.DutyStatusLog), args.Error(1)
}

// Attendance & Schedule Methods
func (m *MockDriverRepository) CreateAttendance(ctx context.Context, attendance *models.DriverAttendance) error {
	args := m.Called(ctx, attendance)
	return args.Error(0)
}

func (m *MockDriverRepository) GetAttendanceByID(ctx context.Context, id uuid.UUID) (*models.DriverAttendance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DriverAttendance), args.Error(1)
}

func (m *MockDriverRepository) GetAttendanceByDriver(ctx context.Context, driverID, orgID uuid.UUID, from, to time.Time) ([]*models.DriverAttendance, error) {
	args := m.Called(ctx, driverID, orgID, from, to)
	return args.Get(0).([]*models.DriverAttendance), args.Error(1)
}

func (m *MockDriverRepository) UpdateCheckOut(ctx context.Context, id, orgID uuid.UUID, checkOut time.Time, remarks string) error {
	args := m.Called(ctx, id, orgID, checkOut, remarks)
	return args.Error(0)
}

func (m *MockDriverRepository) CreateSchedule(ctx context.Context, schedule *models.DriverSchedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockDriverRepository) GetScheduleByID(ctx context.Context, id uuid.UUID) (*models.DriverSchedule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DriverSchedule), args.Error(1)
}

func (m *MockDriverRepository) GetSchedulesByDriver(ctx context.Context, driverID, orgID uuid.UUID) ([]*models.DriverSchedule, error) {
	args := m.Called(ctx, driverID, orgID)
	return args.Get(0).([]*models.DriverSchedule), args.Error(1)
}

func (m *MockDriverRepository) UpdateSchedule(ctx context.Context, schedule *models.DriverSchedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockDriverRepository) DeleteSchedule(ctx context.Context, id, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}

func TestUpdateDutyStatus(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewDriverService(mockRepo, logger)

	ctx := context.Background()
	driverID := uuid.New().String()
	orgID := uuid.New()
	req := &models.DutyStatusLog{
		Status:   "ON_DUTY",
		Location: &models.Location{Latitude: 28.6, Longitude: 77.2},
	}

	// Mock for DutyStatusService inside DriverService
	mockRepo.On("GetLatestDutyStatus", ctx, driverID).Return(&models.DutyStatusLog{Status: "OFF_DUTY"}, nil)
	mockRepo.On("EndDutyStatusLog", ctx, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("CreateDutyStatusLog", ctx, mock.MatchedBy(func(log *models.DutyStatusLog) bool {
		return log.DriverID == driverID && log.Status == "ON_DUTY"
	})).Return(nil)

	err := service.UpdateDutyStatus(ctx, driverID, orgID, req)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSendMessage(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewDriverService(mockRepo, logger)

	ctx := context.Background()
	driverID := "driver-123"
	orgID := "org-123"

	mockRepo.On("GetByID", ctx, driverID).Return(&models.Driver{ID: driverID, OrgID: orgID}, nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*models.DriverMessage")).Return(nil)

	msg, err := service.SendMessage(ctx, driverID, orgID, "Hello Dispatch", "CHAT", "")
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
	orgID := "org-123"
	req := &models.DriverExpense{Amount: 50.0, Currency: "USD"}

	mockRepo.On("GetByID", ctx, driverID).Return(&models.Driver{ID: driverID, OrgID: orgID}, nil)
	mockRepo.On("CreateExpense", ctx, mock.AnythingOfType("*models.DriverExpense")).Return(nil)

	res, err := service.SubmitExpense(ctx, driverID, orgID, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "PENDING", res.Status)
	mockRepo.AssertExpectations(t)
}
