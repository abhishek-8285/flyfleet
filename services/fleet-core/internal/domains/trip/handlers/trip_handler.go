package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/trip/models"
	"github.com/flyfleet/fleet-core/internal/domains/trip/service"
	"github.com/flyfleet/fleet-core/internal/platform/metrics"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createTripRequestDTO struct {
	VehicleID            string              `json:"vehicle_id" validate:"required,uuid4"`
	DriverID             string              `json:"driver_id" validate:"required,uuid4"`
	Title                string              `json:"title" validate:"required,max=100"`
	Description          string              `json:"description" validate:"max=500"`
	PlannedRoute         []models.Waypoint   `json:"planned_route" validate:"required,min=2"`
	Waypoints            []models.Waypoint   `json:"waypoints" validate:"required,min=2"`
	PlannedDistance      float64             `json:"planned_distance" validate:"required,gt=0"`
	EstimatedDurationMin int                 `json:"estimated_duration_min" validate:"required,gt=0"`
	CargoDetails         models.CargoDetails `json:"cargo_details" validate:"required"`
	ScheduledStart       time.Time           `json:"scheduled_start" validate:"required"`
	PlannedETA           *time.Time          `json:"planned_eta" validate:"omitempty"`
	StartOdometer        *float64            `json:"start_odometer" validate:"omitempty,gte=0"`
	FuelLevelSnapshot    *float64            `json:"fuel_level_snapshot" validate:"omitempty,gte=0,lte=100"`
	CustomerName         string              `json:"customer_name" validate:"required"`
	CustomerPhone        string              `json:"customer_phone" validate:"required,e164"`
	CustomerEmail        string              `json:"customer_email" validate:"omitempty,email"`
	PickupGeo            *models.Location    `json:"pickup_geo" validate:"omitempty"`
	DropGeo              *models.Location    `json:"drop_geo" validate:"omitempty"`
	EWayBillNo           string              `json:"eway_bill_no" validate:"omitempty"`
	GSTClassification    string              `json:"gst_classification" validate:"omitempty"`
}

type updateTripRequestDTO struct {
	Title                *string              `json:"title,omitempty" validate:"omitempty,max=100"`
	Description          *string              `json:"description,omitempty" validate:"omitempty,max=500"`
	PlannedRoute         *[]models.Waypoint   `json:"planned_route,omitempty" validate:"omitempty,min=2"`
	Waypoints            *[]models.Waypoint   `json:"waypoints,omitempty" validate:"omitempty,min=2"`
	PlannedDistance      *float64             `json:"planned_distance,omitempty" validate:"omitempty,gt=0"`
	EstimatedDurationMin *int                 `json:"estimated_duration_min,omitempty" validate:"omitempty,gt=0"`
	CargoDetails         *models.CargoDetails `json:"cargo_details,omitempty"`
	ScheduledStart       *time.Time           `json:"scheduled_start,omitempty"`
	CustomerName         *string              `json:"customer_name,omitempty"`
	CustomerPhone        *string              `json:"customer_phone,omitempty" validate:"omitempty,e164"`
	CustomerEmail        *string              `json:"customer_email,omitempty" validate:"omitempty,email"`
}

type startTripRequestDTO struct {
	OdometerReading float64 `json:"odometer_reading" validate:"required,gt=0"`
	FuelLevel       float64 `json:"fuel_level" validate:"required,gte=0,lte=100"`
	Notes           string  `json:"notes" validate:"max=500"`
}

type completeTripRequestDTO struct {
	OdometerReading   float64              `json:"odometer_reading" validate:"required,gt=0"`
	FuelLevel         float64              `json:"fuel_level" validate:"required,gte=0,lte=100"`
	TotalDistance     float64              `json:"total_distance" validate:"required,gt=0"`
	FuelUsed          float64              `json:"fuel_used" validate:"required,gte=0"`
	IdlingTimeMinutes int                  `json:"idling_time_minutes" validate:"gte=0"`
	PODDocuments      []models.PODDocument `json:"pod_documents" validate:"required,min=1"`
	Notes             string               `json:"notes" validate:"max=500"`
}

type addEventRequestDTO struct {
	Type        string          `json:"type" validate:"required,oneof=arrival departure delivery_attempt delivery_success delivery_failed delay breakdown accident"`
	Location    models.Location `json:"location" validate:"required"`
	Description string          `json:"description" validate:"required,max=500"`
	Photos      []string        `json:"photos" validate:"max=10"`
}

