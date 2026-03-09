package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core-service/internal/domains/geofencing/models"
)

type GeofenceRepository struct {
	db *pgxpool.Pool
}

func NewGeofenceRepository(db *pgxpool.Pool) *GeofenceRepository {
	return &GeofenceRepository{db: db}
}

func (r *GeofenceRepository) DB() *pgxpool.Pool { return r.db }

// Create creates a new geofence
func (r *GeofenceRepository) Create(ctx context.Context, geofence *models.Geofence) error {
	query := `
		INSERT INTO geofences (
			id, org_id, name, type,
			center, radius_meters, coordinates,
			route_polyline, buffer_meters,
			entry_alert, exit_alert, speed_limit,
			vehicles, metadata,
			enabled, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		geofence.ID,
		geofence.OrgID,
		geofence.Name,
		geofence.Type,
		geofence.Center,
		geofence.RadiusMeters,
		geofence.Coordinates,
		geofence.RoutePolyline,
		geofence.BufferMeters,
		geofence.EntryAlert,
		geofence.ExitAlert,
		geofence.SpeedLimit,
		geofence.Vehicles,
		geofence.Metadata,
		geofence.Active,
		geofence.CreatedBy,
	).Scan(&geofence.CreatedAt, &geofence.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create geofence: %w", err)
	}
	return nil
}

// GetByID retrieves a geofence by ID
func (r *GeofenceRepository) GetByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.Geofence, error) {
	query := `
		SELECT
			id, org_id, name, type,
			center, radius_meters, coordinates,
			route_polyline, buffer_meters,
			entry_alert, exit_alert, speed_limit,
			vehicles, metadata,
			enabled, created_by, created_at, updated_at
		FROM geofences
		WHERE id = $1 AND org_id = $2 AND enabled = true`

	row := r.db.QueryRow(ctx, query, id, orgID)

	var geofence models.Geofence
	err := row.Scan(
		&geofence.ID,
		&geofence.OrgID,
		&geofence.Name,
		&geofence.Type,
		&geofence.Center,
		&geofence.RadiusMeters,
		&geofence.Coordinates,
		&geofence.RoutePolyline,
		&geofence.BufferMeters,
		&geofence.EntryAlert,
		&geofence.ExitAlert,
		&geofence.SpeedLimit,
		&geofence.Vehicles,
		&geofence.Metadata,
		&geofence.Active,
		&geofence.CreatedBy,
		&geofence.CreatedAt,
		&geofence.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get geofence: %w", err)
	}

	return &geofence, nil
}

// List retrieves geofences with pagination and filtering
func (r *GeofenceRepository) List(ctx context.Context, orgID uuid.UUID, req *models.ListGeofencesRequest) ([]models.Geofence, int64, error) {
	baseQuery := `SELECT id, org_id, name, type, 
	                 center, radius_meters, coordinates,
	                 route_polyline, buffer_meters,
	                 entry_alert, exit_alert, speed_limit,
	                 vehicles, metadata,
	                 enabled, created_by, created_at, updated_at,
	                 COUNT(*) OVER() as total
	          FROM geofences 
	          WHERE org_id = $1 AND enabled = true`

	args := []interface{}{orgID}
	argCount := 2

	// Add search filter
	if req.Search != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR type ILIKE $%d)", argCount, argCount+1)
		args = append(args, "%"+req.Search+"%", "%"+req.Search+"%")
		argCount += 2
	}

	// Add type filter
	if req.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, req.Type)
		argCount++
	}

	// Add active filter
	if req.Active != nil {
		baseQuery += fmt.Sprintf(" AND enabled = $%d", argCount)
		args = append(args, *req.Active)
		argCount++
	}

	// Add sorting
	orderClause := "ORDER BY created_at DESC"
	if req.Sort != "" {
		// Parse sort parameter (e.g., "created_at:desc", "name:asc")
		parts := strings.SplitN(req.Sort, ":", 2)
		if len(parts) == 2 {
			field := parts[0]
			direction := parts[1]
			if direction != "asc" && direction != "desc" {
				direction = "desc"
			}
			// Whitelist allowed sort fields to prevent SQL injection
			allowedFields := map[string]bool{"created_at": true, "name": true, "type": true, "updated_at": true}
			if allowedFields[field] {
				orderClause = fmt.Sprintf("ORDER BY %s %s", field, direction)
			}
		}
	}
	baseQuery += " " + orderClause

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.Limit, (req.Page-1)*req.Limit)

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query geofences: %w", err)
	}
	defer rows.Close()

	var geofences []models.Geofence
	var total int64

	for rows.Next() {
		var geofence models.Geofence
		var tempTotal int64

		err := rows.Scan(
			&geofence.ID,
			&geofence.OrgID,
			&geofence.Name,
			&geofence.Type,
			&geofence.Center,
			&geofence.RadiusMeters,
			&geofence.Coordinates,
			&geofence.RoutePolyline,
			&geofence.BufferMeters,
			&geofence.EntryAlert,
			&geofence.ExitAlert,
			&geofence.SpeedLimit,
			&geofence.Vehicles,
			&geofence.Metadata,
			&geofence.Active,
			&geofence.CreatedBy,
			&geofence.CreatedAt,
			&geofence.UpdatedAt,
			&tempTotal,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan geofence: %w", err)
		}

		// Only set total from the first row
		if total == 0 {
			total = tempTotal
		}

		geofences = append(geofences, geofence)
	}

	return geofences, total, nil
}

// Update updates an existing geofence
func (r *GeofenceRepository) Update(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req *models.UpdateGeofenceRequest) error {
	// Build dynamic update query
	query := "UPDATE geofences SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argCount+1)
		args = append(args, *req.Name)
		argCount++
	}
	if req.Type != nil {
		query += fmt.Sprintf(", type = $%d", argCount+1)
		args = append(args, *req.Type)
		argCount++
	}
	if req.RadiusMeters != nil {
		query += fmt.Sprintf(", radius_meters = $%d", argCount+1)
		args = append(args, *req.RadiusMeters)
		argCount++
	}
	if req.Metadata != nil {
		query += fmt.Sprintf(", metadata = $%d", argCount+1)
		args = append(args, req.Metadata)
		argCount++
	}
	if req.Active != nil {
		query += fmt.Sprintf(", enabled = $%d", argCount+1)
		args = append(args, *req.Active)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND org_id = $%d", argCount+1, argCount+2)
	args = append(args, id, orgID)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update geofence: %w", err)
	}
	return nil
}

// Delete marks a geofence as inactive
func (r *GeofenceRepository) Delete(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	query := `UPDATE geofences SET enabled = false, updated_at = NOW() WHERE id = $1 AND org_id = $2`
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete geofence: %w", err)
	}
	return nil
}

// GetEvents retrieves geofence events with filtering
func (r *GeofenceRepository) GetEvents(ctx context.Context, filter *models.GeofenceEventFilter) ([]models.GeofenceEvent, int64, error) {
	baseQuery := `SELECT ge.id, ge.geofence_id, g.name as geofence_name, 
	                 ge.vehicle_id, v.registration_number as vehicle_reg_number,
	                 ge.event_type, ST_X(ge.location::geometry) as lng, ST_Y(ge.location::geometry) as lat,
	                 ge.duration_minutes, ge.time as timestamp, ge.created_at, COUNT(*) OVER() as total
	              FROM geofence_events ge
	              JOIN geofences g ON ge.geofence_id = g.id
	              LEFT JOIN vehicles v ON ge.vehicle_id = v.id
	              WHERE g.org_id = $1`

	args := []interface{}{filter.OrgID}
	argCount := 2

	// Add filters
	if filter.GeofenceID != uuid.Nil {
		baseQuery += fmt.Sprintf(" AND ge.geofence_id = $%d", argCount)
		args = append(args, filter.GeofenceID)
		argCount++
	}
	if filter.VehicleID != uuid.Nil {
		baseQuery += fmt.Sprintf(" AND ge.vehicle_id = $%d", argCount)
		args = append(args, filter.VehicleID)
		argCount++
	}
	if filter.EventType != "" {
		baseQuery += fmt.Sprintf(" AND ge.event_type = $%d", argCount)
		args = append(args, filter.EventType)
		argCount++
	}
	if !filter.From.IsZero() {
		baseQuery += fmt.Sprintf(" AND ge.time >= $%d", argCount)
		args = append(args, filter.From)
		argCount++
	}
	if !filter.To.IsZero() {
		baseQuery += fmt.Sprintf(" AND ge.time <= $%d", argCount)
		args = append(args, filter.To)
		argCount++
	}

	baseQuery += " ORDER BY ge.time DESC"
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query geofence events: %w", err)
	}
	defer rows.Close()

	var events []models.GeofenceEvent
	var total int64

	for rows.Next() {
		var event models.GeofenceEvent
		var tempTotal int64
		var lng, lat float64

		err := rows.Scan(
			&event.ID,
			&event.GeofenceID,
			&event.GeofenceName,
			&event.VehicleID,
			&event.VehicleRegNumber,
			&event.EventType,
			&lng,
			&lat,
			&event.DurationMinutes,
			&event.Timestamp,
			&event.CreatedAt,
			&tempTotal,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan geofence event: %w", err)
		}

		event.Location = models.Location{Lng: lng, Lat: lat}

		// Only set total from the first row
		if total == 0 {
			total = tempTotal
		}

		events = append(events, event)
	}

	return events, total, nil
}
