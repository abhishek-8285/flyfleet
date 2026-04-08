package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics"
	"github.com/flyfleet/fleet-core/internal/domains/analytics/cache"
	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core/internal/domains/analytics/queue"
	"github.com/flyfleet/fleet-core/internal/domains/analytics/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxQueryPeriod = 90 * 24 * time.Hour
	maxSyncRows    = 1000
)

func validateQueryPeriod(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return nil
	}
	if end.Sub(start) > maxQueryPeriod {
		return fmt.Errorf("query period exceeds 90 days; please use Async Analytics Jobs for larger ranges")
	}
	return nil
}

// AnalyticsServiceImpl implements the AnalyticsService interface
type AnalyticsServiceImpl struct {
	repo    repository.AnalyticsRepository
	cache   *cache.AnalyticsCache
	queue   *queue.AnalyticsQueueClient
	metrics *analytics.Metrics
	logger  *zap.Logger
}

// NewAnalyticsService creates a new instance of AnalyticsService
func NewAnalyticsService(repo repository.AnalyticsRepository, cache *cache.AnalyticsCache, queue *queue.AnalyticsQueueClient, metrics *analytics.Metrics, logger *zap.Logger) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{
		repo:    repo,
		cache:   cache,
		queue:   queue,
		metrics: metrics,
		logger:  logger,
	}
}

// Validation functions
func validatePredictiveModel(model *models.PredictiveModel) error {
	if model.Name == "" {
		return errors.New("model name is required")
	}
	if model.Version == "" {
		return errors.New("model version is required")
	}
	if model.ModelType == "" {
		return errors.New("model type is required")
	}
	if model.Status == "" {
		return errors.New("model status is required")
	}
	if model.TrainedAt.IsZero() {
		return errors.New("trained at timestamp is required")
	}
	return nil
}

func validatePredictionResult(result *models.PredictionResult) error {
	if result.ModelID == "" {
		return errors.New("model ID is required")
	}
	if result.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		return errors.New("confidence score must be between 0 and 1")
	}
	return nil
}

func validateAnalyticsJob(job *models.AnalyticsJob) error {
	if job.Name == "" {
		return errors.New("job name is required")
	}
	if job.Type == "" {
		return errors.New("job type is required")
	}
	if job.Status == "" {
		return errors.New("job status is required")
	}
	if job.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if job.CreatedBy == uuid.Nil {
		return errors.New("created by ID is required")
	}
	return nil
}

func validateFleetAnalytics(analytics *models.FleetAnalytics) error {
	if analytics.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if analytics.StartDate.IsZero() || analytics.EndDate.IsZero() {
		return errors.New("start and end dates are required")
	}
	if analytics.StartDate.After(analytics.EndDate) {
		return errors.New("start date cannot be after end date")
	}
	return nil
}

func validateDriverBehaviorAnalytics(analytics *models.DriverBehaviorAnalytics) error {
	if analytics.DriverID == uuid.Nil {
		return errors.New("driver ID is required")
	}
	if analytics.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if analytics.AnalysisPeriodStart.IsZero() || analytics.AnalysisPeriodEnd.IsZero() {
		return errors.New("analysis period start and end dates are required")
	}
	if analytics.AnalysisPeriodStart.After(analytics.AnalysisPeriodEnd) {
		return errors.New("analysis period start cannot be after end date")
	}
	return nil
}

func validatePredictiveMaintenanceAnalytics(analytics *models.PredictiveMaintenanceAnalytics) error {
	if analytics.VehicleID == uuid.Nil {
		return errors.New("vehicle ID is required")
	}
	if analytics.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if analytics.AnalysisDate.IsZero() {
		return errors.New("analysis date is required")
	}
	return nil
}
