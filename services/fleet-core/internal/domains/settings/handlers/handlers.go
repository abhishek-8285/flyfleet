package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/settings/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/service"
	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DTOs for request payloads (separate from models to prevent mass assignment)

type platformSettingsDTO struct {
	SystemName      string     `json:"system_name"`
	SupportEmail    string     `json:"support_email"`
	DefaultTimezone string     `json:"default_timezone"`
	DefaultCurrency string     `json:"default_currency"`
	DefaultLanguage string     `json:"default_language"`
	WhiteLabel      bool       `json:"white_label_enabled"`
	SMTPConfigID    *uuid.UUID `json:"smtp_config_id,omitempty"`
}

type featureFlagDTO struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Scope       string   `json:"scope"` // GLOBAL, ORG
	Enabled     bool     `json:"enabled"`
	Tags        []string `json:"tags"`
}

type localizationSettingsDTO struct {
	LanguageCode string `json:"language_code"`
	IsActive     bool   `json:"is_active"`
	IsDefault    bool   `json:"is_default"`
}

type templateDTO struct {
	Key       string   `json:"key"`
	Channel   string   `json:"channel"` // EMAIL, SMS
	Subject   string   `json:"subject"`
	Body      string   `json:"body"`
	Variables []string `json:"variables"`
}

type notificationChannelDTO struct {
	ChannelType string         `json:"channel_type"`
	Provider    string         `json:"provider"`
	Config      map[string]any `json:"config"`
	Enabled     bool           `json:"enabled"`
}

type systemLimitDTO struct {
	APIRateLimit      int    `json:"api_rate_limit"`
	BurstLimit        int    `json:"burst_limit"`
	AuditRetention    string `json:"audit_retention"`
	ActivityRetention string `json:"activity_retention"`
}

