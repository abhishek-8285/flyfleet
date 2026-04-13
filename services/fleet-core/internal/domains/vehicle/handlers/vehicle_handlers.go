package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/service"
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createVehicleRequestDTO struct {
	RegistrationNumber string     `json:"registration_number" validate:"required"`
	VehicleType        string     `json:"vehicle_type"        validate:"required,oneof=TRUCK BUS VAN CAR TRAILER TANKER TIPPER CRANE OTHER"`
	FuelType           string     `json:"fuel_type"           validate:"required,oneof=PETROL DIESEL CNG ELECTRIC HYBRID LPG"`
	OwnershipType      string     `json:"ownership_type"      validate:"required,oneof=OWNED LEASED FINANCED VENDOR"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	ImmobilizerEnabled bool       `json:"immobilizer_enabled"`
	AIS140Compliant    bool       `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year" validate:"omitempty,min=1900,max=2030"`
	Color              *string    `json:"color"`
	// Legacy fields (still accepted from old clients)
	LicensePlate *string `json:"license_plate"`
}

type updateVehicleRequestDTO struct {
	VehicleType        *string    `json:"vehicle_type"      validate:"omitempty,oneof=TRUCK BUS VAN CAR TRAILER TANKER TIPPER CRANE OTHER"`
	FuelType           *string    `json:"fuel_type"         validate:"omitempty,oneof=PETROL DIESEL CNG ELECTRIC HYBRID LPG"`
	OwnershipType      *string    `json:"ownership_type"    validate:"omitempty,oneof=OWNED LEASED FINANCED VENDOR"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	ImmobilizerEnabled *bool      `json:"immobilizer_enabled"`
	AIS140Compliant    *bool      `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year"    validate:"omitempty,min=1900,max=2030"`
	Color              *string    `json:"color"`
	// Legacy fields
	Status    *string `json:"status"   validate:"omitempty,oneof=ACTIVE MAINTENANCE SUSPENDED DECOMMISSIONED active inactive maintenance retired"`
	Odometer  *int64  `json:"odometer" validate:"omitempty,min=0"`
	Insurance *string `json:"insurance"`
}

type assignDriverRequestDTO struct {
	DriverID uuid.UUID `json:"driver_id" validate:"required"`
}

type VehicleHandlers struct {
	service  service.VehicleServiceInterface
	logger   *zap.Logger
	validate *validator.Validate
}

func NewVehicleHandlers(service service.VehicleServiceInterface, logger *zap.Logger) *VehicleHandlers {
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
	r.Get("/search", h.SearchVehicles)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetVehicle)
		r.Put("/", h.UpdateVehicle)
		r.Delete("/", h.DeleteVehicle)
		r.Post("/assign-driver", h.AssignDriver)
	})
}

