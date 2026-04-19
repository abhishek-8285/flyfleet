package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
)

// InventoryRepository handles inventory database operations
type InventoryRepository struct {
	db *pgxpool.Pool
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// Create inserts a new inventory part
func (r *InventoryRepository) Create(ctx context.Context, part *models.InventoryPart) error {
	query := `
		INSERT INTO inventory_parts (
			tenant_id, part_number, name, description,
			stock_location, quantity_available, reorder_level, unit_cost
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		part.OrgID, part.PartNumber, part.Name, part.Description,
		part.StockLocation, part.QuantityAvailable, part.ReorderLevel, part.UnitCost,
	).Scan(&part.ID, &part.CreatedAt, &part.UpdatedAt)
}

// GetByID retrieves an inventory part by ID
func (r *InventoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.InventoryPart, error) {
	query := `
		SELECT id, tenant_id, part_number, name, COALESCE(description, ''),
			   COALESCE(stock_location, ''), quantity_available, reorder_level, unit_cost,
			   created_at, updated_at
		FROM inventory_parts WHERE id = $1
	`
	part := &models.InventoryPart{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&part.ID, &part.OrgID, &part.PartNumber, &part.Name, &part.Description,
		&part.StockLocation, &part.QuantityAvailable, &part.ReorderLevel, &part.UnitCost,
		&part.CreatedAt, &part.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return part, err
}

// Update updates an inventory part
func (r *InventoryRepository) Update(ctx context.Context, part *models.InventoryPart) error {
	query := `
		UPDATE inventory_parts SET
			part_number = $2, name = $3, description = $4,
			stock_location = $5, quantity_available = $6,
			reorder_level = $7, unit_cost = $8
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		part.ID, part.PartNumber, part.Name, part.Description,
		part.StockLocation, part.QuantityAvailable, part.ReorderLevel, part.UnitCost,
	).Scan(&part.UpdatedAt)
}

// Delete deletes an inventory part
func (r *InventoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM inventory_parts WHERE id = $1", id)
	return err
}

// List retrieves inventory parts with filters
func (r *InventoryRepository) List(ctx context.Context, filter models.InventoryFilter) ([]models.InventoryPart, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, filter.OrgID)
	argIdx++

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR part_number ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.StockLocation != "" {
		conditions = append(conditions, fmt.Sprintf("stock_location = $%d", argIdx))
		args = append(args, filter.StockLocation)
		argIdx++
	}

	if filter.LowStockOnly {
		conditions = append(conditions, "quantity_available <= reorder_level")
	}

	if filter.AvailableOnly {
		conditions = append(conditions, "quantity_available > 0")
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory_parts WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get items
	offset := (filter.Page - 1) * filter.PageSize
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, part_number, name, COALESCE(description, ''),
			   COALESCE(stock_location, ''), quantity_available, reorder_level, unit_cost,
			   created_at, updated_at
		FROM inventory_parts WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var parts []models.InventoryPart
	for rows.Next() {
		var part models.InventoryPart
		if err := rows.Scan(
			&part.ID, &part.OrgID, &part.PartNumber, &part.Name, &part.Description,
			&part.StockLocation, &part.QuantityAvailable, &part.ReorderLevel, &part.UnitCost,
			&part.CreatedAt, &part.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		parts = append(parts, part)
	}

	return parts, total, rows.Err()
}

// DeductStock reduces quantity available (for work order part assignment)
func (r *InventoryRepository) DeductStock(ctx context.Context, id uuid.UUID, quantity int) error {
	query := `
		UPDATE inventory_parts SET
			quantity_available = quantity_available - $2
		WHERE id = $1 AND quantity_available >= $2
	`
	result, err := r.db.Exec(ctx, query, id, quantity)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient stock for part %s", id)
	}
	return nil
}

// RestoreStock adds back quantity (when removing part from work order)
func (r *InventoryRepository) RestoreStock(ctx context.Context, id uuid.UUID, quantity int) error {
	query := `
		UPDATE inventory_parts SET
			quantity_available = quantity_available + $2
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, quantity)
	return err
}

// GetLowStockCount returns count of parts below reorder level
func (r *InventoryRepository) GetLowStockCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM inventory_parts WHERE tenant_id = $1 AND quantity_available <= reorder_level",
		orgID,
	).Scan(&count)
	return count, err
}
