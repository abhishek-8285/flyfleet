package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/settings/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SettingsService struct {
	repo   repository.Repository
	logger *zap.Logger
}

func NewSettingsService(repo repository.Repository, logger *zap.Logger) *SettingsService {
	return &SettingsService{repo: repo, logger: logger}
}

// Helper to log audit events
func (s *SettingsService) logAudit(ctx context.Context, entityType string, entityID *uuid.UUID, action string, changes map[string]any) {
	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		s.logger.Warn("Audit log missing user ID")
		userID = uuid.Nil
	}

	ip, _ := ctx.Value(middleware.UserIPKey).(string)

	log := &models.AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		ActorID:    userID,
		ActorIP:    ip,
		Changes:    sanitizeAuditChanges(changes),
	}

	if err := s.repo.CreateAuditLog(ctx, log); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}
}

// Platform Settings
func (s *SettingsService) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	settings, err := s.repo.GetPlatformSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform settings: %w", err)
	}
	return settings, nil
}

func (s *SettingsService) ListAuditLogs(ctx context.Context, query models.AuditLogQuery) ([]models.AuditLog, error) {
	if query.Limit == 0 {
		query.Limit = 50
	}
	if query.Limit < 1 || query.Limit > 500 {
		return nil, fmt.Errorf("invalid limit: must be between 1 and 500: %w", ErrInvalidValue)
	}
	if query.Offset < 0 {
		return nil, fmt.Errorf("invalid offset: must be >= 0: %w", ErrInvalidValue)
	}
	query.EntityType = strings.ToUpper(strings.TrimSpace(query.EntityType))
	query.Action = strings.ToUpper(strings.TrimSpace(query.Action))

	logs, err := s.repo.ListAuditLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	return logs, nil
}

func (s *SettingsService) GetSettingsExport(ctx context.Context) (*models.SettingsExport, error) {
	platform, err := s.repo.GetPlatformSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform settings: %w", err)
	}
	flags, err := s.repo.ListFeatureFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}
	localization, err := s.repo.GetLocalizationSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get localization settings: %w", err)
	}
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	channels, err := s.ListNotificationChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification channels: %w", err)
	}
	limits, err := s.repo.GetSystemLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system limits: %w", err)
	}

	return &models.SettingsExport{
		PlatformSettings:     platform,
		FeatureFlags:         flags,
		LocalizationSettings: localization,
		Templates:            templates,
		NotificationChannels: channels,
		SystemLimits:         limits,
		ExportedAt:           time.Now().UTC(),
	}, nil
}

