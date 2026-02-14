package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
	CreateCommand(ctx context.Context, cmd *models.DeviceCommand) error
	UpdateCommandStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}, errorMsg *string) error

	// Firmware
	ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error)
	CreateFirmware(ctx context.Context, fw *models.FirmwarePackage) error
	GetFirmware(ctx context.Context, id uuid.UUID) (*models.FirmwarePackage, error)

	// Rollouts
	ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error)
	CreateRollout(ctx context.Context, r *models.FirmwareRollout) error
	UpdateRolloutProgress(ctx context.Context, id uuid.UUID, completed, failed int, status string) error

	// Logs
	ListDiagnosticsLogs(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DiagnosticsLog, error)
	CreateDiagnosticsLog(ctx context.Context, log *models.DiagnosticsLog) error
}

type postgresRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &postgresRepository{db: db, logger: logger}
}

// Device implementations

func (r *postgresRepository) GetDeviceSummary(ctx context.Context, orgID uuid.UUID) (*models.DeviceSummary, error) {
	summary := &models.DeviceSummary{}

	// Total devices
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false`, orgID).Scan(&summary.TotalDevices)
	if err != nil {
		return nil, err
	}

	// Online (heartbeat within 5 minutes)
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND last_heartbeat > NOW() - INTERVAL '5 minutes'`, orgID).Scan(&summary.OnlineDevices)
	if err != nil {
		return nil, err
	}

	summary.OfflineDevices = summary.TotalDevices - summary.OnlineDevices

	// Low battery (< 20%)
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND battery_level IS NOT NULL AND battery_level < 20`, orgID).Scan(&summary.LowBattery)
	if err != nil {
		return nil, err
	}

	// Weak signal (< -80 dBm)
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND signal_strength IS NOT NULL AND signal_strength < -80`, orgID).Scan(&summary.WeakSignal)
	if err != nil {
		return nil, err
	}

	// System errors (status = 'error')
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = false AND status = 'error'`, orgID).Scan(&summary.SystemErrors)
	if err != nil {
		return nil, err
	}

	// New devices this week
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND created_at > NOW() - INTERVAL '7 days'`, orgID).Scan(&summary.NewDevicesWeek)
	if err != nil {
		return nil, err
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

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+len(args)+1)) + ` OFFSET $` + string(rune('0'+len(args)+2))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d := &models.Device{}
		if err := rows.Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.SerialNumber, &d.Type, &d.Model, &d.Manufacturer, &d.FirmwareVersion, &d.HardwareRevision, &d.BatteryLevel, &d.SignalStrength, &d.Status, &d.LastPingAt, &d.LastHeartbeat, &d.IsDeleted, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, nil
}

func (r *postgresRepository) GetDevice(ctx context.Context, id, orgID uuid.UUID) (*models.Device, error) {
	d := &models.Device{}
	err := r.db.QueryRowContext(ctx, `SELECT id, org_id, vehicle_id, serial_number, type, model, manufacturer, firmware_version, hardware_revision, battery_level, signal_strength, status, last_ping_at, last_heartbeat, is_deleted, created_at, updated_at 
	          FROM devices WHERE id = $1 AND org_id = $2`, id, orgID).Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.SerialNumber, &d.Type, &d.Model, &d.Manufacturer, &d.FirmwareVersion, &d.HardwareRevision, &d.BatteryLevel, &d.SignalStrength, &d.Status, &d.LastPingAt, &d.LastHeartbeat, &d.IsDeleted, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *postgresRepository) UpdateDeviceHealth(ctx context.Context, deviceID uuid.UUID, battery, signal *int, firmware *string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET battery_level = COALESCE($2, battery_level), signal_strength = COALESCE($3, signal_strength), firmware_version = COALESCE($4, firmware_version), last_heartbeat = NOW(), updated_at = NOW() WHERE id = $1`, deviceID, battery, signal, firmware)
	return err
}

// Telemetry implementations

func (r *postgresRepository) ListTelemetry(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceTelemetry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, device_id, org_id, payload, packet_type, raw_hex, timestamp FROM device_telemetry WHERE device_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`, deviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.DeviceTelemetry
	for rows.Next() {
		t := &models.DeviceTelemetry{}
		var payloadJSON []byte
		if err := rows.Scan(&t.ID, &t.DeviceID, &t.OrgID, &payloadJSON, &t.PacketType, &t.RawHex, &t.Timestamp); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(payloadJSON, &t.Payload)
		results = append(results, t)
	}
	return results, nil
}

