package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// DeviceCommandHandler handles POST /v1/devices/{id}/command.
type DeviceCommandHandler struct {
	svc       *service.DeviceCommandService
	validator *validator.Validate
	logger    *zap.Logger
}

// NewDeviceCommandHandler creates a new command handler.
func NewDeviceCommandHandler(svc *service.DeviceCommandService, logger *zap.Logger) *DeviceCommandHandler {
	return &DeviceCommandHandler{
		svc:       svc,
		validator: validator.New(),
		logger:    logger,
	}
}

// RegisterCommandRoutes mounts the command endpoints.
func (h *DeviceCommandHandler) RegisterCommandRoutes(r chi.Router) {
	r.Post("/v1/devices/{id}/command", h.SendCommand)
	r.Get("/v1/devices/{id}/commands", h.GetPendingCommands)
}

// SendCommand enqueues a command for a device.
//
//	POST /v1/devices/{id}/command
//	Body: { "type": "ping|reboot|immobilize|config_update", "params": {} }
func (h *DeviceCommandHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	if deviceID == "" {
		jsonError(w, "device_id is required", http.StatusBadRequest)
		return
	}

	var req models.SendCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		jsonError(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	commandID, err := h.svc.SendCommand(r.Context(), deviceID, req.Type, req.Params)
	if err != nil {
		h.logger.Error("send command failed",
			zap.String("device_id", deviceID),
			zap.String("type", req.Type),
			zap.Error(err),
		)
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]any{
		"command_id": commandID,
		"device_id":  deviceID,
		"type":       req.Type,
		"status":     "pending",
	}, http.StatusAccepted)
}

// GetPendingCommands returns all pending commands for a device.
//
//	GET /v1/devices/{id}/commands
func (h *DeviceCommandHandler) GetPendingCommands(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	cmds, err := h.svc.GetPendingCommands(r.Context(), deviceID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if cmds == nil {
		cmds = []models.DeviceCommand{}
	}
	jsonOK(w, map[string]any{
		"device_id": deviceID,
		"count":     len(cmds),
		"data":      cmds,
	}, http.StatusOK)
}
