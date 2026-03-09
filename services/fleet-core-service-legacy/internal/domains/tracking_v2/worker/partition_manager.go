package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	PartitionLockKey = 1002 // Postgres advisory lock key for partitions
)

// PartitionManager ensures PostgreSQL partitions for telemetry_points exist
// for the current and upcoming months.
// Multi-pod safe: Uses PostgreSQL Advisory Locks to ensure only one leader handles DDL.
type PartitionManager struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	ticker *time.Ticker
}

func NewPartitionManager(db *pgxpool.Pool, logger *zap.Logger) *PartitionManager {
	return &PartitionManager{
		db:     db,
		logger: logger,
		ticker: time.NewTicker(24 * time.Hour), // Check once a day
	}
}

func (w *PartitionManager) Start(ctx context.Context) {
	w.logger.Info("Starting Telemetry Partition Manager")
	defer w.ticker.Stop()

	// Initial run
	w.ensurePartitions(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping Telemetry Partition Manager (Graceful)")
			return
		case <-w.ticker.C:
			w.ensurePartitions(ctx)
		}
	}
}

func (w *PartitionManager) ensurePartitions(ctx context.Context) {
	// Leader Election: Try to acquire advisory lock
	var locked bool
	err := w.db.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", PartitionLockKey).Scan(&locked)
	if err != nil {
		w.logger.Error("failed to check partition lock", zap.Error(err))
		return
	}
	if !locked {
		w.logger.Debug("another instance is handling partitions; skipping")
		return
	}
	defer w.db.Exec(ctx, "SELECT pg_advisory_unlock($1)", PartitionLockKey)

	w.logger.Info("Running Partition Maintenance (Leader)")
	now := time.Now().UTC()
	// Ensure current month and next 2 months
	for i := 0; i < 3; i++ {
		target := now.AddDate(0, i, 0)
		if err := w.createPartitionIfMissing(ctx, target); err != nil {
			w.logger.Error("failed to create partition", zap.Time("target", target), zap.Error(err))
		}
	}
}

func (w *PartitionManager) createPartitionIfMissing(ctx context.Context, t time.Time) error {
	year := t.Year()
	month := int(t.Month())
	tableName := fmt.Sprintf("telemetry_points_y%04dm%02d", year, month)
	
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	w.logger.Debug("checking partition", zap.String("table", tableName))

	// 1. Create partition table
	q := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		PARTITION OF telemetry_points
		FOR VALUES FROM ('%s') TO ('%s')`,
		tableName,
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
	)

	if _, err := w.db.Exec(ctx, q); err != nil {
		return fmt.Errorf("create partition %s failed: %w", tableName, err)
	}

	// 2. Ensure indexes on partition (PostGIS GIST and BRIN)
	idxGist := fmt.Sprintf("idx_%s_geo", tableName)
	qGist := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING GIST (geo)`, idxGist, tableName)
	
	idxBrin := fmt.Sprintf("idx_%s_recorded_at_brin", tableName)
	qBrin := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING BRIN (recorded_at)`, idxBrin, tableName)

	if _, err := w.db.Exec(ctx, qGist); err != nil {
		w.logger.Warn("failed to create GIST index", zap.String("table", tableName), zap.Error(err))
	}
	if _, err := w.db.Exec(ctx, qBrin); err != nil {
		w.logger.Warn("failed to create BRIN index", zap.String("table", tableName), zap.Error(err))
	}

	return nil
}