func (r *postgresRepository) CreateTelemetry(ctx context.Context, t *models.DeviceTelemetry) error {
	payloadJSON, _ := json.Marshal(t.Payload)
	_, err := r.db.ExecContext(ctx, `INSERT INTO device_telemetry (id, device_id, org_id, payload, packet_type, raw_hex) VALUES ($1, $2, $3, $4, $5, $6)`, t.ID, t.DeviceID, t.OrgID, payloadJSON, t.PacketType, t.RawHex)
	return err
}

func (r *postgresRepository) GetLatestTelemetry(ctx context.Context, deviceID uuid.UUID) (*models.DeviceTelemetry, error) {
	var t models.DeviceTelemetry
	var payloadJSON []byte
	err := r.db.QueryRowContext(ctx, 
		`SELECT id, device_id, org_id, payload, packet_type, raw_hex, timestamp 
		FROM device_telemetry 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`, deviceID).Scan(
		&t.ID, &t.DeviceID, &t.OrgID, &payloadJSON, &t.PacketType, &t.RawHex, &t.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No telemetry found, return nil without error
		}
		return nil, err
	}
	_ = json.Unmarshal(payloadJSON, &t.Payload)
	return &t, nil
}

// Command implementations

func (r *postgresRepository) ListCommands(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, device_id, org_id, command, category, params, status, result, error_message, created_by, sent_at, executed_at, created_at FROM device_commands WHERE device_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, deviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.DeviceCommand
	for rows.Next() {
		c := &models.DeviceCommand{}
		var paramsJSON, resultJSON []byte
		if err := rows.Scan(&c.ID, &c.DeviceID, &c.OrgID, &c.Command, &c.Category, &paramsJSON, &c.Status, &resultJSON, &c.ErrorMessage, &c.CreatedBy, &c.SentAt, &c.ExecutedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(paramsJSON, &c.Params)
		_ = json.Unmarshal(resultJSON, &c.Result)
		results = append(results, c)
	}
	return results, nil
}

func (r *postgresRepository) CreateCommand(ctx context.Context, cmd *models.DeviceCommand) error {
	paramsJSON, _ := json.Marshal(cmd.Params)
	return r.db.QueryRowContext(ctx, `INSERT INTO device_commands (id, device_id, org_id, command, category, params, status, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`, cmd.ID, cmd.DeviceID, cmd.OrgID, cmd.Command, cmd.Category, paramsJSON, cmd.Status, cmd.CreatedBy).Scan(&cmd.CreatedAt)
}

func (r *postgresRepository) UpdateCommandStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}, errorMsg *string) error {
	resultJSON, _ := json.Marshal(result)
	executedAt := time.Now()
	_, err := r.db.ExecContext(ctx, `UPDATE device_commands SET status = $2, result = $3, error_message = $4, executed_at = $5 WHERE id = $1`, id, status, resultJSON, errorMsg, executedAt)
	return err
}

// Firmware implementations

func (r *postgresRepository) ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical, released_at, created_at, updated_at FROM firmware_packages ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.FirmwarePackage
	for rows.Next() {
		f := &models.FirmwarePackage{}
		if err := rows.Scan(&f.ID, &f.Version, &f.HardwareTarget, &f.FileURL, &f.FileSize, &f.Checksum, &f.ReleaseNotes, &f.Status, &f.IsCritical, &f.ReleasedAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, f)
	}
	return results, nil
}

