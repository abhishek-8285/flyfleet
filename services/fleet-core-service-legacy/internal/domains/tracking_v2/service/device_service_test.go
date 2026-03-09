//go:build integration
// +build integration

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockDeviceRepository is a mock implementation for testing
type MockDeviceRepository struct {
	Devices        map[string]*models.Device
	Commands       map[string][]models.Command
	DriverSessions map[string]*DriverSession
	HOSLogs        map[string][]*HOSLog
	MediaUploads   map[string]*MediaUploadData
	Attestations   map[string]*AttestationData
	LastHeartbeat  map[string]time.Time
}

// Helper structs for mock storage
type DriverSession struct {
	DeviceID   string
	DriverID   string
	RFIDToken  string
	DutyStatus string
	ShiftStart int64
}

type HOSLog struct {
	DeviceID   string
	DriverID   string
	DutyStatus string
	Timestamp  int64
	Latitude   float64
	Longitude  float64
}

type MediaUploadData struct {
	ClipID      string
	ObjectKey   string
	Checksum    string
	StorageMode string
	SyncStatus  string
}

type AttestationData struct {
	AttestationID string
	DeviceID      string
	EvidenceHash  string
	ChainHash     string
	Signature     string
	Timestamp     int64
}

func NewMockDeviceRepository() *MockDeviceRepository {
	return &MockDeviceRepository{
		Devices:        make(map[string]*models.Device),
		Commands:       make(map[string][]models.Command),
		DriverSessions: make(map[string]*DriverSession),
		HOSLogs:        make(map[string][]*HOSLog),
		MediaUploads:   make(map[string]*MediaUploadData),
		Attestations:   make(map[string]*AttestationData),
		LastHeartbeat:  make(map[string]time.Time),
	}
}

func (m *MockDeviceRepository) Register(ctx context.Context, device *models.Device) error {
	m.Devices[device.DeviceID] = device
	return nil
}

func (m *MockDeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	device, ok := m.Devices[deviceID]
	if !ok {
		return nil, nil
	}
	return device, nil
}

func (m *MockDeviceRepository) UpdateHeartbeat(ctx context.Context, deviceID string, status string) error {
	m.LastHeartbeat[deviceID] = time.Now()
	if device, ok := m.Devices[deviceID]; ok {
		device.Status = status
	}
	return nil
}

func (m *MockDeviceRepository) GetConfigVersion(ctx context.Context, deviceID string) (string, error) {
	return "v1", nil
}

func (m *MockDeviceRepository) StoreTelemetryBatch(ctx context.Context, deviceID string, batchID string, telemetry []models.TelemetryPoint) error {
	return nil
}

func (m *MockDeviceRepository) StoreMediaUpload(ctx context.Context, deviceID, clipID, objectKey, checksum string, durationSeconds int, uploadedBytes int64, storageMode, syncStatus string) error {
	m.MediaUploads[clipID] = &MediaUploadData{
		ClipID:      clipID,
		ObjectKey:   objectKey,
		Checksum:    checksum,
		StorageMode: storageMode,
		SyncStatus:  syncStatus,
	}
	return nil
}

func (m *MockDeviceRepository) StoreAttestation(ctx context.Context, deviceID, attestationID, evidenceHash, chainHash, signature string, timestamp int64) error {
	m.Attestations[attestationID] = &AttestationData{
		AttestationID: attestationID,
		DeviceID:      deviceID,
		EvidenceHash:  evidenceHash,
		ChainHash:     chainHash,
		Signature:     signature,
		Timestamp:     timestamp,
	}
	return nil
}

func (m *MockDeviceRepository) GetPendingCommands(ctx context.Context, deviceID string, lastCommandID string) ([]models.Command, error) {
	commands, ok := m.Commands[deviceID]
	if !ok {
		return []models.Command{}, nil
	}
	return commands, nil
}

func (m *MockDeviceRepository) MarkCommandExecuted(ctx context.Context, commandID string) error {
	return nil
}

func (m *MockDeviceRepository) StoreDriverSession(ctx context.Context, deviceID, driverID, rfidToken string, dutyStatus string, shiftStart int64) error {
	m.DriverSessions[deviceID+"-"+driverID] = &DriverSession{
		DeviceID:   deviceID,
		DriverID:   driverID,
		RFIDToken:  rfidToken,
		DutyStatus: dutyStatus,
		ShiftStart: shiftStart,
	}
	return nil
}

