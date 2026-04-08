package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LifecycleHandlers handles vehicle lifecycle state transitions.
type LifecycleHandlers struct {
	lifecycleSvc *service.LifecycleService
	logger       *zap.Logger
}

func NewLifecycleHandlers(svc *service.LifecycleService, logger *zap.Logger) *LifecycleHandlers {
	return &LifecycleHandlers{lifecycleSvc: svc, logger: logger}
}

// RegisterRoutes mounts lifecycle routes inside the vehicles/{id} subrouter.
func (h *LifecycleHandlers) RegisterRoutes(r chi.Router) {
	r.Post("/{id}/activate", h.Activate)
	r.Post("/{id}/suspend", h.Suspend)
	r.Post("/{id}/set-maintenance", h.SetMaintenance)
	r.Post("/{id}/decommission", h.Decommission)
}

// Activate godoc — POST /vehicles/{id}/activate
func (h *LifecycleHandlers) Activate(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	vehicle, err := h.lifecycleSvc.Activate(r.Context(), vehicleID)
	if err != nil {
		switch err {
		case service.ErrVehicleNotFound:
			h.writeError(w, http.StatusNotFound, "Vehicle not found")
		case service.ErrDecommissionedIrreversible:
			h.writeError(w, http.StatusConflict, err.Error())
		default:
			h.logger.Error("activate vehicle failed", zap.Error(err))
			h.writeError(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"data": vehicle, "status": "ACTIVE"})
}

// Suspend godoc — POST /vehicles/{id}/suspend
func (h *LifecycleHandlers) Suspend(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	vehicle, err := h.lifecycleSvc.Suspend(r.Context(), vehicleID, body.Reason)
	if err != nil {
		switch err {
		case service.ErrVehicleNotFound:
			h.writeError(w, http.StatusNotFound, "Vehicle not found")
		case service.ErrDecommissionedIrreversible:
			h.writeError(w, http.StatusConflict, err.Error())
		default:
			h.logger.Error("suspend vehicle failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to suspend vehicle")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"data": vehicle, "status": "SUSPENDED"})
}

// SetMaintenance godoc — POST /vehicles/{id}/set-maintenance
func (h *LifecycleHandlers) SetMaintenance(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	vehicle, err := h.lifecycleSvc.SetMaintenance(r.Context(), vehicleID)
	if err != nil {
		switch err {
		case service.ErrVehicleNotFound:
			h.writeError(w, http.StatusNotFound, "Vehicle not found")
		case service.ErrDecommissionedIrreversible:
			h.writeError(w, http.StatusConflict, err.Error())
		default:
			h.logger.Error("set-maintenance failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to set maintenance status")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"data": vehicle, "status": "MAINTENANCE"})
}

// Decommission godoc — POST /vehicles/{id}/decommission
func (h *LifecycleHandlers) Decommission(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var body struct {
		Reason string `json:"reason" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	vehicle, err := h.lifecycleSvc.Decommission(r.Context(), vehicleID, body.Reason)
	if err != nil {
		if err.Error() == "vehicle not found" {
			h.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		h.logger.Error("decommission vehicle failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to decommission vehicle")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"data": vehicle, "status": "DECOMMISSIONED"})
}

func (h *LifecycleHandlers) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (h *LifecycleHandlers) writeError(w http.ResponseWriter, code int, msg string) {
	h.writeJSON(w, code, map[string]string{"error": msg})
}
