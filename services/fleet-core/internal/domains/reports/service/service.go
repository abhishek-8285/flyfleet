package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/reports/models"
	"github.com/flyfleet/fleet-core/internal/domains/reports/queue"
	"github.com/flyfleet/fleet-core/internal/domains/reports/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReportService struct {
	repo   repository.Repository
	queue  *queue.QueueClient
	logger *zap.Logger
}

func NewReportService(repo repository.Repository, queue *queue.QueueClient, logger *zap.Logger) *ReportService {
	return &ReportService{
		repo:   repo,
		queue:  queue,
		logger: logger,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, report *models.ScheduledReport) (uuid.UUID, error) {
	report.ID = uuid.New()
	report.Status = "ACTIVE"
	report.CreatedAt = time.Now()

	userID := s.getUserID(ctx)
	report.CreatedBy = userID
	report.UpdatedBy = userID

	if err := s.repo.CreateReport(ctx, report); err != nil {
		return uuid.Nil, err
	}
	s.logger.Info("Created new scheduled report", zap.String("id", report.ID.String()))
	return report.ID, nil
}

func (s *ReportService) GetReport(ctx context.Context, id uuid.UUID) (*models.ScheduledReport, error) {
	return s.repo.GetReport(ctx, id)
}

func (s *ReportService) UpdateReport(ctx context.Context, report *models.ScheduledReport) error {
	report.UpdatedBy = s.getUserID(ctx)
	return s.repo.UpdateReport(ctx, report)
}

func (s *ReportService) ListReports(ctx context.Context, page, limit int) ([]models.ScheduledReport, int, error) {
	return s.repo.ListReports(ctx, page, limit)
}

func (s *ReportService) PauseReport(ctx context.Context, id uuid.UUID) error {
	return s.repo.PauseReport(ctx, id)
}

func (s *ReportService) ResumeReport(ctx context.Context, id uuid.UUID) error {
	return s.repo.ResumeReport(ctx, id)
}

// Execution Logic (Simplified for Prototype)

func (s *ReportService) ExecuteReport(ctx context.Context, reportID uuid.UUID) (uuid.UUID, error) {
	// 1. Fetch Report Config
	report, err := s.repo.GetReport(ctx, reportID)
	if err != nil {
		return uuid.Nil, err
	}
	if report == nil {
		return uuid.Nil, fmt.Errorf("report not found")
	}

	userID := s.getUserID(ctx)
	// 2. Create Execution Record
	execID := uuid.New()
	startTime := time.Now()
	exec := &models.ReportExecution{
		ID:        execID,
		ReportID:  reportID,
		StartedAt: &startTime,
		Status:    "RUNNING",
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return uuid.Nil, err
	}

	// 3. Enqueue Job
	if err := s.queue.EnqueueReportGeneration(execID, reportID, userID); err != nil {
		s.logger.Error("Failed to enqueue report generation task", zap.Error(err))

		// Fail the execution record
		errMsg := err.Error()
		if updateErr := s.repo.UpdateExecutionStatus(ctx, execID, "FAILED", &errMsg, nil, userID); updateErr != nil {
			s.logger.Error("Failed to update execution status to FAILED", zap.Error(updateErr))
		}

		return uuid.Nil, err
	}

	s.logger.Info("Enqueued report generation task", zap.String("exec_id", execID.String()))
	return execID, nil
}

func (s *ReportService) ListExecutions(ctx context.Context, reportID uuid.UUID, page, limit int) ([]models.ReportExecution, int, error) {
	return s.repo.ListExecutions(ctx, reportID, page, limit)
}

func (s *ReportService) GetSettings(ctx context.Context) (*models.GlobalSettings, error) {
	return s.repo.GetSettings(ctx)
}

func (s *ReportService) UpdateSettings(ctx context.Context, settings *models.GlobalSettings) error {
	settings.UpdatedBy = s.getUserID(ctx)
	return s.repo.UpdateSettings(ctx, settings)
}

func (s *ReportService) getUserID(ctx context.Context) uuid.UUID {
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
