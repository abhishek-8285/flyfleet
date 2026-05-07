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
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strings"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createVehicleRequestDTO struct {
	RegistrationNumber string     `json:"registration_number" validate:"required"`
	VehicleType        string     `json:"vehicle_type"        validate:"required,oneof=TRUCK BUS VAN CAR TRAILER TANKER TIPPER CRANE OTHER truck bus van car trailer tanker tipper crane other"`
	FuelType           string     `json:"fuel_type"           validate:"required,oneof=PETROL DIESEL CNG ELECTRIC HYBRID LPG petrol diesel cng electric hybrid lpg"`
	OwnershipType      string     `json:"ownership_type"      validate:"required,oneof=OWNED LEASED FINANCED VENDOR owned leased financed vendor"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	GPSEnabled         bool       `json:"gps_enabled"`
	GPSDeviceType      *string    `json:"gps_device_type"`
	ImmobilizerEnabled bool       `json:"immobilizer_enabled"`
	AIS140Compliant    bool       `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year" validate:"omitempty,min=1900,max=2030"`
	Color              *string    `json:"color"`
	Odometer           *int64     `json:"odometer"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"`
	SpeedLimit         *float64   `json:"speed_limit"`
	// Legacy fields (still accepted from old clients)
	LicensePlate *string `json:"license_plate"`
}

type updateVehicleRequestDTO struct {
	VehicleType        *string    `json:"vehicle_type"      validate:"omitempty,oneof=TRUCK BUS VAN CAR TRAILER TANKER TIPPER CRANE OTHER truck bus van car trailer tanker tipper crane other"`
	FuelType           *string    `json:"fuel_type"         validate:"omitempty,oneof=PETROL DIESEL CNG ELECTRIC HYBRID LPG petrol diesel cng electric hybrid lpg"`
	OwnershipType      *string    `json:"ownership_type"    validate:"omitempty,oneof=OWNED LEASED FINANCED VENDOR owned leased financed vendor"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	GPSEnabled         *bool      `json:"gps_enabled"`
	GPSDeviceType      *string    `json:"gps_device_type"`
	ImmobilizerEnabled *bool      `json:"immobilizer_enabled"`
	AIS140Compliant    *bool      `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year"    validate:"omitempty,min=1900,max=2030"`
	Color              *string    `json:"color"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"`
	// Legacy fields
	Status    *string `json:"status"   validate:"omitempty,oneof=ACTIVE MAINTENANCE SUSPENDED DECOMMISSIONED active inactive maintenance retired"`
	Odometer  *int64  `json:"odometer" validate:"omitempty,min=0"`
	Insurance *string `json:"insurance"`
	SpeedLimit *float64 `json:"speed_limit"`
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
		r.Post("/unassign-driver", h.UnassignDriver)
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
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID not found in context", nil)
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
		GPSEnabled:         dto.GPSEnabled,
		GPSDeviceType:      dto.GPSDeviceType,
		ImmobilizerEnabled: dto.ImmobilizerEnabled,
		AIS140Compliant:    dto.AIS140Compliant,
		Make:               dto.Make,
		Model:              dto.Model,
		Year:               dto.Year,
		Color:              dto.Color,
		Odometer:           dto.Odometer,
		CurrentDriverID:    dto.CurrentDriverID,
		SpeedLimit:         0, // Default
		LicensePlate:       dto.LicensePlate,
	}

	if dto.SpeedLimit != nil {
		req.SpeedLimit = *dto.SpeedLimit
	}

	vehicle, err := h.service.CreateVehicle(r.Context(), req, tenantID)
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
	tenantID := h.getTenantID(r.Context())
	if vehicle.TenantID != tenantID {
		h.writeErrorResponse(w, http.StatusNotFound, "Vehicle not found or access denied", nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": vehicle})
}

// ListVehicles godoc
func (h *VehicleHandlers) ListVehicles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	tenantID, _ := ctxkeys.GetTenantID(r.Context())
	role := ctxkeys.GetRole(r.Context())

	// Only super_admin can override tenant_id via query param
	if role == "super_admin" {
		if tidStr := query.Get("tenant_id"); tidStr != "" {
			if tid, err := uuid.Parse(tidStr); err == nil {
				tenantID = tid
			}
		}
	}

	if tenantID == uuid.Nil && role != "super_admin" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant context required", nil)
		return
	}

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	vehicles, total, err := h.service.ListVehicles(r.Context(), tenantID, page, limit)
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
		GPSEnabled:         dto.GPSEnabled,
		GPSDeviceType:      dto.GPSDeviceType,
		ImmobilizerEnabled: dto.ImmobilizerEnabled,
		AIS140Compliant:    dto.AIS140Compliant,
		Make:               dto.Make,
		Model:              dto.Model,
		Year:               dto.Year,
		Color:              dto.Color,
		CurrentDriverID:    dto.CurrentDriverID,
		Status:             dto.Status,
		Odometer:           dto.Odometer,
		Insurance:          dto.Insurance,
		SpeedLimit:         dto.SpeedLimit,
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

	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID required", nil)
		return
	}

	err = h.service.DeleteVehicle(r.Context(), vehicleID, tenantID)
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

	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID required", nil)
		return
	}

	var dto assignDriverRequestDTO
	if err := h.decodeAndValidate(r.Body, &dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	vehicle, err := h.service.AssignDriver(r.Context(), vehicleID, tenantID, dto.DriverID)
	if err != nil {
		// Pass through custom error messages (e.g. "please unmap first")
		h.writeErrorResponse(w, http.StatusConflict, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vehicle)
}

// UnassignDriver godoc
func (h *VehicleHandlers) UnassignDriver(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", err)
		return
	}

	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID required", nil)
		return
	}

	vehicle, err := h.service.UnassignDriver(r.Context(), vehicleID, tenantID)
	if err != nil {
		h.logger.Error("failed to unassign driver", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to unassign driver", err)
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
		Status:      models.VehicleStatus(strings.ToUpper(q.Get("status"))),
		VehicleType: models.VehicleType(strings.ToUpper(q.Get("vehicle_type"))),
		FuelType:    models.FuelType(strings.ToUpper(q.Get("fuel_type"))),
	}

	h.logger.Info("Vehicle Search Triggered",
		zap.String("q", params.Query),
		zap.Any("gps_enabled", params.GPSEnabled),
	)

	if gpsEnabledStr := q.Get("gps_enabled"); gpsEnabledStr != "" {
		if b, err := strconv.ParseBool(gpsEnabledStr); err == nil {
			params.GPSEnabled = &b
		}
	}

	if tenantIDStr := q.Get("tenant_id"); tenantIDStr != "" {
		if id, err := uuid.Parse(tenantIDStr); err == nil {
			params.TenantID = &id
		}
	} else {
		tenantID := h.getTenantID(r.Context())
		if tenantID != uuid.Nil {
			params.TenantID = &tenantID
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
	tenantID := h.getTenantID(r.Context())
	vehicles, total, err := h.service.ListVehicles(r.Context(), tenantID, page, limit)
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
	resp := map[string]interface{}{"error": message}
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			details := make(map[string]string)
			for _, f := range ve {
				details[f.Field()] = fmt.Sprintf("failed validation: %s", f.Tag())
			}
			resp["details"] = details
		} else {
			resp["details"] = err.Error()
		}
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *VehicleHandlers) getTenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctxkeys.GetTenantID(ctx); ok {
		return id
	}
	return uuid.Nil
}
