package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/cache"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/queue"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/repository"
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

// CreatePredictiveModel creates a new predictive model with validation
func (s *AnalyticsServiceImpl) CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error {
	if err := validatePredictiveModel(model); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now()
	}
	if model.UpdatedAt.IsZero() {
		model.UpdatedAt = time.Now()
	}

	return s.repo.CreatePredictiveModel(ctx, model)
}

// GetPredictiveModel retrieves a predictive model by ID
func (s *AnalyticsServiceImpl) GetPredictiveModel(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveModel, error) {
	start := time.Now()
	var err error
	var res *models.PredictiveModel
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"id": id}
		s.logQuery(ctx, orgID, "get_predictive_model", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	if s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "predictive_model", params)
		if found, _ := s.cache.Get(ctx, cacheKey, &res); found {
			cacheHit = true
			return res, nil
		}

		res, err = s.repo.GetPredictiveModel(ctx, id, orgID)
		if err == nil && res != nil {
			params := map[string]interface{}{"id": id}
			cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "predictive_model", params)
			_ = s.cache.Set(ctx, cacheKey, res, 0)
		}
		return res, err
	}

	res, err = s.repo.GetPredictiveModel(ctx, id, orgID)
	return res, err
}

// GetPredictiveModelByVersion retrieves a predictive model by name and version
func (s *AnalyticsServiceImpl) GetPredictiveModelByVersion(ctx context.Context, name, version string, orgID uuid.UUID) (*models.PredictiveModel, error) {
	start := time.Now()
	var err error
	var res *models.PredictiveModel

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"name": name, "version": version}
		s.logQuery(ctx, orgID, "get_predictive_model_version", params, time.Since(start), rowCount, false, err)
	}()

	if name == "" {
		err = errors.New("name is required")
		return nil, err
	}
	if version == "" {
		err = errors.New("version is required")
		return nil, err
	}

	res, err = s.repo.GetPredictiveModelByVersion(ctx, name, version, orgID)
	return res, err
}

