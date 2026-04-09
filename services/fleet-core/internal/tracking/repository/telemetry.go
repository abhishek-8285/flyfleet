package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
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
	q := `INSERT INTO telemetry_points (time, vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, raw_data, created_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`

	_, err := r.pool.Exec(ctx, q,
		event.RecordedAt,
		event.VehicleID,
		event.Lat,
		event.Lng,
		event.Speed,
		event.Heading,
		event.Ignition,
		event.FuelLevel,
		event.RawData,
	)
	if err != nil {
		return fmt.Errorf("insert telemetry: %w", err)
	}
	return nil
}

// GetLatestByVehicle returns the most recent telemetry point for a
// single vehicle. Used as DB fallback when Redis cache misses.
func (r *TelemetryRepository) GetLatestByVehicle(ctx context.Context, vehicleID uuid.UUID) (*models.TelemetryEvent, error) {
	q := `SELECT vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time
	      FROM telemetry_points
	      WHERE vehicle_id = $1
	      ORDER BY time DESC
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
			time_bucket('%s', time) AS bucket_time,
			AVG(latitude)  AS latitude,
			AVG(longitude) AS longitude,
			AVG(speed_kmh) AS speed_kmh,
			AVG(heading)   AS heading,
			BOOL_OR(ignition) AS ignition,
			MAX(speed_kmh) AS max_speed
		FROM telemetry_points
		WHERE vehicle_id = $1
		  AND time >= $2
		  AND time <= $3
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
				to_timestamp(floor(extract(epoch from time) / $4) * $4) AS bucket_time,
				AVG(latitude)  AS latitude,
				AVG(longitude) AS longitude,
				AVG(speed_kmh) AS speed_kmh,
				AVG(heading)   AS heading,
				BOOL_OR(ignition) AS ignition,
				MAX(speed_kmh) AS max_speed
			FROM telemetry_points
			WHERE vehicle_id = $1
			  AND time >= $2
			  AND time <= $3
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
			vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time
		FROM telemetry_points
		WHERE vehicle_id = ANY($1)
		ORDER BY vehicle_id, time DESC`

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
