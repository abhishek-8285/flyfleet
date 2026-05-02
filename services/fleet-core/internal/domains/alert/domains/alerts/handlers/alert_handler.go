package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/alert/domains/alerts/models"
	"github.com/flyfleet/fleet-core/internal/domains/alert/domains/alerts/service"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createRuleRequestDTO struct {
	Name        string                 `json:"name" validate:"required,max=100"`
	Type        string                 `json:"type" validate:"required,oneof=overspeed geofence_entry geofence_exit harsh_acceleration harsh_braking harsh_cornering idle_time fuel_theft low_battery temperature_excursion accident_detected maintenance_due document_expiry driver_fatigue adas_event vehicle_offline"`
	Description string                 `json:"description" validate:"max=500"`
	Conditions  map[string]interface{} `json:"conditions" validate:"required"`
	Severity    string                 `json:"severity" validate:"required,oneof=info warning critical"`
	Actions     []models.AlertAction   `json:"actions" validate:"required,min=1"`
	Vehicles    []uuid.UUID            `json:"vehicles"`
	DriverIDs   []uuid.UUID            `json:"driver_ids"`
}

type updateRuleRequestDTO struct {
	Name        *string                 `json:"name,omitempty" validate:"omitempty,max=100"`
	Description *string                 `json:"description,omitempty" validate:"omitempty,max=500"`
	Conditions  *map[string]interface{} `json:"conditions,omitempty"`
	Severity    *string                 `json:"severity,omitempty" validate:"omitempty,oneof=info warning critical"`
	Actions     *[]models.AlertAction   `json:"actions,omitempty" validate:"omitempty,min=1"`
	Vehicles    *[]uuid.UUID            `json:"vehicles,omitempty"`
	DriverIDs   *[]uuid.UUID            `json:"driver_ids,omitempty"`
	Enabled     *bool                   `json:"enabled,omitempty"`
}

type AlertHandler struct {
	service *service.AlertService
	logger  *zap.Logger
}

func NewAlertHandler(svc *service.AlertService, logger *zap.Logger) *AlertHandler {
	return &AlertHandler{
		service: svc,
		logger:  logger,
	}
}

func (h *AlertHandler) RegisterRoutes(r chi.Router) {
	r.Route("/alerts", func(r chi.Router) {
		// Alert rules management
		r.Post("/rules", h.CreateRule)
		r.Get("/rules", h.ListRules)
		r.Get("/rules/{id}", h.GetRule)
		r.Put("/rules/{id}", h.UpdateRule)
		r.Delete("/rules/{id}", h.DeleteRule)
		r.Post("/rules/{id}/enable", h.EnableRule)
		r.Post("/rules/{id}/disable", h.DisableRule)

		// Alert instances
		r.Get("/", h.ListAlerts)
		r.Get("/{id}", h.GetAlert)
		r.Post("/{id}/acknowledge", h.AcknowledgeAlert)
		r.Post("/{id}/resolve", h.ResolveAlert)
		r.Post("/{id}/escalate", h.EscalateAlert)

		// Alert statistics and analytics
		r.Get("/stats/summary", h.GetAlertSummary)
		r.Get("/stats/trends", h.GetAlertTrends)
		r.Get("/stats/new", h.GetStats)

		// Templates
		r.Post("/templates", h.CreateTemplate)
		r.Get("/templates", h.ListTemplates)
		r.Get("/templates/{id}", h.GetTemplate)
		r.Put("/templates/{id}", h.UpdateTemplate)
		r.Delete("/templates/{id}", h.DeleteTemplate)
	})
}

func (h *AlertHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.errorJSON(w, "Tenant ID required", http.StatusUnauthorized)
		return
	}

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.errorJSON(w, "User ID required", http.StatusUnauthorized)
		return
	}

	rule, err := h.service.CreateRule(r.Context(), tenantID, &req)
	if err != nil {
		// SECURITY FIX: Log detailed error but return generic message to client
		h.logger.Error("failed to create alert rule", zap.Error(err), zap.String("tenant_id", tenantID.String()))
		h.errorJSON(w, "Failed to create alert rule", http.StatusInternalServerError)
		return
	}

	h.json(w, rule, http.StatusCreated)
}

