package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"
	"github.com/go-chi/chi/v5"
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

func (h *VehicleHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListVehicles)
	r.Post("/", h.CreateVehicle)
	r.Get("/stats", h.GetStats)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetVehicle)
		r.Put("/", h.UpdateVehicle)
		r.Delete("/", h.DeleteVehicle)
		r.Post("/assign-driver", h.AssignDriver)
	})
}

// CreateVehicle godoc
func (h *VehicleHandlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req models.CreateVehicleRequest
	if err := h.decodeAndValidate(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get Organization ID from context (set by auth middleware)
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found in context", nil)
		return
	}

	vehicle, err := h.service.CreateVehicle(r.Context(), &req, orgID)
	if err != nil {
		h.logger.Error("failed to create vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": vehicle})
}

// GetVehicle godoc
func (h *VehicleHandlers) GetVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	vehicle, err := h.service.GetVehicle(r.Context(), vehicleID)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to get vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": vehicle})
}

// ListVehicles godoc
func (h *VehicleHandlers) ListVehicles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid org_id format", err)
			return
		}
	} else {
		// Fallback to context if not in query
		orgID = h.getOrgID(r)
	}

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	vehicles, total, err := h.service.ListVehicles(r.Context(), orgID, page, limit)
	if err != nil {
		h.logger.Error("failed to list vehicles", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list vehicles", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.VehicleListResponse{
		Vehicles: vehicles,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// UpdateVehicle godoc
func (h *VehicleHandlers) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	var req models.UpdateVehicleRequest
	if err := h.decodeAndValidate(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	vehicle, err := h.service.UpdateVehicle(r.Context(), vehicleID, &req)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found", nil)
			return
		}
		h.logger.Error("failed to update vehicle", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": vehicle})
}

// DeleteVehicle godoc
func (h *VehicleHandlers) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	err = h.service.DeleteVehicle(r.Context(), vehicleID)
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
func (h *VehicleHandlers) AssignDriver(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	var req models.AssignDriverRequest
	if err := h.decodeAndValidate(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	vehicle, err := h.service.AssignDriver(r.Context(), vehicleID, req.DriverID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to assign driver", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vehicle)
}

// GetStats godoc
func (h *VehicleHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get vehicle stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get vehicle stats", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

func (h *VehicleHandlers) decodeAndValidate(body io.ReadCloser, dest interface{}) error {
	decoder := json.NewDecoder(body)
	defer body.Close()
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return h.validate.Struct(dest)
}

func (h *VehicleHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *VehicleHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	h.writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

func (h *VehicleHandlers) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}