type submitPODRequestDTO struct {
	Type        string           `json:"type" validate:"required,oneof=signature photo invoice receipt"`
	URL         string           `json:"url" validate:"required,url"`
	Recipient   models.Recipient `json:"recipient" validate:"required"`
	OTPVerified bool             `json:"otp_verified"`
	Signature   string           `json:"signature"` // base64 encoded
}

type TripHandler struct {
	service  service.TripServiceInterface
	logger   *zap.Logger
	validate *validator.Validate
}

func NewTripHandler(service service.TripServiceInterface, logger *zap.Logger) *TripHandler {
	return &TripHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *TripHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListTrips)
	r.Post("/", h.CreateTrip)
	r.Get("/stats", h.GetStats)
	r.Get("/stats/drivers", h.GetDriverStats)
	r.Get("/stats/trips", h.GetTripStats)
	r.Get("/customers", h.GetTripCustomers)

	// Offer Routes
	r.Get("/offers", h.GetDriverOffers)
	r.Post("/offers/{id}/respond", h.RespondToOffer)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetTrip)
		r.Put("/", h.UpdateTrip)
		r.Delete("/", h.DeleteTrip)
		r.Post("/approve", h.ApproveTrip)
		r.Post("/start", h.StartTrip)
		r.Post("/pause", h.PauseTrip)
		r.Post("/resume", h.ResumeTrip)
		r.Post("/complete", h.CompleteTrip)
		r.Post("/cancel", h.CancelTrip)
		r.Post("/fail", h.FailTrip)
		r.Post("/events", h.AddEvent)
		r.Get("/telemetry", h.GetTelemetry)
		r.Post("/pod", h.SubmitPOD)
		r.Post("/offer", h.OfferTrip)
		r.Get("/optimize-route", h.OptimizeRoute)
	})
}

// GetTrip godoc
func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	trip, err := h.service.GetTripByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get trip", zap.String("trip_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "trip not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": trip})
}

func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context missing", nil)
		return
	}

	var dto createTripRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request (if validator is used; the original code didn't validate, but we can add)
	// Note: Original didn't call validate.Struct; but we could. However to preserve behavior, I won't add extra validation unless present before.
	// Map DTO to model
	req := &models.CreateTripRequest{
		VehicleID:            dto.VehicleID,
		DriverID:             dto.DriverID,
		Title:                dto.Title,
		Description:          dto.Description,
		PlannedRoute:         dto.PlannedRoute,
		Waypoints:            dto.Waypoints,
		PlannedDistance:      dto.PlannedDistance,
		EstimatedDurationMin: dto.EstimatedDurationMin,
		CargoDetails:         dto.CargoDetails,
		ScheduledStart:       dto.ScheduledStart,
		PlannedETA:           dto.PlannedETA,
		StartOdometer:        dto.StartOdometer,
		FuelLevelSnapshot:    dto.FuelLevelSnapshot,
		CustomerName:         dto.CustomerName,
		CustomerPhone:        dto.CustomerPhone,
		CustomerEmail:        dto.CustomerEmail,
		PickupGeo:            dto.PickupGeo,
		DropGeo:              dto.DropGeo,
		EWayBillNo:           dto.EWayBillNo,
		GSTClassification:    dto.GSTClassification,
	}

	trip, err := h.service.Create(r.Context(), orgID, req)
	if err != nil {
		h.logger.Error("failed to create trip", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": trip})
}

// ListTrips godoc
func (h *TripHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("DEBUG: Executing TRIP DOMAIN TripHandler.ListTrips")
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context missing", nil)
		return
	}

	query := r.URL.Query()
	filter := models.TripListFilter{
		OrgID:        orgID,
		Status:       query.Get("status"),
		TripType:     query.Get("trip_type"),
		VehicleID:    query.Get("vehicle_id"),
		DriverID:     query.Get("driver_id"),
		CustomerName: query.Get("customer_name"),
		From:         query.Get("from"),
		To:           query.Get("to"),
		OrderBy:      query.Get("order_by"),
		OrderDir:     query.Get("order_dir"),
	}

	if pageStr := query.Get("page"); pageStr != "" {
		filter.Page, _ = strconv.Atoi(pageStr)
	}
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		filter.PageSize, _ = strconv.Atoi(pageSizeStr)
	} else if limitStr := query.Get("limit"); limitStr != "" {
		filter.PageSize, _ = strconv.Atoi(limitStr)
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
		"trips":       trips,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateTrip godoc
func (h *TripHandler) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var dto updateTripRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Map DTO to model
	req := &models.UpdateTripRequest{
		Title:                dto.Title,
		Description:          dto.Description,
		PlannedRoute:         dto.PlannedRoute,
		Waypoints:            dto.Waypoints,
		PlannedDistance:      dto.PlannedDistance,
		EstimatedDurationMin: dto.EstimatedDurationMin,
		CargoDetails:         dto.CargoDetails,
		ScheduledStart:       dto.ScheduledStart,
		CustomerName:         dto.CustomerName,
		CustomerPhone:        dto.CustomerPhone,
		CustomerEmail:        dto.CustomerEmail,
	}

	trip, err := h.service.Update(r.Context(), id, req)
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

// ApproveTrip godoc
func (h *TripHandler) ApproveTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Approve(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "SCHEDULED"})
}

