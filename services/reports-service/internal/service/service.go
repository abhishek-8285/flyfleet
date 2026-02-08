package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/flyfleet/reports-service/internal/models"
	"github.com/flyfleet/reports-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReportService struct {
	repo   repository.Repository
	logger *zap.Logger
}

func NewReportService(repo repository.Repository, logger *zap.Logger) *ReportService {
	return &ReportService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, report *models.ScheduledReport) (uuid.UUID, error) {
	report.ID = uuid.New()
	report.Status = "ACTIVE"
	report.CreatedAt = time.Now()
	
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

	// 2. Create Execution Record
	execID := uuid.New()
	startTime := time.Now()
	exec := &models.ReportExecution{
		ID:        execID,
		ReportID:  reportID,
		StartedAt: &startTime,
		Status:    "RUNNING",
	}
	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return uuid.Nil, err
	}

	// 3. Simulate Async Job (In real system, push to Redis/Kafka)
	go func() {
		// Verify safety: In a real app we'd need context handling here.
		// Emulate processing time
		time.Sleep(2 * time.Second)
		
		// Simulate Success/Failure
		outcome := "SUCCESS"
		if rand.Float32() < 0.1 {
			outcome = "FAILED"
		}

		s.logger.Info("Completed report execution", zap.String("exec_id", execID.String()), zap.String("outcome", outcome))
		
		// Update DB (Ignoring error for prototype simplicity)
		// access postgres directly or add UpdateExecution to repo interface if needed.
		// For now, we assume it's done. 
	}()

	return execID, nil
}

func (s *ReportService) ListExecutions(ctx context.Context, reportID uuid.UUID, page, limit int) ([]models.ReportExecution, int, error) {
	return s.repo.ListExecutions(ctx, reportID, page, limit)
}

func (s *ReportService) GetSettings(ctx context.Context) (*models.GlobalSettings, error) {
	return s.repo.GetSettings(ctx)
}

func (s *ReportService) UpdateSettings(ctx context.Context, settings *models.GlobalSettings) error {
	return s.repo.UpdateSettings(ctx, settings)
}
