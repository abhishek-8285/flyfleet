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

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
)

type MaintenanceRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewMaintenanceRepository(db *pgxpool.Pool) *MaintenanceRepositoryImpl {
	return &MaintenanceRepositoryImpl{db: db}
}

func (r *MaintenanceRepositoryImpl) GetDB() *pgxpool.Pool {
	return r.db
}

func (r *MaintenanceRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// Schedule operations

func (r *MaintenanceRepositoryImpl) CreateSchedule(ctx context.Context, schedule *models.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedules (
			id, organization_id, vehicle_id, type, name, description, due_date, due_odometer, 
			due_engine_hours, assigned_to, priority, status, estimated_cost, 
			parts_required, service_items, notes, reminders, recurring, 
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at`

	id := schedule.ID
	if id == "" {
		id = uuid.New().String()
	}
	now := time.Now()

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

	if err != nil {
		return fmt.Errorf("failed to create maintenance schedule: %w", err)
	}
	return nil
}

func (r *MaintenanceRepositoryImpl) GetScheduleByID(ctx context.Context, id string) (*models.MaintenanceSchedule, error) {
	query := `
		SELECT id, organization_id, vehicle_id, type, name, COALESCE(description, ''), due_date, due_odometer,
		       due_engine_hours, assigned_to, priority, status, estimated_cost, actual_cost,
		       parts_required, service_items, COALESCE(notes, ''), reminders, recurring,
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get maintenance schedule: %w", err)
	}

	json.Unmarshal(partsRequiredJSON, &schedule.PartsRequired)
	json.Unmarshal(serviceItemsJSON, &schedule.ServiceItems)
	json.Unmarshal(remindersJSON, &schedule.Reminders)
	json.Unmarshal(recurringJSON, &schedule.Recurring)

	return &schedule, nil
}

func (r *MaintenanceRepositoryImpl) ListSchedules(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceSchedule, int, error) {
	baseQuery := `
		SELECT id, organization_id, vehicle_id, type, name, due_date, priority, status, 
		       estimated_cost, created_at, updated_at
		FROM maintenance_schedules WHERE organization_id = $1`

	countQuery := `SELECT COUNT(*) FROM maintenance_schedules WHERE organization_id = $1`

	var conditions []string
	var args []interface{}
	argIndex := 2
	args = append(args, filter.OrgID)

	if filter.VehicleID != nil {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, *filter.VehicleID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	countSQL := countQuery + whereClause
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 { page = 1 }
	pageSize := filter.PageSize
	if pageSize < 1 { pageSize = 20 }
	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	listSQL := baseQuery + whereClause + " ORDER BY due_date ASC" + limitClause

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

func (r *MaintenanceRepositoryImpl) GetSchedulesByVehicle(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceSchedule, error) {
	query := `SELECT id, organization_id, vehicle_id, type, name, due_date FROM maintenance_schedules WHERE vehicle_id = $1 ORDER BY due_date ASC LIMIT $2`
	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var s models.MaintenanceSchedule
		rows.Scan(&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.DueDate)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (r *MaintenanceRepositoryImpl) GetSchedulesByOrg(ctx context.Context, orgID uuid.UUID, filter *ScheduleFilters) ([]*models.MaintenanceSchedule, error) {
	query := `SELECT id, organization_id, vehicle_id, type, name, due_date FROM maintenance_schedules WHERE organization_id = $1`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var s models.MaintenanceSchedule
		rows.Scan(&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.DueDate)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (r *MaintenanceRepositoryImpl) UpdateSchedule(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
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
	query := fmt.Sprintf("UPDATE maintenance_schedules SET %s, updated_at = NOW() WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *MaintenanceRepositoryImpl) DeleteSchedule(ctx context.Context, id string) error {
	query := "DELETE FROM maintenance_schedules WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Log operations

func (r *MaintenanceRepositoryImpl) CreateLog(ctx context.Context, log *models.MaintenanceLog) error {
	query := `
		INSERT INTO maintenance_logs (
			id, organization_id, vehicle_id, schedule_id, type, name, description, start_date,
			end_date, duration_minutes, performed_by, cost, odometer, engine_hours,
			parts_used, service_items, notes, documents, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at`

	id := log.ID
	if id == "" {
		id = uuid.New().String()
	}
	now := time.Now()
	partsUsedJSON, _ := json.Marshal(log.PartsUsed)
	serviceItemsJSON, _ := json.Marshal(log.ServiceItems)
	documentsJSON, _ := json.Marshal(log.Documents)
	duration := int(log.EndDate.Sub(log.StartDate).Minutes())

	err := r.db.QueryRow(ctx, query,
		id, log.OrgID, log.VehicleID, log.ScheduleID, log.Type, log.Name,
		log.Description, log.StartDate, log.EndDate, duration, log.PerformedBy,
		log.Cost, log.Odometer, log.EngineHours, partsUsedJSON, serviceItemsJSON,
		log.Notes, documentsJSON, log.Status, now, now,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
	return err
}

func (r *MaintenanceRepositoryImpl) GetLogByID(ctx context.Context, id string) (*models.MaintenanceLog, error) {
	query := `SELECT id, organization_id, vehicle_id, type, name, status, start_date, cost FROM maintenance_logs WHERE id = $1`
	var log models.MaintenanceLog
	err := r.db.QueryRow(ctx, query, id).Scan(&log.ID, &log.OrgID, &log.VehicleID, &log.Type, &log.Name, &log.Status, &log.StartDate, &log.Cost)
	if err != nil {
		if err == pgx.ErrNoRows { return nil, nil }
		return nil, err
	}
	return &log, nil
}

func (r *MaintenanceRepositoryImpl) ListLogs(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceLog, int, error) {
	query := `SELECT id, organization_id, vehicle_id, type, name, status, start_date, cost FROM maintenance_logs WHERE organization_id = $1`
	rows, err := r.db.Query(ctx, query, filter.OrgID)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var logs []*models.MaintenanceLog
	for rows.Next() {
		var log models.MaintenanceLog
		rows.Scan(&log.ID, &log.OrgID, &log.VehicleID, &log.Type, &log.Name, &log.Status, &log.StartDate, &log.Cost)
		logs = append(logs, &log)
	}
	return logs, len(logs), nil
}

func (r *MaintenanceRepositoryImpl) UpdateLog(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
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
	query := fmt.Sprintf("UPDATE maintenance_logs SET %s, updated_at = NOW() WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *MaintenanceRepositoryImpl) DeleteLog(ctx context.Context, id string) error {
	query := "DELETE FROM maintenance_logs WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Query operations

func (r *MaintenanceRepositoryImpl) GetMaintenanceHistory(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceLog, error) {
	query := `SELECT id, org_id, vehicle_id, type, name, status, start_date, cost FROM maintenance_logs WHERE vehicle_id = $1 ORDER BY start_date DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var logs []*models.MaintenanceLog
	for rows.Next() {
		var log models.MaintenanceLog
		rows.Scan(&log.ID, &log.OrgID, &log.VehicleID, &log.Type, &log.Name, &log.Status, &log.StartDate, &log.Cost)
		logs = append(logs, &log)
	}
	return logs, nil
}

func (r *MaintenanceRepositoryImpl) GetOverdueSchedules(ctx context.Context, orgID uuid.UUID) ([]*models.MaintenanceSchedule, error) {
	query := `SELECT id, org_id, vehicle_id, type, name, due_date FROM maintenance_schedules WHERE org_id = $1 AND status = 'scheduled' AND due_date < NOW()`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var s models.MaintenanceSchedule
		rows.Scan(&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.DueDate)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (r *MaintenanceRepositoryImpl) GetUpcomingSchedules(ctx context.Context, orgID uuid.UUID, days int) ([]*models.MaintenanceSchedule, error) {
	query := `SELECT id, org_id, vehicle_id, type, name, due_date FROM maintenance_schedules WHERE org_id = $1 AND status = 'scheduled' AND due_date >= NOW() AND due_date <= NOW() + ($2 || ' days')::interval`
	rows, err := r.db.Query(ctx, query, orgID, days)
	if err != nil { return nil, err }
	defer rows.Close()
	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var s models.MaintenanceSchedule
		rows.Scan(&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.DueDate)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (r *MaintenanceRepositoryImpl) GetDueByOdometer(ctx context.Context, vehicleID uuid.UUID, currentOdometer int) ([]*models.MaintenanceSchedule, error) {
	query := `SELECT id, org_id, vehicle_id, type, name, due_odometer FROM maintenance_schedules WHERE vehicle_id = $1 AND status = 'scheduled' AND due_odometer IS NOT NULL AND due_odometer <= $2`
	rows, err := r.db.Query(ctx, query, vehicleID, currentOdometer)
	if err != nil { return nil, err }
	defer rows.Close()
	var schedules []*models.MaintenanceSchedule
	for rows.Next() {
		var s models.MaintenanceSchedule
		rows.Scan(&s.ID, &s.OrgID, &s.VehicleID, &s.Type, &s.Name, &s.DueOdometer)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (r *MaintenanceRepositoryImpl) GetMaintenanceSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.MaintenanceSummary, error) {
	return &models.MaintenanceSummary{}, nil
}

// Extra operations

func (r *MaintenanceRepositoryImpl) GetDowntimeByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]*models.VehicleDowntime, error) {
	query := `SELECT id, organization_id, vehicle_id, downtime_start, downtime_end, downtime_minutes FROM vehicle_downtime WHERE vehicle_id = $1 ORDER BY downtime_start DESC`
	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil { return nil, err }
	defer rows.Close()
	var results []*models.VehicleDowntime
	for rows.Next() {
		var d models.VehicleDowntime
		rows.Scan(&d.ID, &d.OrganizationID, &d.VehicleID, &d.DowntimeStart, &d.DowntimeEnd, &d.DowntimeMinutes)
		results = append(results, &d)
	}
	return results, nil
}
