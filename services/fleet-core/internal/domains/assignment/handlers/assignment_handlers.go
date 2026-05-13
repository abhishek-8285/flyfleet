package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/assignment/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type assignmentRequestDTO struct {
	VehicleID      uuid.UUID `json:"vehicle_id"      validate:"required"`
	VehicleVersion int       `json:"vehicle_version"`
	DriverID       uuid.UUID `json:"driver_id"       validate:"required"`
	DriverVersion  int       `json:"driver_version"`
	Remarks        string    `json:"remarks"`
	BypassPolicy   bool      `json:"bypass_policy"`
	BypassReason   string    `json:"bypass_reason"`
}

type unassignRequestDTO struct {
	VehicleID uuid.UUID `json:"vehicle_id" validate:"required"`
	Remarks   string    `json:"remarks"`
}

type AssignmentHandlers struct {
	service  *service.AssignmentService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewAssignmentHandlers(service *service.AssignmentService, logger *zap.Logger) *AssignmentHandlers {
	return &AssignmentHandlers{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *AssignmentHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/assignments", func(r chi.Router) {
		r.Post("/assign", h.Assign)
		r.Post("/unassign", h.Unassign)
		r.Post("/reconcile", h.Reconcile)
	})
}

func (h *AssignmentHandlers) Assign(w http.ResponseWriter, r *http.Request) {
	var req assignmentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("assignment request validation failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract Context (Tenant, Actor, Idempotency)
	tenantID, _ := ctxkeys.GetTenantID(r.Context())
	actorID, _ := ctxkeys.GetUserID(r.Context())
	requestID, _ := ctxkeys.GetRequestID(r.Context())
	
	opCtx := service.AssignmentContext{
		TenantID:       tenantID,
		ActorID:        actorID,
		ActorType:      "ADMIN",
		SourceIP:       r.RemoteAddr,
		RequestID:      requestID,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
		VehicleVersion: req.VehicleVersion,
		DriverVersion:  req.DriverVersion,
		BypassPolicy:   req.BypassPolicy,
		BypassReason:   req.BypassReason,
	}

	err := h.service.Assign(r.Context(), opCtx, req.VehicleID, req.DriverID, req.Remarks)
	if err != nil {
		var conflictErr *service.ConflictError
		if errors.As(err, &conflictErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(conflictErr)
			return
		}
		
		h.logger.Error("assignment failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "assignment_orchestrated"})
}

func (h *AssignmentHandlers) Unassign(w http.ResponseWriter, r *http.Request) {
	var req unassignRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenantID, _ := ctxkeys.GetTenantID(r.Context())
	actorID, _ := ctxkeys.GetUserID(r.Context())
	requestID, _ := ctxkeys.GetRequestID(r.Context())

	opCtx := service.AssignmentContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		ActorType: "ADMIN",
		SourceIP:  r.RemoteAddr,
		RequestID: requestID,
	}

	err := h.service.Unassign(r.Context(), opCtx, req.VehicleID, req.Remarks)
	if err != nil {
		h.logger.Error("unassignment failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AssignmentHandlers) Reconcile(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(ctxkeys.TenantIDKey).(uuid.UUID)
	
	summary, err := h.service.Reconcile(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("reconciliation failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(summary)
}
