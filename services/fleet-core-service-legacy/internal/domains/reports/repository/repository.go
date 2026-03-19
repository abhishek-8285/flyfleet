package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/reports/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateReport(ctx context.Context, report *models.ScheduledReport) error
	GetReport(ctx context.Context, id uuid.UUID) (*models.ScheduledReport, error)
	UpdateReport(ctx context.Context, report *models.ScheduledReport) error
	ListReports(ctx context.Context, page, limit int) ([]models.ScheduledReport, int, error)
	PauseReport(ctx context.Context, id uuid.UUID) error
	ResumeReport(ctx context.Context, id uuid.UUID) error
	
	CreateExecution(ctx context.Context, exec *models.ReportExecution) error
	UpdateExecutionStatus(ctx context.Context, id uuid.UUID, status string, errorMessage *string, fileURL *string, updatedBy uuid.UUID) error
	ListExecutions(ctx context.Context, reportID uuid.UUID, page, limit int) ([]models.ReportExecution, int, error)
	
	GetSettings(ctx context.Context) (*models.GlobalSettings, error)
	UpdateSettings(ctx context.Context, settings *models.GlobalSettings) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateReport(ctx context.Context, report *models.ScheduledReport) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO scheduled_reports (id, organization_id, name, report_type, format, frequency, cron_expression, delivery_time, timezone, filters, status, created_by, updated_by, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	
	_, err = tx.Exec(ctx, query, report.ID, report.OrganizationID, report.Name, report.ReportType, report.Format, report.Frequency, report.CronExpression, report.DeliveryTime, report.Timezone, report.Filters, report.Status, report.CreatedBy, report.UpdatedBy, report.CreatedAt, report.UpdatedAt)
	if err != nil {
		return err
	}

	for _, recipient := range report.Recipients {
		_, err = tx.Exec(ctx, "INSERT INTO report_recipients (report_id, email) VALUES ($1, $2)", report.ID, recipient)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetReport(ctx context.Context, id uuid.UUID) (*models.ScheduledReport, error) {
	var report models.ScheduledReport
	query := `SELECT id, organization_id, name, report_type, format, frequency, cron_expression, delivery_time::text, timezone, filters, status, created_by, updated_by, created_at, updated_at
	          FROM scheduled_reports WHERE id = $1`
	
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&report.ID, &report.OrganizationID, &report.Name, &report.ReportType, &report.Format, &report.Frequency, &report.CronExpression, &report.DeliveryTime, &report.Timezone, &report.Filters, &report.Status, &report.CreatedBy, &report.UpdatedBy, &report.CreatedAt, &report.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Fetch recipients
	rows, err := r.pool.Query(ctx, "SELECT email FROM report_recipients WHERE report_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		report.Recipients = append(report.Recipients, email)
	}

	return &report, nil
}

func (r *PostgresRepository) UpdateReport(ctx context.Context, report *models.ScheduledReport) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE scheduled_reports SET name=$1, format=$2, frequency=$3, cron_expression=$4, filters=$5, updated_by=$6, updated_at=NOW() WHERE id=$7`
	_, err = tx.Exec(ctx, query, report.Name, report.Format, report.Frequency, report.CronExpression, report.Filters, report.UpdatedBy, report.ID)
	if err != nil {
		return err
	}

	// Update recipients (delete all and re-insert for simplicity)
	_, err = tx.Exec(ctx, "DELETE FROM report_recipients WHERE report_id = $1", report.ID)
	if err != nil {
		return err
	}

	for _, recipient := range report.Recipients {
		_, err = tx.Exec(ctx, "INSERT INTO report_recipients (report_id, email) VALUES ($1, $2)", report.ID, recipient)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) ListReports(ctx context.Context, page, limit int) ([]models.ScheduledReport, int, error) {
	offset := (page - 1) * limit
	var reports []models.ScheduledReport

	// Count total
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM scheduled_reports").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, report_type, format, frequency, status FROM scheduled_reports ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var rep models.ScheduledReport
		if err := rows.Scan(&rep.ID, &rep.Name, &rep.ReportType, &rep.Format, &rep.Frequency, &rep.Status); err != nil {
			return nil, 0, err
		}
		reports = append(reports, rep)
	}

	return reports, total, nil
}

func (r *PostgresRepository) PauseReport(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE scheduled_reports SET status = 'PAUSED', updated_at = NOW() WHERE id = $1", id)
	return err
}

func (r *PostgresRepository) ResumeReport(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE scheduled_reports SET status = 'ACTIVE', updated_at = NOW() WHERE id = $1", id)
	return err
}

func (r *PostgresRepository) CreateExecution(ctx context.Context, exec *models.ReportExecution) error {
	query := `INSERT INTO report_executions (id, report_id, started_at, status, created_at, created_by, updated_at, updated_by) VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), $5)`
	_, err := r.pool.Exec(ctx, query, exec.ID, exec.ReportID, exec.StartedAt, exec.Status, exec.CreatedBy)
	return err
}

func (r *PostgresRepository) UpdateExecutionStatus(ctx context.Context, id uuid.UUID, status string, errorMessage *string, fileURL *string, updatedBy uuid.UUID) error {
	query := `UPDATE report_executions 
	          SET status = $1, error_message = $2, file_url = $3, completed_at = CASE WHEN $1 IN ('COMPLETED', 'FAILED') THEN NOW() ELSE completed_at END, updated_by = $4, updated_at = NOW()
	          WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, status, errorMessage, fileURL, updatedBy, id)
	return err
}

