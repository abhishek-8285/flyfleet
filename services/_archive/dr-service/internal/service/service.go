package service

import (
	"context"
	"time"

	"github.com/flyfleet/dr-service/internal/models"
	"github.com/flyfleet/dr-service/internal/queue"
	"github.com/flyfleet/dr-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DRService struct {
	repo   *repository.PostgresRepository
	logger *zap.Logger
	queue  *queue.QueueClient
}

func NewDRService(repo *repository.PostgresRepository, logger *zap.Logger, queue *queue.QueueClient) *DRService {
	return &DRService{repo: repo, logger: logger, queue: queue}
}

func (s *DRService) ListBackups(ctx context.Context) ([]models.Backup, error) {
	return s.repo.ListBackups(ctx)
}

func (s *DRService) GetBackupDetails(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	return s.repo.GetBackup(ctx, id)
}

func (s *DRService) CreateBackup(ctx context.Context, req *models.CreateBackupRequest) (*models.Backup, error) {
	id := uuid.New()
	initiatedBy := req.InitiatedBy
	if initiatedBy == "" {
		initiatedBy = "System"
	}
	remarks := req.Remarks

	backup := &models.Backup{
		ID:            id,
		BackupType:    req.BackupType,
		Status:        "SCHEDULED",
		Location:      &req.Location,
		RetentionDays: req.RetentionDays,
		Remarks:       &remarks,
		InitiatedBy:   &initiatedBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateBackup(ctx, backup); err != nil {
		return nil, err
	}

	// Enqueue job to Redis
	if err := s.queue.EnqueueBackup(backup.ID); err != nil {
		s.logger.Error("Failed to enqueue backup job", zap.Error(err))
		// We might want to revert the DB status here, but skipping for now
		return nil, err
	}
	s.logger.Info("Backup job enqueued", zap.String("id", id.String()))

	return backup, nil
}

func (s *DRService) ListStoragePools(ctx context.Context) ([]models.StoragePool, error) {
	return s.repo.ListStoragePools(ctx)
}

// --- Restore Operations ---

func (s *DRService) ListRestores(ctx context.Context) ([]models.Restore, error) {
	return s.repo.ListRestores(ctx)
}

func (s *DRService) GetRestoreDetails(ctx context.Context, id uuid.UUID) (*models.Restore, error) {
	// TODO: Fetch logs separately or embed them
	return s.repo.GetRestore(ctx, id)
}

func (s *DRService) GetRestoreLogs(ctx context.Context, id uuid.UUID) ([]models.RestoreLog, error) {
	return s.repo.GetRestoreLogs(ctx, id)
}

func (s *DRService) InitiateRestore(ctx context.Context, req *models.CreateRestoreRequest) (*models.Restore, error) {
	backupID, err := uuid.Parse(req.BackupID)
	if err != nil {
		return nil, err
	}

	restore := &models.Restore{
		ID:          uuid.New(),
		BackupID:    backupID,
		RestoreType: "FULL",
		TargetEnv:   req.TargetEnv,
		Status:      "PENDING",
		InitiatedBy: &req.RequestedBy,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateRestore(ctx, restore); err != nil {
		return nil, err
	}

	if err := s.queue.EnqueueRestore(restore.ID); err != nil {
		s.logger.Error("Failed to enqueue restore job", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Restore job enqueued", zap.String("id", restore.ID.String()))

	return restore, nil
}

func ptr(s string) *string { return &s }

// ========== DR Testing Service Methods ==========

func (s *DRService) ListDRTests(ctx context.Context) ([]models.DRTest, error) {
	return s.repo.ListDRTests(ctx)
}

func (s *DRService) GetDRTestDetails(ctx context.Context, id uuid.UUID) (*models.DRTest, []models.DRTestStep, error) {
	test, err := s.repo.GetDRTest(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	steps, err := s.repo.GetDRTestSteps(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return test, steps, nil
}

func (s *DRService) GetDRTestLogs(ctx context.Context, id uuid.UUID) ([]models.DRTestLog, error) {
	return s.repo.GetDRTestLogs(ctx, id)
}

func (s *DRService) CreateDRTest(ctx context.Context, req *models.CreateDRTestRequest) (*models.DRTest, error) {
	id := uuid.New()
	testID := time.Now().Format("DR-2006-0102")

	initiatedBy := req.InitiatedBy
	if initiatedBy == "" {
		initiatedBy = "System"
	}

	var scheduledAt *time.Time
	if req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err == nil {
			scheduledAt = &t
		}
	}

	test := &models.DRTest{
		ID:              id,
		TestID:          testID,
		Name:            req.Name,
		Description:     &req.Description,
		TestType:        req.TestType,
		TargetEnv:       req.TargetEnv,
		Status:          "SCHEDULED",
		ProgressPercent: 0,
		InitiatedBy:     &initiatedBy,
		SuccessCriteria: &req.SuccessCriteria,
		ScheduledAt:     scheduledAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create default steps based on test type
	steps := []models.DRTestStep{
		{ID: uuid.New(), StepOrder: 1, StepName: "Initialization", Status: "PENDING"},
		{ID: uuid.New(), StepOrder: 2, StepName: "Storage Replication", Status: "PENDING"},
		{ID: uuid.New(), StepOrder: 3, StepName: "Service Failover", Status: "PENDING"},
		{ID: uuid.New(), StepOrder: 4, StepName: "Data Integrity", Status: "PENDING"},
		{ID: uuid.New(), StepOrder: 5, StepName: "Final Report", Status: "PENDING"},
	}

	if err := s.repo.CreateDRTest(ctx, test, steps); err != nil {
		return nil, err
	}

	// Enqueue the DR test job
	if err := s.queue.EnqueueDRTest(test.ID); err != nil {
		s.logger.Error("Failed to enqueue DR test job", zap.Error(err))
		return nil, err
	}
	s.logger.Info("DR test job enqueued", zap.String("id", id.String()))

	return test, nil
}

func (s *DRService) AbortDRTest(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateDRTestStatus(ctx, id, "ABORTED", 0)
}

func (s *DRService) GetDRTestMetrics(ctx context.Context) (*models.DRTestMetrics, error) {
	return s.repo.GetDRTestMetrics(ctx)
}

func (s *DRService) GetDRTestSettings(ctx context.Context) (*models.DRTestSettings, error) {
	return s.repo.GetDRTestSettings(ctx)
}

func (s *DRService) UpdateDRTestSettings(ctx context.Context, req *models.UpdateDRTestSettingsRequest) error {
	settings := &models.DRTestSettings{
		ID:                 uuid.New(),
		ScheduleEnabled:    req.ScheduleEnabled,
		Frequency:          req.Frequency,
		ScheduleDay:        req.ScheduleDay,
		ScheduleTime:       req.ScheduleTime,
		Intensity:          req.Intensity,
		NotificationEmails: req.NotificationEmails,
	}
	return s.repo.UpdateDRTestSettings(ctx, settings)
}
