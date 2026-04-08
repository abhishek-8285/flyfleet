package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreditNoteRepository handles credit note persistence
type CreditNoteRepository struct {
	db *pgxpool.Pool
}

// NewCreditNoteRepository creates a new credit note repository
func NewCreditNoteRepository(db *pgxpool.Pool) *CreditNoteRepository {
	return &CreditNoteRepository{db: db}
}

// Create creates a new credit note with validation
func (r *CreditNoteRepository) Create(ctx context.Context, cn *models.CreditNote, items []models.CreditNoteItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// BUSINESS LOGIC: Validate original invoice exists and is issued
	var invoiceStatus string
	err = tx.QueryRow(ctx, "SELECT status FROM invoices WHERE id = $1", cn.OriginalInvoiceID).Scan(&invoiceStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("original invoice not found")
		}
		return fmt.Errorf("failed to get original invoice: %w", err)
	}

	if invoiceStatus != "ISSUED" && invoiceStatus != "PAID" && invoiceStatus != "PARTIAL" {
		return errors.New("can only create credit note for issued/paid invoices")
	}

	// BUSINESS LOGIC: Validate credit note amount doesn't exceed outstanding
	var outstanding float64
	err = tx.QueryRow(ctx, "SELECT balance_due FROM invoices WHERE id = $1", cn.OriginalInvoiceID).Scan(&outstanding)
	if err != nil {
		return fmt.Errorf("failed to get outstanding amount: %w", err)
	}

	// Get sum of existing credit notes
	var existingCredits float64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_amount), 0) 
		FROM credit_notes 
		WHERE original_invoice_id = $1 AND status != 'cancelled' AND deleted_at IS NULL`,
		cn.OriginalInvoiceID).Scan(&existingCredits)
	if err != nil {
		return fmt.Errorf("failed to get existing credits: %w", err)
	}

	if cn.TotalAmount > (outstanding - existingCredits) {
		return errors.New("credit note amount exceeds outstanding invoice balance")
	}

	// Insert credit note
	cn.ID = uuid.New()
	cn.CreatedAt = time.Now()
	cn.UpdatedAt = time.Now()

	query := `
		INSERT INTO credit_notes (
			id, credit_note_no, org_id, original_invoice_id, customer_id, customer_name,
			customer_gstin, customer_pan, taxable_value, cgst, sgst, igst, total_amount,
			status, reason, remarks, issue_date, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING created_at, updated_at`

	err = tx.QueryRow(ctx, query,
		cn.ID, cn.CreditNoteNo, cn.OrgID, cn.OriginalInvoiceID, cn.CustomerID, cn.CustomerName,
		cn.CustomerGSTIN, cn.CustomerPAN, cn.TaxableValue, cn.CGST, cn.SGST, cn.IGST, cn.TotalAmount,
		cn.Status, cn.Reason, cn.Remarks, cn.IssueDate, cn.CreatedBy, cn.UpdatedBy).Scan(&cn.CreatedAt, &cn.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create credit note: %w", err)
	}

	// Batch insert items to avoid N+1 queries
	if len(items) > 0 {
		var args []interface{}
		var placeholders []string
		for i, item := range items {
			item.ID = uuid.New()
			item.CreditNoteID = cn.ID
			item.CreatedAt = time.Now()

			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*10+1, i*10+2, i*10+3, i*10+4, i*10+5, i*10+6, i*10+7, i*10+8, i*10+9, i*10+10))
			args = append(args,
				item.ID, item.CreditNoteID, item.OriginalInvoiceItemID, item.Description, item.HSNCode,
				item.Quantity, item.UnitPrice, item.TaxPercent, item.TaxAmount, item.TotalAmount)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO credit_note_items (id, credit_note_id, original_invoice_item_id, description, hsn_code, "+
				"quantity, unit_price, tax_percent, tax_amount, total_amount) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err = tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert credit note items: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a credit note by ID
func (r *CreditNoteRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CreditNote, error) {
	query := `
		SELECT id, credit_note_no, org_id, original_invoice_id, customer_id, customer_name,
		       customer_gstin, customer_pan, taxable_value, cgst, sgst, igst, total_amount,
		       status, reason, remarks, applied_to_invoice_id, applied_amount, applied_at,
		       issue_date, created_at, created_by, updated_at, updated_by
		FROM credit_notes
		WHERE id = $1 AND deleted_at IS NULL`

	var cn models.CreditNote
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cn.ID, &cn.CreditNoteNo, &cn.OrgID, &cn.OriginalInvoiceID, &cn.CustomerID, &cn.CustomerName,
		&cn.CustomerGSTIN, &cn.CustomerPAN, &cn.TaxableValue, &cn.CGST, &cn.SGST, &cn.IGST, &cn.TotalAmount,
		&cn.Status, &cn.Reason, &cn.Remarks, &cn.AppliedToInvoiceID, &cn.AppliedAmount, &cn.AppliedAt,
		&cn.IssueDate, &cn.CreatedAt, &cn.CreatedBy, &cn.UpdatedAt, &cn.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get credit note: %w", err)
	}

	return &cn, nil
}

// GetItems retrieves items for a credit note
func (r *CreditNoteRepository) GetItems(ctx context.Context, creditNoteID uuid.UUID) ([]models.CreditNoteItem, error) {
	query := `
		SELECT id, credit_note_id, original_invoice_item_id, description, hsn_code,
		       quantity, unit_price, tax_percent, tax_amount, total_amount, created_at
		FROM credit_note_items
		WHERE credit_note_id = $1
		ORDER BY id`

	rows, err := r.db.Query(ctx, query, creditNoteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CreditNoteItem
	for rows.Next() {
		var item models.CreditNoteItem
		err := rows.Scan(
			&item.ID, &item.CreditNoteID, &item.OriginalInvoiceItemID, &item.Description, &item.HSNCode,
			&item.Quantity, &item.UnitPrice, &item.TaxPercent, &item.TaxAmount, &item.TotalAmount, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// Update updates a credit note with validation
func (r *CreditNoteRepository) Update(ctx context.Context, cn *models.CreditNote) error {
	// BUSINESS LOGIC: Cannot edit non-draft credit notes
	existing, err := r.GetByID(ctx, cn.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("credit note not found")
	}

	if !existing.CanEdit() {
		return errors.New("cannot edit credit note that is not in draft status")
	}

	cn.UpdatedAt = time.Now()
	query := `
		UPDATE credit_notes SET
			customer_name = $1, customer_gstin = $2, customer_pan = $3,
			taxable_value = $4, cgst = $5, sgst = $6, igst = $7, total_amount = $8,
			reason = $9, remarks = $10, updated_at = $11, updated_by = $12
		WHERE id = $13`

	_, err = r.db.Exec(ctx, query,
		cn.CustomerName, cn.CustomerGSTIN, cn.CustomerPAN,
		cn.TaxableValue, cn.CGST, cn.SGST, cn.IGST, cn.TotalAmount,
		cn.Reason, cn.Remarks, cn.UpdatedAt, cn.UpdatedBy, cn.ID)

	if err != nil {
		return fmt.Errorf("failed to update credit note: %w", err)
	}

	return nil
}

// Issue issues a credit note (changes status from draft to issued)
func (r *CreditNoteRepository) Issue(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	// BUSINESS LOGIC: Can only issue draft credit notes
	cn, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cn == nil {
		return errors.New("credit note not found")
	}
	if cn.Status != string(models.CreditNoteDraft) {
		return errors.New("can only issue draft credit notes")
	}

	query := `UPDATE credit_notes SET status = 'issued', updated_at = CURRENT_TIMESTAMP, updated_by = $1 WHERE id = $2`
	_, err = r.db.Exec(ctx, query, updatedBy, id)
	return err
}

// Apply applies a credit note to an invoice
func (r *CreditNoteRepository) Apply(ctx context.Context, id uuid.UUID, appliedToInvoiceID uuid.UUID, amount float64, updatedBy uuid.UUID) error {
	// BUSINESS LOGIC: Validate credit note can be applied
	cn, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cn == nil {
		return errors.New("credit note not found")
	}
	if !cn.CanApply() {
		return errors.New("credit note cannot be applied (already applied or not issued)")
	}

	// BUSINESS LOGIC: Amount cannot exceed credit note total
	if amount > cn.TotalAmount {
		return errors.New("application amount exceeds credit note total")
	}

	now := time.Now()
	query := `
		UPDATE credit_notes SET
			status = 'applied',
			applied_to_invoice_id = $1,
			applied_amount = $2,
			applied_at = $3,
			updated_at = $3,
			updated_by = $4
		WHERE id = $5`

	_, err = r.db.Exec(ctx, query, appliedToInvoiceID, amount, now, updatedBy, id)
	return err
}

// Cancel cancels a credit note
func (r *CreditNoteRepository) Cancel(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	// BUSINESS LOGIC: Can only cancel draft or issued credit notes
	cn, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cn == nil {
		return errors.New("credit note not found")
	}
	if !cn.CanCancel() {
		return errors.New("can only cancel draft or issued credit notes")
	}

	query := `UPDATE credit_notes SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP, updated_by = $1 WHERE id = $2`
	_, err = r.db.Exec(ctx, query, updatedBy, id)
	return err
}

// Delete soft-deletes a credit note (only draft)
func (r *CreditNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cn, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cn == nil {
		return errors.New("credit note not found")
	}
	if cn.Status != string(models.CreditNoteDraft) {
		return errors.New("can only delete draft credit notes")
	}

	query := `UPDATE credit_notes SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = r.db.Exec(ctx, query, id)
	return err
}

