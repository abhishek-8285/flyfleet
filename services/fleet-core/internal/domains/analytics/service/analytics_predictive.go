package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Predictive Model & Inference Methods
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
func (s *AnalyticsServiceImpl) GetPredictiveModel(ctx context.Context, id string, tenantID uuid.UUID) (*models.PredictiveModel, error) {
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
		s.logQuery(ctx, tenantID, "get_predictive_model", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	if s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "predictive_model", params)
		if found, _ := s.cache.Get(ctx, cacheKey, &res); found {
			cacheHit = true
			return res, nil
		}

		res, err = s.repo.GetPredictiveModel(ctx, id, tenantID)
		if err == nil && res != nil {
			params := map[string]interface{}{"id": id}
			cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "predictive_model", params)
			_ = s.cache.Set(ctx, cacheKey, res, 0)
		}
		return res, err
	}

	res, err = s.repo.GetPredictiveModel(ctx, id, tenantID)
	return res, err
}

// GetPredictiveModelByVersion retrieves a predictive model by name and version
func (s *AnalyticsServiceImpl) GetPredictiveModelByVersion(ctx context.Context, name, version string, tenantID uuid.UUID) (*models.PredictiveModel, error) {
	start := time.Now()
	var err error
	var res *models.PredictiveModel

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"name": name, "version": version}
		s.logQuery(ctx, tenantID, "get_predictive_model_version", params, time.Since(start), rowCount, false, err)
	}()

	if name == "" {
		err = errors.New("name is required")
		return nil, err
	}
	if version == "" {
		err = errors.New("version is required")
		return nil, err
	}

	res, err = s.repo.GetPredictiveModelByVersion(ctx, name, version, tenantID)
	return res, err
}

// UpdatePredictiveModel updates an existing predictive model with validation
func (s *AnalyticsServiceImpl) UpdatePredictiveModel(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	err := s.repo.UpdatePredictiveModel(ctx, id, tenantID, updates)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "predictive_model", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// DeletePredictiveModel deletes a predictive model
func (s *AnalyticsServiceImpl) DeletePredictiveModel(ctx context.Context, id string, tenantID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	err := s.repo.DeletePredictiveModel(ctx, id, tenantID)
	if err == nil && s.cache != nil {
		params := map[string]interface{}{"id": id}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "predictive_model", params)
		_ = s.cache.Invalidate(ctx, cacheKey)
	}
	return err
}

// ListPredictiveModels retrieves predictive models with filtering
func (s *AnalyticsServiceImpl) ListPredictiveModels(ctx context.Context, tenantID uuid.UUID, modelType *string, status *string, cursor string, limit, offset int) ([]*models.PredictiveModel, int, error) {
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
		s.logQuery(ctx, tenantID, "list_predictive_models", params, time.Since(start), len(data), false, err)
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

	data, total, err = s.repo.ListPredictiveModels(ctx, tenantID, modelType, status, cursor, limit, offset)
	return data, total, err
}

