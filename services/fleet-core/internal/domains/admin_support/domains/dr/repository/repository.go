package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DrRepository interface {
	CreateBackup(ctx context.Context, backup *models.Backup) error
	GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error)
	ListBackups(ctx context.Context, limit, offset int) ([]*models.Backup, int, error)
	UpdateBackupStatus(ctx context.Context, id uuid.UUID, status models.BackupStatus, s3Key string, sizeBytes int64, updatedBy uuid.UUID) error

	CreateRestoreJob(ctx context.Context, job *models.RestoreJob) error
	GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error)
	UpdateRestoreJobStatus(ctx context.Context, id uuid.UUID, status models.RestoreStatus, logs string, updatedBy uuid.UUID) error
	ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error)

	ListStoragePools(ctx context.Context) ([]*models.StoragePool, error)
	CreateBackupVerification(ctx context.Context, verification *models.BackupVerification) error
	GetBackupVerification(ctx context.Context, backupID uuid.UUID) (*models.BackupVerification, error)

	CreateDRTest(ctx context.Context, test *models.DRTest) error
	ListDRTests(ctx context.Context, limit, offset int) ([]*models.DRTest, int, error)
	GetDRTest(ctx context.Context, id uuid.UUID) (*models.DRTest, error)
	UpdateDRTestStatus(ctx context.Context, id uuid.UUID, status string, progress int, completedAt *time.Time) error
	AbortDRTest(ctx context.Context, id uuid.UUID) error

	CreateDRTestStep(ctx context.Context, step *models.DRTestStep) error
	UpdateDRTestStepStatus(ctx context.Context, stepID uuid.UUID, status string, startedAt, completedAt *time.Time) error
	MarkPendingDRTestStepsSkipped(ctx context.Context, testID uuid.UUID) error
	ListDRTestSteps(ctx context.Context, testID uuid.UUID) ([]*models.DRTestStep, error)

	CreateDRTestLog(ctx context.Context, log *models.DRTestLog) error
	ListDRTestLogs(ctx context.Context, testID uuid.UUID, limit int) ([]*models.DRTestLog, error)

	GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error)
	GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error)
	UpsertDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error
}

type drRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewDrRepository(db *pgxpool.Pool, logger *zap.Logger) DrRepository {
	return &drRepository{db: db, logger: logger}
}

func (r *drRepository) CreateBackup(ctx context.Context, backup *models.Backup) error {
	query := `
		INSERT INTO backups (id, type, status, s3_key, size_bytes, created_by, updated_by, created_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	if backup.CreatedAt.IsZero() {
		backup.CreatedAt = time.Now()
	}

	err := r.db.QueryRow(ctx, query,
		backup.ID, backup.Type, backup.Status, backup.S3Key, backup.SizeBytes,
		backup.CreatedBy, backup.UpdatedBy, backup.CreatedAt, backup.Metadata,
	).Scan(&backup.ID)

	if err != nil {
		r.logger.Error("Failed to create backup", zap.Error(err))
		return fmt.Errorf("failed to create backup: %w", err)
	}
	return nil
}

func (r *drRepository) GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	query := `SELECT id, type, status, s3_key, size_bytes, created_by, updated_by, created_at, completed_at, metadata FROM backups WHERE id = $1`
	var backup models.Backup
	err := r.db.QueryRow(ctx, query, id).Scan(
		&backup.ID, &backup.Type, &backup.Status, &backup.S3Key, &backup.SizeBytes,
		&backup.CreatedBy, &backup.UpdatedBy, &backup.CreatedAt, &backup.CompletedAt, &backup.Metadata,
	)
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *drRepository) ListBackups(ctx context.Context, limit, offset int) ([]*models.Backup, int, error) {
	countQuery := `SELECT COUNT(*) FROM backups`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, type, status, s3_key, size_bytes, created_by, updated_by, created_at, completed_at, metadata
		FROM backups
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var backups []*models.Backup
	for rows.Next() {
		var b models.Backup
		if err := rows.Scan(
			&b.ID, &b.Type, &b.Status, &b.S3Key, &b.SizeBytes,
			&b.CreatedBy, &b.UpdatedBy, &b.CreatedAt, &b.CompletedAt, &b.Metadata,
		); err != nil {
			return nil, 0, err
		}
		backups = append(backups, &b)
	}
	return backups, total, nil
}

func (r *drRepository) UpdateBackupStatus(ctx context.Context, id uuid.UUID, status models.BackupStatus, s3Key string, sizeBytes int64, updatedBy uuid.UUID) error {
	query := `UPDATE backups SET status = $1, s3_key = $2, size_bytes = $3, completed_at = CASE WHEN $1 IN ('COMPLETED', 'FAILED') THEN NOW() ELSE completed_at END, updated_by = $4, updated_at = NOW() WHERE id = $5`
	_, err := r.db.Exec(ctx, query, status, s3Key, sizeBytes, updatedBy, id)
	return err
}

func (r *drRepository) CreateRestoreJob(ctx context.Context, job *models.RestoreJob) error {
	query := `
		INSERT INTO restore_jobs (id, backup_id, status, target_environment, started_at, performed_by, updated_by, logs)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now()
	}

	err := r.db.QueryRow(ctx, query,
		job.ID, job.BackupID, job.Status, job.TargetEnvironment,
		job.StartedAt, job.PerformedBy, job.UpdatedBy, job.Logs,
	).Scan(&job.ID)

	if err != nil {
		return fmt.Errorf("failed to create restore job: %w", err)
	}
	return nil
}

