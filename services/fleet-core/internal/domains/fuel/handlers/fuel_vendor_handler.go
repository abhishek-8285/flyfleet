package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
	var req models.CreateFuelVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vendor, err := h.svc.CreateVendor(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vendor)
}

func (h *FuelVendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
	vendors, err := h.svc.ListVendors(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(vendors)
}

func (h *FuelVendorHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed fallback to X-Organization-ID header to prevent header spoofing
	// The org ID must come from validated JWT context only
	h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}
