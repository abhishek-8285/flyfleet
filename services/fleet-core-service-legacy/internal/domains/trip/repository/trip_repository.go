package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
)

type TripRepository struct {
	db *pgxpool.Pool
}

func NewTripRepository(db *pgxpool.Pool) *TripRepository {
	return &TripRepository{db: db}
}

// Create inserts a new trip into the database
func (r *TripRepository) Create(ctx context.Context, trip *models.Trip) error {
	query := `
		INSERT INTO trips (
			id, organization_id, vehicle_id, driver_id, status, title, description,
			planned_route, waypoints, estimated_distance_km, estimated_duration_min,
			cargo_details, scheduled_start, customer_name, customer_phone, customer_email,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	plannedRouteJSON, _ := json.Marshal(trip.PlannedRoute)
	waypointsJSON, _ := json.Marshal(trip.Waypoints)
	cargoDetailsJSON, _ := json.Marshal(trip.CargoDetails)

	err := r.db.QueryRow(ctx, query,
		id, trip.OrgID, trip.VehicleID, trip.DriverID, trip.Status, trip.Title, trip.Description,
		plannedRouteJSON, waypointsJSON, trip.EstimatedDistanceKm, trip.EstimatedDurationMin,
		cargoDetailsJSON, trip.ScheduledStart, trip.CustomerName, trip.CustomerPhone, trip.CustomerEmail,
		now, now,
	).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)

	return err
}

// GetByID retrieves a trip by its ID
func (r *TripRepository) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	query := `
		SELECT id, organization_id, vehicle_id, driver_id, status, title, description,
		       planned_route, waypoints, estimated_distance_km, actual_distance_km,
		       estimated_duration_min, actual_duration_min, cargo_details, scheduled_start,
		       started_at, ended_at, start_location, end_location, fuel_consumed_liters,
		       idling_time_minutes, events, pod_documents, customer_name, customer_phone,
		       customer_email, cancellation_reason, created_at, updated_at
		FROM trips WHERE id = $1`

	var trip models.Trip
	var plannedRouteJSON, waypointsJSON, cargoDetailsJSON, eventsJSON, podDocsJSON []byte
	var startLocJSON, endLocJSON []byte
	var cancellationReason sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status, &trip.Title, &trip.Description,
		&plannedRouteJSON, &waypointsJSON, &trip.EstimatedDistanceKm, &trip.ActualDistanceKm,
		&trip.EstimatedDurationMin, &trip.ActualDurationMin, &cargoDetailsJSON, &trip.ScheduledStart,
		&trip.StartedAt, &trip.EndedAt, &startLocJSON, &endLocJSON, &trip.FuelConsumedLiters,
		&trip.IdlingTimeMinutes, &eventsJSON, &podDocsJSON, &trip.CustomerName, &trip.CustomerPhone,
		&trip.CustomerEmail, &cancellationReason, &trip.CreatedAt, &trip.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if cancellationReason.Valid {
		trip.CancellationReason = cancellationReason.String
	}

	// Parse JSON fields
	json.Unmarshal(plannedRouteJSON, &trip.PlannedRoute)
	json.Unmarshal(waypointsJSON, &trip.Waypoints)
	json.Unmarshal(cargoDetailsJSON, &trip.CargoDetails)
	json.Unmarshal(eventsJSON, &trip.Events)
	json.Unmarshal(podDocsJSON, &trip.PODDocuments)
	json.Unmarshal(startLocJSON, &trip.StartLocation)
	json.Unmarshal(endLocJSON, &trip.EndLocation)

	return &trip, nil
}

// List retrieves trips with filtering and pagination
func (r *TripRepository) List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error) {
	baseQuery := `
		SELECT id, organization_id, vehicle_id, driver_id, status, title, description,
		       estimated_distance_km, actual_distance_km, estimated_duration_min,
		       actual_duration_min, scheduled_start, started_at, ended_at,
		       customer_name, created_at, updated_at
		FROM trips`

	countQuery := `SELECT COUNT(*) FROM trips`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.VehicleID != "" {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, filter.VehicleID)
		argIndex++
	}

	if filter.DriverID != "" {
		conditions = append(conditions, fmt.Sprintf("driver_id = $%d", argIndex))
		args = append(args, filter.DriverID)
		argIndex++
	}

	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("scheduled_start >= $%d", argIndex))
		args = append(args, filter.DateFrom)
		argIndex++
	}

	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("scheduled_start <= $%d", argIndex))
		args = append(args, filter.DateTo)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query
	var totalCount int
	countSQL := countQuery + whereClause
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// List query with pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		allowedColumns := map[string]bool{
			"created_at": true, "scheduled_start": true, "status": true,
			"vehicle_id": true, "driver_id": true,
		}
		if allowedColumns[filter.OrderBy] {
			direction := "ASC"
			if strings.ToUpper(filter.OrderDir) == "DESC" {
				direction = "DESC"
			}
			orderBy = fmt.Sprintf("%s %s", filter.OrderBy, direction)
		}
	}

	listSQL := baseQuery + whereClause + " ORDER BY " + orderBy + limitClause

	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var trips []*models.Trip
	for rows.Next() {
		var trip models.Trip
		err := rows.Scan(
			&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status,
			&trip.Title, &trip.Description, &trip.EstimatedDistanceKm, &trip.ActualDistanceKm,
			&trip.EstimatedDurationMin, &trip.ActualDurationMin, &trip.ScheduledStart,
			&trip.StartedAt, &trip.EndedAt, &trip.CustomerName, &trip.CreatedAt, &trip.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		trips = append(trips, &trip)
	}

	return trips, totalCount, nil
}

// Update modifies an existing trip
func (r *TripRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		
		// Handle special JSON fields
		switch key {
		case "planned_route", "waypoints", "cargo_details", "events", "pod_documents", "start_location", "end_location":
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE trips SET %s, updated_at = $%d WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete removes a trip (soft delete by setting status to deleted)
func (r *TripRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE trips SET status = 'deleted', updated_at = NOW() WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// AddEvent adds an event to a trip
func (r *TripRepository) AddEvent(ctx context.Context, tripID string, event models.TripEvent) error {
	query := `
		UPDATE trips 
		SET events = CASE 
			WHEN events IS NULL THEN $1::jsonb
			ELSE (events::jsonb || $1::jsonb)
		END,
		updated_at = NOW()
		WHERE id = $2`

	eventJSON, _ := json.Marshal([]models.TripEvent{event})
	_, err := r.db.Exec(ctx, query, eventJSON, tripID)
	return err
}

// AddPODDocument adds a POD document to a trip
func (r *TripRepository) AddPODDocument(ctx context.Context, tripID string, pod models.PODDocument) error {
	query := `
		UPDATE trips 
		SET pod_documents = CASE 
			WHEN pod_documents IS NULL THEN $1::jsonb
			ELSE (pod_documents::jsonb || $1::jsonb)
		END,
		updated_at = NOW()
		WHERE id = $2`

	podJSON, _ := json.Marshal([]models.PODDocument{pod})
	_, err := r.db.Exec(ctx, query, podJSON, tripID)
	return err
}

// GetActiveTripsForVehicle retrieves active trips for a vehicle
func (r *TripRepository) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	query := `
		SELECT id, org_id, vehicle_id, driver_id, status, title, description,
		       estimated_distance_km, actual_distance_km, estimated_duration_min,
		       actual_duration_min, scheduled_start, started_at, ended_at,
		       customer_name, created_at, updated_at
		FROM trips 
		WHERE vehicle_id = $1 AND status IN ('planned', 'in_progress', 'paused')
		ORDER BY scheduled_start ASC`

	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*models.Trip
	for rows.Next() {
		var trip models.Trip
		err := rows.Scan(
			&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status,
			&trip.Title, &trip.Description, &trip.EstimatedDistanceKm, &trip.ActualDistanceKm,
			&trip.EstimatedDurationMin, &trip.ActualDurationMin, &trip.ScheduledStart,
			&trip.StartedAt, &trip.EndedAt, &trip.CustomerName, &trip.CreatedAt, &trip.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, &trip)
	}

	return trips, nil
}

// GetTripSummary retrieves trip analytics for an organization
func (r *TripRepository) GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_trips,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_trips,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_trips,
			COALESCE(SUM(actual_distance_km), 0) as total_distance_km,
			COALESCE(SUM(actual_duration_min), 0) as total_duration_minutes,
			COALESCE(AVG(actual_distance_km / NULLIF(fuel_consumed_liters, 0)), 0) as avg_fuel_efficiency,
			COALESCE(SUM(fuel_consumed_liters), 0) as total_fuel_consumed,
			COALESCE(SUM(idling_time_minutes), 0) as total_idle_time_minutes
		FROM trips 
		WHERE organization_id = $1 
		  AND created_at >= $2 
		  AND created_at <= $3
		  AND status IN ('completed', 'cancelled')`

	var summary models.TripSummary
	var totalDurationMinutes float64

	err := r.db.QueryRow(ctx, query, orgID, startDate, endDate).Scan(
		&summary.TotalTrips, &summary.CompletedTrips, &summary.CancelledTrips,
		&summary.TotalDistanceKm, &totalDurationMinutes, &summary.AverageFuelEfficiency,
		&summary.TotalFuelConsumed, &summary.TotalIdleTimeMinutes,
	)

	if err != nil {
		return nil, err
	}

	summary.TotalDurationHours = totalDurationMinutes / 60.0
	// TotalIdleTimeMinutes is already in minutes, no conversion needed
	
	if summary.TotalTrips > 0 {
		summary.OnTimeRate = float64(summary.CompletedTrips) / float64(summary.TotalTrips) * 100
		summary.AverageTripDuration = totalDurationMinutes / float64(summary.CompletedTrips) / 60.0
	}

	return &summary, nil
}
func (r *TripRepository) GetStats(ctx context.Context) (*models.TripStats, error) {
	query := `SELECT COUNT(*) FROM trips WHERE status IN ('in_progress', 'paused') AND is_deleted = FALSE`
	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &models.TripStats{ActiveTrips: count}, nil
}