func (r *postgresRepository) CreateFirmware(ctx context.Context, fw *models.FirmwarePackage) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO firmware_packages (id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING created_at, updated_at`, fw.ID, fw.Version, fw.HardwareTarget, fw.FileURL, fw.FileSize, fw.Checksum, fw.ReleaseNotes, fw.Status, fw.IsCritical).Scan(&fw.CreatedAt, &fw.UpdatedAt)
}

func (r *postgresRepository) GetFirmware(ctx context.Context, id uuid.UUID) (*models.FirmwarePackage, error) {
	f := &models.FirmwarePackage{}
	err := r.db.QueryRowContext(ctx, `SELECT id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical, released_at, created_at, updated_at FROM firmware_packages WHERE id = $1`, id).Scan(&f.ID, &f.Version, &f.HardwareTarget, &f.FileURL, &f.FileSize, &f.Checksum, &f.ReleaseNotes, &f.Status, &f.IsCritical, &f.ReleasedAt, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

// Rollout implementations

func (r *postgresRepository) ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, firmware_id, org_id, name, device_group, target_filter, total_devices, completed_devices, failed_devices, status, rollout_strategy, max_concurrent, created_by, started_at, completed_at, created_at FROM firmware_rollouts WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.FirmwareRollout
	for rows.Next() {
		ro := &models.FirmwareRollout{}
		var filterJSON []byte
		if err := rows.Scan(&ro.ID, &ro.FirmwareID, &ro.OrgID, &ro.Name, &ro.DeviceGroup, &filterJSON, &ro.TotalDevices, &ro.CompletedDevices, &ro.FailedDevices, &ro.Status, &ro.RolloutStrategy, &ro.MaxConcurrent, &ro.CreatedBy, &ro.StartedAt, &ro.CompletedAt, &ro.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(filterJSON, &ro.TargetFilter)
		results = append(results, ro)
	}
	return results, nil
}

func (r *postgresRepository) CreateRollout(ctx context.Context, ro *models.FirmwareRollout) error {
	filterJSON, _ := json.Marshal(ro.TargetFilter)
	return r.db.QueryRowContext(ctx, `INSERT INTO firmware_rollouts (id, firmware_id, org_id, name, device_group, target_filter, total_devices, status, rollout_strategy, max_concurrent, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING created_at`, ro.ID, ro.FirmwareID, ro.OrgID, ro.Name, ro.DeviceGroup, filterJSON, ro.TotalDevices, ro.Status, ro.RolloutStrategy, ro.MaxConcurrent, ro.CreatedBy).Scan(&ro.CreatedAt)
}

func (r *postgresRepository) UpdateRolloutProgress(ctx context.Context, id uuid.UUID, completed, failed int, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE firmware_rollouts SET completed_devices = $2, failed_devices = $3, status = $4 WHERE id = $1`, id, completed, failed, status)
	return err
}

// Diagnostics logs implementations

func (r *postgresRepository) ListDiagnosticsLogs(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DiagnosticsLog, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, device_id, org_id, log_type, source, message, metadata, timestamp FROM device_diagnostics_logs WHERE device_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`, deviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.DiagnosticsLog
	for rows.Next() {
		l := &models.DiagnosticsLog{}
		var metaJSON []byte
		if err := rows.Scan(&l.ID, &l.DeviceID, &l.OrgID, &l.LogType, &l.Source, &l.Message, &metaJSON, &l.Timestamp); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metaJSON, &l.Metadata)
		results = append(results, l)
	}
	return results, nil
}

func (r *postgresRepository) CreateDiagnosticsLog(ctx context.Context, log *models.DiagnosticsLog) error {
	metaJSON, _ := json.Marshal(log.Metadata)
	_, err := r.db.ExecContext(ctx, `INSERT INTO device_diagnostics_logs (id, device_id, org_id, log_type, source, message, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7)`, log.ID, log.DeviceID, log.OrgID, log.LogType, log.Source, log.Message, metaJSON)
	return err
}

// Suppress unused import warning
var _ = pq.Array