// UpdatePredictiveModel updates an existing predictive model with validation
func (s *AnalyticsServiceImpl) UpdatePredictiveModel(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	err := s.repo.UpdatePredictiveModel(ctx, id, orgID, updates)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "predictive_model", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// DeletePredictiveModel deletes a predictive model
func (s *AnalyticsServiceImpl) DeletePredictiveModel(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	err := s.repo.DeletePredictiveModel(ctx, id, orgID)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "predictive_model", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// ListPredictiveModels retrieves predictive models with filtering
func (s *AnalyticsServiceImpl) ListPredictiveModels(ctx context.Context, orgID uuid.UUID, modelType *string, status *string, cursor string, limit, offset int) ([]*models.PredictiveModel, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictiveModel
	var total int

	defer func() {
		params := map[string]interface{}{"limit": limit, "offset": offset}
		if modelType != nil {
			params["model_type"] = *modelType
		}
		if status != nil {
			params["status"] = *status
		}
		s.logQuery(ctx, orgID, "list_predictive_models", params, time.Since(start), len(data), false, err)
	}()

	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	data, total, err = s.repo.ListPredictiveModels(ctx, orgID, modelType, status, cursor, limit, offset)
	return data, total, err
}

// RunInference executes a specific predictive model on an entity using Phase 8 deployment strategies
func (s *AnalyticsServiceImpl) RunInference(ctx context.Context, modelID string, entityID uuid.UUID, orgID uuid.UUID, input map[string]interface{}) (*models.PredictionResult, error) {
	start := time.Now()
	
	// Quota Check (Phase 8)
	if allowed, _ := s.CheckQuota(ctx, orgID, 1); !allowed {
		return nil, errors.New("organization analytics quota exceeded for today")
	}

	// Fetch Model
	model, err := s.repo.GetPredictiveModel(ctx, modelID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model: %w", err)
	}
	if model.Status != "active" {
		return nil, fmt.Errorf("model %s is not active (status: %s)", modelID, model.Status)
	}

	// 1. Compliance Audit (Phase 11)
	s.LogAccess(ctx, &models.AnalyticsAccessLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         uuid.Nil, // Should be extracted from context in real implementation
		ResourceID:     modelID,
		ResourceType:   "MODEL",
		AccessContext:  map[string]interface{}{"action": "inference", "version": model.Version},
		CreatedAt:      time.Now(),
	})

	// 2. Automated Rollback Check (Phase 11)
	if err := s.CheckModelHealth(ctx, modelID, orgID); err != nil {
		s.logger.Error("Model health check failed, blocking inference", zap.String("model_id", modelID), zap.Error(err))
		return nil, fmt.Errorf("model safety guard triggered: %w", err)
	}

	// Implementation of reproducible inference (Phase 10)
	// If a specific feature version is requested, we fetch that specifically
	if requestedVersion, ok := input["feature_version"].(string); ok && requestedVersion != "" {
		s.logger.Info("Executing reproducible inference with pinned features", 
			zap.String("version", requestedVersion),
			zap.String("model_id", modelID))
		
		features, err := s.GetFeaturesByVersion(ctx, orgID, "VEHICLE", entityID, requestedVersion)
		if err == nil && len(features) > 0 {
			// Inject these features into the input map for the specific scoring logic
			for _, f := range features {
				input[f.FeatureName] = f.FeatureValue
			}
		} else if err != nil {
			s.logger.Warn("Pinned features not found, falling back to latest", zap.String("version", requestedVersion), zap.Error(err))
		}
	}

	// Check for Deployment Strategy (Phase 8)
	deployment, err := s.repo.GetActiveModelDeployment(ctx, modelID, orgID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Warn("failed to retrieve deployment info, falling back to standard", zap.Error(err))
	}

	var primaryResult *models.PredictionResult
	if deployment != nil && deployment.IsActive {
		switch deployment.DeploymentStrategy {
		case "shadow":
			// Execute primary (specified in deployment or current active model)
			primaryResult, err = s.executeInference(ctx, model, entityID, orgID, input)
			// Execute Shadow (Log/Observe but don't return)
			go s.executeShadowInference(ctx, deployment.ModelID, entityID, orgID, input)
			
		case "ab":
			// Deterministic split based on entityID to ensure consistency for that vehicle/driver
			hashBytes := sha256.Sum256([]byte(entityID.String()))
			hash := uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
			if float64(hash % 100) < deployment.TrafficSplit {
				// Traffic routed to variant
				primaryResult, err = s.executeInference(ctx, model, entityID, orgID, input)
			} else {
				// Traffic routed to control
				primaryResult, err = s.executeInference(ctx, model, entityID, orgID, input)
			}
		default:
			primaryResult, err = s.executeInference(ctx, model, entityID, orgID, input)
		}
	} else {
		primaryResult, err = s.executeInference(ctx, model, entityID, orgID, input)
	}

	if err != nil {
		s.metrics.PredictionErrorCounter.WithLabelValues(orgID.String(), modelID, "orchestration_error").Inc()
		return nil, err
	}

	// Quota tracking update (Phase 8)
	_ = s.repo.UpdateUsedCredits(ctx, orgID, 1)
	
	duration := time.Since(start)
	s.metrics.InferenceLatencyHistogram.WithLabelValues(orgID.String(), modelID).Observe(duration.Seconds())

	s.logger.Info("Inference orchestration complete", 
		zap.String("model_id", modelID), 
		zap.String("entity_id", entityID.String()),
		zap.Duration("duration", duration))

	return primaryResult, nil
}

func (s *AnalyticsServiceImpl) executeInference(ctx context.Context, model *models.PredictiveModel, entityID uuid.UUID, orgID uuid.UUID, input map[string]interface{}) (*models.PredictionResult, error) {
	var predictedValue interface{}
	var confidence float64
	var anomaly bool
	var aScore *float64

	switch model.ModelType {
	case "MAINTENANCE":
		predictedValue, confidence = s.scoreMaintenanceModel(model, input)
	case "RISK":
		predictedValue, confidence, anomaly, aScore = s.scoreRiskModel(model, input)
	case "FUEL_EFFICIENCY":
		predictedValue, confidence = s.scoreFuelEfficiencyModel(model, input)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", model.ModelType)
	}

	result := &models.PredictionResult{
		ID:              uuid.New().String(),
		ModelID:         model.ID,
		OrganizationID:  orgID,
		PredictionInput: input,
		PredictedValue:  predictedValue,
		ConfidenceScore: confidence,
		AnomalyDetected: anomaly,
		AnomalyScore:    aScore,
		PredictedAt:     time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.CreatePredictionResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save prediction result: %w", err)
	}

	s.metrics.PredictiveAnalyticCounter.WithLabelValues(orgID.String(), model.ModelType, "success").Inc()
	return result, nil
}

func (s *AnalyticsServiceImpl) executeShadowInference(ctx context.Context, modelID string, entityID uuid.UUID, orgID uuid.UUID, input map[string]interface{}) {
	// Shadow execution should not affect primary flow
	// Placeholder for Drift Detection comparison
	s.logger.Debug("Shadow inference execution (Phase 8 Observer)", zap.String("model_id", modelID))
}

func (s *AnalyticsServiceImpl) scoreMaintenanceModel(model *models.PredictiveModel, input map[string]interface{}) (interface{}, float64) {
	// Simplified production scoring logic
	// Feature weighting based on technical spec
	engineTemp, _ := input["engine_temp"].(float64)
	vibrationLevel, _ := input["vibration_level"].(float64)
	lastServiceDays, _ := input["last_service_days"].(float64)

	// Formula: RUL prediction
	score := 100.0 - (engineTemp * 0.1) - (vibrationLevel * 0.5) - (lastServiceDays * 0.05)
	if score < 0 { score = 0 }
	
	status := "healthy"
	if score < 40 {
		status = "critical"
	} else if score < 70 {
		status = "warning"
	}

	return map[string]interface{}{
		"rul_score": score,
		"status":    status,
	}, 0.85 // Static confidence for demo
}

func (s *AnalyticsServiceImpl) scoreRiskModel(model *models.PredictiveModel, input map[string]interface{}) (interface{}, float64, bool, *float64) {
	safetyScore, _ := input["safety_score"].(float64)
	hoursDriving, _ := input["hours_driving"].(float64)
	
	riskProbability := (100.0 - safetyScore)*0.5 + (hoursDriving * 2.0)
	if riskProbability > 100 { riskProbability = 100 }

	anomaly := false
	var aScore *float64
	if safetyScore < 20 && hoursDriving > 12 {
		anomaly = true
		val := 0.95
		aScore = &val
	}

	return map[string]interface{}{
		"risk_probability": riskProbability / 100.0,
		"risk_tier": s.getRiskTier(riskProbability),
	}, 0.92, anomaly, aScore
}

func (s *AnalyticsServiceImpl) getRiskTier(prob float64) string {
	if prob > 80 { return "CRITICAL" }
	if prob > 50 { return "HIGH" }
	if prob > 20 { return "MEDIUM" }
	return "LOW"
}

func (s *AnalyticsServiceImpl) scoreFuelEfficiencyModel(model *models.PredictiveModel, input map[string]interface{}) (interface{}, float64) {
	distance, _ := input["distance"].(float64)
	fuel, _ := input["fuel"].(float64)
	if fuel == 0 { return 0.0, 0.0 }
	
	return map[string]interface{}{
		"predicted_kml": distance / fuel,
	}, 0.98
}

// CreatePredictionResult creates a new prediction result with validation
func (s *AnalyticsServiceImpl) CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error {
	if err := validatePredictionResult(result); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if result.ID == "" {
		result.ID = uuid.New().String()
	}
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}
	if result.UpdatedAt.IsZero() {
		result.UpdatedAt = time.Now()
	}
	if result.PredictedAt.IsZero() {
		result.PredictedAt = time.Now()
	}

	return s.repo.CreatePredictionResult(ctx, result)
}

