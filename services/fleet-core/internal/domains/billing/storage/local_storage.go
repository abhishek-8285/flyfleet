package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"go.uber.org/zap"
)

// LocalStorage handles local file storage for invoices, implementing ObjectStorage interface
type LocalStorage struct {
	basePath string
	logger   *zap.Logger
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(basePath string, logger *zap.Logger) (*LocalStorage, error) {
	if basePath == "" {
		basePath = "./data/invoices"
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// GetPath returns the file path for an invoice
func (s *LocalStorage) GetPath(invoice *models.Invoice) string {
	// Organize by year/month for better management
	year := invoice.InvoiceDate.Format("2006")
	month := invoice.InvoiceDate.Format("01")

	return filepath.Join(
		s.basePath,
		year,
		month,
		fmt.Sprintf("%s.pdf", invoice.InvoiceNumber),
	)
}

// Save saves a PDF to local storage
func (s *LocalStorage) Save(ctx context.Context, invoice *models.Invoice, pdfData []byte) error {
	path := s.GetPath(invoice)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, pdfData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	s.logger.Info("Invoice PDF saved",
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("path", path),
		zap.Int("size_bytes", len(pdfData)),
	)

	return nil
}

// Get retrieves a PDF from local storage
func (s *LocalStorage) Get(ctx context.Context, invoice *models.Invoice) ([]byte, error) {
	path := s.GetPath(invoice)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// GetPresignedURL returns a local file path (mocking S3 presigned URL)
func (s *LocalStorage) GetPresignedURL(ctx context.Context, invoice *models.Invoice, expires time.Duration) (string, error) {
	return "file://" + s.GetPath(invoice), nil
}

// Exists checks if a PDF exists in storage
func (s *LocalStorage) Exists(ctx context.Context, invoice *models.Invoice) (bool, error) {
	path := s.GetPath(invoice)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Delete removes a PDF from storage (soft delete - renames)
func (s *LocalStorage) Delete(ctx context.Context, invoice *models.Invoice) error {
	path := s.GetPath(invoice)

	// Rename instead of delete for audit trail
	deletedPath := path + ".deleted"
	if err := os.Rename(path, deletedPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Info("Invoice PDF deleted",
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("path", path),
	)

	return nil
}

// List returns all invoices in a date range
func (s *LocalStorage) List(ctx context.Context, startDate, endDate time.Time) ([]string, error) {
	var files []string

	// Walk through directory
	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and deleted files
		if info.IsDir() || filepath.Ext(path) != ".pdf" {
			return nil
		}

		// Check if file is in date range
		modTime := info.ModTime()
		if modTime.Before(startDate) || modTime.After(endDate) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// GetStorageStats returns storage statistics
func (s *LocalStorage) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	var totalSize int64
	var fileCount int64

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".pdf" {
			totalSize += info.Size()
			fileCount++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	return &StorageStats{
		TotalSizeBytes: totalSize,
		TotalSizeMB:    float64(totalSize) / 1024 / 1024,
		FileCount:      fileCount,
		BasePath:       s.basePath,
	}, nil
}

// CleanupOldFiles removes files older than specified duration
func (s *LocalStorage) CleanupOldFiles(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	deletedCount := 0

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process PDF files
		if filepath.Ext(path) != ".pdf" {
			return nil
		}

		// Check if file is older than cutoff
		if info.ModTime().Before(cutoff) {
			// Move to archive instead of deleting
			archivePath := path + ".archived"
			if err := os.Rename(path, archivePath); err != nil {
				s.logger.Error("Failed to archive file",
					zap.String("path", path),
					zap.Error(err),
				)
				return nil // Continue with other files
			}

			deletedCount++
			s.logger.Info("Archived old invoice",
				zap.String("path", path),
				zap.Time("modified", info.ModTime()),
			)
		}

		return nil
	})

	if err != nil {
		return deletedCount, fmt.Errorf("failed to cleanup files: %w", err)
	}

	return deletedCount, nil
}

// HealthCheck checks if storage is healthy
func (s *LocalStorage) HealthCheck(ctx context.Context) error {
	// Check if base path exists and is writable
	testFile := filepath.Join(s.basePath, ".healthcheck")

	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}

	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to cleanup healthcheck file: %w", err)
	}

	return nil
}

// VerifyChecksum verifies the integrity of the stored file against an expected SHA256 checksum
func (s *LocalStorage) VerifyChecksum(ctx context.Context, invoice *models.Invoice, expectedChecksum string) (bool, error) {
	path := s.GetPath(invoice)

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}

	actualChecksum := hex.EncodeToString(h.Sum(nil))
	return actualChecksum == expectedChecksum, nil
}
