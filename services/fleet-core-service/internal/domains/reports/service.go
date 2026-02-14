package reports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	CreateReport(ctx context.Context, req CreateReportRequest) (*Report, error)
	GetReport(ctx context.Context, id uuid.UUID) (*Report, error)
	ListReports(ctx context.Context, limit, offset int) ([]Report, error)
}

type service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) CreateReport(ctx context.Context, req CreateReportRequest) (*Report, error) {
	report := &Report{
		ID:         uuid.New(),
		Type:       req.Type,
		Status:     ReportStatusPending,
		Parameters: req.Parameters,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	// Simulate background processing
	go func() {
		time.Sleep(2 * time.Second)
		s.repo.UpdateStatus(context.Background(), report.ID, ReportStatusCompleted, "https://example.com/report.pdf")
	}()

	return report, nil
}

func (s *service) GetReport(ctx context.Context, id uuid.UUID) (*Report, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListReports(ctx context.Context, limit, offset int) ([]Report, error) {
	return s.repo.List(ctx, limit, offset)
}
