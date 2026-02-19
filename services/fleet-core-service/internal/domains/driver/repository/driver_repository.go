package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
)

type DriverRepository struct {
	db *pgxpool.Pool
}

func NewDriverRepository(db *pgxpool.Pool) *DriverRepository {
	return &DriverRepository{db: db}
}

// Create inserts a new driver into the database
func (r *DriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	query := `
		INSERT INTO drivers (
			id, org_id, first_name, last_name, full_name, email, phone,
			license_number, license_type, license_expiry, license_authority,
			date_of_birth, address, emergency_contact, status, joining_date,
			experience_years, rating, total_trips, total_distance_km,
			total_hours_driven, behavior_score, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()
	fullName := fmt.Sprintf("%s %s", driver.FirstName, driver.LastName)

	// Convert structs to JSON for storage
	addressJSON, _ := json.Marshal(driver.Address)
	emergencyContactJSON, _ := json.Marshal(driver.EmergencyContact)
	metadataJSON, _ := json.Marshal(driver.Metadata)

	err := r.db.QueryRow(ctx, query,
		id, driver.OrgID, driver.FirstName, driver.LastName, fullName, driver.Email, driver.Phone,
		driver.LicenseNumber, driver.LicenseType, driver.LicenseExpiry, driver.LicenseAuthority,
		driver.DateOfBirth, addressJSON, emergencyContactJSON, driver.Status, driver.JoiningDate,
		driver.ExperienceYears, driver.Rating, driver.TotalTrips, driver.TotalDistanceKm,
		driver.TotalHoursDriven, driver.BehaviorScore, metadataJSON, now, now,
	).Scan(&driver.ID, &driver.CreatedAt, &driver.UpdatedAt)

	return err
}

// GetByEmail retrieves a driver by email address
func (r *DriverRepository) GetByEmail(ctx context.Context, email string) (*models.Driver, error) {
	query := `SELECT id, org_id, first_name, last_name, full_name, email, status, created_at
	          FROM drivers WHERE email = $1 AND deleted_at IS NULL LIMIT 1`
	var driver models.Driver
	err := r.db.QueryRow(ctx, query, email).Scan(
		&driver.ID, &driver.OrgID, &driver.FirstName, &driver.LastName,
		&driver.FullName, &driver.Email, &driver.Status, &driver.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetByLicenseNumber retrieves a driver by license number
func (r *DriverRepository) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error) {
	query := `SELECT id, org_id, first_name, last_name, full_name, email, license_number, status, created_at
	          FROM drivers WHERE license_number = $1 AND deleted_at IS NULL LIMIT 1`
	var driver models.Driver
	err := r.db.QueryRow(ctx, query, licenseNumber).Scan(
		&driver.ID, &driver.OrgID, &driver.FirstName, &driver.LastName,
		&driver.FullName, &driver.Email, &driver.LicenseNumber, &driver.Status, &driver.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetByID retrieves a driver by its ID
func (r *DriverRepository) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	query := `
		SELECT id, org_id, first_name, last_name, full_name, email, phone,
		       license_number, license_type, license_expiry, license_authority,
		       date_of_birth, address, emergency_contact, status, current_vehicle_id,
		       current_trip_id, joining_date, resignation_date, experience_years,
		       rating, total_trips, total_distance_km, total_hours_driven,
		       behavior_score, documents, attendance_records, schedules,
		       metadata, created_at, updated_at, deleted_at
		FROM drivers WHERE id = $1 AND deleted_at IS NULL`

	var driver models.Driver
	var addressJSON, emergencyContactJSON, documentsJSON, attendanceJSON, schedulesJSON, metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&driver.ID, &driver.OrgID, &driver.FirstName, &driver.LastName, &driver.FullName,
		&driver.Email, &driver.Phone, &driver.LicenseNumber, &driver.LicenseType,
		&driver.LicenseExpiry, &driver.LicenseAuthority, &driver.DateOfBirth,
		&addressJSON, &emergencyContactJSON, &driver.Status, &driver.CurrentVehicleID,
		&driver.CurrentTripID, &driver.JoiningDate, &driver.ResignationDate,
		&driver.ExperienceYears, &driver.Rating, &driver.TotalTrips, &driver.TotalDistanceKm,
		&driver.TotalHoursDriven, &driver.BehaviorScore, &documentsJSON,
		&attendanceJSON, &schedulesJSON, &metadataJSON, &driver.CreatedAt,
		&driver.UpdatedAt, &driver.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(addressJSON, &driver.Address)
	json.Unmarshal(emergencyContactJSON, &driver.EmergencyContact)
	json.Unmarshal(documentsJSON, &driver.Documents)
	json.Unmarshal(attendanceJSON, &driver.AttendanceRecords)
	json.Unmarshal(schedulesJSON, &driver.Schedules)
	json.Unmarshal(metadataJSON, &driver.Metadata)

	return &driver, nil
}

// List retrieves drivers with filtering and pagination
func (r *DriverRepository) List(ctx context.Context, filter models.DriverListFilter) ([]*models.Driver, int, error) {
	baseQuery := `
		SELECT id, org_id, first_name, last_name, full_name, email, phone,
		       license_number, license_type, status, current_vehicle_id,
		       rating, total_trips, behavior_score, created_at, updated_at
		FROM drivers WHERE deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM drivers WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.VehicleID != "" {
		conditions = append(conditions, fmt.Sprintf("current_vehicle_id = $%d", argIndex))
		args = append(args, filter.VehicleID)
		argIndex++
	}

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d OR full_name ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Name+"%")
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
			"created_at": true, "full_name": true, "rating": true,
			"total_trips": true, "behavior_score": true, "status": true,
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

	var drivers []*models.Driver
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.OrgID, &driver.FirstName, &driver.LastName, &driver.FullName,
			&driver.Email, &driver.Phone, &driver.LicenseNumber, &driver.LicenseType,
			&driver.Status, &driver.CurrentVehicleID, &driver.Rating, &driver.TotalTrips,
			&driver.BehaviorScore, &driver.CreatedAt, &driver.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		drivers = append(drivers, &driver)
	}

	return drivers, totalCount, nil
}

