package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *MaintenanceRepositoryImpl) CreatePO(ctx context.Context, po *models.PurchaseOrder) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if po.ID == uuid.Nil {
		po.ID = uuid.New()
	}
	now := time.Now()
	po.CreatedAt = now
	po.UpdatedAt = now

	poQuery := `
		INSERT INTO purchase_orders (id, tenant_id, vendor_id, order_number, status, total_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(ctx, poQuery,
		po.ID, po.TenantID, po.VendorID, po.OrderNumber, po.Status, po.TotalCost, po.CreatedAt, po.UpdatedAt,
	)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO purchase_order_items (id, purchase_order_id, part_id, quantity, unit_cost, total_cost)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for i := range po.Items {
		item := &po.Items[i]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.PurchaseOrderID = po.ID
		_, err = tx.Exec(ctx, itemQuery,
			item.ID, item.PurchaseOrderID, item.PartID, item.Quantity, item.UnitCost, item.TotalCost,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *MaintenanceRepositoryImpl) GetPOByID(ctx context.Context, id uuid.UUID) (*models.PurchaseOrder, error) {
	query := `SELECT id, tenant_id, vendor_id, order_number, status, total_cost, created_at FROM purchase_orders WHERE id = $1`
	var po models.PurchaseOrder
	err := r.db.QueryRow(ctx, query, id).Scan(&po.ID, &po.TenantID, &po.VendorID, &po.OrderNumber, &po.Status, &po.TotalCost, &po.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Load items
	itemQuery := `SELECT id, purchase_order_id, part_id, quantity, unit_cost, total_cost FROM purchase_order_items WHERE purchase_order_id = $1`
	rows, err := r.db.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.PurchaseOrderItem
		if err := rows.Scan(&item.ID, &item.PurchaseOrderID, &item.PartID, &item.Quantity, &item.UnitCost, &item.TotalCost); err != nil {
			return nil, err
		}
		po.Items = append(po.Items, item)
	}

	return &po, nil
}

func (r *MaintenanceRepositoryImpl) UpdatePOStatus(ctx context.Context, id uuid.UUID, status models.POStatus) error {
	query := `UPDATE purchase_orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}
