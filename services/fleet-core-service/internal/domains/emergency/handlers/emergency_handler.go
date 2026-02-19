package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmergencyHandlers struct {
	service   *service.EmergencyService
	logger    *zap.Logger
	validator *validator.Validate
}

func NewEmergencyHandlers(s *service.EmergencyService, logger *zap.Logger) *EmergencyHandlers {
	return &EmergencyHandlers{
		service:   s,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *EmergencyHandlers) RegisterRoutes(r chi.Router) {
	r.Post("/sos-trigger", h.TriggerSOS)
	r.Get("/incidents", h.ListActiveIncidents)
	r.Post("/incidents/{id}/acknowledge", h.AcknowledgeIncident)
	r.Post("/incidents/{id}/resolve", h.ResolveIncident)
}

func (h *EmergencyHandlers) TriggerSOS(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSOSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	incident, err := h.service.TriggerSOS(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to trigger SOS", zap.Error(err))
		http.Error(w, "Failed to trigger SOS", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusCreated, incident)
}

func (h *EmergencyHandlers) ListActiveIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.service.ListActiveIncidents(r.Context())
	if err != nil {
		h.logger.Error("Failed to list incidents", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": incidents})
}

func (h *EmergencyHandlers) AcknowledgeIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	userID := getActorUserID(r)

	var req models.UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AcknowledgeIncident(r.Context(), id, userID, req.Notes); err != nil {
		h.logger.Error("Failed to acknowledge incident", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Incident acknowledged"})
}

func (h *EmergencyHandlers) ResolveIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	userID := getActorUserID(r)

	var req models.UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	status := models.SOSStatus(req.Status)
	if err := h.service.ResolveIncident(r.Context(), id, userID, req.Notes, status); err != nil {
		h.logger.Error("Failed to resolve incident", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Incident resolved"})
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getActorUserID(r *http.Request) uuid.UUID {
	if fromCtx, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		return fromCtx
	}
	if fromHeader := r.Header.Get("X-User-ID"); fromHeader != "" {
		if parsed, err := uuid.Parse(fromHeader); err == nil {
			return parsed
		}
	}
	return uuid.Nil
}
