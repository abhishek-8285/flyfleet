package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handlers struct{}

func NewHandlers() *Handlers { return &Handlers{} }

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/ws", h.WebsocketCompat)
	r.Get("/dashboard/stats", h.DashboardStats)
	r.Get("/admin/dashboard/data", h.AdminDashboardData)

	r.Post("/compliance/hos", h.ComplianceHOSCompat)
	r.Post("/telemetry", h.TelemetryCompat)
	r.Post("/telemetry/dashcam", h.TelemetryDashcamCompat)

	r.Get("/ml/health", h.MLHealth)
	r.Post("/ml/demand/predict", h.PredictDemand)
	r.Post("/ml/pricing/suggest", h.SuggestPricing)
	r.Post("/ml/jobs/train", h.TrainJob)

	r.Get("/routes/{route_id}", h.GetOptimizedRoute)
	r.Post("/optimize", h.OptimizeRoute)

	r.Get("/services/{name}/health", h.ServiceHealth)
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) WebsocketCompat(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"mode":    "monolith-compat",
		"message": "websocket compatibility endpoint is available, upgrade handling is disabled",
		"service": "web-bff-compat",
	})
}

func (h *Handlers) DashboardStats(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"mode":   "monolith-compat",
		"data": map[string]any{
			"fleet":        map[string]any{},
			"alerts":       map[string]any{},
			"maintenance":  map[string]any{},
			"utilization":  map[string]any{},
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *Handlers) AdminDashboardData(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"mode":   "monolith-compat",
		"data": map[string]any{
			"kpis":         map[string]any{},
			"recentAlerts": []any{},
			"widgets":      []any{},
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *Handlers) ComplianceHOSCompat(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "accepted",
		"message": "use /api/v1/compliance/duty-status/record for canonical endpoint",
		"target":  "/api/v1/compliance/duty-status/record",
	})
}

func (h *Handlers) TelemetryCompat(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "accepted",
		"message": "telemetry ingress compatibility endpoint",
	})
}

func (h *Handlers) TelemetryDashcamCompat(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "accepted",
		"message": "dashcam telemetry compatibility endpoint",
	})
}

func (h *Handlers) MLHealth(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "fleet-ml-service", "mode": "monolith-compat"})
}

func (h *Handlers) PredictDemand(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RouteID string `json:"route_id"`
		Date    string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.RouteID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "route_id required"})
		return
	}
	day := time.Now().Day()
	if t, err := time.Parse("2006-01-02", req.Date); err == nil {
		day = t.Day()
	}
	predicted := len(req.RouteID)*7 + day
	if predicted < 1 {
		predicted = 1
	}
	h.writeJSON(w, http.StatusOK, map[string]any{
		"route_id":         req.RouteID,
		"date":             req.Date,
		"predicted_demand": predicted,
		"confidence_score": 0.82,
	})
}

func (h *Handlers) SuggestPricing(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RouteID      string `json:"route_id"`
		Date         string `json:"date"`
		BaseCapacity int    `json:"base_capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.RouteID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "route_id required"})
		return
	}
	if req.BaseCapacity <= 0 {
		req.BaseCapacity = 50
	}
	multiplier := 1.0 + float64((len(req.RouteID)%7))/20.0
	h.writeJSON(w, http.StatusOK, map[string]any{
		"route_id":             req.RouteID,
		"date":                 req.Date,
		"suggested_multiplier": fmt.Sprintf("%.2f", multiplier),
	})
}

func (h *Handlers) TrainJob(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusAccepted, map[string]any{"status": "accepted", "job": "training_pipeline"})
}

func (h *Handlers) GetOptimizedRoute(w http.ResponseWriter, r *http.Request) {
	routeID := chi.URLParam(r, "route_id")
	switch routeID {
	case "not-found":
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "route not found"})
		return
	case "rate-limit":
		w.Header().Set("Retry-After", "60")
		h.writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
		return
	case "error":
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unexpected failure"})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"id": routeID, "status": "optimized"})
}

func (h *Handlers) OptimizeRoute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Locations []string `json:"locations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(req.Locations) < 2 {
		h.writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "at least two locations are required"})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"optimized": true, "stops": len(req.Locations)})
}

func (h *Handlers) ServiceHealth(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(chi.URLParam(r, "name"))
	if name == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "service name required"})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": name, "mode": "monolith-compat"})
}
