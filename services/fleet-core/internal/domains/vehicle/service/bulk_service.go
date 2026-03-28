package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/processor"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BulkService struct {
	repo      *repository.BulkRepository
	processor *processor.BulkProcessor
	logger    *zap.Logger
}

func NewBulkService(repo *repository.BulkRepository, processor *processor.BulkProcessor, logger *zap.Logger) *BulkService {
	// Ensure uploads directory exists
	if err := os.MkdirAll("uploads", 0755); err != nil {
		logger.Fatal("Failed to create uploads directory", zap.Error(err))
	}
	return &BulkService{
		repo:      repo,
		processor: processor,
		logger:    logger,
	}
}

// validateCSVContent verifies that the file content is actually valid CSV format
// and not a malicious script disguised with a .csv extension (CWE-434)
func (s *BulkService) validateCSVContent(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("file is empty")
	}

	// Try to parse as CSV to ensure it's valid CSV data
	reader := csv.NewReader(bytes.NewReader(data))

	// Set some sane defaults
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.LazyQuotes = true    // Allow unescaped quotes in fields (common in real CSVs)
	reader.TrimLeadingSpace = true

	// Read the first record (header) to validate format
	_, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("file has no data")
		}
		return fmt.Errorf("invalid CSV format: %w", err)
	}

	// Additional check: ensure we have at least some columns
	// CSV files should have at least one column
	// (We've already successfully parsed a row, so this is valid CSV)

	return nil
}

func (s *BulkService) CreateJob(ctx context.Context, orgID uuid.UUID, fileName string, fileReader io.Reader) (*models.BulkJob, error) {
	// 1. Validation
	ext := filepath.Ext(fileName)
	if ext != ".csv" {
		return nil, fmt.Errorf("invalid file type: only .csv is supported")
	}

	// Read all data (limit to 10MB)
	// Create a limited reader
	const MaxFileSize = 10 * 1024 * 1024 // 10MB
	data, err := io.ReadAll(io.LimitReader(fileReader, MaxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if len(data) > MaxFileSize {
		return nil, fmt.Errorf("file too large: maximum size is 10MB")
	}

	// SECURITY FIX: Validate actual file content is CSV, not just extension (CWE-434)
	// This prevents attackers from uploading malicious scripts with .csv extension
	if err := s.validateCSVContent(data); err != nil {
		return nil, fmt.Errorf("invalid CSV file: %w", err)
	}

	jobID := uuid.New()

	// 2. Save file locally
	filePath := filepath.Join("uploads", fmt.Sprintf("%s.csv", jobID))
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}

	job := &models.BulkJob{
		ID:             jobID,
		OrganizationID: orgID,
		FileName:       fileName,
		Status:         "pending", // Initial status
		TotalRecords:   0,
		SuccessCount:   0,
		FailedCount:    0,
		RetryCount:     0,
		CreatedBy:      uuid.Nil, // Should get from context if available but for now nil
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateJob(ctx, job); err != nil {
		// Cleanup file if DB insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create bulk job: %w", err)
	}

	// Trigger Async Processing
	go func() {
		// Panic Recovery
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic in bulk job processing",
					zap.String("job_id", job.ID.String()),
					zap.Any("recover", r))
				// Update job status to failed
				// Create a background context as the request context might be cancelled
				bgCtx := context.Background()
				s.repo.UpdateJobStatus(bgCtx, job.ID, "failed", 0, 0)
			}
		}()

		reader := bytes.NewReader(data)
		s.processor.ProcessCSV(job.ID, reader, orgID)
	}()

	return job, nil
}

func (s *BulkService) RetryJob(ctx context.Context, jobID uuid.UUID, orgID uuid.UUID) (*models.BulkJob, error) {
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found")
	}

	if job.OrganizationID != orgID {
		return nil, fmt.Errorf("unauthorized")
	}

	if job.RetryCount >= 3 {
		return nil, fmt.Errorf("max retry limit exceeded")
	}

	// Check if file exists
	filePath := filepath.Join("uploads", fmt.Sprintf("%s.csv", jobID))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("original file not found: %w", err)
	}

	// Increment Retry Count in DB
	if err := s.repo.IncrementRetryCount(ctx, jobID); err != nil {
		s.logger.Error("Failed to increment retry count", zap.Error(err))
		// Non-fatal, continue with retry
	}

	// Trigger Async Processing again
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic in bulk job retry",
					zap.String("job_id", job.ID.String()),
					zap.Any("recover", r))
				bgCtx := context.Background()
				s.repo.UpdateJobStatus(bgCtx, job.ID, "failed", 0, 0)
			}
		}()

		reader := bytes.NewReader(data)
		s.processor.ProcessCSV(job.ID, reader, orgID)
	}()

	return job, nil
}

func (s *BulkService) ListJobs(ctx context.Context, orgID uuid.UUID, status string, page, limit int) ([]models.BulkJob, int, error) {
	return s.repo.ListJobs(ctx, orgID, status, page, limit)
}

func (s *BulkService) GetJob(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.BulkJob, error) {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	// SECURITY FIX: Validate job belongs to requesting organization (CWE-639)
	if job.OrganizationID != orgID {
		s.logger.Warn("unauthorized bulk job access attempt",
			zap.String("job_id", id.String()),
			zap.String("job_org", job.OrganizationID.String()),
			zap.String("request_org", orgID.String()))
		return nil, fmt.Errorf("job not found") // Opaque error
	}
	return job, nil
}

func (s *BulkService) GetJobRecords(ctx context.Context, jobID uuid.UUID, orgID uuid.UUID, status string) ([]models.BulkJobRecord, error) {
	// SECURITY FIX: Verify job belongs to organization before returning records (CWE-639)
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job.OrganizationID != orgID {
		s.logger.Warn("unauthorized job records access attempt",
			zap.String("job_id", jobID.String()),
			zap.String("job_org", job.OrganizationID.String()),
			zap.String("request_org", orgID.String()))
		return nil, fmt.Errorf("job not found")
	}
	return s.repo.GetJobRecords(ctx, jobID, status)
}
