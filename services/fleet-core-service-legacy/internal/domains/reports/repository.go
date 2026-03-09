package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository interface {
	CreateReport(ctx context.Context, report *Report) error
	GetReport(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Report, error)
	ListReports(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]Report, error)
	UpdateReportStatus(ctx context.Context, id uuid.UUID, status ReportStatus, progress int, fileURL, errorMessage string) error
	DeleteExpiredReports(ctx context.Context, olderThanDays int) (int64, error)
	ListExpiredReports(ctx context.Context, olderThanDays int) ([]Report, error)
	CountTenantReportsInWindow(ctx context.Context, tenantID string, since time.Time) (int64, error)
	CountTenantRunningJobs(ctx context.Context, tenantID string) (int64, error)

	EnqueueJob(ctx context.Context, job *ReportJob) error
	AcquireJobLock(ctx context.Context, workerID string) (*ReportJob, error)
	UpdateJobStatus(ctx context.Context, id uuid.UUID, status JobStatus) error
	TouchJobHeartbeat(ctx context.Context, id uuid.UUID) error
	ResetStalledJobs(ctx context.Context, olderThan time.Duration) (int64, error)
	SendToDeadLetterQueue(ctx context.Context, jobID uuid.UUID, reason string) error
	ListenForJobs(ctx context.Context) <-chan struct{}
	CountDLQEntries(ctx context.Context) (int64, error)
	RequeueDLQJob(ctx context.Context, jobID uuid.UUID) error

	CreateReportTemplate(ctx context.Context, tmpl *ReportTemplate) error
	GetReportTemplate(ctx context.Context, id uuid.UUID) (*ReportTemplate, error)
	ListReportTemplates(ctx context.Context) ([]ReportTemplate, error)

	ListEnabledScheduledReports(ctx context.Context) ([]ScheduledReport, error)
	AcquireSchedulerLock(ctx context.Context) (bool, error)
	ReleaseSchedulerLock(ctx context.Context)

	RecordAuditLog(ctx context.Context, log *ReportAuditLog) error
}

type repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

func (r *repository) CreateReport(ctx context.Context, report *Report) error {
	query := `
		INSERT INTO reports (
			id, tenant_id, report_type, status, parameters, requested_by, requested_at, progress, retry_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	paramsJSON, err := json.Marshal(report.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		report.ID, report.TenantID, report.Type, report.Status, paramsJSON,
		report.RequestedBy, report.RequestedAt, report.Progress, report.RetryCount,
	)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}

func (r *repository) GetReport(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Report, error) {
	query := `
		SELECT id, tenant_id, report_type, status, parameters, requested_by, requested_at, 
		       started_at, completed_at, progress, file_url, error_message, retry_count, checksum
		FROM reports
		WHERE id = $1 AND tenant_id = $2
	`

	var report Report
	var paramsJSON []byte
	var fileUrl, errorMessage, checksum *string

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&report.ID, &report.TenantID, &report.Type, &report.Status, &paramsJSON,
		&report.RequestedBy, &report.RequestedAt, &report.StartedAt, &report.CompletedAt,
		&report.Progress, &fileUrl, &errorMessage, &report.RetryCount, &checksum,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil 
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	if fileUrl != nil { report.FileURL = *fileUrl }
	if errorMessage != nil { report.ErrorMessage = *errorMessage }
	if checksum != nil { report.Checksum = *checksum }

	if len(paramsJSON) > 0 {
		if err := json.Unmarshal(paramsJSON, &report.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}
	}

	return &report, nil
}

func (r *repository) ListReports(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]Report, error) {
	query := `
		SELECT id, tenant_id, report_type, status, requested_at, progress
		FROM reports
		WHERE tenant_id = $1
		ORDER BY requested_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var report Report

		err := rows.Scan(
			&report.ID, &report.TenantID, &report.Type, &report.Status, 
			&report.RequestedAt, &report.Progress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("report iteration error: %w", err)
	}

	return reports, nil
}

func (r *repository) UpdateReportStatus(ctx context.Context, id uuid.UUID, status ReportStatus, progress int, fileURL, errorMessage string) error {
	query := `
		UPDATE reports
		SET status = $1, progress = $2, file_url = $3, error_message = $4
	`
	var args []interface{}
	args = append(args, status, progress)

	var fileURLArg interface{} = nil
	if fileURL != "" { fileURLArg = fileURL }
	args = append(args, fileURLArg)

	var errorMsgArg interface{} = nil
	if errorMessage != "" { errorMsgArg = errorMessage }
	args = append(args, errorMsgArg)

	if status == ReportStatusRunning {
		query += ", started_at = NOW()"
	} else if status == ReportStatusCompleted || status == ReportStatusFailed {
		query += ", completed_at = NOW()"
	}

	query += ` WHERE id = $5`
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}
	return nil
}

