package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
)

func (r *AnalyticsRepositoryImpl) CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		INSERT INTO analytics_jobs (id, name, type, description, status, job_type, parameters, 
			input_source, output_destination, start_time, end_time, progress, error_details, 
			result_key, organization_id, created_by, created_at, updated_at, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	parametersJSON, err := json.Marshal(job.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Handle nullable fields
	var startTime *time.Time
	if job.StartTime != nil {
		startTime = job.StartTime
	}

	var endTime *time.Time
	if job.EndTime != nil {
		endTime = job.EndTime
	}

	var errorDetails *string
	if job.ErrorDetails != nil {
		errorDetails = job.ErrorDetails
	}

	_, err = r.db.Exec(ctx, query, job.ID, job.Name, job.Type, job.Description, job.Status, job.JobType,
		parametersJSON, job.InputSource, job.OutputDestination, startTime, endTime, job.Progress,
		errorDetails, job.ResultKey, job.OrganizationID, job.CreatedBy, job.CreatedAt, job.UpdatedAt, job.IdempotencyKey)

	if err != nil {
		if strings.Contains(err.Error(), "idx_analytics_jobs_dedup") || strings.Contains(err.Error(), "unique constraint") {
			return ErrDuplicateJob
		}
		return fmt.Errorf("failed to create analytics job: %w", err)
	}
	return nil
}

// GetAnalyticsJob retrieves an analytics job by ID
func (r *AnalyticsRepositoryImpl) GetAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) (*models.AnalyticsJob, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, name, type, description, status, job_type, parameters, 
			input_source, output_destination, start_time, end_time, progress, error_details, 
			result_key, organization_id, created_by, created_at, updated_at, idempotency_key
		FROM analytics_jobs WHERE id = $1 AND organization_id = $2
	`

	var job models.AnalyticsJob
	var parametersJSON []byte
	var startTime, endTime *time.Time
	var errorDetails *string

	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&job.ID, &job.Name, &job.Type, &job.Description, &job.Status, &job.JobType, &parametersJSON,
		&job.InputSource, &job.OutputDestination, &startTime, &endTime, &job.Progress, &errorDetails,
		&job.ResultKey, &job.OrganizationID, &job.CreatedBy, &job.CreatedAt, &job.UpdatedAt, &job.IdempotencyKey,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analytics job %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get analytics job: %w", err)
	}

	err = json.Unmarshal(parametersJSON, &job.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	job.StartTime = startTime
	job.EndTime = endTime
	job.ErrorDetails = errorDetails

	return &job, nil
}

// UpdateAnalyticsJob updates an existing analytics job
func (r *AnalyticsRepositoryImpl) UpdateAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}

	args = append(args, id, orgID)
	query := fmt.Sprintf("UPDATE analytics_jobs SET %s WHERE id = $%d AND organization_id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteAnalyticsJob deletes an analytics job
func (r *AnalyticsRepositoryImpl) DeleteAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM analytics_jobs WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete analytics job %s: %w", id, err)
	}
	return nil
}

// ListAnalyticsJobs retrieves analytics jobs with filtering
func (r *AnalyticsRepositoryImpl) ListAnalyticsJobs(ctx context.Context, jobType, status *string, orgID uuid.UUID, cursor string, limit, offset int) ([]*models.AnalyticsJob, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	baseQuery := `
		SELECT id, name, type, description, status, job_type, parameters,
		       input_source, output_destination, start_time, end_time, progress, error_details,
		       result_key, organization_id, created_by, created_at, updated_at, idempotency_key
		FROM analytics_jobs
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM analytics_jobs WHERE organization_id = $1`

	args := []interface{}{orgID}
	argIndex := 2
	var conditions []string

	if jobType != nil && *jobType != "" {
		conditions = append(conditions, fmt.Sprintf("job_type = $%d", argIndex))
		args = append(args, *jobType)
		argIndex++
	}

	if status != nil && *status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Count total
	var totalCount int
	countSQL := countQuery + whereClause
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count analytics jobs: %w", err)
	}

	// List with pagination and cursor
	query := baseQuery + whereClause
	if cursor != "" {
		query += " AND created_at < $" + fmt.Sprintf("%d", argIndex)
		args = append(args, cursor)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query analytics jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.AnalyticsJob
	for rows.Next() {
		var job models.AnalyticsJob
		var parametersJSON []byte
		var startTime, endTime *time.Time
		var errorDetails *string

		err := rows.Scan(
			&job.ID, &job.Name, &job.Type, &job.Description, &job.Status, &job.JobType,
			&parametersJSON, &job.InputSource, &job.OutputDestination,
			&startTime, &endTime, &job.Progress, &errorDetails,
			&job.ResultKey, &job.OrganizationID, &job.CreatedBy, &job.CreatedAt, &job.UpdatedAt, &job.IdempotencyKey,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan analytics job: %w", err)
		}

		if err := json.Unmarshal(parametersJSON, &job.Parameters); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal job parameters: %w", err)
		}

		job.StartTime = startTime
		job.EndTime = endTime
		job.ErrorDetails = errorDetails

		jobs = append(jobs, &job)
	}

	return jobs, totalCount, nil
}
