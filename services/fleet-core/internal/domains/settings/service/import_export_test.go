package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/settings/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/settings/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Extend the existing MockRepository from service_test.go with missing methods
func (m *MockRepository) CleanupOldAuditLogs(ctx context.Context, retentionPeriod string) (int64, error) {
	// For testing, return a fixed count based on retention period
	return 100, nil
}

func setupTestService(t *testing.T) (*SettingsService, *MockRepository) {
	logger := zap.NewNop()
	mockRepo := &MockRepository{
		FeatureFlags:         make(map[uuid.UUID]models.FeatureFlag),
		LocalizationSettings: make(map[string]models.LocalizationSettings),
		Templates:            make(map[uuid.UUID]models.Template),
		NotificationChannels: make(map[uuid.UUID]models.NotificationChannel),
	}
	service := NewSettingsService(mockRepo, logger)
	return service, mockRepo
}

func TestGetSettingsExport_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	now := time.Now()
	userID := uuid.New()

	// Setup test data
	mockRepo.PlatformSettings = &models.PlatformSettings{
		ID:              models.PlatformSettingsSingletonID,
		SystemName:      "Test Platform",
		SupportEmail:    "test@example.com",
		DefaultTimezone: "UTC",
		CreatedAt:       now,
		CreatedBy:       userID,
	}
	mockRepo.FeatureFlags = map[uuid.UUID]models.FeatureFlag{
		uuid.New(): {
			ID:          uuid.New(),
			Key:         "feature_one",
			Description: "First feature",
			Enabled:     true,
		},
	}
	mockRepo.LocalizationSettings = map[string]models.LocalizationSettings{
		"en-US": {
			ID:           uuid.New(),
			LanguageCode: "en-US",
			IsActive:     true,
			IsDefault:    true,
		},
	}
	mockRepo.Templates = map[uuid.UUID]models.Template{
		uuid.New(): {
			ID:      uuid.New(),
			Key:     "welcome_email",
			Channel: "EMAIL",
			Subject: "Welcome",
		},
	}
	mockRepo.NotificationChannels = map[uuid.UUID]models.NotificationChannel{
		uuid.New(): {
			ID:          uuid.New(),
			ChannelType: "EMAIL",
			Provider:    "SendGrid",
			Enabled:     true,
		},
	}
	mockRepo.SystemLimits = &models.SystemLimit{
		ID:             uuid.New(),
		APIRateLimit:   10000,
		BurstLimit:     500,
		AuditRetention: "1y",
	}

	// Execute
	export, err := service.GetSettingsExport(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, export)
	assert.Equal(t, "Test Platform", export.PlatformSettings.SystemName)
	assert.GreaterOrEqual(t, len(export.FeatureFlags), 1)
	assert.GreaterOrEqual(t, len(export.LocalizationSettings), 1)
	assert.GreaterOrEqual(t, len(export.Templates), 1)
	assert.GreaterOrEqual(t, len(export.NotificationChannels), 1)
	assert.NotNil(t, export.SystemLimits)
	assert.False(t, export.ExportedAt.IsZero())
}

func TestGetSettingsExport_EmptyData(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	// Setup minimal data
	mockRepo.PlatformSettings = &models.PlatformSettings{
		ID:              models.PlatformSettingsSingletonID,
		SystemName:      "Default Platform",
		SupportEmail:    "default@fleet-system.com",
		DefaultTimezone: "UTC",
		DefaultCurrency: "USD",
		DefaultLanguage: "en-US",
	}
	mockRepo.SystemLimits = &models.SystemLimit{
		ID:             uuid.New(),
		APIRateLimit:   10000,
		BurstLimit:     500,
		AuditRetention: "1y",
	}

	export, err := service.GetSettingsExport(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, export)
	assert.Equal(t, "Default Platform", export.PlatformSettings.SystemName)
}

func TestImportSettings_NilPayload(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	result, err := service.ImportSettings(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "import payload is required")
}