func (r *repository) EnqueueJob(ctx context.Context, job *ReportJob) error {
	query := `
		INSERT INTO report_jobs (
			id, report_id, job_status, attempts, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		job.ID, job.ReportID, job.JobStatus, job.Attempts, job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	return nil
}

func (r *repository) AcquireJobLock(ctx context.Context, workerID string) (*ReportJob, error) {
	query := `
		UPDATE report_jobs
		SET job_status = $1, worker_id = $2, updated_at = NOW(), attempts = attempts + 1
		WHERE id = (
			SELECT id FROM report_jobs
			WHERE job_status = 'QUEUED' OR (job_status = 'FAILED' AND attempts < 3 AND updated_at < NOW() - INTERVAL '5 minutes')
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, report_id, job_status, worker_id, attempts, created_at, updated_at
	`
	
	var job ReportJob
	var retWorkerID *string
	err := r.db.QueryRow(ctx, query, JobStatusRunning, workerID).Scan(
		&job.ID, &job.ReportID, &job.JobStatus, &retWorkerID, &job.Attempts, &job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Queue empty
		}
		return nil, fmt.Errorf("failed to acquire job lock: %w", err)
	}

	if retWorkerID != nil { job.WorkerID = *retWorkerID }

	return &job, nil
}

func (r *repository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status JobStatus) error {
	query := `
		UPDATE report_jobs
		SET job_status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

func (r *repository) RecordAuditLog(ctx context.Context, log *ReportAuditLog) error {
	query := `
		INSERT INTO report_audit_logs (
			id, report_id, user_id, action, timestamp, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	var ipArg interface{} = nil
	if log.IPAddress != "" { ipArg = log.IPAddress }
	_, err := r.db.Exec(ctx, query, log.ID, log.ReportID, log.UserID, log.Action, log.Timestamp, ipArg)
	if err != nil {
		return fmt.Errorf("failed to record audit log: %w", err)
	}
	return nil
}
func (r *repository) DeleteExpiredReports(ctx context.Context, olderThanDays int) (int64, error) {
	query := `DELETE FROM reports WHERE requested_at < NOW() - ($1 || ' days')::INTERVAL`
	result, err := r.db.Exec(ctx, query, olderThanDays)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired reports: %w", err)
	}
	return result.RowsAffected(), nil
}

func (r *repository) CreateReportTemplate(ctx context.Context, tmpl *ReportTemplate) error {
	filtersJSON, err := json.Marshal(tmpl.AllowedFilters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}
	query := `
		INSERT INTO report_templates (id, name, description, query_template, allowed_filters, output_format, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.Exec(ctx, query,
		tmpl.ID, tmpl.Name, tmpl.Description, tmpl.QueryTemplate, filtersJSON, tmpl.OutputFormat, tmpl.Version)
	if err != nil {
		return fmt.Errorf("failed to create report template: %w", err)
	}
	return nil
}

func (r *repository) GetReportTemplate(ctx context.Context, id uuid.UUID) (*ReportTemplate, error) {
	query := `
		SELECT id, name, description, query_template, allowed_filters, output_format, version
		FROM report_templates WHERE id = $1
	`
	var tmpl ReportTemplate
	var filtersJSON []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tmpl.ID, &tmpl.Name, &tmpl.Description, &tmpl.QueryTemplate, &filtersJSON, &tmpl.OutputFormat, &tmpl.Version,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get report template: %w", err)
	}
	if len(filtersJSON) > 0 {
		json.Unmarshal(filtersJSON, &tmpl.AllowedFilters)
	}
	return &tmpl, nil
}

func (r *repository) ListReportTemplates(ctx context.Context) ([]ReportTemplate, error) {
	query := `SELECT id, name, description, output_format, version FROM report_templates ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list report templates: %w", err)
	}
	defer rows.Close()
	var templates []ReportTemplate
	for rows.Next() {
		var tmpl ReportTemplate
		if err := rows.Scan(&tmpl.ID, &tmpl.Name, &tmpl.Description, &tmpl.OutputFormat, &tmpl.Version); err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}
	return templates, rows.Err()
}

func (r *repository) ListEnabledScheduledReports(ctx context.Context) ([]ScheduledReport, error) {
	query := `
		SELECT id, tenant_id, report_template_id, cron_expression, parameters, enabled, created_at
		FROM scheduled_reports WHERE enabled = true
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled reports: %w", err)
	}
	defer rows.Close()
	var schedules []ScheduledReport
	for rows.Next() {
		var s ScheduledReport
		var paramsJSON []byte
		if err := rows.Scan(&s.ID, &s.TenantID, &s.ReportTemplateID, &s.CronExpression, &paramsJSON, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		if len(paramsJSON) > 0 {
			json.Unmarshal(paramsJSON, &s.Parameters)
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// ─── Heartbeat & Crash Recovery ────────────────────────────────────────────

func (r *repository) TouchJobHeartbeat(ctx context.Context, id uuid.UUID) error {
	// Update last_heartbeat (distinct column) so that updated_at is not conflated
	// with heartbeat, allowing long-running jobs to stay alive.
	_, err := r.db.Exec(ctx,
		"UPDATE report_jobs SET last_heartbeat = NOW() WHERE id = $1", id)
	return err
}

func (r *repository) ResetStalledJobs(ctx context.Context, olderThan time.Duration) (int64, error) {
	// Use last_heartbeat (not updated_at) to determine staleness.
	// Timeout = olderThan (e.g. 2×heartbeatInterval) + 15s margin to avoid
	// premature resets of legitimately slow jobs (e.g. heavy aggregation).
	marginSeconds := int(olderThan.Seconds()) + 15
	result, err := r.db.Exec(ctx,
		"UPDATE report_jobs SET job_status = 'QUEUED', worker_id = NULL, updated_at = NOW(), last_heartbeat = NULL"+
			" WHERE job_status = 'RUNNING'"+
			" AND (last_heartbeat IS NULL OR last_heartbeat < NOW() - ($1 || ' seconds')::INTERVAL)",
		marginSeconds)
	if err != nil {
		return 0, fmt.Errorf("failed to reset stalled jobs: %w", err)
	}
	return result.RowsAffected(), nil
}

// ─── Dead-Letter Queue ──────────────────────────────────────────────────────

func (r *repository) SendToDeadLetterQueue(ctx context.Context, jobID uuid.UUID, reason string) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO report_dlq (id, job_id, reason, created_at) VALUES ($1, $2, $3, NOW())"+
			" ON CONFLICT (job_id) DO UPDATE SET reason = EXCLUDED.reason, created_at = NOW()",
		uuid.New(), jobID, reason)
	if err != nil {
		return fmt.Errorf("failed to send job to DLQ: %w", err)
	}
	r.db.Exec(ctx,
		"UPDATE report_jobs SET job_status = 'FAILED', updated_at = NOW() WHERE id = $1", jobID)
	return nil
}

// ─── LISTEN/NOTIFY ──────────────────────────────────────────────────────────

// ListenForJobs subscribes to the 'report_jobs' NOTIFY channel.
// On connection drop, it automatically reconnects with 5s backoff.
// The returned channel receives a signal whenever a new job is enqueued.
func (r *repository) ListenForJobs(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)

	go func() {
		defer close(ch)
		for {
			if ctx.Err() != nil {
				return
			}
			if err := r.listenOnce(ctx, ch); err != nil {
				if ctx.Err() != nil {
					return
				}
				r.logger.Warn("LISTEN/NOTIFY connection lost, reconnecting in 5s", zap.Error(err))
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
			}
		}
	}()
	return ch
}

// listenOnce acquires a connection, issues LISTEN, and polls until context cancel or error.
func (r *repository) listenOnce(ctx context.Context, ch chan struct{}) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "LISTEN report_jobs"); err != nil {
		return err
	}
	r.logger.Info("LISTEN report_jobs established")

	for {
		_, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err // triggers reconnect in outer loop
		}
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// ─── Scheduler Advisory Lock ────────────────────────────────────────────────

const schedulerAdvisoryLockID = 2200000022

func (r *repository) AcquireSchedulerLock(ctx context.Context) (bool, error) {
	var acquired bool
	err := r.db.QueryRow(ctx,
		"SELECT pg_try_advisory_lock($1)", schedulerAdvisoryLockID).Scan(&acquired)
	return acquired, err
}

func (r *repository) ReleaseSchedulerLock(ctx context.Context) {
	r.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", schedulerAdvisoryLockID)
}

// ─── Tenant Quota ────────────────────────────────────────────────────────────

func (r *repository) CountTenantReportsInWindow(ctx context.Context, tenantID string, since time.Time) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM reports WHERE tenant_id::text = $1 AND requested_at >= $2",
		tenantID, since,
	).Scan(&count)
	return count, err
}

