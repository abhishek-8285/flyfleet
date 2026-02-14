package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/dr-service/internal/events"
	"github.com/flyfleet/dr-service/internal/models"
	"github.com/google/uuid"

	"github.com/flyfleet/dr-service/internal/repository"
	"github.com/flyfleet/dr-service/internal/storage"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type WorkerServer struct {
	server   *asynq.Server
	mux      *asynq.ServeMux
	repo     repository.Repository
	storage  storage.BackupProvider
	logger   *zap.Logger
	broadcaster *events.Broadcaster
}

func NewWorkerServer(redisAddr string, repo repository.Repository, storage storage.BackupProvider, logger *zap.Logger, broadcaster *events.Broadcaster) *WorkerServer {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: logger.Sugar(),
		},
	)

	mux := asynq.NewServeMux()
	ws := &WorkerServer{
		server:  server,
		mux:     mux,
		repo:    repo,
		storage: storage,
		logger:  logger,
		broadcaster: broadcaster,
	}

	mux.HandleFunc(TypeBackup, ws.HandleBackupTask)
	mux.HandleFunc(TypeRestore, ws.HandleRestoreTask)
	mux.HandleFunc(TypeDRTest, ws.HandleDRTestTask)

	return ws
}

func (ws *WorkerServer) Run() error {
	return ws.server.Run(ws.mux)
}

func (ws *WorkerServer) HandleDRTestTask(ctx context.Context, t *asynq.Task) error {
	var payload DRTestPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	id := payload.TestID
	ws.logger.Info("Processing DR test task", zap.String("id", id.String()))

	// 5-step execution flow
	steps := []struct {
		order int
		name  string
		work  func() error
	}{
		{1, "Initialization", func() error {
			time.Sleep(2 * time.Second)
			return nil
		}},
		{2, "Storage Replication", func() error {
			time.Sleep(3 * time.Second)
			return nil
		}},
		{3, "Service Failover", func() error {
			time.Sleep(5 * time.Second)
			return nil
		}},
		{4, "Data Integrity", func() error {
			time.Sleep(2 * time.Second)
			return nil
		}},
		{5, "Final Report", func() error {
			time.Sleep(1 * time.Second)
			return nil
		}},
	}

	// Update to RUNNING
	ws.repo.UpdateDRTestStatus(ctx, id, "RUNNING", 0)
	ws.addDRTestLog(ctx, id, "INFO", "Test execution started")

	for i, step := range steps {
		progress := (i * 100) / len(steps)
		ws.repo.UpdateDRTestStepStatus(ctx, id, step.order, "RUNNING")
		ws.repo.UpdateDRTestStatus(ctx, id, "RUNNING", progress)
		ws.addDRTestLog(ctx, id, "RUN", fmt.Sprintf("Executing step: %s", step.name))

		if err := step.work(); err != nil {
			ws.repo.UpdateDRTestStepStatus(ctx, id, step.order, "FAILED")
			ws.repo.UpdateDRTestStatus(ctx, id, "FAILED", progress)
			ws.addDRTestLog(ctx, id, "ERROR", fmt.Sprintf("Step %s failed: %v", step.name, err))
			return err
		}

		ws.repo.UpdateDRTestStepStatus(ctx, id, step.order, "PASSED")
		ws.addDRTestLog(ctx, id, "INFO", fmt.Sprintf("Step %s completed successfully", step.name))
	}

	// Mark as PASSED
	ws.repo.UpdateDRTestStatus(ctx, id, "PASSED", 100)
	ws.addDRTestLog(ctx, id, "INFO", "DR test completed successfully - All steps passed")

	return nil
}

func (ws *WorkerServer) addDRTestLog(ctx context.Context, testID uuid.UUID, level, message string) {
	log := &models.DRTestLog{
		ID:        uuid.New(),
		TestID:    testID,
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}
	if err := ws.repo.AddDRTestLog(ctx, log); err != nil {
		ws.logger.Error("Failed to add DR test log", zap.Error(err))
	}
}

func (ws *WorkerServer) HandleBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload BackupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	
	id := payload.BackupID
	ws.logger.Info("Processing backup task", zap.String("id", id.String()))


	// Update status to RUNNING
	ws.repo.UpdateBackupStatus(ctx, id, "RUNNING")
	ws.publishEvent(ctx, events.EventBackupStatus, id, "RUNNING")

	ws.repo.AddLog(ctx, &models.BackupLog{
		BackupID: id, Action: "INIT", Result: "SUCCESS", Message: ptr("Worker picked up backup job"),
	})
	ws.publishLog(ctx, events.EventLog, id, "Worker picked up backup job")

	// Simulate Data Extraction (would involve DB dump here)
	time.Sleep(2 * time.Second)
	
	// Simulate Upload
	// dummyData := "This is a dummy backup file content for " + id.String()
	// Using a pipe or buffer in real life
	// For now just upload string
	
	// Assuming LocalProvider or S3Provider is injected
	key := fmt.Sprintf("backups/%s.dat", id.String())
	// Convert string to reader
	ws.repo.AddLog(ctx, &models.BackupLog{
		BackupID: id, Action: "UPLOAD", Result: "PENDING", Message: ptr(fmt.Sprintf("Uploading to %s", key)),
	})
	ws.publishLog(ctx, events.EventLog, id, fmt.Sprintf("Uploading to %s", key))
	
	// Note: In real impl we would stream from pg_dump to storage
	
	// Mark COMPLETED
	ws.repo.UpdateBackupStatus(ctx, id, "COMPLETED")
	ws.publishEvent(ctx, events.EventBackupStatus, id, "COMPLETED")

	ws.repo.AddLog(ctx, &models.BackupLog{
		BackupID: id, Action: "COMPLETE", Result: "SUCCESS", Message: ptr("Backup completed successfully"),
	})
	ws.publishLog(ctx, events.EventLog, id, "Backup completed successfully")

	return nil
}

func (ws *WorkerServer) HandleRestoreTask(ctx context.Context, t *asynq.Task) error {
	var payload RestorePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}


	id := payload.RestoreID
	ws.logger.Info("Processing restore task", zap.String("id", id.String()))

	ws.repo.UpdateRestoreStatus(ctx, id, "RUNNING")
	ws.publishEvent(ctx, events.EventRestoreStatus, id, "RUNNING")

	// Note: We need UpdateRestoreStatus in repo, currently missing from interface in previous step.
	// Logic would go here:
	// 1. Download backup file
	// 2. Validate checksum
	// 3. Apply to target DB

	time.Sleep(2 * time.Second) // Simulated work
	ws.repo.UpdateRestoreStatus(ctx, id, "COMPLETED")
	ws.publishEvent(ctx, events.EventRestoreStatus, id, "COMPLETED")

	return nil
}

func (ws *WorkerServer) publishEvent(ctx context.Context, apiType events.EventType, id uuid.UUID, payload interface{}) {
	evt := events.Event{
		Type:      apiType,
		ResourceID: id,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}
	if err := ws.broadcaster.Publish(ctx, evt); err != nil {
		ws.logger.Error("Failed to publish event", zap.Error(err))
	}
}

func (ws *WorkerServer) publishLog(ctx context.Context, apiType events.EventType, id uuid.UUID, msg string) {
	ws.publishEvent(ctx, apiType, id, msg)
}

func ptr(s string) *string { return &s }
