package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/service"
	coremiddleware "github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
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
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	config, err := h.svc.GetConfig(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get billing config", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, config, http.StatusOK)
}

// UpdateConfig updates billing configuration
func (h *BillingConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)

	var req models.BillingConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Validate
	if err := h.svc.ValidateConfig(&req); err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := h.svc.UpdateConfig(r.Context(), orgID, userID, &req)
	if err != nil {
		h.logger.Error("Failed to update billing config", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, config, http.StatusOK)
}

// DeleteConfig deletes billing configuration
func (h *BillingConfigHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteConfig(r.Context(), orgID); err != nil {
		h.logger.Error("Failed to delete billing config", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
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

func (h *BillingConfigHandler) errorJSON(w http.ResponseWriter, r *http.Request, message string, status int) {
	h.json(w, map[string]string{"error": message, "trace_id": coremiddleware.GetTraceID(r.Context())}, status)
}

func (h *BillingConfigHandler) getOrgID(r *http.Request) uuid.UUID {
	ctx := r.Context()

	// Exhaustive list of all known organization context keys
	keys := []interface{}{
		coremiddleware.CtxOrgID, // internal/middleware
		ctxkeys.OrgIDKey,        // internal/pkg/ctxkeys
		ctxkeys.TenantIDKey,     // internal/pkg/ctxkeys
		middleware.OrgIDKey,     // shared/pkg/middleware
		"tenant_id",
		"tenant_id",
	}

	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			// Try uuid.UUID first
			if oid, ok := val.(uuid.UUID); ok && oid != uuid.Nil {
				return oid
			}
			// Try string fallback
			if s, ok := val.(string); ok && s != "" {
				if oid, err := uuid.Parse(s); err == nil && oid != uuid.Nil {
					return oid
				}
			}
		}
	}

	return uuid.Nil
}

func (h *BillingConfigHandler) getUserID(r *http.Request) uuid.UUID {
	ctx := r.Context()

	keys := []interface{}{
		coremiddleware.CtxUserID,
		ctxkeys.UserIDKey,
		middleware.UserIDKey,
		"user_id",
		"sub",
	}

	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if uid, ok := val.(uuid.UUID); ok && uid != uuid.Nil {
				return uid
			}
			if s, ok := val.(string); ok && s != "" {
				if uid, err := uuid.Parse(s); err == nil && uid != uuid.Nil {
					return uid
				}
			}
		}
	}

	return uuid.Nil
}
