package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/device-diagnostics-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	service service.DeviceDiagnosticsService
}

func NewHandlers(s service.DeviceDiagnosticsService) *Handlers {
	return &Handlers{service: s}
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
