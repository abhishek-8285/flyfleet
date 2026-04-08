package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core/internal/domains/monitoring/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// MonitoringHandler handles monitoring HTTP requests
type MonitoringHandler struct {
	service service.PulseService
	logger  *zap.Logger
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(service service.PulseService, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers monitoring routes
// RegisterRoutes registers monitoring routes
func (h *MonitoringHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.GetSystemHealth)
	r.Get("/sla", h.GetSLAMetrics)
	r.Get("/metrics", h.GetSystemMetrics)
	r.Get("/alerts", h.GetActiveAlerts)
	r.Post("/alerts/acknowledge", h.AcknowledgeAlert)
	r.Post("/alerts/resolve", h.ResolveAlert)

	// SRE Endpoints
	r.Get("/synthetics", h.GetSyntheticChecks)
	r.Get("/slos", h.GetSLOs)
	r.Get("/forecasts", h.GetCapacityForecasts)
	r.Get("/business-metrics", h.GetBusinessKPIs)
	r.Get("/global-health", h.GetGlobalHealth)
}

// GetSystemHealth handles GET /health
func (h *MonitoringHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	health := h.service.CheckHealth(ctx)

	response := models.SystemHealth{
		Status:      models.HealthStatusUp,
		Components:  make(map[string]models.HealthCheck),
		Uptime:      h.service.GetUptime(),
		LastChecked: time.Now(),
	}

	// Determine overall status
	hasDown := false
	hasDegraded := false

	for component, status := range health {
		check := models.HealthCheck{
			Component:   component,
			Status:      models.HealthStatus(status),
			LastChecked: time.Now(),
		}

		if status == "down" {
			hasDown = true
		} else if status == "degraded" {
			hasDegraded = true
		}

		response.Components[component] = check
	}

	if hasDown {
		response.Status = models.HealthStatusDown
	} else if hasDegraded {
		response.Status = models.HealthStatusDegraded
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetSLAMetrics handles GET /sla
func (h *MonitoringHandler) GetSLAMetrics(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	metrics, err := h.service.GetSLAMetrics(r.Context(), period)
	if err != nil {
		h.logger.Error("Failed to get SLA metrics", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, metrics)
}

// GetSystemMetrics handles GET /metrics
func (h *MonitoringHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetSystemMetrics(r.Context())
	if err != nil {
		h.logger.Error("Failed to get system metrics", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, metrics)
}

// GetActiveAlerts handles GET /alerts
func (h *MonitoringHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.service.GetActiveAlerts(r.Context())
	if err != nil {
		h.logger.Error("Failed to get active alerts", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// AcknowledgeAlert handles POST /alerts/acknowledge
func (h *MonitoringHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertID string `json:"alert_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AlertID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "alert_id is required")
		return
	}

	if err := h.service.AcknowledgeAlert(r.Context(), req.AlertID); err != nil {
		h.logger.Error("Failed to acknowledge alert", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Alert acknowledged",
	})
}

// ResolveAlert handles POST /alerts/resolve
func (h *MonitoringHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertID string `json:"alert_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AlertID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "alert_id is required")
		return
	}

	if err := h.service.ResolveAlert(r.Context(), req.AlertID); err != nil {
		h.logger.Error("Failed to resolve alert", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Alert resolved",
	})
}

// SRE Handlers

func (h *MonitoringHandler) GetSyntheticChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.GetSyntheticChecks(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, checks)
}

func (h *MonitoringHandler) GetSLOs(w http.ResponseWriter, r *http.Request) {
	slos, err := h.service.GetSLOs(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, slos)
}

func (h *MonitoringHandler) GetCapacityForecasts(w http.ResponseWriter, r *http.Request) {
	forecasts, err := h.service.GetCapacityForecasts(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, forecasts)
}

func (h *MonitoringHandler) GetBusinessKPIs(w http.ResponseWriter, r *http.Request) {
	kpis, err := h.service.GetBusinessKPIs(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, kpis)
}

func (h *MonitoringHandler) GetGlobalHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.GetGlobalHealth(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, health)
}

// Helper methods

func (h *MonitoringHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (h *MonitoringHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{
		"error": message,
	})
}
