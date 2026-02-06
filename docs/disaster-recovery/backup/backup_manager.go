package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fleet-platform/shared/pkg/logger"
)

type BackupManager struct {
	logger      logger.Logger
	config      BackupConfig
	storage     BackupStorage
	notification NotificationService
}

type BackupConfig struct {
	BackupDir        string
	RetentionDays    int
	FullBackupCron   string
	IncrementalCron  string
	EncryptionKey    string
	CompressionLevel int
}

type BackupStorage interface {
	Upload(ctx context.Context, localPath, remotePath string) error
	Download(ctx context.Context, remotePath, localPath string) error
	List(prefix string) ([]BackupMetadata, error)
	Delete(remotePath string) error
}

type NotificationService interface {
	SendBackupNotification(notification BackupNotification) error
}

type BackupMetadata struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // full, incremental, differential
	Status      string    `json:"status"` // created, uploading, completed, failed
	Size        int64     `json:"size"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    string    `json:"duration"`
	Checksum    string    `json:"checksum"`
	Encrypted   bool      `json:"encrypted"`
	Compressed  bool      `json:"compressed"`
}

type BackupNotification struct {
	BackupID    string    `json:"backup_id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Recipients  []string  `json:"recipients"`
}

const (
	BackupTypeFull        = "full"
	BackupTypeIncremental = "incremental"
	BackupTypeDifferential = "differential"
	
	BackupStatusCreated   = "created"
	BackupStatusUploading = "uploading"
	BackupStatusCompleted = "completed"
	BackupStatusFailed    = "failed"
)

func NewBackupManager(logger logger.Logger, config BackupConfig, storage BackupStorage, notification NotificationService) *BackupManager {
	return &BackupManager{
		logger:      logger,
		config:      config,
		storage:     storage,
		notification: notification,
	}
}

