package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BulkRepository struct {
	db *pgxpool.Pool
}

func NewBulkRepository(db *pgxpool.Pool) *BulkRepository {
	return &BulkRepository{db: db}
}

func (r *BulkRepository) CreateJob(ctx context.Context, job *models.BulkJob) error {
	query := `
		INSERT INTO bulk_jobs (id, organization_id, file_name, status, total_records, success_count, failed_count, retry_count, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		job.ID, job.OrganizationID, job.FileName, job.Status, job.TotalRecords, job.SuccessCount, job.FailedCount, job.RetryCount, job.CreatedBy, job.CreatedAt, job.UpdatedAt)
	return err
}

func (r *BulkRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status string, success, failed int) error {
	query := `
		UPDATE bulk_jobs 
		SET status = $1, success_count = $2, failed_count = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, status, success, failed, id)
	return err
}

func (r *BulkRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE bulk_jobs 
		SET retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *BulkRepository) GetJob(ctx context.Context, id uuid.UUID) (*models.BulkJob, error) {
	query := `SELECT id, organization_id, file_name, status, total_records, success_count, failed_count, retry_count, created_by, created_at, updated_at FROM bulk_jobs WHERE id = $1`
	
	var job models.BulkJob
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.OrganizationID, &job.FileName, &job.Status, 
		&job.TotalRecords, &job.SuccessCount, &job.FailedCount, &job.RetryCount,
		&job.CreatedBy, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *BulkRepository) AddJobRecord(ctx context.Context, record *models.BulkJobRecord) error {
	recordDataJSON, err := json.Marshal(record.RecordData)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO bulk_job_records (id, job_id, row_number, record_data, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.Exec(ctx, query,
		record.ID, record.JobID, record.RowNumber, recordDataJSON, record.Status, record.ErrorMessage, record.CreatedAt)
	return err
}

func (r *BulkRepository) GetJobRecords(ctx context.Context, jobID uuid.UUID, status string) ([]models.BulkJobRecord, error) {
	query := `SELECT id, job_id, row_number, record_data, status, error_message, created_at FROM bulk_job_records WHERE job_id = $1`
	var args []interface{}
	args = append(args, jobID)
	
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	
	query += " ORDER BY row_number ASC"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []models.BulkJobRecord
	for rows.Next() {
		var rec models.BulkJobRecord
		var recordDataBytes []byte
		if err := rows.Scan(
			&rec.ID, &rec.JobID, &rec.RowNumber, &recordDataBytes, 
			&rec.Status, &rec.ErrorMessage, &rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(recordDataBytes, &rec.RecordData)
		records = append(records, rec)
	}
	return records, nil
}

func (r *BulkRepository) ListJobs(ctx context.Context, orgID uuid.UUID, status string, page, limit int) ([]models.BulkJob, int, error) {
    offset := (page - 1) * limit
    
	// List jobs
	// Base query
    query := `SELECT id, organization_id, file_name, status, total_records, success_count, failed_count, retry_count, created_by, created_at, updated_at FROM bulk_jobs WHERE organization_id = $1`
    countQuery := `SELECT COUNT(*) FROM bulk_jobs WHERE organization_id = $1`
    
    args := []interface{}{orgID}
    if status != "" {
        query += ` AND status = $2`
        countQuery += ` AND status = $2`
        args = append(args, status)
    }
    
    // Add sorting and pagination
    query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
    args = append(args, limit, offset)
    
    // Execute count query
    var total int
    countArgs := args[:len(args)-2] // Remove limit and offset
    err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    // Execute list query
    rows, err := r.db.Query(ctx, query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var jobs []models.BulkJob
    for rows.Next() {
        var job models.BulkJob
        if err := rows.Scan(
            &job.ID, &job.OrganizationID, &job.FileName, &job.Status, 
            &job.TotalRecords, &job.SuccessCount, &job.FailedCount, &job.RetryCount,
            &job.CreatedBy, &job.CreatedAt, &job.UpdatedAt,
        ); err != nil {
            return nil, 0, err
        }
        jobs = append(jobs, job)
    }
    
    return jobs, total, nil
}
