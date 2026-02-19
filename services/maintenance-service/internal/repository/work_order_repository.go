package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/maintenance-service/internal/models"
)

// WorkOrderRepository handles work order database operations
type WorkOrderRepository struct {
	db *pgxpool.Pool
}

// NewWorkOrderRepository creates a new work order repository
func NewWorkOrderRepository(db *pgxpool.Pool) *WorkOrderRepository {
	return &WorkOrderRepository{db: db}
}

// Create inserts a new work order
func (r *WorkOrderRepository) Create(ctx context.Context, wo *models.WorkOrder) error {
	query := `
		INSERT INTO work_orders (
			organization_id, vehicle_id, title, description,
			maintenance_type, priority, status, assigned_to,
			scheduled_date, estimated_cost, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		wo.OrganizationID, wo.VehicleID, wo.Title, wo.Description,
		wo.MaintenanceType, wo.Priority, wo.Status, wo.AssignedTo,
		wo.ScheduledDate, wo.EstimatedCost, wo.CreatedBy,
	).Scan(&wo.ID, &wo.CreatedAt, &wo.UpdatedAt)
}

// GetByID retrieves a work order by ID
func (r *WorkOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkOrder, error) {
	query := `
		SELECT id, organization_id, vehicle_id, title, description,
			   maintenance_type, priority, status, assigned_to,
			   scheduled_date, completed_date, estimated_cost, actual_cost,
			   created_by, created_at, updated_at
		FROM work_orders WHERE id = $1
	`
	wo := &models.WorkOrder{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&wo.ID, &wo.OrganizationID, &wo.VehicleID, &wo.Title, &wo.Description,
		&wo.MaintenanceType, &wo.Priority, &wo.Status, &wo.AssignedTo,
		&wo.ScheduledDate, &wo.CompletedDate, &wo.EstimatedCost, &wo.ActualCost,
		&wo.CreatedBy, &wo.CreatedAt, &wo.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return wo, err
}

// Update updates a work order
func (r *WorkOrderRepository) Update(ctx context.Context, wo *models.WorkOrder) error {
	query := `
		UPDATE work_orders SET
			title = $2, description = $3, maintenance_type = $4,
			priority = $5, status = $6, assigned_to = $7,
			scheduled_date = $8, completed_date = $9,
			estimated_cost = $10, actual_cost = $11
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		wo.ID, wo.Title, wo.Description, wo.MaintenanceType,
		wo.Priority, wo.Status, wo.AssignedTo,
		wo.ScheduledDate, wo.CompletedDate,
		wo.EstimatedCost, wo.ActualCost,
	).Scan(&wo.UpdatedAt)
}

// Delete deletes a work order
func (r *WorkOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM work_orders WHERE id = $1", id)
	return err
}

// List retrieves work orders with filters
func (r *WorkOrderRepository) List(ctx context.Context, filter models.WorkOrderFilter) ([]models.WorkOrder, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIdx))
	args = append(args, filter.OrganizationID)
	argIdx++

	if filter.VehicleID != nil {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIdx))
		args = append(args, *filter.VehicleID)
		argIdx++
	}

	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, s := range filter.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, s)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Priority) > 0 {
		placeholders := make([]string, len(filter.Priority))
		for i, p := range filter.Priority {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, p)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("priority IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, *filter.AssignedTo)
		argIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM work_orders WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get items
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortDir := "DESC"
	if filter.SortDir == "asc" {
		sortDir = "ASC"
	}

	offset := (filter.Page - 1) * filter.PageSize
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, vehicle_id, title, description,
			   maintenance_type, priority, status, assigned_to,
			   scheduled_date, completed_date, estimated_cost, actual_cost,
			   created_by, created_at, updated_at
		FROM work_orders WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortDir, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var workOrders []models.WorkOrder
	for rows.Next() {
		var wo models.WorkOrder
		if err := rows.Scan(
			&wo.ID, &wo.OrganizationID, &wo.VehicleID, &wo.Title, &wo.Description,
			&wo.MaintenanceType, &wo.Priority, &wo.Status, &wo.AssignedTo,
			&wo.ScheduledDate, &wo.CompletedDate, &wo.EstimatedCost, &wo.ActualCost,
			&wo.CreatedBy, &wo.CreatedAt, &wo.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		workOrders = append(workOrders, wo)
	}

	return workOrders, total, rows.Err()
}

// UpdateStatus updates only the status field
func (r *WorkOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.WorkOrderStatus) error {
	query := "UPDATE work_orders SET status = $2 WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

// Close marks a work order as completed
func (r *WorkOrderRepository) Close(ctx context.Context, id uuid.UUID, actualCost float64) error {
	query := `
		UPDATE work_orders SET
			status = 'COMPLETED',
			completed_date = $2,
			actual_cost = $3
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, time.Now(), actualCost)
	return err
}

// GetStats retrieves work order statistics for an organization
func (r *WorkOrderRepository) GetStats(ctx context.Context, orgID uuid.UUID) (*models.WorkOrderStats, error) {
	stats := &models.WorkOrderStats{
		ByStatus:   make(map[string]int),
		ByPriority: make(map[string]int),
	}

	// Active work orders (not completed)
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM work_orders WHERE organization_id = $1 AND status != 'COMPLETED'",
		orgID,
	).Scan(&stats.ActiveWorkOrders)
	if err != nil {
		return nil, err
	}

	// Status breakdown
	rows, err := r.db.Query(ctx,
		"SELECT status, COUNT(*) FROM work_orders WHERE organization_id = $1 GROUP BY status",
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.ByStatus[status] = count
	}

	// Priority breakdown
	rows, err = r.db.Query(ctx,
		"SELECT priority, COUNT(*) FROM work_orders WHERE organization_id = $1 AND status != 'COMPLETED' GROUP BY priority",
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		stats.ByPriority[priority] = count
	}

	return stats, nil
}
