package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
)

// DeviceRepository defines the interface for device data persistence
type DeviceRepository interface {
	Register(ctx context.Context, device *models.Device) error
	GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error)
	UpdateHeartbeat(ctx context.Context, deviceID string, status string) error
	GetConfigVersion(ctx context.Context, deviceID string) (string, error)
	StoreTelemetryBatch(ctx context.Context, deviceID string, batchID string, telemetry []models.TelemetryPoint) error
	StoreMediaUpload(ctx context.Context, deviceID, clipID, objectKey, checksum string, durationSeconds int, uploadedBytes int64, storageMode, syncStatus string) error
	StoreAttestation(ctx context.Context, deviceID, attestationID, evidenceHash, chainHash, signature string, timestamp int64) error
	GetPendingCommands(ctx context.Context, deviceID string, lastCommandID string) ([]models.Command, error)
	MarkCommandExecuted(ctx context.Context, commandID string) error
	StoreDriverSession(ctx context.Context, deviceID, driverID, rfidToken string, dutyStatus string, shiftStart int64) error
	EndDriverSession(ctx context.Context, deviceID, driverID string, shiftDuration float64, endDutyStatus string) error
	UpdateDriverHOS(ctx context.Context, deviceID, driverID, dutyStatus string, timestamp int64, latitude, longitude float64) error
	GetDriverHOSStatus(ctx context.Context, driverID string) (*models.HOSStatus, error)
}

// PostgresDeviceRepository implements DeviceRepository with PostgreSQL
type PostgresDeviceRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *pgxpool.Pool, logger *zap.Logger) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{
		db:     db,
		logger: logger,
	}
}

// Register creates a new device registration
func (r *PostgresDeviceRepository) Register(ctx context.Context, device *models.Device) error {
	query := `
		INSERT INTO devices (
			id, device_id, organization_id, vehicle_id, hardware_attestation,
			android_version, device_model, app_version, status, jwt_token,
			refresh_token, token_expires_at, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), $13, NOW(), $14)
		ON CONFLICT (device_id) DO UPDATE SET
			organization_id = EXCLUDED.organization_id,
			vehicle_id = EXCLUDED.vehicle_id,
			hardware_attestation = EXCLUDED.hardware_attestation,
			android_version = EXCLUDED.android_version,
			device_model = EXCLUDED.device_model,
			app_version = EXCLUDED.app_version,
			status = EXCLUDED.status,
			jwt_token = EXCLUDED.jwt_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
	`

	_, err := r.db.Exec(ctx, query,
		device.ID,
		device.DeviceID,
		device.OrganizationID,
		device.VehicleID,
		device.HardwareAttestation,
		device.AndroidVersion,
		device.DeviceModel,
		device.AppVersion,
		device.Status,
		device.JWTToken,
		device.RefreshToken,
		device.TokenExpiresAt,
		device.CreatedBy,
		device.UpdatedBy,
	)

	return err
}

