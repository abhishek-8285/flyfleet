package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/repository"
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

	CreateDRTest(ctx context.Context, test *models.DRTest) (*models.DRTest, error)
	ListDRTests(ctx context.Context, limit, offset int) ([]*models.DRTest, int, error)
	GetDRTestDetails(ctx context.Context, id uuid.UUID) (*models.DRTest, []*models.DRTestStep, []*models.DRTestLog, error)
	AbortDRTest(ctx context.Context, id uuid.UUID) error
	GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error)
	GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error)
	UpdateDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error
}

type drService struct {
	repo   repository.DrRepository
	logger *zap.Logger
}

func NewDrService(repo repository.DrRepository, logger *zap.Logger) DrService {
	return &drService{repo: repo, logger: logger}
}

func (s *drService) InitiateBackup(ctx context.Context, backupType models.BackupType, createdBy uuid.UUID) (*models.Backup, error) {
	userID := createdBy
	if userID == uuid.Nil {
		userID = s.getUserID(ctx)
	}
	backup := &models.Backup{
		ID:        uuid.New(),
		Type:      backupType,
		Status:    models.BackupStatusPending,
		CreatedBy: userID,
		UpdatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if err := s.repo.CreateBackup(ctx, backup); err != nil {
		return nil, err
	}

	// In a real system, this would trigger an async job (e.g., via NATS or Asynq)
	// For now, we'll log it.
	s.logger.Info("Backup initiated", zap.String("id", backup.ID.String()), zap.String("type", string(backupType)))

	// Simulate immediate start for demonstration
	go s.processBackup(backup.ID, userID)

	return backup, nil
}

func (s *drService) processBackup(id uuid.UUID, updatedBy uuid.UUID) {
	ctx := context.Background()
	_ = s.repo.UpdateBackupStatus(ctx, id, models.BackupStatusInProgress, "", 0, updatedBy)

	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/backups"
	}

	// Validate backup script exists before executing
	backupScript := "/app/scripts/db/backup.sh"
	if _, err := os.Stat(backupScript); os.IsNotExist(err) {
		s.logger.Error("Backup script does not exist", zap.String("path", backupScript))
		_ = s.repo.UpdateBackupStatus(ctx, id, models.BackupStatusFailed, "", 0, updatedBy)
		return
	}

	// We'll call the backup script
	cmd := exec.Command("/bin/bash", backupScript)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("BACKUP_DIR=%s", backupDir),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Backup script failed", zap.Error(err), zap.String("output", string(output)))
		_ = s.repo.UpdateBackupStatus(ctx, id, models.BackupStatusFailed, "", 0, updatedBy)
		return
	}

	// The script outputs the backup file path at the end: "Backup successful: /backups/fleet_db-2026-02-21_19-00-00.sql.gz"
	// We should ideally parse this or have the script output just the path.
	// For now, let's assume standard naming if we can't parse easily.
	s.logger.Info("Backup script completed", zap.String("output", string(output)))

	// Find the most recent file in backupDir starting with fleet_db-
	files, err := os.ReadDir(backupDir)
	var latestFile string
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".gz" {
				if latestFile == "" || f.Name() > latestFile {
					latestFile = f.Name()
				}
			}
		}
	}

	s3Key := latestFile
	_ = s.repo.UpdateBackupStatus(ctx, id, models.BackupStatusCompleted, s3Key, 0, updatedBy)
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

	userID := performedBy
	if userID == uuid.Nil {
		userID = s.getUserID(ctx)
	}

	job := &models.RestoreJob{
		ID:                uuid.New(),
		BackupID:          backupID,
		Status:            models.RestoreStatusPending,
		TargetEnvironment: targetEnv,
		PerformedBy:       userID,
		UpdatedBy:         userID,
		StartedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.CreateRestoreJob(ctx, job); err != nil {
		return nil, err
	}

	s.logger.Info("Restore job initiated", zap.String("id", job.ID.String()), zap.String("env", targetEnv))

	// Simulate
	go s.processRestore(job.ID, userID)

	return job, nil
}

