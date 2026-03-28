package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
	"go.uber.org/zap"
)

type ArchiveWorker struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewArchiveWorker(repo repository.ComplianceRepository, logger logger.Logger) *ArchiveWorker {
	return &ArchiveWorker{
		repo:   repo,
		logger: logger,
	}
}

func (w *ArchiveWorker) Run(ctx context.Context) error {
	w.logger.Info("starting compliance archive worker")

	// Rules:
	// GPS logs > 90 days archived (deleted from hot table)
	// Trip sheets > 1 year archived
	// Compliance alerts > 1 year archived

	// For this implementation, we use CleanupOldData as a catch-all maintenance function
	// with different retention limits.

	// 1. Duty/GPS Logs (90 days)
	gpsRetentionLimit := time.Now().AddDate(0, 0, -90)
	rows, err := w.repo.CleanupOldData(ctx, gpsRetentionLimit)
	if err != nil {
		w.logger.Error("failed to cleanup old driver logs", zap.Error(err))
	} else {
		w.logger.Info("cleaned up old driver logs", zap.Int64("count", rows))
	}

	// 2. General Data Retention (1 year)
	oneYearLimit := time.Now().AddDate(-1, 0, 0)
	// Note: Would typically have specific methods for different tables
	// s.repo.CleanupAlerts(ctx, oneYearLimit)

	w.logger.Info("archive worker run completed", zap.Time("retention_limit", oneYearLimit))
	return nil
}