func (s *SettingsService) ImportSettings(ctx context.Context, payload *models.SettingsImportRequest) (*models.SettingsImportResult, error) {
	if payload == nil {
		return nil, errors.New("import payload is required")
	}
	if payload.PlatformSettings == nil &&
		len(payload.FeatureFlags) == 0 &&
		len(payload.LocalizationSettings) == 0 &&
		len(payload.Templates) == 0 &&
		len(payload.NotificationChannels) == 0 &&
		payload.SystemLimits == nil {
		return nil, errors.New("at least one settings section is required")
	}

	seenFlagKeys := make(map[string]struct{}, len(payload.FeatureFlags))
	for i := range payload.FeatureFlags {
		normalizedKey := strings.ToLower(strings.TrimSpace(payload.FeatureFlags[i].Key))
		if normalizedKey == "" {
			return nil, errors.New("flag key is required")
		}
		if _, exists := seenFlagKeys[normalizedKey]; exists {
			return nil, errors.New("duplicate flag key in import payload")
		}
		seenFlagKeys[normalizedKey] = struct{}{}
		payload.FeatureFlags[i].Key = normalizedKey
	}

	seenTemplateKeys := make(map[string]struct{}, len(payload.Templates))
	for i := range payload.Templates {
		normalizedKey := strings.ToLower(strings.TrimSpace(payload.Templates[i].Key))
		if normalizedKey == "" {
			return nil, errors.New("template key is required")
		}
		if _, exists := seenTemplateKeys[normalizedKey]; exists {
			return nil, errors.New("duplicate template key in import payload")
		}
		seenTemplateKeys[normalizedKey] = struct{}{}
		payload.Templates[i].Key = normalizedKey
	}

	seenLanguageCodes := make(map[string]struct{}, len(payload.LocalizationSettings))
	for i := range payload.LocalizationSettings {
		languageCode := strings.TrimSpace(payload.LocalizationSettings[i].LanguageCode)
		if languageCode == "" {
			return nil, errors.New("language code is required")
		}
		normalizedLang := strings.ToLower(languageCode)
		if _, exists := seenLanguageCodes[normalizedLang]; exists {
			return nil, errors.New("duplicate language code in import payload")
		}
		seenLanguageCodes[normalizedLang] = struct{}{}
		payload.LocalizationSettings[i].LanguageCode = languageCode
	}

	seenChannelTypes := make(map[string]struct{}, len(payload.NotificationChannels))
	for i := range payload.NotificationChannels {
		channelType := strings.ToUpper(strings.TrimSpace(payload.NotificationChannels[i].ChannelType))
		if channelType == "" {
			return nil, errors.New("channel type is required")
		}
		if _, exists := seenChannelTypes[channelType]; exists {
			return nil, errors.New("duplicate channel type in import payload")
		}
		seenChannelTypes[channelType] = struct{}{}
		payload.NotificationChannels[i].ChannelType = channelType
	}

	var result *models.SettingsImportResult
	if err := s.repo.WithTx(ctx, func(txRepo repository.Repository) error {
		txService := &SettingsService{
			repo:   txRepo,
			logger: s.logger,
		}
		var err error
		result, err = txService.importSettingsMutating(ctx, payload)
		return err
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SettingsService) importSettingsMutating(ctx context.Context, payload *models.SettingsImportRequest) (*models.SettingsImportResult, error) {
	result := &models.SettingsImportResult{}

	if payload.PlatformSettings != nil {
		if err := s.UpdatePlatformSettings(ctx, payload.PlatformSettings); err != nil {
			return nil, err
		}
		result.PlatformUpdated = true
	}

	existingFlags, err := s.repo.ListFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}
	existingFlagsByKey := make(map[string]models.FeatureFlag, len(existingFlags))
	for _, f := range existingFlags {
		existingFlagsByKey[strings.ToLower(strings.TrimSpace(f.Key))] = f
	}
	importedFlagKeys := make(map[string]struct{}, len(payload.FeatureFlags))

	for i := range payload.FeatureFlags {
		flag := payload.FeatureFlags[i]
		normalizedKey := flag.Key
		importedFlagKeys[normalizedKey] = struct{}{}

		if existing, ok := existingFlagsByKey[normalizedKey]; ok {
			flag.ID = existing.ID
			if err := s.UpdateFeatureFlag(ctx, &flag); err != nil {
				return nil, err
			}
			result.FeatureFlagsUpdated++
			continue
		}
		if err := s.CreateFeatureFlag(ctx, &flag); err != nil {
			return nil, err
		}
		result.FeatureFlagsCreated++
	}

	if payload.ReplaceExisting {
		for _, existing := range existingFlags {
			normalizedKey := strings.ToLower(strings.TrimSpace(existing.Key))
			if _, keep := importedFlagKeys[normalizedKey]; keep {
				continue
			}
			if err := s.DeleteFeatureFlag(ctx, existing.ID); err != nil {
				return nil, err
			}
			result.FeatureFlagsDeleted++
		}
	}

	existingTemplates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	existingTemplatesByKey := make(map[string]models.Template, len(existingTemplates))
	for _, tpl := range existingTemplates {
		existingTemplatesByKey[strings.ToLower(strings.TrimSpace(tpl.Key))] = tpl
	}
	for i := range payload.Templates {
		template := payload.Templates[i]
		normalizedKey := template.Key
		if existing, ok := existingTemplatesByKey[normalizedKey]; ok {
			template.ID = existing.ID
			if err := s.UpdateTemplate(ctx, &template); err != nil {
				return nil, err
			}
			result.TemplatesUpdated++
			continue
		}
		if err := s.CreateTemplate(ctx, &template); err != nil {
			return nil, err
		}
		result.TemplatesCreated++
	}

	for i := range payload.LocalizationSettings {
		localization := payload.LocalizationSettings[i]
		if err := s.UpdateLocalizationSettings(ctx, &localization); err != nil {
			return nil, err
		}
		result.LocalizationUpserted++
	}

	for i := range payload.NotificationChannels {
		channel := payload.NotificationChannels[i]
		if err := s.UpdateNotificationChannel(ctx, &channel); err != nil {
			return nil, err
		}
		result.ChannelsUpserted++
	}

	if payload.SystemLimits != nil {
		if err := s.UpdateSystemLimits(ctx, payload.SystemLimits); err != nil {
			return nil, err
		}
		result.SystemLimitsUpdated = true
	}

	s.logAudit(ctx, "SETTINGS_IMPORT", nil, "IMPORT", map[string]any{
		"replace_existing": payload.ReplaceExisting,
		"result":           result,
	})

	return result, nil
}

func (s *SettingsService) UpdatePlatformSettings(ctx context.Context, settings *models.PlatformSettings) error {
	settings.SystemName = strings.TrimSpace(settings.SystemName)
	settings.SupportEmail = strings.TrimSpace(strings.ToLower(settings.SupportEmail))
	settings.DefaultTimezone = strings.TrimSpace(settings.DefaultTimezone)
	settings.DefaultCurrency = strings.TrimSpace(strings.ToUpper(settings.DefaultCurrency))
	settings.DefaultLanguage = strings.TrimSpace(settings.DefaultLanguage)

	if settings.SystemName == "" {
		return errors.New("system name is required")
	}
	if settings.SupportEmail == "" {
		return errors.New("support email is required")
	}
	if !isValidEmail(settings.SupportEmail) {
		return errors.New("invalid support email format")
	}
	if !isValidTimezone(settings.DefaultTimezone) {
		return errors.New("invalid timezone")
	}
	// Enforce singleton identity for global platform settings.
	settings.ID = models.PlatformSettingsSingletonID
	userID := s.getUserID(ctx)
	settings.UpdatedBy = userID
	
	if err := s.repo.UpdatePlatformSettings(ctx, settings); err != nil {
		return err
	}
	s.logAudit(ctx, "PLATFORM_SETTINGS", &settings.ID, "UPDATE", map[string]any{"system_name": settings.SystemName})
	return nil
}

// Feature Flags
func (s *SettingsService) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	flags, err := s.repo.ListFeatureFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}
	return flags, nil
}

