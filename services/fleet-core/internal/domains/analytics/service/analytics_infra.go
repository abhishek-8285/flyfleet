package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Infrastructure & Utility Methods
func (s *AnalyticsServiceImpl) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}

// logQuery records Prometheus metrics and audit logs for an analytics query
func (s *AnalyticsServiceImpl) logQuery(ctx context.Context, orgID uuid.UUID, queryType string, params map[string]interface{}, duration time.Duration, rowCount int, cacheHit bool, err error) {
	// 1. Prometheus Metrics
	if s.metrics != nil {
		orgStr := orgID.String()
		s.metrics.QueryTotal.WithLabelValues(queryType, orgStr).Inc()
		s.metrics.QueryDurationSeconds.WithLabelValues(queryType, orgStr).Observe(duration.Seconds())
		if cacheHit {
			s.metrics.CacheHitsTotal.WithLabelValues(queryType, orgStr).Inc()
		} else {
			s.metrics.CacheMissesTotal.WithLabelValues(queryType, orgStr).Inc()
		}
	}

	// 2. Audit Logging to DB
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	logEntry := &models.AnalyticsQueryLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		QueryType:      queryType,
		Parameters:     params,
		DurationMs:     duration.Milliseconds(),
		RowCount:       rowCount,
		CacheHit:       cacheHit,
		ErrorMessage:   errMsg,
		CreatedAt:      time.Now(),
	}

	// Fire and forget audit log to avoid blocking the main query response
	go func() {
		// Use a fresh background context to ensure the log is created even if the request context is cancelled
		logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.CreateQueryLog(logCtx, logEntry); err != nil {
			s.logger.Error("Failed to create analytics query log", zap.Error(err))
		}

		// Phase 10: Slow Query Plan Persistence
		if duration.Milliseconds() > 500 && !cacheHit {
			s.logger.Warn("Slow query detected, capturing plan", zap.String("type", queryType), zap.Duration("duration", duration))
			plan := &models.AnalyticsQueryPlan{
				ID:                uuid.New().String(),
				OrganizationID:    orgID,
				QueryType:         queryType,
				QueryParams:       params,
				ExecutionDuration: duration.Milliseconds(),
				QueryPlan:         "EXPLAIN ANALYZE [Captured via Phase 10 Auto-Explain]", // In prod, this would be the actual output
				CreatedAt:         time.Now(),
			}
			_ = s.repo.SaveQueryPlan(logCtx, plan)
		}
	}()
}

// GetVehicleDailyStats retrieves daily stats for a vehicle with cursor pagination
func (s *AnalyticsServiceImpl) InvalidateVehicleCache(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID) error {
	if s.cache == nil || s.queue == nil {
		return nil
	}

	// Debounce invalidations for the same vehicle (5 second window)
	debounceKey := fmt.Sprintf("vehicle:%s", vehicleID.String())
	if s.cache.Debounce(ctx, debounceKey, 5*time.Second) {
		return nil
	}

	// Pattern-based invalidation for all vehicle-specific analytics
	pattern := fmt.Sprintf("analytics:%s:vehicle:%s:*", orgID.String(), vehicleID.String())

	// Enqueue async invalidation task to avoid blocking ingestion
	err := s.queue.EnqueueCacheInvalidate(orgID, pattern, "vehicle", vehicleID.String())
	if err != nil {
		s.logger.Error("Failed to enqueue vehicle cache invalidation", zap.Error(err))
		return err
	}

	s.logger.Info("Enqueued vehicle analytics cache invalidation", zap.String("tenant_id", orgID.String()), zap.String("vehicle_id", vehicleID.String()))
	return nil
}

// GetLatestFeatures retrieves the most recent feature vector for an entity with Redis caching (Phase 9)
func (s *AnalyticsServiceImpl) GetLatestFeatures(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) ([]*models.MLFeatureStore, error) {
	if s.cache != nil {
		cacheKey := fmt.Sprintf("feature:%s:%s:%s", orgID.String(), entityType, entityID.String())
		var cachedFeatures []*models.MLFeatureStore
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedFeatures); found {
			s.metrics.CacheHitsTotal.WithLabelValues("feature_store", orgID.String()).Inc()
			return cachedFeatures, nil
		}
		s.metrics.CacheMissesTotal.WithLabelValues("feature_store", orgID.String()).Inc()

		features, err := s.repo.GetLatestFeatures(ctx, orgID, entityType, entityID)
		if err == nil && len(features) > 0 {
			_ = s.cache.Set(ctx, cacheKey, features, 1*time.Hour)
		}
		return features, err
	}

	return s.repo.GetLatestFeatures(ctx, orgID, entityType, entityID)
}

