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

	"github.com/flyfleet/maintenance-service/internal/models"
)


type MaintenanceRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewMaintenanceRepository(db *pgxpool.Pool) *MaintenanceRepositoryImpl {
	return &MaintenanceRepositoryImpl{db: db}
}

// CreateSchedule inserts a new maintenance schedule into the database
func (r *MaintenanceRepositoryImpl) CreateSchedule(ctx context.Context, schedule *models.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedules (
			id, org_id, vehicle_id, type, name, description, due_date, due_odometer, 
			due_engine_hours, assigned_to, priority, status, estimated_cost, 
			parts_required, service_items, notes, reminders, recurring, 
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	partsRequiredJSON, _ := json.Marshal(schedule.PartsRequired)
	serviceItemsJSON, _ := json.Marshal(schedule.ServiceItems)
	remindersJSON, _ := json.Marshal(schedule.Reminders)
	recurringJSON, _ := json.Marshal(schedule.Recurring)

	err := r.db.QueryRow(ctx, query,
		id, schedule.OrgID, schedule.VehicleID, schedule.Type, schedule.Name,
		schedule.Description, schedule.DueDate, schedule.DueOdometer,
		schedule.DueEngineHours, schedule.AssignedTo, schedule.Priority,
		schedule.Status, schedule.EstimatedCost, partsRequiredJSON,
		serviceItemsJSON, schedule.Notes, remindersJSON, recurringJSON,
		schedule.CreatedBy, now, now,
	).Scan(&schedule.ID, &schedule.CreatedAt, &schedule.UpdatedAt)

	return err
}

// GetScheduleByID retrieves a maintenance schedule by its ID
func (r *MaintenanceRepositoryImpl) GetScheduleByID(ctx context.Context, id string) (*models.MaintenanceSchedule, error) {
	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer, 
		       due_engine_hours, assigned_to, priority, status, estimated_cost, actual_cost,
		       parts_required, service_items, notes, reminders, recurring, 
		       created_by, created_at, updated_at, completed_at, completed_by
		FROM maintenance_schedules WHERE id = $1`

	var schedule models.MaintenanceSchedule
	var partsRequiredJSON, serviceItemsJSON, remindersJSON, recurringJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&schedule.ID, &schedule.OrgID, &schedule.VehicleID, &schedule.Type,
		&schedule.Name, &schedule.Description, &schedule.DueDate, &schedule.DueOdometer,
		&schedule.DueEngineHours, &schedule.AssignedTo, &schedule.Priority,
		&schedule.Status, &schedule.EstimatedCost, &schedule.ActualCost,
		&partsRequiredJSON, &serviceItemsJSON, &schedule.Notes, &remindersJSON,
		&recurringJSON, &schedule.CreatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
		&schedule.CompletedAt, &schedule.CompletedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(partsRequiredJSON, &schedule.PartsRequired)
	json.Unmarshal(serviceItemsJSON, &schedule.ServiceItems)
	json.Unmarshal(remindersJSON, &schedule.Reminders)
	json.Unmarshal(recurringJSON, &schedule.Recurring)

	return &schedule, nil
}

// ListSchedules retrieves maintenance schedules with filtering and pagination
func (r *MaintenanceRepositoryImpl) ListSchedules(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceSchedule, int, error) {
	baseQuery := `
		SELECT id, org_id, vehicle_id, type, name, due_date, priority, status, 
		       estimated_cost, created_at, updated_at
		FROM maintenance_schedules WHERE org_id = $1`

	countQuery := `SELECT COUNT(*) FROM maintenance_schedules WHERE org_id = $1`

	var conditions []string
	var args []interface{}
	argIndex := 2
	args = append(args, filter.OrgID)

	// Add filters
	if filter.VehicleID != nil {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, *filter.VehicleID)
		argIndex++
	}

	if len(filter.Type) > 0 {
		typePlaceholders := make([]string, len(filter.Type))
		for i, scheduleType := range filter.Type {
			typePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, scheduleType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(typePlaceholders, ",")))
	}

	if len(filter.Status) > 0 {
		statusPlaceholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusPlaceholders, ",")))
	}

	if len(filter.Priority) > 0 {
		priorityPlaceholders := make([]string, len(filter.Priority))
		for i, priority := range filter.Priority {
			priorityPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, priority)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("priority IN (%s)", strings.Join(priorityPlaceholders, ",")))
	}

	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIndex))
		args = append(args, *filter.AssignedTo)
		argIndex++
	}

	if !filter.FromDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argIndex))
		args = append(args, filter.FromDate)
		argIndex++
	}

	if !filter.ToDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIndex))
		args = append(args, filter.ToDate)
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

	orderBy := "due_date ASC"
	if filter.OrderBy != "" {
		allowedColumns := map[string]bool{
			"due_date": true, "created_at": true, "name": true, "type": true,
			"priority": true, "status": true,
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

	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var schedule models.MaintenanceSchedule
		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.VehicleID, &schedule.Type,
			&schedule.Name, &schedule.DueDate, &schedule.Priority, &schedule.Status,
			&schedule.EstimatedCost, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, totalCount, nil
}

// UpdateSchedule modifies an existing maintenance schedule
func (r *MaintenanceRepositoryImpl) UpdateSchedule(ctx context.Context, id string, updates map[string]interface{}) error {
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
		case "parts_required", "service_items", "reminders", "recurring":
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE maintenance_schedules SET %s, updated_at = $%d WHERE id = $%d",
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

// DeleteSchedule removes a maintenance schedule
func (r *MaintenanceRepositoryImpl) DeleteSchedule(ctx context.Context, id string) error {
	query := "DELETE FROM maintenance_schedules WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// CreateLog inserts a new maintenance log into the database
func (r *MaintenanceRepositoryImpl) CreateLog(ctx context.Context, log *models.MaintenanceLog) error {
	query := `
		INSERT INTO maintenance_logs (
			id, org_id, vehicle_id, schedule_id, type, name, description, start_date, 
			end_date, duration_minutes, performed_by, cost, odometer, engine_hours,
			parts_used, service_items, notes, documents, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	partsUsedJSON, _ := json.Marshal(log.PartsUsed)
	serviceItemsJSON, _ := json.Marshal(log.ServiceItems)
	documentsJSON, _ := json.Marshal(log.Documents)

	// Calculate duration in minutes
	duration := int(log.EndDate.Sub(log.StartDate).Minutes())

	err := r.db.QueryRow(ctx, query,
		id, log.OrgID, log.VehicleID, log.ScheduleID, log.Type, log.Name,
		log.Description, log.StartDate, log.EndDate, duration, log.PerformedBy,
		log.Cost, log.Odometer, log.EngineHours, partsUsedJSON, serviceItemsJSON,
		log.Notes, documentsJSON, log.Status, now, now,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)

	return err
}

// GetLogByID retrieves a maintenance log by its ID
func (r *MaintenanceRepositoryImpl) GetLogByID(ctx context.Context, id string) (*models.MaintenanceLog, error) {
	query := `
		SELECT id, org_id, vehicle_id, schedule_id, type, name, description, start_date, 
		       end_date, duration_minutes, performed_by, cost, odometer, engine_hours,
		       parts_used, service_items, notes, documents, status, created_at, updated_at
		FROM maintenance_logs WHERE id = $1`

	var log models.MaintenanceLog
	var partsUsedJSON, serviceItemsJSON, documentsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.OrgID, &log.VehicleID, &log.ScheduleID, &log.Type,
		&log.Name, &log.Description, &log.StartDate, &log.EndDate,
		&log.DurationMinutes, &log.PerformedBy, &log.Cost, &log.Odometer,
		&log.EngineHours, &partsUsedJSON, &serviceItemsJSON, &log.Notes,
		&documentsJSON, &log.Status, &log.CreatedAt, &log.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(partsUsedJSON, &log.PartsUsed)
	json.Unmarshal(serviceItemsJSON, &log.ServiceItems)
	json.Unmarshal(documentsJSON, &log.Documents)

	return &log, nil
}

// ListLogs retrieves maintenance logs with filtering and pagination
func (r *MaintenanceRepositoryImpl) ListLogs(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceLog, int, error) {
	baseQuery := `
		SELECT id, org_id, vehicle_id, type, name, start_date, performed_by, 
		       cost, odometer, status, created_at, updated_at
		FROM maintenance_logs WHERE org_id = $1`

	countQuery := `SELECT COUNT(*) FROM maintenance_logs WHERE org_id = $1`

	var conditions []string
	var args []interface{}
	argIndex := 2
	args = append(args, filter.OrgID)

	// Add filters
	if filter.VehicleID != nil {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, *filter.VehicleID)
		argIndex++
	}

	if len(filter.Type) > 0 {
		typePlaceholders := make([]string, len(filter.Type))
		for i, logType := range filter.Type {
			typePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, logType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(typePlaceholders, ",")))
	}

	if len(filter.Status) > 0 {
		statusPlaceholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusPlaceholders, ",")))
	}

	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("performed_by = $%d", argIndex))
		args = append(args, *filter.AssignedTo)
		argIndex++
	}

	if !filter.FromDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_date >= $%d", argIndex))
		args = append(args, filter.FromDate)
		argIndex++
	}

	if !filter.ToDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_date <= $%d", argIndex))
		args = append(args, filter.ToDate)
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

	orderBy := "start_date DESC"
	if filter.OrderBy != "" {
		allowedColumns := map[string]bool{
			"start_date": true, "created_at": true, "name": true, "type": true,
			"cost": true, "status": true,
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

	var logs []*models.MaintenanceLog
	for rows.Next() {
		var log models.MaintenanceLog
		err := rows.Scan(
			&log.ID, &log.OrgID, &log.VehicleID, &log.Type, &log.Name,
			&log.StartDate, &log.PerformedBy, &log.Cost, &log.Odometer,
			&log.Status, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}

	return logs, totalCount, nil
}

// UpdateLog modifies an existing maintenance log
func (r *MaintenanceRepositoryImpl) UpdateLog(ctx context.Context, id string, updates map[string]interface{}) error {
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
		case "parts_used", "service_items", "documents":
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE maintenance_logs SET %s, updated_at = $%d WHERE id = $%d",
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

// DeleteLog removes a maintenance log
func (r *MaintenanceRepositoryImpl) DeleteLog(ctx context.Context, id string) error {
	query := "DELETE FROM maintenance_logs WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetMaintenanceHistory retrieves maintenance history for a vehicle
func (r *MaintenanceRepositoryImpl) GetMaintenanceHistory(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceLog, error) {
	query := `
		SELECT id, org_id, vehicle_id, schedule_id, type, name, description, start_date, 
		       end_date, duration_minutes, performed_by, cost, odometer, engine_hours,
		       parts_used, service_items, notes, documents, status, created_at, updated_at
		FROM maintenance_logs 
		WHERE vehicle_id = $1 
		ORDER BY start_date DESC 
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.MaintenanceLog
	for rows.Next() {
		var log models.MaintenanceLog
		var partsUsedJSON, serviceItemsJSON, documentsJSON []byte

		err := rows.Scan(
			&log.ID, &log.OrgID, &log.VehicleID, &log.ScheduleID, &log.Type,
			&log.Name, &log.Description, &log.StartDate, &log.EndDate,
			&log.DurationMinutes, &log.PerformedBy, &log.Cost, &log.Odometer,
			&log.EngineHours, &partsUsedJSON, &serviceItemsJSON, &log.Notes,
			&documentsJSON, &log.Status, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(partsUsedJSON, &log.PartsUsed)
		json.Unmarshal(serviceItemsJSON, &log.ServiceItems)
		json.Unmarshal(documentsJSON, &log.Documents)

		logs = append(logs, &log)
	}

	return logs, nil
}

// GetOverdueSchedules retrieves overdue maintenance schedules
func (r *MaintenanceRepositoryImpl) GetOverdueSchedules(ctx context.Context, orgID uuid.UUID) ([]*models.MaintenanceSchedule, error) {
	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer, 
		       due_engine_hours, assigned_to, priority, status, estimated_cost, 
		       parts_required, service_items, notes, created_by, created_at, updated_at
		FROM maintenance_schedules 
		WHERE org_id = $1 
		  AND status = $2 
		  AND due_date < $3
		ORDER BY due_date ASC`

	rows, err := r.db.Query(ctx, query, orgID, models.StatusScheduled, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var schedule models.MaintenanceSchedule
		var partsRequiredJSON, serviceItemsJSON []byte

		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.VehicleID, &schedule.Type,
			&schedule.Name, &schedule.Description, &schedule.DueDate, &schedule.DueOdometer,
			&schedule.DueEngineHours, &schedule.AssignedTo, &schedule.Priority,
			&schedule.Status, &schedule.EstimatedCost, &partsRequiredJSON,
			&serviceItemsJSON, &schedule.Notes, &schedule.CreatedBy,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(partsRequiredJSON, &schedule.PartsRequired)
		json.Unmarshal(serviceItemsJSON, &schedule.ServiceItems)

		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// GetUpcomingSchedules retrieves upcoming maintenance schedules
func (r *MaintenanceRepositoryImpl) GetUpcomingSchedules(ctx context.Context, orgID uuid.UUID, days int) ([]*models.MaintenanceSchedule, error) {
	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer, 
		       due_engine_hours, assigned_to, priority, status, estimated_cost, 
		       parts_required, service_items, notes, created_by, created_at, updated_at
		FROM maintenance_schedules 
		WHERE org_id = $1 
		  AND status = $2 
		  AND due_date >= $3 
		  AND due_date <= $4
		ORDER BY due_date ASC`

	fromDate := time.Now()
	toDate := fromDate.AddDate(0, 0, days)

	rows, err := r.db.Query(ctx, query, orgID, models.StatusScheduled, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var schedule models.MaintenanceSchedule
		var partsRequiredJSON, serviceItemsJSON []byte

		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.VehicleID, &schedule.Type,
			&schedule.Name, &schedule.Description, &schedule.DueDate, &schedule.DueOdometer,
			&schedule.DueEngineHours, &schedule.AssignedTo, &schedule.Priority,
			&schedule.Status, &schedule.EstimatedCost, &partsRequiredJSON,
			&serviceItemsJSON, &schedule.Notes, &schedule.CreatedBy,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(partsRequiredJSON, &schedule.PartsRequired)
		json.Unmarshal(serviceItemsJSON, &schedule.ServiceItems)

		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// GetMaintenanceSummary retrieves maintenance statistics for an organization
func (r *MaintenanceRepositoryImpl) GetMaintenanceSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.MaintenanceSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_schedules,
			COUNT(CASE WHEN status = 'scheduled' THEN 1 END) as scheduled,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled,
			COUNT(CASE WHEN due_date < $4 AND status = 'scheduled' THEN 1 END) as overdue,
			COUNT(CASE WHEN due_date >= $4 AND due_date <= $5 AND status = 'scheduled' THEN 1 END) as upcoming,
			SUM(CASE WHEN status = 'completed' THEN cost ELSE 0 END) as total_cost
		FROM maintenance_logs 
		WHERE org_id = $1 
		  AND start_date >= $2 
		  AND start_date <= $3`

	var summary models.MaintenanceSummary
	var totalCost *float64
	err := r.db.QueryRow(ctx, query, orgID, fromDate, toDate, time.Now(), time.Now().AddDate(0, 0, 30)).Scan(
		&summary.TotalSchedules, &summary.Scheduled, &summary.InProgress,
		&summary.Completed, &summary.Cancelled, &summary.Overdue,
		&summary.Upcoming, &totalCost,
	)

	if err != nil {
		return nil, err
	}

	if totalCost != nil {
		summary.TotalCost = *totalCost
		if summary.Completed > 0 {
			summary.AverageCost = summary.TotalCost / float64(summary.Completed)
		}
	}

	// Initialize maps
	summary.ByType = make(map[string]int)
	summary.ByPriority = make(map[string]int)
	summary.ByStatus = make(map[string]int)

	// Get breakdown by type
	typeQuery := `
		SELECT type, COUNT(*) 
		FROM maintenance_logs 
		WHERE org_id = $1 AND start_date >= $2 AND start_date <= $3
		GROUP BY type`

	typeRows, err := r.db.Query(ctx, typeQuery, orgID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var logType string
		var count int
		typeRows.Scan(&logType, &count)
		summary.ByType[logType] = count
	}

	// Get maintenance history
	history, err := r.GetMaintenanceHistory(ctx, uuid.Nil, 10) // Get last 10 records
	if err != nil {
		return nil, err
	}
	for _, log := range history {
		summary.MaintenanceHistory = append(summary.MaintenanceHistory, *log)
	}

	return &summary, nil
}

// HealthCheck performs a health check on the database
func (r *MaintenanceRepositoryImpl) HealthCheck(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *MaintenanceRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// GetSchedulesByVehicle retrieves maintenance schedules for a specific vehicle
func (r *MaintenanceRepositoryImpl) GetSchedulesByVehicle(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceSchedule, error) {
	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer,
			due_engine_hours, assigned_to, priority, status, estimated_cost,
			parts_required, service_items, notes, reminders, recurring,
			created_by, created_at, updated_at
		FROM maintenance_schedules
		WHERE vehicle_id = $1
		ORDER BY due_date ASC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var schedule models.MaintenanceSchedule
		var partsRequiredJSON, serviceItemsJSON, remindersJSON, recurringJSON []byte

		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.VehicleID, &schedule.Type,
			&schedule.Name, &schedule.Description, &schedule.DueDate, &schedule.DueOdometer,
			&schedule.DueEngineHours, &schedule.AssignedTo, &schedule.Priority, &schedule.Status,
			&schedule.EstimatedCost, &partsRequiredJSON, &serviceItemsJSON, &schedule.Notes,
			&remindersJSON, &recurringJSON, &schedule.CreatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(partsRequiredJSON, &schedule.PartsRequired)
		json.Unmarshal(serviceItemsJSON, &schedule.ServiceItems)
		json.Unmarshal(remindersJSON, &schedule.Reminders)
		json.Unmarshal(recurringJSON, &schedule.Recurring)

		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// GetSchedulesByOrg retrieves maintenance schedules for an organization with filters
func (r *MaintenanceRepositoryImpl) GetSchedulesByOrg(ctx context.Context, orgID uuid.UUID, filters *ScheduleFilters) ([]*models.MaintenanceSchedule, error) {
	query := `
		SELECT id, org_id, vehicle_id, type, name, description, due_date, due_odometer,
			due_engine_hours, assigned_to, priority, status, estimated_cost,
			parts_required, service_items, notes, reminders, recurring,
			created_by, created_at, updated_at
		FROM maintenance_schedules
		WHERE org_id = $1`

	args := []interface{}{orgID}
	argIndex := 2

	if filters != nil {
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, *filters.Status)
			argIndex++
		}
		if filters.Type != nil {
			query += fmt.Sprintf(" AND type = $%d", argIndex)
			args = append(args, *filters.Type)
			argIndex++
		}
		if filters.Priority != nil {
			query += fmt.Sprintf(" AND priority = $%d", argIndex)
			args = append(args, *filters.Priority)
			argIndex++
		}
		if filters.FromDate != nil {
			query += fmt.Sprintf(" AND due_date >= $%d", argIndex)
			args = append(args, *filters.FromDate)
			argIndex++
		}
		if filters.ToDate != nil {
			query += fmt.Sprintf(" AND due_date <= $%d", argIndex)
			args = append(args, *filters.ToDate)
			argIndex++
		}
		if filters.Overdue != nil && *filters.Overdue {
			query += fmt.Sprintf(" AND due_date < $%d AND status = 'scheduled'", argIndex)
			args = append(args, time.Now())
			argIndex++
		}
		if filters.Upcoming != nil && *filters.Upcoming {
			now := time.Now()
			future := now.AddDate(0, 0, 30) // Next 30 days
			query += fmt.Sprintf(" AND due_date >= $%d AND due_date <= $%d AND status = 'scheduled'", argIndex, argIndex+1)
			args = append(args, now, future)
			argIndex += 2
		}
	}

	query += " ORDER BY due_date ASC"

	if filters != nil && filters.Page > 0 && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filters.PageSize, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var schedule models.MaintenanceSchedule
		var partsRequiredJSON, serviceItemsJSON, remindersJSON, recurringJSON []byte

		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.VehicleID, &schedule.Type,
			&schedule.Name, &schedule.Description, &schedule.DueDate, &schedule.DueOdometer,
			&schedule.DueEngineHours, &schedule.AssignedTo, &schedule.Priority, &schedule.Status,
			&schedule.EstimatedCost, &partsRequiredJSON, &serviceItemsJSON, &schedule.Notes,
			&remindersJSON, &recurringJSON, &schedule.CreatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(partsRequiredJSON, &schedule.PartsRequired)
		json.Unmarshal(serviceItemsJSON, &schedule.ServiceItems)
		json.Unmarshal(remindersJSON, &schedule.Reminders)
		json.Unmarshal(recurringJSON, &schedule.Recurring)

		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

