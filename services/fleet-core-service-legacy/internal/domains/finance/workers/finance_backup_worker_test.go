package workers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestFinanceBackupWorker_RunBackupCreatesSnapshotAndCleansRetention(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}

	// Stub pg_dump in PATH; worker pipes it to gzip.
	pgDumpPath := filepath.Join(binDir, "pg_dump")
	if err := os.WriteFile(pgDumpPath, []byte("#!/usr/bin/env sh\necho 'mock-backup'\n"), 0o755); err != nil {
		t.Fatalf("write pg_dump stub: %v", err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	oldFile := filepath.Join(backupDir, "old.sql.gz")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("mkdir backups: %v", err)
	}
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old file: %v", err)
	}
	oldTs := time.Now().AddDate(-8, 0, 0)
	if err := os.Chtimes(oldFile, oldTs, oldTs); err != nil {
		t.Fatalf("set old file mtime: %v", err)
	}

	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("FINANCE_BACKUP_DIR", backupDir)
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	w := NewFinanceBackupWorker(zap.NewNop())
	w.runBackup(context.Background())

	matches, err := filepath.Glob(filepath.Join(backupDir, "finance_snapshot_*.sql.gz"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected snapshot file to be created")
	}
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("expected old retention file to be removed, stat err=%v", err)
	}
}
