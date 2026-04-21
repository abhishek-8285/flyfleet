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

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
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
			id, tenant_id, first_name, last_name, email, phone,
			employee_code, aadhaar_number, pan_number,
			license_number, license_type, license_issue_date, license_expiry, license_authority,
			date_of_birth, address, emergency_contact, status, employment_type, joining_date,
			experience_years, rating, total_trips, total_distance_km,
			total_hours_driven, behavior_score, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	addressJSON, err := json.Marshal(driver.Address)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}
	emergencyContactJSON, err := json.Marshal(driver.EmergencyContact)
	if err != nil {
		return fmt.Errorf("failed to marshal emergency contact: %w", err)
	}
	metadataJSON, err := json.Marshal(driver.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	start := time.Now()
	err = r.db.QueryRow(ctx, query,
		id, driver.TenantID, driver.FirstName, driver.LastName, driver.Email, driver.Phone,
		driver.EmployeeCode, driver.AadhaarNumber, driver.PanNumber,
		driver.LicenseNumber, driver.LicenseType, driver.LicenseIssueDate, driver.LicenseExpiry, driver.LicenseAuthority,
		driver.DateOfBirth, addressJSON, emergencyContactJSON, driver.Status, driver.EmploymentType, driver.JoiningDate,
		driver.ExperienceYears, driver.Rating, driver.TotalTrips, driver.TotalDistanceKm,
		driver.TotalHoursDriven, driver.BehaviorScore, metadataJSON, now, now,
	).Scan(&driver.ID, &driver.CreatedAt, &driver.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	return nil
}

// GetByEmail retrieves a driver by email address
func (r *DriverRepository) GetByEmail(ctx context.Context, email string) (*models.Driver, error) {
	query := `SELECT id, tenant_id, first_name, last_name, full_name, email, status, created_at
	          FROM drivers WHERE email = $1 AND deleted_at IS NULL LIMIT 1`
	var driver models.Driver
	start := time.Now()
	err := r.db.QueryRow(ctx, query, email).Scan(
		&driver.ID, &driver.TenantID, &driver.FirstName, &driver.LastName,
		&driver.FullName, &driver.Email, &driver.Status, &driver.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver by email: %w", err)
	}
	return &driver, nil
}

// GetByLicenseNumber retrieves a driver by license number
func (r *DriverRepository) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error) {
	query := `SELECT id, tenant_id, first_name, last_name, full_name, email, license_number, status, created_at
	          FROM drivers WHERE license_number = $1 AND deleted_at IS NULL LIMIT 1`
	var driver models.Driver
	start := time.Now()
	err := r.db.QueryRow(ctx, query, licenseNumber).Scan(
		&driver.ID, &driver.TenantID, &driver.FirstName, &driver.LastName,
		&driver.FullName, &driver.Email, &driver.LicenseNumber, &driver.Status, &driver.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver by license number: %w", err)
	}
	return &driver, nil
}