func (r *drRepository) GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error) {
	query := `SELECT id, backup_id, status, target_environment, started_at, completed_at, performed_by, updated_by, updated_at, logs FROM restore_jobs WHERE id = $1`
	var job models.RestoreJob
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.BackupID, &job.Status, &job.TargetEnvironment,
		&job.StartedAt, &job.CompletedAt, &job.PerformedBy, &job.UpdatedBy, &job.UpdatedAt, &job.Logs,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *drRepository) UpdateRestoreJobStatus(ctx context.Context, id uuid.UUID, status models.RestoreStatus, logs string, updatedBy uuid.UUID) error {
	query := `UPDATE restore_jobs SET status = $1, logs = $2, completed_at = CASE WHEN $1 IN ('COMPLETED', 'FAILED') THEN NOW() ELSE completed_at END, updated_by = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.Exec(ctx, query, status, logs, updatedBy, id)
	return err
}

func (r *drRepository) ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error) {
	countQuery := `SELECT COUNT(*) FROM restore_jobs`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, backup_id, status, target_environment, started_at, completed_at, performed_by, updated_by, updated_at, logs
		FROM restore_jobs
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []*models.RestoreJob
	for rows.Next() {
		var j models.RestoreJob
		if err := rows.Scan(
			&j.ID, &j.BackupID, &j.Status, &j.TargetEnvironment,
			&j.StartedAt, &j.CompletedAt, &j.PerformedBy, &j.UpdatedBy, &j.UpdatedAt, &j.Logs,
		); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, &j)
	}
	return jobs, total, nil
}

func (r *drRepository) ListStoragePools(ctx context.Context) ([]*models.StoragePool, error) {
	query := `
		SELECT id, name, status, total_bytes, used_bytes, available_bytes, location, backup_count, created_at, created_by, updated_at, updated_by
		FROM storage_pools
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %w", err)
	}
	defer rows.Close()

	var pools []*models.StoragePool
	for rows.Next() {
		var p models.StoragePool
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Status, &p.TotalBytes, &p.UsedBytes,
			&p.AvailableBytes, &p.Location, &p.BackupCount, &p.CreatedAt, &p.CreatedBy, &p.UpdatedAt, &p.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan storage pool: %w", err)
		}
		pools = append(pools, &p)
	}
	return pools, nil
}

func (r *drRepository) CreateBackupVerification(ctx context.Context, verification *models.BackupVerification) error {
	query := `
		INSERT INTO backup_verifications (id, backup_id, status, checksum, verified_at, verified_by, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (backup_id) DO UPDATE SET status = $3, checksum = $4, verified_at = $5, verified_by = $6, error_message = $7`

	_, err := r.db.Exec(ctx, query,
		verification.ID, verification.BackupID, verification.Status, verification.Checksum,
		verification.VerifiedAt, verification.VerifiedBy, verification.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("failed to create backup verification: %w", err)
	}
	return nil
}

