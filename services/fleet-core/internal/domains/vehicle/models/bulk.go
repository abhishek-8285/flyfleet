package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BulkJob struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"organization_id"`
	FileName       string    `json:"file_name"`
	Status         string    `json:"status"` // processing, completed, failed, partial_success
	TotalRecords   int       `json:"total_records"`
	SuccessCount   int       `json:"success_count"`
	FailedCount    int       `json:"failed_count"`
	RetryCount     int       `json:"retry_count"`
	CreatedBy      uuid.UUID `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BulkJobRecord struct {
	ID           uuid.UUID       `json:"id"`
	JobID        uuid.UUID       `json:"job_id"`
	RowNumber    int             `json:"row_number"`
	RecordData   json.RawMessage `json:"record_data"`
	Status       string          `json:"status"` // success, failed
	ErrorMessage string          `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type BulkConfig struct {
	TenantID          uuid.UUID       `json:"organization_id"`
	MaxRecordsPerFile int             `json:"max_records_per_file"`
	ConcurrentThreads int             `json:"concurrent_threads"`
	RetentionMonths   int             `json:"retention_months"`
	MandatoryFields   json.RawMessage `json:"mandatory_fields"` // Array of strings
	UpdatedAt         time.Time       `json:"updated_at"`
}
