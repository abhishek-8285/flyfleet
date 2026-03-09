package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository interface {
	// Devices
	GetDeviceSummary(ctx context.Context, orgID uuid.UUID) (*models.DeviceSummary, error)
	ListDevices(ctx context.Context, orgID uuid.UUID, status string, limit, offset int) ([]*models.Device, error)
	GetDevice(ctx context.Context, id, orgID uuid.UUID) (*models.Device, error)
	UpdateDeviceHealth(ctx context.Context, deviceID uuid.UUID, battery, signal *int, firmware *string) error

	// Telemetry
	GetLatestTelemetry(ctx context.Context, deviceID uuid.UUID) (*models.DeviceTelemetry, error)
	ListTelemetry(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceTelemetry, error)
	CreateTelemetry(ctx context.Context, t *models.DeviceTelemetry) error

	// Commands
	ListCommands(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error)
	GetCommand(ctx context.Context, commandID uuid.UUID) (*models.DeviceCommand, error)
	CreateCommand(ctx context.Context, cmd *models.DeviceCommand) error
	UpdateCommandStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}, errorMsg *string) error

	// Device twins
	ListDeviceTwins(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DeviceTwin, error)
	GetDeviceTwin(ctx context.Context, deviceID, orgID uuid.UUID) (*models.DeviceTwin, error)
	UpsertDesiredState(ctx context.Context, deviceID, orgID uuid.UUID, desiredState map[string]interface{}, lastCommandID *uuid.UUID, lastCommandStatus *string) error
	UpsertReportedState(ctx context.Context, deviceID, orgID uuid.UUID, reportedState map[string]interface{}, lastSeen *time.Time) error

	// Firmware
	ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error)
	CreateFirmware(ctx context.Context, fw *models.FirmwarePackage) error
	GetFirmware(ctx context.Context, id uuid.UUID) (*models.FirmwarePackage, error)

	// Rollouts
	ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error)
	GetRollout(ctx context.Context, rolloutID uuid.UUID) (*models.FirmwareRollout, error)
	CreateRollout(ctx context.Context, r *models.FirmwareRollout) error
	UpdateRolloutProgress(ctx context.Context, id uuid.UUID, completed, failed int, status string) error
	UpdateRolloutStatus(ctx context.Context, id uuid.UUID, status string) error
	ListRolloutEvents(ctx context.Context, rolloutID, orgID uuid.UUID, eventType string, limit, offset int) ([]*models.FirmwareRolloutEvent, error)
	CreateRolloutEvent(ctx context.Context, event *models.FirmwareRolloutEvent) error

	// Logs
	ListDiagnosticsLogs(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DiagnosticsLog, error)
	CreateDiagnosticsLog(ctx context.Context, log *models.DiagnosticsLog) error

	// Audit logs
	CreateAuditLog(ctx context.Context, audit *models.DiagnosticsAuditLog) error
	ListAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.DiagnosticsAuditLog, int, error)
}

type postgresRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) Repository {
	return &postgresRepository{db: db, logger: logger}
}

// Device implementations

func (r *postgresRepository) GetDeviceSummary(ctx context.Context, orgID uuid.UUID) (*models.DeviceSummary, error) {
	summary := &models.DeviceSummary{}

	// Total devices
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false`, orgID).Scan(&summary.TotalDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to count total devices: %w", err)
	}

	// Online (heartbeat within 5 minutes)
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND last_heartbeat > NOW() - INTERVAL '5 minutes'`, orgID).Scan(&summary.OnlineDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to count online devices: %w", err)
	}

	summary.OfflineDevices = summary.TotalDevices - summary.OnlineDevices

	// Low battery (< 20%)
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND battery_level IS NOT NULL AND battery_level < 20`, orgID).Scan(&summary.LowBattery)
	if err != nil {
		return nil, fmt.Errorf("failed to count low battery devices: %w", err)
	}

	// Weak signal (< -80 dBm)
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND signal_strength IS NOT NULL AND signal_strength < -80`, orgID).Scan(&summary.WeakSignal)
	if err != nil {
		return nil, fmt.Errorf("failed to count weak signal devices: %w", err)
	}

	// System errors (status = 'error')
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND status = 'error'`, orgID).Scan(&summary.SystemErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to count error devices: %w", err)
	}

	// New devices this week
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND created_at > NOW() - INTERVAL '7 days'`, orgID).Scan(&summary.NewDevicesWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to count new devices: %w", err)
	}

	return summary, nil
}

