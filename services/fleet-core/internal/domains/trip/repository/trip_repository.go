package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
)

// TripRepository defines the interface for trip data persistence
type TripRepository interface {
	// Trip CRUD
	Create(ctx context.Context, trip *models.Trip) error
	GetByID(ctx context.Context, id string) (*models.Trip, error)
	List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error

	// Trip lifecycle
	Approve(ctx context.Context, id string) error
	AddEvent(ctx context.Context, tripID string, event models.TripEvent) error
	GetLastEvent(ctx context.Context, tripID string) (*models.TripEvent, error)
	AddPODDocument(ctx context.Context, tripID string, pod models.PODDocument) error
	GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error)
	GetActiveTripsForDriver(ctx context.Context, driverID string) ([]*models.Trip, error)
	GetLastCompletedTripForVehicle(ctx context.Context, vehicleID string) (*models.Trip, error)
	CheckOverlappingTrips(ctx context.Context, vehicleID, driverID string, start, end time.Time) (bool, error)

	// Telemetry
	AddTelemetryPoint(ctx context.Context, point models.TelemetryPoint) error
	GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error)

	// Analytics
	GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error)
	GetStats(ctx context.Context) (*models.TripStats, error)
	GetDriverStats(ctx context.Context, orgID string) ([]models.DriverPerformance, error)
	GetTripStats(ctx context.Context, orgID string) ([]models.DailyTripTrend, error)

	// Trip Offers
	CreateOffer(ctx context.Context, offer *models.TripOffer) error
	GetOffer(ctx context.Context, offerID string) (*models.TripOffer, error)
	RespondToOffer(ctx context.Context, offerID string, status, reason string) error
	GetOffersForDriver(ctx context.Context, driverID string) ([]*models.TripOffer, error)

	// Proof of Delivery
	SaveProofOfDelivery(ctx context.Context, pod *models.ProofOfDelivery) error
	GetPODsForTrip(ctx context.Context, tripID string) ([]*models.ProofOfDelivery, error)

	// Transactions
	WithTx(ctx context.Context, fn func(TripRepository) error) error
}

// PostgresTripRepository implements TripRepository with PostgreSQL
type PostgresTripRepository struct {
	db *pgxpool.Pool
}

// NewTripRepository creates a new trip repository
func NewTripRepository(db *pgxpool.Pool) *PostgresTripRepository {
	return &PostgresTripRepository{db: db}
}