func (m *MockDeviceRepository) EndDriverSession(ctx context.Context, deviceID, driverID string, shiftDuration float64, endDutyStatus string) error {
	key := deviceID + "-" + driverID
	if session, ok := m.DriverSessions[key]; ok {
		session.DutyStatus = endDutyStatus
	}
	return nil
}

func (m *MockDeviceRepository) UpdateDriverHOS(ctx context.Context, deviceID, driverID, dutyStatus string, timestamp int64, latitude, longitude float64) error {
	key := deviceID + "-" + driverID
	m.HOSLogs[key] = append(m.HOSLogs[key], &HOSLog{
		DeviceID:   deviceID,
		DriverID:   driverID,
		DutyStatus: dutyStatus,
		Timestamp:  timestamp,
		Latitude:   latitude,
		Longitude:  longitude,
	})
	return nil
}

func (m *MockDeviceRepository) GetDriverHOSStatus(ctx context.Context, driverID string) (*models.HOSStatus, error) {
	return &models.HOSStatus{
		DutyStatus:           "on_duty",
		DrivingTimeRemaining: 11.0,
		OnDutyTimeRemaining:  14.0,
		BreakRequired:        false,
		Violations:           []string{},
	}, nil
}

func setupTestService(t *testing.T) (*DeviceService, *MockDeviceRepository) {
	logger := zap.NewNop()
	mockRepo := NewMockDeviceRepository()
	service := NewDeviceService(mockRepo, "test-secret-key", nil, logger)
	return service, mockRepo
}

func TestRegisterDevice_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), "user_id", uuid.New().String())

	req := &models.DeviceRegisterRequest{
		DeviceID:            "DEV-001",
		HardwareAttestation: "attestation-hash",
		AndroidVersion:      "13",
		DeviceModel:         "Fleet Tab Pro",
		AppVersion:          "2.1.0",
	}

	resp, err := service.RegisterDevice(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "DEV-001", resp.DeviceID)
	assert.NotEmpty(t, resp.JWTToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Greater(t, resp.ExpiresAt, time.Now().Unix())

	// Verify device was stored
	storedDevice := mockRepo.Devices["DEV-001"]
	assert.NotNil(t, storedDevice)
	assert.Equal(t, "active", storedDevice.Status)
}

func TestRegisterDevice_MissingDeviceID(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.DeviceRegisterRequest{
		HardwareAttestation: "attestation-hash",
		AndroidVersion:      "13",
		DeviceModel:         "Fleet Tab Pro",
	}

	// Service layer doesn't validate - that's the handler's job
	// The service will create a device with empty device_id
	resp, err := service.RegisterDevice(ctx, req)

	// The service generates tokens even for empty device_id
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.JWTToken)
}

func TestRegisterDevice_WithOrganizationAndVehicle(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.WithValue(context.Background(), "user_id", uuid.New().String())

	orgID := uuid.New()
	vehicleID := uuid.New()

	req := &models.DeviceRegisterRequest{
		DeviceID:            "DEV-002",
		HardwareAttestation: "attestation-hash",
		AndroidVersion:      "13",
		DeviceModel:         "Fleet Tab Pro",
		OrganizationID:      orgID.String(),
		VehicleID:           vehicleID.String(),
	}

	resp, err := service.RegisterDevice(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	storedDevice := mockRepo.Devices["DEV-002"]
	assert.NotNil(t, storedDevice)
	assert.Equal(t, orgID, storedDevice.OrganizationID)
	assert.Equal(t, vehicleID, storedDevice.VehicleID)
}

func TestUpdateHeartbeat_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	// First register a device
	regReq := &models.DeviceRegisterRequest{
		DeviceID:            "DEV-003",
		HardwareAttestation: "attestation-hash",
		AndroidVersion:      "13",
		DeviceModel:         "Fleet Tab Pro",
	}
	_, _ = service.RegisterDevice(ctx, regReq)

	req := &models.HeartbeatRequest{
		DeviceID: "DEV-003",
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

	resp, err := service.UpdateHeartbeat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "v1", resp.ConfigVersion)

	// Verify heartbeat was updated
	assert.NotNil(t, mockRepo.LastHeartbeat["DEV-003"])
}

