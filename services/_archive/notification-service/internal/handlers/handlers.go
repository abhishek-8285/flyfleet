package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/notification-service/internal/dto"
	"github.com/flyfleet/notification-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationHandlers struct {
	service *service.NotificationService
	logger  *zap.Logger
}

func NewNotificationHandlers(service *service.NotificationService, logger *zap.Logger) *NotificationHandlers {
	return &NotificationHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *NotificationHandlers) bindStrictJSON(c *gin.Context, obj any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return binding.Validator.ValidateStruct(obj)
}

func (h *NotificationHandlers) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	channel := c.Query("channel")
	search := c.Query("search")

	logs, total, err := h.service.ListLogs(c.Request.Context(), page, limit, status, channel, search)
	if err != nil {
		h.logger.Error("Failed to list logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list logs"})
		return
	}

	c.JSON(http.StatusOK, dto.ListLogsResponse{
		Data:  logs,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *NotificationHandlers) GetLogDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}

	log, err := h.service.GetLogDetail(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get log detail", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get log detail"})
		return
	}

	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

func (h *NotificationHandlers) SendNotification(c *gin.Context) {
	var req dto.SendNotificationRequest
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	id, err := h.service.SendNotification(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to send notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "Notification queued successfully",
	})
}

func (h *NotificationHandlers) Resend(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}

	var req dto.ResendRequest
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	newID, err := h.service.ResendNotification(c.Request.Context(), id, req.Reason, req.Force)
	if err != nil {
		h.logger.Error("Failed to resend notification", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resend notification"})
		return
	}

	c.JSON(http.StatusOK, dto.ResendResponse{
		NewNotificationID: newID,
		Message:           "Resend initiated successfully",
	})
}

func (h *NotificationHandlers) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get settings"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *NotificationHandlers) UpdateSettings(c *gin.Context) {
	var req dto.SettingsUpdateRequest
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	if err := h.service.UpdateSettings(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *NotificationHandlers) ListPreferences(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	prefs, total, err := h.service.ListPreferences(c.Request.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("Failed to list preferences", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list preferences"})
		return
	}

	c.JSON(http.StatusOK, dto.PreferenceListResponse{
		Items: prefs,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *NotificationHandlers) GetUserPreference(c *gin.Context) {
	idStr := c.Param("id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	pref, err := h.service.GetUserPreference(c.Request.Context(), userId)
	if err != nil {
		h.logger.Error("Failed to get user preference", zap.Error(err), zap.String("userId", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user preference"})
		return
	}

	if pref == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "preference not found"})
		return
	}

	c.JSON(http.StatusOK, pref)
}

func (h *NotificationHandlers) UpdateUserPreference(c *gin.Context) {
	idStr := c.Param("id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.PreferenceUpdateRequest
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	// In a real system, we'd get the admin ID from JWT context
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Placeholder
	
	if err := h.service.UpdateUserPreference(c.Request.Context(), userId, req, adminID); err != nil {
		h.logger.Error("Failed to update user preference", zap.Error(err), zap.String("userId", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user preference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *NotificationHandlers) GetGlobalDefaults(c *gin.Context) {
	defaults, err := h.service.GetGlobalDefaults(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get global defaults", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get global defaults"})
		return
	}
	c.JSON(http.StatusOK, defaults)
}

func (h *NotificationHandlers) UpdateGlobalDefaults(c *gin.Context) {
	var req dto.GlobalDefaultUpdateRequest
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Placeholder

	if err := h.service.UpdateGlobalDefaults(c.Request.Context(), req, adminID); err != nil {
		h.logger.Error("Failed to update global defaults", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update global defaults"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *NotificationHandlers) GetPreferenceAudit(c *gin.Context) {
	idStr := c.Param("id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	audits, err := h.service.GetPreferenceAudit(c.Request.Context(), userId)
	if err != nil {
		h.logger.Error("Failed to get preference audit", zap.Error(err), zap.String("userId", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get preference audit"})
		return
	}

	c.JSON(http.StatusOK, audits)
}
