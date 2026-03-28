package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/integrations/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	service service.IntegrationsService
}

func NewHandlers(s service.IntegrationsService) *Handlers {
	return &Handlers{service: s}
}

// getOrgID extracts and parses the organization ID from context
func (h *Handlers) getOrgID(r *http.Request) (uuid.UUID, error) {
	orgIDStr := middleware.GetOrgID(r.Context())
	if orgIDStr == "" {
		return uuid.Nil, fmt.Errorf("unauthorized: organization context required")
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid organization ID")
	}
	return orgID, nil
}

// Integrations Handlers

func (h *Handlers) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	integrations, err := h.service.ListIntegrations(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, integrations)
}

func (h *Handlers) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Provider string `json:"provider"`
		Config   any    `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	integration, err := h.service.CreateIntegration(r.Context(), orgID, req.Name, req.Type, req.Provider, req.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, integration)
}

func (h *Handlers) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Provider string `json:"provider"`
		Status   string `json:"status"`
		Config   any    `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = h.service.UpdateIntegration(r.Context(), id, orgID, req.Name, req.Type, req.Provider, req.Status, req.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handlers) TestConnection(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Extract ID from path... (simplified for now)
	idStr := r.URL.Query().Get("id")
	id, _ := uuid.Parse(idStr)

	success, msg, err := h.service.TestConnection(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"success": success, "message": msg})
}

// API Client Handlers

func (h *Handlers) ListAPIClients(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	clients, err := h.service.ListAPIClients(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, clients)
}

func (h *Handlers) CreateAPIClient(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req struct {
		Name       string   `json:"name"`
		Scopes     []string `json:"scopes"`
		ExpiryDays int      `json:"expiry_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	client, rawKey, err := h.service.CreateAPIClient(r.Context(), orgID, req.Name, req.Scopes, req.ExpiryDays)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]any{"client": client, "client_secret": rawKey})
}

// Webhook Handlers

func (h *Handlers) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	webhooks, err := h.service.ListWebhooks(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, webhooks)
}

func (h *Handlers) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	webhook, err := h.service.CreateWebhook(r.Context(), orgID, req.URL, req.Events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, webhook)
}

func (h *Handlers) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteWebhook(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Log Handlers

func (h *Handlers) ListLogs(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var integrationID *uuid.UUID
	if iID := r.URL.Query().Get("integration_id"); iID != "" {
		parsed, _ := uuid.Parse(iID)
		integrationID = &parsed
	}

	logs, err := h.service.ListLogs(r.Context(), orgID, integrationID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, logs)
}

// Helpers

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
