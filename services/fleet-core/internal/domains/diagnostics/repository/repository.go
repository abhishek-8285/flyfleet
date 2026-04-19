package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DiagnosticsRepository struct {
	db *pgxpool.Pool
}

func NewDiagnosticsRepository(db *pgxpool.Pool) *DiagnosticsRepository {
	return &DiagnosticsRepository{db: db}
}

// Device Operations
func (r *DiagnosticsRepository) CreateDevice(ctx context.Context, device *models.Device) error {
	query := `INSERT INTO devices (id, tenant_id, vehicle_id, device_imei, device_serial, device_model, firmware_version, hardware_version, status, last_seen_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, device.ID, device.OrgID, device.VehicleID, device.DeviceIMEI, device.DeviceSerial, device.DeviceModel, device.FirmwareVersion, device.HardwareVersion, device.Status, device.LastSeenAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("create device: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) UpdateDevice(ctx context.Context, device *models.Device) error {
	query := `UPDATE devices SET vehicle_id = $1, status = $2, last_seen_at = $3, firmware_version = $4 WHERE id = $5 AND tenant_id = $6`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, device.VehicleID, device.Status, device.LastSeenAt, device.FirmwareVersion, device.ID, device.OrgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("update device: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) GetDevice(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Device, error) {
	query := `SELECT id, tenant_id, vehicle_id, device_imei, device_serial, device_model, firmware_version, hardware_version, status, last_seen_at, registered_at
			  FROM devices WHERE id = $1 AND tenant_id = $2`
	start := time.Now()
	var d models.Device
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.DeviceIMEI, &d.DeviceSerial, &d.DeviceModel, &d.FirmwareVersion, &d.HardwareVersion, &d.Status, &d.LastSeenAt, &d.RegisteredAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	return &d, nil
}

func (r *DiagnosticsRepository) ListDevices(ctx context.Context, orgID uuid.UUID) ([]*models.Device, error) {
	query := `SELECT id, tenant_id, vehicle_id, device_imei, device_serial, device_model, firmware_version, hardware_version, status, last_seen_at, registered_at
			  FROM devices WHERE tenant_id = $1`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.DeviceIMEI, &d.DeviceSerial, &d.DeviceModel, &d.FirmwareVersion, &d.HardwareVersion, &d.Status, &d.LastSeenAt, &d.RegisteredAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, &d)
	}
	return devices, nil
}

// Telemetry Operations
func (r *DiagnosticsRepository) InsertTelemetry(ctx context.Context, tel *models.DeviceTelemetry) error {
	query := `INSERT INTO device_telemetry (device_id, timestamp, gps_lat, gps_lon, speed, battery_voltage, signal_strength, engine_rpm, temperature, payload_json)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	start := time.Now()
	err := r.db.QueryRow(ctx, query, tel.DeviceID, tel.Timestamp, tel.GPSLat, tel.GPSLon, tel.Speed, tel.BatteryVoltage, tel.SignalStrength, tel.EngineRPM, tel.Temperature, tel.PayloadJSON).Scan(&tel.ID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("insert telemetry: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) GetTelemetry(ctx context.Context, deviceID uuid.UUID, limit int) ([]*models.DeviceTelemetry, error) {
	query := `SELECT id, device_id, timestamp, gps_lat, gps_lon, speed, battery_voltage, signal_strength, engine_rpm, temperature, payload_json
			  FROM device_telemetry WHERE device_id = $1 ORDER BY timestamp DESC LIMIT $2`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, deviceID, limit)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("get telemetry: %w", err)
	}
	defer rows.Close()

	var telemetries []*models.DeviceTelemetry
	for rows.Next() {
		var t models.DeviceTelemetry
		if err := rows.Scan(&t.ID, &t.DeviceID, &t.Timestamp, &t.GPSLat, &t.GPSLon, &t.Speed, &t.BatteryVoltage, &t.SignalStrength, &t.EngineRPM, &t.Temperature, &t.PayloadJSON); err != nil {
			return nil, fmt.Errorf("scan telemetry: %w", err)
		}
		telemetries = append(telemetries, &t)
	}
	return telemetries, nil
}

