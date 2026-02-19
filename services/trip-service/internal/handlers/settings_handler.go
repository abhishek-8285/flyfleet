package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/flyfleet/trip-service/internal/service"
	"go.uber.org/zap"
)

type SettingsHandler struct {
	service *service.SettingsService
	logger  *zap.Logger
}

func NewSettingsHandler(service *service.SettingsService, logger *zap.Logger) *SettingsHandler {
	return &SettingsHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SettingsHandler) GetTripSettings(w http.ResponseWriter, r *http.Request) {
	// org_id would come from auth middleware via context
	// In the absence of it, we use the fallback as in original code
	orgID := "" // value, ok := r.Context().Value("org_id").(string)
	if orgID == "" {
		orgID = "550e8400-e29b-41d4-a716-446655440002" // Fallback
	}

	resp, err := h.service.GetSettings(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, resp)
}

func (h *SettingsHandler) UpdateTripSettings(w http.ResponseWriter, r *http.Request) {
	orgID := ""
	if orgID == "" {
		orgID = "550e8400-e29b-41d4-a716-446655440002"
	}

	userID := ""
	if userID == "" {
		userID = "00000000-0000-0000-0000-000000000001"
	}

	var req models.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.UpdateSettings(r.Context(), orgID, req, userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return updated settings (same format as GET)
	resp, err := h.service.GetSettings(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, resp)
}

// Helper methods

func (h *SettingsHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *SettingsHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}
