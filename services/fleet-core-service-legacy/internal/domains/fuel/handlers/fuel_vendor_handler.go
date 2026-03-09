package handlers

import (
	"encoding/json"
	"net/http"

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
	vendors, err := h.svc.ListVendors(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(vendors)
}

func (h *FuelVendorHandler) getOrgID(r *http.Request) uuid.UUID {
	// Simple helper for now, assuming orgID is in header or context
	val := r.Header.Get("X-Organization-ID")
	id, _ := uuid.Parse(val)
	return id
}
