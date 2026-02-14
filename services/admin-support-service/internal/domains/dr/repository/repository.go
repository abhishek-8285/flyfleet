package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/admin-support-service/internal/domains/dr/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DrRepository interface {
	CreateBackup(ctx context.Context, backup *models.Backup) error
	GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error)
	ListBackups(ctx context.Context, limit, offset int) ([]*models.Backup, int, error)
	UpdateBackupStatus(ctx context.Context, id uuid.UUID, status models.BackupStatus, s3Key string, sizeBytes int64) error

	CreateRestoreJob(ctx context.Context, job *models.RestoreJob) error
	GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error)
	UpdateRestoreJobStatus(ctx context.Context, id uuid.UUID, status models.RestoreStatus, logs string) error
	ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error)

	ListStoragePools(ctx context.Context) ([]*models.StoragePool, error)
	CreateBackupVerification(ctx context.Context, verification *models.BackupVerification) error
	GetBackupVerification(ctx context.Context, backupID uuid.UUID) (*models.BackupVerification, error)
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
		INSERT INTO backups (id, type, status, s3_key, size_bytes, created_by, created_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	
	if backup.CreatedAt.IsZero() {
		backup.CreatedAt = time.Now()
	}

	err := r.db.QueryRow(ctx, query,
		backup.ID, backup.Type, backup.Status, backup.S3Key, backup.SizeBytes,
		backup.CreatedBy, backup.CreatedAt, backup.Metadata,
	).Scan(&backup.ID)

	if err != nil {
		r.logger.Error("Failed to create backup", zap.Error(err))
		return fmt.Errorf("failed to create backup: %w", err)
	}
	return nil
}

func (r *drRepository) GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	query := `SELECT id, type, status, s3_key, size_bytes, created_by, created_at, completed_at, metadata FROM backups WHERE id = $1`
	var backup models.Backup
	err := r.db.QueryRow(ctx, query, id).Scan(
		&backup.ID, &backup.Type, &backup.Status, &backup.S3Key, &backup.SizeBytes,
		&backup.CreatedBy, &backup.CreatedAt, &backup.CompletedAt, &backup.Metadata,
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
		SELECT id, type, status, s3_key, size_bytes, created_by, created_at, completed_at, metadata
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
			&b.CreatedBy, &b.CreatedAt, &b.CompletedAt, &b.Metadata,
		); err != nil {
			return nil, 0, err
		}
		backups = append(backups, &b)
	}
	return backups, total, nil
}

func (r *drRepository) UpdateBackupStatus(ctx context.Context, id uuid.UUID, status models.BackupStatus, s3Key string, sizeBytes int64) error {
	query := `UPDATE backups SET status = $1, s3_key = $2, size_bytes = $3, completed_at = CASE WHEN $1 IN ('COMPLETED', 'FAILED') THEN NOW() ELSE completed_at END WHERE id = $4`
	_, err := r.db.Exec(ctx, query, status, s3Key, sizeBytes, id)
	return err
}

func (r *drRepository) CreateRestoreJob(ctx context.Context, job *models.RestoreJob) error {
	query := `
		INSERT INTO restore_jobs (id, backup_id, status, target_environment, started_at, performed_by, logs)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	
	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now()
	}

	err := r.db.QueryRow(ctx, query,
		job.ID, job.BackupID, job.Status, job.TargetEnvironment,
		job.StartedAt, job.PerformedBy, job.Logs,
	).Scan(&job.ID)

	if err != nil {
		return fmt.Errorf("failed to create restore job: %w", err)
	}
	return nil
}

func (r *drRepository) GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error) {
	query := `SELECT id, backup_id, status, target_environment, started_at, completed_at, performed_by, logs FROM restore_jobs WHERE id = $1`
	var job models.RestoreJob
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.BackupID, &job.Status, &job.TargetEnvironment,
		&job.StartedAt, &job.CompletedAt, &job.PerformedBy, &job.Logs,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *drRepository) UpdateRestoreJobStatus(ctx context.Context, id uuid.UUID, status models.RestoreStatus, logs string) error {
	query := `UPDATE restore_jobs SET status = $1, logs = $2, completed_at = CASE WHEN $1 IN ('COMPLETED', 'FAILED') THEN NOW() ELSE completed_at END WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, logs, id)
	return err
}

func (r *drRepository) ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error) {
	countQuery := `SELECT COUNT(*) FROM restore_jobs`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, backup_id, status, target_environment, started_at, completed_at, performed_by, logs
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
			&j.StartedAt, &j.CompletedAt, &j.PerformedBy, &j.Logs,
		); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, &j)
	}
	return jobs, total, nil
}

func (r *drRepository) ListStoragePools(ctx context.Context) ([]*models.StoragePool, error) {
	query := `
		SELECT id, name, status, total_bytes, used_bytes, available_bytes, location, backup_count, created_at, updated_at
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
			&p.AvailableBytes, &p.Location, &p.BackupCount, &p.CreatedAt, &p.UpdatedAt,
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
