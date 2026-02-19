package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/pagination"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/flyfleet/trip-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type TripHandler struct {
	service *service.TripService
	logger  *zap.Logger
}

func NewTripHandler(service *service.TripService, logger *zap.Logger) *TripHandler {
	return &TripHandler{
		service: service,
		logger:  logger,
	}
}

// Helper methods

func (h *TripHandler) decodeJSON(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

func (h *TripHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("failed to encode JSON response", zap.Error(err))
		}
	}
}

func (h *TripHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}

// CreateTrip godoc
func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTripRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get org_id from context/fallback
	orgID := "" // value, ok := r.Context().Value("org_id").(string)
	if orgID == "" {
		orgID = "550e8400-e29b-41d4-a716-446655440002" // fallback for testing
	}

	trip, err := h.service.Create(r.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create trip", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, trip)
}

// GetTrip godoc
func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	trip, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get trip", zap.String("trip_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "trip not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// ListTrips godoc
func (h *TripHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	p := pagination.GetParams(r)

	query := r.URL.Query()
	filter := models.TripListFilter{
		Status:    query.Get("status"),
		VehicleID: query.Get("vehicle_id"),
		DriverID:  query.Get("driver_id"),
		DateFrom:  query.Get("date_from"),
		DateTo:    query.Get("date_to"),
		Page:      p.Page,
		PageSize:  p.Limit,
	}

	trips, totalCount, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list trips", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, pagination.NewResponse(trips, totalCount, p.Page, p.Limit))
}

// UpdateTrip godoc
func (h *TripHandler) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.UpdateTripRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	trip, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// DeleteTrip godoc
func (h *TripHandler) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	if err := h.service.Delete(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartTrip godoc
func (h *TripHandler) StartTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.StartTripRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get driver_id from context/fallback
	driverID := "" // value, ok := r.Context().Value("driver_id").(string)
	if driverID == "" {
		driverID = "550e8400-e29b-41d4-a716-446655440001" // fallback for testing
	}

	trip, err := h.service.Start(r.Context(), id, driverID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// PauseTrip godoc
func (h *TripHandler) PauseTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// Get driver_id from context
	driverID := ""
	if driverID == "" {
		driverID = "default-driver"
	}

	trip, err := h.service.Pause(r.Context(), id, driverID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// ResumeTrip godoc
func (h *TripHandler) ResumeTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// Get driver_id from context
	driverID := ""
	if driverID == "" {
		driverID = "default-driver"
	}

	trip, err := h.service.Resume(r.Context(), id, driverID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// CompleteTrip godoc
func (h *TripHandler) CompleteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.CompleteTripRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get driver_id from context
	driverID := ""
	if driverID == "" {
		driverID = "default-driver"
	}

	trip, err := h.service.Complete(r.Context(), id, driverID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, trip)
}

// CancelTrip godoc
func (h *TripHandler) CancelTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	reason := r.URL.Query().Get("reason")
	
	if reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "reason is required")
		return
	}

	if err := h.service.Cancel(r.Context(), id, reason); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddEvent godoc
func (h *TripHandler) AddEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.AddEventRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get driver_id from context
	driverID := ""
	if driverID == "" {
		driverID = "default-driver"
	}

	if err := h.service.AddEvent(r.Context(), id, driverID, &req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SubmitPOD godoc
func (h *TripHandler) SubmitPOD(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var req models.SubmitPODRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get driver_id from context
	driverID := ""
	if driverID == "" {
		driverID = "default-driver"
	}

	if err := h.service.SubmitPOD(r.Context(), id, driverID, &req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// OptimizeRoute godoc
func (h *TripHandler) OptimizeRoute(w http.ResponseWriter, r *http.Request) {
	// id := chi.URLParam(r, "id") // Not used in mock implementation
	
	// This would integrate with OSRM or Google Maps API
	// For now, returning mock data
	result := &models.RouteOptimizationResult{
		Routes: []models.OptimizedRoute{
			{
				Polyline:        "mock_polyline_data",
				DistanceKm:      150.5,
				DurationMinutes: 120,
				Tolls: []models.TollInfo{
					{Name: "Mumbai-Pune Expressway", CostINR: 450.0},
				},
				FuelCostEstimate: 1200.0,
				WeatherAlerts:    []string{},
				Instructions:     []string{"Take NH48 toward Pune", "Continue for 150km"},
			},
		},
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetDriverStats godoc
func (h *TripHandler) GetDriverStats(w http.ResponseWriter, r *http.Request) {
	orgID := ""
	if orgID == "" {
		orgID = "default-org"
	}

	stats, err := h.service.GetDriverStats(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get driver stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetTripStats godoc
func (h *TripHandler) GetTripStats(w http.ResponseWriter, r *http.Request) {
	orgID := ""
	if orgID == "" {
		orgID = "default-org"
	}

	stats, err := h.service.GetTripStats(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get trip stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetTripTelemetry godoc
func (h *TripHandler) GetTripTelemetry(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "id")
	
	// Validate UUID format roughly
	if _, err := uuid.Parse(tripID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid trip_id format")
		return
	}

	orgID := ""
	if orgID == "" {
		orgID = "550e8400-e29b-41d4-a716-446655440002" // Fallback for safety
	}

	query := r.URL.Query()
	filter := models.TelemetryFilter{
		Aggregation: query.Get("aggregation"),
	}

	if st := query.Get("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			filter.StartTime = &t
		}
	}
	if et := query.Get("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			filter.EndTime = &t
		}
	}
	if l := query.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			filter.Limit = v
		}
	}

	resp, err := h.service.GetTripTelemetry(r.Context(), tripID, orgID, filter)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "unauthorized") {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, resp)
}

func (h *TripHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get trip stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}