func TestUpdateHeartbeat_WithPendingCommands(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	// Register device
	regReq := &models.DeviceRegisterRequest{
		DeviceID:            "DEV-004",
		HardwareAttestation: "attestation-hash",
		AndroidVersion:      "13",
		DeviceModel:         "Fleet Tab Pro",
	}
	_, _ = service.RegisterDevice(ctx, regReq)

	// Add pending command
	mockRepo.Commands["DEV-004"] = []models.Command{
		{
			ID:        "cmd-001",
			Type:      models.CommandUpdateConfig,
			Params:    map[string]interface{}{"heartbeat_seconds": 30},
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
			Priority:  1,
			CreatedAt: time.Now().Unix(),
		},
	}

	req := &models.HeartbeatRequest{
		DeviceID: "DEV-004",
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

	resp, err := service.UpdateHeartbeat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.PendingCommands, 1)
	assert.Equal(t, "cmd-001", resp.PendingCommands[0].ID)
}

func TestGetConfig_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	resp, err := service.GetConfig(ctx, "DEV-001")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "v1", resp.Version)
	assert.NotNil(t, resp.Config)
	assert.NotEmpty(t, resp.Signature)
	assert.Greater(t, resp.ExpiresAt, time.Now().Unix())

	// Verify config structure
	assert.Equal(t, "v1", resp.Config.SchemaVersion)
	assert.Greater(t, resp.Config.HeartbeatSeconds, 0)
	assert.True(t, resp.Config.AI.Enabled)
	assert.True(t, resp.Config.FeatureFlags.CameraEnabled)
}

func TestStoreTelemetryBatch_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.TelemetryBatchRequest{
		DeviceID: "DEV-001",
		BatchID:  "BATCH-001",
		Telemetry: []models.TelemetryPoint{
			{
				Timestamp: time.Now().UnixMilli(),
				GPS: models.GPSData{
					Latitude:  37.7749,
					Longitude: -122.4194,
					Accuracy:  5.0,
				},
				SpeedKph: 60.0,
				Heading:  180.0,
			},
			{
				Timestamp: time.Now().Add(1 * time.Second).UnixMilli(),
				GPS: models.GPSData{
					Latitude:  37.7750,
					Longitude: -122.4195,
					Accuracy:  5.0,
				},
				SpeedKph: 62.0,
				Heading:  185.0,
			},
		},
	}

	resp, err := service.StoreTelemetryBatch(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "accepted", resp.Status)
	assert.Equal(t, 2, resp.ReceivedCount)
	assert.NotEmpty(t, resp.NextBatchID)
}

func TestStoreTelemetryBatch_EmptyTelemetry(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.TelemetryBatchRequest{
		DeviceID:  "DEV-001",
		BatchID:   "BATCH-002",
		Telemetry: []models.TelemetryPoint{},
	}

	resp, err := service.StoreTelemetryBatch(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "accepted", resp.Status)
	assert.Equal(t, 0, resp.ReceivedCount)
}

func TestGetMediaUploadURL_LocalStorage(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.UploadUrlRequest{
		DeviceID:    "DEV-001",
		ClipID:      "CLIP-001",
		SizeBytes:   10485760,
		ContentType: "video/mp4",
		StorageMode: "local",
	}

	resp, err := service.GetMediaUploadURL(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "local", resp.StorageMode)
	assert.NotEmpty(t, resp.LocalPath)
	assert.Contains(t, resp.LocalPath, "CLIP-001.mp4")
	assert.Greater(t, resp.MaxLocalSizeMB, int64(0))
}

func TestGetMediaUploadURL_CloudStorage(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.UploadUrlRequest{
		DeviceID:    "DEV-001",
		ClipID:      "CLIP-002",
		SizeBytes:   10485760,
		ContentType: "video/mp4",
		StorageMode: "cloud",
	}

	resp, err := service.GetMediaUploadURL(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "cloud", resp.StorageMode)
	assert.NotEmpty(t, resp.UploadURL)
	assert.Greater(t, resp.ExpiresAt, time.Now().Unix())
}

func TestConfirmMediaUpload_LocalStorage(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.ConfirmUploadRequest{
		ClipID:          "CLIP-003",
		ChecksumSHA256:  "abc123def456",
		DurationSeconds: 30,
		UploadedBytes:   10485760,
		StorageMode:     "local",
		LocalPath:       "/sdcard/flyfleet/clips/CLIP-003.mp4",
	}

	resp, err := service.ConfirmMediaUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "confirmed", resp.Status)
	assert.Equal(t, "local", resp.StorageMode)
	assert.Equal(t, "pending_sync", resp.SyncStatus)
	assert.Contains(t, resp.StorageLocation, "CLIP-003.mp4")
}

