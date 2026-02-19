package service

import (
	"context"
	"errors"
	"testing"

	"github.com/flyfleet/settings-service/internal/models"
	"github.com/flyfleet/settings-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Mock Repository
type MockRepository struct {
	PlatformSettings     *models.PlatformSettings
	FeatureFlags         map[uuid.UUID]models.FeatureFlag
	LocalizationSettings map[string]models.LocalizationSettings
	Templates            map[uuid.UUID]models.Template
	NotificationChannels map[uuid.UUID]models.NotificationChannel
	SystemLimits         *models.SystemLimit
	AuditLogs            []models.AuditLog
	LastAuditQuery       models.AuditLogQuery
	FailCreateTemplate   bool
}

func cloneFeatureFlags(src map[uuid.UUID]models.FeatureFlag) map[uuid.UUID]models.FeatureFlag {
	if src == nil {
		return nil
	}
	out := make(map[uuid.UUID]models.FeatureFlag, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func cloneLocalization(src map[string]models.LocalizationSettings) map[string]models.LocalizationSettings {
	if src == nil {
		return nil
	}
	out := make(map[string]models.LocalizationSettings, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func cloneTemplates(src map[uuid.UUID]models.Template) map[uuid.UUID]models.Template {
	if src == nil {
		return nil
	}
	out := make(map[uuid.UUID]models.Template, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func cloneChannels(src map[uuid.UUID]models.NotificationChannel) map[uuid.UUID]models.NotificationChannel {
	if src == nil {
		return nil
	}
	out := make(map[uuid.UUID]models.NotificationChannel, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (m *MockRepository) WithTx(ctx context.Context, fn func(repository.Repository) error) error {
	staged := &MockRepository{
		PlatformSettings:     m.PlatformSettings,
		FeatureFlags:         cloneFeatureFlags(m.FeatureFlags),
		LocalizationSettings: cloneLocalization(m.LocalizationSettings),
		Templates:            cloneTemplates(m.Templates),
		NotificationChannels: cloneChannels(m.NotificationChannels),
		SystemLimits:         m.SystemLimits,
		AuditLogs:            append([]models.AuditLog{}, m.AuditLogs...),
		LastAuditQuery:       m.LastAuditQuery,
		FailCreateTemplate:   m.FailCreateTemplate,
	}
	if err := fn(staged); err != nil {
		return err
	}
	m.PlatformSettings = staged.PlatformSettings
	m.FeatureFlags = staged.FeatureFlags
	m.LocalizationSettings = staged.LocalizationSettings
	m.Templates = staged.Templates
	m.NotificationChannels = staged.NotificationChannels
	m.SystemLimits = staged.SystemLimits
	m.AuditLogs = staged.AuditLogs
	m.LastAuditQuery = staged.LastAuditQuery
	return nil
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
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if m.FeatureFlags == nil {
		m.FeatureFlags = make(map[uuid.UUID]models.FeatureFlag)
	}
	m.FeatureFlags[f.ID] = *f
	return nil
}

func (m *MockRepository) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	if m.FeatureFlags == nil {
		m.FeatureFlags = make(map[uuid.UUID]models.FeatureFlag)
	}
	m.FeatureFlags[f.ID] = *f
	return nil
}

func (m *MockRepository) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	delete(m.FeatureFlags, id)
	return nil
}

func (m *MockRepository) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	if m.LocalizationSettings == nil {
		return nil, nil
	}
	out := make([]models.LocalizationSettings, 0, len(m.LocalizationSettings))
	for _, l := range m.LocalizationSettings {
		out = append(out, l)
	}
	return out, nil
}

func (m *MockRepository) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	if m.LocalizationSettings == nil {
		m.LocalizationSettings = make(map[string]models.LocalizationSettings)
	}
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	m.LocalizationSettings[l.LanguageCode] = *l
	return nil
}

func (m *MockRepository) ListTemplates(ctx context.Context) ([]models.Template, error) {
	if m.Templates == nil {
		return nil, nil
	}
	out := make([]models.Template, 0, len(m.Templates))
	for _, tpl := range m.Templates {
		out = append(out, tpl)
	}
	return out, nil
}

func (m *MockRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	return nil, nil
}

func (m *MockRepository) CreateTemplate(ctx context.Context, t *models.Template) error {
	if m.FailCreateTemplate {
		return errors.New("template insert failed")
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if m.Templates == nil {
		m.Templates = make(map[uuid.UUID]models.Template)
	}
	m.Templates[t.ID] = *t
	return nil
}

func (m *MockRepository) UpdateTemplate(ctx context.Context, t *models.Template) error {
	if m.Templates == nil {
		m.Templates = make(map[uuid.UUID]models.Template)
	}
	m.Templates[t.ID] = *t
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
	if m.NotificationChannels == nil {
		m.NotificationChannels = make(map[uuid.UUID]models.NotificationChannel)
	}
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	m.NotificationChannels[c.ID] = *c
	return nil
}

func (m *MockRepository) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	return m.SystemLimits, nil
}

func (m *MockRepository) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	m.SystemLimits = l
	return nil
}

func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.AuditLogs = append(m.AuditLogs, *log)
	return nil
}

func (m *MockRepository) ListAuditLogs(ctx context.Context, query models.AuditLogQuery) ([]models.AuditLog, error) {
	m.LastAuditQuery = query
	end := query.Offset + query.Limit
	if query.Offset > len(m.AuditLogs) {
		return []models.AuditLog{}, nil
	}
	if end > len(m.AuditLogs) {
		end = len(m.AuditLogs)
	}
	return append([]models.AuditLog{}, m.AuditLogs[query.Offset:end]...), nil
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
		assert  func(t *testing.T, repo *MockRepository)
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
			name: "Normalizes Input Casing and Whitespace",
			input: &models.PlatformSettings{
				SystemName:      "  Enterprise Ops  ",
				SupportEmail:    "  SUPPORT@EXAMPLE.COM  ",
				DefaultTimezone: "  UTC  ",
				DefaultCurrency: " usd ",
				DefaultLanguage: " en-US ",
			},
			wantErr: "",
			assert: func(t *testing.T, repo *MockRepository) {
				got := repo.PlatformSettings
				if got == nil {
					t.Fatal("expected normalized settings to be stored")
				}
				if got.SystemName != "Enterprise Ops" {
					t.Fatalf("unexpected normalized system name: %q", got.SystemName)
				}
				if got.SupportEmail != "support@example.com" {
					t.Fatalf("unexpected normalized support email: %q", got.SupportEmail)
				}
				if got.DefaultTimezone != "UTC" {
					t.Fatalf("unexpected normalized timezone: %q", got.DefaultTimezone)
				}
				if got.DefaultCurrency != "USD" {
					t.Fatalf("unexpected normalized currency: %q", got.DefaultCurrency)
				}
			},
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
				if repo.PlatformSettings == nil {
					t.Fatal("expected platform settings to be stored")
				}
				if repo.PlatformSettings.ID != models.PlatformSettingsSingletonID {
					t.Errorf(
						"expected singleton ID %s, got %s",
						models.PlatformSettingsSingletonID.String(),
						repo.PlatformSettings.ID.String(),
					)
				}
				if len(repo.AuditLogs) != 1 {
					t.Errorf("expected 1 audit log, got %d", len(repo.AuditLogs))
				}
				if tt.assert != nil {
					tt.assert(t, repo)
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

func TestImportSettings_UpsertAndReplace(t *testing.T) {
	existingFlagID := uuid.New()
	existingTemplateID := uuid.New()
	existingFlag := models.FeatureFlag{
		ID:          existingFlagID,
		Key:         "legacy-flag",
		Description: "legacy",
		Scope:       "GLOBAL",
	}
	existingTemplate := models.Template{
		ID:      existingTemplateID,
		Key:     "trip-summary",
		Channel: "EMAIL",
		Body:    "old-body",
	}
	repo := &MockRepository{
		FeatureFlags: map[uuid.UUID]models.FeatureFlag{
			existingFlagID: existingFlag,
		},
		Templates: map[uuid.UUID]models.Template{
			existingTemplateID: existingTemplate,
		},
		NotificationChannels: make(map[uuid.UUID]models.NotificationChannel),
	}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	input := &models.SettingsImportRequest{
		PlatformSettings: &models.PlatformSettings{
			SystemName:      "  Ops Console  ",
			SupportEmail:    "  SUPPORT@EXAMPLE.COM  ",
			DefaultTimezone: "UTC",
		},
		FeatureFlags: []models.FeatureFlag{
			{Key: "new-flag", Scope: "GLOBAL", Enabled: true},
		},
		Templates: []models.Template{
			{Key: "trip-summary", Channel: "EMAIL", Subject: "Updated", Body: "new-body"},
		},
		LocalizationSettings: []models.LocalizationSettings{
			{LanguageCode: "en-US", IsActive: true, IsDefault: true},
		},
		NotificationChannels: []models.NotificationChannel{
			{ChannelType: "email", Provider: "SendGrid", Config: map[string]any{"api_key": "secret"}},
		},
		SystemLimits: &models.SystemLimit{
			APIRateLimit:      1000,
			BurstLimit:        100,
			AuditRetention:    "365d",
			ActivityRetention: "90d",
		},
		ReplaceExisting: true,
	}

	result, err := svc.ImportSettings(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error importing settings: %v", err)
	}

	if !result.PlatformUpdated {
		t.Fatal("expected platform settings to be updated")
	}
	if result.FeatureFlagsCreated != 1 || result.FeatureFlagsDeleted != 1 {
		t.Fatalf("unexpected feature flag import result: %+v", result)
	}
	if result.TemplatesUpdated != 1 {
		t.Fatalf("expected templates updated=1, got %+v", result)
	}
	if result.LocalizationUpserted != 1 || result.ChannelsUpserted != 1 || !result.SystemLimitsUpdated {
		t.Fatalf("unexpected import result counters: %+v", result)
	}
	if repo.PlatformSettings == nil || repo.PlatformSettings.SupportEmail != "support@example.com" {
		t.Fatalf("expected normalized platform support email, got %+v", repo.PlatformSettings)
	}
	if len(repo.FeatureFlags) != 1 {
		t.Fatalf("expected exactly one feature flag after replace, got %d", len(repo.FeatureFlags))
	}
	if len(repo.AuditLogs) == 0 {
		t.Fatal("expected audit logs to be created")
	}
}

func TestImportSettings_EmptyPayloadReturnsError(t *testing.T) {
	repo := &MockRepository{}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	_, err := svc.ImportSettings(context.Background(), &models.SettingsImportRequest{})
	if err == nil {
		t.Fatal("expected validation error for empty payload")
	}
	if err.Error() != "at least one settings section is required" {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.AuditLogs) != 0 {
		t.Fatalf("expected no audit logs on validation failure, got %d", len(repo.AuditLogs))
	}
}

func TestImportSettings_DuplicateFeatureFlagKeyReturnsError(t *testing.T) {
	repo := &MockRepository{}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	_, err := svc.ImportSettings(context.Background(), &models.SettingsImportRequest{
		FeatureFlags: []models.FeatureFlag{
			{Key: "dispatch-v2", Scope: "GLOBAL", Enabled: true},
			{Key: "  DISPATCH-v2  ", Scope: "GLOBAL", Enabled: false},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate key validation error")
	}
	if err.Error() != "duplicate flag key in import payload" {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.FeatureFlags) != 0 {
		t.Fatalf("expected no feature flags created, got %d", len(repo.FeatureFlags))
	}
}

func TestListAuditLogs_Validation(t *testing.T) {
	repo := &MockRepository{}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	if _, err := svc.ListAuditLogs(context.Background(), models.AuditLogQuery{Limit: 0, Offset: -1}); err == nil {
		t.Fatal("expected validation error for negative offset")
	}
	if _, err := svc.ListAuditLogs(context.Background(), models.AuditLogQuery{Limit: 501}); err == nil {
		t.Fatal("expected validation error for out-of-range limit")
	}
}

func TestListAuditLogs_DefaultsAndNormalization(t *testing.T) {
	repo := &MockRepository{
		AuditLogs: []models.AuditLog{
			{EntityType: "SETTINGS_IMPORT", Action: "IMPORT"},
		},
	}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	_, err := svc.ListAuditLogs(context.Background(), models.AuditLogQuery{
		Limit:      0,
		Offset:     0,
		EntityType: "settings_import",
		Action:     "import",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.LastAuditQuery.Limit != 50 {
		t.Fatalf("expected default limit 50, got %d", repo.LastAuditQuery.Limit)
	}
	if repo.LastAuditQuery.EntityType != "SETTINGS_IMPORT" {
		t.Fatalf("expected normalized entity_type SETTINGS_IMPORT, got %q", repo.LastAuditQuery.EntityType)
	}
	if repo.LastAuditQuery.Action != "IMPORT" {
		t.Fatalf("expected normalized action IMPORT, got %q", repo.LastAuditQuery.Action)
	}
}

func TestImportSettings_TransactionRollbackOnFailure(t *testing.T) {
	repo := &MockRepository{
		FeatureFlags:       map[uuid.UUID]models.FeatureFlag{},
		FailCreateTemplate: true,
	}
	logger := zap.NewNop()
	svc := NewSettingsService(repo, logger)

	_, err := svc.ImportSettings(context.Background(), &models.SettingsImportRequest{
		FeatureFlags: []models.FeatureFlag{
			{Key: "flag-before-failure", Scope: "GLOBAL", Enabled: true},
		},
		Templates: []models.Template{
			{Key: "template-failure", Channel: "EMAIL", Body: "body"},
		},
	})
	if err == nil {
		t.Fatal("expected error from template creation failure")
	}
	if len(repo.FeatureFlags) != 0 {
		t.Fatalf("expected feature flag rollback, got %d rows", len(repo.FeatureFlags))
	}
	if len(repo.AuditLogs) != 0 {
		t.Fatalf("expected audit rollback, got %d rows", len(repo.AuditLogs))
	}
}
