package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountingRepository struct {
	db *pgxpool.Pool
}

func NewAccountingRepository(db *pgxpool.Pool) *AccountingRepository {
	return &AccountingRepository{db: db}
}

func (r *AccountingRepository) CreateExport(ctx context.Context, export *models.AccountingExport) error {
	if export.ID == uuid.Nil {
		export.ID = uuid.New()
	}
	if export.CreatedAt.IsZero() {
		export.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO billing_accounting_exports (
			id, org_id, external_system, export_type, status, file_key, records_count, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, export.ID, export.OrgID, export.ExternalSystem, export.ExportType, export.Status, export.FileKey, export.RecordsCount, export.CreatedAt, export.CreatedBy)

	return err
}

func (r *AccountingRepository) UpdateExportStatus(ctx context.Context, id uuid.UUID, status models.ExportStatus, fileKey string, recordsCount int, errStr string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE billing_accounting_exports
		SET status = $1, file_key = $2, records_count = $3, error_message = $4, completed_at = $5
		WHERE id = $6
	`, status, fileKey, recordsCount, errStr, now, id)
	return err
}

func (r *AccountingRepository) GetExportHistory(ctx context.Context, orgID uuid.UUID, limit int) ([]models.AccountingExport, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, org_id, external_system, export_type, status, file_key, records_count, error_message, created_at, completed_at, created_by
		FROM billing_accounting_exports
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exports []models.AccountingExport
	for rows.Next() {
		var e models.AccountingExport
		err := rows.Scan(
			&e.ID, &e.OrgID, &e.ExternalSystem, &e.ExportType, &e.Status, &e.FileKey, &e.RecordsCount, &e.ErrorMessage, &e.CreatedAt, &e.CompletedAt, &e.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		exports = append(exports, e)
	}
	return exports, nil
}
