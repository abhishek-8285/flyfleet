package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/service"
	"github.com/gin-gonic/gin"
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
func (h *VehicleHandlers) CreateVehicle(c *gin.Context) {
	var req models.CreateVehicleRequest
	if err := h.decodeAndValidate(c.Request.Body, &req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Placeholder

	vehicle, err := h.service.CreateVehicle(c.Request.Context(), &req, orgID)
	if err != nil {
		h.logger.Error("failed to create vehicle", zap.Error(err))
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to create vehicle", err)
		return
	}

	h.writeJSONResponse(c, http.StatusCreated, vehicle)
}

// GetVehicle godoc
func (h *VehicleHandlers) GetVehicle(c *gin.Context) {
	idStr := c.Param("id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	vehicle, err := h.service.GetVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(c, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to get vehicle", zap.Error(err))
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get vehicle", err)
		return
	}

	h.writeJSONResponse(c, http.StatusOK, vehicle)
}

// ListVehicles godoc
func (h *VehicleHandlers) ListVehicles(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.writeErrorResponse(c, http.StatusBadRequest, "Invalid org_id format", err)
			return
		}
	} else {
		orgID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	
	if page <= 0 { page = 1 }
	if limit <= 0 || limit > 100 { limit = 20 }

	vehicles, total, err := h.service.ListVehicles(c.Request.Context(), orgID, page, limit)
	if err != nil {
		h.logger.Error("failed to list vehicles", zap.Error(err))
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to list vehicles", err)
		return
	}

	h.writeJSONResponse(c, http.StatusOK, models.VehicleListResponse{
		Vehicles: vehicles,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// UpdateVehicle godoc
func (h *VehicleHandlers) UpdateVehicle(c *gin.Context) {
	idStr := c.Param("id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	var req models.UpdateVehicleRequest
	if err := h.decodeAndValidate(c.Request.Body, &req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	vehicle, err := h.service.UpdateVehicle(c.Request.Context(), vehicleID, &req)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(c, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to update vehicle", zap.Error(err))
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to update vehicle", err)
		return
	}

	h.writeJSONResponse(c, http.StatusOK, vehicle)
}

// DeleteVehicle godoc
func (h *VehicleHandlers) DeleteVehicle(c *gin.Context) {
	idStr := c.Param("id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	err = h.service.DeleteVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(c, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to delete vehicle", zap.Error(err))
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to delete vehicle", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignDriver godoc
func (h *VehicleHandlers) AssignDriver(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	var req models.AssignDriverRequest
	if err := h.decodeAndValidate(c.Request.Body, &req); err != nil {
		h.writeErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	vehicle, err := h.service.AssignDriver(c.Request.Context(), vehicleID, req.DriverID)
	if err != nil {
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to assign driver", err)
		return
	}

	h.writeJSONResponse(c, http.StatusOK, vehicle)
}

// GetStats godoc
func (h *VehicleHandlers) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get vehicle stats", zap.Error(err))
		h.writeErrorResponse(c, http.StatusInternalServerError, "Failed to get vehicle stats", err)
		return
	}

	h.writeJSONResponse(c, http.StatusOK, stats)
}

func (h *VehicleHandlers) decodeAndValidate(body io.ReadCloser, dest interface{}) error {
	decoder := json.NewDecoder(body)
	defer body.Close()
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return h.validate.Struct(dest)
}

func (h *VehicleHandlers) writeJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func (h *VehicleHandlers) writeErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	c.JSON(statusCode, gin.H{"error": message})
}