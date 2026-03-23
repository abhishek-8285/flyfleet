package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockDeviceService is a mock implementation for testing
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) RegisterDevice(ctx context.Context, req *models.DeviceRegisterRequest) (*models.DeviceRegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeviceRegisterResponse), args.Error(1)
}

func (m *MockDeviceService) UpdateHeartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HeartbeatResponse), args.Error(1)
}

func (m *MockDeviceService) GetConfig(ctx context.Context, deviceID string) (*models.ConfigResponse, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConfigResponse), args.Error(1)
}

func (m *MockDeviceService) StoreTelemetryBatch(ctx context.Context, req *models.TelemetryBatchRequest) (*models.TelemetryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TelemetryResponse), args.Error(1)
}

func (m *MockDeviceService) GetMediaUploadURL(ctx context.Context, req *models.UploadUrlRequest) (*models.UploadUrlResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UploadUrlResponse), args.Error(1)
}

func (m *MockDeviceService) ConfirmMediaUpload(ctx context.Context, req *models.ConfirmUploadRequest) (*models.ConfirmUploadResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConfirmUploadResponse), args.Error(1)
}

func (m *MockDeviceService) AttestEvidence(ctx context.Context, req *models.AttestationRequest) (*models.AttestationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AttestationResponse), args.Error(1)
}

func (m *MockDeviceService) GetPendingCommands(ctx context.Context, deviceID, lastCommandID string) ([]models.Command, error) {
	args := m.Called(ctx, deviceID, lastCommandID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Command), args.Error(1)
}

func (m *MockDeviceService) DriverLogin(ctx context.Context, req *models.DriverLoginRequest) (*models.DriverLoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DriverLoginResponse), args.Error(1)
}

func (m *MockDeviceService) DriverLogout(ctx context.Context, req *models.DriverLogoutRequest) (*models.DriverLogoutResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DriverLogoutResponse), args.Error(1)
}

func (m *MockDeviceService) UpdateDriverHOS(ctx context.Context, req *models.HOSUpdateRequest) (*models.HOSUpdateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HOSUpdateResponse), args.Error(1)
}

func setupHandlerTest(t *testing.T) (*DeviceHandler, *MockDeviceService) {
	logger := zap.NewNop()
	mockService := &MockDeviceService{}
	handler := NewDeviceHandler(mockService, logger)
	return handler, mockService
}

func TestDeviceHandler_Register_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.DeviceRegisterRequest{
		DeviceID:            "DEV-001",
		HardwareAttestation: "attestation-hash",
		AndroidVersion:      "13",
		DeviceModel:         "Fleet Tab Pro",
		AppVersion:          "2.1.0",
	}

	mockService.On("RegisterDevice", mock.Anything, mock.Anything).Return(&models.DeviceRegisterResponse{
		DeviceID:     "DEV-001",
		JWTToken:     "jwt-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresAt:    1234567890,
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/device/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "DEV-001")
	assert.Contains(t, rr.Body.String(), "jwt-token-123")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_Register_InvalidJSON(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/device/register", bytes.NewReader([]byte("invalid-json")))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid request body")
}