// GetPredictionResult retrieves a prediction result by ID
func (s *AnalyticsServiceImpl) GetPredictionResult(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictionResult, error) {
	start := time.Now()
	var err error
	var res *models.PredictionResult

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"id": id}
		s.logQuery(ctx, orgID, "get_prediction_result", params, time.Since(start), rowCount, false, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetPredictionResult(ctx, id, orgID)
	return res, err
}

// GetPredictionResultsByModel retrieves prediction results by model ID
func (s *AnalyticsServiceImpl) GetPredictionResultsByModel(ctx context.Context, modelID string, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictionResult
	var total int

	defer func() {
		params := map[string]interface{}{"model_id": modelID, "limit": limit, "offset": offset}
		s.logQuery(ctx, orgID, "get_prediction_results_model", params, time.Since(start), len(data), false, err)
	}()

	if modelID == "" {
		err = errors.New("model ID is required")
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

	data, total, err = s.repo.GetPredictionResultsByModel(ctx, modelID, orgID, limit, offset)
	return data, total, err
}

// GetPredictionResultsByOrg retrieves prediction results by organization
func (s *AnalyticsServiceImpl) GetPredictionResultsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictionResult
	var total int

	defer func() {
		params := map[string]interface{}{"limit": limit, "offset": offset}
		s.logQuery(ctx, orgID, "get_prediction_results_org", params, time.Since(start), len(data), false, err)
	}()

	if orgID == uuid.Nil {
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

	data, total, err = s.repo.GetPredictionResultsByOrg(ctx, orgID, limit, offset)
	return data, total, err
}

// UpdatePredictionResult updates an existing prediction result with validation
func (s *AnalyticsServiceImpl) UpdatePredictionResult(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictionResult(ctx, id, orgID, updates)
}

// DeletePredictionResult deletes a prediction result
func (s *AnalyticsServiceImpl) DeletePredictionResult(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictionResult(ctx, id, orgID)
}

// ListPredictionResults retrieves prediction results with filtering
func (s *AnalyticsServiceImpl) ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, cursor string, limit, offset int) ([]*models.PredictionResult, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictionResult
	var total int

	defer func() {
		params := map[string]interface{}{"limit": limit, "offset": offset}
		if modelID != nil {
			params["model_id"] = *modelID
		}
		if anomalyDetected != nil {
			params["anomaly_detected"] = *anomalyDetected
		}
		s.logQuery(ctx, orgID, "list_prediction_results", params, time.Since(start), len(data), false, err)
	}()

	if orgID == uuid.Nil {
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

	data, total, err = s.repo.ListPredictionResults(ctx, orgID, modelID, anomalyDetected, cursor, limit, offset)
	return data, total, err
}

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
func (s *AnalyticsServiceImpl) GetAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) (*models.AnalyticsJob, error) {
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
		s.logQuery(ctx, orgID, "get_analytics_job", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	if s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "analytics_job", params)
		if exists, _ := s.cache.Get(ctx, cacheKey, &res); exists {
			cacheHit = true
			return res, nil
		}

		res, err = s.repo.GetAnalyticsJob(ctx, id, orgID)
		if err == nil && res != nil {
			params := map[string]interface{}{"id": id}
			cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "analytics_job", params)
			_ = s.cache.Set(ctx, cacheKey, res, 0)
		}
		return res, err
	}

	res, err = s.repo.GetAnalyticsJob(ctx, id, orgID)
	return res, err
}

