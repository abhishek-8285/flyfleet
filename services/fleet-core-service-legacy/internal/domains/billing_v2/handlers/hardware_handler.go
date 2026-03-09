package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"go.uber.org/zap"
	"github.com/google/uuid"
)

func (h *BillingHandler) ListHardwareProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.hwService.ListProducts(r.Context())
	if err != nil {
		h.logger.Error("Failed to list hardware products", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": products}, http.StatusOK)
}

func (h *BillingHandler) AssignHardware(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var req models.DeviceAssignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	req.OrgID = orgID

	if err := h.validator.Struct(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.hwService.AssignHardware(r.Context(), &req); err != nil {
		h.logger.Error("Failed to assign hardware", zap.Error(err))
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, req, http.StatusCreated)
}
