package reports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler fires scheduled reports using cron expressions stored in the DB.
type Scheduler struct {
	repo       Repository
	jobManager JobManager
	logger     *zap.Logger
	cron       *cron.Cron
}

func NewScheduler(repo Repository, jobManager JobManager, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		repo:       repo,
		jobManager: jobManager,
		logger:     logger,
		cron:       cron.New(cron.WithSeconds()),
	}
}

// Start loads all enabled scheduled reports from DB and registers them,
// but only runs on the instance that holds the advisory leader lock.
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("Report Scheduler trying to acquire leader lock")

	acquired, err := s.repo.AcquireSchedulerLock(ctx)
	if err != nil || !acquired {
		s.logger.Info("Scheduler lock not acquired — another replica is leader, skipping")
		return
	}
	s.logger.Info("Scheduler leader lock acquired")
	defer s.repo.ReleaseSchedulerLock(ctx)

	schedules, err := s.repo.ListEnabledScheduledReports(ctx)
	if err != nil {
		s.logger.Error("Failed to load scheduled reports", zap.Error(err))
		return
	}

	for _, sched := range schedules {
		sc := sched // capture loop var
		_, err := s.cron.AddFunc(sc.CronExpression, func() {
			s.fire(ctx, sc)
		})
		if err != nil {
			s.logger.Error("Invalid cron expression",
				zap.String("id", sc.ID.String()),
				zap.String("cron", sc.CronExpression),
				zap.Error(err))
		}
	}

	s.cron.Start()
	s.logger.Info("Scheduler started", zap.Int("schedules", len(schedules)))

	go func() {
		<-ctx.Done()
		s.cron.Stop()
		s.logger.Info("Scheduler stopped")
	}()
}

func (s *Scheduler) fire(ctx context.Context, sched ScheduledReport) {
	s.logger.Info("Firing scheduled report",
		zap.String("scheduled_id", sched.ID.String()),
		zap.String("tenant_id", sched.TenantID.String()))

	// Determine report type from the template
	tmpl, err := s.repo.GetReportTemplate(ctx, sched.ReportTemplateID)
	if err != nil || tmpl == nil {
		s.logger.Error("Scheduled report template not found", zap.String("id", sched.ID.String()))
		return
	}

	report := &Report{
		ID:          uuid.New(),
		TenantID:    sched.TenantID,
		Type:        ReportType(tmpl.OutputFormat),
		Status:      ReportStatusCreated,
		Parameters:  sched.Parameters,
		RequestedBy: sched.TenantID, // system-generated
		RequestedAt: time.Now(),
	}

	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.Error("Failed to create scheduled report record", zap.Error(err))
		return
	}

	if err := s.jobManager.EnqueueReport(ctx, report.ID); err != nil {
		s.logger.Error("Failed to enqueue scheduled report", zap.Error(err))
		s.repo.UpdateReportStatus(ctx, report.ID, ReportStatusFailed, 0, "", "Failed to enqueue")
		return
	}
	s.repo.UpdateReportStatus(ctx, report.ID, ReportStatusQueued, 0, "", "")
	s.logger.Info("Scheduled report enqueued", zap.String("report_id", report.ID.String()))
}