func (s *SettingsService) GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error) {
	flag, err := s.repo.GetFeatureFlag(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature flag: %w", err)
	}
	return flag, nil
}

func (s *SettingsService) CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	if f.Key == "" {
		return fmt.Errorf("flag key is required: %w", ErrRequiredField)
	}
	userID := s.getUserID(ctx)
	f.CreatedBy = userID
	f.UpdatedBy = userID
	if err := s.repo.CreateFeatureFlag(ctx, f); err != nil {
		return fmt.Errorf("failed to create feature flag: %w", err)
	}
	s.logAudit(ctx, "FEATURE_FLAG", &f.ID, "CREATE", map[string]any{"key": f.Key})
	return nil
}

func (s *SettingsService) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	f.UpdatedBy = s.getUserID(ctx)
	if err := s.repo.UpdateFeatureFlag(ctx, f); err != nil {
		return fmt.Errorf("failed to update feature flag: %w", err)
	}
	s.logAudit(ctx, "FEATURE_FLAG", &f.ID, "UPDATE", map[string]any{"key": f.Key, "enabled": f.Enabled})
	return nil
}

func (s *SettingsService) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteFeatureFlag(ctx, id); err != nil {
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}
	s.logAudit(ctx, "FEATURE_FLAG", &id, "DELETE", nil)
	return nil
}

func (s *SettingsService) ToggleFeatureFlag(ctx context.Context, id uuid.UUID) error {
	flag, err := s.repo.GetFeatureFlag(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get feature flag: %w", err)
	}
	flag.Enabled = !flag.Enabled
	flag.UpdatedBy = s.getUserID(ctx)
	if err := s.repo.UpdateFeatureFlag(ctx, flag); err != nil {
		return fmt.Errorf("failed to toggle feature flag: %w", err)
	}
	s.logAudit(ctx, "FEATURE_FLAG", &flag.ID, "TOGGLE", map[string]any{"enabled": flag.Enabled})
	return nil
}

// Localization
func (s *SettingsService) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	settings, err := s.repo.GetLocalizationSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get localization settings: %w", err)
	}
	return settings, nil
}

func (s *SettingsService) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	if l.LanguageCode == "" {
		return fmt.Errorf("language code is required: %w", ErrRequiredField)
	}
	userID := s.getUserID(ctx)
	l.CreatedBy = userID
	l.UpdatedBy = userID
	if err := s.repo.UpdateLocalizationSettings(ctx, l); err != nil {
		return fmt.Errorf("failed to update localization settings: %w", err)
	}
	s.logAudit(ctx, "LOCALIZATION", &l.ID, "UPDATE", map[string]any{"code": l.LanguageCode})
	return nil
}

