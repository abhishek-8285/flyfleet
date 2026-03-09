package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BillingConfigHandler handles billing configuration HTTP requests
type BillingConfigHandler struct {
	svc    *service.BillingConfigService
	logger *zap.Logger
}

// NewBillingConfigHandler creates a new handler
func NewBillingConfigHandler(svc *service.BillingConfigService, logger *zap.Logger) *BillingConfigHandler {
	return &BillingConfigHandler{
		svc:    svc,
		logger: logger,
	}
}

// RegisterRoutes registers billing config routes
func (h *BillingConfigHandler) RegisterRoutes(r chi.Router) {
	r.Route("/billing-config", func(r chi.Router) {
		r.Get("/", h.GetConfig)
		r.Put("/", h.UpdateConfig)
		r.Delete("/", h.DeleteConfig)
	})
}

// GetConfig retrieves billing configuration
func (h *BillingConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	config, err := h.svc.GetConfig(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get billing config", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, config, http.StatusOK)
}

// UpdateConfig updates billing configuration
func (h *BillingConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	var req models.BillingConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Validate
	if err := h.svc.ValidateConfig(&req); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := h.svc.UpdateConfig(r.Context(), orgID, userID, &req)
	if err != nil {
		h.logger.Error("Failed to update billing config", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, config, http.StatusOK)
}

// DeleteConfig deletes billing configuration
func (h *BillingConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteConfig(r.Context(), orgID); err != nil {
		h.logger.Error("Failed to delete billing config", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "deleted"}, http.StatusOK)
}

// Helper methods
func (h *BillingConfigHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *BillingConfigHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *BillingConfigHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *BillingConfigHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
