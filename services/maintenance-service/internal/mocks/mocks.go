package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/flyfleet/maintenance-service/internal/models"
)

// MockMaintenanceRepositoryImpl is a mock implementation of MaintenanceRepositoryImpl
type MockMaintenanceRepositoryImpl struct {
	mock.Mock
}

func NewMockMaintenanceRepositoryImpl() *MockMaintenanceRepositoryImpl {
	return &MockMaintenanceRepositoryImpl{}
}

func (m *MockMaintenanceRepositoryImpl) CreateSchedule(ctx context.Context, schedule *models.MaintenanceSchedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockMaintenanceRepositoryImpl) GetScheduleByID(ctx context.Context, id string) (*models.MaintenanceSchedule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MaintenanceSchedule), args.Error(1)
}

func (m *MockMaintenanceRepositoryImpl) ListSchedules(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceSchedule, int, error) {
	args := m.Called(ctx, filter)
	schedules, _ := args.Get(0).([]*models.MaintenanceSchedule)
	count, _ := args.Get(1).(int)
	return schedules, count, args.Error(2)
}

func (m *MockMaintenanceRepositoryImpl) UpdateSchedule(ctx context.Context, id string, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockMaintenanceRepositoryImpl) DeleteSchedule(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMaintenanceRepositoryImpl) CreateLog(ctx context.Context, log *models.MaintenanceLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockMaintenanceRepositoryImpl) GetLogByID(ctx context.Context, id string) (*models.MaintenanceLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MaintenanceLog), args.Error(1)
}

func (m *MockMaintenanceRepositoryImpl) ListLogs(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceLog, int, error) {
	args := m.Called(ctx, filter)
	logs, _ := args.Get(0).([]*models.MaintenanceLog)
	count, _ := args.Get(1).(int)
	return logs, count, args.Error(2)
}

func (m *MockMaintenanceRepositoryImpl) UpdateLog(ctx context.Context, id string, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockMaintenanceRepositoryImpl) DeleteLog(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMaintenanceRepositoryImpl) GetMaintenanceHistory(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceLog, error) {
	args := m.Called(ctx, vehicleID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MaintenanceLog), args.Error(1)
}

func (m *MockMaintenanceRepositoryImpl) GetOverdueSchedules(ctx context.Context, orgID uuid.UUID) ([]*models.MaintenanceSchedule, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MaintenanceSchedule), args.Error(1)
}

func (m *MockMaintenanceRepositoryImpl) GetUpcomingSchedules(ctx context.Context, orgID uuid.UUID, days int) ([]*models.MaintenanceSchedule, error) {
	args := m.Called(ctx, orgID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MaintenanceSchedule), args.Error(1)
}

func (m *MockMaintenanceRepositoryImpl) GetMaintenanceSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.MaintenanceSummary, error) {
	args := m.Called(ctx, orgID, fromDate, toDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MaintenanceSummary), args.Error(1)
}