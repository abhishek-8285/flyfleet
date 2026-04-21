package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/repository"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// DeviceService handles device business logic
type DeviceService struct {
	repo          repository.DeviceRepository
	jwtSecret     string
	configSigner  ConfigSigner
	defaultConfig models.DeviceRuntimeConfig
	analytics     analytics.AnalyticsService // optional
	nc            *nats.Conn
	logger        *zap.Logger
}

// ConfigSigner signs configuration payloads
type ConfigSigner interface {
	Sign(config []byte) (string, error)
}

// NewDeviceService creates a new device service
func NewDeviceService(
	repo repository.DeviceRepository,
	jwtSecret string,
	configSigner ConfigSigner,
	analytics analytics.AnalyticsService,
	nc *nats.Conn,
	logger *zap.Logger,
) *DeviceService {
	return &DeviceService{
		repo:         repo,
		jwtSecret:    jwtSecret,
		configSigner: configSigner,
		analytics:    analytics,
		nc:           nc,
		defaultConfig: models.DeviceRuntimeConfig{
			SchemaVersion:     "v1",
			HeartbeatSeconds:  45,
			RingBufferMinutes: 10,
			GPS: models.GPSConfig{
				ActiveIntervalSeconds: 1,
				IdleIntervalSeconds:   30,
			},
			AI: models.AIConfig{
				Enabled:           true,
				FrameSampleStride: 6,
				SeverityThresholds: models.SeverityThresholds{
					WarnMin:     0.6,
					CriticalMin: 0.85,
				},
			},
			Retention: models.RetentionConfig{
				CriticalClipDays:    14,
				MinStorageReserveMB: 512,
				LegalHold:           false,
			},
			FeatureFlags: models.FeatureFlags{
				CameraEnabled:           true,
				AudioEventChecksEnabled: false,
				AIEnabled:               true,
				CANBusEnabled:           true,
				DriverLockoutEnabled:    true,
			},
		},
		logger: logger,
	}
}

// RegisterDevice registers a new device
func (s *DeviceService) RegisterDevice(ctx context.Context, req *models.DeviceRegisterRequest) (*models.DeviceRegisterResponse, error) {
	// Generate device tokens
	jwtToken, refreshToken, expiresAt, err := s.generateDeviceTokens(req.DeviceID)
	if err != nil {
		s.logger.Error("Failed to generate device tokens", zap.Error(err))
		return nil, err
	}

	userID := s.getUserID(ctx)
	device := &models.Device{
		ID:                  uuid.New(),
		DeviceID:            req.DeviceID,
		HardwareAttestation: req.HardwareAttestation,
		AndroidVersion:      req.AndroidVersion,
		DeviceModel:         req.DeviceModel,
		AppVersion:          req.AppVersion,
		Status:              "active",
		JWTToken:            jwtToken,
		RefreshToken:        refreshToken,
		TokenExpiresAt:      time.Unix(expiresAt, 0),
		CreatedBy:           userID,
		UpdatedBy:           userID,
	}

	// Parse optional IDs
	if req.OrganizationID != "" {
		if tenantID, err := uuid.Parse(req.OrganizationID); err == nil {
			device.OrganizationID = tenantID
		}
	}

	if req.VehicleID != "" {
		if vehicleID, err := uuid.Parse(req.VehicleID); err == nil {
			device.VehicleID = vehicleID
		}
	}

	// Store device
	if err := s.repo.Register(ctx, device); err != nil {
		s.logger.Error("Failed to register device", zap.Error(err))
		return nil, err
	}

	// Publish domain event
	if s.nc != nil {
		eventData := map[string]interface{}{
			"device_id":       req.DeviceID,
			"organization_id": req.OrganizationID,
			"vehicle_id":      req.VehicleID,
			"timestamp":       time.Now().Unix(),
		}
		eventBytes, _ := json.Marshal(eventData)
		if err := s.nc.Publish("fleet.device.registered", eventBytes); err != nil {
			s.logger.Warn("Failed to publish device registration event", zap.Error(err))
		}
	}

	return &models.DeviceRegisterResponse{
		DeviceID:     req.DeviceID,
		JWTToken:     jwtToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *DeviceService) getUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctxkeys.GetUserID(ctx); ok {
		return id
	}
	return uuid.Nil
}

