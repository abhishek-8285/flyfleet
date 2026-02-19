package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/vehicle-service/internal/middleware"
	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TripHandlers struct {
	service   *service.TripService
	logger    *zap.Logger
	validator *middleware.Validator
}

func NewTripHandlers(service *service.TripService, logger *zap.Logger) *TripHandlers {
	return &TripHandlers{
		service:   service,
		logger:    logger,
		validator: middleware.NewValidator(),
	}
}

// ListTrips handles GET /api/v1/trips
func (h *TripHandlers) ListTrips(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := h.getOrgID(r)
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	status := r.URL.Query().Get("status")

	// Validate query parameters
	if validationErrors := h.validator.ValidateListParams(page, limit, status); len(validationErrors) > 0 {
		middleware.WriteValidationError(w, validationErrors)
		return
	}

	params := models.TripListParams{
		OrganizationID: orgID,
		Status:         status,
		Page:           page,
		Limit:          limit,
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

	// Decode to map for validation
	var rawData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	// Validate request
	if validationErrors := h.validator.ValidateCreateTripRequest(rawData); len(validationErrors) > 0 {
		middleware.WriteValidationError(w, validationErrors)
		return
	}

	// Parse validated data into struct
	var req models.CreateTripRequest
	reqBytes, _ := json.Marshal(rawData)
	json.Unmarshal(reqBytes, &req)

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

	idStr := h.getTripID(r)
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

	idStr := h.getTripID(r)
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid trip ID", err)
		return
	}

	// Decode to map for validation
	var rawData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	// Validate request
	if validationErrors := h.validator.ValidateUpdateTripRequest(rawData); len(validationErrors) > 0 {
		middleware.WriteValidationError(w, validationErrors)
		return
	}

	// Parse validated data into struct
	var req models.UpdateTripRequest
	reqBytes, _ := json.Marshal(rawData)
	json.Unmarshal(reqBytes, &req)

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

	idStr := h.getTripID(r)
	tripID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid trip ID", err)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	// Validate telemetry params
	if validationErrors := h.validator.ValidateTelemetryParams(limit); len(validationErrors) > 0 {
		middleware.WriteValidationError(w, validationErrors)
		return
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
func (h *TripHandlers) getTripID(r *http.Request) string {
	return chi.URLParam(r, "id")
}

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
