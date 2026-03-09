package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// TelemetryRepository defines the storage contract for telemetry events.
type TelemetryRepository interface {
	IngestEvent(ctx context.Context, tx pgx.Tx, event *models.TelemetryEvent) error
	IsDuplicate(ctx context.Context, eventID, deviceID string) (bool, error)
	GetHistory(ctx context.Context, tenantID uuid.UUID, q models.HistoryQuery) ([]models.TelemetryEvent, error)
	GetNearby(ctx context.Context, tenantID uuid.UUID, q models.NearbyQuery) ([]models.LiveLocation, error)
	GetLatestByVehicle(ctx context.Context, tenantID uuid.UUID, vehicleID uuid.UUID) (*models.TelemetryEvent, error)
	GetTrackingConfig(ctx context.Context, targetID uuid.UUID) (map[string]any, error)
}

// PostgresTelemetryRepository implements TelemetryRepository using pgxpool + PostGIS.
type PostgresTelemetryRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewTelemetryRepository(db *pgxpool.Pool, logger *zap.Logger) *PostgresTelemetryRepository {
	return &PostgresTelemetryRepository{db: db, logger: logger}
}

// IngestEvent inserts a single telemetry event into the partitioned table.
// Supports being part of a transaction (for dual-write outbox).
func (r *PostgresTelemetryRepository) IngestEvent(ctx context.Context, tx pgx.Tx, event *models.TelemetryEvent) error {
	const q = `
		INSERT INTO telemetry_points (
			event_id, vehicle_id, device_id, tenant_id,
			lat, lng, speed, heading, ignition,
			fuel_level, odometer, event_type, sequence_no,
			recorded_at, received_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		) ON CONFLICT DO NOTHING`

	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, q,
			event.EventID, event.VehicleID, event.DeviceID, event.TenantID,
			event.Lat, event.Lng, event.Speed, event.Heading, event.Ignition,
			event.FuelLevel, event.Odometer, event.EventType, event.SequenceNo,
			event.RecordedAt, event.ReceivedAt,
		)
	} else {
		_, err = r.db.Exec(ctx, q,
			event.EventID, event.VehicleID, event.DeviceID, event.TenantID,
			event.Lat, event.Lng, event.Speed, event.Heading, event.Ignition,
			event.FuelLevel, event.Odometer, event.EventType, event.SequenceNo,
			event.RecordedAt, event.ReceivedAt,
		)
	}
	if err != nil {
		r.logger.Error("failed to ingest telemetry event",
			zap.String("event_id", event.EventID),
			zap.Error(err),
		)
	}
	return err
}

func (r *PostgresTelemetryRepository) IsDuplicate(ctx context.Context, eventID, deviceID string) (bool, error) {
	const q = `SELECT 1 FROM telemetry_points WHERE event_id = $1 AND device_id = $2 LIMIT 1`
	var dummy int
	err := r.db.QueryRow(ctx, q, eventID, deviceID).Scan(&dummy)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// GetHistory returns paginated telemetry history for a vehicle in the given time range.
func (r *PostgresTelemetryRepository) GetHistory(ctx context.Context, tenantID uuid.UUID, q models.HistoryQuery) ([]models.TelemetryEvent, error) {
	if q.PageSize <= 0 {
		q.PageSize = 100
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize

	const query = `
		SELECT event_id, vehicle_id, device_id, tenant_id,
		       lat, lng, speed, heading, ignition,
		       fuel_level, odometer, event_type, sequence_no,
		       recorded_at, received_at
		FROM telemetry_points
		WHERE vehicle_id = $1 AND tenant_id = $2
		  AND recorded_at BETWEEN $3 AND $4
		ORDER BY recorded_at DESC
		LIMIT $5 OFFSET $6`

	rows, err := r.db.Query(ctx, query,
		q.VehicleID, tenantID, q.From, q.To, q.PageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.TelemetryEvent
	for rows.Next() {
		var e models.TelemetryEvent
		err := rows.Scan(
			&e.EventID, &e.VehicleID, &e.DeviceID, &e.TenantID,
			&e.Lat, &e.Lng, &e.Speed, &e.Heading, &e.Ignition,
			&e.FuelLevel, &e.Odometer, &e.EventType, &e.SequenceNo,
			&e.RecordedAt, &e.ReceivedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetNearby returns the latest live location for vehicles within radiusM metres
// of the given lat/lng using PostGIS ST_DWithin. Mandatory tenantID isolation.
// Uses bbox pre-filter (&&) for GIST index optimization.
func (r *PostgresTelemetryRepository) GetNearby(ctx context.Context, tenantID uuid.UUID, q models.NearbyQuery) ([]models.LiveLocation, error) {
	const query = `
		SELECT DISTINCT ON (vehicle_id)
		       vehicle_id, lat, lng, speed, heading, ignition, recorded_at
		FROM telemetry_points
		WHERE received_at > $1 AND tenant_id = $2
		  AND geo && ST_Expand(ST_SetSRID(ST_MakePoint($4, $3), 4326)::geography::geometry, $5)
		  AND ST_DWithin(
		        geo::geography,
		        ST_SetSRID(ST_MakePoint($4, $3), 4326)::geography,
		        $5
		      )
		ORDER BY vehicle_id, recorded_at DESC`

	since := time.Now().Add(-10 * time.Minute)
	rows, err := r.db.Query(ctx, query, since, tenantID, q.Lat, q.Lng, q.RadiusM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []models.LiveLocation
	for rows.Next() {
		var l models.LiveLocation
		if err := rows.Scan(&l.VehicleID, &l.Lat, &l.Lng, &l.Speed, &l.Heading, &l.Ignition, &l.LastSeen); err != nil {
			return nil, err
		}
		locs = append(locs, l)
	}
	return locs, rows.Err()
}

// GetLatestByVehicle returns the single most recent telemetry event for a vehicle.
func (r *PostgresTelemetryRepository) GetLatestByVehicle(ctx context.Context, tenantID uuid.UUID, vehicleID uuid.UUID) (*models.TelemetryEvent, error) {
	const query = `
		SELECT event_id, vehicle_id, device_id, tenant_id,
		       lat, lng, speed, heading, ignition,
		       fuel_level, odometer, event_type, sequence_no,
		       recorded_at, received_at
		FROM telemetry_points
		WHERE vehicle_id = $1 AND tenant_id = $2
		ORDER BY recorded_at DESC
		LIMIT 1`

	var e models.TelemetryEvent
	err := r.db.QueryRow(ctx, query, vehicleID, tenantID).Scan(
		&e.EventID, &e.VehicleID, &e.DeviceID, &e.TenantID,
		&e.Lat, &e.Lng, &e.Speed, &e.Heading, &e.Ignition,
		&e.FuelLevel, &e.Odometer, &e.EventType, &e.SequenceNo,
		&e.RecordedAt, &e.ReceivedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

// GetTrackingConfig fetches custom thresholds for a vehicle or tenant.
func (r *PostgresTelemetryRepository) GetTrackingConfig(ctx context.Context, targetID uuid.UUID) (map[string]any, error) {
	const query = `SELECT config FROM tracking_config WHERE target_id = $1 LIMIT 1`
	var config map[string]any
	err := r.db.QueryRow(ctx, query, targetID).Scan(&config)
	if err == pgx.ErrNoRows {
		return nil, nil // Fallback to defaults
	}
	return config, err
}
