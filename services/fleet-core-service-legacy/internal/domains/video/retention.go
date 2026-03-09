package video

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// RetentionEnforcer deletes expired videos and their stored files.
// It respects evidence locks and legal holds.
type RetentionEnforcer struct {
	repo    Repository
	storage VideoStorage
	logger  *zap.Logger
}

func NewRetentionEnforcer(repo Repository, storage VideoStorage, logger *zap.Logger) *RetentionEnforcer {
	return &RetentionEnforcer{repo: repo, storage: storage, logger: logger}
}

// Run starts the daily retention sweep.
func (e *RetentionEnforcer) Run(ctx context.Context) {
	e.logger.Info("Video retention enforcer started")
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	e.sweep(ctx) // immediate first run
	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Video retention enforcer stopped")
			return
		case <-ticker.C:
			e.sweep(ctx)
		}
	}
}

func (e *RetentionEnforcer) sweep(ctx context.Context) {
	e.logger.Info("Video retention sweep started")

	// Policy: continuous videos kept for 30 days
	e.deleteExpired(ctx, VideoTypeContinuous, RetentionDays(VideoTypeContinuous))
	// Policy: event videos kept for 365 days
	e.deleteExpired(ctx, VideoTypeEvent, RetentionDays(VideoTypeEvent))
	// Policy: manual clips kept for 90 days
	e.deleteExpired(ctx, VideoTypeManual, RetentionDays(VideoTypeManual))
	// Policy: live capture kept for 7 days
	e.deleteExpired(ctx, VideoTypeLiveCapture, RetentionDays(VideoTypeLiveCapture))
	// Legal evidence (evidence_locked=TRUE or legal_hold=TRUE) is NEVER deleted by this enforcer.

	e.logger.Info("Video retention sweep complete")
}

func (e *RetentionEnforcer) deleteExpired(ctx context.Context, vt VideoType, days int) {
	expired, err := e.repo.ListExpiredVideos(ctx, vt, days)
	if err != nil {
		e.logger.Error("Failed to list expired videos",
			zap.String("type", string(vt)), zap.Error(err))
		return
	}

	for _, v := range expired {
		// Delete primary file from storage
		if v.FilePath != "" {
			if err := e.storage.Delete(ctx, v.FilePath); err != nil {
				e.logger.Warn("Failed to delete video file",
					zap.String("key", v.FilePath), zap.Error(err))
			}
		}
		// Delete HLS segments if transcoded
		if v.HLSPlaylistPath != "" {
			if err := e.storage.Delete(ctx, v.HLSPlaylistPath); err != nil {
				e.logger.Warn("Failed to delete HLS files",
					zap.String("key", v.HLSPlaylistPath), zap.Error(err))
			}
		}
		// Soft-delete the DB record (sets status=deleted)
		if err := e.repo.DeleteVideo(ctx, v.ID, v.TenantID); err != nil {
			e.logger.Warn("Failed to delete video DB record",
				zap.String("video_id", v.ID.String()), zap.Error(err))
		} else {
			e.logger.Info("Expired video deleted",
				zap.String("video_id", v.ID.String()),
				zap.String("type", string(vt)),
				zap.Int("retention_days", days))
		}
	}
}
