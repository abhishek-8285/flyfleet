package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) ListHardwareProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.hwService.ListProducts(r.Context())
	if err != nil {
		h.logger.Error("Failed to list hardware products", zap.Error(err))
		h.errorJSON(w, r, "Failed to list hardware products: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": products}, http.StatusOK)
}

func (h *BillingHandler) AssignHardware(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var req models.DeviceAssignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	req.TenantID = tenantID

	if err := h.validator.Struct(&req); err != nil {
		h.errorJSON(w, r, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.hwService.AssignHardware(r.Context(), &req); err != nil {
		h.logger.Error("Failed to assign hardware", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, req, http.StatusCreated)
}