// GetByID retrieves a driver by its ID
func (r *DriverRepository) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	query := `
		SELECT id, tenant_id, first_name, last_name, full_name, email, phone,
		       employee_code, aadhaar_number, pan_number,
		       license_number, license_type, license_issue_date, license_expiry, license_authority,
		       date_of_birth, address, emergency_contact, status, employment_type, current_vehicle_id,
		       current_trip_id, joining_date, resignation_date, experience_years,
		       rating, total_trips, total_distance_km, total_hours_driven,
		       behavior_score, documents, metadata, created_at, updated_at, deleted_at
		FROM drivers WHERE id = $1 AND deleted_at IS NULL`

	var driver models.Driver
	var addressJSON, emergencyContactJSON, documentsJSON, metadataJSON []byte

	start := time.Now()
	err := r.db.QueryRow(ctx, query, id).Scan(
		&driver.ID, &driver.TenantID, &driver.FirstName, &driver.LastName, &driver.FullName,
		&driver.Email, &driver.Phone, &driver.EmployeeCode, &driver.AadhaarNumber,
		&driver.PanNumber, &driver.LicenseNumber, &driver.LicenseType,
		&driver.LicenseIssueDate, &driver.LicenseExpiry, &driver.LicenseAuthority,
		&driver.DateOfBirth, &addressJSON, &emergencyContactJSON, &driver.Status, &driver.EmploymentType,
		&driver.CurrentVehicleID, &driver.CurrentTripID, &driver.JoiningDate, &driver.ResignationDate,
		&driver.ExperienceYears, &driver.Rating, &driver.TotalTrips, &driver.TotalDistanceKm,
		&driver.TotalHoursDriven, &driver.BehaviorScore, &documentsJSON,
		&metadataJSON, &driver.CreatedAt, &driver.UpdatedAt, &driver.DeletedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver by ID: %w", err)
	}

	// Parse JSON fields
	if len(addressJSON) > 0 {
		if err := json.Unmarshal(addressJSON, &driver.Address); err != nil {
			return nil, fmt.Errorf("failed to unmarshal address: %w", err)
		}
	}
	if len(emergencyContactJSON) > 0 {
		if err := json.Unmarshal(emergencyContactJSON, &driver.EmergencyContact); err != nil {
			return nil, fmt.Errorf("failed to unmarshal emergency contact: %w", err)
		}
	}
	if len(documentsJSON) > 0 {
		if err := json.Unmarshal(documentsJSON, &driver.Documents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal documents: %w", err)
		}
	}
	// Fetch attendance and schedules from normalized tables
	attendance, err := r.GetAttendance(ctx, driver.ID)
	if err != nil {
		// Log error but don't fail - return empty attendance
		attendance = []models.AttendanceRecord{}
	}
	driver.AttendanceRecords = attendance

	schedules, err := r.GetSchedules(ctx, driver.ID)
	if err != nil {
		schedules = []models.WorkSchedule{}
	}
	driver.Schedules = schedules

	return &driver, nil
}

