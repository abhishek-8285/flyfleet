package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Log(ctx context.Context, entry models.AuditLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (user_id, tenant_id, action, resource, data, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.UserID, entry.TenantID, entry.Action, entry.Resource,
		entry.Data, entry.IP, entry.UserAgent,
	)
	return err
}

func (r *AuditRepository) GetUserAuditLog(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, tenant_id, action, resource, data,
			   ip_address::text, user_agent, created_at
		FROM audit_logs WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.AuditLog
	for rows.Next() {
		var e models.AuditLog
		if err := rows.Scan(&e.ID, &e.UserID, &e.TenantID, &e.Action, &e.Resource,
			&e.Data, &e.IP, &e.UserAgent, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
