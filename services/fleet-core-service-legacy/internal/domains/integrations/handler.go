package integrations

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateIntegration)
	r.Get("/", h.ListIntegrations)
	r.Get("/{id}", h.GetIntegration)
	r.Delete("/{id}", h.DeleteIntegration)
}

func (h *Handler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	var req CreateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	integration, err := h.service.CreateIntegration(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create integration", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create integration")
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, integration)
}

func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	integration, err := h.service.GetIntegration(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Integration not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, integration)
}

func (h *Handler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, _ := strconv.Atoi(limitStr)

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, _ := strconv.Atoi(offsetStr)

	integrations, err := h.service.ListIntegrations(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list integrations", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list integrations")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": integrations})
}

func (h *Handler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if err := h.service.DeleteIntegration(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete integration", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete integration")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}
