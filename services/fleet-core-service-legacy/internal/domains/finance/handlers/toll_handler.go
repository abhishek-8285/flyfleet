package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/finance/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type TollHandler struct {
	svc    *service.TollService
	logger *zap.Logger
}

func NewTollHandler(svc *service.TollService, logger *zap.Logger) *TollHandler {
	return &TollHandler{svc: svc, logger: logger}
}

func (h *TollHandler) RegisterRoutes(r chi.Router) {
	r.Post("/ingest", h.IngestToll)
}

func (h *TollHandler) IngestToll(w http.ResponseWriter, r *http.Request) {
	var req models.IngestTollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.IngestToll(r.Context(), req); err != nil {
		h.logger.Error("Failed to ingest toll", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Toll ingested"})
}
