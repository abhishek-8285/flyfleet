package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gonotify/gonotify/internal/tenant"
	"github.com/gonotify/gonotify/internal/template"
)

type AdminTemplateHandler struct {
	repo   *template.Repository
	logger *slog.Logger
}

func NewAdminTemplateHandler(repo *template.Repository, logger *slog.Logger) *AdminTemplateHandler {
	return &AdminTemplateHandler{
		repo:   repo,
		logger: logger,
	}
}

type createTemplateRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Channel     string `json:"channel"`
	Category    string `json:"category"`
}

type createVersionRequest struct {
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

func (h *AdminTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in context")
		return
	}

	var req createTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Key == "" || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_failed", "key and name are required")
		return
	}

	tpl, err := h.repo.CreateTemplate(
		r.Context(),
		tenantID,
		req.Key,
		req.Name,
		req.Description,
		req.Channel,
		req.Category,
	)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create template", "error", err)
		WriteError(w, http.StatusInternalServerError, "create_failed", "failed to create template")
		return
	}

	WriteJSON(w, http.StatusCreated, tpl)
}

func (h *AdminTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in context")
		return
	}

	templates, err := h.repo.ListTemplates(r.Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list templates", "error", err)
		WriteError(w, http.StatusInternalServerError, "list_failed", "failed to list templates")
		return
	}

	WriteJSON(w, http.StatusOK, templates)
}

func (h *AdminTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in context")
		return
	}

	templateID := chi.URLParam(r, "template_id")
	if templateID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "template_id is required")
		return
	}

	tpl, err := h.repo.GetTemplateByID(r.Context(), tenantID, templateID)
	if errors.Is(err, template.ErrTemplateNotFound) {
		WriteError(w, http.StatusNotFound, "template_not_found", "template not found")
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get template", "template_id", templateID, "error", err)
		WriteError(w, http.StatusInternalServerError, "get_failed", "failed to get template")
		return
	}

	WriteJSON(w, http.StatusOK, tpl)
}

func (h *AdminTemplateHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in context")
		return
	}

	templateID := chi.URLParam(r, "template_id")
	if templateID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "template_id is required")
		return
	}

	var req createVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	ver, err := h.repo.CreateTemplateVersion(
		r.Context(),
		tenantID,
		templateID,
		req.Subject,
		req.HTMLBody,
		req.TextBody,
	)

	if errors.Is(err, template.ErrTemplateNotFound) {
		WriteError(w, http.StatusNotFound, "template_not_found", "template not found")
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create template version", "template_id", templateID, "error", err)
		WriteError(w, http.StatusInternalServerError, "create_version_failed", "failed to create template version")
		return
	}

	WriteJSON(w, http.StatusCreated, ver)
}

func (h *AdminTemplateHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenant.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusInternalServerError, "missing_tenant", "tenant not found in context")
		return
	}

	templateID := chi.URLParam(r, "template_id")
	versionID := chi.URLParam(r, "version_id")

	if templateID == "" || versionID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "template_id and version_id are required")
		return
	}

	ver, err := h.repo.PublishTemplateVersion(
		r.Context(),
		tenantID,
		templateID,
		versionID,
	)

	if errors.Is(err, template.ErrTemplateNotFound) {
		WriteError(w, http.StatusNotFound, "template_not_found", "template not found")
		return
	}

	if errors.Is(err, template.ErrVersionNotFound) {
		WriteError(w, http.StatusNotFound, "version_not_found", "template version not found")
		return
	}

	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to publish template version", "template_id", templateID, "version_id", versionID, "error", err)
		WriteError(w, http.StatusInternalServerError, "publish_version_failed", "failed to publish template version")
		return
	}

	WriteJSON(w, http.StatusOK, ver)
}
