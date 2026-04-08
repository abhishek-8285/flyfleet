package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/service"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createFuelLogRequestDTO struct {
	VehicleID       uuid.UUID `json:"vehicle_id" validate:"required,uuid"`
	DriverID        uuid.UUID `json:"driver_id" validate:"required,uuid"`
	StationName     string    `json:"station_name" validate:"required,max=255"`
	FuelType        string    `json:"fuel_type" validate:"required,max=50"`
	Quantity        float64   `json:"quantity" validate:"required,gt=0"`
	UnitPrice       float64   `json:"unit_price" validate:"required,gt=0"`
	TotalCost       float64   `json:"total_cost" validate:"required,gt=0"`
	Odometer        float64   `json:"odometer" validate:"required,gte=0"`
	TransactionDate time.Time `json:"transaction_date" validate:"required"`
	GSTNumber       *string   `json:"gst_number,omitempty" validate:"omitempty,len=15"`
	GSTRate         *float64  `json:"gst_rate,omitempty"`
	InvoiceNumber   *string   `json:"invoice_number,omitempty"`
	VendorGSTIN     *string   `json:"vendor_gstin,omitempty" validate:"omitempty,len=15"`
	ReceiptURL      *string   `json:"receipt_url,omitempty" validate:"omitempty,url"`
	Notes           *string   `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type updateFuelLogRequestDTO struct {
	StationName     *string    `json:"station_name,omitempty" validate:"omitempty,max=255"`
	FuelType        *string    `json:"fuel_type,omitempty" validate:"omitempty,max=50"`
	Quantity        *float64   `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice       *float64   `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	TotalCost       *float64   `json:"total_cost,omitempty" validate:"omitempty,gt=0"`
	Odometer        *float64   `json:"odometer,omitempty" validate:"omitempty,gte=0"`
	TransactionDate *time.Time `json:"transaction_date,omitempty"`
	GSTNumber       *string    `json:"gst_number,omitempty" validate:"omitempty,len=15"`
	GSTRate         *float64   `json:"gst_rate,omitempty"`
	InvoiceNumber   *string    `json:"invoice_number,omitempty"`
	VendorGSTIN     *string    `json:"vendor_gstin,omitempty" validate:"omitempty,len=15"`
	ReceiptURL      *string    `json:"receipt_url,omitempty" validate:"omitempty,url"`
	Notes           *string    `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type changeFuelLogStatusRequestDTO struct {
	Status models.FuelLogStatus `json:"status" validate:"required,oneof=APPROVED REJECTED"`
	Notes  *string              `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type FuelHandler struct {
	svc      service.FuelService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewFuelHandler(svc service.FuelService, logger *zap.Logger) *FuelHandler {
	return &FuelHandler{
		svc:      svc,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *FuelHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListLogs)
	r.Post("/", h.CreateLog)
	r.Get("/stats", h.GetStats)
	r.Get("/settings", h.GetSettings)
	r.Put("/settings", h.UpdateSettings)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetLog)
		r.Put("/", h.UpdateLog)
		r.Delete("/", h.DeleteLog)
		r.Patch("/status", h.ChangeStatus)
	})
}

// CreateLog godoc
func (h *FuelHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}

	var dto createFuelLogRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Map DTO to model
	req := models.CreateFuelLogRequest{
		VehicleID:       dto.VehicleID,
		DriverID:        dto.DriverID,
		StationName:     dto.StationName,
		FuelType:        dto.FuelType,
		Quantity:        dto.Quantity,
		UnitPrice:       dto.UnitPrice,
		TotalCost:       dto.TotalCost,
		Odometer:        dto.Odometer,
		TransactionDate: dto.TransactionDate,
		GSTNumber:       dto.GSTNumber,
		GSTRate:         dto.GSTRate,
		InvoiceNumber:   dto.InvoiceNumber,
		VendorGSTIN:     dto.VendorGSTIN,
		ReceiptURL:      dto.ReceiptURL,
		Notes:           dto.Notes,
	}

	log, err := h.svc.CreateLog(r.Context(), orgID, userID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, log)
}

// GetLog godoc
func (h *FuelHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	log, err := h.svc.GetLog(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, log)
}

// ListLogs godoc
func (h *FuelHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}

	query := r.URL.Query()
	filter := models.FuelLogFilter{
		OrganizationID: orgID,
		Search:         query.Get("search"),
	}

	if vID := query.Get("vehicle_id"); vID != "" {
		if u, err := uuid.Parse(vID); err == nil {
			filter.VehicleID = &u
		}
	}

	if dID := query.Get("driver_id"); dID != "" {
		if u, err := uuid.Parse(dID); err == nil {
			filter.DriverID = &u
		}
	}

	status := query["status"]
	for _, s := range status {
		filter.Status = append(filter.Status, models.FuelLogStatus(s))
	}

	res, err := h.svc.ListLogs(r.Context(), filter)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, res)
}

// UpdateLog godoc
func (h *FuelHandler) UpdateLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	var dto updateFuelLogRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Map DTO to model
	req := models.UpdateFuelLogRequest{
		StationName:     dto.StationName,
		FuelType:        dto.FuelType,
		Quantity:        dto.Quantity,
		UnitPrice:       dto.UnitPrice,
		TotalCost:       dto.TotalCost,
		Odometer:        dto.Odometer,
		TransactionDate: dto.TransactionDate,
		GSTNumber:       dto.GSTNumber,
		GSTRate:         dto.GSTRate,
		InvoiceNumber:   dto.InvoiceNumber,
		VendorGSTIN:     dto.VendorGSTIN,
		ReceiptURL:      dto.ReceiptURL,
		Notes:           dto.Notes,
	}

	log, err := h.svc.UpdateLog(r.Context(), id, orgID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, log)
}

// DeleteLog godoc
func (h *FuelHandler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	if err := h.svc.DeleteLog(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangeStatus godoc
func (h *FuelHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	var dto changeFuelLogStatusRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Map DTO to model
	req := models.ChangeFuelLogStatusRequest{
		Status: dto.Status,
		Notes:  dto.Notes,
	}

	if err := h.svc.ChangeStatus(r.Context(), id, orgID, req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "status updated"})
}

// GetStats godoc
func (h *FuelHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	stats, err := h.svc.GetStats(r.Context(), orgID, models.FuelLogFilter{})
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetSettings godoc
func (h *FuelHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	settings, err := h.svc.GetSettings(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, settings)
}

// UpdateSettings godoc
func (h *FuelHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}

	// SECURITY FIX: Check user role - only platform admins can modify fuel policy
	role := h.getUserRole(r)
	if role != "admin" && role != "platform_admin" {
		h.writeErrorResponse(w, http.StatusForbidden, "insufficient permissions: only platform administrators can modify fuel settings", nil)
		return
	}

	var req models.UpdateFuelSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	settings, err := h.svc.UpdateSettings(r.Context(), orgID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, settings)
}

// Helper methods

func (h *FuelHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *FuelHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *FuelHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed fallback to X-Organization-ID header to prevent header spoofing
	// The org ID must come from validated JWT context only
	h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}

func (h *FuelHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed fallback to X-User-ID header to prevent header spoofing
	// The user ID must come from validated JWT context only
	h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}

func (h *FuelHandler) getUserRole(r *http.Request) string {
	if val := r.Context().Value(middleware.RoleKey); val != nil {
		if role, ok := val.(string); ok {
			return role
		}
	}
	// Default should be empty to indicate no role, not "admin"
	h.logger.Warn("Missing role in context - request likely bypassed authentication middleware")
	return ""
}

func (h *FuelHandler) ensureAuthIDs(w http.ResponseWriter, orgID uuid.UUID, userID uuid.UUID) bool {
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "missing organization id", nil)
		return false
	}
	if userID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "missing user id", nil)
		return false
	}
	return true
}
