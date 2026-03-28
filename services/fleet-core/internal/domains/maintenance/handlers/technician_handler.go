package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/service"
)

type TechnicianHandler struct {
	service *service.WorkOrderService // or TechnicianService if we split it
	logger  *zap.Logger
}

func NewTechnicianHandler(svc *service.WorkOrderService, logger *zap.Logger) *TechnicianHandler {
	return &TechnicianHandler{service: svc, logger: logger}
}

func (h *TechnicianHandler) RegisterRoutes(r chi.Router) {
	r.Route("/technicians", func(r chi.Router) {
		r.Get("/", h.ListTechnicians)
		// Add more endpoints as needed
	})
}

func (h *TechnicianHandler) ListTechnicians(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	w.WriteHeader(http.StatusNotImplemented)
}
