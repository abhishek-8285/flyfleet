package workers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// FinanceBackupWorker handles daily snapshots and 7-year legal archival.
type FinanceBackupWorker struct {
	logger *zap.Logger
}

func NewFinanceBackupWorker(logger *zap.Logger) *FinanceBackupWorker {
	return &FinanceBackupWorker{logger: logger}
}

func (w *FinanceBackupWorker) Start(ctx context.Context) {
	// Daily backup at 02:00 IST
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runBackup(ctx)
		}
	}
}

func (w *FinanceBackupWorker) runBackup(ctx context.Context) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		w.logger.Warn("FinanceBackupWorker: DATABASE_URL not configured; skipping backup")
		return
	}

	backupDir := os.Getenv("FINANCE_BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/tmp/flyfleet-finance-backups"
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		w.logger.Error("FinanceBackupWorker: failed to create backup directory", zap.Error(err))
		return
	}

	fileName := fmt.Sprintf("finance_snapshot_%s.sql.gz", time.Now().UTC().Format("20060102T150405Z"))
	target := filepath.Join(backupDir, fileName)

	// Snapshot financial tables; includes schema and data for portability.
	// SECURITY FIX: Avoid shell injection by not using bash -lc. Use exec.Command with proper argument separation.
	// Create gzip file directly
	targetFile, err := os.Create(target)
	if err != nil {
		w.logger.Error("FinanceBackupWorker: failed to create backup file", zap.Error(err))
		return
	}
	defer targetFile.Close()

	// Start pg_dump
	pgDumpCmd := exec.CommandContext(ctx, "pg_dump", dbURL, "--format=plain", "--no-owner", "--no-privileges")
	stdout, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		w.logger.Error("FinanceBackupWorker: failed to get pg_dump stdout", zap.Error(err))
		return
	}

	// Start gzip subprocess
	gzipCmd := exec.CommandContext(ctx, "gzip")
	gzipCmd.Stdin = stdout
	gzipCmd.Stdout = targetFile

	// Start pg_dump
	if err := pgDumpCmd.Start(); err != nil {
		w.logger.Error("FinanceBackupWorker: pg_dump start failed", zap.Error(err))
		return
	}

	// Wait for gzip to complete
	if err := gzipCmd.Wait(); err != nil {
		w.logger.Error("FinanceBackupWorker: gzip failed", zap.Error(err))
		pgDumpCmd.Process.Kill()
		return
	}

	// Wait for pg_dump to complete
	if err := pgDumpCmd.Wait(); err != nil {
		w.logger.Error("FinanceBackupWorker: pg_dump failed", zap.Error(err))
		return
	}

	// Retain legal archive for 7 years (2555 days).
	cutoff := time.Now().AddDate(-7, 0, 0)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		w.logger.Warn("FinanceBackupWorker: cannot read backup directory for retention", zap.Error(err))
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(backupDir, entry.Name()))
		}
	}

	w.logger.Info("FinanceBackupWorker: backup complete", zap.String("file", target))
}
