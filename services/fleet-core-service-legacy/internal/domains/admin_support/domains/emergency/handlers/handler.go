package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/emergency/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Global rate limiter: 3 halt commands per 5 minutes per issuer
var haltRateLimiter = NewRateLimiter(3, 5*time.Minute)

type EmergencyHandler struct {
	svc    service.EmergencyService
	logger *zap.Logger
}

func NewEmergencyHandler(svc service.EmergencyService, logger *zap.Logger) *EmergencyHandler {
	return &EmergencyHandler{svc: svc, logger: logger}
}

func (h *EmergencyHandler) RegisterRoutes(r chi.Router) {
	r.Post("/halt", h.GlobalHalt)
	r.Post("/clear", h.ClearHalt)
	r.Get("/status", h.GetStatus)
}

func (h *EmergencyHandler) GlobalHalt(w http.ResponseWriter, r *http.Request) {
	var req models.EmergencyHaltRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		h.errorJSON(w, "reason is required", http.StatusBadRequest)
		return
	}

	// In a real system, issuedBy would come from JWT context
	issuedBy := req.AuthorizedBy
	if issuedBy == "" {
		issuedBy = "Administrator"
	}

	// Rate limiting: Check if issuer has exceeded halt command limit
	if !haltRateLimiter.Allow(issuedBy) {
		h.logger.Warn("Rate limit exceeded for halt command", zap.String("issuer", issuedBy))
		h.errorJSON(w, "Rate limit exceeded: Maximum 3 halt commands per 5 minutes", http.StatusTooManyRequests)
		return
	}

	command, err := h.svc.GlobalHalt(r.Context(), req.Reason, issuedBy)
	if err != nil {
		h.logger.Error("Failed to issue global halt", zap.Error(err))
		h.errorJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.json(w, command, http.StatusAccepted)
}

func (h *EmergencyHandler) ClearHalt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AuthorizedBy string `json:"authorized_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clearedBy := req.AuthorizedBy
	if clearedBy == "" {
		clearedBy = "Administrator"
	}

	if err := h.svc.ClearHalt(r.Context(), clearedBy); err != nil {
		h.logger.Error("Failed to clear halt", zap.Error(err))
		h.errorJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "cleared"}, http.StatusOK)
}

func (h *EmergencyHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.GetStatus(r.Context())
	if err != nil {
		h.errorJSON(w, "Failed to get emergency status", http.StatusInternalServerError)
		return
	}
	h.json(w, status, http.StatusOK)
}

func (h *EmergencyHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EmergencyHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}