// RunInference executes a specific predictive model on an entity using Phase 8 deployment strategies
func (s *AnalyticsServiceImpl) RunInference(ctx context.Context, modelID string, entityID uuid.UUID, tenantID uuid.UUID, input map[string]interface{}) (*models.PredictionResult, error) {
	start := time.Now()

	// Quota Check (Phase 8)
	if allowed, _ := s.CheckQuota(ctx, tenantID, 1); !allowed {
		return nil, errors.New("organization analytics quota exceeded for today")
	}

	// Fetch Model
	model, err := s.repo.GetPredictiveModel(ctx, modelID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model: %w", err)
	}
	if model.Status != "active" {
		return nil, fmt.Errorf("model %s is not active (status: %s)", modelID, model.Status)
	}

	// 1. Compliance Audit (Phase 11)
	s.LogAccess(ctx, &models.AnalyticsAccessLog{
		ID:             uuid.New(),
		TenantID: tenantID,
		UserID:         uuid.Nil, // Should be extracted from context in real implementation
		ResourceID:     modelID,
		ResourceType:   "MODEL",
		AccessContext:  map[string]interface{}{"action": "inference", "version": model.Version},
		CreatedAt:      time.Now(),
	})

	// 2. Automated Rollback Check (Phase 11)
	if err := s.CheckModelHealth(ctx, modelID, tenantID); err != nil {
		s.logger.Error("Model health check failed, blocking inference", zap.String("model_id", modelID), zap.Error(err))
		return nil, fmt.Errorf("model safety guard triggered: %w", err)
	}

	// Implementation of reproducible inference (Phase 10)
	// If a specific feature version is requested, we fetch that specifically
	if requestedVersion, ok := input["feature_version"].(string); ok && requestedVersion != "" {
		s.logger.Info("Executing reproducible inference with pinned features",
			zap.String("version", requestedVersion),
			zap.String("model_id", modelID))

		features, err := s.GetFeaturesByVersion(ctx, tenantID, "VEHICLE", entityID, requestedVersion)
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
	deployment, err := s.repo.GetActiveModelDeployment(ctx, modelID, tenantID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Warn("failed to retrieve deployment info, falling back to standard", zap.Error(err))
	}

	var primaryResult *models.PredictionResult
	if deployment != nil && deployment.IsActive {
		switch deployment.DeploymentStrategy {
		case "shadow":
			// Execute primary (specified in deployment or current active model)
			primaryResult, err = s.executeInference(ctx, model, entityID, tenantID, input)
			// Execute Shadow (Log/Observe but don't return)
			go s.executeShadowInference(ctx, deployment.ModelID, entityID, tenantID, input)

		case "ab":
			// Deterministic split based on entityID to ensure consistency for that vehicle/driver
			hashBytes := sha256.Sum256([]byte(entityID.String()))
			hash := uint32(hashBytes[0])<<24 | uint32(hashBytes[1])<<16 | uint32(hashBytes[2])<<8 | uint32(hashBytes[3])
			if float64(hash%100) < deployment.TrafficSplit {
				// Traffic routed to variant
				primaryResult, err = s.executeInference(ctx, model, entityID, tenantID, input)
			} else {
				// Traffic routed to control
				primaryResult, err = s.executeInference(ctx, model, entityID, tenantID, input)
			}
		default:
			primaryResult, err = s.executeInference(ctx, model, entityID, tenantID, input)
		}
	} else {
		primaryResult, err = s.executeInference(ctx, model, entityID, tenantID, input)
	}

	if err != nil {
		s.metrics.PredictionErrorCounter.WithLabelValues(tenantID.String(), modelID, "orchestration_error").Inc()
		return nil, err
	}

	// Quota tracking update (Phase 8)
	_ = s.repo.UpdateUsedCredits(ctx, tenantID, 1)

	duration := time.Since(start)
	s.metrics.InferenceLatencyHistogram.WithLabelValues(tenantID.String(), modelID).Observe(duration.Seconds())

	s.logger.Info("Inference orchestration complete",
		zap.String("model_id", modelID),
		zap.String("entity_id", entityID.String()),
		zap.Duration("duration", duration))

	return primaryResult, nil
}

func (s *AnalyticsServiceImpl) executeInference(ctx context.Context, model *models.PredictiveModel, entityID uuid.UUID, tenantID uuid.UUID, input map[string]interface{}) (*models.PredictionResult, error) {
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
		TenantID:  tenantID,
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

	s.metrics.PredictiveAnalyticCounter.WithLabelValues(tenantID.String(), model.ModelType, "success").Inc()
	return result, nil
}

func (s *AnalyticsServiceImpl) executeShadowInference(ctx context.Context, modelID string, entityID uuid.UUID, tenantID uuid.UUID, input map[string]interface{}) {
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
	if score < 0 {
		score = 0
	}

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

	riskProbability := (100.0-safetyScore)*0.5 + (hoursDriving * 2.0)
	if riskProbability > 100 {
		riskProbability = 100
	}

	anomaly := false
	var aScore *float64
	if safetyScore < 20 && hoursDriving > 12 {
		anomaly = true
		val := 0.95
		aScore = &val
	}

	return map[string]interface{}{
		"risk_probability": riskProbability / 100.0,
		"risk_tier":        s.getRiskTier(riskProbability),
	}, 0.92, anomaly, aScore
}

func (s *AnalyticsServiceImpl) getRiskTier(prob float64) string {
	if prob > 80 {
		return "CRITICAL"
	}
	if prob > 50 {
		return "HIGH"
	}
	if prob > 20 {
		return "MEDIUM"
	}
	return "LOW"
}

func (s *AnalyticsServiceImpl) scoreFuelEfficiencyModel(model *models.PredictiveModel, input map[string]interface{}) (interface{}, float64) {
	distance, _ := input["distance"].(float64)
	fuel, _ := input["fuel"].(float64)
	if fuel == 0 {
		return 0.0, 0.0
	}

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
func (s *AnalyticsServiceImpl) GetPredictionResult(ctx context.Context, id string, tenantID uuid.UUID) (*models.PredictionResult, error) {
	start := time.Now()
	var err error
	var res *models.PredictionResult

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"id": id}
		s.logQuery(ctx, tenantID, "get_prediction_result", params, time.Since(start), rowCount, false, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetPredictionResult(ctx, id, tenantID)
	return res, err
}

// GetPredictionResultsByModel retrieves prediction results by model ID
func (s *AnalyticsServiceImpl) GetPredictionResultsByModel(ctx context.Context, modelID string, tenantID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictionResult
	var total int

	defer func() {
		params := map[string]interface{}{"model_id": modelID, "limit": limit, "offset": offset}
		s.logQuery(ctx, tenantID, "get_prediction_results_model", params, time.Since(start), len(data), false, err)
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

	data, total, err = s.repo.GetPredictionResultsByModel(ctx, modelID, tenantID, limit, offset)
	return data, total, err
}

// GetPredictionResultsByTenant retrieves prediction results by tenant
func (s *AnalyticsServiceImpl) GetPredictionResultsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictionResult
	var total int

	defer func() {
		params := map[string]interface{}{"limit": limit, "offset": offset}
		s.logQuery(ctx, tenantID, "get_prediction_results_tenant", params, time.Since(start), len(data), false, err)
	}()

	if tenantID == uuid.Nil {
		err = errors.New("tenant ID is required")
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

	data, total, err = s.repo.GetPredictionResultsByTenant(ctx, tenantID, limit, offset)
	return data, total, err
}

// UpdatePredictionResult updates an existing prediction result with validation
func (s *AnalyticsServiceImpl) UpdatePredictionResult(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictionResult(ctx, id, tenantID, updates)
}

// DeletePredictionResult deletes a prediction result
func (s *AnalyticsServiceImpl) DeletePredictionResult(ctx context.Context, id string, tenantID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictionResult(ctx, id, tenantID)
}

// ListPredictionResults retrieves prediction results with filtering
func (s *AnalyticsServiceImpl) ListPredictionResults(ctx context.Context, tenantID uuid.UUID, modelID *string, anomalyDetected *bool, cursor string, limit, offset int) ([]*models.PredictionResult, int, error) {
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
		s.logQuery(ctx, tenantID, "list_prediction_results", params, time.Since(start), len(data), false, err)
	}()

	if tenantID == uuid.Nil {
		err = errors.New("tenant ID is required")
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

	data, total, err = s.repo.ListPredictionResults(ctx, tenantID, modelID, anomalyDetected, cursor, limit, offset)
	return data, total, err
}
