package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/flyfleet/admin-support-service/internal/domains/dr/models"
	"github.com/flyfleet/admin-support-service/internal/domains/dr/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DrService interface {
	InitiateBackup(ctx context.Context, backupType models.BackupType, createdBy uuid.UUID) (*models.Backup, error)
	GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error)
	ListBackups(ctx context.Context, limit, offset int) ([]*models.Backup, int, error)
	VerifyBackup(ctx context.Context, backupID uuid.UUID, verifiedBy uuid.UUID) (*models.BackupVerification, error)

	InitiateRestore(ctx context.Context, backupID uuid.UUID, targetEnv string, performedBy uuid.UUID) (*models.RestoreJob, error)
	GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error)
	ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error)

	ListStoragePools(ctx context.Context) ([]*models.StoragePool, error)
}

type drService struct {
	repo   repository.DrRepository
	logger *zap.Logger
}

func NewDrService(repo repository.DrRepository, logger *zap.Logger) DrService {
	return &drService{repo: repo, logger: logger}
}

func (s *drService) InitiateBackup(ctx context.Context, backupType models.BackupType, createdBy uuid.UUID) (*models.Backup, error) {
	backup := &models.Backup{
		ID:        uuid.New(),
		Type:      backupType,
		Status:    models.BackupStatusPending,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if err := s.repo.CreateBackup(ctx, backup); err != nil {
		return nil, err
	}

	// In a real system, this would trigger an async job (e.g., via NATS or Asynq)
	// For now, we'll log it.
	s.logger.Info("Backup initiated", zap.String("id", backup.ID.String()), zap.String("type", string(backupType)))

	// Simulate immediate start for demonstration
	go s.processBackup(backup.ID)

	return backup, nil
}

func (s *drService) processBackup(id uuid.UUID) {
	// Simulate background work
	ctx := context.Background()
	time.Sleep(2 * time.Second)
	
	// Start
	_ = s.repo.UpdateBackupStatus(ctx, id, models.BackupStatusInProgress, "", 0)
	
	time.Sleep(5 * time.Second)
	
	// Complete
	s3Key := fmt.Sprintf("backups/%s/%s.tar.gz", time.Now().Format("2006-01-02"), id.String())
	_ = s.repo.UpdateBackupStatus(ctx, id, models.BackupStatusCompleted, s3Key, 1024*1024*500) // 500MB dummy
}

func (s *drService) GetBackup(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	return s.repo.GetBackup(ctx, id)
}

func (s *drService) ListBackups(ctx context.Context, limit, offset int) ([]*models.Backup, int, error) {
	return s.repo.ListBackups(ctx, limit, offset)
}

func (s *drService) InitiateRestore(ctx context.Context, backupID uuid.UUID, targetEnv string, performedBy uuid.UUID) (*models.RestoreJob, error) {
	// Check if backup exists
	backup, err := s.repo.GetBackup(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %w", err)
	}
	if backup.Status != models.BackupStatusCompleted {
		return nil, fmt.Errorf("cannot restore from incomplete backup (status: %s)", backup.Status)
	}

	job := &models.RestoreJob{
		ID:                uuid.New(),
		BackupID:          backupID,
		Status:            models.RestoreStatusPending,
		TargetEnvironment: targetEnv,
		PerformedBy:       performedBy,
		StartedAt:         time.Now(),
	}

	if err := s.repo.CreateRestoreJob(ctx, job); err != nil {
		return nil, err
	}

	s.logger.Info("Restore job initiated", zap.String("id", job.ID.String()), zap.String("env", targetEnv))
	
	// Simulate
	go s.processRestore(job.ID)

	return job, nil
}

func (s *drService) processRestore(id uuid.UUID) {
	ctx := context.Background()
	time.Sleep(2 * time.Second)
	
	_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusInProgress, "Starting restore process...")
	
	time.Sleep(5 * time.Second)
	
	_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusCompleted, "Restore completed successfully.")
}

func (s *drService) GetRestoreJob(ctx context.Context, id uuid.UUID) (*models.RestoreJob, error) {
	return s.repo.GetRestoreJob(ctx, id)
}

func (s *drService) ListRestoreJobs(ctx context.Context, limit, offset int) ([]*models.RestoreJob, int, error) {
	return s.repo.ListRestoreJobs(ctx, limit, offset)
}

func (s *drService) VerifyBackup(ctx context.Context, backupID uuid.UUID, verifiedBy uuid.UUID) (*models.BackupVerification, error) {
	// Get the backup first
	backup, err := s.repo.GetBackup(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %w", err)
	}

	if backup.Status != models.BackupStatusCompleted {
		return nil, fmt.Errorf("cannot verify incomplete backup (status: %s)", backup.Status)
	}

	// Create verification record
	verification := &models.BackupVerification{
		ID:         uuid.New(),
		BackupID:   backupID,
		Status:     "VERIFIED",
		VerifiedAt: time.Now(),
		VerifiedBy: verifiedBy,
	}

	// Simulate checksum calculation
	hash := sha256.Sum256([]byte(backup.S3Key + backup.ID.String()))
	verification.Checksum = hex.EncodeToString(hash[:])

	if err := s.repo.CreateBackupVerification(ctx, verification); err != nil {
		return nil, fmt.Errorf("failed to create verification record: %w", err)
	}

	s.logger.Info("Backup verified", zap.String("backup_id", backupID.String()), zap.String("checksum", verification.Checksum))

	return verification, nil
}

func (s *drService) ListStoragePools(ctx context.Context) ([]*models.StoragePool, error) {
	// For demo purposes, return mock data if no pools exist
	pools, err := s.repo.ListStoragePools(ctx)
	if err != nil {
		s.logger.Error("Failed to list storage pools", zap.Error(err))
		// Return demo data as fallback
		return []*models.StoragePool{
			{
				ID:             uuid.New(),
				Name:           "Primary Storage",
				Status:         models.StoragePoolStatusActive,
				TotalBytes:     1000 * 1024 * 1024 * 1024, // 1TB
				UsedBytes:      350 * 1024 * 1024 * 1024,   // 350GB
				AvailableBytes: 650 * 1024 * 1024 * 1024,   // 650GB
				Location:       "us-east-1",
				BackupCount:    24,
				CreatedAt:      time.Now().Add(-90 * 24 * time.Hour),
				UpdatedAt:      time.Now(),
			},
			{
				ID:             uuid.New(),
				Name:           "DR Storage",
				Status:         models.StoragePoolStatusActive,
				TotalBytes:     2000 * 1024 * 1024 * 1024, // 2TB
				UsedBytes:      800 * 1024 * 1024 * 1024,   // 800GB
				AvailableBytes: 1200 * 1024 * 1024 * 1024,  // 1.2TB
				Location:       "us-west-2",
				BackupCount:    48,
				CreatedAt:      time.Now().Add(-180 * 24 * time.Hour),
				UpdatedAt:      time.Now(),
			},
			{
				ID:             uuid.New(),
				Name:           "Archive Storage",
				Status:         models.StoragePoolStatusWarning,
				TotalBytes:     5000 * 1024 * 1024 * 1024, // 5TB
				UsedBytes:      4200 * 1024 * 1024 * 1024, // 4.2TB
				AvailableBytes: 800 * 1024 * 1024 * 1024,   // 800GB
				Location:       "eu-central-1",
				BackupCount:    120,
				CreatedAt:      time.Now().Add(-365 * 24 * time.Hour),
				UpdatedAt:      time.Now().Add(-1 * 24 * time.Hour),
			},
		}, nil
	}
	return pools, nil
}
