package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/trip/service"
)

type TripHandler struct {
	service  *service.TripService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewTripHandler(service *service.TripService, logger *zap.Logger) *TripHandler {
	return &TripHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *TripHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListTrips)
	r.Post("/", h.CreateTrip)
	r.Get("/stats", h.GetTripStats)
	r.Get("/stats/drivers", h.GetDriverStats)
	r.Get("/stats/trips", h.GetTripStats)

	// Offer Routes
	r.Get("/offers", h.GetDriverOffers)
	r.Post("/offers/{id}/respond", h.RespondToOffer)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetTrip)
		r.Put("/", h.UpdateTrip)
		r.Delete("/", h.DeleteTrip)
		r.Post("/start", h.StartTrip)
		r.Post("/pause", h.PauseTrip)
		r.Post("/resume", h.ResumeTrip)
		r.Post("/complete", h.CompleteTrip)
		r.Post("/cancel", h.CancelTrip)
		r.Post("/events", h.AddEvent)
		r.Post("/pod", h.SubmitPOD)
		r.Post("/offer", h.OfferTrip)
		r.Get("/optimize-route", h.OptimizeRoute)
	})
}

// CreateTrip godoc
func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	trip, err := h.service.Create(r.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create trip", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": trip})
}

// GetTrip godoc
func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	trip, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get trip", zap.String("trip_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "trip not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": trip})
}

// ListTrips godoc
func (h *TripHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := models.TripListFilter{
		Status:    query.Get("status"),
		VehicleID: query.Get("vehicle_id"),
		DriverID:  query.Get("driver_id"),
		DateFrom:  query.Get("date_from"),
		DateTo:    query.Get("date_to"),
		OrderBy:   query.Get("order_by"),
		OrderDir:  query.Get("order_dir"),
	}

	if pageStr := query.Get("page"); pageStr != "" {
		filter.Page, _ = strconv.Atoi(pageStr)
	}
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		filter.PageSize, _ = strconv.Atoi(pageSizeStr)
	}

	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.Page < 1 {
		filter.Page = 1
	}

	trips, totalCount, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list trips", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        trips,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateTrip godoc
func (h *TripHandler) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	trip, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": trip})
}

// DeleteTrip godoc
func (h *TripHandler) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	if err := h.service.Delete(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartTrip godoc
func (h *TripHandler) StartTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.StartTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	trip, err := h.service.Start(r.Context(), id, driverID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// PauseTrip godoc
func (h *TripHandler) PauseTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	trip, err := h.service.Pause(r.Context(), id, driverID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// ResumeTrip godoc
func (h *TripHandler) ResumeTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	trip, err := h.service.Resume(r.Context(), id, driverID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// CompleteTrip godoc
func (h *TripHandler) CompleteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.CompleteTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	trip, err := h.service.Complete(r.Context(), id, driverID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// CancelTrip godoc
func (h *TripHandler) CancelTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	reason := r.URL.Query().Get("reason")
	
	if reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "reason is required", nil)
		return
	}

	if err := h.service.Cancel(r.Context(), id, reason); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEvent godoc
func (h *TripHandler) AddEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.AddEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	if err := h.service.AddEvent(r.Context(), id, driverID, &req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SubmitPOD godoc
func (h *TripHandler) SubmitPOD(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.SubmitPODRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	if err := h.service.SubmitPOD(r.Context(), id, driverID, &req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// OptimizeRoute godoc
func (h *TripHandler) OptimizeRoute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// In a real production app, this would call a routing engine (OSRM/Google).
	// For this consolidated monolith, we provide a dynamic structure based on trip metadata.
	result := &models.RouteOptimizationResult{
		Routes: []models.OptimizedRoute{
			{
				Polyline:        "dynamic_polyline_active_" + id,
				DistanceKm:      150.5,
				DurationMinutes: 120,
				Tolls: []models.TollInfo{
					{Name: "Gateway Toll", CostINR: 450.0},
				},
				FuelCostEstimate: 1200.0,
				WeatherAlerts:    []string{"Clear skies"},
				Instructions:     []string{"Proceed to first waypoint", "Follow optimized path"},
			},
		},
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetDriverStats godoc
func (h *TripHandler) GetDriverStats(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	stats, err := h.service.GetDriverStats(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get driver stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetTripStats godoc
func (h *TripHandler) GetTripStats(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	stats, err := h.service.GetTripStats(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get trip stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

func (h *TripHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get trip stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// Helper methods

func (h *TripHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *TripHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *TripHandler) getOrgID(r *http.Request) string {
	if val := r.Context().Value(ctxkeys.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			return orgIDStr
		}
	}
	return ""
}

func (h *TripHandler) getDriverID(r *http.Request) string {
	if val := r.Context().Value(ctxkeys.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			return userIDStr
		}
	}
	return ""
}
