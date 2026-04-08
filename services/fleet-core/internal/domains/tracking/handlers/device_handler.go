package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
)

// DeviceServiceInterface defines the interface for device service
type DeviceServiceInterface interface {
	RegisterDevice(ctx context.Context, req *models.DeviceRegisterRequest) (*models.DeviceRegisterResponse, error)
	UpdateHeartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error)
	GetConfig(ctx context.Context, deviceID string) (*models.ConfigResponse, error)
	StoreTelemetryBatch(ctx context.Context, req *models.TelemetryBatchRequest) (*models.TelemetryResponse, error)
	GetMediaUploadURL(ctx context.Context, req *models.UploadUrlRequest) (*models.UploadUrlResponse, error)
	ConfirmMediaUpload(ctx context.Context, req *models.ConfirmUploadRequest) (*models.ConfirmUploadResponse, error)
	AttestEvidence(ctx context.Context, req *models.AttestationRequest) (*models.AttestationResponse, error)
	GetPendingCommands(ctx context.Context, deviceID, lastCommandID string) ([]models.Command, error)
	DriverLogin(ctx context.Context, req *models.DriverLoginRequest) (*models.DriverLoginResponse, error)
	DriverLogout(ctx context.Context, req *models.DriverLogoutRequest) (*models.DriverLogoutResponse, error)
	UpdateDriverHOS(ctx context.Context, req *models.HOSUpdateRequest) (*models.HOSUpdateResponse, error)
}

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct {
	deviceService DeviceServiceInterface
	logger        *zap.Logger
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceService DeviceServiceInterface, logger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		logger:        logger,
	}
}

// RegisterRoutes registers device routes
func (h *DeviceHandler) RegisterRoutes(router chi.Router) {
	router.Post("/device/register", h.Register)
	router.Post("/device/heartbeat", h.Heartbeat)
	router.Get("/device/config", h.GetConfig)
	router.Post("/telemetry/batch", h.UploadTelemetry)
	router.Post("/media/upload-url", h.GetMediaUploadURL)
	router.Post("/media/confirm", h.ConfirmMediaUpload)
	router.Post("/forensics/attest", h.AttestEvidence)
	router.Get("/device/commands", h.GetPendingCommands)
	router.Post("/driver/login", h.DriverLogin)
	router.Post("/driver/logout", h.DriverLogout)
	router.Post("/driver/hos", h.UpdateDriverHOS)
}