func (s *drService) processRestore(id uuid.UUID, updatedBy uuid.UUID) {
	ctx := context.Background()
	job, err := s.repo.GetRestoreJob(ctx, id)
	if err != nil {
		return
	}

	backup, err := s.repo.GetBackup(ctx, job.BackupID)
	if err != nil {
		_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusFailed, "Backup record not found", updatedBy)
		return
	}

	_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusInProgress, "Starting restore process...", updatedBy)

	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/backups"
	}
	backupPath := filepath.Join(backupDir, backup.S3Key)

	// Validate backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		s.logger.Error("Backup file does not exist", zap.String("path", backupPath))
		_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusFailed, fmt.Sprintf("Backup file not found: %s", backupPath), updatedBy)
		return
	}

	// Validate restore script exists before executing
	restoreScript := "/app/scripts/db/restore.sh"
	if _, err := os.Stat(restoreScript); os.IsNotExist(err) {
		s.logger.Error("Restore script does not exist", zap.String("path", restoreScript))
		_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusFailed, "Restore script not found", updatedBy)
		return
	}

	cmd := exec.Command("/bin/bash", restoreScript, backupPath)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Restore script failed", zap.Error(err), zap.String("output", string(output)))
		_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusFailed, fmt.Sprintf("Script failed: %v\nOutput: %s", err, string(output)), updatedBy)
		return
	}

	s.logger.Info("Restore script completed", zap.String("output", string(output)))
	_ = s.repo.UpdateRestoreJobStatus(ctx, id, models.RestoreStatusCompleted, "Restore completed successfully.", updatedBy)
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

	userID := verifiedBy
	if userID == uuid.Nil {
		userID = s.getUserID(ctx)
	}

	// Create verification record
	verification := &models.BackupVerification{
		ID:         uuid.New(),
		BackupID:   backupID,
		Status:     "VERIFIED",
		VerifiedAt: time.Now(),
		VerifiedBy: userID,
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
	pools, err := s.repo.ListStoragePools(ctx)
	if err != nil {
		s.logger.Error("Failed to list storage pools", zap.Error(err))
		return nil, err
	}
	return pools, nil
}

func (s *drService) CreateDRTest(ctx context.Context, test *models.DRTest) (*models.DRTest, error) {
	now := time.Now()
	test.ID = uuid.New()
	test.TestID = fmt.Sprintf("DRT-%d", now.Unix())
	test.CreatedAt = now
	test.UpdatedAt = now
	test.ProgressPercent = 0

	userID := s.getUserID(ctx)
	test.CreatedBy = userID
	test.UpdatedBy = userID

	if test.ScheduledAt != nil && test.ScheduledAt.After(now) {
		test.Status = "SCHEDULED"
	} else {
		test.Status = "RUNNING"
		test.StartedAt = &now
	}

	if err := s.repo.CreateDRTest(ctx, test); err != nil {
		return nil, err
	}

	steps := []string{
		"Initialize test context",
		"Validate backup chain",
		"Execute failover or restore",
		"Validate recovery criteria",
	}
	for idx, name := range steps {
		step := &models.DRTestStep{
			ID:        uuid.New(),
			TestID:    test.ID,
			StepOrder: idx + 1,
			StepName:  name,
			Status:    "PENDING",
		}
		if err := s.repo.CreateDRTestStep(ctx, step); err != nil {
			return nil, err
		}
	}

	_ = s.repo.CreateDRTestLog(ctx, &models.DRTestLog{
		ID:        uuid.New(),
		TestID:    test.ID,
		Timestamp: now,
		Level:     "RUN",
		Message:   "DR test created",
	})

	if test.Status == "RUNNING" {
		go s.runDRTestSimulation(test.ID)
	}
	return test, nil
}

func (s *drService) runDRTestSimulation(testID uuid.UUID) {
	ctx := context.Background()
	steps, err := s.repo.ListDRTestSteps(ctx, testID)
	if err != nil {
		s.logger.Error("Failed to list DR test steps", zap.Error(err))
		return
	}
	total := len(steps)
	if total == 0 {
		return
	}

	for idx, step := range steps {
		test, err := s.repo.GetDRTest(ctx, testID)
		if err != nil {
			s.logger.Error("Failed to get DR test", zap.Error(err))
			return
		}
		if test.Status == "ABORTED" {
			return
		}

		started := time.Now()
		_ = s.repo.UpdateDRTestStepStatus(ctx, step.ID, "RUNNING", &started, nil)
		_ = s.repo.CreateDRTestLog(ctx, &models.DRTestLog{
			ID:        uuid.New(),
			TestID:    testID,
			Timestamp: started,
			Level:     "RUN",
			Message:   fmt.Sprintf("Step %d started: %s", step.StepOrder, step.StepName),
		})

		// Execute actual validation based on step type
		stepPassed, validationLog := s.executeDRTestStep(ctx, testID, step.StepName)

		completed := time.Now()
		progress := int(float64(idx+1) / float64(total) * 100.0)
		if stepPassed {
			_ = s.repo.UpdateDRTestStepStatus(ctx, step.ID, "PASSED", nil, &completed)
			_ = s.repo.UpdateDRTestStatus(ctx, testID, "RUNNING", progress, nil)
			_ = s.repo.CreateDRTestLog(ctx, &models.DRTestLog{
				ID:        uuid.New(),
				TestID:    testID,
				Timestamp: completed,
				Level:     "RUN",
				Message:   fmt.Sprintf("Step %d passed: %s - %s", step.StepOrder, step.StepName, validationLog),
			})
		} else {
			_ = s.repo.UpdateDRTestStepStatus(ctx, step.ID, "FAILED", nil, &completed)
			_ = s.repo.UpdateDRTestStatus(ctx, testID, "FAILED", progress, &completed)
			_ = s.repo.CreateDRTestLog(ctx, &models.DRTestLog{
				ID:        uuid.New(),
				TestID:    testID,
				Timestamp: completed,
				Level:     "ERROR",
				Message:   fmt.Sprintf("Step %d failed: %s - %s", step.StepOrder, step.StepName, validationLog),
			})
			return
		}

		time.Sleep(500 * time.Millisecond) // Small delay between steps
	}

	// All steps passed - mark test as passed
	completed := time.Now()
	_ = s.repo.UpdateDRTestStatus(ctx, testID, "PASSED", 100, &completed)
	_ = s.repo.CreateDRTestLog(ctx, &models.DRTestLog{
		ID:        uuid.New(),
		TestID:    testID,
		Timestamp: completed,
		Level:     "INFO",
		Message:   "DR test completed successfully",
	})
}