func (r *postgresRepository) ListDevices(ctx context.Context, orgID uuid.UUID, status string, limit, offset int) ([]*models.Device, error) {
	query := `SELECT id, org_id, vehicle_id, serial_number, type, model, manufacturer, firmware_version, hardware_revision, battery_level, signal_strength, status, last_ping_at, last_heartbeat, is_deleted, created_at, updated_at
	          FROM devices WHERE org_id = $1 AND is_deleted = false`
	args := []interface{}{orgID}

	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d := &models.Device{}
		if err := rows.Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.SerialNumber, &d.Type, &d.Model, &d.Manufacturer, &d.FirmwareVersion, &d.HardwareRevision, &d.BatteryLevel, &d.SignalStrength, &d.Status, &d.LastPingAt, &d.LastHeartbeat, &d.IsDeleted, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("device iteration error: %w", err)
	}

	return devices, nil
}

func (r *postgresRepository) GetDevice(ctx context.Context, id, orgID uuid.UUID) (*models.Device, error) {
	d := &models.Device{}
	err := r.db.QueryRow(ctx, `SELECT id, org_id, vehicle_id, serial_number, type, model, manufacturer, firmware_version, hardware_revision, battery_level, signal_strength, status, last_ping_at, last_heartbeat, is_deleted, created_at, updated_at
	          FROM devices WHERE id = $1 AND org_id = $2`, id, orgID).Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.SerialNumber, &d.Type, &d.Model, &d.Manufacturer, &d.FirmwareVersion, &d.HardwareRevision, &d.BatteryLevel, &d.SignalStrength, &d.Status, &d.LastPingAt, &d.LastHeartbeat, &d.IsDeleted, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return d, err
}

func (r *postgresRepository) UpdateDeviceHealth(ctx context.Context, deviceID uuid.UUID, battery, signal *int, firmware *string) error {
	_, err := r.db.Exec(ctx, `UPDATE devices SET battery_level = COALESCE($2, battery_level), signal_strength = COALESCE($3, signal_strength), firmware_version = COALESCE($4, firmware_version), last_heartbeat = NOW(), updated_at = NOW() WHERE id = $1`, deviceID, battery, signal, firmware)
	if err != nil {
		return fmt.Errorf("failed to update device health: %w", err)
	}
	return nil
}

// Telemetry implementations