func (r *drRepository) GetBackupVerification(ctx context.Context, backupID uuid.UUID) (*models.BackupVerification, error) {
	query := `SELECT id, backup_id, status, checksum, verified_at, verified_by, error_message FROM backup_verifications WHERE backup_id = $1`
	var v models.BackupVerification
	err := r.db.QueryRow(ctx, query, backupID).Scan(
		&v.ID, &v.BackupID, &v.Status, &v.Checksum, &v.VerifiedAt, &v.VerifiedBy, &v.ErrorMessage,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *drRepository) CreateDRTest(ctx context.Context, test *models.DRTest) error {
	query := `
		INSERT INTO dr_tests (
			id, test_id, name, description, test_type, target_env, status, progress_percent,
			initiated_by, success_criteria, scheduled_at, started_at, completed_at, created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15, $16, $17
		)`
	_, err := r.db.Exec(ctx, query,
		test.ID, test.TestID, test.Name, test.Description, test.TestType, test.TargetEnv, test.Status, test.ProgressPercent,
		test.InitiatedBy, test.SuccessCriteria, test.ScheduledAt, test.StartedAt, test.CompletedAt, test.CreatedAt, test.CreatedBy, test.UpdatedAt, test.UpdatedBy,
	)
	return err
}

func (r *drRepository) ListDRTests(ctx context.Context, limit, offset int) ([]*models.DRTest, int, error) {
	countQuery := `SELECT COUNT(*) FROM dr_tests`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, test_id, name, description, test_type, target_env, status, progress_percent,
		       initiated_by, success_criteria, scheduled_at, started_at, completed_at, created_at, created_by, updated_at, updated_by
		FROM dr_tests
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tests []*models.DRTest
	for rows.Next() {
		var t models.DRTest
		if err := rows.Scan(
			&t.ID, &t.TestID, &t.Name, &t.Description, &t.TestType, &t.TargetEnv, &t.Status, &t.ProgressPercent,
			&t.InitiatedBy, &t.SuccessCriteria, &t.ScheduledAt, &t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.CreatedBy, &t.UpdatedAt, &t.UpdatedBy,
		); err != nil {
			return nil, 0, err
		}
		tests = append(tests, &t)
	}
	return tests, total, nil
}

func (r *drRepository) GetDRTest(ctx context.Context, id uuid.UUID) (*models.DRTest, error) {
	query := `
		SELECT id, test_id, name, description, test_type, target_env, status, progress_percent,
		       initiated_by, success_criteria, scheduled_at, started_at, completed_at, created_at, created_by, updated_at, updated_by
		FROM dr_tests WHERE id = $1`
	var t models.DRTest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TestID, &t.Name, &t.Description, &t.TestType, &t.TargetEnv, &t.Status, &t.ProgressPercent,
		&t.InitiatedBy, &t.SuccessCriteria, &t.ScheduledAt, &t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.CreatedBy, &t.UpdatedAt, &t.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *drRepository) UpdateDRTestStatus(ctx context.Context, id uuid.UUID, status string, progress int, completedAt *time.Time) error {
	query := `
		UPDATE dr_tests
		SET status = $1, progress_percent = $2, completed_at = COALESCE($3, completed_at), updated_at = NOW()
		WHERE id = $4`
	_, err := r.db.Exec(ctx, query, status, progress, completedAt, id)
	return err
}

func (r *drRepository) AbortDRTest(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE dr_tests
		SET status = 'ABORTED', completed_at = $1, updated_at = NOW()
		WHERE id = $2 AND status IN ('SCHEDULED', 'RUNNING')`
	_, err := r.db.Exec(ctx, query, now, id)
	return err
}

func (r *drRepository) CreateDRTestStep(ctx context.Context, step *models.DRTestStep) error {
	query := `
		INSERT INTO dr_test_steps (id, test_id, step_order, step_name, status, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, step.ID, step.TestID, step.StepOrder, step.StepName, step.Status, step.StartedAt, step.CompletedAt)
	return err
}

func (r *drRepository) UpdateDRTestStepStatus(ctx context.Context, stepID uuid.UUID, status string, startedAt, completedAt *time.Time) error {
	query := `
		UPDATE dr_test_steps
		SET status = $1, started_at = COALESCE($2, started_at), completed_at = COALESCE($3, completed_at)
		WHERE id = $4`
	_, err := r.db.Exec(ctx, query, status, startedAt, completedAt, stepID)
	return err
}

func (r *drRepository) MarkPendingDRTestStepsSkipped(ctx context.Context, testID uuid.UUID) error {
	query := `
		UPDATE dr_test_steps
		SET status = 'SKIPPED', completed_at = NOW()
		WHERE test_id = $1 AND status IN ('PENDING', 'RUNNING')`
	_, err := r.db.Exec(ctx, query, testID)
	return err
}

func (r *drRepository) ListDRTestSteps(ctx context.Context, testID uuid.UUID) ([]*models.DRTestStep, error) {
	query := `
		SELECT id, test_id, step_order, step_name, status, started_at, completed_at
		FROM dr_test_steps
		WHERE test_id = $1
		ORDER BY step_order ASC`
	rows, err := r.db.Query(ctx, query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*models.DRTestStep
	for rows.Next() {
		var s models.DRTestStep
		if err := rows.Scan(&s.ID, &s.TestID, &s.StepOrder, &s.StepName, &s.Status, &s.StartedAt, &s.CompletedAt); err != nil {
			return nil, err
		}
		steps = append(steps, &s)
	}
	return steps, nil
}

func (r *drRepository) CreateDRTestLog(ctx context.Context, log *models.DRTestLog) error {
	query := `INSERT INTO dr_test_logs (id, test_id, ts, level, message) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, log.ID, log.TestID, log.Timestamp, log.Level, log.Message)
	return err
}

func (r *drRepository) ListDRTestLogs(ctx context.Context, testID uuid.UUID, limit int) ([]*models.DRTestLog, error) {
	query := `
		SELECT id, test_id, ts, level, message
		FROM dr_test_logs
		WHERE test_id = $1
		ORDER BY ts DESC
		LIMIT $2`
	rows, err := r.db.Query(ctx, query, testID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.DRTestLog
	for rows.Next() {
		var l models.DRTestLog
		if err := rows.Scan(&l.ID, &l.TestID, &l.Timestamp, &l.Level, &l.Message); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}

	// Reverse to chronological order for UI readability.
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
	return logs, nil
}

func (r *drRepository) GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dr_tests`).Scan(&total); err != nil {
		return nil, err
	}

	var passed int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dr_tests WHERE status = 'PASSED'`).Scan(&passed); err != nil {
		return nil, err
	}

	var avgSeconds *float64
	if err := r.db.QueryRow(ctx, `
		SELECT AVG(EXTRACT(EPOCH FROM (completed_at - started_at)))
		FROM dr_tests
		WHERE status IN ('PASSED', 'FAILED', 'ABORTED') AND started_at IS NOT NULL AND completed_at IS NOT NULL`).Scan(&avgSeconds); err != nil {
		return nil, err
	}

	var nextScheduled *time.Time
	if err := r.db.QueryRow(ctx, `
		SELECT MIN(scheduled_at)
		FROM dr_tests
		WHERE status = 'SCHEDULED' AND scheduled_at IS NOT NULL`).Scan(&nextScheduled); err != nil {
		return nil, err
	}

	successRate := 0.0
	if total > 0 {
		successRate = (float64(passed) / float64(total)) * 100
	}

	avgRecovery := "—"
	if avgSeconds != nil {
		avgRecovery = fmt.Sprintf("%.0fs", *avgSeconds)
	}

	return &models.DRTestMetrics{
		TotalTests:      total,
		SuccessRate:     successRate,
		AvgRecoveryTime: avgRecovery,
		NextScheduled:   nextScheduled,
	}, nil
}

func (r *drRepository) GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error) {
	query := `
		SELECT schedule_enabled, frequency, schedule_day, schedule_time, intensity, notification_emails
		FROM dr_test_settings
		WHERE id = TRUE`
	var s models.DRTestSettings
	var emailsRaw []byte
	if err := r.db.QueryRow(ctx, query).Scan(
		&s.ScheduleEnabled, &s.Frequency, &s.ScheduleDay, &s.ScheduleTime, &s.Intensity, &emailsRaw,
	); err != nil {
		return nil, err
	}
	if len(emailsRaw) > 0 {
		if err := json.Unmarshal(emailsRaw, &s.NotificationEmails); err != nil {
			return nil, err
		}
	}
	if s.NotificationEmails == nil {
		s.NotificationEmails = []string{}
	}
	return &s, nil
}

func (r *drRepository) UpsertDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error {
	emailsJSON, err := json.Marshal(settings.NotificationEmails)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO dr_test_settings (id, schedule_enabled, frequency, schedule_day, schedule_time, intensity, notification_emails, updated_at)
		VALUES (TRUE, $1, $2, $3, $4, $5, $6::jsonb, NOW())
		ON CONFLICT (id) DO UPDATE SET
			schedule_enabled = EXCLUDED.schedule_enabled,
			frequency = EXCLUDED.frequency,
			schedule_day = EXCLUDED.schedule_day,
			schedule_time = EXCLUDED.schedule_time,
			intensity = EXCLUDED.intensity,
			notification_emails = EXCLUDED.notification_emails,
			updated_at = NOW()`
	_, err = r.db.Exec(ctx, query,
		settings.ScheduleEnabled, settings.Frequency, settings.ScheduleDay, settings.ScheduleTime, settings.Intensity, string(emailsJSON),
	)
	return err
}
