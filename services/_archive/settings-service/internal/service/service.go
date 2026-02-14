package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/flyfleet/settings-service/internal/middleware"
	"github.com/flyfleet/settings-service/internal/models"
	"github.com/flyfleet/settings-service/internal/repository"
	"github.com/flyfleet/settings-service/internal/utils"
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
		Changes:    changes,
	}

	if err := s.repo.CreateAuditLog(ctx, log); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}
}

// Platform Settings
func (s *SettingsService) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	return s.repo.GetPlatformSettings(ctx)
}

func (s *SettingsService) UpdatePlatformSettings(ctx context.Context, settings *models.PlatformSettings) error {
	if settings.SystemName == "" {
		return errors.New("system name is required")
	}
	if settings.SupportEmail == "" {
		return errors.New("support email is required")
	}
	if !isValidEmail(settings.SupportEmail) {
		return errors.New("invalid support email format")
	}
	if !isValidEmail(settings.SupportEmail) {
		return errors.New("invalid support email format")
	}
	if !isValidTimezone(settings.DefaultTimezone) {
		return errors.New("invalid timezone")
	}
	if err := s.repo.UpdatePlatformSettings(ctx, settings); err != nil {
		return err
	}
	s.logAudit(ctx, "PLATFORM_SETTINGS", &settings.ID, "UPDATE", map[string]any{"system_name": settings.SystemName})
	return nil
}

// Feature Flags
func (s *SettingsService) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	return s.repo.ListFeatureFlags(ctx)
}

func (s *SettingsService) GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error) {
	return s.repo.GetFeatureFlag(ctx, id)
}

func (s *SettingsService) CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	if f.Key == "" {
		return errors.New("flag key is required")
	}
	if err := s.repo.CreateFeatureFlag(ctx, f); err != nil {
		return err
	}
	s.logAudit(ctx, "FEATURE_FLAG", &f.ID, "CREATE", map[string]any{"key": f.Key})
	return nil
}

func (s *SettingsService) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error {
	if err := s.repo.UpdateFeatureFlag(ctx, f); err != nil {
		return err
	}
	s.logAudit(ctx, "FEATURE_FLAG", &f.ID, "UPDATE", map[string]any{"key": f.Key, "enabled": f.Enabled})
	return nil
}

func (s *SettingsService) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteFeatureFlag(ctx, id); err != nil {
		return err
	}
	s.logAudit(ctx, "FEATURE_FLAG", &id, "DELETE", nil)
	return nil
}

func (s *SettingsService) ToggleFeatureFlag(ctx context.Context, id uuid.UUID) error {
	flag, err := s.repo.GetFeatureFlag(ctx, id)
	if err != nil {
		return err
	}
	flag.Enabled = !flag.Enabled
	if err := s.repo.UpdateFeatureFlag(ctx, flag); err != nil {
		return err
	}
	s.logAudit(ctx, "FEATURE_FLAG", &flag.ID, "TOGGLE", map[string]any{"enabled": flag.Enabled})
	return nil
}

// Localization
func (s *SettingsService) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	return s.repo.GetLocalizationSettings(ctx)
}

func (s *SettingsService) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	if l.LanguageCode == "" {
		return errors.New("language code is required")
	}
	if err := s.repo.UpdateLocalizationSettings(ctx, l); err != nil {
		return err
	}
	s.logAudit(ctx, "LOCALIZATION", &l.ID, "UPDATE", map[string]any{"code": l.LanguageCode})
	return nil
}

// Templates
func (s *SettingsService) ListTemplates(ctx context.Context) ([]models.Template, error) {
	return s.repo.ListTemplates(ctx)
}

func (s *SettingsService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	return s.repo.GetTemplate(ctx, id)
}

func (s *SettingsService) CreateTemplate(ctx context.Context, t *models.Template) error {
	if t.Key == "" {
		return errors.New("template key is required")
	}
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return err
	}
	s.logAudit(ctx, "TEMPLATE", &t.ID, "CREATE", map[string]any{"key": t.Key})
	return nil
}

func (s *SettingsService) UpdateTemplate(ctx context.Context, t *models.Template) error {
	if err := s.repo.UpdateTemplate(ctx, t); err != nil {
		return err
	}
	s.logAudit(ctx, "TEMPLATE", &t.ID, "UPDATE", map[string]any{"subject": t.Subject})
	return nil
}

// Notification Channels
func (s *SettingsService) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	channels, err := s.repo.ListNotificationChannels(ctx)
	if err != nil {
		return nil, err
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
		return errors.New("channel type is required")
	}

	// Encrypt Config
	if len(c.Config) > 0 {
		if _, ok := c.Config["_encrypted"]; !ok {
			configBytes, err := json.Marshal(c.Config)
			if err != nil {
				return err
			}
			encrypted, err := utils.Encrypt(string(configBytes))
			if err != nil {
				return err
			}
			c.Config = map[string]any{"_encrypted": encrypted}
		}
	}

	if err := s.repo.UpdateNotificationChannel(ctx, c); err != nil {
		return err
	}
	s.logAudit(ctx, "NOTIFICATION_CHANNEL", &c.ID, "UPDATE", map[string]any{"provider": c.Provider})
	return nil
}

// System Limits
func (s *SettingsService) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	return s.repo.GetSystemLimits(ctx)
}

func (s *SettingsService) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	if l.APIRateLimit < 0 {
		return errors.New("api rate limit cannot be negative")
	}
	if err := s.repo.UpdateSystemLimits(ctx, l); err != nil {
		return err
	}
	s.logAudit(ctx, "SYSTEM_LIMITS", &l.ID, "UPDATE", map[string]any{"rate_limit": l.APIRateLimit})
	return nil
}