func TestDeviceHandler_Register_MissingDeviceID(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	reqBody := map[string]string{
		"hardware_attestation": "hash",
		"android_version":      "13",
		"device_model":         "Model",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/device/register", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "device_id is required")
}

func TestDeviceHandler_Register_MissingAttestation(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	reqBody := map[string]string{
		"device_id":       "DEV-001",
		"android_version": "13",
		"device_model":    "Model",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/device/register", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "hardware_attestation is required")
}

func TestDeviceHandler_Heartbeat_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.HeartbeatRequest{
		DeviceID: "DEV-001",
		Status:   "ok",
		ModuleHealth: models.ModuleHealth{
			Camera:  models.ModuleStatus{Status: "ok"},
			AI:      models.ModuleStatus{Status: "ok"},
			GPS:     models.ModuleStatus{Status: "ok"},
			CANBus:  models.ModuleStatus{Status: "ok"},
			Storage: models.ModuleStatus{Status: "ok"},
			Network: models.ModuleStatus{Status: "ok"},
		},
	}

	mockService.On("UpdateHeartbeat", mock.Anything, mock.Anything).Return(&models.HeartbeatResponse{
		Status:        "ok",
		ConfigVersion: "v1",
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/device/heartbeat", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Heartbeat(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_Heartbeat_MissingDeviceID(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	reqBody := map[string]string{
		"status": "ok",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/device/heartbeat", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.Heartbeat(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "device_id is required")
}

func TestDeviceHandler_GetConfig_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	mockService.On("GetConfig", mock.Anything, "DEV-001").Return(&models.ConfigResponse{
		Config: models.DeviceRuntimeConfig{
			SchemaVersion:     "v1",
			HeartbeatSeconds:  45,
			RingBufferMinutes: 10,
		},
		Signature: "sig-123",
		Version:   "v1",
		ExpiresAt: 1234567890,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/device/config?device_id=DEV-001", nil)
	rr := httptest.NewRecorder()

	handler.GetConfig(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "v1")
	assert.Contains(t, rr.Body.String(), "sig-123")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_GetConfig_MissingDeviceID(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/device/config", nil)
	rr := httptest.NewRecorder()

	handler.GetConfig(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "device_id is required")
}

func TestDeviceHandler_UploadTelemetry_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.TelemetryBatchRequest{
		DeviceID: "DEV-001",
		BatchID:  "BATCH-001",
		Telemetry: []models.TelemetryPoint{
			{
				Timestamp: 1234567890000,
				GPS: models.GPSData{
					Latitude:  37.7749,
					Longitude: -122.4194,
				},
				SpeedKph: 60.0,
			},
		},
	}

	mockService.On("StoreTelemetryBatch", mock.Anything, mock.Anything).Return(&models.TelemetryResponse{
		Status:        "accepted",
		ReceivedCount: 1,
		NextBatchID:   "BATCH-002",
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/telemetry/batch", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.UploadTelemetry(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "accepted")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_UploadTelemetry_EmptyTelemetry(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	reqBody := map[string]interface{}{
		"device_id":  "DEV-001",
		"batch_id":   "BATCH-001",
		"telemetry":  []interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/telemetry/batch", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.UploadTelemetry(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "cannot be empty")
}

func TestDeviceHandler_GetMediaUploadURL_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.UploadUrlRequest{
		DeviceID:    "DEV-001",
		ClipID:      "CLIP-001",
		SizeBytes:   10485760,
		ContentType: "video/mp4",
		StorageMode: "local",
	}

	mockService.On("GetMediaUploadURL", mock.Anything, mock.Anything).Return(&models.UploadUrlResponse{
		LocalPath:      "/sdcard/flyfleet/clips/CLIP-001.mp4",
		ObjectKey:      "clips/CLIP-001.mp4",
		StorageMode:    "local",
		MaxLocalSizeMB: 5120,
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/media/upload-url", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.GetMediaUploadURL(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "local")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_ConfirmMediaUpload_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.ConfirmUploadRequest{
		ClipID:          "CLIP-001",
		ChecksumSHA256:  "abc123def456789012345678901234567890",
		DurationSeconds: 30,
		UploadedBytes:   10485760,
		StorageMode:     "local",
	}

	mockService.On("ConfirmMediaUpload", mock.Anything, mock.Anything).Return(&models.ConfirmUploadResponse{
		Status:          "confirmed",
		StorageLocation: "/sdcard/flyfleet/clips/CLIP-001.mp4",
		StorageMode:     "local",
		SyncStatus:      "pending_sync",
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/media/confirm", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ConfirmMediaUpload(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "confirmed")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_AttestEvidence_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.AttestationRequest{
		DeviceID:     "DEV-001",
		EvidenceHash: "evidence-hash-abc123def456789012345678901234567890",
		ChainHash:    "chain-hash-xyz789012345678901234567890123456789012",
		Timestamp:    1234567890,
	}

	mockService.On("AttestEvidence", mock.Anything, mock.Anything).Return(&models.AttestationResponse{
		AttestationID: "ATT-001",
		Signature:     "signature-123",
		Timestamp:     1234567890,
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/forensics/attest", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.AttestEvidence(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ATT-001")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_GetPendingCommands_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	mockService.On("GetPendingCommands", mock.Anything, "DEV-001", "").Return([]models.Command{
		{
			ID:        "cmd-001",
			Type:      models.CommandUpdateConfig,
			Params:    map[string]interface{}{"key": "value"},
			ExpiresAt: 1234567890,
			Priority:  1,
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/device/commands?device_id=DEV-001", nil)
	rr := httptest.NewRecorder()

	handler.GetPendingCommands(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "cmd-001")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_DriverLogin_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.DriverLoginRequest{
		DeviceID:   "DEV-001",
		DriverID:   "DRV-001",
		DutyStatus: "on_duty",
	}

	mockService.On("DriverLogin", mock.Anything, mock.Anything).Return(&models.DriverLoginResponse{
		Status:     "success",
		ShiftStart: 1234567890000,
		DriverProfile: &models.DriverProfile{
			DriverID:      "DRV-001",
			Name:          "Test Driver",
			LicenseNumber: "LIC-123",
		},
		HOSStatus: &models.HOSStatus{
			DutyStatus:           "on_duty",
			DrivingTimeRemaining: 11.0,
			OnDutyTimeRemaining:  14.0,
		},
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/driver/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.DriverLogin(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "success")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_DriverLogout_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.DriverLogoutRequest{
		DeviceID:      "DEV-001",
		DriverID:      "DRV-001",
		ShiftDuration: 8.0,
		EndDutyStatus: "off_duty",
	}

	mockService.On("DriverLogout", mock.Anything, mock.Anything).Return(&models.DriverLogoutResponse{
		Status: "success",
		HOSSummary: models.HOSSummary{
			ShiftStart:   1234567890,
			ShiftEnd:     1234567890 + 8*3600,
			TotalDriving: 6.4,
			TotalOnDuty:  8.0,
		},
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/driver/logout", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.DriverLogout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "success")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_UpdateDriverHOS_Success(t *testing.T) {
	handler, mockService := setupHandlerTest(t)

	reqBody := &models.HOSUpdateRequest{
		DeviceID:   "DEV-001",
		DriverID:   "DRV-001",
		DutyStatus: "driving",
		Timestamp:  1234567890000,
		Location: &models.Location{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	mockService.On("UpdateDriverHOS", mock.Anything, mock.Anything).Return(&models.HOSUpdateResponse{
		Status:              "accepted",
		RemainingDriveTime:  10.5,
		RemainingOnDutyTime: 13.0,
	}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/driver/hos", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.UpdateDriverHOS(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "accepted")
	mockService.AssertExpectations(t)
}

func TestDeviceHandler_UpdateDriverHOS_InvalidDutyStatus(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	reqBody := map[string]string{
		"device_id":   "DEV-001",
		"driver_id":   "DRV-001",
		"duty_status": "invalid_status",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/driver/hos", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.UpdateDriverHOS(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid duty_status")
}

func TestDeviceHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockDeviceService{}
	handler := NewDeviceHandler(mockService, logger)

	// Verify handler is created successfully
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.deviceService)
}