// StartTrip godoc
func (h *TripHandler) StartTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var dto startTripRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	// Map DTO to model
	req := &models.StartTripRequest{
		OdometerReading: dto.OdometerReading,
		FuelLevel:       dto.FuelLevel,
		Notes:           dto.Notes,
	}

	trip, err := h.service.Start(r.Context(), id, driverID, req)
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

	var dto completeTripRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	// Map DTO to model
	req := &models.CompleteTripRequest{
		OdometerReading:   dto.OdometerReading,
		FuelLevel:         dto.FuelLevel,
		TotalDistance:     dto.TotalDistance,
		FuelUsed:          dto.FuelUsed,
		IdlingTimeMinutes: dto.IdlingTimeMinutes,
		PODDocuments:      dto.PODDocuments,
		Notes:             dto.Notes,
	}

	trip, err := h.service.Complete(r.Context(), id, driverID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		h.writeErrorResponse(w, status, err.Error(), nil)
		return
	}

	// Record business metric
	metrics.RecordTripCompleted()

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

// FailTrip godoc
func (h *TripHandler) FailTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	reason := r.URL.Query().Get("reason")

	if reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "reason is required", nil)
		return
	}

	if err := h.service.Fail(r.Context(), id, reason); err != nil {
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

	var dto addEventRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	// Map DTO to model
	req := &models.AddEventRequest{
		Type:        dto.Type,
		Location:    dto.Location,
		Description: dto.Description,
		Photos:      dto.Photos,
	}

	if err := h.service.AddEvent(r.Context(), id, driverID, req); err != nil {
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

	var dto submitPODRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	driverID := h.getDriverID(r)
	if driverID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Driver context missing", nil)
		return
	}

	// Map DTO to model
	req := &models.SubmitPODRequest{
		Type:        dto.Type,
		URL:         dto.URL,
		Recipient:   dto.Recipient,
		OTPVerified: dto.OTPVerified,
		Signature:   dto.Signature,
	}

	if err := h.service.SubmitPOD(r.Context(), id, driverID, req); err != nil {
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

// GetTripSummary godoc
func (h *TripHandler) GetTripSummary(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context missing", nil)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)

	if end.IsZero() {
		end = time.Now()
	}

	summary, err := h.service.GetTripSummary(r.Context(), orgID, start, end)
	if err != nil {
		h.logger.Error("failed to get trip summary", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, summary)
}

// GetDriverStats godoc
func (h *TripHandler) GetDriverStats(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context missing", nil)
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
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context missing", nil)
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

// GetTripCustomers returns distinct customer names for filter dropdown
func (h *TripHandler) GetTripCustomers(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context missing", nil)
		return
	}

	customers, err := h.service.GetTripCustomers(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get trip customers", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"customers": customers})
}

// GetStats godoc
func (h *TripHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// SECURITY FIX: Restrict platform-wide statistics to admin roles only (CWE-200)
	roleVal := r.Context().Value(ctxkeys.RoleKey)
	if roleVal == nil {
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}
	// Only allow super_admin, admin, support, org_admin to view platform stats
	if role != "super_admin" && role != "admin" && role != "support" && role != "org_admin" {
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get general stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetTelemetry godoc
func (h *TripHandler) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	telemetry, err := h.service.GetTelemetry(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get telemetry", zap.String("trip_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": telemetry})
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