// UpdateAnalyticsJob updates an existing analytics job with validation
func (s *AnalyticsServiceImpl) UpdateAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	err := s.repo.UpdateAnalyticsJob(ctx, id, orgID, updates)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "analytics_job", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// DeleteAnalyticsJob deletes an analytics job
func (s *AnalyticsServiceImpl) DeleteAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	err := s.repo.DeleteAnalyticsJob(ctx, id, orgID)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "analytics_job", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// ListAnalyticsJobs retrieves analytics jobs with filtering
func (s *AnalyticsServiceImpl) ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, cursor string, limit, offset int) ([]*models.AnalyticsJob, int, error) {
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
		s.logQuery(ctx, orgID, "list_analytics_jobs", params, time.Since(start), len(data), cacheHit, err)
	}()

	if orgID == uuid.Nil {
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
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "list_analytics_jobs", params)
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

		data, total, err = s.repo.ListAnalyticsJobs(ctx, jobType, status, orgID, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListAnalyticsJobs(ctx, jobType, status, orgID, cursor, limit, offset)
	return data, total, err
}

// CreateFleetAnalytics creates new fleet analytics with validation
func (s *AnalyticsServiceImpl) CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error {
	if err := validateFleetAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Calculate KPIs if context/signals allow, otherwise use provided
	if analytics.AverageFuelEfficiency == 0 && analytics.TotalDistance > 0 && analytics.TotalFuelConsumption > 0 {
		analytics.AverageFuelEfficiency = s.calculateFuelEfficiency(analytics.TotalDistance, analytics.TotalFuelConsumption)
	}

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreateFleetAnalytics(ctx, analytics)
}