func (r *postgresRepository) ListTelemetry(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceTelemetry, error) {
	rows, err := r.db.Query(ctx, `SELECT id, device_id, org_id, payload, packet_type, raw_hex, timestamp FROM device_telemetry WHERE device_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`, deviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry: %w", err)
	}
	defer rows.Close()

	var results []*models.DeviceTelemetry
	for rows.Next() {
		t := &models.DeviceTelemetry{}
		var payloadJSON []byte
		if err := rows.Scan(&t.ID, &t.DeviceID, &t.OrgID, &payloadJSON, &t.PacketType, &t.RawHex, &t.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan telemetry: %w", err)
		}
		if err := json.Unmarshal(payloadJSON, &t.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		results = append(results, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("telemetry iteration error: %w", err)
	}

	return results, nil
}

func (r *postgresRepository) CreateTelemetry(ctx context.Context, t *models.DeviceTelemetry) error {
	payloadJSON, err := json.Marshal(t.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	_, err = r.db.Exec(ctx, `INSERT INTO device_telemetry (id, device_id, org_id, payload, packet_type, raw_hex) VALUES ($1, $2, $3, $4, $5, $6)`, t.ID, t.DeviceID, t.OrgID, payloadJSON, t.PacketType, t.RawHex)
	if err != nil {
		return fmt.Errorf("failed to insert telemetry: %w", err)
	}
	return err
}

func (r *postgresRepository) GetLatestTelemetry(ctx context.Context, deviceID uuid.UUID) (*models.DeviceTelemetry, error) {
	var t models.DeviceTelemetry
	var payloadJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, device_id, org_id, payload, packet_type, raw_hex, timestamp
		FROM device_telemetry
		WHERE device_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`, deviceID).Scan(
		&t.ID, &t.DeviceID, &t.OrgID, &payloadJSON, &t.PacketType, &t.RawHex, &t.Timestamp,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest telemetry: %w", err)
	}
	if err := json.Unmarshal(payloadJSON, &t.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return &t, nil
}

// Command implementations

func (r *postgresRepository) ListCommands(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error) {
	rows, err := r.db.Query(ctx, `SELECT id, device_id, org_id, command, category, params, status, result, error_message, created_by, sent_at, executed_at, created_at FROM device_commands WHERE device_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, deviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query commands: %w", err)
	}
	defer rows.Close()

	var results []*models.DeviceCommand
	for rows.Next() {
		c := &models.DeviceCommand{}
		var paramsJSON, resultJSON []byte
		if err := rows.Scan(&c.ID, &c.DeviceID, &c.OrgID, &c.Command, &c.Category, &paramsJSON, &c.Status, &resultJSON, &c.ErrorMessage, &c.CreatedBy, &c.SentAt, &c.ExecutedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		if err := json.Unmarshal(paramsJSON, &c.Params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}
		if err := json.Unmarshal(resultJSON, &c.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		results = append(results, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("command iteration error: %w", err)
	}

	return results, nil
}

func (r *postgresRepository) GetCommand(ctx context.Context, commandID uuid.UUID) (*models.DeviceCommand, error) {
	c := &models.DeviceCommand{}
	var paramsJSON, resultJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, device_id, org_id, command, category, params, status, result, error_message, created_by, sent_at, executed_at, created_at
		FROM device_commands
		WHERE id = $1
	`, commandID).Scan(&c.ID, &c.DeviceID, &c.OrgID, &c.Command, &c.Category, &paramsJSON, &c.Status, &resultJSON, &c.ErrorMessage, &c.CreatedBy, &c.SentAt, &c.ExecutedAt, &c.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get command: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &c.Params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}
	if err := json.Unmarshal(resultJSON, &c.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}
	return c, nil
}

func (r *postgresRepository) CreateCommand(ctx context.Context, cmd *models.DeviceCommand) error {
	paramsJSON, err := json.Marshal(cmd.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	err = r.db.QueryRow(ctx, `INSERT INTO device_commands (id, device_id, org_id, command, category, params, status, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`, cmd.ID, cmd.DeviceID, cmd.OrgID, cmd.Command, cmd.Category, paramsJSON, cmd.Status, cmd.CreatedBy).Scan(&cmd.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateCommandStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}, errorMsg *string) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	_, err = r.db.Exec(ctx, `
		UPDATE device_commands
		SET status = $2,
		    result = $3,
		    error_message = $4,
		    sent_at = CASE
		        WHEN $2 IN ('delivered', 'acked', 'applied', 'failed', 'timed_out') THEN COALESCE(sent_at, NOW())
		        ELSE sent_at
		    END,
		    executed_at = CASE
		        WHEN $2 IN ('applied', 'failed', 'timed_out') THEN NOW()
		        ELSE executed_at
		    END
		WHERE id = $1
	`, id, status, resultJSON, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update command status: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListDeviceTwins(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DeviceTwin, error) {
	rows, err := r.db.Query(ctx, `
		SELECT device_id, org_id, desired_state, reported_state, last_seen, drift_status, drift_reason, last_command_id, last_command_status, created_at, updated_at
		FROM device_twins
		WHERE org_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query device twins: %w", err)
	}
	defer rows.Close()

	var twins []*models.DeviceTwin
	for rows.Next() {
		twin, err := scanDeviceTwin(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device twin: %w", err)
		}
		twins = append(twins, twin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("device twin iteration error: %w", err)
	}

	return twins, nil
}

func (r *postgresRepository) GetDeviceTwin(ctx context.Context, deviceID, orgID uuid.UUID) (*models.DeviceTwin, error) {
	row := r.db.QueryRow(ctx, `
		SELECT device_id, org_id, desired_state, reported_state, last_seen, drift_status, drift_reason, last_command_id, last_command_status, created_at, updated_at
		FROM device_twins
		WHERE device_id = $1 AND org_id = $2
	`, deviceID, orgID)
	twin, err := scanDeviceTwinRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get device twin: %w", err)
	}
	return twin, nil
}

func (r *postgresRepository) UpsertDesiredState(ctx context.Context, deviceID, orgID uuid.UUID, desiredState map[string]interface{}, lastCommandID *uuid.UUID, lastCommandStatus *string) error {
	desiredJSON, err := json.Marshal(desiredState)
	if err != nil {
		return fmt.Errorf("failed to marshal desired state: %w", err)
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO device_twins (device_id, org_id, desired_state, drift_status, drift_reason, last_command_id, last_command_status, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, 'unknown', NULL, $4, $5, NOW(), NOW())
		ON CONFLICT (device_id) DO UPDATE SET
			org_id = EXCLUDED.org_id,
			desired_state = EXCLUDED.desired_state,
			drift_status = 'unknown',
			drift_reason = NULL,
			last_command_id = EXCLUDED.last_command_id,
			last_command_status = EXCLUDED.last_command_status,
			updated_at = NOW()
	`, deviceID, orgID, string(desiredJSON), lastCommandID, lastCommandStatus)
	if err != nil {
		return fmt.Errorf("failed to upsert desired state: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpsertReportedState(ctx context.Context, deviceID, orgID uuid.UUID, reportedState map[string]interface{}, lastSeen *time.Time) error {
	reportedJSON, err := json.Marshal(reportedState)
	if err != nil {
		return fmt.Errorf("failed to marshal reported state: %w", err)
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO device_twins (device_id, org_id, reported_state, last_seen, drift_status, drift_reason, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, COALESCE($4, NOW()), 'unknown', NULL, NOW(), NOW())
		ON CONFLICT (device_id) DO UPDATE SET
			org_id = EXCLUDED.org_id,
			reported_state = EXCLUDED.reported_state,
			last_seen = COALESCE(EXCLUDED.last_seen, device_twins.last_seen),
			drift_status = 'unknown',
			drift_reason = NULL,
			updated_at = NOW()
	`, deviceID, orgID, string(reportedJSON), lastSeen)
	if err != nil {
		return fmt.Errorf("failed to upsert reported state: %w", err)
	}
	return nil
}

// Firmware implementations

func (r *postgresRepository) ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error) {
	rows, err := r.db.Query(ctx, `SELECT id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical, released_at, created_at, updated_at FROM firmware_packages ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query firmware: %w", err)
	}
	defer rows.Close()

	var results []*models.FirmwarePackage
	for rows.Next() {
		f := &models.FirmwarePackage{}
		if err := rows.Scan(&f.ID, &f.Version, &f.HardwareTarget, &f.FileURL, &f.FileSize, &f.Checksum, &f.ReleaseNotes, &f.Status, &f.IsCritical, &f.ReleasedAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan firmware: %w", err)
		}
		results = append(results, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("firmware iteration error: %w", err)
	}

	return results, nil
}

func (r *postgresRepository) CreateFirmware(ctx context.Context, fw *models.FirmwarePackage) error {
	err := r.db.QueryRow(ctx, `INSERT INTO firmware_packages (id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING created_at, updated_at`, fw.ID, fw.Version, fw.HardwareTarget, fw.FileURL, fw.FileSize, fw.Checksum, fw.ReleaseNotes, fw.Status, fw.IsCritical).Scan(&fw.CreatedAt, &fw.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create firmware: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetFirmware(ctx context.Context, id uuid.UUID) (*models.FirmwarePackage, error) {
	f := &models.FirmwarePackage{}
	err := r.db.QueryRow(ctx, `SELECT id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical, released_at, created_at, updated_at FROM firmware_packages WHERE id = $1`, id).Scan(&f.ID, &f.Version, &f.HardwareTarget, &f.FileURL, &f.FileSize, &f.Checksum, &f.ReleaseNotes, &f.Status, &f.IsCritical, &f.ReleasedAt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get firmware: %w", err)
	}
	return f, err
}

// Rollout implementations

func (r *postgresRepository) ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error) {
	rows, err := r.db.Query(ctx, `SELECT id, firmware_id, org_id, name, device_group, target_filter, total_devices, completed_devices, failed_devices, status, rollout_strategy, max_concurrent, created_by, started_at, completed_at, created_at FROM firmware_rollouts WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query rollouts: %w", err)
	}
	defer rows.Close()

	var results []*models.FirmwareRollout
	for rows.Next() {
		ro := &models.FirmwareRollout{}
		var filterJSON []byte
		if err := rows.Scan(&ro.ID, &ro.FirmwareID, &ro.OrgID, &ro.Name, &ro.DeviceGroup, &filterJSON, &ro.TotalDevices, &ro.CompletedDevices, &ro.FailedDevices, &ro.Status, &ro.RolloutStrategy, &ro.MaxConcurrent, &ro.CreatedBy, &ro.StartedAt, &ro.CompletedAt, &ro.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan rollout: %w", err)
		}
		if err := json.Unmarshal(filterJSON, &ro.TargetFilter); err != nil {
			return nil, fmt.Errorf("failed to unmarshal target filter: %w", err)
		}
		results = append(results, ro)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rollout iteration error: %w", err)
	}

	return results, nil
}

func (r *postgresRepository) CreateRollout(ctx context.Context, ro *models.FirmwareRollout) error {
	filterJSON, err := json.Marshal(ro.TargetFilter)
	if err != nil {
		return fmt.Errorf("failed to marshal target filter: %w", err)
	}
	err = r.db.QueryRow(ctx, `INSERT INTO firmware_rollouts (id, firmware_id, org_id, name, device_group, target_filter, total_devices, status, rollout_strategy, max_concurrent, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING created_at`, ro.ID, ro.FirmwareID, ro.OrgID, ro.Name, ro.DeviceGroup, filterJSON, ro.TotalDevices, ro.Status, ro.RolloutStrategy, ro.MaxConcurrent, ro.CreatedBy).Scan(&ro.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create rollout: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetRollout(ctx context.Context, rolloutID uuid.UUID) (*models.FirmwareRollout, error) {
	ro := &models.FirmwareRollout{}
	var filterJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, firmware_id, org_id, name, device_group, target_filter, total_devices, completed_devices, failed_devices, status, rollout_strategy, max_concurrent, created_by, started_at, completed_at, created_at
		FROM firmware_rollouts
		WHERE id = $1
	`, rolloutID).Scan(&ro.ID, &ro.FirmwareID, &ro.OrgID, &ro.Name, &ro.DeviceGroup, &filterJSON, &ro.TotalDevices, &ro.CompletedDevices, &ro.FailedDevices, &ro.Status, &ro.RolloutStrategy, &ro.MaxConcurrent, &ro.CreatedBy, &ro.StartedAt, &ro.CompletedAt, &ro.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get rollout: %w", err)
	}
	if err := json.Unmarshal(filterJSON, &ro.TargetFilter); err != nil {
		return nil, fmt.Errorf("failed to unmarshal target filter: %w", err)
	}
	return ro, nil
}

func (r *postgresRepository) UpdateRolloutProgress(ctx context.Context, id uuid.UUID, completed, failed int, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE firmware_rollouts SET completed_devices = $2, failed_devices = $3, status = $4 WHERE id = $1`, id, completed, failed, status)
	if err != nil {
		return fmt.Errorf("failed to update rollout progress: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateRolloutStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE firmware_rollouts
		SET status = $2,
		    started_at = CASE
		        WHEN $3 = 'in_progress' THEN COALESCE(started_at, NOW())
		        ELSE started_at
		    END,
		    completed_at = CASE
		        WHEN $3 IN ('completed', 'cancelled') THEN NOW()
		        ELSE completed_at
		    END
		WHERE id = $1
	`, id, status, status)
	if err != nil {
		return fmt.Errorf("failed to update rollout status: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListRolloutEvents(ctx context.Context, rolloutID, orgID uuid.UUID, eventType string, limit, offset int) ([]*models.FirmwareRolloutEvent, error) {
	query := `
		SELECT id, rollout_id, org_id, event_type, from_status, to_status, actor_user_id, message, metadata, created_at
		FROM firmware_rollout_events
		WHERE rollout_id = $1 AND org_id = $2`
	args := []interface{}{rolloutID, orgID}
	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", len(args)+1)
		args = append(args, eventType)
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rollout events: %w", err)
	}
	defer rows.Close()

	var events []*models.FirmwareRolloutEvent
	for rows.Next() {
		event := &models.FirmwareRolloutEvent{}
		var metadataJSON []byte
		if err := rows.Scan(
			&event.ID,
			&event.RolloutID,
			&event.OrgID,
			&event.EventType,
			&event.FromStatus,
			&event.ToStatus,
			&event.ActorUserID,
			&event.Message,
			&metadataJSON,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan rollout event: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rollout event iteration error: %w", err)
	}

	return events, nil
}

func (r *postgresRepository) CreateRolloutEvent(ctx context.Context, event *models.FirmwareRolloutEvent) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO firmware_rollout_events
			(id, rollout_id, org_id, event_type, from_status, to_status, actor_user_id, message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`,
		event.ID,
		event.RolloutID,
		event.OrgID,
		event.EventType,
		event.FromStatus,
		event.ToStatus,
		event.ActorUserID,
		event.Message,
		metadataJSON,
	).Scan(&event.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create rollout event: %w", err)
	}
	return nil
}

// Diagnostics logs implementations

func (r *postgresRepository) ListDiagnosticsLogs(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DiagnosticsLog, error) {
	rows, err := r.db.Query(ctx, `SELECT id, device_id, org_id, log_type, source, message, metadata, timestamp FROM device_diagnostics_logs WHERE device_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`, deviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query diagnostics logs: %w", err)
	}
	defer rows.Close()

	var results []*models.DiagnosticsLog
	for rows.Next() {
		l := &models.DiagnosticsLog{}
		var metaJSON []byte
		if err := rows.Scan(&l.ID, &l.DeviceID, &l.OrgID, &l.LogType, &l.Source, &l.Message, &metaJSON, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan diagnostics log: %w", err)
		}
		if err := json.Unmarshal(metaJSON, &l.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		results = append(results, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("diagnostics log iteration error: %w", err)
	}

	return results, nil
}

func (r *postgresRepository) CreateDiagnosticsLog(ctx context.Context, log *models.DiagnosticsLog) error {
	metaJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	_, err = r.db.Exec(ctx, `INSERT INTO device_diagnostics_logs (id, device_id, org_id, log_type, source, message, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7)`, log.ID, log.DeviceID, log.OrgID, log.LogType, log.Source, log.Message, metaJSON)
	if err != nil {
		return fmt.Errorf("failed to create diagnostics log: %w", err)
	}
	return nil
}

func scanDeviceTwin(rows pgx.Rows) (*models.DeviceTwin, error) {
	twin := &models.DeviceTwin{}
	var desiredJSON, reportedJSON []byte
	if err := rows.Scan(
		&twin.DeviceID, &twin.OrgID, &desiredJSON, &reportedJSON, &twin.LastSeen, &twin.DriftStatus,
		&twin.DriftReason, &twin.LastCommandID, &twin.LastCommandStatus, &twin.CreatedAt, &twin.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if len(desiredJSON) > 0 {
		_ = json.Unmarshal(desiredJSON, &twin.DesiredState)
	}
	if len(reportedJSON) > 0 {
		_ = json.Unmarshal(reportedJSON, &twin.ReportedState)
	}
	return twin, nil
}

func scanDeviceTwinRow(row pgx.Row) (*models.DeviceTwin, error) {
	twin := &models.DeviceTwin{}
	var desiredJSON, reportedJSON []byte
	if err := row.Scan(
		&twin.DeviceID, &twin.OrgID, &desiredJSON, &reportedJSON, &twin.LastSeen, &twin.DriftStatus,
		&twin.DriftReason, &twin.LastCommandID, &twin.LastCommandStatus, &twin.CreatedAt, &twin.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to scan device twin: %w", err)
	}
	if len(desiredJSON) > 0 {
		if err := json.Unmarshal(desiredJSON, &twin.DesiredState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal desired state: %w", err)
		}
	}
	if len(reportedJSON) > 0 {
		if err := json.Unmarshal(reportedJSON, &twin.ReportedState); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reported state: %w", err)
		}
	}
	return twin, nil
}

// Audit log implementations

func (r *postgresRepository) CreateAuditLog(ctx context.Context, audit *models.DiagnosticsAuditLog) error {
	var metadataJSON []byte
	var err error
	if audit.Metadata != nil {
		metadataJSON, err = json.Marshal(audit.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	var oldValueJSON, newValueJSON []byte
	if audit.OldValue != nil {
		oldValueJSON, err = json.Marshal(audit.OldValue)
		if err != nil {
			return fmt.Errorf("failed to marshal old value: %w", err)
		}
	}
	if audit.NewValue != nil {
		newValueJSON, err = json.Marshal(audit.NewValue)
		if err != nil {
			return fmt.Errorf("failed to marshal new value: %w", err)
		}
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO diagnostics_audit_log (
			org_id, entity_type, entity_id, action, old_value, new_value,
			user_id, user_type, user_email, ip_address, user_agent, reason, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		audit.OrgID, audit.EntityType, audit.EntityID, audit.Action,
		oldValueJSON, newValueJSON, audit.UserID, audit.UserType, audit.UserEmail,
		audit.IPAddress, audit.UserAgent, audit.Reason, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]*models.DiagnosticsAuditLog, int, error) {
	baseQuery := `
		SELECT id, org_id, entity_type, entity_id, action, old_value, new_value,
		       user_id, user_type, user_email, ip_address, user_agent, reason, metadata, created_at
		FROM diagnostics_audit_log
		WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM diagnostics_audit_log WHERE 1=1`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.OrgID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", argIndex))
		args = append(args, filter.OrgID)
		argIndex++
	}

	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argIndex))
		args = append(args, filter.EntityType)
		argIndex++
	}

	if filter.EntityID != nil {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIndex))
		args = append(args, filter.EntityID)
		argIndex++
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, filter.Action)
		argIndex++
	}

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.UserType != "" {
		conditions = append(conditions, fmt.Sprintf("user_type = $%d", argIndex))
		args = append(args, filter.UserType)
		argIndex++
	}

	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.ToDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Count
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery + whereClause, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// List
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 50
	}
	offset := (filter.Page - 1) * filter.PageSize

	limitClause := fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, baseQuery + whereClause + limitClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var audits []*models.DiagnosticsAuditLog
	for rows.Next() {
		audit := &models.DiagnosticsAuditLog{}
		var oldValueJSON, newValueJSON, metadataJSON []byte
		if err := rows.Scan(
			&audit.ID, &audit.OrgID, &audit.EntityType, &audit.EntityID, &audit.Action,
			&oldValueJSON, &newValueJSON, &audit.UserID, &audit.UserType, &audit.UserEmail,
			&audit.IPAddress, &audit.UserAgent, &audit.Reason, &metadataJSON, &audit.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if len(oldValueJSON) > 0 {
			if err := json.Unmarshal(oldValueJSON, &audit.OldValue); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal old value: %w", err)
			}
		}
		if len(newValueJSON) > 0 {
			if err := json.Unmarshal(newValueJSON, &audit.NewValue); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal new value: %w", err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &audit.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		audits = append(audits, audit)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit log iteration error: %w", err)
	}

	return audits, totalCount, nil
}
