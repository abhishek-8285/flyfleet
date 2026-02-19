package postgres

import (
	"context"
	"fmt"

	"github.com/flyfleet/billing-service/internal/domain/invoice"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepository struct {
	pool *pgxpool.Pool
}

func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{pool: pool}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *invoice.Invoice) (*invoice.Invoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO invoices (
			org_id, invoice_number, customer_name, customer_email, billing_address,
			due_date, status, subtotal, tax_amount, total_amount, currency, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		inv.OrgID, inv.InvoiceNumber, inv.CustomerName, inv.CustomerEmail, inv.BillingAddress,
		inv.DueDate, inv.Status, inv.Subtotal, inv.TaxAmount, inv.TotalAmount, inv.Currency, inv.Notes,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert invoice: %w", err)
	}

	for i := range inv.Items {
		item := &inv.Items[i]
		itemQuery := `
			INSERT INTO invoice_items (
				org_id, invoice_id, description, quantity, unit_price, amount
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at
		`
		err = tx.QueryRow(ctx, itemQuery,
			inv.OrgID, inv.ID, item.Description, item.Quantity, item.UnitPrice, item.Amount,
		).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to insert invoice item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return inv, nil
}

func (r *InvoiceRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*invoice.Invoice, error) {
	query := `
		SELECT id, org_id, invoice_number, customer_name, customer_email, billing_address,
		       due_date, status, subtotal, tax_amount, total_amount, amount_paid, amount_remaining,
		       currency, notes, created_at, updated_at
		FROM invoices
		WHERE id = $1 AND org_id = $2 AND is_deleted = false
	`
	inv := &invoice.Invoice{}
	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(
		&inv.ID, &inv.OrgID, &inv.InvoiceNumber, &inv.CustomerName, &inv.CustomerEmail, &inv.BillingAddress,
		&inv.DueDate, &inv.Status, &inv.Subtotal, &inv.TaxAmount, &inv.TotalAmount, &inv.AmountPaid, &inv.AmountRemaining,
		&inv.Currency, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Fetch items
	itemQuery := `
		SELECT id, invoice_id, description, quantity, unit_price, amount, created_at, updated_at
		FROM invoice_items
		WHERE invoice_id = $1 AND org_id = $2 AND is_deleted = false
	`
	rows, err := r.pool.Query(ctx, itemQuery, id, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item invoice.InvoiceItem
		err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice, &item.Amount, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inv.Items = append(inv.Items, item)
	}

	return inv, nil
}

func (r *InvoiceRepository) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]invoice.Invoice, int64, error) {
	countQuery := `SELECT COUNT(*) FROM invoices WHERE org_id = $1 AND is_deleted = false`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, org_id, invoice_number, customer_name, status, total_amount, due_date, created_at
		FROM invoices
		WHERE org_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []invoice.Invoice
	for rows.Next() {
		var inv invoice.Invoice
		err := rows.Scan(&inv.ID, &inv.OrgID, &inv.InvoiceNumber, &inv.CustomerName, &inv.Status, &inv.TotalAmount, &inv.DueDate, &inv.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status invoice.InvoiceStatus) error {
	query := `UPDATE invoices SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND org_id = $3 AND is_deleted = false`
	_, err := r.pool.Exec(ctx, query, status, id, orgID)
	return err
}

func (r *InvoiceRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	query := `UPDATE invoices SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND org_id = $2`
	_, err := r.pool.Exec(ctx, query, id, orgID)
	return err
}