// Command Operations
func (r *DiagnosticsRepository) CreateCommand(ctx context.Context, cmd *models.DeviceCommand) error {
	query := `INSERT INTO device_commands (id, device_id, command_type, command_payload, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, cmd.ID, cmd.DeviceID, cmd.CommandType, cmd.CommandPayload, cmd.Status, cmd.CreatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("create command: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) UpdateCommandStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE device_commands SET status = $1 WHERE id = $2`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, id)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("update command status: %w", err)
	}
	return nil
}

// Firmware Packages
func (r *DiagnosticsRepository) CreateFirmware(ctx context.Context, fw *models.FirmwarePackage) error {
	query := `INSERT INTO firmware_packages (id, firmware_version, device_model, checksum, file_url, size_bytes, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, fw.ID, fw.FirmwareVersion, fw.DeviceModel, fw.Checksum, fw.FileURL, fw.SizeBytes, fw.Status, fw.CreatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("create firmware: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) PublishFirmware(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE firmware_packages SET status = 'PUBLISHED' WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("publish firmware: %w", err)
	}
	return nil
}

// Firmware Rollouts
func (r *DiagnosticsRepository) CreateRollout(ctx context.Context, rollout *models.FirmwareRollout) error {
	query := `INSERT INTO firmware_rollouts (id, firmware_id, tenant_id, rollout_strategy, target_device_count, status, failure_threshold, batch_size, rollback_enabled, scheduled_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, rollout.ID, rollout.FirmwareID, rollout.OrgID, rollout.RolloutStrategy, rollout.TargetDeviceCount, rollout.Status, rollout.FailureThreshold, rollout.BatchSize, rollout.RollbackEnabled, rollout.ScheduledAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("create rollout: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) UpdateRolloutStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE firmware_rollouts SET status = $1 WHERE id = $2`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, id)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("update rollout status: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) GetRolloutStats(ctx context.Context, id uuid.UUID) (map[string]int, error) {
	query := `SELECT status, count(*) FROM rollout_device_status WHERE rollout_id = $1 GROUP BY status`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("get rollout stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, nil
}

func (r *DiagnosticsRepository) InsertRolloutEvent(ctx context.Context, ev *models.FirmwareRolloutEvent) error {
	query := `INSERT INTO firmware_rollout_events (rollout_id, device_id, event_type, metadata, created_at)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	start := time.Now()
	err := r.db.QueryRow(ctx, query, ev.RolloutID, ev.DeviceID, ev.EventType, ev.Metadata, ev.CreatedAt).Scan(&ev.ID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("insert rollout event: %w", err)
	}
	return nil
}

// Logs and Audit
func (r *DiagnosticsRepository) InsertDiagnosticsLog(ctx context.Context, log *models.DeviceDiagnosticsLog) error {
	query := `INSERT INTO device_diagnostics_logs (device_id, log_level, log_message, created_at)
			  VALUES ($1, $2, $3, $4) RETURNING id`
	start := time.Now()
	err := r.db.QueryRow(ctx, query, log.DeviceID, log.LogLevel, log.LogMessage, log.CreatedAt).Scan(&log.ID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("insert diagnostics log: %w", err)
	}
	return nil
}

func (r *DiagnosticsRepository) InsertAuditLog(ctx context.Context, log *models.DiagnosticsAuditLog) error {
	query := `INSERT INTO diagnostics_audit_logs (entity_type, entity_id, action, actor_id, metadata, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	start := time.Now()
	err := r.db.QueryRow(ctx, query, log.EntityType, log.EntityID, log.Action, log.ActorID, log.Metadata, log.CreatedAt).Scan(&log.ID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
