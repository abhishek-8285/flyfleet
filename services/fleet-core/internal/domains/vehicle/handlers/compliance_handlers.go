package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ComplianceHandlers struct {
	compSvc  *service.ComplianceService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewComplianceHandlers(compSvc *service.ComplianceService, logger *zap.Logger) *ComplianceHandlers {
	return &ComplianceHandlers{
		compSvc:  compSvc,
		logger:   logger,
		validate: validator.New(),
	}
}

// RegisterRoutes mounts compliance routes inside the vehicles subrouter.
// Expected to be called inside: r.Route("/vehicles", ...)
func (h *ComplianceHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/expiring", h.GetExpiringVehicles)
	r.Post("/{id}/verify-compliance", h.VerifyCompliance)
	r.Get("/{id}/compliance", h.GetCompliance)
}

// VerifyCompliance godoc — POST /vehicles/{id}/verify-compliance
func (h *ComplianceHandlers) VerifyCompliance(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	orgID := h.getTenantID(r)
	if orgID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Organization ID required")
		return
	}

	var req models.VerifyComplianceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	comp, status, err := h.compSvc.VerifyCompliance(r.Context(), vehicleID, orgID, &req)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		h.logger.Error("verify compliance failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to verify compliance")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"compliance": comp,
		"status":     status,
	})
}

// GetCompliance godoc — GET /vehicles/{id}/compliance
func (h *ComplianceHandlers) GetCompliance(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	orgID := h.getTenantID(r)
	if orgID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Organization ID required")
		return
	}

	comp, status, err := h.compSvc.GetCompliance(r.Context(), vehicleID, orgID)
	if err != nil {
		h.logger.Error("get compliance failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to get compliance")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"compliance": comp,
		"status":     status,
	})
}

// GetExpiringVehicles godoc — GET /vehicles/expiring?days=30&tenant_id=<uuid>
func (h *ComplianceHandlers) GetExpiringVehicles(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Tenant ID not found in context")
		return
	}

	alerts, err := h.compSvc.GetExpiringVehicles(r.Context(), tenantID, days)
	if err != nil {
		h.logger.Error("get expiring vehicles failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to get expiring vehicles")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
		"days":   days,
	})
}

// --- helpers ---

func (h *ComplianceHandlers) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *ComplianceHandlers) writeError(w http.ResponseWriter, code int, msg string) {
	h.writeJSON(w, code, map[string]string{"error": msg})
}

func (h *ComplianceHandlers) getTenantID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if s, ok := val.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
