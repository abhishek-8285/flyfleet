package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
	currency := r.URL.Query().Get("currency")
	if currency == "" {
		currency = "USD"
	}

	balance, err := h.ledgerService.GetBalance(r.Context(), orgID, currency)
	if err != nil {
		h.logger.Error("Failed to get balance", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, balance, http.StatusOK)
}

func (h *BillingHandler) ApplyCredit(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}

	var req struct {
		Amount   float64 `json:"amount" validate:"required"`
		Currency string  `json:"currency"`
		Type     string  `json:"type"`
		Desc     string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Type == "" {
		req.Type = "MANUAL"
	}

	if err := h.ledgerService.ApplyCredit(r.Context(), orgID, req.Amount, req.Currency, req.Type, nil, &req.Desc); err != nil {
		h.logger.Error("Failed to apply credit", zap.Error(err))
		h.errorJSON(w, r, "Failed to apply credit", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "credit applied successfully"}, http.StatusOK)
}