type settingsImportRequestDTO struct {
	PlatformSettings     *platformSettingsDTO      `json:"platform_settings"`
	FeatureFlags         []featureFlagDTO          `json:"feature_flags"`
	LocalizationSettings []localizationSettingsDTO `json:"localization_settings"`
	Templates            []templateDTO             `json:"templates"`
	NotificationChannels []notificationChannelDTO  `json:"notification_channels"`
	SystemLimits         *systemLimitDTO           `json:"system_limits"`
	ReplaceExisting      bool                      `json:"replace_existing"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type SettingsHandlers struct {
	service *service.SettingsService
	logger  *zap.Logger
}

func NewSettingsHandlers(service *service.SettingsService, logger *zap.Logger) *SettingsHandlers {
	return &SettingsHandlers{service: service, logger: logger}
}

// validatePlatformSettings validates platform settings input
func (h *SettingsHandlers) validatePlatformSettings(settings *models.PlatformSettings) error {
	if settings.SystemName == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "system name is required", Field: "system_name"}
	}
	if len(settings.SystemName) > 255 {
		return &service.SettingsError{Code: "INVALID_LENGTH", Message: "system name must be less than 255 characters", Field: "system_name"}
	}
	if settings.SupportEmail == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "support email is required", Field: "support_email"}
	}
	if !emailRegex.MatchString(settings.SupportEmail) {
		return &service.SettingsError{Code: "INVALID_FORMAT", Message: "invalid email format", Field: "support_email"}
	}
	if settings.DefaultTimezone == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "default timezone is required", Field: "default_timezone"}
	}
	if _, err := time.LoadLocation(settings.DefaultTimezone); err != nil {
		return &service.SettingsError{Code: "INVALID_TIMEZONE", Message: "invalid timezone", Field: "default_timezone"}
	}
	if settings.DefaultCurrency == "" || len(settings.DefaultCurrency) != 3 {
		return &service.SettingsError{Code: "INVALID_FORMAT", Message: "currency must be 3 characters", Field: "default_currency"}
	}
	if settings.DefaultLanguage == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "default language is required", Field: "default_language"}
	}
	return nil
}

// validateFeatureFlag validates feature flag input
func (h *SettingsHandlers) validateFeatureFlag(flag *models.FeatureFlag) error {
	if flag.Key == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "flag key is required", Field: "key"}
	}
	if len(flag.Key) > 100 {
		return &service.SettingsError{Code: "INVALID_LENGTH", Message: "key must be less than 100 characters", Field: "key"}
	}
	if flag.Scope != "" && flag.Scope != "GLOBAL" && flag.Scope != "ORG" {
		return &service.SettingsError{Code: "INVALID_SCOPE", Message: "scope must be GLOBAL or ORG", Field: "scope"}
	}
	return nil
}

// validateTemplate validates template input
func (h *SettingsHandlers) validateTemplate(template *models.Template) error {
	if template.Key == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "template key is required", Field: "key"}
	}
	if len(template.Key) > 100 {
		return &service.SettingsError{Code: "INVALID_LENGTH", Message: "key must be less than 100 characters", Field: "key"}
	}
	if template.Channel == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "channel is required", Field: "channel"}
	}
	if template.Channel != "EMAIL" && template.Channel != "SMS" {
		return &service.SettingsError{Code: "INVALID_CHANNEL", Message: "channel must be EMAIL or SMS", Field: "channel"}
	}
	if template.Body == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "body is required", Field: "body"}
	}
	return nil
}

// validateNotificationChannel validates notification channel input
func (h *SettingsHandlers) validateNotificationChannel(channel *models.NotificationChannel) error {
	if channel.ChannelType == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "channel type is required", Field: "channel_type"}
	}
	if channel.ChannelType != "EMAIL" && channel.ChannelType != "SMS" && channel.ChannelType != "SLACK" {
		return &service.SettingsError{Code: "INVALID_CHANNEL_TYPE", Message: "invalid channel type", Field: "channel_type"}
	}
	if channel.Provider == "" {
		return &service.SettingsError{Code: "REQUIRED_FIELD", Message: "provider is required", Field: "provider"}
	}
	return nil
}

// validateSystemLimits validates system limits input
func (h *SettingsHandlers) validateSystemLimits(limits *models.SystemLimit) error {
	if limits.APIRateLimit < 0 {
		return &service.SettingsError{Code: "INVALID_VALUE", Message: "rate limit cannot be negative", Field: "api_rate_limit"}
	}
	if limits.BurstLimit < 0 {
		return &service.SettingsError{Code: "INVALID_VALUE", Message: "burst limit cannot be negative", Field: "burst_limit"}
	}
	return nil
}

// requireAdminRole ensures the user has admin or super_admin role
func (h *SettingsHandlers) requireAdminRole(w http.ResponseWriter, r *http.Request) bool {
	role, ok := r.Context().Value(ctxkeys.RoleKey).(string)
	if !ok || role == "" {
		h.logger.Warn("Access denied: no role found in context",
			zap.String("path", r.URL.Path))
		http.Error(w, "Forbidden: No Role Assigned", http.StatusForbidden)
		return false
	}

	allowedRoles := map[string]bool{
		"super_admin": true,
		"org_admin":   true,
		"admin":       true,
	}

	if !allowedRoles[role] {
		h.logger.Warn("Access denied: insufficient role for settings operation",
			zap.String("user_role", role),
			zap.String("path", r.URL.Path))
		http.Error(w, "Forbidden: Insufficient Permissions - Admin Role Required", http.StatusForbidden)
		return false
	}

	return true
}

// Helper for JSON responses
func (h *SettingsHandlers) respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			h.logger.Error("Failed to encode response", zap.Error(err))
		}
	}
}

func (h *SettingsHandlers) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	validationHints := []string{
		"required",
		"invalid",
		"cannot be negative",
		"duplicate",
		"at least one",
	}
	for _, hint := range validationHints {
		if strings.Contains(msg, hint) {
			return true
		}
	}
	return false
}

func (h *SettingsHandlers) handleServiceError(w http.ResponseWriter, err error, fallbackMessage string) {
	if err == nil {
		return
	}
	if isValidationError(err) {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondError(w, http.StatusInternalServerError, fallbackMessage)
}

// Platform Settings
func (h *SettingsHandlers) GetSettingsExport(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for settings export
	if !h.requireAdminRole(w, r) {
		return
	}

	exportPayload, err := h.service.GetSettingsExport(r.Context())
	if err != nil {
		h.logger.Error("Failed to build settings export", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, exportPayload)
}

func (h *SettingsHandlers) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for audit logs
	if !h.requireAdminRole(w, r) {
		return
	}

	limit := 50
	offset := 0

	if limitRaw := strings.TrimSpace(r.URL.Query().Get("limit")); limitRaw != "" {
		parsed, err := strconv.Atoi(limitRaw)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		limit = parsed
	}
	if offsetRaw := strings.TrimSpace(r.URL.Query().Get("offset")); offsetRaw != "" {
		parsed, err := strconv.Atoi(offsetRaw)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "Invalid offset parameter")
			return
		}
		offset = parsed
	}

	logs, err := h.service.ListAuditLogs(r.Context(), models.AuditLogQuery{
		Limit:      limit,
		Offset:     offset,
		EntityType: r.URL.Query().Get("entity_type"),
		Action:     r.URL.Query().Get("action"),
	})
	if err != nil {
		h.logger.Error("Failed to list settings audit logs", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]any{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SettingsHandlers) ImportSettings(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for settings import
	if !h.requireAdminRole(w, r) {
		return
	}

	var dto settingsImportRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Map DTO to model
	req := &models.SettingsImportRequest{
		PlatformSettings:     mapPlatformSettings(dto.PlatformSettings),
		FeatureFlags:         mapFeatureFlags(dto.FeatureFlags),
		LocalizationSettings: mapLocalizationSettings(dto.LocalizationSettings),
		Templates:            mapTemplates(dto.Templates),
		NotificationChannels: mapNotificationChannels(dto.NotificationChannels),
		SystemLimits:         mapSystemLimit(dto.SystemLimits),
		ReplaceExisting:      dto.ReplaceExisting,
	}

	result, err := h.service.ImportSettings(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to import settings", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *SettingsHandlers) GetPlatformSettings(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for platform settings (exposes sensitive configuration)
	if !h.requireAdminRole(w, r) {
		return
	}

	settings, err := h.service.GetPlatformSettings(r.Context())
	if err != nil {
		h.logger.Error("Failed to get platform settings", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdatePlatformSettings(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for platform settings modification
	if !h.requireAdminRole(w, r) {
		return
	}

	var dto platformSettingsDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Map DTO to model
	settings := &models.PlatformSettings{
		SystemName:      dto.SystemName,
		SupportEmail:    dto.SupportEmail,
		DefaultTimezone: dto.DefaultTimezone,
		DefaultCurrency: dto.DefaultCurrency,
		DefaultLanguage: dto.DefaultLanguage,
		WhiteLabel:      dto.WhiteLabel,
		SMTPConfigID:    dto.SMTPConfigID,
	}

	if err := h.validatePlatformSettings(settings); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdatePlatformSettings(r.Context(), settings); err != nil {
		h.logger.Error("Failed to update platform settings", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

// Feature Flags
func (h *SettingsHandlers) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for feature flag visibility (could expose internal roadmap)
	if !h.requireAdminRole(w, r) {
		return
	}

	flags, err := h.service.ListFeatureFlags(r.Context())
	if err != nil {
		h.logger.Error("Failed to list feature flags", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, flags)
}

func (h *SettingsHandlers) CreateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for feature flag management
	if !h.requireAdminRole(w, r) {
		return
	}

	var dto featureFlagDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Map DTO to model
	flag := &models.FeatureFlag{
		Key:         dto.Key,
		Description: dto.Description,
		Scope:       dto.Scope,
		Enabled:     dto.Enabled,
		Tags:        dto.Tags,
	}

	if err := h.validateFeatureFlag(flag); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.CreateFeatureFlag(r.Context(), flag); err != nil {
		h.logger.Error("Failed to create feature flag", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusCreated, flag)
}

func (h *SettingsHandlers) ToggleFeatureFlag(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for feature flag modification
	if !h.requireAdminRole(w, r) {
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/settings/flags/")
	idStr = strings.TrimSuffix(idStr, "/toggle")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.service.ToggleFeatureFlag(r.Context(), id); err != nil {
		h.logger.Error("Failed to toggle feature flag", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "toggled"})
}

// Localization
func (h *SettingsHandlers) GetLocalizationSettings(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for localization settings (exposes global configuration)
	if !h.requireAdminRole(w, r) {
		return
	}

	settings, err := h.service.GetLocalizationSettings(r.Context())
	if err != nil {
		h.logger.Error("Failed to get localization settings", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdateLocalizationSettings(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for localization settings (global configuration)
	if !h.requireAdminRole(w, r) {
		return
	}

	var dto localizationSettingsDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Map DTO to model
	settings := &models.LocalizationSettings{
		LanguageCode: dto.LanguageCode,
		IsActive:     dto.IsActive,
		IsDefault:    dto.IsDefault,
	}

	if err := h.service.UpdateLocalizationSettings(r.Context(), settings); err != nil {
		h.logger.Error("Failed to update localization settings", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

// Templates
func (h *SettingsHandlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for template listing (exposes notification templates)
	if !h.requireAdminRole(w, r) {
		return
	}

	templates, err := h.service.ListTemplates(r.Context())
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, templates)
}

func (h *SettingsHandlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for template modification
	if !h.requireAdminRole(w, r) {
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/settings/templates/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var dto templateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// Map DTO to model
	template := &models.Template{
		Key:       dto.Key,
		Channel:   dto.Channel,
		Subject:   dto.Subject,
		Body:      dto.Body,
		Variables: dto.Variables,
	}
	template.ID = id

	if err := h.validateTemplate(template); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateTemplate(r.Context(), template); err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, template)
}

// Channels
func (h *SettingsHandlers) ListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for notification channel listing (exposes channel configurations)
	if !h.requireAdminRole(w, r) {
		return
	}

	channels, err := h.service.ListNotificationChannels(r.Context())
	if err != nil {
		h.logger.Error("Failed to list notification channels", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, channels)
}

func (h *SettingsHandlers) UpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for notification channel configuration
	if !h.requireAdminRole(w, r) {
		return
	}

	var dto notificationChannelDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Map DTO to model
	channel := &models.NotificationChannel{
		ChannelType: dto.ChannelType,
		Provider:    dto.Provider,
		Config:      dto.Config,
		Enabled:     dto.Enabled,
	}
	if err := h.validateNotificationChannel(channel); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateNotificationChannel(r.Context(), channel); err != nil {
		h.logger.Error("Failed to update notification channel", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, channel)
}

// Limits
func (h *SettingsHandlers) GetSystemLimits(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for system limits (could expose DoS protection settings)
	if !h.requireAdminRole(w, r) {
		return
	}

	limits, err := h.service.GetSystemLimits(r.Context())
	if err != nil {
		h.logger.Error("Failed to get system limits", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, limits)
}

func (h *SettingsHandlers) UpdateSystemLimits(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for system limits modification
	if !h.requireAdminRole(w, r) {
		return
	}

	var dto systemLimitDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Map DTO to model
	limits := &models.SystemLimit{
		APIRateLimit:      dto.APIRateLimit,
		BurstLimit:        dto.BurstLimit,
		AuditRetention:    dto.AuditRetention,
		ActivityRetention: dto.ActivityRetention,
	}
	if err := h.validateSystemLimits(limits); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateSystemLimits(r.Context(), limits); err != nil {
		h.logger.Error("Failed to update system limits", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, limits)
}

// CleanupAuditLogs removes old audit logs based on configured retention policy
func (h *SettingsHandlers) CleanupAuditLogs(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require admin role for audit log cleanup
	if !h.requireAdminRole(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	deleted, err := h.service.CleanupOldAuditLogs(r.Context())
	if err != nil {
		h.logger.Error("Failed to cleanup audit logs", zap.Error(err))
		h.handleServiceError(w, err, "Internal Server Error")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":       "Audit logs cleaned up successfully",
		"deleted_count": deleted,
	})
}

// Mapping functions for DTO to model conversion

func mapPlatformSettings(dto *platformSettingsDTO) *models.PlatformSettings {
	if dto == nil {
		return nil
	}
	return &models.PlatformSettings{
		SystemName:      dto.SystemName,
		SupportEmail:    dto.SupportEmail,
		DefaultTimezone: dto.DefaultTimezone,
		DefaultCurrency: dto.DefaultCurrency,
		DefaultLanguage: dto.DefaultLanguage,
		WhiteLabel:      dto.WhiteLabel,
		SMTPConfigID:    dto.SMTPConfigID,
	}
}

func mapFeatureFlags(dtos []featureFlagDTO) []models.FeatureFlag {
	result := make([]models.FeatureFlag, len(dtos))
	for i, dto := range dtos {
		result[i] = models.FeatureFlag{
			Key:         dto.Key,
			Description: dto.Description,
			Scope:       dto.Scope,
			Enabled:     dto.Enabled,
			Tags:        dto.Tags,
		}
	}
	return result
}

func mapLocalizationSettings(dtos []localizationSettingsDTO) []models.LocalizationSettings {
	result := make([]models.LocalizationSettings, len(dtos))
	for i, dto := range dtos {
		result[i] = models.LocalizationSettings{
			LanguageCode: dto.LanguageCode,
			IsActive:     dto.IsActive,
			IsDefault:    dto.IsDefault,
		}
	}
	return result
}

func mapTemplates(dtos []templateDTO) []models.Template {
	result := make([]models.Template, len(dtos))
	for i, dto := range dtos {
		result[i] = models.Template{
			Key:       dto.Key,
			Channel:   dto.Channel,
			Subject:   dto.Subject,
			Body:      dto.Body,
			Variables: dto.Variables,
		}
	}
	return result
}

func mapNotificationChannels(dtos []notificationChannelDTO) []models.NotificationChannel {
	result := make([]models.NotificationChannel, len(dtos))
	for i, dto := range dtos {
		result[i] = models.NotificationChannel{
			ChannelType: dto.ChannelType,
			Provider:    dto.Provider,
			Config:      dto.Config,
			Enabled:     dto.Enabled,
		}
	}
	return result
}

func mapSystemLimit(dto *systemLimitDTO) *models.SystemLimit {
	if dto == nil {
		return nil
	}
	return &models.SystemLimit{
		APIRateLimit:      dto.APIRateLimit,
		BurstLimit:        dto.BurstLimit,
		AuditRetention:    dto.AuditRetention,
		ActivityRetention: dto.ActivityRetention,
	}
}
