package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
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

	query := `SELECT id, org_id, invoice_number, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due, created_at, updated_at
	          FROM invoices WHERE org_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []models.Invoice
	for rows.Next() {
		var i models.Invoice
		err := rows.Scan(&i.ID, &i.OrgID, &i.InvoiceNumber, &i.CustomerName, &i.InvoiceDate, &i.DueDate, &i.Status, &i.Currency, &i.TotalAmount, &i.AmountPaid, &i.BalanceDue, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, i)
	}

	start2 := time.Now()
	var total int
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM invoices WHERE org_id = $1 AND deleted_at IS NULL", orgID).Scan(&total)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM invoices WHERE org_id = $1 AND deleted_at IS NULL", time.Since(start2))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	return invoices, total, nil
}

func (r *BillingRepository) CreateInvoice(ctx context.Context, inv *models.Invoice) error {
	startTx := time.Now()
	tx, err := r.db.Begin(ctx)
	metrics.RecordDBQuery("BEGIN", time.Since(startTx))
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	inv.ID = uuid.New()
	query := `INSERT INTO invoices (id, org_id, invoice_number, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due, notes, custom_fields)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING created_at, updated_at`

	start1 := time.Now()
	err = tx.QueryRow(ctx, query, inv.ID, inv.OrgID, inv.InvoiceNumber, inv.CustomerName, inv.InvoiceDate, inv.DueDate, inv.Status, inv.Currency, inv.TotalAmount, inv.AmountPaid, inv.BalanceDue, inv.Notes, inv.CustomFields).Scan(&inv.CreatedAt, &inv.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start1))
	if err != nil {
		return fmt.Errorf("failed to insert invoice: %w", err)
	}

	// Batch insert invoice items to avoid N+1 queries
	if len(inv.Items) > 0 {
		var args []interface{}
		var placeholders []string
		for i, item := range inv.Items {
			item.ID = uuid.New()
			item.InvoiceID = inv.ID
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
			args = append(args, item.ID, item.InvoiceID, item.Description, item.Quantity, item.UnitPrice, item.TaxPercent, item.Total)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO invoice_items (id, invoice_id, description, quantity, unit_price, tax_percent, total) VALUES %s RETURNING created_at, updated_at",
			strings.Join(placeholders, ", "),
		)
		start2 := time.Now()
		rows, err := tx.Query(ctx, batchQuery, args...)
		metrics.RecordDBQuery(batchQuery, time.Since(start2))
		if err != nil {
			return fmt.Errorf("failed to batch insert invoice items: %w", err)
		}
		defer rows.Close()
		// Scan results to populate timestamps
		for i := range inv.Items {
			if rows.Next() {
				rows.Scan(&inv.Items[i].CreatedAt, &inv.Items[i].UpdatedAt)
			}
		}
	}

	start3 := time.Now()
	if err := tx.Commit(ctx); err != nil {
		metrics.RecordDBQuery("COMMIT", time.Since(start3))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	metrics.RecordDBQuery("COMMIT", time.Since(start3))
	return nil
}

func (r *BillingRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	var i models.Invoice
	var notes sql.NullString
	var customFields []byte
	query := `SELECT id, org_id, invoice_number, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due, notes, custom_fields, created_at, updated_at
	          FROM invoices WHERE id = $1 AND deleted_at IS NULL`

	start := time.Now()
	err := r.db.QueryRow(ctx, query, id).Scan(&i.ID, &i.OrgID, &i.InvoiceNumber, &i.CustomerName, &i.InvoiceDate, &i.DueDate, &i.Status, &i.Currency, &i.TotalAmount, &i.AmountPaid, &i.BalanceDue, &notes, &customFields, &i.CreatedAt, &i.UpdatedAt)
	if notes.Valid {
		i.Notes = notes.String
	}
	if len(customFields) > 0 {
		i.CustomFields = make(map[string]interface{})
		json.Unmarshal(customFields, &i.CustomFields)
	}
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Fetch items
	itemsQuery := "SELECT id, invoice_id, description, quantity, unit_price, tax_percent, total, created_at, updated_at FROM invoice_items WHERE invoice_id = $1 ORDER BY id"
	start2 := time.Now()
	rows, err := r.db.Query(ctx, itemsQuery, id)
	metrics.RecordDBQuery(itemsQuery, time.Since(start2))
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.InvoiceItem
		err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice, &item.TaxPercent, &item.Total, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice item: %w", err)
		}
		i.Items = append(i.Items, item)
	}

	return &i, nil
}

func (r *BillingRepository) UpdateInvoice(ctx context.Context, inv *models.Invoice) error {
	startTx := time.Now()
	tx, err := r.db.Begin(ctx)
	metrics.RecordDBQuery("BEGIN", time.Since(startTx))
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `UPDATE invoices SET customer_name = $1, invoice_date = $2, due_date = $3, status = $4, currency = $5, total_amount = $6, amount_paid = $7, balance_due = $8, notes = $9, custom_fields = $10, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $11`

	start1 := time.Now()
	_, err = tx.Exec(ctx, query, inv.CustomerName, inv.InvoiceDate, inv.DueDate, inv.Status, inv.Currency, inv.TotalAmount, inv.AmountPaid, inv.BalanceDue, inv.Notes, inv.CustomFields, inv.ID)
	metrics.RecordDBQuery(query, time.Since(start1))
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// Delete existing items and recreate
	start2 := time.Now()
	_, err = tx.Exec(ctx, "DELETE FROM invoice_items WHERE invoice_id = $1", inv.ID)
	metrics.RecordDBQuery("DELETE FROM invoice_items WHERE invoice_id = $1", time.Since(start2))
	if err != nil {
		return fmt.Errorf("failed to delete invoice items: %w", err)
	}

	// Batch insert invoice items to avoid N+1 queries
	if len(inv.Items) > 0 {
		var args []interface{}
		var placeholders []string
		for i, item := range inv.Items {
			item.ID = uuid.New()
			item.InvoiceID = inv.ID
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
			args = append(args, item.ID, item.InvoiceID, item.Description, item.Quantity, item.UnitPrice, item.TaxPercent, item.Total)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO invoice_items (id, invoice_id, description, quantity, unit_price, tax_percent, total) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		start3 := time.Now()
		_, err = tx.Exec(ctx, batchQuery, args...)
		metrics.RecordDBQuery(batchQuery, time.Since(start3))
		if err != nil {
			return fmt.Errorf("failed to batch insert invoice items: %w", err)
		}
	}

	start4 := time.Now()
	if err := tx.Commit(ctx); err != nil {
		metrics.RecordDBQuery("COMMIT", time.Since(start4))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	metrics.RecordDBQuery("COMMIT", time.Since(start4))
	return nil
}

func (r *BillingRepository) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	start := time.Now()
	_, err := r.db.Exec(ctx, "UPDATE invoices SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	metrics.RecordDBQuery("UPDATE invoices SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

func (r *BillingRepository) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error) {
	var s models.BillingSettings
	query := `SELECT org_id, default_tax_rate, base_currency, grace_period_days, auto_invoicing_enabled, tax_country, tax_state, invoice_prefix, next_invoice_number, render_config, created_at, updated_at
	          FROM billing_settings WHERE org_id = $1`

	start := time.Now()
	err := r.db.QueryRow(ctx, query, orgID).Scan(&s.OrgID, &s.DefaultTaxRate, &s.BaseCurrency, &s.GracePeriodDays, &s.AutoInvoicingEnabled, &s.TaxCountry, &s.TaxState, &s.InvoicePrefix, &s.NextInvoiceNumber, &s.RenderConfig, &s.CreatedAt, &s.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get billing settings: %w", err)
	}
	return &s, nil
}

func (r *BillingRepository) UpdateSettings(ctx context.Context, s *models.BillingSettings) error {
	query := `INSERT INTO billing_settings (org_id, default_tax_rate, base_currency, grace_period_days, auto_invoicing_enabled, tax_country, tax_state, invoice_prefix, next_invoice_number, render_config, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP)
	          ON CONFLICT (org_id) DO UPDATE SET
	          default_tax_rate = EXCLUDED.default_tax_rate,
	          base_currency = EXCLUDED.base_currency,
	          grace_period_days = EXCLUDED.grace_period_days,
	          auto_invoicing_enabled = EXCLUDED.auto_invoicing_enabled,
	          tax_country = EXCLUDED.tax_country,
	          tax_state = EXCLUDED.tax_state,
	          invoice_prefix = EXCLUDED.invoice_prefix,
	          next_invoice_number = EXCLUDED.next_invoice_number,
	          render_config = EXCLUDED.render_config,
	          updated_at = EXCLUDED.updated_at`

	start := time.Now()
	_, err := r.db.Exec(ctx, query, s.OrgID, s.DefaultTaxRate, s.BaseCurrency, s.GracePeriodDays, s.AutoInvoicingEnabled, s.TaxCountry, s.TaxState, s.InvoicePrefix, s.NextInvoiceNumber, s.RenderConfig)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to update billing settings: %w", err)
	}
	return nil
}

func (r *BillingRepository) GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error) {
	var m models.BillingMetrics

	// Total Invoiced MTD
	mtdQuery := `SELECT COALESCE(SUM(total_amount), 0) FROM invoices WHERE org_id = $1 AND invoice_date >= date_trunc('month', CURRENT_DATE) AND deleted_at IS NULL`
	start1 := time.Now()
	err := r.db.QueryRow(ctx, mtdQuery, orgID).Scan(&m.TotalInvoicedMTD)
	metrics.RecordDBQuery(mtdQuery, time.Since(start1))
	if err != nil {
		return nil, fmt.Errorf("failed to get MTD invoiced: %w", err)
	}

	// Pending Payment
	pendingQuery := `SELECT COALESCE(SUM(balance_due), 0), COUNT(*) FROM invoices WHERE org_id = $1 AND (status = 'PENDING' OR status = 'PARTIAL') AND deleted_at IS NULL`
	start2 := time.Now()
	err = r.db.QueryRow(ctx, pendingQuery, orgID).Scan(&m.PendingPayment, &m.OpenCount)
	metrics.RecordDBQuery(pendingQuery, time.Since(start2))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}

	// Overdue Amount
	overdueQuery := `SELECT COALESCE(SUM(balance_due), 0) FROM invoices WHERE org_id = $1 AND status = 'OVERDUE' AND deleted_at IS NULL`
	start3 := time.Now()
	err = r.db.QueryRow(ctx, overdueQuery, orgID).Scan(&m.OverdueAmount)
	metrics.RecordDBQuery(overdueQuery, time.Since(start3))
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue amount: %w", err)
	}

	// Paid Count (last 30 days)
	paidQuery := `SELECT COUNT(*) FROM invoices WHERE org_id = $1 AND status = 'PAID' AND updated_at >= CURRENT_DATE - INTERVAL '30 days' AND deleted_at IS NULL`
	start4 := time.Now()
	err = r.db.QueryRow(ctx, paidQuery, orgID).Scan(&m.PaidCount)
	metrics.RecordDBQuery(paidQuery, time.Since(start4))
	if err != nil {
		return nil, fmt.Errorf("failed to get paid count: %w", err)
	}

	return &m, nil
}

func (r *BillingRepository) RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	startTx := time.Now()
	tx, err := r.db.Begin(ctx)
	metrics.RecordDBQuery("BEGIN", time.Since(startTx))
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	start1 := time.Now()
	var inv models.Invoice
	err = tx.QueryRow(ctx, "SELECT total_amount, amount_paid FROM invoices WHERE id = $1 FOR UPDATE", id).Scan(&inv.TotalAmount, &inv.AmountPaid)
	metrics.RecordDBQuery("SELECT total_amount, amount_paid FROM invoices WHERE id = $1 FOR UPDATE", time.Since(start1))
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	inv.AmountPaid += amount
	inv.BalanceDue = inv.TotalAmount - inv.AmountPaid
	if inv.BalanceDue <= 0 {
		inv.BalanceDue = 0
		inv.Status = models.StatusPaid
	} else {
		inv.Status = models.StatusPartial
	}

	start2 := time.Now()
	_, err = tx.Exec(ctx, "UPDATE invoices SET amount_paid = $1, balance_due = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4", inv.AmountPaid, inv.BalanceDue, inv.Status, id)
	metrics.RecordDBQuery("UPDATE invoices SET amount_paid = $1, balance_due = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4", time.Since(start2))
	if err != nil {
		return fmt.Errorf("failed to update invoice payment: %w", err)
	}

	// Record payment in audit trail
	paymentQuery := `INSERT INTO payment_records (id, invoice_id, amount, payment_method, payment_date, created_by) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, $5)`
	start3 := time.Now()
	_, err = tx.Exec(ctx, paymentQuery, uuid.New(), id, amount, "MANUAL", uuid.Nil)
	metrics.RecordDBQuery(paymentQuery, time.Since(start3))
	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	start4 := time.Now()
	if err := tx.Commit(ctx); err != nil {
		metrics.RecordDBQuery("COMMIT", time.Since(start4))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	metrics.RecordDBQuery("COMMIT", time.Since(start4))
	return nil
}
func (r *BillingRepository) CreateTaxLines(ctx context.Context, lines []models.InvoiceTaxLine) error {
	startTx := time.Now()
	tx, err := r.db.Begin(ctx)
	metrics.RecordDBQuery("BEGIN", time.Since(startTx))
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Batch insert tax lines to avoid N+1 queries
	if len(lines) > 0 {
		var args []interface{}
		var placeholders []string
		for i, line := range lines {
			if line.ID == uuid.Nil {
				line.ID = uuid.New()
			}
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
			args = append(args, line.ID, line.InvoiceID, line.TaxRateID, line.Name, line.Amount, line.Percentage)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO billing_invoice_tax_lines (id, invoice_id, tax_rate_id, name, amount, percentage) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		start1 := time.Now()
		_, err = tx.Exec(ctx, batchQuery, args...)
		metrics.RecordDBQuery(batchQuery, time.Since(start1))
		if err != nil {
			return fmt.Errorf("failed to batch insert tax lines: %w", err)
		}
	}

	start2 := time.Now()
	if err := tx.Commit(ctx); err != nil {
		metrics.RecordDBQuery("COMMIT", time.Since(start2))
		return err
	}
	metrics.RecordDBQuery("COMMIT", time.Since(start2))
	return nil
}
