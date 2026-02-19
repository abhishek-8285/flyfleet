package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TelemetryRepository struct {
	db *pgxpool.Pool
}

func NewTelemetryRepository(db *pgxpool.Pool) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

func (r *TelemetryRepository) GetTelemetryByTripID(ctx context.Context, tripID, orgID string, filter models.TelemetryFilter) ([]models.TelemetryPoint, error) {
	// First, get trip details to validate org_id and get time range + vehicle_id
	var vehicleID string
	var startTime time.Time
	var endTime *time.Time
	
	tripQuery := `SELECT vehicle_id, started_at, ended_at FROM trips WHERE id = $1 AND organization_id = $2`
	err := r.db.QueryRow(ctx, tripQuery, tripID, orgID).Scan(&vehicleID, &startTime, &endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trip for telemetry: %w", err)
	}

	if startTime.IsZero() {
		return []models.TelemetryPoint{}, nil // Trip hasn't started
	}

	// Use filter time range if provided, else fall back to trip times
	queryStart := startTime
	if filter.StartTime != nil {
		queryStart = *filter.StartTime
	}

	queryEnd := time.Now()
	if filter.EndTime != nil {
		queryEnd = *filter.EndTime
	} else if endTime != nil {
		queryEnd = *endTime
	}

	limit := filter.Limit
	if limit == 0 {
		limit = 1000
	}

	var query string
	var args []interface{}

	switch filter.Aggregation {
	case "hourly":
		query = `
			SELECT 
				time_bucket('1 hour', time) AS time,
				AVG(latitude) as latitude,
				AVG(longitude) as longitude,
				AVG(speed_kph) as speed_kph,
				CAST(AVG(heading) AS SMALLINT) as heading,
				MAX(odometer_km) as odometer_km,
				CAST(AVG(fuel_level_percent) AS SMALLINT) as fuel_level_percent,
				CAST(AVG(engine_temp_celsius) AS SMALLINT) as engine_temp_celsius,
				'[]'::jsonb as alerts
			FROM vehicle_telemetry
			WHERE vehicle_id = $1 AND time >= $2 AND time <= $3
			GROUP BY time
			ORDER BY time ASC
			LIMIT $4`
		args = []interface{}{vehicleID, queryStart, queryEnd, limit}
	case "daily":
		query = `
			SELECT 
				time_bucket('1 day', time) AS time,
				AVG(latitude) as latitude,
				AVG(longitude) as longitude,
				AVG(speed_kph) as speed_kph,
				CAST(AVG(heading) AS SMALLINT) as heading,
				MAX(odometer_km) as odometer_km,
				CAST(AVG(fuel_level_percent) AS SMALLINT) as fuel_level_percent,
				CAST(AVG(engine_temp_celsius) AS SMALLINT) as engine_temp_celsius,
				'[]'::jsonb as alerts
			FROM vehicle_telemetry
			WHERE vehicle_id = $1 AND time >= $2 AND time <= $3
			GROUP BY time
			ORDER BY time ASC
			LIMIT $4`
		args = []interface{}{vehicleID, queryStart, queryEnd, limit}
	default: // raw
		query = `
			SELECT 
				time, latitude, longitude, speed_kph, heading, 
				odometer_km, fuel_level_percent, engine_temp_celsius, alerts
			FROM vehicle_telemetry
			WHERE vehicle_id = $1 AND time >= $2 AND time <= $3
			ORDER BY time ASC
			LIMIT $4`
		args = []interface{}{vehicleID, queryStart, queryEnd, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry: %w", err)
	}
	defer rows.Close()

	var points []models.TelemetryPoint
	for rows.Next() {
		var p models.TelemetryPoint
		var alertsJSON []byte
		err := rows.Scan(
			&p.Timestamp, &p.Latitude, &p.Longitude, &p.SpeedKph, &p.Heading,
			&p.OdometerKm, &p.FuelLevelPercent, &p.EngineTempCelsius, &alertsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry point: %w", err)
		}
		// In a real scenario, we'd unmarshal alertsJSON to p.Alerts
		// For now, keeping it simple as per schema
		points = append(points, p)
	}

	return points, nil
}
