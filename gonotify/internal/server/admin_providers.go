package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/tenant"
)

type AdminProviderHandler struct {
	repo   *provider.Repository
	logger *slog.Logger
}

func NewAdminProviderHandler(
	repo *provider.Repository,
	logger *slog.Logger,
) *AdminProviderHandler {
	return &AdminProviderHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *AdminProviderHandler) ListSystemProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.repo.ListSystemProviders(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list providers", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list system providers")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"providers": providers,
	})
}

func (h *AdminProviderHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	var req provider.CreateProviderConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.ProviderID == "" {
		WriteError(w, http.StatusBadRequest, "validation_failed", "provider_id is required")
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_failed", "name is required")
		return
	}

	configView, err := h.repo.CreateProviderConfig(r.Context(), tenantID, req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create provider config", "error", err)
		WriteError(w, http.StatusInternalServerError, "create_failed", "failed to create provider config")
		return
	}

	WriteJSON(w, http.StatusCreated, configView)
}

func (h *AdminProviderHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	configs, err := h.repo.ListProviderConfigs(r.Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list provider configs", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list provider configs")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"provider_configs": configs,
	})
}

func (h *AdminProviderHandler) EnableConfig(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, true)
}

func (h *AdminProviderHandler) DisableConfig(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, false)
}

func (h *AdminProviderHandler) setStatus(w http.ResponseWriter, r *http.Request, enabled bool) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in request context")
		return
	}

	configID := chi.URLParam(r, "configID")
	pc, err := h.repo.SetProviderConfigStatus(r.Context(), tenantID, configID, enabled)
	if errors.Is(err, provider.ErrProviderConfigNotFound) {
		WriteError(w, http.StatusNotFound, "provider_config_not_found", "provider config not found")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to set provider config status", "error", err)
		WriteError(w, http.StatusInternalServerError, "update_failed", "failed to update provider config status")
		return
	}

	WriteJSON(w, http.StatusOK, pc)
}