// Update modifies an existing driver
func (r *DriverRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
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
		case "address", "emergency_contact", "documents", "attendance_records", "schedules", "metadata":
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE drivers SET %s, updated_at = $%d WHERE id = $%d AND deleted_at IS NULL",
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

// Delete performs soft delete of a driver
func (r *DriverRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE drivers SET deleted_at = NOW(), status = 'inactive' WHERE id = $1 AND deleted_at IS NULL"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// AssignVehicle assigns a vehicle to a driver
func (r *DriverRepository) AssignVehicle(ctx context.Context, driverID, vehicleID string) error {
	query := `
		UPDATE drivers 
		SET current_vehicle_id = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`
	
	result, err := r.db.Exec(ctx, query, vehicleID, driverID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UnassignVehicle removes vehicle assignment from a driver
func (r *DriverRepository) UnassignVehicle(ctx context.Context, driverID string) error {
	query := `
		UPDATE drivers 
		SET current_vehicle_id = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	
	result, err := r.db.Exec(ctx, query, driverID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateBehaviorScore updates driver's behavior score
func (r *DriverRepository) UpdateBehaviorScore(ctx context.Context, driverID string, scoreChange int) error {
	query := `
		UPDATE drivers 
		SET behavior_score = behavior_score + $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`
	
	result, err := r.db.Exec(ctx, query, scoreChange, driverID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// IncrementStats increments driver statistics
func (r *DriverRepository) IncrementStats(ctx context.Context, driverID string, distanceKm, hoursDriven float64) error {
	query := `
		UPDATE drivers 
		SET total_trips = total_trips + 1,
		    total_distance_km = total_distance_km + $1,
		    total_hours_driven = total_hours_driven + $2,
		    updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL`
	
	result, err := r.db.Exec(ctx, query, distanceKm, hoursDriven, driverID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetDriverAnalytics retrieves driver analytics
func (r *DriverRepository) GetDriverAnalytics(ctx context.Context, driverID string) (*models.DriverAnalytics, error) {
	query := `
		SELECT 
			d.total_trips,
			d.total_distance_km,
			d.total_hours_driven,
			d.rating as average_rating,
			d.behavior_score as current_behavior_score,
			COALESCE(COUNT(t.id) FILTER (WHERE t.status = 'completed'), 0) as completed_trips,
			COALESCE(COUNT(t.id) FILTER (WHERE t.status = 'cancelled'), 0) as cancelled_trips,
			COALESCE(AVG(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) * 100, 0) as attendance_rate
		FROM drivers d
		LEFT JOIN trips t ON d.id = t.driver_id AND t.created_at >= CURRENT_DATE - INTERVAL '30 days'
		LEFT JOIN attendance_records a ON d.id = a.driver_id AND a.date >= CURRENT_DATE - INTERVAL '30 days'
		WHERE d.id = $1 AND d.deleted_at IS NULL
		GROUP BY d.total_trips, d.total_distance_km, d.total_hours_driven, d.rating, d.behavior_score`

	var analytics models.DriverAnalytics
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&analytics.TotalTrips, &analytics.TotalDistanceKm, &analytics.TotalHoursDriven,
		&analytics.AverageRating, &analytics.CurrentBehaviorScore, &analytics.CompletedTrips,
		&analytics.CancelledTrips, &analytics.AttendanceRate,
	)

	if err != nil {
		return nil, err
	}

	// Calculate derived metrics
	if analytics.TotalTrips > 0 {
		analytics.CompletionRate = float64(analytics.CompletedTrips) / float64(analytics.TotalTrips) * 100
		analytics.OnTimeRate = analytics.AttendanceRate // Simplified for now
	}
	
	if analytics.TotalHoursDriven > 0 {
		analytics.FuelEfficiency = analytics.TotalDistanceKm / analytics.TotalHoursDriven // Simplified
	}

	return &analytics, nil
}

// GetActiveDriversForVehicle retrieves active drivers for a vehicle
func (r *DriverRepository) GetActiveDriversForVehicle(ctx context.Context, vehicleID string) ([]*models.Driver, error) {
	query := `
		SELECT id, org_id, first_name, last_name, full_name, email, phone,
		       license_number, license_type, status, rating, behavior_score,
		       created_at, updated_at
		FROM drivers 
		WHERE current_vehicle_id = $1 
		  AND status = 'active' 
		  AND deleted_at IS NULL
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []*models.Driver
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.OrgID, &driver.FirstName, &driver.LastName, &driver.FullName,
			&driver.Email, &driver.Phone, &driver.LicenseNumber, &driver.LicenseType,
			&driver.Status, &driver.Rating, &driver.BehaviorScore, &driver.CreatedAt, &driver.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		drivers = append(drivers, &driver)
	}

	return drivers, nil
}

// CreateBehaviorRecord creates a new driver behavior record
func (r *DriverRepository) CreateBehaviorRecord(ctx context.Context, behavior *models.DriverBehavior) error {
	query := `
		INSERT INTO driver_behavior (
			id, driver_id, org_id, period_start, period_end, score, safety_score,
			efficiency_score, compliance_score, harsh_accelerations, harsh_brakings,
			harsh_cornerings, speeding_incidents, phone_usage, seatbelt_violations,
			fuel_consumption, distance_driven, idle_time, eco_driving_score,
			hours_of_service, rest_periods, document_compliance, maintenance_adherence,
			night_driving_hours, highway_percentage, city_driving_percentage,
			risk_level, risk_factors, improvement_areas, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert slices and maps to JSON
	riskFactorsJSON, _ := json.Marshal(behavior.RiskFactors)
	improvementAreasJSON, _ := json.Marshal(behavior.ImprovementAreas)
	metadataJSON, _ := json.Marshal(behavior.Metadata)

	err := r.db.QueryRow(ctx, query,
		id, behavior.DriverID, behavior.OrgID, behavior.PeriodStart, behavior.PeriodEnd,
		behavior.Score, behavior.SafetyScore, behavior.EfficiencyScore, behavior.ComplianceScore,
		behavior.HarshAccelerations, behavior.HarshBrakings, behavior.HarshCornerings,
		behavior.SpeedingIncidents, behavior.PhoneUsage, behavior.SeatbeltViolations,
		behavior.FuelConsumption, behavior.DistanceDriven, behavior.IdleTime, behavior.EcoDrivingScore,
		behavior.HoursOfService, behavior.RestPeriods, behavior.DocumentCompliance, behavior.MaintenanceAdherence,
		behavior.NightDrivingHours, behavior.HighwayPercentage, behavior.CityDrivingPercentage,
		behavior.RiskLevel, riskFactorsJSON, improvementAreasJSON, metadataJSON, now, now,
	).Scan(&behavior.ID, &behavior.CreatedAt, &behavior.UpdatedAt)

	return err
}

// GetBehaviorRecords retrieves behavior records for a driver
func (r *DriverRepository) GetBehaviorRecords(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DriverBehavior, error) {
	query := `
		SELECT id, driver_id, org_id, period_start, period_end, score, safety_score,
			efficiency_score, compliance_score, harsh_accelerations, harsh_brakings,
			harsh_cornerings, speeding_incidents, phone_usage, seatbelt_violations,
			fuel_consumption, distance_driven, idle_time, eco_driving_score,
			hours_of_service, rest_periods, document_compliance, maintenance_adherence,
			night_driving_hours, highway_percentage, city_driving_percentage,
			risk_level, risk_factors, improvement_areas, metadata, created_at, updated_at
		FROM driver_behavior 
		WHERE driver_id = $1 
			AND period_start >= $2 
			AND period_end <= $3
		ORDER BY period_start DESC`

	rows, err := r.db.Query(ctx, query, driverID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var behaviors []*models.DriverBehavior
	for rows.Next() {
		var behavior models.DriverBehavior
		var riskFactorsJSON, improvementAreasJSON, metadataJSON []byte

		err := rows.Scan(
			&behavior.ID, &behavior.DriverID, &behavior.OrgID, &behavior.PeriodStart, &behavior.PeriodEnd,
			&behavior.Score, &behavior.SafetyScore, &behavior.EfficiencyScore, &behavior.ComplianceScore,
			&behavior.HarshAccelerations, &behavior.HarshBrakings, &behavior.HarshCornerings,
			&behavior.SpeedingIncidents, &behavior.PhoneUsage, &behavior.SeatbeltViolations,
			&behavior.FuelConsumption, &behavior.DistanceDriven, &behavior.IdleTime, &behavior.EcoDrivingScore,
			&behavior.HoursOfService, &behavior.RestPeriods, &behavior.DocumentCompliance, &behavior.MaintenanceAdherence,
			&behavior.NightDrivingHours, &behavior.HighwayPercentage, &behavior.CityDrivingPercentage,
			&behavior.RiskLevel, &riskFactorsJSON, &improvementAreasJSON, &metadataJSON,
			&behavior.CreatedAt, &behavior.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(riskFactorsJSON, &behavior.RiskFactors)
		json.Unmarshal(improvementAreasJSON, &behavior.ImprovementAreas)
		json.Unmarshal(metadataJSON, &behavior.Metadata)

		behaviors = append(behaviors, &behavior)
	}

	return behaviors, nil
}

// CreateIncident creates a new driver incident
func (r *DriverRepository) CreateIncident(ctx context.Context, incident *models.DriverIncident) error {
	query := `
		INSERT INTO driver_incidents (
			id, driver_id, org_id, type, severity, timestamp, location,
			speed, details, resolved, resolution_notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON
	locationJSON, _ := json.Marshal(incident.Location)
	detailsJSON, _ := json.Marshal(incident.Details)

	err := r.db.QueryRow(ctx, query,
		id, incident.DriverID, incident.OrgID, incident.Type, incident.Severity,
		incident.Timestamp, locationJSON, incident.Speed, detailsJSON,
		incident.Resolved, incident.ResolutionNotes, now,
	).Scan(&incident.ID, &incident.CreatedAt)

	return err
}

// GetIncidents retrieves incidents for a driver
func (r *DriverRepository) GetIncidents(ctx context.Context, driverID string, filter models.BehaviorReportRequest) ([]*models.DriverIncident, error) {
	baseQuery := `
		SELECT id, driver_id, org_id, type, severity, timestamp, location,
			speed, details, resolved, resolution_notes, created_at
		FROM driver_incidents WHERE driver_id = $1`

	var conditions []string
	var args []interface{}
	args = append(args, driverID)
	argIndex := 2

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	query := baseQuery + whereClause + " ORDER BY timestamp DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*models.DriverIncident
	for rows.Next() {
		var incident models.DriverIncident
		var locationJSON, detailsJSON []byte

		err := rows.Scan(
			&incident.ID, &incident.DriverID, &incident.OrgID, &incident.Type,
			&incident.Severity, &incident.Timestamp, &locationJSON, &incident.Speed,
			&detailsJSON, &incident.Resolved, &incident.ResolutionNotes, &incident.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(locationJSON, &incident.Location)
		json.Unmarshal(detailsJSON, &incident.Details)

		incidents = append(incidents, &incident)
	}

	return incidents, nil
}

// UpdateIncidentResolution updates incident resolution status
func (r *DriverRepository) UpdateIncidentResolution(ctx context.Context, incidentID string, resolved bool, notes string) error {
	query := `UPDATE driver_incidents SET resolved = $1, resolution_notes = $2, updated_at = NOW() WHERE id = $3`
	result, err := r.db.Exec(ctx, query, resolved, notes, incidentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// CreateBehaviorScoreUpdate records a behavior score change
func (r *DriverRepository) CreateBehaviorScoreUpdate(ctx context.Context, update *models.BehaviorScoreUpdate) error {
	query := `
		INSERT INTO behavior_score_updates (
			id, driver_id, score_change, reason, event_type, details, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	id := uuid.New().String()
	now := time.Now()
	detailsJSON, _ := json.Marshal(update.Details)

	err := r.db.QueryRow(ctx, query,
		id, update.DriverID, update.ScoreChange, update.Reason,
		update.EventType, detailsJSON, update.CreatedBy, now,
	).Scan(&update.ID, &update.CreatedAt)

	return err
}

// GetBehaviorSummary retrieves behavior summary for a driver
func (r *DriverRepository) GetBehaviorSummary(ctx context.Context, driverID string) (*models.DriverBehavior, error) {
	query := `
		SELECT 
			MAX(score) as current_score,
			MAX(safety_score) as current_safety_score,
			MAX(efficiency_score) as current_efficiency_score,
			MAX(compliance_score) as current_compliance_score,
			SUM(harsh_accelerations) as total_harsh_accelerations,
			SUM(harsh_brakings) as total_harsh_brakings,
			SUM(speeding_incidents) as total_speeding_incidents,
			AVG(fuel_consumption) as avg_fuel_consumption,
			SUM(distance_driven) as total_distance_driven,
			MAX(risk_level) as current_risk_level
		FROM driver_behavior 
		WHERE driver_id = $1 
			AND period_end >= CURRENT_DATE - INTERVAL '30 days'`

	var summary models.DriverBehavior
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&summary.Score, &summary.SafetyScore, &summary.EfficiencyScore, &summary.ComplianceScore,
		&summary.HarshAccelerations, &summary.HarshBrakings, &summary.SpeedingIncidents,
		&summary.FuelConsumption, &summary.DistanceDriven, &summary.RiskLevel,
	)

	if err != nil {
		return nil, err
	}

	summary.DriverID = driverID
	return &summary, nil
}
