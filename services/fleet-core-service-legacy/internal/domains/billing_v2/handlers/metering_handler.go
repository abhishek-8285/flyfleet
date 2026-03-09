package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	var event models.ConsumptionEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Ensure OrgID is set from context if missing
	if event.OrgID == uuid.Nil {
		event.OrgID = h.getOrgID(r)
	}

	err := h.meterService.RecordUsage(r.Context(), &event)
	if err != nil {
		h.logger.Error("Failed to record usage", zap.Error(err))
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Usage recorded successfully"}, http.StatusCreated)
}

func (h *BillingHandler) GenerateDraftInvoice(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.service.GenerateDraftInvoice(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to generate draft invoice", zap.Error(err))
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, invoice, http.StatusCreated)
}
