package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepository struct {
	pool *pgxpool.Pool
}

func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{pool: pool}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *invoicing.Invoice) (*invoicing.Invoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO invoices (
			tenant_id, invoice_number, customer_name, customer_email, customer_phone, billing_address,
			due_date, status, subtotal, tax_amount, total_amount, currency, notes,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		inv.OrgID, inv.InvoiceNumber, inv.CustomerName, inv.CustomerEmail, inv.CustomerPhone, inv.BillingAddress,
		inv.DueDate, inv.Status, inv.Subtotal, inv.TaxAmount, inv.TotalAmount, inv.Currency, inv.Notes,
		inv.CreatedBy, inv.UpdatedBy,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert invoice: %w", err)
	}

	// Batch insert items
	if len(inv.Items) > 0 {
		var args []interface{}
		var placeholders []string
		for i, item := range inv.Items {
			item.ID = uuid.New()
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()

			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
			args = append(args,
				item.ID,
				inv.OrgID,
				inv.ID,
				item.Description,
				item.Quantity,
				item.UnitPrice,
				item.Amount,
				inv.CreatedBy,
				inv.UpdatedBy,
			)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO invoice_items (id, tenant_id, invoice_id, description, quantity, unit_price, amount, created_by, updated_by) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err = tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to batch insert invoice items: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return inv, nil
}

func (r *InvoiceRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*invoicing.Invoice, error) {
	query := `
		SELECT id, tenant_id, COALESCE(invoice_number, '') as invoice_number, 
		       COALESCE(customer_name, '') as customer_name, 
		       COALESCE(customer_email, '') as customer_email, 
		       COALESCE(customer_phone, '') as customer_phone, 
		       COALESCE(billing_address, '') as billing_address,
		       due_date, status, subtotal, tax_amount, total_amount, amount_paid, amount_remaining,
		       COALESCE(currency, 'INR') as currency, COALESCE(notes, '') as notes, 
		       created_at, COALESCE(created_by, '00000000-0000-0000-0000-000000000000') as created_by, 
		       updated_at, COALESCE(updated_by, '00000000-0000-0000-0000-000000000000') as updated_by
		FROM invoices
		WHERE id = $1 AND tenant_id = $2 AND is_deleted = false
	`
	inv := &invoicing.Invoice{}
	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(
		&inv.ID, &inv.OrgID, &inv.InvoiceNumber, &inv.CustomerName, &inv.CustomerEmail, &inv.CustomerPhone, &inv.BillingAddress,
		&inv.DueDate, &inv.Status, &inv.Subtotal, &inv.TaxAmount, &inv.TotalAmount, &inv.AmountPaid, &inv.AmountRemaining,
		&inv.Currency, &inv.Notes, &inv.CreatedAt, &inv.CreatedBy, &inv.UpdatedAt, &inv.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Fetch items
	itemQuery := `
		SELECT id, invoice_id, description, quantity, unit_price, amount, created_at, created_by, updated_at, updated_by
		FROM invoice_items
		WHERE invoice_id = $1 AND tenant_id = $2 AND is_deleted = false
	`
	rows, err := r.pool.Query(ctx, itemQuery, id, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item invoicing.InvoiceItem
		err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice, &item.Amount, &item.CreatedAt, &item.CreatedBy, &item.UpdatedAt, &item.UpdatedBy)
		if err != nil {
			return nil, err
		}
		inv.Items = append(inv.Items, item)
	}

	return inv, nil
}

func (r *InvoiceRepository) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error) {
	countQuery := `SELECT COUNT(*) FROM invoices WHERE tenant_id = $1 AND is_deleted = false`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, COALESCE(invoice_number, '') as invoice_number, 
		       COALESCE(customer_name, '') as customer_name, 
		       status, total_amount, due_date, created_at, 
		       COALESCE(created_by, '00000000-0000-0000-0000-000000000000') as created_by, 
		       updated_at, COALESCE(updated_by, '00000000-0000-0000-0000-000000000000') as updated_by
		FROM invoices
		WHERE tenant_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []invoicing.Invoice
	for rows.Next() {
		var inv invoicing.Invoice
		err := rows.Scan(&inv.ID, &inv.OrgID, &inv.InvoiceNumber, &inv.CustomerName, &inv.Status, &inv.TotalAmount, &inv.DueDate, &inv.CreatedAt, &inv.CreatedBy, &inv.UpdatedAt, &inv.UpdatedBy)
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status invoicing.InvoiceStatus, updatedBy uuid.UUID) error {
	query := `UPDATE invoices SET status = $1, updated_at = CURRENT_TIMESTAMP, updated_by = $2 WHERE id = $3 AND tenant_id = $4 AND is_deleted = false`
	_, err := r.pool.Exec(ctx, query, status, updatedBy, id, orgID)
	return err
}

func (r *InvoiceRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	query := `UPDATE invoices SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, orgID)
	return err
}