// Templates
func (s *SettingsService) ListTemplates(ctx context.Context) ([]models.Template, error) {
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}

func (s *SettingsService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	template, err := s.repo.GetTemplate(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return template, nil
}

func (s *SettingsService) CreateTemplate(ctx context.Context, t *models.Template) error {
	if t.Key == "" {
		return fmt.Errorf("template key is required: %w", ErrRequiredField)
	}
	userID := s.getUserID(ctx)
	t.CreatedBy = userID
	t.UpdatedBy = userID
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	s.logAudit(ctx, "TEMPLATE", &t.ID, "CREATE", map[string]any{"key": t.Key})
	return nil
}

func (s *SettingsService) UpdateTemplate(ctx context.Context, t *models.Template) error {
	t.UpdatedBy = s.getUserID(ctx)
	if err := s.repo.UpdateTemplate(ctx, t); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}
	s.logAudit(ctx, "TEMPLATE", &t.ID, "UPDATE", map[string]any{"subject": t.Subject})
	return nil
}

// Notification Channels
func (s *SettingsService) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	channels, err := s.repo.ListNotificationChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification channels: %w", err)
	}
	// Decrypt configs
	for i := range channels {
		if encrypted, ok := channels[i].Config["_encrypted"].(string); ok {
			decryptedJSON, err := utils.Decrypt(encrypted)
			if err == nil {
				var decryptedMap map[string]any
				if err := json.Unmarshal([]byte(decryptedJSON), &decryptedMap); err == nil {
					channels[i].Config = decryptedMap
				}
			}
		}
	}
	return channels, nil
}

func (s *SettingsService) UpdateNotificationChannel(ctx context.Context, c *models.NotificationChannel) error {
	if c.ChannelType == "" {
		return fmt.Errorf("channel type is required: %w", ErrRequiredField)
	}

	// Encrypt Config
	if len(c.Config) > 0 {
		if _, ok := c.Config["_encrypted"]; !ok {
			configBytes, err := json.Marshal(c.Config)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			encrypted, err := utils.Encrypt(string(configBytes))
			if err != nil {
				return fmt.Errorf("failed to encrypt config: %w", err)
			}
			c.Config = map[string]any{"_encrypted": encrypted}
		}
	}

	userID := s.getUserID(ctx)
	c.CreatedBy = userID
	c.UpdatedBy = userID

	if err := s.repo.UpdateNotificationChannel(ctx, c); err != nil {
		return fmt.Errorf("failed to update notification channel: %w", err)
	}
	s.logAudit(ctx, "NOTIFICATION_CHANNEL", &c.ID, "UPDATE", map[string]any{"provider": c.Provider})
	return nil
}

// System Limits
func (s *SettingsService) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	limits, err := s.repo.GetSystemLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system limits: %w", err)
	}
	return limits, nil
}

func (s *SettingsService) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	if l.APIRateLimit < 0 {
		return fmt.Errorf("api rate limit cannot be negative: %w", ErrNegativeValue)
	}
	userID := s.getUserID(ctx)
	l.CreatedBy = userID
	l.UpdatedBy = userID
	if err := s.repo.UpdateSystemLimits(ctx, l); err != nil {
		return fmt.Errorf("failed to update system limits: %w", err)
	}
	s.logAudit(ctx, "SYSTEM_LIMITS", &l.ID, "UPDATE", map[string]any{"rate_limit": l.APIRateLimit})
	return nil
}

func (s *SettingsService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value(middleware.UserIDKey)
	if val == nil {
		return uuid.Nil
	}
	uid, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return uid
}

// CleanupOldAuditLogs removes audit logs older than the configured retention period
func (s *SettingsService) CleanupOldAuditLogs(ctx context.Context) (int64, error) {
	// Get system limits to determine retention period
	limits, err := s.repo.GetSystemLimits(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get system limits: %w", err)
	}

	// Default to 1 year if not specified
	retentionPeriod := "1 year"
	if limits.AuditRetention != "" {
		retentionPeriod = limits.AuditRetention
	}

	deleted, err := s.repo.CleanupOldAuditLogs(ctx, retentionPeriod)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup audit logs: %w", err)
	}

	s.logAudit(ctx, "AUDIT_CLEANUP", nil, "CLEANUP", map[string]any{
		"deleted_count": deleted,
		"retention":     retentionPeriod,
	})

	s.logger.Info("Audit log cleanup completed", zap.Int64("deleted", deleted), zap.String("retention", retentionPeriod))
	return deleted, nil
}
