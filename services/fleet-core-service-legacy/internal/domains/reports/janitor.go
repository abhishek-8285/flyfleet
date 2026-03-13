package reports

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const (
	retentionDays     = 30
	archiveDays       = 90
	dlqAlertThreshold = 10 // Log a warning if DLQ has more than this many entries
)


// Janitor deletes/archives old reports and their stored files on a regular schedule.
type Janitor struct {
	repo    Repository
	storage ObjectStorage
	logger  *zap.Logger
}

func NewJanitor(repo Repository, storage ObjectStorage, logger *zap.Logger) *Janitor {
	return &Janitor{repo: repo, storage: storage, logger: logger}
}

// Run starts the janitor loop: daily storage cleanup + every 5m DLQ monitoring.
func (j *Janitor) Run(ctx context.Context) {
	j.logger.Info("Janitor started", zap.Int("retention_days", retentionDays))

	// File + DB cleanup: runs daily
	cleanupTicker := time.NewTicker(24 * time.Hour)
	// DLQ monitoring: runs every 5 minutes for rapid failure alerting
	dlqTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()
	defer dlqTicker.Stop()

	j.sweep(ctx)      // immediate first run
	j.checkDLQ(ctx)   // immediate DLQ check

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("Janitor stopped")
			return
		case <-cleanupTicker.C:
			j.sweep(ctx)
		case <-dlqTicker.C:
			j.checkDLQ(ctx)
		}
	}
}

func (j *Janitor) sweep(ctx context.Context) {
	// 1. Delete orphaned storage files for expired reports first
	expired, err := j.repo.ListExpiredReports(ctx, retentionDays)
	if err != nil {
		j.logger.Error("Janitor: failed to list expired reports", zap.Error(err))
	} else {
		for _, report := range expired {
			if report.FileURL != "" {
				// Pattern: tenant-id/report-id/export.csv.gz
				key := report.TenantID.String() + "/" + report.ID.String() + "/export.csv.gz"
				if delErr := j.storage.Delete(ctx, key); delErr != nil {
					j.logger.Warn("Janitor: failed to delete storage file",
						zap.String("report_id", report.ID.String()),
						zap.Error(delErr))
				}
			}
		}
	}

	// 2. Delete the DB rows
	deleted, err := j.repo.DeleteExpiredReports(ctx, retentionDays)
	if err != nil {
		j.logger.Error("Janitor cleanup failed", zap.Error(err))
	} else {
		j.logger.Info("Janitor cleanup complete", zap.Int64("deleted_db_records", deleted))
	}
}

// checkDLQ runs every 5 minutes to detect permanently failed jobs quickly.
func (j *Janitor) checkDLQ(ctx context.Context) {
	dlqSize, err := j.repo.CountDLQEntries(ctx)
	if err != nil {
		j.logger.Warn("Janitor: failed to check DLQ size", zap.Error(err))
		return
	}
	if dlqSize >= dlqAlertThreshold {
		j.logger.Warn("ALERT: Report DLQ has entries — investigate permanently failed jobs",
			zap.Int64("dlq_size", dlqSize),
			zap.Int("threshold", dlqAlertThreshold))
	} else if dlqSize > 0 {
		j.logger.Info("DLQ has entries (below alert threshold)",
			zap.Int64("dlq_size", dlqSize))
	}
}