// GetByDeviceID retrieves a device by device ID
func (r *PostgresDeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	query := `
		SELECT id, device_id, organization_id, vehicle_id, hardware_attestation,
		       android_version, device_model, app_version, status, jwt_token,
		       refresh_token, token_expires_at, last_heartbeat, created_at, created_by, updated_at, updated_by
		FROM devices
		WHERE device_id = $1
	`

	device := &models.Device{}
	err := r.db.QueryRow(ctx, query, deviceID).Scan(
		&device.ID,
		&device.DeviceID,
		&device.OrganizationID,
		&device.VehicleID,
		&device.HardwareAttestation,
		&device.AndroidVersion,
		&device.DeviceModel,
		&device.AppVersion,
		&device.Status,
		&device.JWTToken,
		&device.RefreshToken,
		&device.TokenExpiresAt,
		&device.LastHeartbeat,
		&device.CreatedAt,
		&device.CreatedBy,
		&device.UpdatedAt,
		&device.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return device, err
}

// UpdateHeartbeat updates the device heartbeat timestamp and health status
func (r *PostgresDeviceRepository) UpdateHeartbeat(ctx context.Context, deviceID string, status string) error {
	query := `
		UPDATE devices
		SET last_heartbeat = NOW(),
		    status = $2,
		    updated_at = NOW()
		WHERE device_id = $1
	`

	_, err := r.db.Exec(ctx, query, deviceID, status)
	return err
}

// GetConfigVersion retrieves the current config version for a device
func (r *PostgresDeviceRepository) GetConfigVersion(ctx context.Context, deviceID string) (string, error) {
	query := `
		SELECT COALESCE(config_version, 'v1')
		FROM device_configs
		WHERE device_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var configVersion string
	err := r.db.QueryRow(ctx, query, deviceID).Scan(&configVersion)
	if err == sql.ErrNoRows {
		return "v1", nil
	}
	return configVersion, err
}

// StoreTelemetryBatch stores a batch of telemetry data
func (r *PostgresDeviceRepository) StoreTelemetryBatch(ctx context.Context, deviceID string, batchID string, telemetry []models.TelemetryPoint) error {
	if len(telemetry) == 0 {
		return nil
	}

	// Batch insert to avoid N+1 queries - critical for high-throughput telemetry
	if len(telemetry) > 0 {
		var args []interface{}
		var placeholders []string
		for i, point := range telemetry {
			var canBusDataJSON []byte
			if point.CANBus != nil {
				// In production, use json.Marshal
				canBusDataJSON = []byte("{}") // Placeholder
			}

			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, NOW())",
				i*13+1, i*13+2, i*13+3, i*13+4, i*13+5, i*13+6, i*13+7, i*13+8, i*13+9, i*13+10, i*13+11, i*13+12))
			args = append(args,
				uuid.New().String(),
				deviceID,
				batchID,
				point.Timestamp,
				point.GPS.Latitude,
				point.GPS.Longitude,
				point.SpeedKph,
				point.Heading,
				point.Altitude,
				point.GPS.Accuracy,
				point.GPS.Satellites,
				canBusDataJSON,
			)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO telemetry_points (id, device_id, batch_id, timestamp, latitude, longitude, "+
				"speed_kph, heading, altitude, accuracy, satellites, can_bus_data, created_at) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err := r.db.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert telemetry: %w", err)
		}
	}

	return nil
}

// StoreMediaUpload stores media upload metadata
func (r *PostgresDeviceRepository) StoreMediaUpload(ctx context.Context, deviceID, clipID, objectKey, checksum string, durationSeconds int, uploadedBytes int64, storageMode, syncStatus string) error {
	query := `
		INSERT INTO media_uploads (
			id, device_id, clip_id, object_key, checksum_sha256,
			duration_seconds, uploaded_bytes, status, storage_mode, sync_status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'uploaded', $8, $9, NOW(), NOW())
		ON CONFLICT (clip_id) DO UPDATE SET
			object_key = EXCLUDED.object_key,
			checksum_sha256 = EXCLUDED.checksum_sha256,
			duration_seconds = EXCLUDED.duration_seconds,
			uploaded_bytes = EXCLUDED.uploaded_bytes,
			storage_mode = EXCLUDED.storage_mode,
			sync_status = EXCLUDED.sync_status,
			status = 'uploaded',
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		uuid.New().String(),
		deviceID,
		clipID,
		objectKey,
		checksum,
		durationSeconds,
		uploadedBytes,
		storageMode,
		syncStatus,
	)

	return err
}