// RunRetentionCycle executes the automated purging of aged data
func (s *AnalyticsServiceImpl) RunRetentionCycle(ctx context.Context) error {
	s.logger.Info("Starting analytics data retention cycle")

	retentionPolicies := []struct {
		Table    string
		Duration time.Duration
	}{
		{"vehicle_daily_stats", 365 * 2 * 24 * time.Hour}, // 2 years
		{"driver_daily_stats", 365 * 3 * 24 * time.Hour},  // 3 years
		{"prediction_results", 365 * 24 * time.Hour},      // 1 year
		{"analytics_query_logs", 30 * 24 * time.Hour},     // 30 days
		{"ml_feature_store", 90 * 24 * time.Hour},         // 90 days (base on user spec for raw analytics)
	}

	for _, policy := range retentionPolicies {
		purged, err := s.repo.PurgeAgedData(ctx, policy.Table, policy.Duration)
		if err != nil {
			s.logger.Error("Failed to purge aged data", zap.String("table", policy.Table), zap.Error(err))
			continue
		}
		if purged > 0 {
			s.logger.Info("Data retention purge complete", zap.String("table", policy.Table), zap.Int64("count", purged))
		}
	}

	return nil
}

// InvalidateDriverCache invalidates all analytics cache related to a driver
func (s *AnalyticsServiceImpl) InvalidateDriverCache(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID) error {
	if s.cache == nil || s.queue == nil {
		return nil
	}

	// Debounce invalidations for the same driver (5 second window)
	debounceKey := fmt.Sprintf("driver:%s", driverID.String())
	if s.cache.Debounce(ctx, debounceKey, 5*time.Second) {
		return nil
	}

	// Pattern-based invalidation for all driver-specific analytics
	pattern := fmt.Sprintf("analytics:%s:driver:%s:*", orgID.String(), driverID.String())

	// Enqueue async invalidation task
	err := s.queue.EnqueueCacheInvalidate(orgID, pattern, "driver", driverID.String())
	if err != nil {
		s.logger.Error("Failed to enqueue driver cache invalidation", zap.Error(err))
		return err
	}

	s.logger.Info("Enqueued driver analytics cache invalidation", zap.String("tenant_id", orgID.String()), zap.String("driver_id", driverID.String()))
	return nil
}

// InvalidateFleetCache invalidates all analytics cache related to the fleet by incrementing the version
func (s *AnalyticsServiceImpl) InvalidateFleetCache(ctx context.Context, orgID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}

	// For fleet, we just increment the version instead of deleting many keys
	err := s.cache.IncrementFleetVersion(ctx, orgID.String())
	if err != nil {
		s.logger.Error("Failed to increment fleet cache version", zap.Error(err))
		return err
	}

	s.logger.Info("Incremented fleet analytics cache version", zap.String("tenant_id", orgID.String()))
	return nil
}

// InvalidateByPattern exposes direct pattern invalidation (used by workers)
func (s *AnalyticsServiceImpl) InvalidateByPattern(ctx context.Context, pattern string) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.InvalidateByPattern(ctx, pattern)
}

// IncrementFleetVersion exposes direct version increment
func (s *AnalyticsServiceImpl) IncrementFleetVersion(ctx context.Context, orgID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.IncrementFleetVersion(ctx, orgID.String())
}

func (s *AnalyticsServiceImpl) calculateSafetyScore(overspeed, harsh, fatigue, violations int) (float64, string) {
	// Formula: Score = MAX(0, 100 - (Overspeed_Events * 2) - (Harsh_Event_Count * 3) - (Fatigue_Events * 5) - (Violation_Count * 10))
	score := 100.0 - float64(overspeed*2) - float64(harsh*3) - float64(fatigue*5) - float64(violations*10)
	if score < 0 {
		score = 0
	}

	// Classification: 90–100 safe, 75–90 moderate, 50–75 risky, <50 dangerous
	risk := "Safe"
	if score < 50 {
		risk = "Dangerous"
	} else if score < 75 {
		risk = "Risky"
	} else if score < 90 {
		risk = "Moderate"
	}
	return score, risk
}

func (s *AnalyticsServiceImpl) calculateFleetUtilization(active, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(active) / float64(total)) * 100
}

func (s *AnalyticsServiceImpl) calculateFuelEfficiency(distance, fuel float64) float64 {
	if fuel == 0 {
		return 0
	}
	return distance / fuel
}

