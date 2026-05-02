package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
)

type FuelVendorHandler struct {
	svc    *service.FuelVendorService
	logger *zap.Logger
}

func NewFuelVendorHandler(svc *service.FuelVendorService, logger *zap.Logger) *FuelVendorHandler {
	return &FuelVendorHandler{svc: svc, logger: logger}
}

func (h *FuelVendorHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateVendor)
	r.Get("/", h.ListVendors)
}

func (h *FuelVendorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: tenant context required", http.StatusUnauthorized)
		return
	}
	var req models.CreateFuelVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vendor, err := h.svc.CreateVendor(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vendor)
}

func (h *FuelVendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: tenant context required", http.StatusUnauthorized)
		return
	}
	vendors, err := h.svc.ListVendors(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(vendors)
}

func (h *FuelVendorHandler) getTenantID(r *http.Request) uuid.UUID {
	if id, ok := ctxkeys.GetTenantID(r.Context()); ok {
		return id
	}
	h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}