// ListByOrg lists credit notes for an organization
func (r *CreditNoteRepository) ListByOrg(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.CreditNote, int, error) {
	offset := (page - 1) * size

	countQuery := `SELECT COUNT(*) FROM credit_notes WHERE org_id = $1 AND deleted_at IS NULL`
	var total int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, credit_note_no, org_id, original_invoice_id, customer_id, customer_name,
		       customer_gstin, customer_pan, taxable_value, cgst, sgst, igst, total_amount,
		       status, reason, remarks, applied_to_invoice_id, applied_amount, applied_at,
		       issue_date, created_at, created_by, updated_at, updated_by
		FROM credit_notes
		WHERE org_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var creditNotes []models.CreditNote
	for rows.Next() {
		var cn models.CreditNote
		err := rows.Scan(
			&cn.ID, &cn.CreditNoteNo, &cn.OrgID, &cn.OriginalInvoiceID, &cn.CustomerID, &cn.CustomerName,
			&cn.CustomerGSTIN, &cn.CustomerPAN, &cn.TaxableValue, &cn.CGST, &cn.SGST, &cn.IGST, &cn.TotalAmount,
			&cn.Status, &cn.Reason, &cn.Remarks, &cn.AppliedToInvoiceID, &cn.AppliedAmount, &cn.AppliedAt,
			&cn.IssueDate, &cn.CreatedAt, &cn.CreatedBy, &cn.UpdatedAt, &cn.UpdatedBy)
		if err != nil {
			return nil, 0, err
		}
		creditNotes = append(creditNotes, cn)
	}

	return creditNotes, total, nil
}

