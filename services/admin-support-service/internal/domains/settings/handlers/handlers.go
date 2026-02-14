package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/admin-support-service/internal/domains/settings/models"
	"github.com/flyfleet/admin-support-service/internal/domains/settings/service"
	"github.com/gin-gonic/gin"
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

func (h *SettingsHandlers) RegisterRoutes(router *gin.RouterGroup) {
	settings := router.Group("/settings")
	{
		settings.GET("/platform", h.GetPlatformSettings)
		settings.PUT("/platform", h.UpdatePlatformSettings)

		settings.GET("/flags", h.ListFeatureFlags)
		settings.POST("/flags", h.CreateFeatureFlag)
		settings.PATCH("/flags/:id/toggle", h.ToggleFeatureFlag)

		settings.GET("/localization", h.GetLocalizationSettings)
		settings.PUT("/localization", h.UpdateLocalizationSettings)

		settings.GET("/templates", h.ListTemplates)
		settings.PUT("/templates/:id", h.UpdateTemplate)

		settings.GET("/channels", h.ListNotificationChannels)
		settings.PUT("/channels", h.UpdateNotificationChannel)

		settings.GET("/limits", h.GetSystemLimits)
		settings.PUT("/limits", h.UpdateSystemLimits)
	}
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
func (h *SettingsHandlers) GetPlatformSettings(c *gin.Context) {
	settings, err := h.service.GetPlatformSettings(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get platform settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdatePlatformSettings(c *gin.Context) {
	var settings models.PlatformSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.service.UpdatePlatformSettings(c.Request.Context(), &settings); err != nil {
		h.logger.Error("Failed to update platform settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// Feature Flags
func (h *SettingsHandlers) ListFeatureFlags(c *gin.Context) {
	flags, err := h.service.ListFeatureFlags(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list feature flags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, flags)
}

func (h *SettingsHandlers) CreateFeatureFlag(c *gin.Context) {
	var flag models.FeatureFlag
	if err := c.ShouldBindJSON(&flag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.service.CreateFeatureFlag(c.Request.Context(), &flag); err != nil {
		h.logger.Error("Failed to create feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusCreated, flag)
}

func (h *SettingsHandlers) ToggleFeatureFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.service.ToggleFeatureFlag(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to toggle feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "toggled"})
}

// Localization
func (h *SettingsHandlers) GetLocalizationSettings(c *gin.Context) {
	settings, err := h.service.GetLocalizationSettings(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get localization settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdateLocalizationSettings(c *gin.Context) {
	var settings models.LocalizationSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.service.UpdateLocalizationSettings(c.Request.Context(), &settings); err != nil {
		h.logger.Error("Failed to update localization settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// Templates
func (h *SettingsHandlers) ListTemplates(c *gin.Context) {
	templates, err := h.service.ListTemplates(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (h *SettingsHandlers) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var template models.Template
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	template.ID = id

	if err := h.service.UpdateTemplate(c.Request.Context(), &template); err != nil {
		h.logger.Error("Failed to update template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, template)
}

// Channels
func (h *SettingsHandlers) ListNotificationChannels(c *gin.Context) {
	channels, err := h.service.ListNotificationChannels(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list notification channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, channels)
}

func (h *SettingsHandlers) UpdateNotificationChannel(c *gin.Context) {
	var channel models.NotificationChannel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.service.UpdateNotificationChannel(c.Request.Context(), &channel); err != nil {
		h.logger.Error("Failed to update notification channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, channel)
}

// Limits
func (h *SettingsHandlers) GetSystemLimits(c *gin.Context) {
	limits, err := h.service.GetSystemLimits(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get system limits", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, limits)
}

func (h *SettingsHandlers) UpdateSystemLimits(c *gin.Context) {
	var limits models.SystemLimit
	if err := c.ShouldBindJSON(&limits); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.service.UpdateSystemLimits(c.Request.Context(), &limits); err != nil {
		h.logger.Error("Failed to update system limits", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, limits)
}