func (bm *BackupManager) CreateFullBackup(ctx context.Context) (*BackupMetadata, error) {
	backupID := uuid.New().String()
	startTime := time.Now()
	
	bm.logger.Info("Starting full backup", zap.String("backup_id", backupID))
	
	// Create backup metadata
	backup := &BackupMetadata{
		ID:        backupID,
		Type:      BackupTypeFull,
		Status:    BackupStatusCreated,
		StartTime: startTime,
		Encrypted: bm.config.EncryptionKey != "",
		Compressed: true,
	}
	
	// Create backup directory
	backupDir := filepath.Join(bm.config.BackupDir, backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Perform database backup
	dbBackupPath := filepath.Join(backupDir, "database.sql")
	if err := bm.backupDatabase(ctx, dbBackupPath); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Backup application data
	appDataPath := filepath.Join(backupDir, "application-data")
	if err := bm.backupApplicationData(ctx, appDataPath); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Compress backup
	compressedPath := filepath.Join(bm.config.BackupDir, backupID+".tar.gz")
	if err := bm.compressBackup(backupDir, compressedPath); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Encrypt if configured
	if bm.config.EncryptionKey != "" {
		encryptedPath := compressedPath + ".enc"
		if err := bm.encryptBackup(compressedPath, encryptedPath); err != nil {
			bm.markBackupFailed(backup, err)
			return backup, err
		}
		// Remove unencrypted file
		os.Remove(compressedPath)
		compressedPath = encryptedPath
	}
	
	// Get file size
	if fileInfo, err := os.Stat(compressedPath); err == nil {
		backup.Size = fileInfo.Size()
	}
	
	// Upload to remote storage
	backup.Status = BackupStatusUploading
	remotePath := fmt.Sprintf("backups/%s/%s", time.Now().Format("2006/01/02"), filepath.Base(compressedPath))
	if err := bm.storage.Upload(ctx, compressedPath, remotePath); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Cleanup local files
	os.RemoveAll(backupDir)
	os.Remove(compressedPath)
	
	// Complete backup
	endTime := time.Now()
	backup.Status = BackupStatusCompleted
	backup.EndTime = endTime
	backup.Duration = endTime.Sub(startTime).String()
	
	bm.logger.Info("Full backup completed",
		zap.String("backup_id", backupID),
		zap.String("duration", backup.Duration),
		zap.Int64("size", backup.Size))
	
	// Send notification
	notification := BackupNotification{
		BackupID:  backupID,
		Type:      BackupTypeFull,
		Status:    BackupStatusCompleted,
		Message:   fmt.Sprintf("Full backup completed successfully in %s", backup.Duration),
		Timestamp: time.Now(),
		Recipients: []string{"admin@flyfleet.com", "ops@flyfleet.com"},
	}
	
	if err := bm.notification.SendBackupNotification(notification); err != nil {
		bm.logger.Error("Failed to send backup notification", zap.Error(err))
	}
	
	return backup, nil
}

func (bm *BackupManager) CreateIncrementalBackup(ctx context.Context, baseBackupID string) (*BackupMetadata, error) {
	backupID := uuid.New().String()
	startTime := time.Now()
	
	bm.logger.Info("Starting incremental backup",
		zap.String("backup_id", backupID),
		zap.String("base_backup_id", baseBackupID))
	
	backup := &BackupMetadata{
		ID:        backupID,
		Type:      BackupTypeIncremental,
		Status:    BackupStatusCreated,
		StartTime: startTime,
		Encrypted: bm.config.EncryptionKey != "",
		Compressed: true,
	}
	
	// Create backup directory
	backupDir := filepath.Join(bm.config.BackupDir, backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Get changes since last backup
	changes, err := bm.getChangesSinceBackup(ctx, baseBackupID)
	if err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Backup only changed data
	if err := bm.backupChanges(ctx, backupDir, changes); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Compress and encrypt (similar to full backup)
	compressedPath := filepath.Join(bm.config.BackupDir, backupID+".tar.gz")
	if err := bm.compressBackup(backupDir, compressedPath); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	if bm.config.EncryptionKey != "" {
		encryptedPath := compressedPath + ".enc"
		if err := bm.encryptBackup(compressedPath, encryptedPath); err != nil {
			bm.markBackupFailed(backup, err)
			return backup, err
		}
		os.Remove(compressedPath)
		compressedPath = encryptedPath
	}
	
	// Upload to remote storage
	backup.Status = BackupStatusUploading
	remotePath := fmt.Sprintf("backups/%s/%s", time.Now().Format("2006/01/02"), filepath.Base(compressedPath))
	if err := bm.storage.Upload(ctx, compressedPath, remotePath); err != nil {
		bm.markBackupFailed(backup, err)
		return backup, err
	}
	
	// Cleanup
	os.RemoveAll(backupDir)
	os.Remove(compressedPath)
	
	// Complete backup
	endTime := time.Now()
	backup.Status = BackupStatusCompleted
	backup.EndTime = endTime
	backup.Duration = endTime.Sub(startTime).String()
	
	if fileInfo, err := os.Stat(compressedPath); err == nil {
		backup.Size = fileInfo.Size()
	}
	
	bm.logger.Info("Incremental backup completed",
		zap.String("backup_id", backupID),
		zap.String("duration", backup.Duration))
	
	return backup, nil
}

func (bm *BackupManager) RestoreBackup(ctx context.Context, backupID, restorePath string) error {
	bm.logger.Info("Starting backup restoration",
		zap.String("backup_id", backupID),
		zap.String("restore_path", restorePath))
	
	// Download backup from storage
	remotePath := fmt.Sprintf("backups/*/%s*", backupID)
	localPath := filepath.Join(bm.config.BackupDir, backupID+".downloaded")
	
	if err := bm.storage.Download(ctx, remotePath, localPath); err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}
	
	// Decrypt if necessary
	if bm.config.EncryptionKey != "" {
		decryptedPath := localPath + ".dec"
		if err := bm.decryptBackup(localPath, decryptedPath); err != nil {
			return fmt.Errorf("failed to decrypt backup: %w", err)
		}
		os.Remove(localPath)
		localPath = decryptedPath
	}
	
	// Decompress
	decompressedPath := filepath.Join(bm.config.BackupDir, backupID+"_extracted")
	if err := bm.decompressBackup(localPath, decompressedPath); err != nil {
		return fmt.Errorf("failed to decompress backup: %w", err)
	}
	
	// Restore database
	dbBackupPath := filepath.Join(decompressedPath, "database.sql")
	if err := bm.restoreDatabase(ctx, dbBackupPath); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}
	
	// Restore application data
	appDataPath := filepath.Join(decompressedPath, "application-data")
	if err := bm.restoreApplicationData(ctx, appDataPath, restorePath); err != nil {
		return fmt.Errorf("failed to restore application data: %w", err)
	}
	
	// Cleanup
	os.Remove(localPath)
	os.RemoveAll(decompressedPath)
	
	bm.logger.Info("Backup restoration completed", zap.String("backup_id", backupID))
	return nil
}

