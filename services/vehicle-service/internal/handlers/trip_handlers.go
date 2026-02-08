package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TripHandlers struct {
	service *service.TripService
	logger  *zap.Logger
}

func NewTripHandlers(service *service.TripService, logger *zap.Logger) *TripHandlers {
	return &TripHandlers{service: service, logger: logger}
}

// ListTrips handles GET /api/v1/trips
func (h *TripHandlers) ListTrips(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := h.getOrgID(r)
	
	params := models.TripListParams{
		OrganizationID: orgID,
		Status:         r.URL.Query().Get("status"),
	}
	
	if page, _ := strconv.Atoi(r.URL.Query().Get("page")); page > 0 {
		params.Page = page
	} else {
		params.Page = 1
	}
	if limit, _ := strconv.Atoi(r.URL.Query().Get("limit")); limit > 0 {
		params.Limit = limit
	} else {
		params.Limit = 20
	}
	
	if vid := r.URL.Query().Get("vehicle_id"); vid != "" {
		if id, err := uuid.Parse(vid); err == nil {
			params.VehicleID = &id
		}
	}
	if did := r.URL.Query().Get("driver_id"); did != "" {
		if id, err := uuid.Parse(did); err == nil {
			params.DriverID = &id
		}
	}

	ctx := context.Background()
	trips, total, err := h.service.ListTrips(ctx, params)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list trips", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"trips": trips,
		"total": total,
		"page":  params.Page,
		"limit": params.Limit,
	})
}

// CreateTrip handles POST /api/v1/trips
func (h *TripHandlers) CreateTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	orgID := h.getOrgID(r)
	ctx := context.Background()
	
	trip, err := h.service.CreateTrip(ctx, orgID, &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create trip", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, trip)
}

// GetTrip handles GET /api/v1/trips/:id
func (h *TripHandlers) GetTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Missing trip ID", nil)
		return
	}

	ctx := context.Background()
	var trip *models.Trip
	var err error

	// Try parsing as UUID first, then as trip_id string
	if id, parseErr := uuid.Parse(idStr); parseErr == nil {
		trip, err = h.service.GetTrip(ctx, id)
	} else {
		trip, err = h.service.GetTripByTripID(ctx, idStr)
	}

	if err != nil {
		h.writeError(w, http.StatusNotFound, "Trip not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, trip)
}

// UpdateTrip handles PUT /api/v1/trips/:id
func (h *TripHandlers) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid trip ID", err)
		return
	}

	var req models.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := context.Background()
	trip, err := h.service.UpdateTrip(ctx, id, &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update trip", err)
		return
	}

	h.writeJSON(w, http.StatusOK, trip)
}

// GetTelemetry handles GET /api/v1/trips/:id/telemetry
func (h *TripHandlers) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	tripID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid trip ID", err)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	ctx := context.Background()
	events, err := h.service.GetTelemetry(ctx, tripID, limit)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get telemetry", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// GetSettings handles GET /api/v1/trips/settings
func (h *TripHandlers) GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := h.getOrgID(r)
	ctx := context.Background()
	
	settings, err := h.service.GetSettings(ctx, orgID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get settings", err)
		return
	}

	h.writeJSON(w, http.StatusOK, settings)
}

// UpdateSettings handles PUT /api/v1/trips/settings
func (h *TripHandlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings models.TripSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	settings.OrganizationID = h.getOrgID(r)
	settings.UpdatedAt = time.Now()

	ctx := context.Background()
	if err := h.service.UpdateSettings(ctx, &settings); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update settings", err)
		return
	}

	h.writeJSON(w, http.StatusOK, settings)
}

// Helpers
func (h *TripHandlers) getOrgID(r *http.Request) uuid.UUID {
	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		orgIDStr = "00000000-0000-0000-0000-000000000000"
	}
	orgID, _ := uuid.Parse(orgIDStr)
	return orgID
}

func (h *TripHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TripHandlers) writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{"error": message}
	if err != nil {
		response["details"] = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