func TestConfirmMediaUpload_CloudStorage(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.ConfirmUploadRequest{
		ClipID:          "CLIP-004",
		ChecksumSHA256:  "xyz789abc123",
		DurationSeconds: 60,
		UploadedBytes:   20971520,
		StorageMode:     "cloud",
	}

	resp, err := service.ConfirmMediaUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "confirmed", resp.Status)
	assert.Equal(t, "cloud", resp.StorageMode)
	assert.Equal(t, "synced", resp.SyncStatus)
}

func TestAttestEvidence_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.AttestationRequest{
		DeviceID:     "DEV-001",
		EvidenceHash: "evidence-hash-abc123",
		ChainHash:    "chain-hash-xyz789",
		Timestamp:    time.Now().Unix(),
	}

	resp, err := service.AttestEvidence(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AttestationID)
	assert.NotEmpty(t, resp.Signature)
	assert.Greater(t, resp.Timestamp, int64(0))
}

func TestGetPendingCommands_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	// Add pending commands
	mockRepo.Commands["DEV-005"] = []models.Command{
		{
			ID:        "cmd-001",
			Type:      models.CommandReboot,
			Params:    map[string]interface{}{},
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
			Priority:  1,
			CreatedAt: time.Now().Unix(),
		},
		{
			ID:        "cmd-002",
			Type:      models.CommandUpdateConfig,
			Params:    map[string]interface{}{"key": "value"},
			ExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			Priority:  2,
			CreatedAt: time.Now().Unix(),
		},
	}

	commands, err := service.GetPendingCommands(ctx, "DEV-005", "")

	assert.NoError(t, err)
	assert.Len(t, commands, 2)
	assert.Equal(t, "cmd-001", commands[0].ID)
}

func TestDriverLogin_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.DriverLoginRequest{
		DeviceID:   "DEV-001",
		DriverID:   "DRV-001",
		RFIDToken:  "rfid-token-123",
		DutyStatus: "on_duty",
	}

	resp, err := service.DriverLogin(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.NotNil(t, resp.DriverProfile)
	assert.Equal(t, "DRV-001", resp.DriverProfile.DriverID)
	assert.NotNil(t, resp.HOSStatus)
	assert.Greater(t, resp.ShiftStart, int64(0))
}

func TestDriverLogout_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	// First login
	loginReq := &models.DriverLoginRequest{
		DeviceID:   "DEV-001",
		DriverID:   "DRV-002",
		DutyStatus: "on_duty",
	}
	_, _ = service.DriverLogin(ctx, loginReq)

	// Then logout
	logoutReq := &models.DriverLogoutRequest{
		DeviceID:      "DEV-001",
		DriverID:      "DRV-002",
		ShiftDuration: 8.0,
		EndDutyStatus: "off_duty",
	}

	resp, err := service.DriverLogout(ctx, logoutReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, 8.0, resp.HOSSummary.TotalOnDuty)
	assert.Empty(t, resp.Violations)
}

func TestUpdateDriverHOS_Success(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.HOSUpdateRequest{
		DeviceID:   "DEV-001",
		DriverID:   "DRV-003",
		DutyStatus: "driving",
		Timestamp:  time.Now().UnixMilli(),
		Location: &models.Location{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	resp, err := service.UpdateDriverHOS(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "accepted", resp.Status)
	assert.Empty(t, resp.Violations)
}

func TestUpdateDriverHOS_InvalidDutyStatus(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &models.HOSUpdateRequest{
		DeviceID:   "DEV-001",
		DriverID:   "DRV-004",
		DutyStatus: "invalid_status",
		Timestamp:  time.Now().UnixMilli(),
	}

	// The handler validates duty status, not the service
	// Service should accept any status
	resp, err := service.UpdateDriverHOS(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGenerateDeviceTokens(t *testing.T) {
	service, _ := setupTestService(t)

	jwtToken, refreshToken, expiresAt, err := service.generateDeviceTokens("DEV-TEST")

	assert.NoError(t, err)
	assert.NotEmpty(t, jwtToken)
	assert.NotEmpty(t, refreshToken)
	assert.Greater(t, expiresAt, time.Now().Unix())

	// Verify JWT structure
	parts := splitString(jwtToken, ".")
	assert.Equal(t, 3, len(parts))

	// Decode and verify claims
	claimsJSON, err := base64Decode(parts[1])
	assert.NoError(t, err)

	var claims map[string]interface{}
	err = json.Unmarshal(claimsJSON, &claims)
	assert.NoError(t, err)
	assert.Equal(t, "DEV-TEST", claims["device_id"])
	assert.Equal(t, "device", claims["type"])
	assert.Contains(t, claims, "exp")
	assert.Contains(t, claims, "iat")
}

// Helper functions
func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func base64Decode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
