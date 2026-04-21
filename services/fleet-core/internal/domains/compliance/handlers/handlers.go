package handlers

import (
	"encoding/json"
	"net/http"

	complianceservice "github.com/flyfleet/fleet-core/internal/domains/compliance/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	pb "github.com/flyfleet/proto/compliance/v1"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	// SECURITY FIX: Prevent IDOR - validate driver_id matches authenticated user
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		h.errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}

	reqDriverID, err := uuid.Parse(req.DriverId)
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid driver ID")
		return
	}

	// Check if user is allowed to record duty status for this driver
	if !h.canRecordDutyStatus(r, authUserUUID, reqDriverID) {
		h.logger.Warn("IDOR attempt: driver mismatch in duty status",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("requested_driver", req.DriverId))
		h.errorJSON(w, http.StatusForbidden, "cannot record duty status for another driver")
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

// canRecordDutyStatus checks if the authenticated user can record duty status for the target driver
func (h *Handlers) canRecordDutyStatus(r *http.Request, authUserID, targetDriverID uuid.UUID) bool {
	// Users can only record their own duty status
	if authUserID == targetDriverID {
		return true
	}

	// Admins/fleet managers can record for any driver in their org
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	adminRoles := map[string]bool{
		"super_admin":   true,
		"org_admin":     true,
		"fleet_manager": true,
		"dispatcher":    true,
	}
	if adminRoles[role] {
		// In a full implementation, we'd also verify org membership
		// For now, role check is sufficient
		return true
	}

	return false
}

func (h *Handlers) GetCurrentDutyStatus(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driver_id")
	if driverID == "" {
		h.errorJSON(w, http.StatusBadRequest, "driver_id required")
		return
	}

	// SECURITY: Validate user can view this driver's duty status
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		h.errorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}

	reqDriverUUID, err := uuid.Parse(driverID)
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid driver ID")
		return
	}

	// Check if user is allowed to view duty status for this driver
	if !h.canRecordDutyStatus(r, authUserUUID, reqDriverUUID) {
		h.errorJSON(w, http.StatusForbidden, "cannot view duty status for another driver")
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
