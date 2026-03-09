package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLogRepository handles audit log persistence
type AuditLogRepository struct {
	db *pgxpool.Pool
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// LogInvoiceChange creates an audit log entry for an invoice change
func (r *AuditLogRepository) LogInvoiceChange(ctx context.Context, log *models.InvoiceAuditLog) error {
	query := `
		INSERT INTO invoice_audit_log (
			id, invoice_id, action, old_value, new_value,
			changed_by, changed_by_email, changed_at, ip_address, user_agent, reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.InvoiceID, log.Action, log.OldValue, log.NewValue,
		log.ChangedBy, log.ChangedByEmail, log.ChangedAt, log.IPAddress, log.UserAgent, log.Reason,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogsByInvoice retrieves all audit logs for an invoice
func (r *AuditLogRepository) GetAuditLogsByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]models.InvoiceAuditLog, error) {
	query := `
		SELECT id, invoice_id, action, old_value, new_value,
		       changed_by, changed_by_email, changed_at, ip_address, user_agent, reason
		FROM invoice_audit_log
		WHERE invoice_id = $1
		ORDER BY changed_at DESC`

	rows, err := r.db.Query(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.InvoiceAuditLog
	for rows.Next() {
		var log models.InvoiceAuditLog
		err := rows.Scan(
			&log.ID, &log.InvoiceID, &log.Action, &log.OldValue, &log.NewValue,
			&log.ChangedBy, &log.ChangedByEmail, &log.ChangedAt, &log.IPAddress, &log.UserAgent, &log.Reason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditLogsByOrg retrieves audit logs for all invoices in an organization
func (r *AuditLogRepository) GetAuditLogsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]models.InvoiceAuditLog, error) {
	query := `
		SELECT ial.id, ial.invoice_id, ial.action, ial.old_value, ial.new_value,
		       ial.changed_by, ial.changed_by_email, ial.changed_at, ial.ip_address, ial.user_agent, ial.reason
		FROM invoice_audit_log ial
		JOIN invoices i ON ial.invoice_id = i.id
		WHERE i.org_id = $1
		ORDER BY ial.changed_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.InvoiceAuditLog
	for rows.Next() {
		var log models.InvoiceAuditLog
		err := rows.Scan(
			&log.ID, &log.InvoiceID, &log.Action, &log.OldValue, &log.NewValue,
			&log.ChangedBy, &log.ChangedByEmail, &log.ChangedAt, &log.IPAddress, &log.UserAgent, &log.Reason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// CreateAuditLogEntry creates an audit log entry with proper serialization
func (r *AuditLogRepository) CreateAuditLogEntry(ctx context.Context, invoiceID uuid.UUID, action models.AuditLogAction, oldValue, newValue interface{}, changedBy uuid.UUID, changedByEmail, ipAddress, userAgent, reason string) error {
	log := &models.InvoiceAuditLog{
		ID:             uuid.New(),
		InvoiceID:      invoiceID,
		Action:         string(action),
		ChangedBy:      changedBy,
		ChangedByEmail: changedByEmail,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		Reason:         reason,
	}

	// Serialize old value
	if oldValue != nil {
		oldValueBytes, err := json.Marshal(oldValue)
		if err != nil {
			return fmt.Errorf("failed to marshal old value: %w", err)
		}
		log.OldValue = oldValueBytes
	}

	// Serialize new value
	if newValue != nil {
		newValueBytes, err := json.Marshal(newValue)
		if err != nil {
			return fmt.Errorf("failed to marshal new value: %w", err)
		}
		log.NewValue = newValueBytes
	}

	return r.LogInvoiceChange(ctx, log)
}

// GetAuditLogSummary retrieves summary of audit logs for an invoice
func (r *AuditLogRepository) GetAuditLogSummary(ctx context.Context, invoiceID uuid.UUID) (*models.AuditLogSummary, error) {
	query := `
		SELECT invoice_id, COUNT(*) as total_changes, MAX(changed_at) as last_changed_at,
		       array_agg(DISTINCT action) as actions_performed
		FROM invoice_audit_log
		WHERE invoice_id = $1
		GROUP BY invoice_id`

	var summary models.AuditLogSummary
	err := r.db.QueryRow(ctx, query, invoiceID).Scan(
		&summary.InvoiceID, &summary.TotalChanges, &summary.LastChangedAt, &summary.ActionsPerformed,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit log summary: %w", err)
	}

	return &summary, nil
}

// ParseIP parses an IP address string
func ParseIP(ipStr string) net.IP {
	if ipStr == "" {
		return nil
	}
	return net.ParseIP(ipStr)
}
