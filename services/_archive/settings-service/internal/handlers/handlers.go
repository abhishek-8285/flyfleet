package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyfleet/settings-service/internal/models"
	"github.com/flyfleet/settings-service/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SettingsHandlers struct {
	service *service.SettingsService
	logger  *zap.Logger
}

func NewSettingsHandlers(service *service.SettingsService, logger *zap.Logger) *SettingsHandlers {
	return &SettingsHandlers{service: service, logger: logger}
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

// Platform Settings
func (h *SettingsHandlers) GetPlatformSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetPlatformSettings(r.Context())
	if err != nil {
		h.logger.Error("Failed to get platform settings", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdatePlatformSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.PlatformSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.UpdatePlatformSettings(r.Context(), &settings); err != nil {
		h.logger.Error("Failed to update platform settings", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

// Feature Flags
func (h *SettingsHandlers) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.service.ListFeatureFlags(r.Context())
	if err != nil {
		h.logger.Error("Failed to list feature flags", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, flags)
}

func (h *SettingsHandlers) CreateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	var flag models.FeatureFlag
	if err := json.NewDecoder(r.Body).Decode(&flag); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.CreateFeatureFlag(r.Context(), &flag); err != nil {
		h.logger.Error("Failed to create feature flag", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusCreated, flag)
}

func (h *SettingsHandlers) ToggleFeatureFlag(w http.ResponseWriter, r *http.Request) {
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
	settings, err := h.service.GetLocalizationSettings(r.Context())
	if err != nil {
		h.logger.Error("Failed to get localization settings", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdateLocalizationSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.LocalizationSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.UpdateLocalizationSettings(r.Context(), &settings); err != nil {
		h.logger.Error("Failed to update localization settings", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, settings)
}

// Templates
func (h *SettingsHandlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.ListTemplates(r.Context())
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, templates)
}

func (h *SettingsHandlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/settings/templates/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var template models.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	template.ID = id

	if err := h.service.UpdateTemplate(r.Context(), &template); err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, template)
}

// Channels
func (h *SettingsHandlers) ListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.service.ListNotificationChannels(r.Context())
	if err != nil {
		h.logger.Error("Failed to list notification channels", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, channels)
}

func (h *SettingsHandlers) UpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	var channel models.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.UpdateNotificationChannel(r.Context(), &channel); err != nil {
		h.logger.Error("Failed to update notification channel", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, channel)
}

// Limits
func (h *SettingsHandlers) GetSystemLimits(w http.ResponseWriter, r *http.Request) {
	limits, err := h.service.GetSystemLimits(r.Context())
	if err != nil {
		h.logger.Error("Failed to get system limits", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, limits)
}

func (h *SettingsHandlers) UpdateSystemLimits(w http.ResponseWriter, r *http.Request) {
	var limits models.SystemLimit
	if err := json.NewDecoder(r.Body).Decode(&limits); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.service.UpdateSystemLimits(r.Context(), &limits); err != nil {
		h.logger.Error("Failed to update system limits", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondJSON(w, http.StatusOK, limits)
}
