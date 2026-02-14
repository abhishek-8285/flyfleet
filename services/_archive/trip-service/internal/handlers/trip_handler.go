package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/flyfleet/trip-service/internal/service"
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

// CreateTrip godoc
// @Summary Create a new trip
// @Description Create a new trip with planned route and waypoints
// @Tags trips
// @Accept json
// @Produce json
// @Param trip body models.CreateTripRequest true "Trip creation request"
// @Success 201 {object} models.Trip
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips [post]
func (h *TripHandler) CreateTrip(c *gin.Context) {
	var req models.CreateTripRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get org_id from context (would come from auth middleware)
	orgID := c.GetString("org_id")
	if orgID == "" {
		orgID = "550e8400-e29b-41d4-a716-446655440002" // fallback for testing
	}

	trip, err := h.service.Create(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create trip", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, trip)
}

// GetTrip godoc
// @Summary Get trip by ID
// @Description Retrieve a trip by its ID
// @Tags trips
// @Produce json
// @Param id path string true "Trip ID"
// @Success 200 {object} models.Trip
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id} [get]
func (h *TripHandler) GetTrip(c *gin.Context) {
	id := c.Param("id")
	
	trip, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get trip", zap.String("trip_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// ListTrips godoc
// @Summary List trips
// @Description List trips with filtering and pagination
// @Tags trips
// @Produce json
// @Param status query string false "Trip status filter"
// @Param vehicle_id query string false "Vehicle ID filter"
// @Param driver_id query string false "Driver ID filter"
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /trips [get]
func (h *TripHandler) ListTrips(c *gin.Context) {
	var filter models.TripListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.Page < 1 {
		filter.Page = 1
	}

	trips, totalCount, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list trips", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       trips,
		"total":      totalCount,
		"page":       filter.Page,
		"page_size":  filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateTrip godoc
// @Summary Update trip
// @Description Update an existing trip (only for planned trips)
// @Tags trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Param trip body models.UpdateTripRequest true "Trip update request"
// @Success 200 {object} models.Trip
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id} [put]
func (h *TripHandler) UpdateTrip(c *gin.Context) {
	id := c.Param("id")
	
	var req models.UpdateTripRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trip, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// DeleteTrip godoc
// @Summary Delete trip
// @Description Delete a planned trip
// @Tags trips
// @Produce json
// @Param id path string true "Trip ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id} [delete]
func (h *TripHandler) DeleteTrip(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartTrip godoc
// @Summary Start trip
// @Description Start a planned or paused trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Param request body models.StartTripRequest true "Start trip request"
// @Success 200 {object} models.Trip
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/start [post]
func (h *TripHandler) StartTrip(c *gin.Context) {
	id := c.Param("id")
	
	var req models.StartTripRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get driver_id from context (would come from auth middleware)
	driverID := c.GetString("driver_id")
	if driverID == "" {
		driverID = "550e8400-e29b-41d4-a716-446655440001" // fallback for testing
	}

	trip, err := h.service.Start(c.Request.Context(), id, driverID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// PauseTrip godoc
// @Summary Pause trip
// @Description Pause an in-progress trip
// @Tags trips
// @Produce json
// @Param id path string true "Trip ID"
// @Success 200 {object} models.Trip
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/pause [post]
func (h *TripHandler) PauseTrip(c *gin.Context) {
	id := c.Param("id")
	
	// Get driver_id from context
	driverID := c.GetString("driver_id")
	if driverID == "" {
		driverID = "default-driver"
	}

	trip, err := h.service.Pause(c.Request.Context(), id, driverID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// ResumeTrip godoc
// @Summary Resume trip
// @Description Resume a paused trip
// @Tags trips
// @Produce json
// @Param id path string true "Trip ID"
// @Success 200 {object} models.Trip
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/resume [post]
func (h *TripHandler) ResumeTrip(c *gin.Context) {
	id := c.Param("id")
	
	// Get driver_id from context
	driverID := c.GetString("driver_id")
	if driverID == "" {
		driverID = "default-driver"
	}

	trip, err := h.service.Resume(c.Request.Context(), id, driverID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// CompleteTrip godoc
// @Summary Complete trip
// @Description Complete an in-progress trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Param request body models.CompleteTripRequest true "Complete trip request"
// @Success 200 {object} models.Trip
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/complete [post]
func (h *TripHandler) CompleteTrip(c *gin.Context) {
	id := c.Param("id")
	
	var req models.CompleteTripRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get driver_id from context
	driverID := c.GetString("driver_id")
	if driverID == "" {
		driverID = "default-driver"
	}

	trip, err := h.service.Complete(c.Request.Context(), id, driverID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// CancelTrip godoc
// @Summary Cancel trip
// @Description Cancel a trip
// @Tags trips
// @Produce json
// @Param id path string true "Trip ID"
// @Param reason query string true "Cancellation reason"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/cancel [post]
func (h *TripHandler) CancelTrip(c *gin.Context) {
	id := c.Param("id")
	reason := c.Query("reason")
	
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	if err := h.service.Cancel(c.Request.Context(), id, reason); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddEvent godoc
// @Summary Add trip event
// @Description Add an event to a trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Param event body models.AddEventRequest true "Add event request"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/events [post]
func (h *TripHandler) AddEvent(c *gin.Context) {
	id := c.Param("id")
	
	var req models.AddEventRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get driver_id from context
	driverID := c.GetString("driver_id")
	if driverID == "" {
		driverID = "default-driver"
	}

	if err := h.service.AddEvent(c.Request.Context(), id, driverID, &req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SubmitPOD godoc
// @Summary Submit proof of delivery
// @Description Submit proof of delivery for a completed trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Param pod body models.SubmitPODRequest true "Submit POD request"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/pod [post]
func (h *TripHandler) SubmitPOD(c *gin.Context) {
	id := c.Param("id")
	
	var req models.SubmitPODRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get driver_id from context
	driverID := c.GetString("driver_id")
	if driverID == "" {
		driverID = "default-driver"
	}

	if err := h.service.SubmitPOD(c.Request.Context(), id, driverID, &req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "trip not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// OptimizeRoute godoc
// @Summary Optimize trip route
// @Description Get optimized route options for a trip
// @Tags trips
// @Produce json
// @Param id path string true "Trip ID"
// @Param avoid_tolls query bool false "Avoid toll roads"
// @Param traffic_aware query bool false "Consider traffic conditions"
// @Success 200 {object} models.RouteOptimizationResult
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trips/{id}/optimize-route [get]
func (h *TripHandler) OptimizeRoute(c *gin.Context) {
	// id := c.Param("id") // Not used in mock implementation
	
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

	c.JSON(http.StatusOK, result)
}

// GetDriverStats godoc
// @Summary Get driver statistics
// @Description Get performance metrics for drivers
// @Tags analytics
// @Produce json
// @Success 200 {object} []map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /stats/drivers [get]
func (h *TripHandler) GetDriverStats(c *gin.Context) {
	orgID := c.GetString("org_id")
	if orgID == "" {
		orgID = "default-org"
	}

	stats, err := h.service.GetDriverStats(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get driver stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTripStats godoc
// @Summary Get trip statistics
// @Description Get daily trip counts and trends
// @Tags analytics
// @Produce json
// @Success 200 {object} []map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /stats/trips [get]
func (h *TripHandler) GetTripStats(c *gin.Context) {
	orgID := c.GetString("org_id")
	if orgID == "" {
		orgID = "default-org"
	}

	stats, err := h.service.GetTripStats(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get trip stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *TripHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get trip stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators if needed
	// validate.RegisterValidation("custom_rule", customValidationFunc)
}
