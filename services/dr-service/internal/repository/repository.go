package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/dr-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateBackup(ctx context.Context, backup *models.Backup) error
	GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error)
	ListBackups(ctx context.Context) ([]models.Backup, error)
	UpdateBackupStatus(ctx context.Context, id uuid.UUID, status string) error
	AddLog(ctx context.Context, log *models.BackupLog) error
	GetLogs(ctx context.Context, backupID uuid.UUID) ([]models.BackupLog, error)
	ListStoragePools(ctx context.Context) ([]models.StoragePool, error)

	CreateRestore(ctx context.Context, restore *models.Restore) error
	GetRestore(ctx context.Context, id uuid.UUID) (*models.Restore, error)
	ListRestores(ctx context.Context) ([]models.Restore, error)
	GetRestoreLogs(ctx context.Context, restoreID uuid.UUID) ([]models.RestoreLog, error)
	UpdateRestoreStatus(ctx context.Context, id uuid.UUID, status string) error

	// DR Testing methods
	CreateDRTest(ctx context.Context, test *models.DRTest, steps []models.DRTestStep) error
	GetDRTest(ctx context.Context, id uuid.UUID) (*models.DRTest, error)
	ListDRTests(ctx context.Context) ([]models.DRTest, error)
	GetDRTestSteps(ctx context.Context, testID uuid.UUID) ([]models.DRTestStep, error)
	AddDRTestLog(ctx context.Context, log *models.DRTestLog) error
	GetDRTestLogs(ctx context.Context, testID uuid.UUID) ([]models.DRTestLog, error)
	UpdateDRTestStatus(ctx context.Context, id uuid.UUID, status string, progress int) error
	UpdateDRTestStepStatus(ctx context.Context, testID uuid.UUID, stepOrder int, status string) error
	GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error)
	GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error)
	UpdateDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateBackup(ctx context.Context, b *models.Backup) error {
	query := `INSERT INTO backups (id, backup_type, status, location, retention_days, remarks, initiated_by, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, b.ID, b.BackupType, b.Status, b.Location, b.RetentionDays, b.Remarks, b.InitiatedBy, time.Now(), time.Now())
	return err
}

func (r *PostgresRepository) GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	query := `SELECT id, backup_type, status, start_time, end_time, duration_seconds, size_bytes, location, retention_days, checksum, initiated_by, remarks, created_at, updated_at FROM backups WHERE id=$1`
	var b models.Backup
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.BackupType, &b.Status, &b.StartTime, &b.EndTime, &b.DurationSeconds, &b.SizeBytes, &b.Location, &b.RetentionDays, &b.Checksum, &b.InitiatedBy, &b.Remarks, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *PostgresRepository) ListBackups(ctx context.Context) ([]models.Backup, error) {
	query := `SELECT id, backup_type, status, start_time, end_time, duration_seconds, size_bytes, location, retention_days, checksum, initiated_by, remarks, created_at, updated_at FROM backups ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []models.Backup
	for rows.Next() {
		var b models.Backup
		if err := rows.Scan(
			&b.ID, &b.BackupType, &b.Status, &b.StartTime, &b.EndTime, &b.DurationSeconds, &b.SizeBytes, &b.Location, &b.RetentionDays, &b.Checksum, &b.InitiatedBy, &b.Remarks, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}

func (r *PostgresRepository) UpdateBackupStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE backups SET status=$1, updated_at=$2 WHERE id=$3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *PostgresRepository) AddLog(ctx context.Context, l *models.BackupLog) error {
	query := `INSERT INTO backup_logs (backup_id, action, result, actor, message, timestamp) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, l.BackupID, l.Action, l.Result, l.Actor, l.Message, time.Now())
	return err
}

func (r *PostgresRepository) GetLogs(ctx context.Context, backupID uuid.UUID) ([]models.BackupLog, error) {
	query := `SELECT id, backup_id, action, result, actor, message, timestamp FROM backup_logs WHERE backup_id=$1 ORDER BY timestamp DESC`
	rows, err := r.db.Query(ctx, query, backupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.BackupLog
	for rows.Next() {
		var l models.BackupLog
		if err := rows.Scan(&l.ID, &l.BackupID, &l.Action, &l.Result, &l.Actor, &l.Message, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *PostgresRepository) ListStoragePools(ctx context.Context) ([]models.StoragePool, error) {
	// Mock implementation for now as table might be empty
	// Or real implementation if we seeded it. Let's do real.
	query := `SELECT id, provider, bucket, region, tier, encryption_enabled, is_active, created_at FROM backup_storage_pools`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []models.StoragePool
	for rows.Next() {
		var p models.StoragePool
		if err := rows.Scan(&p.ID, &p.Provider, &p.Bucket, &p.Region, &p.Tier, &p.EncryptionEnabled, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		pools = append(pools, p)
	}
	
	if len(pools) == 0 {
		// Return hardcoded pools if DB is empty for UI demo
		return []models.StoragePool{
			{ID: 1, Provider: "AWS", Bucket: "fleet-security-dr-prod", Region: ptr("us-east-1"), Tier: "STANDARD", EncryptionEnabled: true, IsActive: true},
			{ID: 2, Provider: "AZURE", Bucket: "fleet-dr-secondary", Region: ptr("westeurope"), Tier: "COOL", EncryptionEnabled: true, IsActive: true},
		}, nil
	}
	
	return pools, nil
}

func ptr(s string) *string { return &s }

func (r *PostgresRepository) CreateRestore(ctx context.Context, res *models.Restore) error {
	query := `INSERT INTO dr_restores (id, backup_id, restore_type, target_env, status, initiated_by, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, res.ID, res.BackupID, res.RestoreType, res.TargetEnv, res.Status, res.InitiatedBy, time.Now())
	return err
}

func (r *PostgresRepository) GetRestore(ctx context.Context, id uuid.UUID) (*models.Restore, error) {
	query := `SELECT id, backup_id, restore_type, target_env, status, started_at, completed_at, initiated_by, duration_seconds, created_at FROM dr_restores WHERE id=$1`
	var res models.Restore
	err := r.db.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.BackupID, &res.RestoreType, &res.TargetEnv, &res.Status, &res.StartTime, &res.CompletedTime, &res.InitiatedBy, &res.DurationSeconds, &res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *PostgresRepository) ListRestores(ctx context.Context) ([]models.Restore, error) {
	query := `SELECT id, backup_id, restore_type, target_env, status, started_at, completed_at, initiated_by, duration_seconds, created_at FROM dr_restores ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restores []models.Restore
	for rows.Next() {
		var res models.Restore
		if err := rows.Scan(
			&res.ID, &res.BackupID, &res.RestoreType, &res.TargetEnv, &res.Status, &res.StartTime, &res.CompletedTime, &res.InitiatedBy, &res.DurationSeconds, &res.CreatedAt,
		); err != nil {
			return nil, err
		}
		restores = append(restores, res)
	}
	return restores, nil
}

func (r *PostgresRepository) GetRestoreLogs(ctx context.Context, restoreID uuid.UUID) ([]models.RestoreLog, error) {
	query := `SELECT id, restore_id, level, module, message, created_at FROM dr_restore_logs WHERE restore_id=$1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, restoreID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.RestoreLog
	for rows.Next() {
		var l models.RestoreLog
		if err := rows.Scan(&l.ID, &l.RestoreID, &l.Level, &l.Module, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *PostgresRepository) UpdateRestoreStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE dr_restores SET status=$1, completed_at=CASE WHEN $1='COMPLETED' OR $1='FAILED' THEN NOW() ELSE completed_at END WHERE id=$2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

// ========== DR Testing Repository Methods ==========

func (r *PostgresRepository) CreateDRTest(ctx context.Context, test *models.DRTest, steps []models.DRTestStep) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO dr_tests (id, test_id, name, description, test_type, target_env, status, progress_percent, initiated_by, success_criteria, scheduled_at, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = tx.Exec(ctx, query, test.ID, test.TestID, test.Name, test.Description, test.TestType, test.TargetEnv, test.Status, test.ProgressPercent, test.InitiatedBy, test.SuccessCriteria, test.ScheduledAt, time.Now(), time.Now())
	if err != nil {
		return err
	}

	for _, step := range steps {
		stepQuery := `INSERT INTO dr_test_steps (id, test_id, step_order, step_name, status, created_at)
                      VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = tx.Exec(ctx, stepQuery, step.ID, test.ID, step.StepOrder, step.StepName, step.Status, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetDRTest(ctx context.Context, id uuid.UUID) (*models.DRTest, error) {
	query := `SELECT id, test_id, name, description, test_type, target_env, status, progress_percent, initiated_by, success_criteria, scheduled_at, started_at, completed_at, created_at, updated_at FROM dr_tests WHERE id=$1`
	var t models.DRTest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TestID, &t.Name, &t.Description, &t.TestType, &t.TargetEnv, &t.Status, &t.ProgressPercent, &t.InitiatedBy, &t.SuccessCriteria, &t.ScheduledAt, &t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) ListDRTests(ctx context.Context) ([]models.DRTest, error) {
	query := `SELECT id, test_id, name, description, test_type, target_env, status, progress_percent, initiated_by, success_criteria, scheduled_at, started_at, completed_at, created_at, updated_at FROM dr_tests ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []models.DRTest
	for rows.Next() {
		var t models.DRTest
		if err := rows.Scan(&t.ID, &t.TestID, &t.Name, &t.Description, &t.TestType, &t.TargetEnv, &t.Status, &t.ProgressPercent, &t.InitiatedBy, &t.SuccessCriteria, &t.ScheduledAt, &t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tests = append(tests, t)
	}
	return tests, nil
}

func (r *PostgresRepository) GetDRTestSteps(ctx context.Context, testID uuid.UUID) ([]models.DRTestStep, error) {
	query := `SELECT id, test_id, step_order, step_name, status, started_at, completed_at, created_at FROM dr_test_steps WHERE test_id=$1 ORDER BY step_order ASC`
	rows, err := r.db.Query(ctx, query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.DRTestStep
	for rows.Next() {
		var s models.DRTestStep
		if err := rows.Scan(&s.ID, &s.TestID, &s.StepOrder, &s.StepName, &s.Status, &s.StartedAt, &s.CompletedAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, nil
}

func (r *PostgresRepository) AddDRTestLog(ctx context.Context, log *models.DRTestLog) error {
	query := `INSERT INTO dr_test_logs (id, test_id, timestamp, level, module, message) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, log.ID, log.TestID, time.Now(), log.Level, log.Module, log.Message)
	return err
}

func (r *PostgresRepository) GetDRTestLogs(ctx context.Context, testID uuid.UUID) ([]models.DRTestLog, error) {
	query := `SELECT id, test_id, timestamp, level, module, message FROM dr_test_logs WHERE test_id=$1 ORDER BY timestamp ASC`
	rows, err := r.db.Query(ctx, query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.DRTestLog
	for rows.Next() {
		var l models.DRTestLog
		if err := rows.Scan(&l.ID, &l.TestID, &l.Timestamp, &l.Level, &l.Module, &l.Message); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *PostgresRepository) UpdateDRTestStatus(ctx context.Context, id uuid.UUID, status string, progress int) error {
	query := `UPDATE dr_tests SET status=$1, progress_percent=$2, updated_at=$3, 
              started_at=CASE WHEN $1='RUNNING' AND started_at IS NULL THEN $3 ELSE started_at END,
              completed_at=CASE WHEN $1 IN ('PASSED', 'FAILED', 'ABORTED') THEN $3 ELSE completed_at END
              WHERE id=$4`
	_, err := r.db.Exec(ctx, query, status, progress, time.Now(), id)
	return err
}

func (r *PostgresRepository) UpdateDRTestStepStatus(ctx context.Context, testID uuid.UUID, stepOrder int, status string) error {
	query := `UPDATE dr_test_steps SET status=$1, 
              started_at=CASE WHEN $1='RUNNING' AND started_at IS NULL THEN NOW() ELSE started_at END,
              completed_at=CASE WHEN $1 IN ('PASSED', 'FAILED', 'SKIPPED') THEN NOW() ELSE completed_at END
              WHERE test_id=$2 AND step_order=$3`
	_, err := r.db.Exec(ctx, query, status, testID, stepOrder)
	return err
}

func (r *PostgresRepository) GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error) {
	metrics := &models.DRTestMetrics{}

	// Total tests
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dr_tests`).Scan(&metrics.TotalTests)

	// Success rate
	var passed, total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dr_tests WHERE status='PASSED'`).Scan(&passed)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dr_tests WHERE status IN ('PASSED', 'FAILED')`).Scan(&total)
	if total > 0 {
		metrics.SuccessRate = float64(passed) / float64(total) * 100
	}

	// Avg recovery time
	var avgSeconds *float64
	r.db.QueryRow(ctx, `SELECT AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) FROM dr_tests WHERE completed_at IS NOT NULL AND started_at IS NOT NULL`).Scan(&avgSeconds)
	if avgSeconds != nil {
		mins := int(*avgSeconds / 60)
		secs := int(*avgSeconds) % 60
		metrics.AvgRecoveryTime = fmt.Sprintf("%dm %ds", mins, secs)
	}

	// Next scheduled
	var nextScheduled *time.Time
	r.db.QueryRow(ctx, `SELECT scheduled_at FROM dr_tests WHERE status='SCHEDULED' AND scheduled_at > NOW() ORDER BY scheduled_at ASC LIMIT 1`).Scan(&nextScheduled)
	if nextScheduled != nil {
		metrics.NextScheduled = nextScheduled.Format(time.RFC3339)
	}

	return metrics, nil
}

func (r *PostgresRepository) GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error) {
	query := `SELECT id, org_id, schedule_enabled, frequency, schedule_day, schedule_time, intensity, notification_emails, created_at, updated_at FROM dr_test_settings LIMIT 1`
	var s models.DRTestSettings
	var emails []byte
	err := r.db.QueryRow(ctx, query).Scan(&s.ID, &s.OrgID, &s.ScheduleEnabled, &s.Frequency, &s.ScheduleDay, &s.ScheduleTime, &s.Intensity, &emails, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		// Return default settings if none exist
		return &models.DRTestSettings{
			ScheduleEnabled: false,
			Frequency:       "WEEKLY",
			ScheduleDay:     "SATURDAY",
			ScheduleTime:    "02:00:00",
			Intensity:       "NORMAL",
		}, nil
	}
	return &s, nil
}

func (r *PostgresRepository) UpdateDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error {
	query := `INSERT INTO dr_test_settings (id, schedule_enabled, frequency, schedule_day, schedule_time, intensity, notification_emails, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
              ON CONFLICT (id) DO UPDATE SET schedule_enabled=$2, frequency=$3, schedule_day=$4, schedule_time=$5, intensity=$6, notification_emails=$7, updated_at=$9`
	_, err := r.db.Exec(ctx, query, settings.ID, settings.ScheduleEnabled, settings.Frequency, settings.ScheduleDay, settings.ScheduleTime, settings.Intensity, settings.NotificationEmails, time.Now(), time.Now())
	return err
}
