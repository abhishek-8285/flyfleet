package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) ProrateUpgrade(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewPlanID uuid.UUID `json:"new_plan_id" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := h.advSubService.UpgradePlan(r.Context(), orgID, req.NewPlanID); err != nil {
		h.logger.Error("Failed to upgrade plan with proration", zap.Error(err), zap.String("tenant_id", orgID.String()))
		h.errorJSON(w, r, "Failed to upgrade plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Plan upgraded successfully with proration"}, http.StatusOK)
}

func (h *BillingHandler) RedeemCoupon(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Code           string     `json:"code" validate:"required"`
		SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
		InvoiceID      *uuid.UUID `json:"invoice_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := h.advSubService.RedeemCoupon(r.Context(), orgID, req.Code, req.SubscriptionID, req.InvoiceID); err != nil {
		h.logger.Error("Failed to redeem coupon", zap.Error(err), zap.String("tenant_id", orgID.String()), zap.String("code", req.Code))
		h.errorJSON(w, r, "Failed to redeem coupon: "+err.Error(), http.StatusBadRequest)
		return
	}

	h.json(w, map[string]string{"message": "Coupon redeemed successfully"}, http.StatusOK)
}