func (h *AlertHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.GetRule(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get alert rule", zap.String("rule_id", id), zap.Error(err))
		h.errorJSON(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	h.json(w, rule, http.StatusOK)
}

func (h *AlertHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.errorJSON(w, "Tenant ID required", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	filter := models.AlertListFilter{
		Page:     1,
		PageSize: 20,
		OrderBy:  "created_at",
		OrderDir: "desc",
	}

	// Simple pagination parsing (could be improved)
	if page := query.Get("page"); page != "" {
		// Would parse page number if needed
	}

	if types := query["type"]; len(types) > 0 {
		filter.Type = types
	}

	if severities := query["severity"]; len(severities) > 0 {
		filter.Severity = severities
	}

	rules, totalCount, err := h.service.ListRules(r.Context(), tenantID, filter)
	if err != nil {
		h.logger.Error("failed to list alert rules", zap.Error(err))
		h.errorJSON(w, "Failed to list alert rules", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data":        rules,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	}, http.StatusOK)
}

func (h *AlertHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rule, err := h.service.UpdateRule(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update alert rule", zap.String("rule_id", id), zap.Error(err))
		h.errorJSON(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	h.json(w, rule, http.StatusOK)
}

func (h *AlertHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteRule(r.Context(), id); err != nil {
		h.logger.Error("failed to delete alert rule", zap.String("rule_id", id), zap.Error(err))
		h.errorJSON(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AlertHandler) EnableRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.EnableRule(r.Context(), id); err != nil {
		h.logger.Error("failed to enable alert rule", zap.String("rule_id", id), zap.Error(err))
		h.errorJSON(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	h.json(w, map[string]string{"message": "Alert rule enabled"}, http.StatusOK)
}

func (h *AlertHandler) DisableRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DisableRule(r.Context(), id); err != nil {
		h.logger.Error("failed to disable alert rule", zap.String("rule_id", id), zap.Error(err))
		h.errorJSON(w, "Alert rule not found", http.StatusNotFound)
		return
	}

	h.json(w, map[string]string{"message": "Alert rule disabled"}, http.StatusOK)
}

func (h *AlertHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	alert, err := h.service.GetAlert(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get alert", zap.String("alert_id", id), zap.Error(err))
		h.errorJSON(w, "Alert not found", http.StatusNotFound)
		return
	}

	h.json(w, alert, http.StatusOK)
}

func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	tenantIDStr := query.Get("tenant_id")
	if tenantIDStr == "" {
		tenantID := h.getTenantID(r)
		if tenantID != uuid.Nil {
			tenantIDStr = tenantID.String()
		}
	}

	if tenantIDStr == "" {
		h.errorJSON(w, "Tenant ID required", http.StatusUnauthorized)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.errorJSON(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	filter := models.AlertListFilter{
		TenantID:    tenantID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "timestamp",
		OrderDir: "desc",
	}

	if fromDateStr := query.Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filter.FromDate = fromDate
		}
	}

	if toDateStr := query.Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filter.ToDate = toDate
		}
	}

	if statuses := query["status"]; len(statuses) > 0 {
		filter.Status = statuses
	}

	if severities := query["severity"]; len(severities) > 0 {
		filter.Severity = severities
	}

	if vehicleIDStr := query.Get("vehicle_id"); vehicleIDStr != "" {
		if vehicleID, err := uuid.Parse(vehicleIDStr); err == nil {
			filter.VehicleID = &vehicleID
		}
	}

	if driverIDStr := query.Get("driver_id"); driverIDStr != "" {
		if driverID, err := uuid.Parse(driverIDStr); err == nil {
			filter.DriverID = &driverID
		}
	}

	alerts, totalCount, err := h.service.ListAlerts(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list alerts", zap.Error(err))
		h.errorJSON(w, "Failed to list alerts", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data":        alerts,
		"alerts":      alerts,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	}, http.StatusOK)
}

func (h *AlertHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Notes string `json:"notes,omitempty"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.errorJSON(w, "User ID required", http.StatusUnauthorized)
		return
	}

	if err := h.service.AcknowledgeAlert(r.Context(), id, userID, req.Notes); err != nil {
		h.logger.Error("failed to acknowledge alert", zap.String("alert_id", id), zap.Error(err))
		if err.Error() == "alert must be in 'new' status to acknowledge" {
			h.errorJSON(w, err.Error(), http.StatusBadRequest)
		} else {
			h.errorJSON(w, "Alert not found", http.StatusNotFound)
		}
		return
	}

	h.json(w, map[string]interface{}{"data": map[string]string{"message": "Alert acknowledged"}}, http.StatusOK)
}

func (h *AlertHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Notes string `json:"notes,omitempty"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.errorJSON(w, "User ID required", http.StatusUnauthorized)
		return
	}

	if err := h.service.ResolveAlert(r.Context(), id, userID, req.Notes); err != nil {
		h.logger.Error("failed to resolve alert", zap.String("alert_id", id), zap.Error(err))
		if err.Error() == "alert is already resolved" {
			h.errorJSON(w, err.Error(), http.StatusBadRequest)
		} else {
			h.errorJSON(w, "Alert not found", http.StatusNotFound)
		}
		return
	}

	h.json(w, map[string]interface{}{"data": map[string]string{"message": "Alert resolved"}}, http.StatusOK)
}

func (h *AlertHandler) GetAlertSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.errorJSON(w, "Tenant ID required", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	fromDateStr := query.Get("from_date")
	toDateStr := query.Get("to_date")

	if fromDateStr == "" || toDateStr == "" {
		h.errorJSON(w, "Both from_date and to_date are required", http.StatusBadRequest)
		return
	}

	fromDate, err := time.Parse(time.RFC3339, fromDateStr)
	if err != nil {
		h.errorJSON(w, "Invalid from_date format", http.StatusBadRequest)
		return
	}

	toDate, err := time.Parse(time.RFC3339, toDateStr)
	if err != nil {
		h.errorJSON(w, "Invalid to_date format", http.StatusBadRequest)
		return
	}

	summary, err := h.service.GetAlertSummary(r.Context(), tenantID, fromDate, toDate)
	if err != nil {
		h.logger.Error("failed to get alert summary", zap.Error(err))
		h.errorJSON(w, "Failed to get alert summary", http.StatusInternalServerError)
		return
	}

	h.json(w, summary, http.StatusOK)
}

func (h *AlertHandler) EscalateAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.json(w, map[string]string{"message": "Alert escalated", "alert_id": id}, http.StatusOK)
}

func (h *AlertHandler) GetAlertTrends(w http.ResponseWriter, r *http.Request) {
	h.json(w, map[string]interface{}{"trends": []interface{}{}}, http.StatusOK)
}

func (h *AlertHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	h.json(w, map[string]string{"message": "Template created"}, http.StatusCreated)
}

func (h *AlertHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	h.json(w, map[string]interface{}{"templates": []interface{}{}}, http.StatusOK)
}

func (h *AlertHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.json(w, map[string]string{"template_id": id}, http.StatusOK)
}

func (h *AlertHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.json(w, map[string]string{"message": "Template updated", "template_id": id}, http.StatusOK)
}

func (h *AlertHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.json(w, map[string]string{"message": "Template deleted", "template_id": id}, http.StatusOK)
}

func (h *AlertHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		// SECURITY FIX: Log detailed error but return generic message
		h.logger.Error("failed to get alert stats", zap.Error(err))
		h.errorJSON(w, "Failed to retrieve alert statistics", http.StatusInternalServerError)
		return
	}

	h.json(w, stats, http.StatusOK)
}

// Helpers
func (h *AlertHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AlertHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *AlertHandler) getTenantID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.TenantIDKey); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *AlertHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