// UpdateHeartbeat processes device heartbeat
func (s *DeviceService) UpdateHeartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error) {
	// Update device heartbeat
	if err := s.repo.UpdateHeartbeat(ctx, req.DeviceID, req.Status); err != nil {
		s.logger.Error("Failed to update heartbeat", zap.Error(err))
		return nil, err
	}

	// Get current config version
	configVersion, err := s.repo.GetConfigVersion(ctx, req.DeviceID)
	if err != nil {
		s.logger.Warn("Failed to get config version", zap.Error(err))
		configVersion = "v1"
	}

	// Get pending commands
	commands, err := s.repo.GetPendingCommands(ctx, req.DeviceID, "")
	if err != nil {
		s.logger.Warn("Failed to get pending commands", zap.Error(err))
		commands = []models.Command{}
	}

	return &models.HeartbeatResponse{
		Status:          "ok",
		ConfigVersion:   configVersion,
		PendingCommands: commands,
	}, nil
}

// GetConfig retrieves device configuration
func (s *DeviceService) GetConfig(ctx context.Context, deviceID string) (*models.ConfigResponse, error) {
	_ = ctx
	_ = deviceID
	// For now, return default config
	// In production, this would fetch device-specific config from database
	config := s.defaultConfig

	// Serialize config
	configJSON, err := json.Marshal(config)
	if err != nil {
		s.logger.Error("Failed to marshal config", zap.Error(err))
		return nil, err
	}

	// Sign config
	signature := ""
	if s.configSigner != nil {
		signature, err = s.configSigner.Sign(configJSON)
		if err != nil {
			s.logger.Error("Failed to sign config", zap.Error(err))
			return nil, err
		}
	} else {
		// Monolith fallback: deterministic hash signature when external signer is not wired.
		digest := sha256.Sum256(configJSON)
		signature = hex.EncodeToString(digest[:])
	}

	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	return &models.ConfigResponse{
		Config:    config,
		Signature: signature,
		Version:   config.SchemaVersion,
		ExpiresAt: expiresAt,
	}, nil
}

// StoreTelemetryBatch stores a batch of telemetry data
func (s *DeviceService) StoreTelemetryBatch(ctx context.Context, req *models.TelemetryBatchRequest) (*models.TelemetryResponse, error) {
	// Store telemetry batch
	if err := s.repo.StoreTelemetryBatch(ctx, req.DeviceID, req.BatchID, req.Telemetry); err != nil {
		s.logger.Error("Failed to store telemetry batch", zap.Error(err))
		return &models.TelemetryResponse{
			Status:        "rejected",
			ReceivedCount: 0,
			NextBatchID:   "",
		}, err
	}

	// Generate next batch ID
	nextBatchID := uuid.New().String()

	return &models.TelemetryResponse{
		Status:        "accepted",
		ReceivedCount: len(req.Telemetry),
		NextBatchID:   nextBatchID,
	}, nil
}

// GetMediaUploadURL generates a storage location for media
// Supports both local-first and cloud storage modes
func (s *DeviceService) GetMediaUploadURL(ctx context.Context, req *models.UploadUrlRequest) (*models.UploadUrlResponse, error) {
	// Default to local storage if not specified
	storageMode := req.StorageMode
	if storageMode == "" {
		storageMode = "local"
	}

	// Generate object key for the clip
	objectKey := "clips/" + req.ClipID + ".mp4"

	if storageMode == "local" {
		// For local storage, return local path on device
		// The Android app will store in: /storage/emulated/0/Android/data/com.flyfleet.vehicledevice/files/clips/
		localPath := "/sdcard/flyfleet/clips/" + req.ClipID + ".mp4"

		return &models.UploadUrlResponse{
			LocalPath:      localPath,
			ObjectKey:      objectKey,
			StorageMode:    "local",
			MaxLocalSizeMB: 5120, // 5GB default local storage quota
		}, nil
	}

	// For cloud storage, generate pre-signed URL (placeholder for now)
	// In production, integrate with S3/GCS SDK
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	uploadURL := "https://storage.flyfleet.local/upload/" + req.ClipID

	return &models.UploadUrlResponse{
		UploadURL:      uploadURL,
		ExpiresAt:      expiresAt,
		ObjectKey:      objectKey,
		StorageMode:    "cloud",
		MaxLocalSizeMB: 0,
	}, nil
}