// CreateVehicle godoc
func (h *VehicleHandlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var dto createVehicleRequestDTO
	if err := h.decodeAndValidate(r.Body, &dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get Organization ID from context (set by auth middleware)
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found in context", nil)
		return
	}

	// Map DTO to model
	req := &models.CreateVehicleRequest{
		RegistrationNumber: dto.RegistrationNumber,
		VehicleType:        dto.VehicleType,
		FuelType:           dto.FuelType,
		OwnershipType:      dto.OwnershipType,
		CapacityTonnage:    dto.CapacityTonnage,
		FuelTankCapacity:   dto.FuelTankCapacity,
		ChassisNumber:      dto.ChassisNumber,
		EngineNumber:       dto.EngineNumber,
		VINNumber:          dto.VINNumber,
		ParentVehicleID:    dto.ParentVehicleID,
		BranchID:           dto.BranchID,
		CostCenterID:       dto.CostCenterID,
		GPSDeviceID:        dto.GPSDeviceID,
		ImmobilizerEnabled: dto.ImmobilizerEnabled,
		AIS140Compliant:    dto.AIS140Compliant,
		Make:               dto.Make,
		Model:              dto.Model,
		Year:               dto.Year,
		Color:              dto.Color,
		LicensePlate:       dto.LicensePlate,
	}

	vehicle, err := h.service.CreateVehicle(r.Context(), req, orgID)
	if err != nil {
		if errors.Is(err, service.ErrVehicleAlreadyExists) {
			h.writeErrorResponse(w, http.StatusConflict, "Vehicle with this registration number already exists", nil)
			return
		}
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

	// SECURITY FIX: Verify vehicle belongs to authenticated organization (CWE-639)
	orgID := h.getOrgID(r)
	if vehicle.TenantID != orgID {
		h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found or access denied", nil)
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

	var dto updateVehicleRequestDTO
	if err := h.decodeAndValidate(r.Body, &dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Map DTO to model
	req := &models.UpdateVehicleRequest{
		VehicleType:        dto.VehicleType,
		FuelType:           dto.FuelType,
		OwnershipType:      dto.OwnershipType,
		CapacityTonnage:    dto.CapacityTonnage,
		FuelTankCapacity:   dto.FuelTankCapacity,
		ChassisNumber:      dto.ChassisNumber,
		EngineNumber:       dto.EngineNumber,
		VINNumber:          dto.VINNumber,
		BranchID:           dto.BranchID,
		CostCenterID:       dto.CostCenterID,
		GPSDeviceID:        dto.GPSDeviceID,
		ImmobilizerEnabled: dto.ImmobilizerEnabled,
		AIS140Compliant:    dto.AIS140Compliant,
		Make:               dto.Make,
		Model:              dto.Model,
		Year:               dto.Year,
		Color:              dto.Color,
		Status:             dto.Status,
		Odometer:           dto.Odometer,
		Insurance:          dto.Insurance,
	}

	vehicle, err := h.service.UpdateVehicle(r.Context(), vehicleID, req)
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

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	err = h.service.DeleteVehicle(r.Context(), vehicleID, orgID)
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

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	var dto assignDriverRequestDTO
	if err := h.decodeAndValidate(r.Body, &dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	vehicle, err := h.service.AssignDriver(r.Context(), vehicleID, orgID, dto.DriverID)
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

// SearchVehicles godoc — GET /vehicles/search?q=&status=&vehicle_type=&fuel_type=&page=&limit=
func (h *VehicleHandlers) SearchVehicles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := models.VehicleSearchParams{
		Query:       q.Get("q"),
		Status:      q.Get("status"),
		VehicleType: q.Get("vehicle_type"),
		FuelType:    q.Get("fuel_type"),
	}

	if tenantIDStr := q.Get("tenant_id"); tenantIDStr != "" {
		if id, err := uuid.Parse(tenantIDStr); err == nil {
			params.TenantID = &id
		}
	} else {
		orgID := h.getOrgID(r)
		if orgID != uuid.Nil {
			params.TenantID = &orgID
		}
	}

	if branchIDStr := q.Get("branch_id"); branchIDStr != "" {
		if id, err := uuid.Parse(branchIDStr); err == nil {
			params.BranchID = &id
		}
	}

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	params.Page = page
	params.Limit = limit

	// Use SearchVehicles if the service impl supports it; fall back to list
	type searchable interface {
		SearchVehicles(ctx context.Context, params models.VehicleSearchParams) ([]models.Vehicle, int64, error)
	}
	if svc, ok := h.service.(searchable); ok {
		vehicles, total, err := svc.SearchVehicles(r.Context(), params)
		if err != nil {
			h.logger.Error("failed to search vehicles", zap.Error(err))
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search vehicles", err)
			return
		}
		h.writeJSONResponse(w, http.StatusOK, models.VehicleListResponse{
			Vehicles: vehicles,
			Total:    total,
			Page:     page,
			Limit:    limit,
		})
		return
	}
	// Fallback: regular list
	orgID := h.getOrgID(r)
	vehicles, total, err := h.service.ListVehicles(r.Context(), orgID, page, limit)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search vehicles", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, models.VehicleListResponse{
		Vehicles: vehicles,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
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
	// Try CtxOrgID (set by fleet-core JWKS middleware) as UUID
	if val := r.Context().Value(middleware.CtxOrgID); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	// Fallback: ctxkeys.OrgIDKey (set as string by JWKS)
	if val := r.Context().Value(ctxkeys.OrgIDKey); val != nil {
		if s, ok := val.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
