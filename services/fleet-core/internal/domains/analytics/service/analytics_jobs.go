package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core/internal/domains/analytics/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Analytics Job Methods
func (s *AnalyticsServiceImpl) CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error {
	if err := validateAnalyticsJob(job); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.Status == "" {
		job.Status = "pending"
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}

	// Generate IdempotencyKey if not provided (Phase 9)
	if job.IdempotencyKey == "" {
		// Use a stable hash of type + parameters as a default idempotency key
		paramsJSON, _ := json.Marshal(job.Parameters)
		hashBytes := sha256.Sum256(paramsJSON)
		job.IdempotencyKey = fmt.Sprintf("%s:%x", job.JobType, hashBytes)
	}

	if err := s.repo.CreateAnalyticsJob(ctx, job); err != nil {
		if errors.Is(err, repository.ErrDuplicateJob) {
			s.logger.Info("Blocking duplicate analytics job", zap.String("idempotency_key", job.IdempotencyKey))
			return nil // Or return the existing job ID if we had a way to fetch it
		}
		return err
	}

	if s.queue != nil {
		if err := s.queue.EnqueueAnalyticsJob(job.ID, job.OrganizationID, job.CreatedBy); err != nil {
			s.logger.Error("Failed to enqueue analytics job", zap.String("job_id", job.ID), zap.Error(err))
		}
	}

	return nil
}

// GetAnalyticsJob retrieves an analytics job by ID
func (s *AnalyticsServiceImpl) GetAnalyticsJob(ctx context.Context, id string, tenantID uuid.UUID) (*models.AnalyticsJob, error) {
	start := time.Now()
	var err error
	var res *models.AnalyticsJob
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"id": id}
		s.logQuery(ctx, tenantID, "get_analytics_job", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	if s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "analytics_job", params)
		if exists, _ := s.cache.Get(ctx, cacheKey, &res); exists {
			cacheHit = true
			return res, nil
		}

		res, err = s.repo.GetAnalyticsJob(ctx, id, tenantID)
		if err == nil && res != nil {
			params := map[string]interface{}{"id": id}
			cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "analytics_job", params)
			_ = s.cache.Set(ctx, cacheKey, res, 0)
		}
		return res, err
	}

	res, err = s.repo.GetAnalyticsJob(ctx, id, tenantID)
	return res, err
}

// UpdateAnalyticsJob updates an existing analytics job with validation
func (s *AnalyticsServiceImpl) UpdateAnalyticsJob(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	err := s.repo.UpdateAnalyticsJob(ctx, id, tenantID, updates)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "analytics_job", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// DeleteAnalyticsJob deletes an analytics job
func (s *AnalyticsServiceImpl) DeleteAnalyticsJob(ctx context.Context, id string, tenantID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	err := s.repo.DeleteAnalyticsJob(ctx, id, tenantID)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "analytics_job", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// ListAnalyticsJobs retrieves analytics jobs with filtering
func (s *AnalyticsServiceImpl) ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, tenantID uuid.UUID, cursor string, limit, offset int) ([]*models.AnalyticsJob, int, error) {
	start := time.Now()
	var err error
	var data []*models.AnalyticsJob
	var total int
	cacheHit := false

	defer func() {
		params := map[string]interface{}{"limit": limit, "offset": offset}
		if jobType != nil {
			params["job_type"] = *jobType
		}
		if status != nil {
			params["status"] = *status
		}
		s.logQuery(ctx, tenantID, "list_analytics_jobs", params, time.Since(start), len(data), cacheHit, err)
	}()

	if tenantID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if jobType != nil {
			params["job_type"] = *jobType
		}
		if status != nil {
			params["status"] = *status
		}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "list_analytics_jobs", params)
		type listResult struct {
			Data  []*models.AnalyticsJob
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListAnalyticsJobs(ctx, jobType, status, tenantID, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListAnalyticsJobs(ctx, jobType, status, tenantID, cursor, limit, offset)
	return data, total, err
}

// CreateFleetAnalytics creates new fleet analytics with validation
