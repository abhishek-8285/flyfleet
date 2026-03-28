package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/service"
	"github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TrackingQueryHandler handles the read-path tracking APIs.
type TrackingQueryHandler struct {
	svc    *service.TelemetryQueryService
	logger *zap.Logger
}

// NewTrackingQueryHandler creates a new query handler.
func NewTrackingQueryHandler(svc *service.TelemetryQueryService, logger *zap.Logger) *TrackingQueryHandler {
	return &TrackingQueryHandler{svc: svc, logger: logger}
}

// RegisterTrackingRoutes mounts all read-path tracking endpoints.
func (h *TrackingQueryHandler) RegisterTrackingRoutes(r chi.Router) {
	r.Get("/v1/vehicles/{id}/location", h.GetLiveLocation)
	r.Get("/v1/vehicles/{id}/history", h.GetHistory)
	r.Get("/v1/vehicles/{id}/replay", h.GetReplay)
	r.Get("/v1/vehicles/nearby", h.GetNearby)
}

// GetLiveLocation returns the current position of a vehicle.
func (h *TrackingQueryHandler) GetLiveLocation(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.error(w, "invalid vehicle_id", http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetOrgID(r.Context())
	if tenantID == uuid.Nil {
		h.error(w, "tenant context missing", http.StatusUnauthorized)
		return
	}

	loc, err := h.svc.GetLiveLocation(r.Context(), tenantID, vehicleID)
	if err != nil {
		h.logger.Warn("live location not found", zap.String("vehicle_id", vehicleID.String()), zap.Error(err))
		h.error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.jsonOK(w, loc)
}

// GetHistory returns paginated telemetry history.
func (h *TrackingQueryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.error(w, "invalid vehicle_id", http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetOrgID(r.Context())
	if tenantID == uuid.Nil {
		h.error(w, "tenant context missing", http.StatusUnauthorized)
		return
	}

	from, to, page, pageSize := h.parseQueryParams(r)

	events, err := h.svc.GetHistory(r.Context(), tenantID, vehicleID, from, to, page, pageSize)
	if err != nil {
		h.logger.Error("history query failed", zap.Error(err))
		h.error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.jsonOK(w, map[string]any{
		"vehicle_id": vehicleID,
		"count":      len(events),
		"data":       events,
	})
}

// GetReplay returns ordered telemetry for map replay.
func (h *TrackingQueryHandler) GetReplay(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.error(w, "invalid vehicle_id", http.StatusBadRequest)
		return
	}

	tenantID := middleware.GetOrgID(r.Context())
	if tenantID == uuid.Nil {
		h.error(w, "tenant context missing", http.StatusUnauthorized)
		return
	}

	from, to, _, _ := h.parseQueryParams(r)

	points, err := h.svc.GetReplay(r.Context(), tenantID, vehicleID, from, to)
	if err != nil {
		h.error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonOK(w, map[string]any{
		"vehicle_id": vehicleID,
		"count":      len(points),
		"data":       points,
	})
}

// GetNearby returns vehicles near a geographic point.
func (h *TrackingQueryHandler) GetNearby(w http.ResponseWriter, r *http.Request) {
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	radius, _ := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)

	tenantID := middleware.GetOrgID(r.Context())
	if tenantID == uuid.Nil {
		h.error(w, "tenant context missing", http.StatusUnauthorized)
		return
	}

	if radius <= 0 {
		radius = 1000 // default 1km
	}

	locs, err := h.svc.GetNearby(r.Context(), tenantID, lat, lng, radius)
	if err != nil {
		h.logger.Error("nearby query failed", zap.Error(err))
		h.error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.jsonOK(w, map[string]any{
		"lat":    lat,
		"lng":    lng,
		"radius": radius,
		"count":  len(locs),
		"data":   locs,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (h *TrackingQueryHandler) parseQueryParams(r *http.Request) (time.Time, time.Time, int, int) {
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	return from, to, page, pageSize
}

func (h *TrackingQueryHandler) error(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *TrackingQueryHandler) jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