func (r *PostgresRepository) ListExecutions(ctx context.Context, reportID uuid.UUID, page, limit int) ([]models.ReportExecution, int, error) {
	offset := (page - 1) * limit
	var execs []models.ReportExecution
	
	whereClause := ""
	args := []any{}
	if reportID != uuid.Nil {
		whereClause = "WHERE report_id = $1"
		args = append(args, reportID)
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM report_executions %s", whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Select
	query := fmt.Sprintf(`SELECT id, report_id, started_at, duration_seconds, status, error_message, file_url, created_at, created_by, updated_at, updated_by 
	                      FROM report_executions %s ORDER BY started_at DESC LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)
	
	args = append(args, limit, offset)
	
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.ReportExecution
		if err := rows.Scan(&e.ID, &e.ReportID, &e.StartedAt, &e.DurationSeconds, &e.Status, &e.ErrorMessage, &e.FileURL, &e.CreatedAt, &e.CreatedBy, &e.UpdatedAt, &e.UpdatedBy); err != nil {
			return nil, 0, err
		}
		execs = append(execs, e)
	}
	return execs, total, nil
}

func (r *PostgresRepository) GetSettings(ctx context.Context) (*models.GlobalSettings, error) {
	var s models.GlobalSettings
	// Ensure default exists
	_, err := r.pool.Exec(ctx, "INSERT INTO report_global_settings (id) VALUES (1) ON CONFLICT DO NOTHING")
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx, "SELECT id, max_concurrent_jobs, max_retry_attempts, retention_days, smtp_host, smtp_port, smtp_user, created_at, created_by, updated_at, updated_by FROM report_global_settings WHERE id = 1").Scan(
		&s.ID, &s.MaxConcurrentJobs, &s.MaxRetryAttempts, &s.RetentionDays, &s.SMTPHost, &s.SMTPPort, &s.SMTPUser, &s.CreatedAt, &s.CreatedBy, &s.UpdatedAt, &s.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) UpdateSettings(ctx context.Context, s *models.GlobalSettings) error {
	query := `UPDATE report_global_settings SET max_concurrent_jobs=$1, max_retry_attempts=$2, retention_days=$3, smtp_host=$4, smtp_port=$5, smtp_user=$6, updated_at=NOW(), updated_by=$7 WHERE id=1`
	_, err := r.pool.Exec(ctx, query, s.MaxConcurrentJobs, s.MaxRetryAttempts, s.RetentionDays, s.SMTPHost, s.SMTPPort, s.SMTPUser, s.UpdatedBy)
	return err
}
