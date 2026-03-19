package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/dto"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationHandlers struct {
	service   *service.NotificationService
	logger    *zap.Logger
	validator *validator.Validate
}

func NewNotificationHandlers(service *service.NotificationService, logger *zap.Logger) *NotificationHandlers {
	return &NotificationHandlers{
		service:   service,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *NotificationHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/logs", h.ListLogs)
		r.Get("/logs/{id}", h.GetLogDetail)
		r.Post("/send", h.SendNotification)
		r.Post("/resend/{id}", h.Resend)
		
		r.Get("/settings", h.GetSettings)
		r.Put("/settings", h.UpdateSettings)
		
		r.Get("/preferences", h.ListPreferences)
		r.Get("/preferences/{id}", h.GetUserPreference)
		r.Put("/preferences/{id}", h.UpdateUserPreference)
		r.Get("/preferences/{id}/audit", h.GetPreferenceAudit)
		
		r.Get("/defaults", h.GetGlobalDefaults)
		r.Put("/defaults", h.UpdateGlobalDefaults)
	})
}

func (h *NotificationHandlers) bindStrictJSON(r *http.Request, obj any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return h.validator.Struct(obj)
}

func (h *NotificationHandlers) ListLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 {
		limit = 20
	}
	status := query.Get("status")
	channel := query.Get("channel")
	search := query.Get("search")

	logs, total, err := h.service.ListLogs(r.Context(), page, limit, status, channel, search)
	if err != nil {
		h.logger.Error("Failed to list logs", zap.Error(err))
		h.errorJSON(w, "failed to list logs", http.StatusInternalServerError)
		return
	}

	h.json(w, dto.ListLogsResponse{
		Data:  logs,
		Total: total,
		Page:  page,
		Limit: limit,
	}, http.StatusOK)
}

func (h *NotificationHandlers) GetLogDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid log id", http.StatusBadRequest)
		return
	}

	log, err := h.service.GetLogDetail(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get log detail", zap.Error(err), zap.String("id", idStr))
		h.errorJSON(w, "failed to get log detail", http.StatusInternalServerError)
		return
	}

	if log == nil {
		h.errorJSON(w, "log not found", http.StatusNotFound)
		return
	}

	h.json(w, log, http.StatusOK)
}

func (h *NotificationHandlers) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req dto.SendNotificationRequest
	if err := h.bindStrictJSON(r, &req); err != nil {
		h.errorJSON(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.SendNotification(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to send notification", zap.Error(err))
		h.errorJSON(w, "failed to send notification", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"id":      id,
		"message": "Notification queued successfully",
	}, http.StatusOK)
}

func (h *NotificationHandlers) Resend(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid log id", http.StatusBadRequest)
		return
	}

	var req dto.ResendRequest
	if err := h.bindStrictJSON(r, &req); err != nil {
		h.errorJSON(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	newID, err := h.service.ResendNotification(r.Context(), id, req.Reason, req.Force)
	if err != nil {
		h.logger.Error("Failed to resend notification", zap.Error(err), zap.String("id", idStr))
		h.errorJSON(w, "failed to resend notification", http.StatusInternalServerError)
		return
	}

	h.json(w, dto.ResendResponse{
		NewNotificationID: newID,
		Message:           "Resend initiated successfully",
	}, http.StatusOK)
}

func (h *NotificationHandlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		h.errorJSON(w, "failed to get settings", http.StatusInternalServerError)
		return
	}
	h.json(w, settings, http.StatusOK)
}

func (h *NotificationHandlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req dto.SettingsUpdateRequest
	if err := h.bindStrictJSON(r, &req); err != nil {
		h.errorJSON(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSettings(r.Context(), req); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		h.errorJSON(w, "failed to update settings", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *NotificationHandlers) ListPreferences(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 {
		limit = 20
	}
	search := query.Get("search")

	prefs, total, err := h.service.ListPreferences(r.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("Failed to list preferences", zap.Error(err))
		h.errorJSON(w, "failed to list preferences", http.StatusInternalServerError)
		return
	}

	h.json(w, dto.PreferenceListResponse{
		Items: prefs,
		Total: total,
		Page:  page,
		Limit: limit,
	}, http.StatusOK)
}

func (h *NotificationHandlers) GetUserPreference(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid user id", http.StatusBadRequest)
		return
	}

	pref, err := h.service.GetUserPreference(r.Context(), userId)
	if err != nil {
		h.logger.Error("Failed to get user preference", zap.Error(err), zap.String("userId", idStr))
		h.errorJSON(w, "failed to get user preference", http.StatusInternalServerError)
		return
	}

	if pref == nil {
		h.errorJSON(w, "preference not found", http.StatusNotFound)
		return
	}

	h.json(w, pref, http.StatusOK)
}

func (h *NotificationHandlers) UpdateUserPreference(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req dto.PreferenceUpdateRequest
	if err := h.bindStrictJSON(r, &req); err != nil {
		h.errorJSON(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// In a real system, we'd get the admin ID from JWT context
	adminID := uuid.Nil // Placeholder
	
	if err := h.service.UpdateUserPreference(r.Context(), userId, req, adminID); err != nil {
		h.logger.Error("Failed to update user preference", zap.Error(err), zap.String("userId", idStr))
		h.errorJSON(w, "failed to update user preference", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *NotificationHandlers) GetGlobalDefaults(w http.ResponseWriter, r *http.Request) {
	defaults, err := h.service.GetGlobalDefaults(r.Context())
	if err != nil {
		h.logger.Error("Failed to get global defaults", zap.Error(err))
		h.errorJSON(w, "failed to get global defaults", http.StatusInternalServerError)
		return
	}
	h.json(w, defaults, http.StatusOK)
}

func (h *NotificationHandlers) UpdateGlobalDefaults(w http.ResponseWriter, r *http.Request) {
	var req dto.GlobalDefaultUpdateRequest
	if err := h.bindStrictJSON(r, &req); err != nil {
		h.errorJSON(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	adminID := uuid.Nil // Placeholder

	if err := h.service.UpdateGlobalDefaults(r.Context(), req, adminID); err != nil {
		h.logger.Error("Failed to update global defaults", zap.Error(err))
		h.errorJSON(w, "failed to update global defaults", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "success"}, http.StatusOK)
}

func (h *NotificationHandlers) GetPreferenceAudit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid user id", http.StatusBadRequest)
		return
	}

	audits, err := h.service.GetPreferenceAudit(r.Context(), userId)
	if err != nil {
		h.logger.Error("Failed to get preference audit", zap.Error(err), zap.String("userId", idStr))
		h.errorJSON(w, "failed to get preference audit", http.StatusInternalServerError)
		return
	}

	h.json(w, audits, http.StatusOK)
}

// Helpers
func (h *NotificationHandlers) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *NotificationHandlers) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}
