package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handlers struct {
	service service.DeviceDiagnosticsService
	logger  *zap.Logger
}

func NewHandlers(s service.DeviceDiagnosticsService, logger *zap.Logger) *Handlers {
	return &Handlers{service: s, logger: logger}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/summary", h.GetSummary)
	r.Get("/status", h.ListDevices)
	r.Get("/twins", h.ListDeviceTwins)
	r.Get("/twins/{id}", h.GetDeviceTwin)
	r.Put("/twins/{id}/desired", h.UpdateDesiredState)
	r.Put("/twins/{id}/reported", h.UpdateReportedState)
	r.Get("/devices", h.ListDevices)
	r.Get("/devices/status", h.ListDevices)
	r.Route("/devices/{id}", func(r chi.Router) {
		r.Get("/", h.GetDevice)
		r.Get("/sensors", h.GetDeviceSensors)
		r.Get("/logs", h.GetDeviceLogs)
		r.Get("/telemetry", h.GetDeviceTelemetry)
		r.Post("/commands", h.SendCommand)
		r.Get("/commands", h.GetCommandHistory)
		r.Patch("/commands/{commandID}/status", h.TransitionCommand)
	})
	r.Get("/firmware", h.ListFirmware)
	r.Post("/firmware", h.CreateFirmware)
	r.Get("/rollouts", h.ListRollouts)
	r.Get("/rollouts/{id}", h.GetRollout)
	r.Get("/rollouts/{id}/events", h.ListRolloutEvents)
	r.Post("/rollouts/{id}/events", h.AddRolloutNote)
	r.Post("/rollouts", h.CreateRollout)
	r.Patch("/rollouts/{id}/status", h.TransitionRollout)
}

// Dashboard handlers

func (h *Handlers) GetSummary(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	summary, err := h.service.GetSummary(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, summary)
}

func (h *Handlers) ListDevices(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}

	devices, err := h.service.ListDevices(r.Context(), orgID, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, devices)
}

// Device detail handlers

func (h *Handlers) GetDevice(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := h.service.GetDevice(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, device)
}

func (h *Handlers) GetDeviceSensors(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	sensors, err := h.service.GetDeviceSensors(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, sensors)
}

func (h *Handlers) GetDeviceLogs(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := h.service.GetDeviceLogs(r.Context(), id, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, logs)
}

// Telemetry handlers

func (h *Handlers) GetDeviceTelemetry(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	telemetry, err := h.service.GetDeviceTelemetry(r.Context(), id, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, telemetry)
}

// Command handlers

