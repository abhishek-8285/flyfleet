package handlers

import (
	"encoding/json"
	"net/http"

	complianceservice "github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/service"
	pb "github.com/flyfleet/proto/compliance/v1"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handlers struct {
	svc    complianceservice.ComplianceService
	logger *zap.Logger
}

func NewHandlers(svc complianceservice.ComplianceService, logger *zap.Logger) *Handlers {
	return &Handlers{svc: svc, logger: logger}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/compliance", func(r chi.Router) {
		r.Post("/duty-status/record", h.RecordDutyStatus)
		r.Get("/duty-status/current/{driver_id}", h.GetCurrentDutyStatus)
	})
}

func (h *Handlers) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handlers) errorJSON(w http.ResponseWriter, status int, msg string) {
	h.json(w, status, map[string]string{"error": msg})
}

func (h *Handlers) RecordDutyStatus(w http.ResponseWriter, r *http.Request) {
	var req pb.RecordDutyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.svc.RecordDutyStatus(r.Context(), &req)
	if err != nil {
		h.logger.Error("record duty status failed", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to record duty status")
		return
	}
	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) GetCurrentDutyStatus(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driver_id")
	if driverID == "" {
		h.errorJSON(w, http.StatusBadRequest, "driver_id required")
		return
	}
	log, err := h.svc.GetCurrentDutyStatus(r.Context(), driverID)
	if err != nil {
		h.logger.Error("get current duty status failed", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to get duty status")
		return
	}
	h.json(w, http.StatusOK, &pb.GetCurrentDutyStatusResponse{
		Log:                     log,
		RemainingDrivingMinutes: 660,
		RemainingShiftMinutes:   840,
		RemainingCycleMinutes:   4200,
		TimeUntilBreakMinutes:   480,
	})
}
