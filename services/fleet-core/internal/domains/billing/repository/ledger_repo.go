package repository

import (
	"context"
	"errors"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientFunds = errors.New("insufficient wallet funds")

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) GetOrCreateBalance(ctx context.Context, orgID uuid.UUID, currency string) (*models.CustomerBalance, error) {
	if currency == "" {
		currency = "USD"
	}

	var b models.CustomerBalance
	err := r.db.QueryRow(ctx, `
		SELECT id, org_id, currency, balance, credit_limit, status, created_at, updated_at
		FROM billing_customer_balances
		WHERE org_id = $1 AND currency = $2
	`, orgID, currency).Scan(
		&b.ID, &b.OrgID, &b.Currency, &b.Balance, &b.CreditLimit, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		// Auto-provision wallet
		b.ID = uuid.New()
		b.OrgID = orgID
		b.Currency = currency
		b.Balance = 0
		b.CreditLimit = 0
		b.Status = "ACTIVE"

		_, err = r.db.Exec(ctx, `
			INSERT INTO billing_customer_balances (id, org_id, currency, balance, credit_limit, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, b.ID, b.OrgID, b.Currency, b.Balance, b.CreditLimit, b.Status)
		if err != nil {
			return nil, err
		}
		return &b, nil
	} else if err != nil {
		return nil, err
	}

	return &b, nil
}

func (r *LedgerRepository) ApplyCreditTransaction(ctx context.Context, tx *models.CreditTransaction) error {
	// Must execute in a SQL transaction to lock the balance row
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer dbTx.Rollback(ctx)

	// 1. Lock and get current balance
	var currentBalance float64
	var creditLimit float64
	err = dbTx.QueryRow(ctx, `
		SELECT balance, credit_limit FROM billing_customer_balances 
		WHERE id = $1 FOR UPDATE
	`, tx.BalanceID).Scan(&currentBalance, &creditLimit)
	
	if err != nil {
		return err
	}

	// 2. Check sufficient funds for debits (negative amounts)
	if tx.Amount < 0 {
		if (currentBalance + creditLimit) < -tx.Amount {
			return ErrInsufficientFunds
		}
	}

	// 3. Insert transaction log
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	_, err = dbTx.Exec(ctx, `
		INSERT INTO billing_credit_transactions (id, balance_id, transaction_type, amount, currency, reference_type, reference_id, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, tx.ID, tx.BalanceID, tx.TransactionType, tx.Amount, tx.Currency, tx.ReferenceType, tx.ReferenceID, tx.Description, tx.CreatedBy)
	if err != nil {
		return err
	}

	// 4. Update balance
	_, err = dbTx.Exec(ctx, `
		UPDATE billing_customer_balances
		SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, tx.Amount, tx.BalanceID)
	if err != nil {
		return err
	}

	return dbTx.Commit(ctx)
}

func (r *LedgerRepository) RecordLedgerEntry(ctx context.Context, entry *models.LedgerEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO billing_ledger_entries (id, org_id, entry_type, amount, currency, invoice_id, payment_id, credit_note_id, ledger_account, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, entry.ID, entry.OrgID, entry.EntryType, entry.Amount, entry.Currency, entry.InvoiceID, entry.PaymentID, entry.CreditNoteID, entry.LedgerAccount, entry.Description)
	
	return err
}

func (r *LedgerRepository) CreateRevenueSchedule(ctx context.Context, s *models.RevenueSchedule) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO billing_revenue_schedules (id, org_id, invoice_line_item_id, total_amount, currency, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, s.ID, s.OrgID, s.InvoiceLineItemID, s.TotalAmount, s.Currency, s.StartDate, s.EndDate, s.Status)
	return err
}

func (r *LedgerRepository) GetEntriesByOrg(ctx context.Context, orgID uuid.UUID, limit int) ([]models.LedgerEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, entry_type, amount, currency, invoice_id, payment_id, credit_note_id, ledger_account, description, created_at
		FROM billing_ledger_entries
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LedgerEntry
	for rows.Next() {
		var e models.LedgerEntry
		err := rows.Scan(
			&e.ID, &e.OrgID, &e.EntryType, &e.Amount, &e.Currency, &e.InvoiceID, &e.PaymentID, &e.CreditNoteID, &e.LedgerAccount, &e.Description, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
