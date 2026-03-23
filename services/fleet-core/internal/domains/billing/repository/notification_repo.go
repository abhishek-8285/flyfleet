package repository

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// RecordPaymentAttempt stores a single retry/attempt for an invoice
func (r *NotificationRepository) RecordPaymentAttempt(ctx context.Context, attempt *models.PaymentAttempt) error {
	if attempt.ID == uuid.Nil {
		attempt.ID = uuid.New()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO billing_payment_attempts (id, org_id, invoice_id, attempt_number, status, failure_reason, processor_response, amount, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, attempt.ID, attempt.OrgID, attempt.InvoiceID, attempt.AttemptNumber, attempt.Status, attempt.FailureReason, attempt.ProcessorResponse, attempt.Amount, attempt.Currency)

	return err
}

// GetLatestAttemptNumber returns the count of previous attempts for an invoice
func (r *NotificationRepository) GetLatestAttemptNumber(ctx context.Context, invoiceID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(attempt_number), 0) FROM billing_payment_attempts WHERE invoice_id = $1`, invoiceID).Scan(&count)
	return count, err
}
