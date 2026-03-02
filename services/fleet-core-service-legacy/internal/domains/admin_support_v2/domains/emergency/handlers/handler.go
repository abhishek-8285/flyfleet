package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/emergency/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type EmergencyHandler struct {
	svc    service.EmergencyService
	logger *zap.Logger
}

func NewEmergencyHandler(svc service.EmergencyService, logger *zap.Logger) *EmergencyHandler {
	return &EmergencyHandler{svc: svc, logger: logger}
}

func (h *EmergencyHandler) RegisterRoutes(r chi.Router) {
	r.Route("/emergency", func(r chi.Router) {
		r.Post("/halt", h.GlobalHalt)
		r.Get("/status", h.GetStatus)
	})
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

	command, err := h.svc.GlobalHalt(r.Context(), req.Reason, issuedBy)
	if err != nil {
		h.logger.Error("Failed to issue global halt", zap.Error(err))
		h.errorJSON(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.json(w, command, http.StatusAccepted)
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