// GetFleetAnalytics retrieves fleet analytics by ID
func (s *AnalyticsServiceImpl) GetFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.FleetAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.FleetAnalytics

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"id": id}
		s.logQuery(ctx, orgID, "get_fleet_analytics", params, time.Since(start), rowCount, false, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetFleetAnalytics(ctx, id, orgID)
	return res, err
}

// GetFleetAnalyticsByPeriod retrieves fleet analytics by organization and date period
func (s *AnalyticsServiceImpl) GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.FleetAnalytics
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		}
		s.logQuery(ctx, orgID, "get_fleet_analytics_period", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, err
	}
	if startDate.IsZero() || endDate.IsZero() {
		err = errors.New("start and end dates are required")
		return nil, err
	}
	if startDate.After(endDate) {
		err = errors.New("start date cannot be after end date")
		return nil, err
	}

	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "fleet_analytics_period", params)

		if found, _ := s.cache.Get(ctx, cacheKey, &res); found {
			cacheHit = true
			return res, nil
		}

		res, err = s.repo.GetFleetAnalyticsByPeriod(ctx, orgID, startDate, endDate)
		if err == nil && res != nil {
			_ = s.cache.Set(ctx, cacheKey, res, 0)
		}
		return res, err
	}

	res, err = s.repo.GetFleetAnalyticsByPeriod(ctx, orgID, startDate, endDate)
	return res, err
}

// UpdateFleetAnalytics updates existing fleet analytics
func (s *AnalyticsServiceImpl) UpdateFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateFleetAnalytics(ctx, id, orgID, updates)
}

// DeleteFleetAnalytics deletes fleet analytics
func (s *AnalyticsServiceImpl) DeleteFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteFleetAnalytics(ctx, id, orgID)
}