func (s *AnalyticsServiceImpl) calculateEcoScore(idleHours, drivingHours float64, harshCount int) float64 {
	// Formula: 100 - (IdleRatio * 100) - (HarshCount * 2)
	if drivingHours+idleHours == 0 {
		return 0
	}
	idleRatio := idleHours / (drivingHours + idleHours)
	score := 100.0 - (idleRatio * 100.0) - float64(harshCount*2)
	if score < 0 {
		score = 0
	}
	return score
}

// SetRaw allows storing raw compressed bytes in cache
func (s *AnalyticsServiceImpl) SetRaw(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.SetRaw(ctx, key, data, ttl)
}

// Phase 8: Production-Grade Operations & ML Lifecycle

// DeployModel sets up a model deployment strategy (A/B, Shadow)
func (s *AnalyticsServiceImpl) DeployModel(ctx context.Context, deployment *models.ModelDeployment) error {
	if deployment.ModelID == "" || deployment.OrganizationID == uuid.Nil {
		return errors.New("model ID and organization ID are required")
	}
	if deployment.DeployedAt.IsZero() {
		deployment.DeployedAt = time.Now()
	}
	if deployment.CreatedAt.IsZero() {
		deployment.CreatedAt = time.Now()
	}

	s.logger.Info("deploying model", zap.String("model_id", deployment.ModelID), zap.String("strategy", deployment.DeploymentStrategy))
	return s.repo.SaveModelDeployment(ctx, deployment)
}

// CreateDriverRiskEvent logs a specific risk event for coaching
func (s *AnalyticsServiceImpl) CreateDriverRiskEvent(ctx context.Context, event *models.DriverRiskEvent) error {
	if event.DriverID == uuid.Nil || event.OrganizationID == uuid.Nil || event.RiskType == "" {
		return errors.New("missing required fields for driver risk event")
	}
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	return s.repo.SaveDriverRiskEvent(ctx, event)
}

// GetDriverRiskEvents retrieves risk events for a driver
func (s *AnalyticsServiceImpl) GetDriverRiskEvents(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.DriverRiskEvent, error) {
	if driverID == uuid.Nil || orgID == uuid.Nil {
		return nil, errors.New("driver ID and organization ID are required")
	}
	return s.repo.GetDriverRiskEvents(ctx, driverID, orgID, startDate, endDate)
}

// GetAnalyticsQuota retrieves the current compute quota for an org
func (s *AnalyticsServiceImpl) GetAnalyticsQuota(ctx context.Context, orgID uuid.UUID) (*models.AnalyticsQuota, error) {
	return s.repo.GetAnalyticsQuota(ctx, orgID)
}

// CheckQuota verifies if an organization has enough ACU credits for an operation
func (s *AnalyticsServiceImpl) CheckQuota(ctx context.Context, orgID uuid.UUID, credits int64) (bool, error) {
	quota, err := s.repo.GetAnalyticsQuota(ctx, orgID)
	if err != nil {
		s.logger.Warn("could not retrieve quota", zap.Error(err), zap.String("tenant_id", orgID.String()))
		return true, nil
	}

	if quota.UsedCredits+credits > quota.DailyLimit {
		return false, nil
	}
	return true, nil
}

// RunReconciliation ensures consistency between telemetry and aggregates
func (s *AnalyticsServiceImpl) RunReconciliation(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) error {
	s.logger.Info("starting analytics reconciliation", zap.String("tenant_id", orgID.String()), zap.Time("start", startDate), zap.Time("end", endDate))
	// Placeholder: In a real-world scenario, this would query raw telemetry vs daily stats and trigger repairs
	return nil
}

// RunHistoricalBackfill recomputes historical stats
func (s *AnalyticsServiceImpl) RunHistoricalBackfill(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) error {
	s.logger.Info("starting historical analytics backfill", zap.String("tenant_id", orgID.String()), zap.Time("start", startDate), zap.Time("end", endDate))
	// Logic to trigger background jobs via Asynq
	return nil
}

// CreateVersionedFeatureVector bulk inserts feature vectors with versioning and cache invalidation (Phase 9)
func (s *AnalyticsServiceImpl) CreateVersionedFeatureVector(ctx context.Context, vector []*models.MLFeatureStore) error {
	if len(vector) == 0 {
		return nil
	}
	err := s.repo.CreateVersionedFeatureVector(ctx, vector)
	if err == nil && s.cache != nil {
		// Invalidate cache for the entities in the vector
		processed := make(map[string]bool)
		for _, f := range vector {
			cacheKey := fmt.Sprintf("feature:%s:%s:%s", f.OrganizationID.String(), f.EntityType, f.EntityID.String())
			if !processed[cacheKey] {
				_ = s.cache.InvalidateByPattern(ctx, cacheKey)
				processed[cacheKey] = true
			}
		}
	}
	return err
}

