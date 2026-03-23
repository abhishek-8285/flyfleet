package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.subService.ListPlans(r.Context())
	if err != nil {
		h.logger.Error("Failed to list plans", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.json(w, plans, http.StatusOK)
}

func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	sub, err := h.subService.GetSubscription(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get subscription", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if sub == nil {
		h.errorJSON(w, r, "No active subscription found", http.StatusNotFound)
		return
	}

	h.json(w, sub, http.StatusOK)
}

func (h *BillingHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PlanID uuid.UUID `json:"plan_id" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	sub, err := h.subService.Subscribe(r.Context(), orgID, req.PlanID)
	if err != nil {
		h.logger.Error("Failed to subscribe", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, sub, http.StatusCreated)
}

func (h *BillingHandler) ChangePlan(w http.ResponseWriter, r *http.Request) {
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

	sub, err := h.subService.ChangePlan(r.Context(), orgID, req.NewPlanID)
	if err != nil {
		h.logger.Error("Failed to change plan", zap.Error(err))
		h.errorJSON(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, sub, http.StatusOK)
}

func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Immediate bool `json:"immediate"`
	}
	// Decode is optional as we have defaults
	_ = json.NewDecoder(r.Body).Decode(&req)

	err := h.subService.CancelSubscription(r.Context(), orgID, req.Immediate)
	if err != nil {
		h.logger.Error("Failed to cancel subscription", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Subscription cancellation processed"}, http.StatusOK)
}
