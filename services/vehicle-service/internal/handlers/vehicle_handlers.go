package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VehicleHandlers struct {
	service  *service.VehicleService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewVehicleHandlers(service *service.VehicleService, logger *zap.Logger) *VehicleHandlers {
	return &VehicleHandlers{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// CreateVehicle godoc
// @Summary Create a new vehicle
// @Description Create a new vehicle record for the organization
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle body models.CreateVehicleRequest true "Vehicle data"
// @Success 201 {object} models.Vehicle
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/vehicles [post]
// @Security BearerAuth
func (h *VehicleHandlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateVehicleRequest
	if err := h.decodeAndValidate(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Extract org_id from context (would come from auth middleware)
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Placeholder

	ctx := context.Background()
	vehicle, err := h.service.CreateVehicle(ctx, &req, orgID)
	if err != nil {
		h.logger.Error("failed to create vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, vehicle)
}

// GetVehicle godoc
// @Summary Get vehicle by ID
// @Description Get detailed information about a specific vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} models.Vehicle
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/vehicles/{id} [get]
// @Security BearerAuth
func (h *VehicleHandlers) GetVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vehicle ID from URL path
	vehicleIDStr := r.URL.Path[len("/api/v1/vehicles/"):]
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	ctx := context.Background()
	vehicle, err := h.service.GetVehicle(ctx, vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to get vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vehicle)
}

// ListVehicles godoc
// @Summary List vehicles
// @Description Get paginated list of vehicles for the organization
// @Tags vehicles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.VehicleListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/vehicles [get]
// @Security BearerAuth
func (h *VehicleHandlers) ListVehicles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract org_id from context
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Placeholder

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	ctx := context.Background()
	vehicles, total, err := h.service.ListVehicles(ctx, orgID, page, limit)
	if err != nil {
		h.logger.Error("failed to list vehicles", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list vehicles", err)
		return
	}

	response := models.VehicleListResponse{
		Vehicles: vehicles,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateVehicle godoc
// @Summary Update vehicle
// @Description Update an existing vehicle's information
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param vehicle body models.UpdateVehicleRequest true "Vehicle updates"
// @Success 200 {object} models.Vehicle
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/vehicles/{id} [put]
// @Security BearerAuth
func (h *VehicleHandlers) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vehicle ID from URL path
	vehicleIDStr := r.URL.Path[len("/api/v1/vehicles/"):]
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	var req models.UpdateVehicleRequest
	if err := h.decodeAndValidate(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := context.Background()
	vehicle, err := h.service.UpdateVehicle(ctx, vehicleID, &req)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to update vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vehicle)
}

// DeleteVehicle godoc
// @Summary Delete vehicle
// @Description Permanently delete a vehicle record
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/vehicles/{id} [delete]
// @Security BearerAuth
func (h *VehicleHandlers) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vehicle ID from URL path
	vehicleIDStr := r.URL.Path[len("/api/v1/vehicles/"):]
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	ctx := context.Background()
	err = h.service.DeleteVehicle(ctx, vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to delete vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete vehicle", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignDriver godoc
// @Summary Assign driver to vehicle
// @Description Assign a driver to operate a specific vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param assignment body models.AssignDriverRequest true "Driver assignment"
// @Success 200 {object} models.Vehicle
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/vehicles/{id}/assign-driver [post]
// @Security BearerAuth
func (h *VehicleHandlers) AssignDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vehicle ID from URL path
	vehicleIDStr := r.URL.Path[len("/api/v1/vehicles/") : len("/api/v1/vehicles/") + 36] // UUID length
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	var req models.AssignDriverRequest
	if err := h.decodeAndValidate(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := context.Background()
	vehicle, err := h.service.AssignDriver(ctx, vehicleID, req.DriverID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to assign driver", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to assign driver", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vehicle)
}

// GetVehicleLocation godoc
// @Summary Get vehicle location
// @Description Get current location of a vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 200 {object} models.VehicleLocation
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/vehicles/{id}/location [get]
// @Security BearerAuth
func (h *VehicleHandlers) GetVehicleLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract vehicle ID from URL path
	vehicleIDStr := r.URL.Path[len("/api/v1/vehicles/") : len("/api/v1/vehicles/") + 36]
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	ctx := context.Background()
	location, err := h.service.GetVehicleLocation(ctx, vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to get vehicle location", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get vehicle location", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, location)
}

func (h *VehicleHandlers) decodeAndValidate(body io.ReadCloser, dest interface{}) error {
	decoder := json.NewDecoder(body)
	defer body.Close()
	
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	if err := h.validate.Struct(dest); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func (h *VehicleHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *VehicleHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"error": message,
	}
	
	if err != nil {
		h.logger.Debug("error details", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}