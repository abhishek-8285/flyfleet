package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/service"
)

// LocationsHandler handles GET /api/v1/vehicles/locations
type LocationsHandler struct {
	batchService *service.BatchLocationService
	logger       *zap.Logger
}

func NewLocationsHandler(bs *service.BatchLocationService, logger *zap.Logger) *LocationsHandler {
	return &LocationsHandler{batchService: bs, logger: logger}
}

func (h *LocationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// FIX: GetOrgID takes context.Context, returns google/uuid — convert to gofrs
	middlewareOrgID := middleware.GetOrgID(r.Context())
	orgID, _ := uuid.FromString(middlewareOrgID.String())
	if orgID == uuid.Nil {
		http.Error(w, `{"error":"missing organization context"}`, http.StatusUnauthorized)
		return
	}

	// Parse query parameters for filtering
	statusFilter := r.URL.Query().Get("status")
	driverIDStr := r.URL.Query().Get("driver_id")
	search := r.URL.Query().Get("search")

	var driverID uuid.UUID
	if driverIDStr != "" {
		driverID, _ = uuid.FromString(driverIDStr)
	}

	// Use filtered method if any filter is set
	var locations []models.VehicleLocationResponse
	var err error
	if statusFilter != "" || driverID != uuid.Nil || search != "" {
		locations, err = h.batchService.GetFilteredOrgLocations(r.Context(), orgID, statusFilter, driverID, search)
	} else {
		locations, err = h.batchService.GetAllOrgLocations(r.Context(), orgID)
	}
	if err != nil {
		h.logger.Error("failed to get org locations", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// Enhanced count by status category
	online, offline, moving, idle, stopped := 0, 0, 0, 0, 0
	for _, loc := range locations {
		switch loc.Status {
		case models.StatusMoving:
			moving++
			online++
		case models.StatusIdle:
			idle++
			online++
		case models.StatusStopped:
			stopped++
			online++
		default:
			offline++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": locations,
		"meta": map[string]interface{}{
			"total":     len(locations),
			"online":    online,
			"offline":   offline,
			"moving":    moving,
			"idle":      idle,
			"stopped":   stopped,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}
