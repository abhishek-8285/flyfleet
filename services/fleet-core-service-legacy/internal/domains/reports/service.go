package reports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	CreateReport(ctx context.Context, req CreateReportRequest, tenantID, userID uuid.UUID) (*CreateReportResponse, error)
	GetReportStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*ReportStatusResponse, error)
	GetReport(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Report, error)
	ListReports(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]Report, error)

	// Templates
	CreateTemplate(ctx context.Context, tmpl *ReportTemplate) error
	ListTemplates(ctx context.Context) ([]ReportTemplate, error)
}

type service struct {
	repo       Repository
	jobManager JobManager
	cache      *ReportCache // nil = cache disabled
	delivery   *DeliveryService
	logger     *zap.Logger
}

func NewService(repo Repository, jobManager JobManager, logger *zap.Logger) Service {
	return &service{
		repo:       repo,
		jobManager: jobManager,
		delivery:   NewDeliveryService(logger),
		logger:     logger,
	}
}

func NewServiceWithCache(repo Repository, jobManager JobManager, cache *ReportCache, logger *zap.Logger) Service {
	return &service{
		repo:       repo,
		jobManager: jobManager,
		cache:      cache,
		delivery:   NewDeliveryService(logger),
		logger:     logger,
	}
}

func (s *service) CreateReport(ctx context.Context, req CreateReportRequest, tenantID, userID uuid.UUID) (*CreateReportResponse, error) {
	if req.ReportType == "" || req.Format == "" {
		return nil, errors.New("report_type and format are required")
	}

	params := map[string]interface{}{
		"format":  req.Format,
		"from":    req.From,
		"to":      req.To,
		"filters": req.Filters,
	}

	// Enforce max report window (90 days)
	if req.From != "" && req.To != "" {
		from, err1 := time.Parse("2006-01-02", req.From)
		to, err2 := time.Parse("2006-01-02", req.To)
		if err1 == nil && err2 == nil {
			if to.Sub(from) > 90*24*time.Hour {
				return nil, errors.New("report window cannot exceed 90 days")
			}
		}
	}

	// Check cache for identical request
	if s.cache != nil {
		cacheKey := s.cache.CacheKey(tenantID.String(), string(req.ReportType), req.Format, params)
		if cachedID, found := s.cache.Get(ctx, cacheKey); found {
			s.logger.Info("Cache hit — returning existing report", zap.String("report_id", cachedID))
			cachedUUID, _ := uuid.Parse(cachedID)
			return &CreateReportResponse{
				ReportID: cachedUUID,
				Status:   ReportStatusQueued,
			}, nil
		}
	}

	report := &Report{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Type:        req.ReportType,
		Status:      ReportStatusCreated,
		Parameters:  params,
		RequestedBy: userID,
		RequestedAt: time.Now(),
		Progress:    0,
		RetryCount:  0,
	}

	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.Error("Failed to create report record", zap.Error(err))
		return nil, err
	}

	// Audit Log
	s.repo.RecordAuditLog(ctx, &ReportAuditLog{
		ID:        uuid.New(),
		ReportID:  report.ID,
		UserID:    userID,
		Action:    "CREATE",
		Timestamp: time.Now(),
	})

	if err := s.jobManager.EnqueueReport(ctx, report.ID); err != nil {
		s.logger.Error("Failed to enqueue report job", zap.Error(err))
		s.repo.UpdateReportStatus(ctx, report.ID, ReportStatusFailed, 0, "", "Queue failure")
		return nil, err
	}

	s.repo.UpdateReportStatus(ctx, report.ID, ReportStatusQueued, 0, "", "")

	// Cache result
	if s.cache != nil {
		cacheKey := s.cache.CacheKey(tenantID.String(), string(req.ReportType), req.Format, params)
		s.cache.Set(ctx, cacheKey, report.ID.String())
	}

	return &CreateReportResponse{
		ReportID: report.ID,
		Status:   ReportStatusQueued,
	}, nil
}

func (s *service) GetReportStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*ReportStatusResponse, error) {
	report, err := s.repo.GetReport(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, errors.New("report not found")
	}
	return &ReportStatusResponse{
		Status:   report.Status,
		Progress: report.Progress,
	}, nil
}

func (s *service) GetReport(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Report, error) {
	report, err := s.repo.GetReport(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, errors.New("report not found")
	}
	return report, nil
}

func (s *service) ListReports(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]Report, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListReports(ctx, tenantID, limit, offset)
}

func (s *service) CreateTemplate(ctx context.Context, tmpl *ReportTemplate) error {
	if tmpl.ID == uuid.Nil {
		tmpl.ID = uuid.New()
	}
	if tmpl.Version == 0 {
		tmpl.Version = 1
	}
	if err := s.repo.CreateReportTemplate(ctx, tmpl); err != nil {
		return err
	}
	// Bump global cache version so previously cached reports using this template
	// are no longer served. See cache.go reportCacheVersion const.
	s.logger.Info("Report template created — cache version should be bumped for stale deduplication",
		zap.String("template_id", tmpl.ID.String()),
		zap.Int("version", tmpl.Version))
	return nil
}

func (s *service) ListTemplates(ctx context.Context) ([]ReportTemplate, error) {
	return s.repo.ListReportTemplates(ctx)
}