// GetAttendance retrieves all attendance records for a driver from the normalized table
func (r *DriverRepository) GetAttendance(ctx context.Context, driverID string) ([]models.AttendanceRecord, error) {
	query := `
		SELECT id, driver_id, tenant_id, attendance_date, status, check_in_time, check_out_time, remarks, created_at
		FROM driver_attendance
		WHERE driver_id = $1
		ORDER BY attendance_date DESC
	`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query attendance: %w", err)
	}
	defer rows.Close()

	var records []models.AttendanceRecord
	for rows.Next() {
		var rec models.AttendanceRecord
		var id uuid.UUID
		var driverID uuid.UUID // dummy, not used
		var tenantID uuid.UUID   // org context, not stored in record
		var date time.Time
		err := rows.Scan(
			&id, &driverID, &tenantID, &date, &rec.Status, &rec.CheckInTime, &rec.CheckOutTime, &rec.Remarks, &rec.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}

// CreateAttendance creates a new attendance record for a driver
func (r *DriverRepository) CreateAttendance(ctx context.Context, attendance *models.DriverAttendance) error {
	query := `
		INSERT INTO driver_attendance (driver_id, tenant_id, attendance_date, status, check_in_time, check_out_time, remarks, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.Exec(ctx, query, attendance.DriverID, attendance.TenantID, attendance.AttendanceDate, attendance.Status, attendance.CheckInTime, attendance.CheckOutTime, attendance.Remarks)
	return err
}

// GetAttendanceByID retrieves an attendance record by ID
func (r *DriverRepository) GetAttendanceByID(ctx context.Context, id uuid.UUID) (*models.DriverAttendance, error) {
	query := `
		SELECT id, driver_id, tenant_id, attendance_date, status, check_in_time, check_out_time, remarks, created_at
		FROM driver_attendance
		WHERE id = $1
	`
	var attendance models.DriverAttendance
	err := r.db.QueryRow(ctx, query, id).Scan(
		&attendance.ID, &attendance.DriverID, &attendance.TenantID, &attendance.AttendanceDate,
		&attendance.Status, &attendance.CheckInTime, &attendance.CheckOutTime, &attendance.Remarks, &attendance.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

// GetAttendanceByDriver retrieves attendance records for a driver within a date range
func (r *DriverRepository) GetAttendanceByDriver(ctx context.Context, driverID, tenantID uuid.UUID, from, to time.Time) ([]*models.DriverAttendance, error) {
	query := `
		SELECT id, driver_id, tenant_id, attendance_date, status, check_in_time, check_out_time, remarks, created_at
		FROM driver_attendance
		WHERE driver_id = $1 AND tenant_id = $2 AND attendance_date BETWEEN $3 AND $4
		ORDER BY attendance_date DESC
	`
	rows, err := r.db.Query(ctx, query, driverID, tenantID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []*models.DriverAttendance
	for rows.Next() {
		var attendance models.DriverAttendance
		err := rows.Scan(
			&attendance.ID, &attendance.DriverID, &attendance.TenantID, &attendance.AttendanceDate,
			&attendance.Status, &attendance.CheckInTime, &attendance.CheckOutTime, &attendance.Remarks, &attendance.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attendances = append(attendances, &attendance)
	}
	return attendances, nil
}

// UpdateCheckOut updates the check-out time for an attendance record
func (r *DriverRepository) UpdateCheckOut(ctx context.Context, id, tenantID uuid.UUID, checkOut time.Time, remarks string) error {
	query := `
		UPDATE driver_attendance
		SET check_out_time = $1, remarks = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
	`
	_, err := r.db.Exec(ctx, query, checkOut, remarks, id, tenantID)
	return err
}

// CreateSchedule creates a new work schedule for a driver
func (r *DriverRepository) CreateSchedule(ctx context.Context, schedule *models.DriverSchedule) error {
	query := `
		INSERT INTO driver_schedules (driver_id, tenant_id, day_of_week, start_time, end_time, is_working, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, schedule.DriverID, schedule.TenantID, schedule.DayOfWeek, schedule.StartTime, schedule.EndTime, schedule.IsWorking)
	return err
}

// GetScheduleByID retrieves a work schedule by ID
func (r *DriverRepository) GetScheduleByID(ctx context.Context, id uuid.UUID) (*models.DriverSchedule, error) {
	query := `
		SELECT id, driver_id, tenant_id, day_of_week, start_time, end_time, is_working, created_at, updated_at
		FROM driver_schedules
		WHERE id = $1
	`
	var schedule models.DriverSchedule
	err := r.db.QueryRow(ctx, query, id).Scan(
		&schedule.ID, &schedule.DriverID, &schedule.TenantID, &schedule.DayOfWeek,
		&schedule.StartTime, &schedule.EndTime, &schedule.IsWorking, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// GetSchedulesByDriver retrieves all work schedules for a driver
func (r *DriverRepository) GetSchedulesByDriver(ctx context.Context, driverID, tenantID uuid.UUID) ([]*models.DriverSchedule, error) {
	query := `
		SELECT id, driver_id, tenant_id, day_of_week, start_time, end_time, is_working, created_at, updated_at
		FROM driver_schedules
		WHERE driver_id = $1 AND tenant_id = $2
		ORDER BY day_of_week
	`
	rows, err := r.db.Query(ctx, query, driverID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.DriverSchedule
	for rows.Next() {
		var schedule models.DriverSchedule
		err := rows.Scan(
			&schedule.ID, &schedule.DriverID, &schedule.TenantID, &schedule.DayOfWeek,
			&schedule.StartTime, &schedule.EndTime, &schedule.IsWorking, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, &schedule)
	}
	return schedules, nil
}

// UpdateSchedule updates a work schedule
func (r *DriverRepository) UpdateSchedule(ctx context.Context, schedule *models.DriverSchedule) error {
	query := `
		UPDATE driver_schedules
		SET day_of_week = $1, start_time = $2, end_time = $3, is_working = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, schedule.DayOfWeek, schedule.StartTime, schedule.EndTime, schedule.IsWorking, schedule.ID)
	return err
}

// DeleteSchedule deletes a work schedule
func (r *DriverRepository) DeleteSchedule(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `DELETE FROM driver_schedules WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

// GetSchedules retrieves all work schedules for a driver from the normalized table
func (r *DriverRepository) GetSchedules(ctx context.Context, driverID string) ([]models.WorkSchedule, error) {
	query := `
		SELECT id, driver_id, tenant_id, day_of_week, start_time, end_time, is_working, created_at, updated_at
		FROM driver_schedules
		WHERE driver_id = $1
		ORDER BY day_of_week
	`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []models.WorkSchedule
	for rows.Next() {
		var ds models.DriverSchedule
		var tenantID string
		err := rows.Scan(
			&ds.ID, &ds.DriverID, &tenantID, &ds.DayOfWeek, &ds.StartTime, &ds.EndTime,
			&ds.IsWorking, &ds.CreatedAt, &ds.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan work schedule: %w", err)
		}
		// Map DriverSchedule to WorkSchedule (Driver model field)
		schedule := models.WorkSchedule{
			ID:        ds.ID.String(),
			DayOfWeek: ds.DayOfWeek,
			StartTime: ds.StartTime,
			EndTime:   ds.EndTime,
			IsWorking: ds.IsWorking,
			CreatedAt: ds.CreatedAt,
			UpdatedAt: ds.UpdatedAt,
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("schedules iteration error: %w", err)
	}
	return schedules, nil
}

// List retrieves drivers with filtering and pagination
func (r *DriverRepository) List(ctx context.Context, filter models.DriverListFilter) ([]*models.Driver, int, error) {
	baseQuery := `
		SELECT id, tenant_id, first_name, last_name, full_name, email, phone,
		       license_number, license_type, status, current_vehicle_id,
		       rating, total_trips, behavior_score, created_at, updated_at
		FROM drivers WHERE deleted_at IS NULL`

	countQuery := `SELECT COUNT(*) FROM drivers WHERE deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

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
	start := time.Now()
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	metrics.RecordDBQuery(countSQL, time.Since(start))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count drivers: %w", err)
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

	start = time.Now()
	rows, err := r.db.Query(ctx, listSQL, args...)
	metrics.RecordDBQuery(listSQL, time.Since(start))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list drivers: %w", err)
	}
	defer rows.Close()

	var drivers []*models.Driver
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.TenantID, &driver.FirstName, &driver.LastName, &driver.FullName,
			&driver.Email, &driver.Phone, &driver.LicenseNumber, &driver.LicenseType,
			&driver.Status, &driver.CurrentVehicleID, &driver.Rating, &driver.TotalTrips,
			&driver.BehaviorScore, &driver.CreatedAt, &driver.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan driver: %w", err)
		}
		drivers = append(drivers, &driver)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("driver iteration error: %w", err)
	}

	return drivers, totalCount, nil
}

// Update modifies an existing driver
func (r *DriverRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Handle normalized table updates separately
	if attendanceVal, hasAttendance := updates["attendance_records"]; hasAttendance {
		// Delete existing attendance records for this driver and insert new ones
		if err := r.replaceAttendance(ctx, id, attendanceVal); err != nil {
			return fmt.Errorf("failed to update attendance records: %w", err)
		}
		delete(updates, "attendance_records")
	}

	if schedulesVal, hasSchedules := updates["schedules"]; hasSchedules {
		// Delete existing schedules for this driver and insert new ones
		if err := r.replaceSchedules(ctx, id, schedulesVal); err != nil {
			return fmt.Errorf("failed to update schedules: %w", err)
		}
		delete(updates, "schedules")
	}

	// If no remaining updates after removing special fields, return early
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))

		// Handle special JSON fields
		switch key {
		case "address", "emergency_contact", "documents", "metadata":
			jsonData, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", key, err)
			}
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE drivers SET %s, updated_at = $%d WHERE id = $%d AND deleted_at IS NULL",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	start := time.Now()
	result, err := r.db.Exec(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update driver: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// replaceAttendance deletes and recreates driver attendance records from a slice

// replaceAttendance deletes and recreates driver attendance records from a slice
func (r *DriverRepository) replaceAttendance(ctx context.Context, driverID string, records interface{}) error {
	// Get driver's tenant_id first (needed for attendance records)
	var tenantID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT tenant_id FROM drivers WHERE id = $1`, driverID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("failed to get driver tenant_id: %w", err)
	}

	// Delete existing attendance records for this driver
	_, err = r.db.Exec(ctx, `DELETE FROM driver_attendance WHERE driver_id = $1`, driverID)
	if err != nil {
		return fmt.Errorf("failed to delete existing attendance: %w", err)
	}

	// Cast records to expected type
	attendanceList, ok := records.([]models.AttendanceRecord)
	if !ok {
		return fmt.Errorf("invalid attendance records type")
	}

	if len(attendanceList) == 0 {
		return nil
	}

	// Build batch insert (9 columns)
	var args []interface{}
	var placeholders []string
	for i, rec := range attendanceList {
		id := uuid.New()
		if rec.ID != "" {
			id = uuid.MustParse(rec.ID)
		}
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
		args = append(args,
			id,
			uuid.MustParse(driverID),
			tenantID,
			rec.Date,
			rec.Status,
			rec.CheckInTime,
			rec.CheckOutTime,
			rec.Remarks,
			time.Now(),
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO driver_attendance (id, driver_id, tenant_id, attendance_date, status, check_in_time, check_out_time, remarks, created_at) VALUES %s`,
		strings.Join(placeholders, ", "),
	)

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert attendance records: %w", err)
	}
	return nil
}

// replaceSchedules deletes and recreates driver work schedules from a slice
func (r *DriverRepository) replaceSchedules(ctx context.Context, driverID string, schedules interface{}) error {
	// Get driver's tenant_id
	var tenantID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT tenant_id FROM drivers WHERE id = $1`, driverID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("failed to get driver tenant_id: %w", err)
	}

	// Delete existing schedules
	_, err = r.db.Exec(ctx, `DELETE FROM driver_schedules WHERE driver_id = $1`, driverID)
	if err != nil {
		return fmt.Errorf("failed to delete existing schedules: %w", err)
	}

	// Cast schedules
	schedList, ok := schedules.([]models.WorkSchedule)
	if !ok {
		return fmt.Errorf("invalid schedules type")
	}

	if len(schedList) == 0 {
		return nil
	}

	// Build batch insert (9 columns)
	var args []interface{}
	var placeholders []string
	for i, s := range schedList {
		id := uuid.New()
		if s.ID != "" {
			id = uuid.MustParse(s.ID)
		}
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
		args = append(args,
			id,
			uuid.MustParse(driverID),
			tenantID,
			s.DayOfWeek,
			s.StartTime,
			s.EndTime,
			s.IsWorking,
			time.Now(),
			time.Now(),
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO driver_schedules (id, driver_id, tenant_id, day_of_week, start_time, end_time, is_working, created_at, updated_at) VALUES %s`,
		strings.Join(placeholders, ", "),
	)

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert schedules: %w", err)
	}
	return nil
}

// Delete performs soft delete of a driver
func (r *DriverRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE drivers SET deleted_at = NOW(), status = 'inactive' WHERE id = $1 AND deleted_at IS NULL"
	start := time.Now()
	result, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to delete driver: %w", err)
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

	start := time.Now()
	result, err := r.db.Exec(ctx, query, vehicleID, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to assign vehicle: %w", err)
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

	start := time.Now()
	result, err := r.db.Exec(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to unassign vehicle: %w", err)
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

	start := time.Now()
	result, err := r.db.Exec(ctx, query, scoreChange, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update behavior score: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// LogBehaviorScoreChange logs a behavior score change event
func (r *DriverRepository) LogBehaviorScoreChange(ctx context.Context, driverID, tenantID uuid.UUID, oldScore, newScore int, reason, eventType string, userID uuid.UUID, userEmail string) error {
	query := `
		INSERT INTO driver_behavior_logs (
			driver_id, tenant_id, old_score, new_score, reason, event_type,
			user_id, user_email, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	_, err := r.db.Exec(ctx, query, driverID, tenantID, oldScore, newScore, reason, eventType, userID, userEmail)
	return err
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

	start := time.Now()
	result, err := r.db.Exec(ctx, query, distanceKm, hoursDriven, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to increment driver stats: %w", err)
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
			COALESCE((
				SELECT AVG(CASE WHEN da.status = 'present' THEN 1 ELSE 0 END) * 100
				FROM driver_attendance da
				WHERE da.driver_id = d.id
				  AND da.attendance_date >= CURRENT_DATE - INTERVAL '30 days'
			), 0) as attendance_rate
		FROM drivers d
		LEFT JOIN trips t ON d.id = t.driver_id AND t.created_at >= CURRENT_DATE - INTERVAL '30 days'
		WHERE d.id = $1 AND d.deleted_at IS NULL
		GROUP BY d.id, d.total_trips, d.total_distance_km, d.total_hours_driven, d.rating, d.behavior_score`

	var analytics models.DriverAnalytics
	start := time.Now()
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&analytics.TotalTrips, &analytics.TotalDistanceKm, &analytics.TotalHoursDriven,
		&analytics.AverageRating, &analytics.CurrentBehaviorScore, &analytics.CompletedTrips,
		&analytics.CancelledTrips, &analytics.AttendanceRate,
	)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver analytics: %w", err)
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
		SELECT id, tenant_id, first_name, last_name, full_name, email, phone,
		       license_number, license_type, status, rating, behavior_score,
		       created_at, updated_at
		FROM drivers
		WHERE current_vehicle_id = $1
		  AND status = 'active'
		  AND deleted_at IS NULL
		ORDER BY created_at ASC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, vehicleID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query active drivers: %w", err)
	}
	defer rows.Close()

	var drivers []*models.Driver
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.TenantID, &driver.FirstName, &driver.LastName, &driver.FullName,
			&driver.Email, &driver.Phone, &driver.LicenseNumber, &driver.LicenseType,
			&driver.Status, &driver.Rating, &driver.BehaviorScore, &driver.CreatedAt, &driver.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		drivers = append(drivers, &driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("driver iteration error: %w", err)
	}

	return drivers, nil
}

// CreateBehaviorRecord creates a new driver behavior record
func (r *DriverRepository) CreateBehaviorRecord(ctx context.Context, behavior *models.DriverBehavior) error {
	query := `
		INSERT INTO driver_behavior (
			id, driver_id, tenant_id, period_start, period_end, score, safety_score,
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
	riskFactorsJSON, err := json.Marshal(behavior.RiskFactors)
	if err != nil {
		return fmt.Errorf("failed to marshal risk factors: %w", err)
	}
	improvementAreasJSON, err := json.Marshal(behavior.ImprovementAreas)
	if err != nil {
		return fmt.Errorf("failed to marshal improvement areas: %w", err)
	}
	metadataJSON, err := json.Marshal(behavior.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	start := time.Now()
	err = r.db.QueryRow(ctx, query,
		id, behavior.DriverID, behavior.TenantID, behavior.PeriodStart, behavior.PeriodEnd,
		behavior.Score, behavior.SafetyScore, behavior.EfficiencyScore, behavior.ComplianceScore,
		behavior.HarshAccelerations, behavior.HarshBrakings, behavior.HarshCornerings,
		behavior.SpeedingIncidents, behavior.PhoneUsage, behavior.SeatbeltViolations,
		behavior.FuelConsumption, behavior.DistanceDriven, behavior.IdleTime, behavior.EcoDrivingScore,
		behavior.HoursOfService, behavior.RestPeriods, behavior.DocumentCompliance, behavior.MaintenanceAdherence,
		behavior.NightDrivingHours, behavior.HighwayPercentage, behavior.CityDrivingPercentage,
		behavior.RiskLevel, riskFactorsJSON, improvementAreasJSON, metadataJSON, now, now,
	).Scan(&behavior.ID, &behavior.CreatedAt, &behavior.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		return fmt.Errorf("failed to create behavior record: %w", err)
	}

	return nil
}

// GetBehaviorRecords retrieves behavior records for a driver
func (r *DriverRepository) GetBehaviorRecords(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DriverBehavior, error) {
	query := `
		SELECT id, driver_id, tenant_id, period_start, period_end, score, safety_score,
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

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID, startDate, endDate)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query behavior records: %w", err)
	}
	defer rows.Close()

	var behaviors []*models.DriverBehavior
	for rows.Next() {
		var behavior models.DriverBehavior
		var riskFactorsJSON, improvementAreasJSON, metadataJSON []byte

		err := rows.Scan(
			&behavior.ID, &behavior.DriverID, &behavior.TenantID, &behavior.PeriodStart, &behavior.PeriodEnd,
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
			return nil, fmt.Errorf("failed to scan behavior record: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(riskFactorsJSON, &behavior.RiskFactors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal risk factors: %w", err)
		}
		if err := json.Unmarshal(improvementAreasJSON, &behavior.ImprovementAreas); err != nil {
			return nil, fmt.Errorf("failed to unmarshal improvement areas: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &behavior.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		behaviors = append(behaviors, &behavior)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("behavior record iteration error: %w", err)
	}

	return behaviors, nil
}

// CreateIncident creates a new driver incident
func (r *DriverRepository) CreateIncident(ctx context.Context, incident *models.DriverIncident) error {
	query := `
		INSERT INTO driver_incidents (
			id, driver_id, tenant_id, type, severity, timestamp, location,
			speed, details, resolved, resolution_notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON
	locationJSON, err := json.Marshal(incident.Location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}
	detailsJSON, err := json.Marshal(incident.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	start := time.Now()
	err = r.db.QueryRow(ctx, query,
		id, incident.DriverID, incident.TenantID, incident.Type, incident.Severity,
		incident.Timestamp, locationJSON, incident.Speed, detailsJSON,
		incident.Resolved, incident.ResolutionNotes, now,
	).Scan(&incident.ID, &incident.CreatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// GetIncidents retrieves incidents for a driver
func (r *DriverRepository) GetIncidents(ctx context.Context, driverID string, filter models.BehaviorReportRequest) ([]*models.DriverIncident, error) {
	baseQuery := `
		SELECT id, driver_id, tenant_id, type, severity, timestamp, location,
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

	start := time.Now()
	rows, err := r.db.Query(ctx, query, args...)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*models.DriverIncident
	for rows.Next() {
		var incident models.DriverIncident
		var locationJSON, detailsJSON []byte

		err := rows.Scan(
			&incident.ID, &incident.DriverID, &incident.TenantID, &incident.Type,
			&incident.Severity, &incident.Timestamp, &locationJSON, &incident.Speed,
			&detailsJSON, &incident.Resolved, &incident.ResolutionNotes, &incident.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(locationJSON, &incident.Location); err != nil {
			return nil, fmt.Errorf("failed to unmarshal location: %w", err)
		}
		if err := json.Unmarshal(detailsJSON, &incident.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		incidents = append(incidents, &incident)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("incident iteration error: %w", err)
	}

	return incidents, nil
}

// UpdateIncidentResolution updates incident resolution status
func (r *DriverRepository) UpdateIncidentResolution(ctx context.Context, incidentID string, resolved bool, notes string) error {
	query := `UPDATE driver_incidents SET resolved = $1, resolution_notes = $2, updated_at = NOW() WHERE id = $3`
	start := time.Now()
	result, err := r.db.Exec(ctx, query, resolved, notes, incidentID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update incident resolution: %w", err)
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
	detailsJSON, err := json.Marshal(update.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	start := time.Now()
	err = r.db.QueryRow(ctx, query,
		id, update.DriverID, update.ScoreChange, update.Reason,
		update.EventType, detailsJSON, update.CreatedBy, now,
	).Scan(&update.ID, &update.CreatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		return fmt.Errorf("failed to create behavior score update: %w", err)
	}

	return nil
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
	start := time.Now()
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&summary.Score, &summary.SafetyScore, &summary.EfficiencyScore, &summary.ComplianceScore,
		&summary.HarshAccelerations, &summary.HarshBrakings, &summary.SpeedingIncidents,
		&summary.FuelConsumption, &summary.DistanceDriven, &summary.RiskLevel,
	)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}

	summary.DriverID = driverID
	return &summary, nil
}

// Audit log methods

// CreateAuditLog creates a new audit log entry
func (r *DriverRepository) CreateAuditLog(ctx context.Context, audit *models.DriverAuditLog) error {
	query := `
		INSERT INTO driver_audit_logs (
			id, driver_id, tenant_id, action, old_value, new_value,
			user_id, user_type, user_email, reason, metadata,
			ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, ($5)::jsonb, ($6)::jsonb, $7, $8, $9, $10, ($11)::jsonb, $12, $13, $14)
	`

	id := uuid.New().String()

	// Convert maps to JSON
	oldValueJSON, err := json.Marshal(audit.OldValue)
	if err != nil {
		return fmt.Errorf("failed to marshal old value: %w", err)
	}
	newValueJSON, err := json.Marshal(audit.NewValue)
	if err != nil {
		return fmt.Errorf("failed to marshal new value: %w", err)
	}
	metadataJSON, err := json.Marshal(audit.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	start := time.Now()
	_, err = r.db.Exec(ctx, query,
		id, audit.DriverID, audit.TenantID, audit.Action,
		oldValueJSON, newValueJSON, audit.UserID, audit.UserType, audit.UserEmail,
		audit.Reason, metadataJSON, audit.IPAddress, audit.UserAgent, audit.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a driver