func (r *repository) CountTenantRunningJobs(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM report_jobs rj"+
			" JOIN reports r ON r.id = rj.report_id"+
			" WHERE r.tenant_id::text = $1 AND rj.job_status = 'RUNNING'",
		tenantID,
	).Scan(&count)
	return count, err
}
func (r *repository) ListExpiredReports(ctx context.Context, olderThanDays int) ([]Report, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, file_url FROM reports WHERE requested_at < NOW() - ($1 || ' days')::INTERVAL AND file_url IS NOT NULL",
		olderThanDays)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired reports: %w", err)
	}
	defer rows.Close()
	var reports []Report
	for rows.Next() {
		var rep Report
		if err := rows.Scan(&rep.ID, &rep.FileURL); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}

func (r *repository) CountDLQEntries(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM report_dlq").Scan(&count)
	return count, err
}

func (r *repository) RequeueDLQJob(ctx context.Context, jobID uuid.UUID) error {
	// Move job back to QUEUED and remove from DLQ
	_, err := r.db.Exec(ctx,
		"UPDATE report_jobs SET job_status = 'QUEUED', worker_id = NULL, attempts = 0, updated_at = NOW() WHERE id = $1",
		jobID)
	if err != nil {
		return fmt.Errorf("failed to requeue job: %w", err)
	}
	r.db.Exec(ctx, "DELETE FROM report_dlq WHERE job_id = $1", jobID)
	return nil
}

