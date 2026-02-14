package service

import (
	"context"
	"errors"
	"testing"

	"github.com/flyfleet/settings-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Mock Repository
type MockRepository struct {
	PlatformSettings     *models.PlatformSettings
	FeatureFlags         map[uuid.UUID]models.FeatureFlag
	NotificationChannels map[uuid.UUID]models.NotificationChannel
	AuditLogs            []models.AuditLog
}

func (m *MockRepository) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	if m.PlatformSettings == nil {
		return nil, errors.New("record not found")
	}
	return m.PlatformSettings, nil
}

func (m *MockRepository) UpdatePlatformSettings(ctx context.Context, s *models.PlatformSettings) error {
	m.PlatformSettings = s
	return nil
}

func (m *MockRepository) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	for _, f := range m.FeatureFlags {
		flags = append(flags, f)
	}
	return flags, nil
}

func (m *MockRepository) GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error) {
	f, ok := m.FeatureFlags[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return &f, nil
}

func (m *MockRepository) CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	m.FeatureFlags[f.ID] = *f
	return nil
}

func (m *MockRepository) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	m.FeatureFlags[f.ID] = *f
	return nil
}

func (m *MockRepository) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	delete(m.FeatureFlags, id)
	return nil
}

func (m *MockRepository) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	return nil, nil
}

func (m *MockRepository) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	return nil
}

func (m *MockRepository) ListTemplates(ctx context.Context) ([]models.Template, error) {
	return nil, nil
}

func (m *MockRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	return nil, nil
}

func (m *MockRepository) CreateTemplate(ctx context.Context, t *models.Template) error {
	return nil
}

func (m *MockRepository) UpdateTemplate(ctx context.Context, t *models.Template) error {
	return nil
}

func (m *MockRepository) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	var channels []models.NotificationChannel
	for _, c := range m.NotificationChannels {
		channels = append(channels, c)
	}
	return channels, nil
}

func (m *MockRepository) UpdateNotificationChannel(ctx context.Context, c *models.NotificationChannel) error {
	m.NotificationChannels[c.ID] = *c
	return nil
}

func (m *MockRepository) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	return nil, nil
}

func (m *MockRepository) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	return nil
}

func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.AuditLogs = append(m.AuditLogs, *log)
	return nil
}

// Tests

func TestUpdatePlatformSettings_Validation(t *testing.T) {
	repo := &MockRepository{}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	tests := []struct {
		name    string
		input   *models.PlatformSettings
		wantErr string
	}{
		{
			name: "Valid Settings",
			input: &models.PlatformSettings{
				SystemName:      "Test System",
				SupportEmail:    "test@example.com",
				DefaultTimezone: "UTC",
			},
			wantErr: "",
		},
		{
			name: "Invalid Email",
			input: &models.PlatformSettings{
				SystemName:      "Test System",
				SupportEmail:    "invalid-email",
				DefaultTimezone: "UTC",
			},
			wantErr: "invalid support email format",
		},
		{
			name: "Invalid Timezone",
			input: &models.PlatformSettings{
				SystemName:      "Test System",
				SupportEmail:    "test@example.com",
				DefaultTimezone: "Invalid/Timezone",
			},
			wantErr: "invalid timezone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdatePlatformSettings(context.Background(), tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(repo.AuditLogs) != 1 {
					t.Errorf("expected 1 audit log, got %d", len(repo.AuditLogs))
				}
			}
			// Reset audit logs
			repo.AuditLogs = []models.AuditLog{}
		})
	}
}

func TestEncryption_EndToEnd(t *testing.T) {
	repo := &MockRepository{
		NotificationChannels: make(map[uuid.UUID]models.NotificationChannel),
	}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	// 1. Update with plain text config
	id := uuid.New()
	apiKey := "secret-api-key-123"
	channel := &models.NotificationChannel{
		ID:          id,
		ChannelType: "EMAIL",
		Provider:    "SendGrid",
		Config:      map[string]any{"api_key": apiKey},
		Enabled:     true,
	}

	err := svc.UpdateNotificationChannel(context.Background(), channel)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify it's stored encrypted in mock repo
	storedChannel := repo.NotificationChannels[id]
	if _, ok := storedChannel.Config["_encrypted"]; !ok {
		t.Error("Config should be encrypted in repository")
	}
	if val, ok := storedChannel.Config["api_key"]; ok {
		t.Errorf("Plain text api_key should not be in repository, got %v", val)
	}

	// 2. List and verify decryption
	channels, err := svc.ListNotificationChannels(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(channels) != 1 {
		t.Fatalf("Expected 1 channel, got %d", len(channels))
	}

	retrievedConfig := channels[0].Config
	if retrievedConfig["api_key"] != apiKey {
		t.Errorf("Expected api_key %q, got %v", apiKey, retrievedConfig["api_key"])
	}
}
