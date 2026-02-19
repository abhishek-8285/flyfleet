package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
)

// VendorHandler handles HTTP requests for maintenance vendors
type VendorHandler struct {
	repo   *repository.VendorRepository
	logger *zap.Logger
}

// NewVendorHandler creates a new vendor handler
func NewVendorHandler(repo *repository.VendorRepository, logger *zap.Logger) *VendorHandler {
	return &VendorHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *VendorHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListVendors)
	r.Post("/", h.CreateVendor)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetVendor)
		r.Put("/", h.UpdateVendor)
		r.Delete("/", h.DeleteVendor)
	})
}

func (h *VendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	search := r.URL.Query().Get("search")
	specialization := r.URL.Query().Get("specialization")

	vendors, err := h.repo.List(r.Context(), orgID, search, specialization)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vendors)
}

func (h *VendorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)

	var vendor models.MaintenanceVendor
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	vendor.OrganizationID = orgID

	if err := h.repo.Create(r.Context(), &vendor); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, vendor)
}

func (h *VendorHandler) GetVendor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid vendor ID")
		return
	}

	vendor, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if vendor == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "vendor not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vendor)
}

func (h *VendorHandler) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid vendor ID")
		return
	}

	var vendor models.MaintenanceVendor
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	vendor.ID = id

	if err := h.repo.Update(r.Context(), &vendor); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, vendor)
}

func (h *VendorHandler) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid vendor ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusNoContent, nil)
}

// Helpers
func (h *VendorHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *VendorHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

func (h *VendorHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}
