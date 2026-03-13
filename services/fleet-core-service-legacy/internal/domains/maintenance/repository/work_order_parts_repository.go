package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
)

// WorkOrderPartsRepository handles work order parts operations
type WorkOrderPartsRepository struct {
	db *pgxpool.Pool
}

// NewWorkOrderPartsRepository creates a new work order parts repository
func NewWorkOrderPartsRepository(db *pgxpool.Pool) *WorkOrderPartsRepository {
	return &WorkOrderPartsRepository{db: db}
}

// AssignPart adds a part to a work order
func (r *WorkOrderPartsRepository) AssignPart(ctx context.Context, workOrderID, partID uuid.UUID, quantity int, unitCost float64) (*models.WorkOrderPart, error) {
	part := &models.WorkOrderPart{
		WorkOrderID: workOrderID,
		PartID:      partID,
		Quantity:    quantity,
		UnitCost:    unitCost,
	}

	totalCost := float64(quantity) * unitCost
	part.TotalCost = totalCost

	query := `
		INSERT INTO work_order_parts (work_order_id, part_id, quantity, unit_cost, total_cost)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		workOrderID, partID, quantity, unitCost, totalCost,
	).Scan(&part.ID, &part.CreatedAt)

	return part, err
}

// RemovePart removes a part from a work order
func (r *WorkOrderPartsRepository) RemovePart(ctx context.Context, workOrderID, partID uuid.UUID) (*models.WorkOrderPart, error) {
	// First get the part details for stock restoration
	part := &models.WorkOrderPart{}
	query := `
		SELECT id, work_order_id, part_id, quantity, unit_cost, total_cost, created_at
		FROM work_order_parts
		WHERE work_order_id = $1 AND part_id = $2
	`
	err := r.db.QueryRow(ctx, query, workOrderID, partID).Scan(
		&part.ID, &part.WorkOrderID, &part.PartID, &part.Quantity,
		&part.UnitCost, &part.TotalCost, &part.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Delete the part assignment
	_, err = r.db.Exec(ctx,
		"DELETE FROM work_order_parts WHERE work_order_id = $1 AND part_id = $2",
		workOrderID, partID,
	)

	return part, err
}

// GetPartsForWorkOrder retrieves all parts for a work order
func (r *WorkOrderPartsRepository) GetPartsForWorkOrder(ctx context.Context, workOrderID uuid.UUID) ([]models.WorkOrderPart, error) {
	query := `
		SELECT wop.id, wop.work_order_id, wop.part_id, wop.quantity, wop.unit_cost, wop.total_cost, wop.created_at,
			   ip.part_number, ip.name, ip.description, ip.stock_location
		FROM work_order_parts wop
		JOIN inventory_parts ip ON wop.part_id = ip.id
		WHERE wop.work_order_id = $1
		ORDER BY wop.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, workOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.WorkOrderPart
	for rows.Next() {
		var wop models.WorkOrderPart
		wop.PartInfo = &models.InventoryPart{}

		if err := rows.Scan(
			&wop.ID, &wop.WorkOrderID, &wop.PartID, &wop.Quantity, &wop.UnitCost, &wop.TotalCost, &wop.CreatedAt,
			&wop.PartInfo.PartNumber, &wop.PartInfo.Name, &wop.PartInfo.Description, &wop.PartInfo.StockLocation,
		); err != nil {
			return nil, err
		}
		wop.PartInfo.ID = wop.PartID
		parts = append(parts, wop)
	}

	return parts, rows.Err()
}

// UpdatePartQuantity updates the quantity of an assigned part
func (r *WorkOrderPartsRepository) UpdatePartQuantity(ctx context.Context, workOrderID, partID uuid.UUID, quantity int) error {
	query := `
		UPDATE work_order_parts SET quantity = $3
		WHERE work_order_id = $1 AND part_id = $2
	`
	_, err := r.db.Exec(ctx, query, workOrderID, partID, quantity)
	return err
}

// GetTotalPartsCost calculates total parts cost for a work order
func (r *WorkOrderPartsRepository) GetTotalPartsCost(ctx context.Context, workOrderID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.QueryRow(ctx,
		"SELECT COALESCE(SUM(total_cost), 0) FROM work_order_parts WHERE work_order_id = $1",
		workOrderID,
	).Scan(&total)
	return total, err
}