// Register handles device registration
// @Summary Register a new device
// @Tags device
// @Accept json
// @Produce json
// @Param request body models.DeviceRegisterRequest true "Registration request"
// @Success 200 {object} models.DeviceRegisterResponse
// @Router /device/register [post]
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.DeviceRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate request
	if req.DeviceID == "" {
		http.Error(w, `{"error": "device_id is required"}`, http.StatusBadRequest)
		return
	}

	if req.HardwareAttestation == "" {
		http.Error(w, `{"error": "hardware_attestation is required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing device registration", zap.String("device_id", req.DeviceID))

	resp, err := h.deviceService.RegisterDevice(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to register device", zap.Error(err))
		http.Error(w, `{"error": "registration failed"}`, http.StatusInternalServerError)
		return
	}

	h.logger.Info("Device registered successfully", zap.String("device_id", req.DeviceID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Heartbeat handles device heartbeat
// @Summary Update device heartbeat
// @Tags device
// @Accept json
// @Produce json
// @Param request body models.HeartbeatRequest true "Heartbeat request"
// @Success 200 {object} models.HeartbeatResponse
// @Router /device/heartbeat [post]
func (h *DeviceHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	var req models.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode heartbeat request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, `{"error": "device_id is required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Processing heartbeat", zap.String("device_id", req.DeviceID))

	resp, err := h.deviceService.UpdateHeartbeat(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update heartbeat", zap.Error(err))
		http.Error(w, `{"error": "heartbeat failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetConfig handles device config fetch
// @Summary Get device configuration
// @Tags device
// @Produce json
// @Param device_id query string true "Device ID"
// @Success 200 {object} models.ConfigResponse
// @Router /device/config [get]
func (h *DeviceHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, `{"error": "device_id is required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Fetching device config", zap.String("device_id", deviceID))

	resp, err := h.deviceService.GetConfig(r.Context(), deviceID)
	if err != nil {
		h.logger.Error("Failed to get config", zap.Error(err))
		http.Error(w, `{"error": "config fetch failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UploadTelemetry handles batch telemetry upload
// @Summary Upload telemetry batch
// @Tags telemetry
// @Accept json
// @Produce json
// @Param request body models.TelemetryBatchRequest true "Telemetry batch"
// @Success 200 {object} models.TelemetryResponse
// @Router /telemetry/batch [post]
func (h *DeviceHandler) UploadTelemetry(w http.ResponseWriter, r *http.Request) {
	var req models.TelemetryBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode telemetry request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.BatchID == "" {
		http.Error(w, `{"error": "device_id and batch_id are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Telemetry) == 0 {
		http.Error(w, `{"error": "telemetry array cannot be empty"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Processing telemetry batch",
		zap.String("device_id", req.DeviceID),
		zap.String("batch_id", req.BatchID),
		zap.Int("points", len(req.Telemetry)))

	resp, err := h.deviceService.StoreTelemetryBatch(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to store telemetry", zap.Error(err))
		http.Error(w, `{"error": "telemetry upload failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMediaUploadURL handles media upload URL requests
// @Summary Get pre-signed media upload URL
// @Tags media
// @Accept json
// @Produce json
// @Param request body models.UploadUrlRequest true "Upload URL request"
// @Success 200 {object} models.UploadUrlResponse
// @Router /media/upload-url [post]
func (h *DeviceHandler) GetMediaUploadURL(w http.ResponseWriter, r *http.Request) {
	var req models.UploadUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode upload URL request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.ClipID == "" {
		http.Error(w, `{"error": "device_id and clip_id are required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Generating upload URL",
		zap.String("device_id", req.DeviceID),
		zap.String("clip_id", req.ClipID))

	resp, err := h.deviceService.GetMediaUploadURL(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to generate upload URL", zap.Error(err))
		http.Error(w, `{"error": "URL generation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ConfirmMediaUpload handles media upload confirmation
// @Summary Confirm media upload completion
// @Tags media
// @Accept json
// @Produce json
// @Param request body models.ConfirmUploadRequest true "Upload confirmation"
// @Success 200 {object} models.ConfirmUploadResponse
// @Router /media/confirm [post]
func (h *DeviceHandler) ConfirmMediaUpload(w http.ResponseWriter, r *http.Request) {
	var req models.ConfirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode confirm upload request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ClipID == "" || req.ChecksumSHA256 == "" {
		http.Error(w, `{"error": "clip_id and checksum_sha256 are required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Confirming media upload",
		zap.String("clip_id", req.ClipID),
		zap.String("checksum", req.ChecksumSHA256[:16]+"..."))

	resp, err := h.deviceService.ConfirmMediaUpload(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to confirm upload", zap.Error(err))
		http.Error(w, `{"error": "upload confirmation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AttestEvidence handles evidence attestation requests
// @Summary Attest evidence integrity
// @Tags forensics
// @Accept json
// @Produce json
// @Param request body models.AttestationRequest true "Attestation request"
// @Success 200 {object} models.AttestationResponse
// @Router /forensics/attest [post]
func (h *DeviceHandler) AttestEvidence(w http.ResponseWriter, r *http.Request) {
	var req models.AttestationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode attestation request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.EvidenceHash == "" || req.ChainHash == "" {
		http.Error(w, `{"error": "device_id, evidence_hash, and chain_hash are required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Processing evidence attestation",
		zap.String("device_id", req.DeviceID),
		zap.String("evidence_hash", req.EvidenceHash[:16]+"..."))

	resp, err := h.deviceService.AttestEvidence(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to attest evidence", zap.Error(err))
		http.Error(w, `{"error": "attestation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPendingCommands handles pending command requests
// @Summary Get pending device commands
// @Tags device
// @Produce json
// @Param device_id query string true "Device ID"
// @Param last_command_id query string false "Last command ID"
// @Success 200 {array} models.Command
// @Router /device/commands [get]
func (h *DeviceHandler) GetPendingCommands(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, `{"error": "device_id is required"}`, http.StatusBadRequest)
		return
	}

	lastCommandID := r.URL.Query().Get("last_command_id")

	h.logger.Debug("Fetching pending commands", zap.String("device_id", deviceID))

	commands, err := h.deviceService.GetPendingCommands(r.Context(), deviceID, lastCommandID)
	if err != nil {
		h.logger.Error("Failed to get pending commands", zap.Error(err))
		http.Error(w, `{"error": "command fetch failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"commands": commands,
	})
}

// DriverLogin handles driver login
// @Summary Driver login with HOS tracking
// @Tags driver
// @Accept json
// @Produce json
// @Param request body models.DriverLoginRequest true "Login request"
// @Success 200 {object} models.DriverLoginResponse
// @Router /driver/login [post]
func (h *DeviceHandler) DriverLogin(w http.ResponseWriter, r *http.Request) {
	var req models.DriverLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode driver login request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.DriverID == "" {
		http.Error(w, `{"error": "device_id and driver_id are required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing driver login",
		zap.String("device_id", req.DeviceID),
		zap.String("driver_id", req.DriverID))

	resp, err := h.deviceService.DriverLogin(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process driver login", zap.Error(err))
		http.Error(w, `{"error": "login failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DriverLogout handles driver logout
// @Summary Driver logout
// @Tags driver
// @Accept json
// @Produce json
// @Param request body models.DriverLogoutRequest true "Logout request"
// @Success 200 {object} models.DriverLogoutResponse
// @Router /driver/logout [post]
func (h *DeviceHandler) DriverLogout(w http.ResponseWriter, r *http.Request) {
	var req models.DriverLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode driver logout request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.DriverID == "" {
		http.Error(w, `{"error": "device_id and driver_id are required"}`, http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing driver logout",
		zap.String("device_id", req.DeviceID),
		zap.String("driver_id", req.DriverID))

	resp, err := h.deviceService.DriverLogout(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process driver logout", zap.Error(err))
		http.Error(w, `{"error": "logout failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateDriverHOS handles HOS duty status updates
// @Summary Update driver HOS duty status
// @Tags driver
// @Accept json
// @Produce json
// @Param request body models.HOSUpdateRequest true "HOS update request"
// @Success 200 {object} models.HOSUpdateResponse
// @Router /driver/hos [post]
func (h *DeviceHandler) UpdateDriverHOS(w http.ResponseWriter, r *http.Request) {
	var req models.HOSUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode HOS update request", zap.Error(err))
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.DriverID == "" || req.DutyStatus == "" {
		http.Error(w, `{"error": "device_id, driver_id, and duty_status are required"}`, http.StatusBadRequest)
		return
	}

	validDutyStatuses := map[string]bool{
		"on_duty": true, "driving": true, "sleeper": true, "off_duty": true,
	}
	if !validDutyStatuses[req.DutyStatus] {
		http.Error(w, `{"error": "invalid duty_status"}`, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Processing HOS update",
		zap.String("device_id", req.DeviceID),
		zap.String("driver_id", req.DriverID),
		zap.String("duty_status", req.DutyStatus))

	resp, err := h.deviceService.UpdateDriverHOS(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update HOS", zap.Error(err))
		http.Error(w, `{"error": "HOS update failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