// ConfirmMediaUpload confirms media upload completion
func (s *DeviceService) ConfirmMediaUpload(ctx context.Context, req *models.ConfirmUploadRequest) (*models.ConfirmUploadResponse, error) {
	// Determine storage location
	var storageLocation string
	var syncStatus string

	if req.StorageMode == "local" {
		// Local storage - mark as pending sync to cloud
		storageLocation = req.LocalPath
		syncStatus = "pending_sync"

		s.logger.Debug("Media stored locally, will sync when connected",
			zap.String("clip_id", req.ClipID),
			zap.String("local_path", storageLocation))
	} else {
		// Cloud storage - already synced
		storageLocation = "clips/" + req.ClipID + ".mp4"
		syncStatus = "synced"
	}

	// Store upload confirmation
	if err := s.repo.StoreMediaUpload(
		ctx,
		"", // deviceID - would be extracted from JWT
		req.ClipID,
		storageLocation,
		req.ChecksumSHA256,
		req.DurationSeconds,
		req.UploadedBytes,
		req.StorageMode,
		syncStatus,
	); err != nil {
		s.logger.Error("Failed to confirm media upload", zap.Error(err))
		return &models.ConfirmUploadResponse{
			Status:          "failed",
			StorageLocation: "",
			StorageMode:     req.StorageMode,
			SyncStatus:      "failed",
		}, err
	}

	return &models.ConfirmUploadResponse{
		Status:          "confirmed",
		StorageLocation: storageLocation,
		StorageMode:     req.StorageMode,
		SyncStatus:      syncStatus,
	}, nil
}

// AttestEvidence creates evidence integrity attestation
func (s *DeviceService) AttestEvidence(ctx context.Context, req *models.AttestationRequest) (*models.AttestationResponse, error) {
	// Generate attestation ID
	attestationID := uuid.New().String()

	// Create signature payload
	signaturePayload := map[string]interface{}{
		"evidence_hash": req.EvidenceHash,
		"chain_hash":    req.ChainHash,
		"timestamp":     req.Timestamp,
		"device_id":     req.DeviceID,
	}

	signatureJSON, _ := json.Marshal(signaturePayload)
	signature := base64.StdEncoding.EncodeToString(signatureJSON) // Placeholder - use real signing

	// Store attestation
	if err := s.repo.StoreAttestation(
		ctx,
		req.DeviceID,
		attestationID,
		req.EvidenceHash,
		req.ChainHash,
		signature,
		req.Timestamp,
	); err != nil {
		s.logger.Error("Failed to store attestation", zap.Error(err))
	}

	return &models.AttestationResponse{
		AttestationID: attestationID,
		Signature:     signature,
		Timestamp:     time.Now().Unix(),
	}, nil
}

// GetPendingCommands retrieves pending commands for a device
func (s *DeviceService) GetPendingCommands(ctx context.Context, deviceID, lastCommandID string) ([]models.Command, error) {
	commands, err := s.repo.GetPendingCommands(ctx, deviceID, lastCommandID)
	if err != nil {
		s.logger.Error("Failed to get pending commands", zap.Error(err))
		return []models.Command{}, err
	}

	return commands, nil
}

// DriverLogin processes driver login
func (s *DeviceService) DriverLogin(ctx context.Context, req *models.DriverLoginRequest) (*models.DriverLoginResponse, error) {
	// In production, verify RFID token and photo hash
	// For now, accept the login

	shiftStart := time.Now().Unix()

	if err := s.repo.StoreDriverSession(
		ctx,
		req.DeviceID,
		req.DriverID,
		req.RFIDToken,
		req.DutyStatus,
		shiftStart,
	); err != nil {
		s.logger.Error("Failed to store driver session", zap.Error(err))
		return nil, err
	}

	// ── Analytics Cache Invalidation ──
	if s.analytics != nil {
		if did, err := uuid.Parse(req.DriverID); err == nil {
			tenantID := s.getTenantID(ctx)
			_ = s.analytics.InvalidateDriverCache(ctx, did, tenantID)
		}
	}

	// Get HOS status
	hosStatus, err := s.repo.GetDriverHOSStatus(ctx, req.DriverID)
	if err != nil {
		s.logger.Warn("Failed to get HOS status", zap.Error(err))
		hosStatus = &models.HOSStatus{
			DutyStatus:           req.DutyStatus,
			DrivingTimeRemaining: 11.0,
			OnDutyTimeRemaining:  14.0,
			BreakRequired:        false,
			Violations:           []string{},
		}
	}

	return &models.DriverLoginResponse{
		Status: "success",
		DriverProfile: &models.DriverProfile{
			DriverID:      req.DriverID,
			Name:          "Driver " + req.DriverID, // Placeholder - fetch from driver service
			LicenseNumber: "LICENSE-" + req.DriverID,
			HOSLimit: models.HOSLimits{
				MaxDrivingHours:    11.0,
				MaxOnDutyHours:     14.0,
				MinBreakAfterHours: 8.0,
				MinBreakDuration:   0.5,
				MaxOnDutyDays:      60,
			},
		},
		HOSStatus:  hosStatus,
		ShiftStart: shiftStart * 1000, // Convert to milliseconds
	}, nil
}