// ListFleetAnalytics retrieves fleet analytics with filtering, validation and caching
func (s *AnalyticsServiceImpl) ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, cursor string, limit, offset int) ([]*models.FleetAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.FleetAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if startDate != nil {
			params["start_date"] = startDate.Format(time.RFC3339)
		}
		if endDate != nil {
			params["end_date"] = endDate.Format(time.RFC3339)
		}
		s.logQuery(ctx, orgID, "list_fleet_analytics", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
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

	if startDate != nil && endDate != nil {
		if err = validateQueryPeriod(*startDate, *endDate); err != nil {
			return nil, 0, err
		}
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if startDate != nil {
			params["start_date"] = startDate.Format(time.RFC3339)
		}
		if endDate != nil {
			params["end_date"] = endDate.Format(time.RFC3339)
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "list_fleet_analytics", params)
		type listResult struct {
			Data  []*models.FleetAnalytics
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListFleetAnalytics(ctx, orgID, startDate, endDate, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListFleetAnalytics(ctx, orgID, startDate, endDate, cursor, limit, offset)
	return data, total, err
}

// CreateDriverBehaviorAnalytics creates new driver behavior analytics with validation
func (s *AnalyticsServiceImpl) CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error {
	if err := validateDriverBehaviorAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Apply Safety Scoring Model from Specification
	harshEvents := analytics.HardBrakingCount + analytics.AggressiveAccelerationCount + analytics.SharpTurnCount
	score, risk := s.calculateSafetyScore(analytics.OverSpeedingCount, harshEvents, analytics.FatigueEventCount, analytics.ViolationCount)
	analytics.SafetyScore = score
	analytics.RiskLevel = risk

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreateDriverBehaviorAnalytics(ctx, analytics)
}

// GetDriverBehaviorAnalytics retrieves driver behavior analytics by ID
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.DriverBehaviorAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.DriverBehaviorAnalytics
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{
			"id": id,
		}
		s.logQuery(ctx, orgID, "get_driver_behavior_analytics", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetDriverBehaviorAnalytics(ctx, id, orgID)
	return res, err
}

// GetDriverBehaviorAnalyticsByDriver retrieves driver behavior analytics by driver ID
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.DriverBehaviorAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"driver_id": driverID.String(),
			"limit":     limit,
			"offset":    offset,
		}
		s.logQuery(ctx, orgID, "get_driver_behavior_by_driver", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if driverID == uuid.Nil {
		err = errors.New("driver ID is required")
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

	data, total, err = s.repo.GetDriverBehaviorAnalyticsByDriver(ctx, driverID, orgID, limit, offset)
	return data, total, err
}

// GetDriverBehaviorAnalyticsByOrg retrieves driver behavior analytics by organization
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalyticsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.DriverBehaviorAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		s.logQuery(ctx, orgID, "get_driver_behavior_by_org", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
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

	data, total, err = s.repo.GetDriverBehaviorAnalyticsByOrg(ctx, orgID, limit, offset)
	return data, total, err
}

// UpdateDriverBehaviorAnalytics updates existing driver behavior analytics
func (s *AnalyticsServiceImpl) UpdateDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateDriverBehaviorAnalytics(ctx, id, orgID, updates)
}

// DeleteDriverBehaviorAnalytics deletes driver behavior analytics
func (s *AnalyticsServiceImpl) DeleteDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteDriverBehaviorAnalytics(ctx, id, orgID)
}

// ListDriverBehaviorAnalytics retrieves driver behavior analytics with filtering, validation and caching
func (s *AnalyticsServiceImpl) ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, cursor string, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.DriverBehaviorAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if periodStart != nil {
			params["period_start"] = periodStart.Format(time.RFC3339)
		}
		if periodEnd != nil {
			params["period_end"] = periodEnd.Format(time.RFC3339)
		}
		s.logQuery(ctx, orgID, "list_driver_behavior", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
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

	if periodStart != nil && periodEnd != nil {
		if err = validateQueryPeriod(*periodStart, *periodEnd); err != nil {
			return nil, 0, err
		}
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if periodStart != nil {
			params["period_start"] = periodStart.Format(time.RFC3339)
		}
		if periodEnd != nil {
			params["period_end"] = periodEnd.Format(time.RFC3339)
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "list_driver_behavior", params)
		type listResult struct {
			Data  []*models.DriverBehaviorAnalytics
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListDriverBehaviorAnalytics(ctx, orgID, periodStart, periodEnd, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListDriverBehaviorAnalytics(ctx, orgID, periodStart, periodEnd, cursor, limit, offset)
	return data, total, err
}

// CreatePredictiveMaintenanceAnalytics creates new predictive maintenance analytics with validation
func (s *AnalyticsServiceImpl) CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error {
	if err := validatePredictiveMaintenanceAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreatePredictiveMaintenanceAnalytics(ctx, analytics)
}

// GetPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics by ID
func (s *AnalyticsServiceImpl) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveMaintenanceAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.PredictiveMaintenanceAnalytics
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{
			"id": id,
		}
		s.logQuery(ctx, orgID, "get_predictive_maintenance_analytics", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetPredictiveMaintenanceAnalytics(ctx, id, orgID)
	return res, err
}

// GetPredictiveMaintenanceAnalyticsByVehicle retrieves predictive maintenance analytics by vehicle ID
func (s *AnalyticsServiceImpl) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictiveMaintenanceAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"vehicle_id": vehicleID.String(),
			"limit":      limit,
			"offset":     offset,
		}
		s.logQuery(ctx, orgID, "get_predictive_maintenance_by_vehicle", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if vehicleID == uuid.Nil {
		err = errors.New("vehicle ID is required")
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

	data, total, err = s.repo.GetPredictiveMaintenanceAnalyticsByVehicle(ctx, vehicleID, orgID, limit, offset)
	return data, total, err
}

// UpdatePredictiveMaintenanceAnalytics updates existing predictive maintenance analytics
func (s *AnalyticsServiceImpl) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictiveMaintenanceAnalytics(ctx, id, orgID, updates)
}

// DeletePredictiveMaintenanceAnalytics deletes predictive maintenance analytics
func (s *AnalyticsServiceImpl) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictiveMaintenanceAnalytics(ctx, id, orgID)
}

// ListPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics with filtering, validation and caching
func (s *AnalyticsServiceImpl) ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, cursor string, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictiveMaintenanceAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if vehicleID != nil {
			params["vehicle_id"] = vehicleID.String()
		}
		s.logQuery(ctx, orgID, "list_predictive_maintenance", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
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
		if vehicleID != nil {
			params["vehicle_id"] = vehicleID.String()
		}
		var cacheKey string
		if vehicleID != nil {
			cacheKey = s.cache.GenerateKey(ctx, orgID.String(), "vehicle", vehicleID.String(), "list_predictive_maintenance", params)
		} else {
			cacheKey = s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "list_predictive_maintenance", params)
		}
		type listResult struct {
			Data  []*models.PredictiveMaintenanceAnalytics
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListPredictiveMaintenanceAnalytics(ctx, orgID, vehicleID, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListPredictiveMaintenanceAnalytics(ctx, orgID, vehicleID, cursor, limit, offset)
	return data, total, err
}

// HealthCheck performs a health check on the service and its dependencies
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
func (s *AnalyticsServiceImpl) GetVehicleDailyStats(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.VehicleDailyStats, string, error) {
	start := time.Now()
	var err error
	var res []*models.VehicleDailyStats
	var nextCursor string
	cacheHit := false
	rowCount := 0

	defer func() {
		params := map[string]interface{}{
			"vehicle_id": vehicleID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		s.logQuery(ctx, orgID, "get_vehicle_daily_stats", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if vehicleID == uuid.Nil {
		err = errors.New("vehicle ID is required")
		return nil, "", err
	}
	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"vehicle_id": vehicleID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "vehicle", vehicleID.String(), "vehicle_daily_stats", params)
		
		type cachedResult struct {
			Data   []*models.VehicleDailyStats `json:"data"`
			Cursor string                      `json:"cursor"`
		}
		var cached cachedResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cached); found {
			cacheHit = true
			res = cached.Data
			nextCursor = cached.Cursor
			rowCount = len(res)
			return res, nextCursor, nil
		}

		res, nextCursor, err = s.repo.GetVehicleDailyStats(ctx, vehicleID, orgID, startDate, endDate, cursor, limit)
		if err == nil {
			rowCount = len(res)
			_ = s.cache.Set(ctx, cacheKey, cachedResult{Data: res, Cursor: nextCursor}, 0)
		}
		return res, nextCursor, err
	}

	res, nextCursor, err = s.repo.GetVehicleDailyStats(ctx, vehicleID, orgID, startDate, endDate, cursor, limit)
	rowCount = len(res)
	return res, nextCursor, err
}

// GetDriverDailyStats retrieves daily stats for a driver with cursor pagination
func (s *AnalyticsServiceImpl) GetDriverDailyStats(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.DriverDailyStats, string, error) {
	start := time.Now()
	var err error
	var res []*models.DriverDailyStats
	var nextCursor string
	cacheHit := false
	rowCount := 0

	defer func() {
		params := map[string]interface{}{
			"driver_id":  driverID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		s.logQuery(ctx, orgID, "get_driver_daily_stats", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if driverID == uuid.Nil {
		err = errors.New("driver ID is required")
		return nil, "", err
	}
	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"driver_id":  driverID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "driver", driverID.String(), "driver_daily_stats", params)
		
		type cachedResult struct {
			Data   []*models.DriverDailyStats `json:"data"`
			Cursor string                     `json:"cursor"`
		}
		var cached cachedResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cached); found {
			cacheHit = true
			res = cached.Data
			nextCursor = cached.Cursor
			rowCount = len(res)
			return res, nextCursor, nil
		}

		res, nextCursor, err = s.repo.GetDriverDailyStats(ctx, driverID, orgID, startDate, endDate, cursor, limit)
		if err == nil {
			rowCount = len(res)
			_ = s.cache.Set(ctx, cacheKey, cachedResult{Data: res, Cursor: nextCursor}, 0)
		}
		return res, nextCursor, err
	}

	res, nextCursor, err = s.repo.GetDriverDailyStats(ctx, driverID, orgID, startDate, endDate, cursor, limit)
	rowCount = len(res)
	return res, nextCursor, err
}

// GetFleetHourlyStats retrieves pre-aggregated hourly stats for the fleet
func (s *AnalyticsServiceImpl) GetFleetHourlyStats(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.FleetHourlyStats, string, error) {
	start := time.Now()
	var err error
	var res []*models.FleetHourlyStats
	var nextCursor string
	cacheHit := false
	rowCount := 0

	defer func() {
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		s.logQuery(ctx, orgID, "get_fleet_hourly_stats", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, "", err
	}
	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "fleet_hourly_stats", params)
		
		type cachedResult struct {
			Data   []*models.FleetHourlyStats `json:"data"`
			Cursor string                     `json:"cursor"`
		}
		var cached cachedResult
		
		if found, _ := s.cache.Get(ctx, cacheKey, &cached); found {
			cacheHit = true
			res = cached.Data
			nextCursor = cached.Cursor
			rowCount = len(res)
			return res, nextCursor, nil
		}

		res, nextCursor, err = s.repo.GetFleetHourlyStats(ctx, orgID, startDate, endDate, cursor, limit)
		if err == nil {
			rowCount = len(res)
			_ = s.cache.Set(ctx, cacheKey, cachedResult{Data: res, Cursor: nextCursor}, 0)
		}
		return res, nextCursor, err
	}

	res, nextCursor, err = s.repo.GetFleetHourlyStats(ctx, orgID, startDate, endDate, cursor, limit)
	rowCount = len(res)
	return res, nextCursor, err
}

// InvalidateVehicleCache invalidates all analytics cache related to a vehicle
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

	s.logger.Info("Enqueued vehicle analytics cache invalidation", zap.String("org_id", orgID.String()), zap.String("vehicle_id", vehicleID.String()))
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
		{"vehicle_daily_stats", 365 * 2 * 24 * time.Hour},  // 2 years
		{"driver_daily_stats", 365 * 3 * 24 * time.Hour},   // 3 years
		{"prediction_results", 365 * 24 * time.Hour},      // 1 year
		{"analytics_query_logs", 30 * 24 * time.Hour},      // 30 days
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

	s.logger.Info("Enqueued driver analytics cache invalidation", zap.String("org_id", orgID.String()), zap.String("driver_id", driverID.String()))
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

	s.logger.Info("Incremented fleet analytics cache version", zap.String("org_id", orgID.String()))
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
	if drivingHours + idleHours == 0 {
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
		s.logger.Warn("could not retrieve quota", zap.Error(err), zap.String("org_id", orgID.String()))
		return true, nil 
	}
	
	if quota.UsedCredits + credits > quota.DailyLimit {
		return false, nil
	}
	return true, nil
}

// RunReconciliation ensures consistency between telemetry and aggregates
func (s *AnalyticsServiceImpl) RunReconciliation(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) error {
	s.logger.Info("starting analytics reconciliation", zap.String("org_id", orgID.String()), zap.Time("start", startDate), zap.Time("end", endDate))
	// Placeholder: In a real-world scenario, this would query raw telemetry vs daily stats and trigger repairs
	return nil
}

// RunHistoricalBackfill recomputes historical stats
func (s *AnalyticsServiceImpl) RunHistoricalBackfill(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) error {
	s.logger.Info("starting historical analytics backfill", zap.String("org_id", orgID.String()), zap.Time("start", startDate), zap.Time("end", endDate))
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
		if baseLimit < 5 { baseLimit = 5 }
		
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
			if finalLimit < 2 { finalLimit = 2 } // Floor limit
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

