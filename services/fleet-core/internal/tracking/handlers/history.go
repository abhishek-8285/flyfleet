package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/tracking/service"
)

// HistoryHandler handles GET /api/v1/vehicles/{id}/telemetry/history
type HistoryHandler struct {
	queryService *service.TelemetryQueryService
	logger       *zap.Logger
}

func NewHistoryHandler(qs *service.TelemetryQueryService, logger *zap.Logger) *HistoryHandler {
	return &HistoryHandler{queryService: qs, logger: logger}
}

func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// FIX: Use chi.URLParam instead of non-existent extractPathParam
	vehicleIDStr := chi.URLParam(r, "id")
	if vehicleIDStr == "" {
		http.Error(w, `{"error":"missing vehicle id"}`, http.StatusBadRequest)
		return
	}

	vehicleID, err := uuid.FromString(vehicleIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid vehicle_id"}`, http.StatusBadRequest)
		return
	}

	// FIX: GetOrgID takes context.Context, returns google/uuid — convert to gofrs
	middlewareOrgID := middleware.GetOrgID(r.Context())
	orgID, _ := uuid.FromString(middlewareOrgID.String())
	if orgID == uuid.Nil {
		http.Error(w, `{"error":"missing organization context"}`, http.StatusUnauthorized)
		return
	}

	// NEW: Tenant isolation check — vehicle must belong to requesting org
	belongs, err := h.queryService.VehicleBelongsToOrg(r.Context(), orgID, vehicleID)
	if err != nil || !belongs {
		http.Error(w, `{"error":"vehicle not found or not accessible"}`, http.StatusNotFound)
		return
	}

	// Parse query parameters
	q := r.URL.Query()
	startStr := q.Get("start")
	endStr := q.Get("end")
	resolution := q.Get("resolution")

	if startStr == "" || endStr == "" {
		http.Error(w, `{"error":"start and end query parameters are required"}`, http.StatusBadRequest)
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, `{"error":"invalid start format, use RFC3339"}`, http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, `{"error":"invalid end format, use RFC3339"}`, http.StatusBadRequest)
		return
	}

	result, err := h.queryService.GetHistoryWithSummary(r.Context(), vehicleID, start, end, resolution)
	if err != nil {
		h.logger.Error("history query failed", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