// DriverLogout processes driver logout
func (s *DeviceService) DriverLogout(ctx context.Context, req *models.DriverLogoutRequest) (*models.DriverLogoutResponse, error) {
	if err := s.repo.EndDriverSession(
		ctx,
		req.DeviceID,
		req.DriverID,
		req.ShiftDuration,
		req.EndDutyStatus,
	); err != nil {
		s.logger.Error("Failed to end driver session", zap.Error(err))
		return nil, err
	}

	// ── Analytics Cache Invalidation ──
	if s.analytics != nil {
		if did, err := uuid.Parse(req.DriverID); err == nil {
			tenantID := s.getTenantID(ctx)
			_ = s.analytics.InvalidateDriverCache(ctx, did, tenantID)
		}
	}

	// Calculate HOS summary (placeholder)
	hosSummary := models.HOSSummary{
		ShiftStart:   time.Now().Add(-time.Duration(req.ShiftDuration) * time.Hour).Unix(),
		ShiftEnd:     time.Now().Unix(),
		TotalDriving: req.ShiftDuration * 0.8, // Placeholder
		TotalOnDuty:  req.ShiftDuration,
		TotalBreaks:  req.ShiftDuration * 0.1,  // Placeholder
		DistanceKm:   req.ShiftDuration * 50.0, // Placeholder - 50 km/h average
	}

	return &models.DriverLogoutResponse{
		Status:     "success",
		HOSSummary: hosSummary,
		Violations: []string{},
	}, nil
}

// UpdateDriverHOS updates driver HOS duty status
func (s *DeviceService) UpdateDriverHOS(ctx context.Context, req *models.HOSUpdateRequest) (*models.HOSUpdateResponse, error) {
	// Update HOS status
	var lat, lng float64
	if req.Location != nil {
		lat = req.Location.Latitude
		lng = req.Location.Longitude
	}

	if err := s.repo.UpdateDriverHOS(
		ctx,
		req.DeviceID,
		req.DriverID,
		req.DutyStatus,
		req.Timestamp,
		lat,
		lng,
	); err != nil {
		s.logger.Error("Failed to update HOS", zap.Error(err))
		return nil, err
	}

	// ── Analytics Cache Invalidation ──
	if s.analytics != nil && (req.DutyStatus == "OFF_DUTY" || req.DutyStatus == "SLEEPER_BERTH") {
		if did, err := uuid.Parse(req.DriverID); err == nil {
			tenantID := s.getTenantID(ctx)
			_ = s.analytics.InvalidateDriverCache(ctx, did, tenantID)
		}
	}

	// Get updated HOS status
	hosStatus, err := s.repo.GetDriverHOSStatus(ctx, req.DriverID)
	if err != nil {
		return &models.HOSUpdateResponse{
			Status:              "accepted",
			Violations:          []string{},
			RemainingDriveTime:  11.0,
			RemainingOnDutyTime: 14.0,
		}, nil
	}

	return &models.HOSUpdateResponse{
		Status:              "accepted",
		Violations:          hosStatus.Violations,
		RemainingDriveTime:  hosStatus.DrivingTimeRemaining,
		RemainingOnDutyTime: hosStatus.OnDutyTimeRemaining,
	}, nil
}

// generateDeviceTokens generates JWT and refresh tokens for device
func (s *DeviceService) generateDeviceTokens(deviceID string) (string, string, int64, error) {
	// Generate JWT token
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"device_id": deviceID,
		"type":      "device",
		"exp":       expiresAt.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", 0, err
	}

	// Generate refresh token (random bytes)
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return "", "", 0, err
	}
	refreshToken := base64.URLEncoding.EncodeToString(refreshBytes)

	return jwtToken, refreshToken, expiresAt.Unix(), nil
}

func (s *DeviceService) getTenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctxkeys.GetTenantID(ctx); ok {
		return id
	}
	return uuid.Nil
}