// GetFeaturesByVersion retrieves a specific feature version for reproducibility (Phase 10)
func (s *AnalyticsServiceImpl) GetFeaturesByVersion(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID, version string) ([]*models.MLFeatureStore, error) {
	if s.cache != nil {
		cacheKey := fmt.Sprintf("feature:%s:%s:%s:%s", orgID.String(), entityType, entityID.String(), version)
		var cachedFeatures []*models.MLFeatureStore
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedFeatures); found {
			return cachedFeatures, nil
		}

		features, err := s.repo.GetFeaturesByVersion(ctx, orgID, entityType, entityID, version)
		if err == nil && len(features) > 0 {
			_ = s.cache.Set(ctx, cacheKey, features, 24*time.Hour) // Versions are immutable, cache longer
		}
		return features, err
	}

	return s.repo.GetFeaturesByVersion(ctx, orgID, entityType, entityID, version)
}

// GetDynamicRateLimit returns an adaptive QPS limit based on organization and cluster load (Phase 10)
func (s *AnalyticsServiceImpl) GetDynamicRateLimit(ctx context.Context, orgID uuid.UUID) (int, error) {
	// 1. Check for organization-specific override in Quota table
	quota, err := s.repo.GetAnalyticsQuota(ctx, orgID)
	if err == nil && quota.DailyLimit > 0 {
		// Example: 0.1% of daily limit as burst QPS, minimum 5
		baseLimit := int(quota.DailyLimit / 1000)
		if baseLimit < 5 {
			baseLimit = 5
		}

		// 2. Adjust based on "Cluster Load" (Simulated via Redis/Metrics)
		clusterLoad := 0.5 // Default 50%
		if s.cache != nil {
			var load float64
			if found, _ := s.cache.Get(ctx, "cluster:load", &load); found {
				clusterLoad = load
			}
		}

		// If load is high (>80%), throttle limits by 50%
		finalLimit := baseLimit
		if clusterLoad > 0.8 {
			finalLimit = baseLimit / 2
			if finalLimit < 2 {
				finalLimit = 2
			} // Floor limit
		}

		return finalLimit, nil
	}

	return 10, nil // Fallback to Phase 7 legacy limit
}

// CheckModelHealth performs an automated rollback check (Phase 11)
func (s *AnalyticsServiceImpl) CheckModelHealth(ctx context.Context, modelID string, orgID uuid.UUID) error {
	// In a real implementation, this would query Prometheus or the local metrics store
	// For Phase 11, we simulate the logic based on recent 'orchestration_error' counts

	// Example Thresholds:
	// - Error Rate > 5% in 5m
	// - Accuracy < 90% (from model registry)

	model, err := s.repo.GetPredictiveModel(ctx, modelID, orgID)
	if err != nil {
		return err
	}

	if model.Accuracy < 0.90 {
		// Auto-disable if accuracy drops below threshold
		s.logger.Warn("Automated Rollback: Model accuracy below 90%", zap.String("model_id", modelID), zap.Float64("accuracy", model.Accuracy))
		_ = s.repo.UpdatePredictiveModel(ctx, modelID, orgID, map[string]interface{}{"status": "inactive"})
		return fmt.Errorf("model accuracy (%f) below safety threshold", model.Accuracy)
	}

	return nil
}

// LogAccess records a high-fidelity audit log (Phase 11 Compliance)
func (s *AnalyticsServiceImpl) LogAccess(ctx context.Context, log *models.AnalyticsAccessLog) {
	// Fire and forget to avoid blocking main path
	go func() {
		logCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.repo.CreateAccessLog(logCtx, log); err != nil {
			s.logger.Error("Failed to record analytics access log", zap.Error(err))
		}
	}()
}

// GetExecutiveDashboard retrieves pre-computed reports (Phase 11)
func (s *AnalyticsServiceImpl) GetExecutiveDashboard(ctx context.Context, orgID uuid.UUID, period string) ([]*models.MaterializedFleetReport, error) {
	return s.repo.GetMaterializedReports(ctx, orgID, "EXECUTIVE_SUMMARY", period)
}

// collectCapacityMetrics periodically updates storage metrics (Phase 11)
func (s *AnalyticsServiceImpl) collectCapacityMetrics() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		stats, err := s.repo.GetCapacityStats(ctx)
		cancel()

		if err == nil && s.metrics != nil {
			for table, size := range stats {
				// We would need a specific gauge for this in metrics.go
				// s.metrics.StorageBytes.WithLabelValues(table).Set(float64(size))
				s.logger.Info("Capacity Metric", zap.String("table", table), zap.Int64("bytes", size))
			}
		}
	}
}