// ListByInvoice lists credit notes for a specific invoice
func (r *CreditNoteRepository) ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]models.CreditNote, error) {
	query := `
		SELECT id, credit_note_no, org_id, original_invoice_id, customer_id, customer_name,
		       customer_gstin, customer_pan, taxable_value, cgst, sgst, igst, total_amount,
		       status, reason, remarks, applied_to_invoice_id, applied_amount, applied_at,
		       issue_date, created_at, created_by, updated_at, updated_by
		FROM credit_notes
		WHERE original_invoice_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creditNotes []models.CreditNote
	for rows.Next() {
		var cn models.CreditNote
		err := rows.Scan(
			&cn.ID, &cn.CreditNoteNo, &cn.OrgID, &cn.OriginalInvoiceID, &cn.CustomerID, &cn.CustomerName,
			&cn.CustomerGSTIN, &cn.CustomerPAN, &cn.TaxableValue, &cn.CGST, &cn.SGST, &cn.IGST, &cn.TotalAmount,
			&cn.Status, &cn.Reason, &cn.Remarks, &cn.AppliedToInvoiceID, &cn.AppliedAmount, &cn.AppliedAt,
			&cn.IssueDate, &cn.CreatedAt, &cn.CreatedBy, &cn.UpdatedAt, &cn.UpdatedBy)
		if err != nil {
			return nil, err
		}
		creditNotes = append(creditNotes, cn)
	}

	return creditNotes, nil
}