// StoreAttestation stores evidence attestation
func (r *PostgresDeviceRepository) StoreAttestation(ctx context.Context, deviceID, attestationID, evidenceHash, chainHash, signature string, timestamp int64) error {
	query := `
		INSERT INTO evidence_attestations (
			id, attestation_id, device_id, evidence_hash, chain_hash,
			signature, timestamp, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	_, err := r.db.Exec(ctx, query,
		uuid.New().String(),
		attestationID,
		deviceID,
		evidenceHash,
		chainHash,
		signature,
		timestamp,
	)

	return err
}

// GetPendingCommands retrieves pending commands for a device
func (r *PostgresDeviceRepository) GetPendingCommands(ctx context.Context, deviceID string, lastCommandID string) ([]models.Command, error) {
	query := `
		SELECT id, type, params, expires_at, priority, created_at
		FROM device_commands
		WHERE device_id = $1
		  AND status = 'pending'
		  AND expires_at > NOW()
		  AND ( $2 = '' OR id > $2 )
		ORDER BY priority ASC, created_at ASC
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, query, deviceID, lastCommandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []models.Command
	for rows.Next() {
		var cmd models.Command
		var paramsJSON []byte

		err := rows.Scan(
			&cmd.ID,
			&cmd.Type,
			&paramsJSON,
			&cmd.ExpiresAt,
			&cmd.Priority,
			&cmd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// In production, use json.Unmarshal
		cmd.Params = make(map[string]interface{})

		commands = append(commands, cmd)
	}

	return commands, rows.Err()
}

// MarkCommandExecuted marks a command as executed
func (r *PostgresDeviceRepository) MarkCommandExecuted(ctx context.Context, commandID string) error {
	query := `
		UPDATE device_commands
		SET status = 'executed',
		    executed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, commandID)
	return err
}

// StoreDriverSession stores driver login session
func (r *PostgresDeviceRepository) StoreDriverSession(ctx context.Context, deviceID, driverID, rfidToken string, dutyStatus string, shiftStart int64) error {
	query := `
		INSERT INTO driver_sessions (
			id, device_id, driver_id, rfid_token, duty_status,
			shift_start, shift_end, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, to_timestamp($6), NULL, 'active', NOW(), NOW())
		ON CONFLICT (device_id, driver_id) DO UPDATE SET
			duty_status = EXCLUDED.duty_status,
			shift_start = EXCLUDED.shift_start,
			status = 'active',
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		uuid.New().String(),
		deviceID,
		driverID,
		rfidToken,
		dutyStatus,
		shiftStart/1000, // Convert ms to seconds
	)

	return err
}

// EndDriverSession ends the current driver session
func (r *PostgresDeviceRepository) EndDriverSession(ctx context.Context, deviceID, driverID string, shiftDuration float64, endDutyStatus string) error {
	query := `
		UPDATE driver_sessions
		SET shift_end = NOW(),
		    duty_status = $3,
		    status = 'completed',
		    updated_at = NOW()
		WHERE device_id = $1 AND driver_id = $2 AND status = 'active'
	`

	_, err := r.db.Exec(ctx, query, deviceID, driverID, endDutyStatus)
	return err
}

// UpdateDriverHOS updates driver HOS duty status
func (r *PostgresDeviceRepository) UpdateDriverHOS(ctx context.Context, deviceID, driverID, dutyStatus string, timestamp int64, latitude, longitude float64) error {
	query := `
		INSERT INTO driver_hos_logs (
			id, device_id, driver_id, duty_status, timestamp,
			location, created_at
		) VALUES ($1, $2, $3, $4, to_timestamp($5), ST_Point($6, $7), NOW())
	`

	_, err := r.db.Exec(ctx, query,
		uuid.New().String(),
		deviceID,
		driverID,
		dutyStatus,
		timestamp/1000, // Convert ms to seconds
		longitude,
		latitude,
	)

	return err
}

// GetDriverHOSStatus retrieves current HOS status for a driver
func (r *PostgresDeviceRepository) GetDriverHOSStatus(ctx context.Context, driverID string) (*models.HOSStatus, error) {
	query := `
		SELECT
			COALESCE(MAX(duty_status), 'off_duty') as current_duty_status,
			11.0 - COALESCE(SUM(CASE WHEN duty_status = 'driving' THEN 1 ELSE 0 END), 0) as driving_remaining,
			14.0 - COALESCE(SUM(CASE WHEN duty_status IN ('driving', 'on_duty') THEN 1 ELSE 0 END), 0) as on_duty_remaining,
			false as break_required,
			ARRAY[]::text[] as violations
		FROM driver_hos_logs
		WHERE driver_id = $1
		  AND timestamp > NOW() - INTERVAL '14 hours'
	`

	hosStatus := &models.HOSStatus{}
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&hosStatus.DutyStatus,
		&hosStatus.DrivingTimeRemaining,
		&hosStatus.OnDutyTimeRemaining,
		&hosStatus.BreakRequired,
		&hosStatus.Violations,
	)

	if err == sql.ErrNoRows {
		return &models.HOSStatus{
			DutyStatus:           "off_duty",
			DrivingTimeRemaining: 11.0,
			OnDutyTimeRemaining:  14.0,
			BreakRequired:        false,
			Violations:           []string{},
		}, nil
	}

	return hosStatus, err
}
