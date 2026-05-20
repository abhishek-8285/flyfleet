package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

// TelemetryRepository handles persistence of GPS telemetry points
// in TimescaleDB.
type TelemetryRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewTelemetryRepository(pool *pgxpool.Pool, logger *zap.Logger) *TelemetryRepository {
	return &TelemetryRepository{pool: pool, logger: logger}
}

// Write inserts a telemetry point into the TimescaleDB hypertable.
func (r *TelemetryRepository) Write(ctx context.Context, event *models.TelemetryEvent) error {
	// Self-healing: ensure table exists
	_, _ = r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS telemetry_points (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			device_id    VARCHAR(100) NOT NULL,
			lat          DOUBLE PRECISION NOT NULL,
			lng          DOUBLE PRECISION NOT NULL,
			speed        DOUBLE PRECISION DEFAULT 0,
			recorded_at  TIMESTAMP WITH TIME ZONE NOT NULL,
			event_id     UUID NOT NULL,
			tenant_id    UUID NOT NULL,
			received_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			vehicle_id   UUID,
			heading      DOUBLE PRECISION,
			ignition     BOOLEAN,
			fuel_level   DOUBLE PRECISION,
			raw_data     JSONB
		)
	`)

	q := `INSERT INTO telemetry_points (recorded_at, device_id, tenant_id, event_id, lat, lng, speed, vehicle_id, received_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`

	_, err := r.pool.Exec(ctx, q,
		event.RecordedAt,
		event.DeviceID,
		event.TenantID,
		event.EventID,
		event.Lat,
		event.Lng,
		event.Speed,
		event.VehicleID,
	)
	if err != nil {
		r.logger.Error("DB insert failed", zap.Error(err), zap.String("query", q))
		return fmt.Errorf("insert telemetry: %w", err)
	}
	return nil
}

// BulkIngest inserts multiple telemetry points efficiently using pgx.CopyFrom.
func (r *TelemetryRepository) BulkIngest(ctx context.Context, events []models.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}

	// CRITICAL TIMESCALEDB OPTIMIZATION: Sort chronologically to prevent chunk/page thrashing
	sort.Slice(events, func(i, j int) bool {
		return events[i].RecordedAt.Before(events[j].RecordedAt)
	})

	rows := make([][]any, len(events))
	for i, e := range events {
		rows[i] = []any{
			e.RecordedAt,
			e.DeviceID,
			e.TenantID,
			e.EventID,
			e.Lat,
			e.Lng,
			e.Speed,
			e.VehicleID,
			time.Now(), // received_at
		}
	}

	_, err := r.pool.CopyFrom(ctx,
		pgx.Identifier{"telemetry_points"},
		[]string{
			"recorded_at", "device_id", "tenant_id", "event_id", "lat", "lng", "speed", "vehicle_id", "received_at",
		},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		r.logger.Error("bulk ingest failed", zap.Int("count", len(events)), zap.Error(err))
		return fmt.Errorf("bulk ingest: %w", err)
	}
	return nil
}

// GetLatestByVehicle returns the most recent telemetry point for a
// single vehicle. Used as DB fallback when Redis cache misses.
func (r *TelemetryRepository) GetLatestByVehicle(ctx context.Context, vehicleID uuid.UUID) (*models.TelemetryEvent, error) {
	q := `SELECT vehicle_id, lat, lng, speed, heading, ignition, fuel_level, recorded_at
	      FROM telemetry_points
	      WHERE vehicle_id = $1
	      ORDER BY recorded_at DESC
	      LIMIT 1`

	var t models.TelemetryEvent
	err := r.pool.QueryRow(ctx, q, vehicleID).Scan(
		&t.VehicleID,
		&t.Lat,
		&t.Lng,
		&t.Speed,
		&t.Heading,
		&t.Ignition,
		&t.FuelLevel,
		&t.RecordedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get latest telemetry: %w", err)
	}
	return &t, nil
}

// GetHistoryBucket returns downsampled telemetry for a vehicle within
// a time range. Uses TimescaleDB time_bucket if available, falls back
// to plain PostgreSQL epoch-based bucketing.
func (r *TelemetryRepository) GetHistoryBucket(ctx context.Context, vehicleID uuid.UUID,
	start, end time.Time, bucketInterval string) ([]models.TelemetryEvent, error) {

	// Try TimescaleDB time_bucket first
	q := fmt.Sprintf(`
		SELECT
			time_bucket('%s', recorded_at) AS bucket_time,
			AVG(lat)  AS lat,
			AVG(lng) AS lng,
			AVG(speed) AS speed,
			AVG(heading)   AS heading,
			BOOL_OR(ignition) AS ignition,
			MAX(speed) AS max_speed
		FROM telemetry_points
		WHERE vehicle_id = $1
		  AND recorded_at >= $2
		  AND recorded_at <= $3
		GROUP BY bucket_time
		ORDER BY bucket_time`, bucketInterval)

	rows, err := r.pool.Query(ctx, q, vehicleID, start, end)
	if err != nil {
		// Fallback: plain PostgreSQL using epoch-based bucketing
		bucketSecs := parseBucketInterval(bucketInterval)
		if bucketSecs <= 0 {
			bucketSecs = 60 // default 60 seconds
		}
		q = `
			SELECT
				to_timestamp(floor(extract(epoch from recorded_at) / $4) * $4) AS bucket_time,
				AVG(lat)  AS lat,
				AVG(lng) AS lng,
				AVG(speed) AS speed,
				AVG(heading)   AS heading,
				BOOL_OR(ignition) AS ignition,
				MAX(speed) AS max_speed
			FROM telemetry_points
			WHERE vehicle_id = $1
			  AND recorded_at >= $2
			  AND recorded_at <= $3
			GROUP BY bucket_time
			ORDER BY bucket_time`
		rows, err = r.pool.Query(ctx, q, vehicleID, start, end, bucketSecs)
		if err != nil {
			return nil, fmt.Errorf("query history bucket: %w", err)
		}
	}
	defer rows.Close()

	var points []models.TelemetryEvent
	for rows.Next() {
		var p models.TelemetryEvent
		var maxSpeed float64
		if err := rows.Scan(
			&p.RecordedAt,
			&p.Lat,
			&p.Lng,
			&p.Speed,
			&p.Heading,
			&p.Ignition,
			&maxSpeed,
		); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		p.VehicleID = vehicleID
		points = append(points, p)
	}
	return points, nil
}

// parseBucketInterval converts a bucket interval string to seconds.
// Examples: "10 seconds" → 10, "5 minutes" → 300, "1 hour" → 3600
func parseBucketInterval(interval string) int64 {
	interval = strings.ToLower(strings.TrimSpace(interval))
	// Extract number and unit
	var secs int64
	var num int64
	var unit string
	fmt.Sscanf(interval, "%d %s", &num, &unit)
	// Handle plural/singular
	unit = strings.TrimSuffix(unit, "s")
	switch unit {
	case "second":
		secs = num
	case "minute":
		secs = num * 60
	case "hour":
		secs = num * 3600
	case "day":
		secs = num * 86400
	default:
		secs = 60
	}
	return secs
}

// GetBatchLatest returns the latest telemetry point for each vehicle
// in the given list. Used for the bulk vehicle locations endpoint.
func (r *TelemetryRepository) GetBatchLatest(ctx context.Context, vehicleIDs []uuid.UUID) (map[uuid.UUID]models.TelemetryEvent, error) {
	if len(vehicleIDs) == 0 {
		return map[uuid.UUID]models.TelemetryEvent{}, nil
	}

	q := `
		SELECT DISTINCT ON (vehicle_id)
			vehicle_id, lat, lng, speed, heading, ignition, fuel_level, recorded_at
		FROM telemetry_points
		WHERE vehicle_id = ANY($1)
		ORDER BY vehicle_id, recorded_at DESC`

	rows, err := r.pool.Query(ctx, q, vehicleIDs)
	if err != nil {
		return nil, fmt.Errorf("batch latest telemetry: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]models.TelemetryEvent, len(vehicleIDs))
	for rows.Next() {
		var t models.TelemetryEvent
		if err := rows.Scan(
			&t.VehicleID,
			&t.Lat,
			&t.Lng,
			&t.Speed,
			&t.Heading,
			&t.Ignition,
			&t.FuelLevel,
			&t.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("scan batch telemetry: %w", err)
		}
		result[t.VehicleID] = t
	}
	return result, nil
}

// ResolveDeviceAssignment resolves tenant_id and vehicle_id assigned to a device_id.
func (r *TelemetryRepository) ResolveDeviceAssignment(ctx context.Context, deviceID string) (uuid.UUID, uuid.UUID, error) {
	// 1. Try devices table
	var tenantID, vehicleID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT tenant_id, vehicle_id FROM devices WHERE device_id = $1`, deviceID).Scan(&tenantID, &vehicleID)
	if err == nil {
		return tenantID, vehicleID, nil
	}

	// 2. Try vehicles table as fallback
	err = r.pool.QueryRow(ctx, `SELECT tenant_id, id FROM vehicles WHERE gps_device_id = $1 AND deleted_at IS NULL`, deviceID).Scan(&tenantID, &vehicleID)
	if err == nil {
		return tenantID, vehicleID, nil
	}

	return uuid.Nil, uuid.Nil, fmt.Errorf("device assignment not found for device_id: %s", deviceID)
}