func (h *Handlers) SendCommand(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	userID := getUserID(r)
	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Command  string                 `json:"command"`
		Category string                 `json:"category"`
		Params   map[string]interface{} `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Command = strings.TrimSpace(req.Command)
	req.Category = strings.TrimSpace(req.Category)
	if req.Command == "" || req.Category == "" {
		http.Error(w, "command and category are required", http.StatusBadRequest)
		return
	}

	cmd, err := h.service.SendCommand(r.Context(), deviceID, orgID, userID, req.Command, req.Category, req.Params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, cmd)
}

func (h *Handlers) GetCommandHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	commands, err := h.service.GetCommandHistory(r.Context(), deviceID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, commands)
}

func (h *Handlers) TransitionCommand(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)

	deviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}
	commandID, err := uuid.Parse(chi.URLParam(r, "commandID"))
	if err != nil {
		http.Error(w, "Invalid command ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status       string                 `json:"status"`
		Result       map[string]interface{} `json:"result"`
		ErrorMessage *string                `json:"error_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	updated, err := h.service.TransitionCommand(r.Context(), deviceID, orgID, commandID, req.Status, req.Result, req.ErrorMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

func (h *Handlers) ListDeviceTwins(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	twins, err := h.service.ListDeviceTwins(r.Context(), orgID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, twins)
}

func (h *Handlers) GetDeviceTwin(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	twin, err := h.service.GetDeviceTwin(r.Context(), deviceID, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, twin)
}

func (h *Handlers) UpdateDesiredState(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req struct {
		DesiredState map[string]interface{} `json:"desired_state"`
		CommandID    *string                `json:"command_id"`
		CommandState *string                `json:"command_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	var commandID *uuid.UUID
	if req.CommandID != nil && strings.TrimSpace(*req.CommandID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*req.CommandID))
		if err != nil {
			http.Error(w, "Invalid command_id", http.StatusBadRequest)
			return
		}
		commandID = &parsed
	}

	if err := h.service.UpdateDesiredState(r.Context(), deviceID, orgID, req.DesiredState, commandID, req.CommandState); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handlers) UpdateReportedState(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ReportedState map[string]interface{} `json:"reported_state"`
		LastSeen      *time.Time             `json:"last_seen"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateReportedState(r.Context(), deviceID, orgID, req.ReportedState, req.LastSeen); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Firmware handlers

func (h *Handlers) ListFirmware(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	firmware, err := h.service.ListFirmware(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, firmware)
}

func (h *Handlers) CreateFirmware(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Version        string  `json:"version"`
		HardwareTarget string  `json:"hardware_target"`
		FileURL        string  `json:"file_url"`
		FileSize       int     `json:"file_size"`
		Checksum       *string `json:"checksum"`
		ReleaseNotes   *string `json:"release_notes"`
		IsCritical     bool    `json:"is_critical"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Version = strings.TrimSpace(req.Version)
	req.HardwareTarget = strings.TrimSpace(req.HardwareTarget)
	req.FileURL = strings.TrimSpace(req.FileURL)
	if req.Version == "" || req.HardwareTarget == "" || req.FileURL == "" || req.FileSize <= 0 {
		http.Error(w, "version, hardware_target, file_url, and file_size are required", http.StatusBadRequest)
		return
	}

	fw, err := h.service.CreateFirmware(r.Context(), req.Version, req.HardwareTarget, req.FileURL, req.FileSize, req.Checksum, req.ReleaseNotes, req.IsCritical)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, fw)
}

// Rollout handlers

func (h *Handlers) ListRollouts(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rollouts, err := h.service.ListRollouts(r.Context(), orgID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, rollouts)
}

func (h *Handlers) GetRollout(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	rolloutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid rollout ID", http.StatusBadRequest)
		return
	}

	rollout, err := h.service.GetRollout(r.Context(), rolloutID, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, http.StatusOK, rollout)
}

func (h *Handlers) ListRolloutEvents(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	rolloutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid rollout ID", http.StatusBadRequest)
		return
	}

	limit := 50
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsedLimit, parseErr := strconv.Atoi(rawLimit)
		if parseErr != nil || parsedLimit <= 0 || parsedLimit > 500 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if rawOffset := strings.TrimSpace(r.URL.Query().Get("offset")); rawOffset != "" {
		parsedOffset, parseErr := strconv.Atoi(rawOffset)
		if parseErr != nil || parsedOffset < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsedOffset
	}
	eventType := strings.TrimSpace(r.URL.Query().Get("event_type"))

	events, err := h.service.ListRolloutEvents(r.Context(), rolloutID, orgID, eventType, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusOK, events)
}

func (h *Handlers) AddRolloutNote(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	userID := getUserID(r)
	rolloutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid rollout ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Message  string                 `json:"message"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	var actorUserID *uuid.UUID
	if userID != uuid.Nil {
		actorUserID = &userID
	}
	event, err := h.service.AddRolloutNote(r.Context(), rolloutID, orgID, actorUserID, req.Message, req.Metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusCreated, event)
}

func (h *Handlers) CreateRollout(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	userID := getUserID(r)

	var req struct {
		FirmwareID   string `json:"firmware_id"`
		Name         string `json:"name"`
		DeviceGroup  string `json:"device_group"`
		TotalDevices int    `json:"total_devices"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.DeviceGroup = strings.TrimSpace(req.DeviceGroup)
	if req.Name == "" || req.DeviceGroup == "" || req.TotalDevices <= 0 {
		http.Error(w, "name, device_group, and total_devices are required", http.StatusBadRequest)
		return
	}

	firmwareID, err := uuid.Parse(req.FirmwareID)
	if err != nil {
		http.Error(w, "Invalid firmware ID", http.StatusBadRequest)
		return
	}

	rollout, err := h.service.CreateRollout(r.Context(), firmwareID, orgID, userID, req.Name, req.DeviceGroup, req.TotalDevices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusCreated, rollout)
}

func (h *Handlers) TransitionRollout(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	userID := getUserID(r)
	rolloutID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid rollout ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string  `json:"status"`
		Reason *string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	var actorUserID *uuid.UUID
	if userID != uuid.Nil {
		actorUserID = &userID
	}
	updated, err := h.service.TransitionRollout(r.Context(), rolloutID, orgID, req.Status, actorUserID, req.Reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

// Helpers

func getOrgID(r *http.Request) uuid.UUID {
	if orgID, ok := r.Context().Value("org_id").(uuid.UUID); ok {
		return orgID
	}
	idStr := r.Header.Get("X-Organization-ID")
	id, _ := uuid.Parse(idStr)
	return id
}

func getUserID(r *http.Request) uuid.UUID {
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		return userID
	}
	idStr := r.Header.Get("X-User-ID")
	id, _ := uuid.Parse(idStr)
	return id
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