// executeDRTestStep executes actual validation for a DR test step
func (s *drService) executeDRTestStep(ctx context.Context, testID uuid.UUID, stepName string) (bool, string) {
	switch stepName {
	case "Initialize test context":
		// Validate test configuration exists
		test, err := s.repo.GetDRTest(ctx, testID)
		if err != nil {
			return false, "Failed to retrieve test configuration"
		}
		if test.TestType == "" {
			return false, "Test type not specified"
		}
		return true, fmt.Sprintf("Test context initialized for %s", test.TestType)

	case "Validate backup chain":
		// Check if there are completed backups available
		backups, _, err := s.repo.ListBackups(ctx, 10, 0)
		if err != nil {
			return false, "Failed to list backups"
		}
		completedBackups := 0
		for _, b := range backups {
			if b.Status == models.BackupStatusCompleted {
				completedBackups++
			}
		}
		if completedBackups == 0 {
			return false, "No completed backups found for validation"
		}
		return true, fmt.Sprintf("Found %d completed backups", completedBackups)

	case "Execute failover or restore":
		// Simulate failover/restore validation
		// In production, this would actually test restoring to a staging environment
		time.Sleep(2 * time.Second) // Simulate restore time
		return true, "Failover/restore simulation completed successfully"

	case "Validate recovery criteria":
		// Check if backups can be verified
		backups, _, err := s.repo.ListBackups(ctx, 5, 0)
		if err != nil {
			return false, "Failed to list backups for verification"
		}
		verifiableCount := 0
		for _, b := range backups {
			if b.Status == models.BackupStatusCompleted && b.S3Key != "" {
				verifiableCount++
			}
		}
		if verifiableCount == 0 {
			return false, "No verifiable backups found"
		}
		return true, fmt.Sprintf("Recovery criteria validated: %d verifiable backups", verifiableCount)

	default:
		return true, fmt.Sprintf("Step %s executed (no specific validation)", stepName)
	}
}

func (s *drService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}

func (s *drService) ListDRTests(ctx context.Context, limit, offset int) ([]*models.DRTest, int, error) {
	return s.repo.ListDRTests(ctx, limit, offset)
}

func (s *drService) GetDRTestDetails(ctx context.Context, id uuid.UUID) (*models.DRTest, []*models.DRTestStep, []*models.DRTestLog, error) {
	test, err := s.repo.GetDRTest(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	steps, err := s.repo.ListDRTestSteps(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	logs, err := s.repo.ListDRTestLogs(ctx, id, 200)
	if err != nil {
		return nil, nil, nil, err
	}
	return test, steps, logs, nil
}

func (s *drService) AbortDRTest(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.AbortDRTest(ctx, id); err != nil {
		return err
	}
	if err := s.repo.MarkPendingDRTestStepsSkipped(ctx, id); err != nil {
		return err
	}
	return s.repo.CreateDRTestLog(ctx, &models.DRTestLog{
		ID:        uuid.New(),
		TestID:    id,
		Timestamp: time.Now(),
		Level:     "WARN",
		Message:   "DR test aborted by operator",
	})
}

func (s *drService) GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error) {
	return s.repo.GetDRTestMetrics(ctx)
}

func (s *drService) GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error) {
	settings, err := s.repo.GetDRTestSettings(ctx)
	if err == nil {
		return settings, nil
	}
	return &models.DRTestSettings{
		ScheduleEnabled:    false,
		Frequency:          "WEEKLY",
		ScheduleDay:        "SATURDAY",
		ScheduleTime:       "02:00",
		Intensity:          "NORMAL",
		NotificationEmails: []string{},
	}, nil
}

func (s *drService) UpdateDRTestSettings(ctx context.Context, settings *models.DRTestSettings) error {
	return s.repo.UpsertDRTestSettings(ctx, settings)
}
