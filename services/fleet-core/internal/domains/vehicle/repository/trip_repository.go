package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository struct {
	db *pgxpool.Pool
}

func NewTripRepository(db *pgxpool.Pool) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) Create(ctx context.Context, trip *models.Trip) error {
	startLocJSON, _ := json.Marshal(trip.StartLocation)
	endLocJSON, _ := json.Marshal(trip.EndLocation)

	query := `
		INSERT INTO trips (id, trip_id, vehicle_id, driver_id, start_location, end_location,
			start_time, end_time, distance, status, tenant_id, fuel_consumed_liters, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		trip.ID, trip.TripID, trip.VehicleID, trip.DriverID,
		startLocJSON, endLocJSON, trip.StartTime, trip.EndTime,
		trip.Distance, trip.Status, trip.OrganizationID, trip.FuelConsumed, trip.CreatedAt, trip.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *TripRepository) scanTrip(row interface {
	Scan(dest ...interface{}) error
}) (*models.Trip, error) {
	var trip models.Trip
	var startLocJSON, endLocJSON []byte
	err := row.Scan(
		&trip.ID, &trip.TripID, &trip.VehicleID, &trip.DriverID,
		&startLocJSON, &endLocJSON, &trip.StartTime, &trip.EndTime,
		&trip.Distance, &trip.Status, &trip.OrganizationID, &trip.FuelConsumed, &trip.CreatedAt, &trip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(startLocJSON, &trip.StartLocation)
	json.Unmarshal(endLocJSON, &trip.EndLocation)
	return &trip, nil
}

const tripSelectCols = `id, trip_id, vehicle_id, driver_id, start_location, end_location,
			start_time, end_time, distance, status, tenant_id, fuel_consumed_liters, created_at, updated_at`

func (r *TripRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
	query := fmt.Sprintf(`SELECT %s FROM trips WHERE id = $1`, tripSelectCols)
	start := time.Now()
	row := r.db.QueryRow(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return r.scanTrip(row)
}

func (r *TripRepository) GetByTripID(ctx context.Context, tripID string) (*models.Trip, error) {
	query := fmt.Sprintf(`SELECT %s FROM trips WHERE trip_id = $1`, tripSelectCols)
	start := time.Now()
	row := r.db.QueryRow(ctx, query, tripID)
	metrics.RecordDBQuery(query, time.Since(start))
	return r.scanTrip(row)
}

func (r *TripRepository) List(ctx context.Context, params models.TripListParams) ([]models.Trip, int, error) {
	offset := (params.Page - 1) * params.Limit

	// Build dynamic query
	baseQuery := `FROM trips WHERE tenant_id = $1`
	args := []interface{}{params.OrganizationID}
	argIdx := 2

	if params.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}
	if params.VehicleID != nil {
		baseQuery += fmt.Sprintf(" AND vehicle_id = $%d", argIdx)
		args = append(args, *params.VehicleID)
		argIdx++
	}
	if params.DriverID != nil {
		baseQuery += fmt.Sprintf(" AND driver_id = $%d", argIdx)
		args = append(args, *params.DriverID)
		argIdx++
	}
	if params.StartDate != nil {
		baseQuery += fmt.Sprintf(" AND start_time >= $%d", argIdx)
		args = append(args, *params.StartDate)
		argIdx++
	}
	if params.EndDate != nil {
		baseQuery += fmt.Sprintf(" AND start_time <= $%d", argIdx)
		args = append(args, *params.EndDate)
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + baseQuery
	start := time.Now()
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		metrics.RecordDBQuery(countQuery, time.Since(start))
		return nil, 0, err
	}
	metrics.RecordDBQuery(countQuery, time.Since(start))

	// List query
	listQuery := fmt.Sprintf(`SELECT %s %s ORDER BY start_time DESC LIMIT $%d OFFSET $%d`,
		tripSelectCols, baseQuery, argIdx, argIdx+1)
	args = append(args, params.Limit, offset)

	listStart := time.Now()
	rows, err := r.db.Query(ctx, listQuery, args...)
	metrics.RecordDBQuery(listQuery, time.Since(listStart))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var trips []models.Trip
	for rows.Next() {
		var trip models.Trip
		var startLocJSON, endLocJSON []byte
		if err := rows.Scan(
			&trip.ID, &trip.TripID, &trip.VehicleID, &trip.DriverID,
			&startLocJSON, &endLocJSON, &trip.StartTime, &trip.EndTime,
			&trip.Distance, &trip.Status, &trip.OrganizationID, &trip.FuelConsumed, &trip.CreatedAt, &trip.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(startLocJSON, &trip.StartLocation)
		json.Unmarshal(endLocJSON, &trip.EndLocation)
		trips = append(trips, trip)
	}
	return trips, total, nil
}

func (r *TripRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateTripRequest) error {
	var sets []string
	args := []interface{}{}
	argIdx := 1

	if req.DriverID != nil {
		sets = append(sets, fmt.Sprintf("driver_id = $%d", argIdx))
		args = append(args, req.DriverID)
		argIdx++
	}
	if req.StartLocation != nil {
		startLocJSON, _ := json.Marshal(req.StartLocation)
		sets = append(sets, fmt.Sprintf("start_location = $%d", argIdx))
		args = append(args, startLocJSON)
		argIdx++
	}
	if req.EndLocation != nil {
		endLocJSON, _ := json.Marshal(req.EndLocation)
		sets = append(sets, fmt.Sprintf("end_location = $%d", argIdx))
		args = append(args, endLocJSON)
		argIdx++
	}
	if req.StartTime != nil {
		sets = append(sets, fmt.Sprintf("start_time = $%d", argIdx))
		args = append(args, req.StartTime)
		argIdx++
	}
	if req.EndTime != nil {
		sets = append(sets, fmt.Sprintf("end_time = $%d", argIdx))
		args = append(args, req.EndTime)
		argIdx++
	}
	if req.Distance != nil {
		sets = append(sets, fmt.Sprintf("distance = $%d", argIdx))
		args = append(args, req.Distance)
		argIdx++
	}
	if req.Status != "" {
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, req.Status)
		argIdx++
	}

	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE trips SET %s WHERE id = $%d", strings.Join(sets, ", "), argIdx)
	args = append(args, id)

	start := time.Now()
	_, err := r.db.Exec(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *TripRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM trips WHERE id = $1"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// Telemetry methods
func (r *TripRepository) AddTelemetry(ctx context.Context, t *models.Telemetry) error {
	query := `
		INSERT INTO telemetry (id, trip_id, latitude, longitude, speed, heading, event_type, event_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, t.ID, t.TripID, t.Latitude, t.Longitude, t.Speed, t.Heading, t.EventType, t.EventTime)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *TripRepository) GetTelemetry(ctx context.Context, tripID uuid.UUID, limit int) ([]models.Telemetry, error) {
	query := `
		SELECT id, trip_id, latitude, longitude, speed, heading, event_type, event_time
		FROM telemetry WHERE trip_id = $1 ORDER BY event_time DESC LIMIT $2
	`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, tripID, limit)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Telemetry
	for rows.Next() {
		var t models.Telemetry
		if err := rows.Scan(&t.ID, &t.TripID, &t.Latitude, &t.Longitude, &t.Speed, &t.Heading, &t.EventType, &t.EventTime); err != nil {
			return nil, err
		}
		events = append(events, t)
	}
	return events, nil
}

// Settings methods
func (r *TripRepository) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.TripSettings, error) {
	query := `SELECT id, tenant_id, retention_period, policies, updated_at FROM trip_settings WHERE tenant_id = $1`
	var s models.TripSettings
	var policiesJSON []byte
	start := time.Now()
	err := r.db.QueryRow(ctx, query, orgID).Scan(&s.ID, &s.OrganizationID, &s.RetentionPeriod, &policiesJSON, &s.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		// Return defaults if not found
		return &models.TripSettings{
			OrganizationID:  orgID,
			RetentionPeriod: 36,
			Policies:        map[string]interface{}{},
			UpdatedAt:       time.Now(),
		}, nil
	}
	json.Unmarshal(policiesJSON, &s.Policies)
	return &s, nil
}

func (r *TripRepository) UpsertSettings(ctx context.Context, s *models.TripSettings) error {
	policiesJSON, _ := json.Marshal(s.Policies)
	query := `
		INSERT INTO trip_settings (id, tenant_id, retention_period, policies, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			retention_period = $3, policies = $4, updated_at = NOW()
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, uuid.New(), s.OrganizationID, s.RetentionPeriod, policiesJSON)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}
