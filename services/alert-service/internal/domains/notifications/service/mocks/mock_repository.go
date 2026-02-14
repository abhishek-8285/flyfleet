package mocks

import (
	"context"

	"github.com/flyfleet/alert-service/internal/domains/notifications/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) ListLogs(ctx context.Context, page, limit int, status, channel, search string) ([]models.NotificationLog, int, error) {
	args := m.Called(ctx, page, limit, status, channel, search)
	return args.Get(0).([]models.NotificationLog), args.Int(1), args.Error(2)
}

func (m *MockNotificationRepository) GetLog(ctx context.Context, id uuid.UUID) (*models.NotificationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationLog), args.Error(1)
}

func (m *MockNotificationRepository) GetSettings(ctx context.Context) ([]models.NotificationSettings, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.NotificationSettings), args.Error(1)
}

func (m *MockNotificationRepository) UpdateSettings(ctx context.Context, s *models.NotificationSettings) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockNotificationRepository) CreateLog(ctx context.Context, l *models.NotificationLog) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockNotificationRepository) AddEvent(ctx context.Context, e *models.NotificationEvent) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockNotificationRepository) ListPreferences(ctx context.Context, page, limit int, search string) ([]models.UserNotificationPreference, int, error) {
	args := m.Called(ctx, page, limit, search)
	return args.Get(0).([]models.UserNotificationPreference), args.Int(1), args.Error(2)
}

func (m *MockNotificationRepository) GetUserPreference(ctx context.Context, userId uuid.UUID) (*models.UserNotificationPreference, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserNotificationPreference), args.Error(1)
}

func (m *MockNotificationRepository) UpdateUserPreference(ctx context.Context, p *models.UserNotificationPreference) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockNotificationRepository) AddPreferenceAudit(ctx context.Context, a *models.PreferenceAudit) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetGlobalDefaults(ctx context.Context) (*models.GlobalNotificationDefault, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GlobalNotificationDefault), args.Error(1)
}

func (m *MockNotificationRepository) UpdateGlobalDefaults(ctx context.Context, d *models.GlobalNotificationDefault) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetPreferenceAudit(ctx context.Context, userId uuid.UUID) ([]models.PreferenceAudit, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]models.PreferenceAudit), args.Error(1)
}
