package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	var event models.ConsumptionEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Ensure TenantID is set from context if missing
	if event.TenantID == uuid.Nil {
		event.TenantID = h.getTenantID(r)
	}

	err := h.meterService.RecordUsage(r.Context(), &event)
	if err != nil {
		h.logger.Error("Failed to record usage", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Usage recorded successfully"}, http.StatusCreated)
}

func (h *BillingHandler) GenerateDraftInvoice(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.service.GenerateDraftInvoice(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to generate draft invoice", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, invoice, http.StatusCreated)
}
