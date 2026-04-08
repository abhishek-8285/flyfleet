package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	platformmetrics "github.com/flyfleet/fleet-core/internal/platform/metrics"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DeviceRegistryRepository defines the device control-plane contract.
type DeviceRegistryRepository interface {
	GetByDeviceID(ctx context.Context, deviceID string) (*models.DeviceRegistry, error)
	UpsertDevice(ctx context.Context, device *models.DeviceRegistry) error
	UpdateLastHeartbeat(ctx context.Context, deviceID string) error
	CreateCommand(ctx context.Context, deviceID, cmdType string, params []byte, ttl time.Duration) (string, error)
	GetPendingCommands(ctx context.Context, deviceID string) ([]models.DeviceCommand, error)
	AckCommand(ctx context.Context, commandID string, result string) error
}

// PostgresDeviceRegistryRepository implements DeviceRegistryRepository.
type PostgresDeviceRegistryRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewDeviceRegistryRepository creates a new device registry repository.
func NewDeviceRegistryRepository(db *pgxpool.Pool, logger *zap.Logger) *PostgresDeviceRegistryRepository {
	return &PostgresDeviceRegistryRepository{db: db, logger: logger}
}

// GetByDeviceID looks up a device in the registry.
func (r *PostgresDeviceRegistryRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.DeviceRegistry, error) {
	const q = `
		SELECT device_id, vehicle_id, firmware_version, sim_iccid, carrier,
		       last_heartbeat, status, certificate_hash, created_at, updated_at
		FROM device_registry
		WHERE device_id = $1`

	start := time.Now()
	d := &models.DeviceRegistry{}
	var vehicleID *uuid.UUID
	var lastHB *time.Time
	err := r.db.QueryRow(ctx, q, deviceID).Scan(
		&d.DeviceID, &vehicleID, &d.FirmwareVersion, &d.SimICCID, &d.Carrier,
		&lastHB, &d.Status, &d.CertificateHash, &d.CreatedAt, &d.UpdatedAt,
	)
	metrics.RecordDBQuery(q, time.Since(start))
	if err != nil {
		return nil, err
	}
	if vehicleID != nil {
		d.VehicleID = vehicleID.String()
	}
	if lastHB != nil {
		d.LastHeartbeat = *lastHB
	}
	return d, nil
}

// UpsertDevice inserts or updates a device registry record.
func (r *PostgresDeviceRegistryRepository) UpsertDevice(ctx context.Context, device *models.DeviceRegistry) error {
	const q = `
		INSERT INTO device_registry (
			device_id, vehicle_id, firmware_version, sim_iccid, carrier,
			status, certificate_hash, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
		ON CONFLICT (device_id) DO UPDATE SET
			vehicle_id       = EXCLUDED.vehicle_id,
			firmware_version = EXCLUDED.firmware_version,
			sim_iccid        = EXCLUDED.sim_iccid,
			carrier          = EXCLUDED.carrier,
			status           = EXCLUDED.status,
			certificate_hash = EXCLUDED.certificate_hash,
			updated_at       = NOW()`

	start := time.Now()
	var vehicleID *string
	if device.VehicleID != "" {
		vehicleID = &device.VehicleID
	}
	_, err := r.db.Exec(ctx, q,
		device.DeviceID, vehicleID, device.FirmwareVersion,
		device.SimICCID, device.Carrier, device.Status,
		device.CertificateHash,
	)
	metrics.RecordDBQuery(q, time.Since(start))
	return err
}

// UpdateLastHeartbeat stamps the device's last_heartbeat and resets status to 'active'.
func (r *PostgresDeviceRegistryRepository) UpdateLastHeartbeat(ctx context.Context, deviceID string) error {
	const q = `
		UPDATE device_registry
		SET last_heartbeat = NOW(), status = 'active', updated_at = NOW()
		WHERE device_id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, q, deviceID)
	metrics.RecordDBQuery(q, time.Since(start))
	return err
}

// CreateCommand creates a new pending command for the device.
// Returns the generated command UUID.
func (r *PostgresDeviceRegistryRepository) CreateCommand(ctx context.Context, deviceID, cmdType string, params []byte, ttl time.Duration) (string, error) {
	const q = `
		INSERT INTO device_commands (device_id, type, params, status, issued_at, expires_at)
		VALUES ($1, $2, $3, 'pending', NOW(), $4)
		RETURNING id`

	expiresAt := time.Now().Add(ttl)
	start := time.Now()
	var commandID string
	err := r.db.QueryRow(ctx, q, deviceID, cmdType, params, expiresAt).Scan(&commandID)
	metrics.RecordDBQuery(q, time.Since(start))
	return commandID, err
}

// GetPendingCommands retrieves all non-expired pending commands for a device.
func (r *PostgresDeviceRegistryRepository) GetPendingCommands(ctx context.Context, deviceID string) ([]models.DeviceCommand, error) {
	const q = `
		SELECT id, device_id, type, params, status, issued_at, expires_at
		FROM device_commands
		WHERE device_id = $1
		  AND status = 'pending'
		  AND expires_at > NOW()
		ORDER BY issued_at ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, q, deviceID)
	metrics.RecordDBQuery(q, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []models.DeviceCommand
	for rows.Next() {
		var cmd models.DeviceCommand
		var paramsJSON []byte
		if err := rows.Scan(
			&cmd.CommandID, &cmd.DeviceID, &cmd.Type,
			&paramsJSON, &cmd.Status, &cmd.IssuedAt, &cmd.ExpiresAt,
		); err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	return cmds, rows.Err()
}

// AckCommand marks a command as executed and stores the device's result payload.
// It also records command acknowledgement latency for SLO monitoring.
func (r *PostgresDeviceRegistryRepository) AckCommand(ctx context.Context, commandID string, result string) error {
	// Fetch command metadata (issued_at, type) before updating
	const getQ = `SELECT issued_at, type FROM device_commands WHERE id = $1`
	startGet := time.Now()
	var issuedAt time.Time
	var cmdType string
	err := r.db.QueryRow(ctx, getQ, commandID).Scan(&issuedAt, &cmdType)
	metrics.RecordDBQuery(getQ, time.Since(startGet))
	if err != nil {
		return fmt.Errorf("failed to get command metadata: %w", err)
	}

	// Update to executed
	const updateQ = `
		UPDATE device_commands
		SET status = 'executed', executed_at = NOW(), result = $2
		WHERE id = $1`
	startUpdate := time.Now()
	_, err = r.db.Exec(ctx, updateQ, commandID, result)
	metrics.RecordDBQuery(updateQ, time.Since(startUpdate))
	if err != nil {
		return fmt.Errorf("failed to ack command: %w", err)
	}

	// Record latency (duration from issuance to execution)
	latency := time.Since(issuedAt)
	platformmetrics.RecordCommandLatency(cmdType, latency)

	return nil
}
