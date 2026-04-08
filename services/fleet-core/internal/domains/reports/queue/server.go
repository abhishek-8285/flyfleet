package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/reports/client"
	"github.com/flyfleet/fleet-core/internal/domains/reports/repository"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type WorkerServer struct {
	server      *asynq.Server
	mux         *asynq.ServeMux
	repo        repository.Repository
	notifClient *client.NotificationClient
	logger      *zap.Logger
}

func NewWorkerServer(redisAddr string, repo repository.Repository, notifClient *client.NotificationClient, logger *zap.Logger) *WorkerServer {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 5, // Limit report generation concurrency
			Queues: map[string]int{
				"default": 1,
			},
			Logger: logger.Sugar(),
		},
	)

	mux := asynq.NewServeMux()
	ws := &WorkerServer{
		server:      server,
		mux:         mux,
		repo:        repo,
		notifClient: notifClient,
		logger:      logger,
	}

	mux.HandleFunc(TypeReportGenerate, ws.HandleReportGenerationTask)

	return ws
}

func (ws *WorkerServer) Run() error {
	return ws.server.Run(ws.mux)
}

func (ws *WorkerServer) HandleReportGenerationTask(ctx context.Context, t *asynq.Task) error {
	var payload ReportGeneratePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	execID := payload.ExecutionID
	reportID := payload.ReportID

	ws.logger.Info("Processing report generation task", zap.String("exec_id", execID.String()), zap.String("report_id", reportID.String()))

	// 1. Update Status to RUNNING
	if err := ws.repo.UpdateExecutionStatus(ctx, execID, "RUNNING", nil, nil, payload.UserID); err != nil {
		ws.logger.Error("Failed to update execution status to RUNNING", zap.Error(err))
	}

	// 2. Fetch Report Details (Mock Fetch)
	// In real app: report, err := ws.repo.GetReport(ctx, reportID)
	// For now we assume we have what we need.

	// 3. Generate Report (Mock)
	time.Sleep(2 * time.Second) // Simulate PDF generation time
	mockFileURL := fmt.Sprintf("https://flyfleet-bucket.s3.amazonaws.com/reports/%s.pdf", execID.String())

	// 4. Send Email
	// In real app: fetch recipients from report.Recipients
	recipient := "admin@flyfleet.com" // Default/Fallback
	subject := "Your Scheduled Report is Ready"
	body := fmt.Sprintf("Hello,\n\nYour report has been generated successfully.\nDownload it here: %s\n\n- FlyFleet Team", mockFileURL)
	metadata := map[string]interface{}{
		"report_id": reportID.String(),
		"exec_id":   execID.String(),
	}

	if err := ws.notifClient.SendEmail(ctx, recipient, subject, body, metadata); err != nil {
		ws.logger.Error("Failed to send report email", zap.Error(err))
		// We verify but don't fail the generation if email fails?
		// Or maybe we treat email delivery failure as partial success.
		// For now, let's log error but mark generation as COMPLETED.
	}

	ws.logger.Info("Report generation completed", zap.String("exec_id", execID.String()))

	// 5. Update Status to COMPLETED
	if err := ws.repo.UpdateExecutionStatus(ctx, execID, "COMPLETED", nil, &mockFileURL, payload.UserID); err != nil {
		ws.logger.Error("Failed to update execution status to COMPLETED", zap.Error(err))
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}