func TestImportSettings_EmptyPayload(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.SettingsImportRequest{
		PlatformSettings:     nil,
		FeatureFlags:         []models.FeatureFlag{},
		LocalizationSettings: []models.LocalizationSettings{},
		Templates:            []models.Template{},
		NotificationChannels: []models.NotificationChannel{},
		SystemLimits:         nil,
	}

	result, err := service.ImportSettings(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one settings section is required")
}

func TestImportSettings_DuplicateFlagKey(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.SettingsImportRequest{
		FeatureFlags: []models.FeatureFlag{
			{Key: "duplicate_key", Enabled: true},
			{Key: "duplicate_key", Enabled: false},
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate flag key")
}

func TestImportSettings_DuplicateTemplateKey(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.SettingsImportRequest{
		Templates: []models.Template{
			{Key: "dup_template", Channel: "EMAIL"},
			{Key: "dup_template", Channel: "SMS"},
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate template key")
}

func TestImportSettings_DuplicateLanguageCode(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.SettingsImportRequest{
		LocalizationSettings: []models.LocalizationSettings{
			{LanguageCode: "en-US", IsActive: true},
			{LanguageCode: "en-US", IsActive: false},
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate language code")
}

func TestImportSettings_DuplicateChannelType(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.SettingsImportRequest{
		NotificationChannels: []models.NotificationChannel{
			{ChannelType: "EMAIL", Provider: "SendGrid"},
			{ChannelType: "EMAIL", Provider: "SES"},
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate channel type")
}

func TestImportSettings_PlatformOnly(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	req := &models.SettingsImportRequest{
		PlatformSettings: &models.PlatformSettings{
			SystemName:   "Imported Platform",
			SupportEmail: "imported@example.com",
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.PlatformUpdated)
	assert.NotNil(t, mockRepo.PlatformSettings)
}

func TestImportSettings_FullImport_ReplaceExisting(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	existingFlagID := uuid.New()

	// Setup existing data
	mockRepo.FeatureFlags = map[uuid.UUID]models.FeatureFlag{
		existingFlagID: {
			ID:      existingFlagID,
			Key:     "old_feature",
			Enabled: false,
		},
	}

	req := &models.SettingsImportRequest{
		PlatformSettings: &models.PlatformSettings{
			SystemName:   "New Platform",
			SupportEmail: "new@example.com",
		},
		FeatureFlags: []models.FeatureFlag{
			{Key: "new_feature", Enabled: true},
		},
		ReplaceExisting: true,
	}

	result, err := service.ImportSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.PlatformUpdated)
	assert.Equal(t, 1, result.FeatureFlagsCreated)
	assert.Equal(t, 1, result.FeatureFlagsDeleted)
	assert.Len(t, mockRepo.FeatureFlags, 1) // Only new_feature should remain
}

func TestImportSettings_UpdateExistingFlag(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	existingFlagID := uuid.New()

	// Setup existing data
	mockRepo.FeatureFlags = map[uuid.UUID]models.FeatureFlag{
		existingFlagID: {
			ID:      existingFlagID,
			Key:     "existing_feature",
			Enabled: false,
		},
	}

	req := &models.SettingsImportRequest{
		FeatureFlags: []models.FeatureFlag{
			{Key: "existing_feature", Enabled: true},
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.FeatureFlagsUpdated)
	assert.Equal(t, 0, result.FeatureFlagsCreated)
	assert.Len(t, mockRepo.FeatureFlags, 1)
}

func TestSettingsExport_JSONSerialization(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	export := &models.SettingsExport{
		PlatformSettings: &models.PlatformSettings{
			ID:              models.PlatformSettingsSingletonID,
			SystemName:      "Test Platform",
			SupportEmail:    "test@example.com",
			DefaultTimezone: "UTC",
			CreatedAt:       now,
			CreatedBy:       userID,
		},
		FeatureFlags: []models.FeatureFlag{
			{
				ID:          uuid.New(),
				Key:         "test_flag",
				Description: "Test feature flag",
				Enabled:     true,
				Scope:       "GLOBAL",
			},
		},
		LocalizationSettings: []models.LocalizationSettings{
			{
				ID:           uuid.New(),
				LanguageCode: "en-US",
				IsActive:     true,
				IsDefault:    true,
			},
		},
		Templates: []models.Template{
			{
				ID:      uuid.New(),
				Key:     "test_template",
				Channel: "EMAIL",
				Subject: "Test Subject",
				Body:    "Test body content",
			},
		},
		NotificationChannels: []models.NotificationChannel{
			{
				ID:          uuid.New(),
				ChannelType: "EMAIL",
				Provider:    "SendGrid",
				Enabled:     true,
			},
		},
		SystemLimits: &models.SystemLimit{
			ID:             uuid.New(),
			APIRateLimit:   10000,
			BurstLimit:     500,
			AuditRetention: "1y",
		},
		ExportedAt: now,
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(export, "", "  ")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize back
	var imported models.SettingsExport
	err = json.Unmarshal(data, &imported)
	assert.NoError(t, err)
	assert.Equal(t, export.PlatformSettings.SystemName, imported.PlatformSettings.SystemName)
	assert.Len(t, imported.FeatureFlags, 1)
	assert.Equal(t, export.FeatureFlags[0].Key, imported.FeatureFlags[0].Key)
}

func TestImportSettings_Normalization(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	req := &models.SettingsImportRequest{
		FeatureFlags: []models.FeatureFlag{
			{Key: "  UPPER_CASE_FLAG  ", Enabled: true},
		},
		LocalizationSettings: []models.LocalizationSettings{
			{LanguageCode: "  EN-US  ", IsActive: true},
		},
		NotificationChannels: []models.NotificationChannel{
			{ChannelType: "  email  ", Provider: "SendGrid"},
		},
	}

	result, err := service.ImportSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Verify normalization happened
	assert.Equal(t, "upper_case_flag", req.FeatureFlags[0].Key)
	assert.Equal(t, "EN-US", req.LocalizationSettings[0].LanguageCode)
	assert.Equal(t, "EMAIL", req.NotificationChannels[0].ChannelType)
	assert.Len(t, mockRepo.FeatureFlags, 1)
	assert.Len(t, mockRepo.LocalizationSettings, 1)
	assert.Len(t, mockRepo.NotificationChannels, 1)
}

func TestCleanupOldAuditLogs_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	mockRepo.SystemLimits = &models.SystemLimit{
		AuditRetention: "1 year",
	}

	deleted, err := service.CleanupOldAuditLogs(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int64(100), deleted)
}

func TestCleanupOldAuditLogs_DefaultRetention(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())

	mockRepo.SystemLimits = &models.SystemLimit{
		AuditRetention: "",
	}

	deleted, err := service.CleanupOldAuditLogs(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int64(100), deleted)
}
