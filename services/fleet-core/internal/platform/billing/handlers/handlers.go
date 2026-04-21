package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/platform/billing/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BillingHandlers struct {
	svc service.BillingService
}

func NewBillingHandlers(svc service.BillingService) *BillingHandlers {
	return &BillingHandlers{svc: svc}
}

func (h *BillingHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/billing", func(r chi.Router) {
		r.Get("/plans", h.GetPlans)
		r.Get("/subscription", h.GetSubscriptionStatus)
	})
}

func (h *BillingHandlers) GetPlans(w http.ResponseWriter, r *http.Request) {
	// For simplicity, we can define a static list or fetch from DB if needed
	// Here we just acknowledge the endpoint structure
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plans": []string{"Starter", "Professional", "Enterprise"},
	})
}

func (h *BillingHandlers) GetSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	var tenantID uuid.UUID
	var ok bool

	if val := r.Context().Value(ctxkeys.TenantIDKey); val != nil {
		if id, isUUID := val.(uuid.UUID); isUUID {
			tenantID = id
			ok = true
		} else if s, isStr := val.(string); isStr {
			if id, err := uuid.Parse(s); err == nil {
				tenantID = id
				ok = true
			}
		}
	}

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	plan, err := h.svc.GetOrganizationPlan(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}
