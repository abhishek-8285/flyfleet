package repository

import (
	"context"

	"github.com/flyfleet/billing-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository struct {
	db *pgxpool.Pool
}

func NewBillingRepository(db *pgxpool.Pool) *BillingRepository {
	return &BillingRepository{db: db}
}

func (r *BillingRepository) ListInvoices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error) {
	offset := (page - 1) * size
	
	query := `SELECT id, org_id, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due, created_at, updated_at 
	          FROM invoices WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []models.Invoice
	for rows.Next() {
		var i models.Invoice
		err := rows.Scan(&i.ID, &i.OrgID, &i.CustomerName, &i.InvoiceDate, &i.DueDate, &i.Status, &i.Currency, &i.TotalAmount, &i.AmountPaid, &i.BalanceDue, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, i)
	}

	var total int
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM invoices WHERE org_id = $1", orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (r *BillingRepository) CreateInvoice(ctx context.Context, inv *models.Invoice) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	inv.ID = uuid.New()
	query := `INSERT INTO invoices (id, org_id, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING created_at, updated_at`
	
	err = tx.QueryRow(ctx, query, inv.ID, inv.OrgID, inv.CustomerName, inv.InvoiceDate, inv.DueDate, inv.Status, inv.Currency, inv.TotalAmount, inv.AmountPaid, inv.BalanceDue).Scan(&inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range inv.Items {
		item := &inv.Items[i]
		item.ID = uuid.New()
		item.InvoiceID = inv.ID
		itemQuery := `INSERT INTO invoice_items (id, invoice_id, description, quantity, unit_price, tax_percent, total) 
		              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
		err = tx.QueryRow(ctx, itemQuery, item.ID, item.InvoiceID, item.Description, item.Quantity, item.UnitPrice, item.TaxPercent, item.Total).Scan(&item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *BillingRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	var i models.Invoice
	query := `SELECT id, org_id, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due, created_at, updated_at 
	          FROM invoices WHERE id = $1`
	
	err := r.db.QueryRow(ctx, query, id).Scan(&i.ID, &i.OrgID, &i.CustomerName, &i.InvoiceDate, &i.DueDate, &i.Status, &i.Currency, &i.TotalAmount, &i.AmountPaid, &i.BalanceDue, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Fetch items
	rows, err := r.db.Query(ctx, "SELECT id, invoice_id, description, quantity, unit_price, tax_percent, total, created_at, updated_at FROM invoice_items WHERE invoice_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.InvoiceItem
		err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice, &item.TaxPercent, &item.Total, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		i.Items = append(i.Items, item)
	}

	return &i, nil
}

func (r *BillingRepository) UpdateInvoice(ctx context.Context, inv *models.Invoice) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE invoices SET customer_name = $1, invoice_date = $2, due_date = $3, status = $4, currency = $5, total_amount = $6, amount_paid = $7, balance_due = $8, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $9`
	
	_, err = tx.Exec(ctx, query, inv.CustomerName, inv.InvoiceDate, inv.DueDate, inv.Status, inv.Currency, inv.TotalAmount, inv.AmountPaid, inv.BalanceDue, inv.ID)
	if err != nil {
		return err
	}

	// Simple approach for items: delete and recreate (can be optimized)
	_, err = tx.Exec(ctx, "DELETE FROM invoice_items WHERE invoice_id = $1", inv.ID)
	if err != nil {
		return err
	}

	for i := range inv.Items {
		item := &inv.Items[i]
		item.ID = uuid.New()
		item.InvoiceID = inv.ID
		itemQuery := `INSERT INTO invoice_items (id, invoice_id, description, quantity, unit_price, tax_percent, total) 
		              VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err = tx.Exec(ctx, itemQuery, item.ID, item.InvoiceID, item.Description, item.Quantity, item.UnitPrice, item.TaxPercent, item.Total)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *BillingRepository) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM invoices WHERE id = $1", id)
	return err
}

func (r *BillingRepository) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error) {
	var s models.BillingSettings
	query := `SELECT org_id, default_tax_rate, base_currency, grace_period_days, auto_invoicing_enabled, created_at, updated_at 
	          FROM billing_settings WHERE org_id = $1`
	
	err := r.db.QueryRow(ctx, query, orgID).Scan(&s.OrgID, &s.DefaultTaxRate, &s.BaseCurrency, &s.GracePeriodDays, &s.AutoInvoicingEnabled, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Initialize default settings
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *BillingRepository) UpdateSettings(ctx context.Context, s *models.BillingSettings) error {
	query := `INSERT INTO billing_settings (org_id, default_tax_rate, base_currency, grace_period_days, auto_invoicing_enabled, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP) 
	          ON CONFLICT (org_id) DO UPDATE SET 
	          default_tax_rate = EXCLUDED.default_tax_rate, 
	          base_currency = EXCLUDED.base_currency, 
	          grace_period_days = EXCLUDED.grace_period_days, 
	          auto_invoicing_enabled = EXCLUDED.auto_invoicing_enabled, 
	          updated_at = EXCLUDED.updated_at`
	
	_, err := r.db.Exec(ctx, query, s.OrgID, s.DefaultTaxRate, s.BaseCurrency, s.GracePeriodDays, s.AutoInvoicingEnabled)
	return err
}

func (r *BillingRepository) GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error) {
	var m models.BillingMetrics
	
	// Total Invoiced MTD
	mtdQuery := `SELECT COALESCE(SUM(total_amount), 0) FROM invoices WHERE org_id = $1 AND invoice_date >= date_trunc('month', CURRENT_DATE)`
	err := r.db.QueryRow(ctx, mtdQuery, orgID).Scan(&m.TotalInvoicedMTD)
	if err != nil { return nil, err }

	// Pending Payment
	pendingQuery := `SELECT COALESCE(SUM(balance_due), 0), COUNT(*) FROM invoices WHERE org_id = $1 AND (status = 'PENDING' OR status = 'PARTIAL')`
	err = r.db.QueryRow(ctx, pendingQuery, orgID).Scan(&m.PendingPayment, &m.OpenCount)
	if err != nil { return nil, err }

	// Overdue Amount
	overdueQuery := `SELECT COALESCE(SUM(balance_due), 0) FROM invoices WHERE org_id = $1 AND status = 'OVERDUE'`
	err = r.db.QueryRow(ctx, overdueQuery, orgID).Scan(&m.OverdueAmount)
	if err != nil { return nil, err }

	// Paid Count
	paidQuery := `SELECT COUNT(*) FROM invoices WHERE org_id = $1 AND status = 'PAID' AND updated_at >= CURRENT_DATE - INTERVAL '30 days'`
	err = r.db.QueryRow(ctx, paidQuery, orgID).Scan(&m.PaidCount)
	if err != nil { return nil, err }

	return &m, nil
}

func (r *BillingRepository) RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	var inv models.Invoice
	err = tx.QueryRow(ctx, "SELECT total_amount, amount_paid FROM invoices WHERE id = $1 FOR UPDATE", id).Scan(&inv.TotalAmount, &inv.AmountPaid)
	if err != nil { return err }

	inv.AmountPaid += amount
	inv.BalanceDue = inv.TotalAmount - inv.AmountPaid
	if inv.BalanceDue <= 0 {
		inv.BalanceDue = 0
		inv.Status = models.StatusPaid
	} else {
		inv.Status = models.StatusPartial
	}

	_, err = tx.Exec(ctx, "UPDATE invoices SET amount_paid = $1, balance_due = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4", inv.AmountPaid, inv.BalanceDue, inv.Status, id)
	if err != nil { return err }

	return tx.Commit(ctx)
}