func (bm *BackupManager) ListBackups(ctx context.Context) ([]BackupMetadata, error) {
	prefix := "backups/"
	metadataList, err := bm.storage.List(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	
	return metadataList, nil
}

func (bm *BackupManager) DeleteOldBackups(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -bm.config.RetentionDays)
	prefix := "backups/"
	
	backups, err := bm.storage.List(prefix)
	if err != nil {
		return fmt.Errorf("failed to list backups for cleanup: %w", err)
	}
	
	for _, backup := range backups {
		if backup.EndTime.Before(cutoffDate) {
			remotePath := fmt.Sprintf("backups/%s/%s", 
				backup.EndTime.Format("2006/01/02"), backup.ID)
			
			if err := bm.storage.Delete(remotePath); err != nil {
				bm.logger.Error("Failed to delete old backup",
					zap.String("backup_id", backup.ID),
					zap.Error(err))
				continue
			}
			
			bm.logger.Info("Deleted old backup", zap.String("backup_id", backup.ID))
		}
	}
	
	return nil
}

func (bm *BackupManager) backupDatabase(ctx context.Context, outputPath string) error {
	cmd := exec.Command("pg_dump", 
		"--dbname=postgresql://username:password@localhost:5432/flyfleet",
		"--format=custom",
		"--file="+outputPath)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("database backup failed: %w", err)
	}
	
	return nil
}

func (bm *BackupManager) backupApplicationData(ctx context.Context, outputPath string) error {
	// Backup configuration files, logs, etc.
	// This is implementation-specific
	return nil
}

func (bm *BackupManager) compressBackup(sourceDir, outputPath string) error {
	cmd := exec.Command("tar", "-czf", outputPath, "-C", filepath.Dir(sourceDir), filepath.Base(sourceDir))
	return cmd.Run()
}

func (bm *BackupManager) encryptBackup(inputPath, outputPath string) error {
	// In real implementation, use proper encryption library
	cmd := exec.Command("openssl", "enc", "-aes-256-cbc", "-salt",
		"-in", inputPath, "-out", outputPath,
		"-k", bm.config.EncryptionKey)
	return cmd.Run()
}

func (bm *BackupManager) decryptBackup(inputPath, outputPath string) error {
	cmd := exec.Command("openssl", "enc", "-d", "-aes-256-cbc", "-salt",
		"-in", inputPath, "-out", outputPath,
		"-k", bm.config.EncryptionKey)
	return cmd.Run()
}

func (bm *BackupManager) decompressBackup(inputPath, outputPath string) error {
	cmd := exec.Command("tar", "-xzf", inputPath, "-C", outputPath)
	return cmd.Run()
}

func (bm *BackupManager) restoreDatabase(ctx context.Context, backupPath string) error {
	cmd := exec.Command("pg_restore",
		"--dbname=postgresql://username:password@localhost:5432/flyfleet",
		"--clean",
		"--if-exists",
		backupPath)
	
	return cmd.Run()
}

func (bm *BackupManager) restoreApplicationData(ctx context.Context, backupPath, restorePath string) error {
	// Restore application data to specified path
	return nil
}

func (bm *BackupManager) getChangesSinceBackup(ctx context.Context, baseBackupID string) ([]string, error) {
	// Implementation depends on change tracking mechanism
	// Could use database transaction logs, file system change logs, etc.
	return []string{}, nil
}

func (bm *BackupManager) backupChanges(ctx context.Context, backupDir string, changes []string) error {
	// Backup only the changed files/data
	return nil
}

func (bm *BackupManager) markBackupFailed(backup *BackupMetadata, err error) {
	backup.Status = BackupStatusFailed
	backup.EndTime = time.Now()
	backup.Duration = backup.EndTime.Sub(backup.StartTime).String()
	
	bm.logger.Error("Backup failed",
		zap.String("backup_id", backup.ID),
		zap.Error(err))
}