// Create inserts a new trip into the database
func (r *PostgresTripRepository) Create(ctx context.Context, trip *models.Trip) error {
	query := `
		INSERT INTO trips (
			id, org_id, vehicle_id, driver_id, status, title, description,
			pickup_geo, drop_geo,
			planned_route, waypoints, planned_distance, estimated_duration_min,
			cargo_details, scheduled_start, start_odometer, fuel_level_snapshot,
			eway_bill_no, gst_classification, customer_name, customer_phone,
			customer_email, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			CASE WHEN $8::text IS NOT NULL THEN ST_GeomFromText($8) ELSE NULL END,
			CASE WHEN $9::text IS NOT NULL THEN ST_GeomFromText($9) ELSE NULL END,
			$10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	plannedRouteJSON, _ := json.Marshal(trip.PlannedRoute)
	waypointsJSON, _ := json.Marshal(trip.Waypoints)
	cargoDetailsJSON, _ := json.Marshal(trip.CargoDetails)

	var pickupWKT, dropWKT *string
	if trip.PickupGeo != nil {
		s := fmt.Sprintf("POINT(%f %f)", trip.PickupGeo.Longitude, trip.PickupGeo.Latitude)
		pickupWKT = &s
	}
	if trip.DropGeo != nil {
		s := fmt.Sprintf("POINT(%f %f)", trip.DropGeo.Longitude, trip.DropGeo.Latitude)
		dropWKT = &s
	}

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		id, trip.OrgID, trip.VehicleID, trip.DriverID, trip.Status, trip.Title, trip.Description,
		pickupWKT, dropWKT,
		plannedRouteJSON, waypointsJSON, trip.PlannedDistance, trip.EstimatedDurationMin,
		cargoDetailsJSON, trip.ScheduledStart, trip.StartOdometer, trip.FuelLevelSnapshot,
		trip.EWayBillNo, trip.GSTClassification, trip.CustomerName, trip.CustomerPhone,
		trip.CustomerEmail, now, now,
	).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	return err
}

// GetByID retrieves a trip by its ID
func (r *PostgresTripRepository) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	query := `
		SELECT
			id, org_id, vehicle_id, driver_id, status, title, description,
			ST_AsText(pickup_geo), ST_AsText(drop_geo),
			planned_route, actual_route, waypoints,
			planned_distance, total_distance, estimated_duration_min, actual_duration_min,
			cargo_details, scheduled_start, start_time, end_time,
			start_odometer, end_odometer, fuel_level_snapshot, fuel_used, idling_time_minutes,
			eway_bill_no, gst_classification, permit_verified, ais140_compliant,
			customer_name, customer_phone, customer_email, cancellation_reason,
			version, created_at, updated_at
		FROM trips
		WHERE id = $1 AND is_deleted = FALSE`

	var trip models.Trip
	var pickupWKT, dropWKT sql.NullString
	var plannedRouteJSON, actualRouteJSON, waypointsJSON, cargoDetailsJSON []byte
	var cancellationReason sql.NullString
	var totalDistance sql.NullFloat64
	var actualDurationMin sql.NullInt32
	var fuelUsed sql.NullFloat64
	var idlingTimeMinutes sql.NullInt32

	start := time.Now()
	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status, &trip.Title, &trip.Description,
		&pickupWKT, &dropWKT,
		&plannedRouteJSON, &actualRouteJSON, &waypointsJSON,
		&trip.PlannedDistance, &totalDistance, &trip.EstimatedDurationMin, &actualDurationMin,
		&cargoDetailsJSON, &trip.ScheduledStart, &trip.StartTime, &trip.EndTime,
		&trip.StartOdometer, &trip.EndOdometer, &trip.FuelLevelSnapshot, &fuelUsed, &idlingTimeMinutes,
		&trip.EWayBillNo, &trip.GSTClassification, &trip.PermitVerified, &trip.AIS140Compliant,
		&trip.CustomerName, &trip.CustomerPhone, &trip.CustomerEmail, &cancellationReason,
		&trip.Version, &trip.CreatedAt, &trip.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		return nil, err
	}

	if cancellationReason.Valid {
		trip.CancellationReason = cancellationReason.String
	}

	// Parse JSON fields
	json.Unmarshal(plannedRouteJSON, &trip.PlannedRoute)
	json.Unmarshal(actualRouteJSON, &trip.ActualRoute)
	json.Unmarshal(waypointsJSON, &trip.Waypoints)
	json.Unmarshal(cargoDetailsJSON, &trip.CargoDetails)

	// Parse Geo fields
	if pickupWKT.Valid {
		var lat, lng float64
		fmt.Sscanf(pickupWKT.String, "POINT(%f %f)", &lng, &lat)
		trip.PickupGeo = &models.Location{Latitude: lat, Longitude: lng}
	}
	if dropWKT.Valid {
		var lat, lng float64
		fmt.Sscanf(dropWKT.String, "POINT(%f %f)", &lng, &lat)
		trip.DropGeo = &models.Location{Latitude: lat, Longitude: lng}
	}

	return &trip, nil
}

// List retrieves trips with filtering and pagination
func (r *PostgresTripRepository) List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error) {
	baseQuery := `
		SELECT
			id, org_id, vehicle_id, driver_id, status, title, description,
			estimated_distance_km, actual_distance_km, estimated_duration_min, actual_duration_min,
			scheduled_start, started_at, ended_at,
			customer_name, created_at, updated_at
		FROM trips
		WHERE org_id = $1 AND is_deleted = FALSE`

	countQuery := `SELECT COUNT(*) FROM trips WHERE org_id = $1 AND is_deleted = FALSE`

	var conditions []string
	args := []interface{}{filter.OrgID}
	argIndex := 2

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

	if filter.From != "" {
		conditions = append(conditions, fmt.Sprintf("scheduled_start >= $%d", argIndex))
		args = append(args, filter.From)
		argIndex++
	}

	if filter.To != "" {
		conditions = append(conditions, fmt.Sprintf("scheduled_start <= $%d", argIndex))
		args = append(args, filter.To)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	var totalCount int
	countSQL := countQuery + whereClause
	start := time.Now()
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	metrics.RecordDBQuery(countSQL, time.Since(start))
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

	start = time.Now()
	rows, err := r.db.Query(ctx, listSQL, args...)
	metrics.RecordDBQuery(listSQL, time.Since(start))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var trips []*models.Trip
	for rows.Next() {
		var trip models.Trip
		var estDist sql.NullFloat64
		var actDist sql.NullFloat64
		var estDur sql.NullInt32
		var actDur sql.NullInt32
		var schedStart sql.NullTime
		var startedAt sql.NullTime
		var endedAt sql.NullTime
		var custName sql.NullString
		
		err := rows.Scan(
			&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status,
			&trip.Title, &trip.Description,
			&estDist, &actDist, &estDur, &actDur,
			&schedStart, &startedAt, &endedAt,
			&custName, &trip.CreatedAt, &trip.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		if estDist.Valid {
			trip.PlannedDistance = estDist.Float64
		}
		if actDist.Valid {
			trip.TotalDistance = &actDist.Float64
		}
		if estDur.Valid {
			trip.EstimatedDurationMin = int(estDur.Int32)
		}
		if actDur.Valid {
			v := int(actDur.Int32)
			trip.ActualDurationMin = &v
		}
		if schedStart.Valid {
			trip.ScheduledStart = schedStart.Time
		}
		if startedAt.Valid {
			t := startedAt.Time
			trip.StartTime = &t
		}
		if endedAt.Valid {
			t := endedAt.Time
			trip.EndTime = &t
		}
		if custName.Valid {
			trip.CustomerName = custName.String
		}
		
		trips = append(trips, &trip)
	}

	return trips, totalCount, nil
}

// Update modifies an existing trip
func (r *PostgresTripRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Mapping Go field names to DB column names if they differ
	fieldMap := map[string]string{
		"OrgID":                "org_id",
		"PlannedDistance":      "planned_distance",
		"TotalDistance":        "total_distance",
		"StartTime":            "start_time",
		"EndTime":              "end_time",
		"FuelUsed":             "fuel_used",
		"GSTClassification":    "gst_classification",
		"PlannedRoute":         "planned_route",
		"ActualRoute":          "actual_route",
		"Waypoints":            "waypoints",
		"CargoDetails":         "cargo_details",
		"StartLocation":        "start_location",
		"EndLocation":          "end_location",
		"PickupGeo":            "pickup_geo",
		"DropGeo":              "drop_geo",
		"ScheduledStart":       "scheduled_start",
		"StartOdometer":        "start_odometer",
		"EndOdometer":          "end_odometer",
		"FuelLevelSnapshot":    "fuel_level_snapshot",
		"EWayBillNo":           "eway_bill_no",
		"PermitVerified":       "permit_verified",
		"AIS140Compliant":      "ais140_compliant",
		"CustomerName":         "customer_name",
		"CustomerPhone":        "customer_phone",
		"CustomerEmail":        "customer_email",
		"CancellationReason":   "cancellation_reason",
		"EstimatedDurationMin": "estimated_duration_min",
		"ActualDurationMin":    "actual_duration_min",
		"IdlingTimeMinutes":    "idling_time_minutes",
		"VehicleID":            "vehicle_id",
		"DriverID":             "driver_id",
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		columnName := key
		if mapped, ok := fieldMap[key]; ok {
			columnName = mapped
		}

		setParts = append(setParts, fmt.Sprintf("%s = $%d", columnName, argIndex))

		// Handle special JSON/Geo fields
		switch columnName {
		case "planned_route", "actual_route", "waypoints", "cargo_details", "start_location", "end_location":
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		case "pickup_geo", "drop_geo":
			if loc, ok := value.(*models.Location); ok && loc != nil {
				wkt := fmt.Sprintf("POINT(%f %f)", loc.Longitude, loc.Latitude)
				args = append(args, wkt)
			} else {
				args = append(args, nil) // Set to NULL if location is nil
			}
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)

	// Check for version to implement optimistic locking
	versionWhere := ""
	if v, ok := updates["version"]; ok {
		versionWhere = fmt.Sprintf(" AND version = $%d", argIndex+1)
		args = append(args, v)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE trips SET %s, updated_at = $%d, version = version + 1 WHERE id = $%d%s",
		strings.Join(setParts, ", "), argIndex, argIndex+1, versionWhere)

	start := time.Now()
	result, err := r.db.Exec(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete removes a trip (soft delete by setting status to deleted)
func (r *PostgresTripRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE trips SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1"
	start := time.Now()
	result, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Approve moves a trip from CREATED to SCHEDULED
func (r *PostgresTripRepository) Approve(ctx context.Context, id string) error {
	query := `UPDATE trips SET status = 'SCHEDULED', updated_at = NOW() WHERE id = $1 AND status = 'CREATED'`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// AddEvent adds an event to a trip with hash chaining
func (r *PostgresTripRepository) AddEvent(ctx context.Context, tripID string, event models.TripEvent) error {
	query := `
		INSERT INTO trip_events (
			id, trip_id, event_type, location, payload, hash, prev_hash, created_at
		) VALUES ($1, $2, $3, ST_GeomFromText($4), $5, $6, $7, $8)`

	id := uuid.New().String()
	now := time.Now()
	locationWKT := fmt.Sprintf("POINT(%f %f)", event.Location.Longitude, event.Location.Latitude)
	payloadJSON, _ := json.Marshal(event.Payload) // Assuming Payload field exists or use JSONB

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		id, tripID, event.Type, locationWKT, payloadJSON, event.Hash, event.PrevHash, now,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// GetLastEvent retrieves the most recent event for a trip to get its hash
func (r *PostgresTripRepository) GetLastEvent(ctx context.Context, tripID string) (*models.TripEvent, error) {
	query := `SELECT id, event_type, hash FROM trip_events WHERE trip_id = $1 ORDER BY created_at DESC LIMIT 1`
	var event models.TripEvent
	start := time.Now()
	err := r.db.QueryRow(ctx, query, tripID).Scan(&event.ID, &event.Type, &event.Hash)
	metrics.RecordDBQuery(query, time.Since(start))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &event, err
}

// AddPODDocument adds a POD document to a trip
func (r *PostgresTripRepository) AddPODDocument(ctx context.Context, tripID string, pod models.PODDocument) error {
	query := `
		UPDATE trips 
		SET pod_documents = CASE 
			WHEN pod_documents IS NULL THEN $1::jsonb
			ELSE (pod_documents::jsonb || $1::jsonb)
		END,
		updated_at = NOW()
		WHERE id = $2`

	podJSON, _ := json.Marshal([]models.PODDocument{pod})
	start := time.Now()
	_, err := r.db.Exec(ctx, query, podJSON, tripID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// GetActiveTripsForDriver retrieves trips for a driver that are not in terminal states
func (r *PostgresTripRepository) GetActiveTripsForDriver(ctx context.Context, driverID string) ([]*models.Trip, error) {
	query := `
		SELECT id, org_id, vehicle_id, driver_id, status, title, created_at, updated_at
		FROM trips
		WHERE driver_id = $1
		  AND status NOT IN ('COMPLETED', 'CANCELLED', 'FAILED')
		  AND is_deleted = FALSE`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*models.Trip
	for rows.Next() {
		var trip models.Trip
		if err := rows.Scan(&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status, &trip.Title, &trip.CreatedAt, &trip.UpdatedAt); err != nil {
			return nil, err
		}
		trips = append(trips, &trip)
	}
	return trips, nil
}

// AddTelemetryPoint adds a telemetry point to the partitioned table
func (r *PostgresTripRepository) AddTelemetryPoint(ctx context.Context, point models.TelemetryPoint) error {
	query := `
		INSERT INTO telemetry_points (
			trip_id, lat, lng, speed, heading, fuel_level, ignition_status, recorded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		point.TripID, point.Latitude, point.Longitude, point.Speed, point.Heading,
		point.FuelLevel, point.IgnitionStatus, point.RecordedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// GetTelemetry retrieves telemetry points for a trip
func (r *PostgresTripRepository) GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error) {
	query := `
		SELECT trip_id, lat, lng, speed, heading, fuel_level, ignition_status, recorded_at
		FROM telemetry_points
		WHERE trip_id = $1
		ORDER BY recorded_at ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, tripID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TelemetryPoint
	for rows.Next() {
		var p models.TelemetryPoint
		if err := rows.Scan(&p.TripID, &p.Latitude, &p.Longitude, &p.Speed, &p.Heading, &p.FuelLevel, &p.IgnitionStatus, &p.RecordedAt); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// GetLastCompletedTripForVehicle retrieves the most recent completed trip for a vehicle
func (r *PostgresTripRepository) GetLastCompletedTripForVehicle(ctx context.Context, vehicleID string) (*models.Trip, error) {
	query := `
		SELECT id, end_odometer, end_time
		FROM trips
		WHERE vehicle_id = $1 AND status = 'COMPLETED' AND is_deleted = FALSE
		ORDER BY end_time DESC
		LIMIT 1`

	var trip models.Trip
	start := time.Now()
	err := r.db.QueryRow(ctx, query, vehicleID).Scan(&trip.ID, &trip.EndOdometer, &trip.EndTime)
	metrics.RecordDBQuery(query, time.Since(start))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

// CheckOverlappingTrips checks if any trip exists in the given time window for a vehicle or driver
func (r *PostgresTripRepository) CheckOverlappingTrips(ctx context.Context, vehicleID, driverID string, start, end time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM trips
			WHERE (vehicle_id = $1 OR driver_id = $2)
			  AND status NOT IN ('CANCELLED', 'FAILED')
			  AND is_deleted = FALSE
			  AND (
				(scheduled_start BETWEEN $3 AND $4) OR
				(planned_eta BETWEEN $3 AND $4) OR
				($3 BETWEEN scheduled_start AND planned_eta)
			  )
		)`

	var exists bool
	queryStart := time.Now()
	err := r.db.QueryRow(ctx, query, vehicleID, driverID, start, end).Scan(&exists)
	metrics.RecordDBQuery(query, time.Since(queryStart))
	return exists, err
}

func (r *PostgresTripRepository) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	query := `
		SELECT id, org_id, vehicle_id, driver_id, status, title, description,
		       planned_distance, total_distance, estimated_duration_min,
		       actual_duration_min, scheduled_start, start_time, end_time,
		       customer_name, created_at, updated_at
		FROM trips
		WHERE vehicle_id = $1 AND status IN ('CREATED', 'STARTED', 'PAUSED', 'RESUMED') AND is_deleted = FALSE
		ORDER BY scheduled_start ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, vehicleID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*models.Trip
	for rows.Next() {
		var trip models.Trip
		err := rows.Scan(
			&trip.ID, &trip.OrgID, &trip.VehicleID, &trip.DriverID, &trip.Status,
			&trip.Title, &trip.Description, &trip.PlannedDistance, &trip.TotalDistance,
			&trip.EstimatedDurationMin, &trip.ActualDurationMin, &trip.ScheduledStart,
			&trip.StartTime, &trip.EndTime, &trip.CustomerName, &trip.CreatedAt, &trip.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, &trip)
	}

	return trips, nil
}

// GetTripSummary retrieves trip analytics for an organization
func (r *PostgresTripRepository) GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_trips,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_trips,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_trips,
			COALESCE(SUM(total_distance), 0) as total_distance_km,
			COALESCE(SUM(actual_duration_min), 0) as total_duration_minutes,
			COALESCE(AVG(total_distance / NULLIF(fuel_used, 0)), 0) as avg_fuel_efficiency,
			COALESCE(SUM(fuel_used), 0) as total_fuel_consumed,
			COALESCE(SUM(idling_time_minutes), 0) as total_idle_time_minutes
		FROM trips
		WHERE org_id = $1
		  AND created_at >= $2
		  AND created_at <= $3
		  AND status IN ('COMPLETED', 'CANCELLED')
		  AND is_deleted = FALSE`

	var summary models.TripSummary
	var totalDurationMinutes float64

	start := time.Now()
	err := r.db.QueryRow(ctx, query, orgID, startDate, endDate).Scan(
		&summary.TotalTrips, &summary.CompletedTrips, &summary.CancelledTrips,
		&summary.TotalDistance, &totalDurationMinutes, &summary.AverageFuelEfficiency,
		&summary.FuelUsed, &summary.TotalIdleTimeMinutes,
	)
	metrics.RecordDBQuery(query, time.Since(start))

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
func (r *PostgresTripRepository) GetStats(ctx context.Context) (*models.TripStats, error) {
	query := `SELECT COUNT(*) FROM trips WHERE status IN ('in_progress', 'paused') AND is_deleted = FALSE`
	var count int
	start := time.Now()
	err := r.db.QueryRow(ctx, query).Scan(&count)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &models.TripStats{ActiveTrips: count}, nil
}

// GetDriverStats aggregates metrics for all drivers in a tenant
func (r *PostgresTripRepository) GetDriverStats(ctx context.Context, orgID string) ([]models.DriverPerformance, error) {
	query := `
		SELECT
			driver_id,
			COALESCE(SUM(total_distance), 0) as total_distance,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as trips_completed,
			COUNT(*) as total_trips,
			AVG(CASE WHEN status = 'completed' THEN 100 ELSE 0 END) as on_time_rate -- Simplified for demo
		FROM trips
		WHERE org_id = $1 AND driver_id IS NOT NULL AND is_deleted = FALSE
		GROUP BY driver_id`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var performance []models.DriverPerformance
	for rows.Next() {
		var p models.DriverPerformance
		err := rows.Scan(&p.DriverID, &p.TotalDistance, &p.TripsCompleted, &p.TotalTrips, &p.OnTimeRate)
		if err != nil {
			return nil, err
		}
		p.Name = "Driver " + p.DriverID[:8]
		p.SafetyScore = 95
		performance = append(performance, p)
	}

	return performance, nil
}

// GetTripStats aggregates trip counts and distance per day for a tenant
func (r *PostgresTripRepository) GetTripStats(ctx context.Context, orgID string) ([]models.DailyTripTrend, error) {
	query := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			COUNT(*) as total_trips,
			COALESCE(SUM(total_distance), 0) as total_distance,
			COALESCE(SUM(actual_duration_min), 0) / 60.0 as active_time_hours
		FROM trips
		WHERE org_id = $1
		  AND created_at >= NOW() - INTERVAL '7 days'
		  AND is_deleted = FALSE
		GROUP BY date
		ORDER BY date ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []models.DailyTripTrend
	for rows.Next() {
		var t models.DailyTripTrend
		err := rows.Scan(&t.Date, &t.TotalTrips, &t.TotalDistance, &t.ActiveTimeHours)
		if err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}

	return trends, nil
}

// WithTx executes a function within a database transaction
func (r *PostgresTripRepository) WithTx(ctx context.Context, fn func(TripRepository) error) error {
	start := time.Now()
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	metrics.RecordDBQuery("BEGIN TRANSACTION", time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a repo that uses the transaction
	txRepo := &txRepository{tx: tx, base: r}

	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	start = time.Now()
	if err := tx.Commit(ctx); err != nil {
		metrics.RecordDBQuery("COMMIT TRANSACTION", time.Since(start))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	metrics.RecordDBQuery("COMMIT TRANSACTION", time.Since(start))

	return nil
}

// txRepository is a transaction-aware repository wrapper
type txRepository struct {
	tx   pgx.Tx
	base *PostgresTripRepository
}

func (r *txRepository) Create(ctx context.Context, trip *models.Trip) error {
	// Implementation would duplicate the base Create logic with tx.Exec
	// For now, delegate to base (which will use db pool, not tx)
	// In production, you'd implement all methods to use r.tx
	return r.base.Create(ctx, trip)
}

func (r *txRepository) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	return r.base.GetByID(ctx, id)
}

func (r *txRepository) List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error) {
	return r.base.List(ctx, filter)
}

func (r *txRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	// Use transaction
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		if key == "planned_route" || key == "waypoints" || key == "cargo_details" || key == "events" || key == "pod_documents" || key == "start_location" || key == "end_location" {
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		} else {
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE trips SET %s, updated_at = $%d WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	start := time.Now()
	_, err := r.tx.Exec(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *txRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE trips SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1"
	start := time.Now()
	_, err := r.tx.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *txRepository) Approve(ctx context.Context, id string) error {
	return r.base.Approve(ctx, id)
}

func (r *txRepository) AddEvent(ctx context.Context, tripID string, event models.TripEvent) error {
	return r.base.AddEvent(ctx, tripID, event)
}

func (r *txRepository) GetLastEvent(ctx context.Context, tripID string) (*models.TripEvent, error) {
	return r.base.GetLastEvent(ctx, tripID)
}

func (r *txRepository) AddPODDocument(ctx context.Context, tripID string, pod models.PODDocument) error {
	podJSON, _ := json.Marshal([]models.PODDocument{pod})
	query := `UPDATE trips SET pod_documents = CASE WHEN pod_documents IS NULL THEN $1::jsonb ELSE (pod_documents::jsonb || $1::jsonb) END, updated_at = NOW() WHERE id = $2`
	start := time.Now()
	_, err := r.tx.Exec(ctx, query, podJSON, tripID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *txRepository) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	return r.base.GetActiveTripsForVehicle(ctx, vehicleID)
}

func (r *txRepository) GetActiveTripsForDriver(ctx context.Context, driverID string) ([]*models.Trip, error) {
	return r.base.GetActiveTripsForDriver(ctx, driverID)
}

func (r *txRepository) GetLastCompletedTripForVehicle(ctx context.Context, vehicleID string) (*models.Trip, error) {
	return r.base.GetLastCompletedTripForVehicle(ctx, vehicleID)
}

func (r *txRepository) CheckOverlappingTrips(ctx context.Context, vehicleID, driverID string, start, end time.Time) (bool, error) {
	return r.base.CheckOverlappingTrips(ctx, vehicleID, driverID, start, end)
}

func (r *txRepository) GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	return r.base.GetTripSummary(ctx, orgID, startDate, endDate)
}

func (r *txRepository) GetStats(ctx context.Context) (*models.TripStats, error) {
	return r.base.GetStats(ctx)
}

func (r *txRepository) GetDriverStats(ctx context.Context, orgID string) ([]models.DriverPerformance, error) {
	return r.base.GetDriverStats(ctx, orgID)
}

func (r *txRepository) GetTripStats(ctx context.Context, orgID string) ([]models.DailyTripTrend, error) {
	return r.base.GetTripStats(ctx, orgID)
}

func (r *txRepository) CreateOffer(ctx context.Context, offer *models.TripOffer) error {
	return r.base.CreateOffer(ctx, offer)
}

func (r *txRepository) GetOffer(ctx context.Context, offerID string) (*models.TripOffer, error) {
	return r.base.GetOffer(ctx, offerID)
}

func (r *txRepository) RespondToOffer(ctx context.Context, offerID string, status, reason string) error {
	query := `UPDATE trip_offers SET status = $1, rejection_reason = $2, responded_at = NOW(), updated_at = NOW() WHERE id = $3`
	start := time.Now()
	_, err := r.tx.Exec(ctx, query, status, reason, offerID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *txRepository) GetOffersForDriver(ctx context.Context, driverID string) ([]*models.TripOffer, error) {
	return r.base.GetOffersForDriver(ctx, driverID)
}

func (r *txRepository) AddTelemetryPoint(ctx context.Context, point models.TelemetryPoint) error {
	return r.base.AddTelemetryPoint(ctx, point)
}

func (r *txRepository) GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error) {
	return r.base.GetTelemetry(ctx, tripID)
}

func (r *txRepository) SaveProofOfDelivery(ctx context.Context, pod *models.ProofOfDelivery) error {
	return r.base.SaveProofOfDelivery(ctx, pod)
}

func (r *txRepository) GetPODsForTrip(ctx context.Context, tripID string) ([]*models.ProofOfDelivery, error) {
	return r.base.GetPODsForTrip(ctx, tripID)
}

func (r *txRepository) WithTx(ctx context.Context, fn func(TripRepository) error) error {
	// Nested transactions not supported - just execute the function
	return fn(r)
}
