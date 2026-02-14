package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsRepositoryImpl implements the AnalyticsRepository interface
type AnalyticsRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewAnalyticsRepository creates a new instance of AnalyticsRepositoryImpl
func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepositoryImpl {
	return &AnalyticsRepositoryImpl{
		db: db,
	}
}

// CreatePredictiveModel creates a new predictive model
func (r *AnalyticsRepositoryImpl) CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error {
	query := `
		INSERT INTO predictive_models (id, name, description, model_type, algorithm, input_features, output_variable, 
			training_data_start, training_data_end, accuracy, precision, recall, f1_score, 
			status, version, trained_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`
	
	inputFeaturesJSON, err := json.Marshal(model.InputFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal input features: %w", err)
	}
	
	// Handle nullable fields
	var precision *float64
	if model.Precision != nil {
		precision = model.Precision
	}
	
	var recall *float64
	if model.Recall != nil {
		recall = model.Recall
	}
	
	var f1Score *float64
	if model.F1Score != nil {
		f1Score = model.F1Score
	}
	
	_, err = r.db.Exec(ctx, query, model.ID, model.Name, model.Description, model.ModelType, model.Algorithm,
		inputFeaturesJSON, model.OutputVariable, model.TrainingDataStart, model.TrainingDataEnd,
		model.Accuracy, precision, recall, f1Score, model.Status, model.Version, 
		model.TrainedAt, model.ExpiresAt, model.CreatedAt, model.UpdatedAt)
	
	return err
}

// GetPredictiveModel retrieves a predictive model by ID
func (r *AnalyticsRepositoryImpl) GetPredictiveModel(ctx context.Context, id string) (*models.PredictiveModel, error) {
	query := `
		SELECT id, name, description, model_type, algorithm, input_features, output_variable, 
			training_data_start, training_data_end, accuracy, precision, recall, f1_score, 
			status, version, trained_at, expires_at, created_at, updated_at
		FROM predictive_models WHERE id = $1
	`
	
	var model models.PredictiveModel
	var inputFeaturesJSON []byte
	var precision, recall, f1Score *float64
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&model.ID, &model.Name, &model.Description, &model.ModelType, &model.Algorithm,
		&inputFeaturesJSON, &model.OutputVariable, &model.TrainingDataStart, &model.TrainingDataEnd,
		&model.Accuracy, &precision, &recall, &f1Score, &model.Status, &model.Version,
		&model.TrainedAt, &model.ExpiresAt, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(inputFeaturesJSON, &model.InputFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal input features: %w", err)
	}
	
	// Set nullable fields
	model.Precision = precision
	model.Recall = recall
	model.F1Score = f1Score
	
	return &model, nil
}

// GetPredictiveModelByVersion retrieves a predictive model by name and version
func (r *AnalyticsRepositoryImpl) GetPredictiveModelByVersion(ctx context.Context, name, version string) (*models.PredictiveModel, error) {
	query := `
		SELECT id, name, description, model_type, algorithm, input_features, output_variable, 
			training_data_start, training_data_end, accuracy, precision, recall, f1_score, 
			status, version, trained_at, expires_at, created_at, updated_at
		FROM predictive_models WHERE name = $1 AND version = $2
	`
	
	var model models.PredictiveModel
	var inputFeaturesJSON []byte
	var precision, recall, f1Score *float64
	
	err := r.db.QueryRow(ctx, query, name, version).Scan(
		&model.ID, &model.Name, &model.Description, &model.ModelType, &model.Algorithm,
		&inputFeaturesJSON, &model.OutputVariable, &model.TrainingDataStart, &model.TrainingDataEnd,
		&model.Accuracy, &precision, &recall, &f1Score, &model.Status, &model.Version,
		&model.TrainedAt, &model.ExpiresAt, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(inputFeaturesJSON, &model.InputFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal input features: %w", err)
	}
	
	// Set nullable fields
	model.Precision = precision
	model.Recall = recall
	model.F1Score = f1Score
	
	return &model, nil
}

// UpdatePredictiveModel updates an existing predictive model
func (r *AnalyticsRepositoryImpl) UpdatePredictiveModel(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE predictive_models SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeletePredictiveModel deletes a predictive model
func (r *AnalyticsRepositoryImpl) DeletePredictiveModel(ctx context.Context, id string) error {
	query := "DELETE FROM predictive_models WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListPredictiveModels retrieves predictive models with filtering
func (r *AnalyticsRepositoryImpl) ListPredictiveModels(ctx context.Context, modelType *string, status *string, limit, offset int) ([]*models.PredictiveModel, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM predictive_models WHERE TRUE"
	countArgs := []interface{}{}
	
	// Filter conditions
	whereConditions := []string{}
	argIndex := 1
	
	if modelType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("model_type = $%d", argIndex))
		countArgs = append(countArgs, *modelType)
		argIndex++
	}
	
	if status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		countArgs = append(countArgs, *status)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 0 {
		finalCountQuery += " AND " + fmt.Sprintf("%s", whereConditions)
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := "SELECT id, name, description, model_type, algorithm, input_features, output_variable, " +
		"training_data_start, training_data_end, accuracy, precision, recall, f1_score, " +
		"status, version, trained_at, expires_at, created_at, updated_at " +
		"FROM predictive_models WHERE TRUE"
	
	args := []interface{}{}
	argIndex = 1
	
	whereConditions = []string{} // Reset for main query
	if modelType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("model_type = $%d", argIndex))
		args = append(args, *modelType)
		argIndex++
	}
	
	if status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}
	
	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + fmt.Sprintf("%s", whereConditions)
	}
	
	finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var modelsList []*models.PredictiveModel
	for rows.Next() {
		var model models.PredictiveModel
		var inputFeaturesJSON []byte
		var precision, recall, f1Score *float64
		
		err := rows.Scan(
			&model.ID, &model.Name, &model.Description, &model.ModelType, &model.Algorithm,
			&inputFeaturesJSON, &model.OutputVariable, &model.TrainingDataStart, &model.TrainingDataEnd,
			&model.Accuracy, &precision, &recall, &f1Score, &model.Status, &model.Version,
			&model.TrainedAt, &model.ExpiresAt, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(inputFeaturesJSON, &model.InputFeatures)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal input features: %w", err)
		}
		
		// Set nullable fields
		model.Precision = precision
		model.Recall = recall
		model.F1Score = f1Score
		
		modelsList = append(modelsList, &model)
	}
	
	return modelsList, totalCount, nil
}

// CreatePredictionResult creates a new prediction result
func (r *AnalyticsRepositoryImpl) CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error {
	query := `
		INSERT INTO prediction_results (id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	
	predictionInputJSON, err := json.Marshal(result.PredictionInput)
	if err != nil {
		return fmt.Errorf("failed to marshal prediction input: %w", err)
	}
	
	probabilityScoresJSON, err := json.Marshal(result.ProbabilityScores)
	if err != nil {
		return fmt.Errorf("failed to marshal probability scores: %w", err)
	}
	
	featureImportanceJSON, err := json.Marshal(result.FeatureImportance)
	if err != nil {
		return fmt.Errorf("failed to marshal feature importance: %w", err)
	}
	
	// Handle nullable fields
	var anomalyScore *float64
	if result.AnomalyScore != nil {
		anomalyScore = result.AnomalyScore
	}
	
	_, err = r.db.Exec(ctx, query, result.ID, result.ModelID, predictionInputJSON, result.PredictedValue,
		result.ConfidenceScore, probabilityScoresJSON, featureImportanceJSON, result.AnomalyDetected,
		anomalyScore, result.Explanation, result.OrganizationID, result.RequestID, result.PredictedAt,
		result.CreatedAt, result.UpdatedAt)
	
	return err
}

// GetPredictionResult retrieves a prediction result by ID
func (r *AnalyticsRepositoryImpl) GetPredictionResult(ctx context.Context, id string) (*models.PredictionResult, error) {
	query := `
		SELECT id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results WHERE id = $1
	`
	
	var result models.PredictionResult
	var predictionInputJSON, probabilityScoresJSON, featureImportanceJSON []byte
	var anomalyScore *float64
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.ModelID, &predictionInputJSON, &result.PredictedValue, &result.ConfidenceScore,
		&probabilityScoresJSON, &featureImportanceJSON, &result.AnomalyDetected, &anomalyScore,
		&result.Explanation, &result.OrganizationID, &result.RequestID, &result.PredictedAt,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(predictionInputJSON, &result.PredictionInput)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal prediction input: %w", err)
	}
	
	err = json.Unmarshal(probabilityScoresJSON, &result.ProbabilityScores)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal probability scores: %w", err)
	}
	
	err = json.Unmarshal(featureImportanceJSON, &result.FeatureImportance)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature importance: %w", err)
	}
	
	result.AnomalyScore = anomalyScore
	
	return &result, nil
}

// GetPredictionResultsByModel retrieves prediction results by model ID
func (r *AnalyticsRepositoryImpl) GetPredictionResultsByModel(ctx context.Context, modelID string, limit, offset int) ([]*models.PredictionResult, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM prediction_results WHERE model_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, modelID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	query := `
		SELECT id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results 
		WHERE model_id = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, modelID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var results []*models.PredictionResult
	for rows.Next() {
		var result models.PredictionResult
		var predictionInputJSON, probabilityScoresJSON, featureImportanceJSON []byte
		var anomalyScore *float64
		
		err := rows.Scan(
			&result.ID, &result.ModelID, &predictionInputJSON, &result.PredictedValue, &result.ConfidenceScore,
			&probabilityScoresJSON, &featureImportanceJSON, &result.AnomalyDetected, &anomalyScore,
			&result.Explanation, &result.OrganizationID, &result.RequestID, &result.PredictedAt,
			&result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(predictionInputJSON, &result.PredictionInput)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal prediction input: %w", err)
		}
		
		err = json.Unmarshal(probabilityScoresJSON, &result.ProbabilityScores)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal probability scores: %w", err)
		}
		
		err = json.Unmarshal(featureImportanceJSON, &result.FeatureImportance)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal feature importance: %w", err)
		}
		
		result.AnomalyScore = anomalyScore
		
		results = append(results, &result)
	}
	
	return results, totalCount, nil
}

// GetPredictionResultsByOrg retrieves prediction results by organization
func (r *AnalyticsRepositoryImpl) GetPredictionResultsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM prediction_results WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	query := `
		SELECT id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results 
		WHERE organization_id = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var results []*models.PredictionResult
	for rows.Next() {
		var result models.PredictionResult
		var predictionInputJSON, probabilityScoresJSON, featureImportanceJSON []byte
		var anomalyScore *float64
		
		err := rows.Scan(
			&result.ID, &result.ModelID, &predictionInputJSON, &result.PredictedValue, &result.ConfidenceScore,
			&probabilityScoresJSON, &featureImportanceJSON, &result.AnomalyDetected, &anomalyScore,
			&result.Explanation, &result.OrganizationID, &result.RequestID, &result.PredictedAt,
			&result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(predictionInputJSON, &result.PredictionInput)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal prediction input: %w", err)
		}
		
		err = json.Unmarshal(probabilityScoresJSON, &result.ProbabilityScores)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal probability scores: %w", err)
		}
		
		err = json.Unmarshal(featureImportanceJSON, &result.FeatureImportance)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal feature importance: %w", err)
		}
		
		result.AnomalyScore = anomalyScore
		
		results = append(results, &result)
	}
	
	return results, totalCount, nil
}

// UpdatePredictionResult updates an existing prediction result
func (r *AnalyticsRepositoryImpl) UpdatePredictionResult(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE prediction_results SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeletePredictionResult deletes a prediction result
func (r *AnalyticsRepositoryImpl) DeletePredictionResult(ctx context.Context, id string) error {
	query := "DELETE FROM prediction_results WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListPredictionResults retrieves prediction results with filtering
func (r *AnalyticsRepositoryImpl) ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, limit, offset int) ([]*models.PredictionResult, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM prediction_results WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
	if modelID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("model_id = $%d", argIndex))
		countArgs = append(countArgs, *modelID)
		argIndex++
	}
	
	if anomalyDetected != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("anomaly_detected = $%d", argIndex))
		countArgs = append(countArgs, *anomalyDetected)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + fmt.Sprintf("%s", whereConditions[1:])
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := `
		SELECT id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results 
		WHERE organization_id = $1
	`
	
	args := []interface{}{orgID}
	argIndex = 2
	
	whereConditions = []string{} // Reset for main query
	if modelID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("model_id = $%d", argIndex))
		args = append(args, *modelID)
		argIndex++
	}
	
	if anomalyDetected != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("anomaly_detected = $%d", argIndex))
		args = append(args, *anomalyDetected)
		argIndex++
	}
	
	finalMainQuery := mainQuery
	if len(whereConditions) > 0 { // Skip first condition since it's in base query
		finalMainQuery += " AND " + fmt.Sprintf("%s", whereConditions)
	}
	
	finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var results []*models.PredictionResult
	for rows.Next() {
		var result models.PredictionResult
		var predictionInputJSON, probabilityScoresJSON, featureImportanceJSON []byte
		var anomalyScore *float64
		
		err := rows.Scan(
			&result.ID, &result.ModelID, &predictionInputJSON, &result.PredictedValue, &result.ConfidenceScore,
			&probabilityScoresJSON, &featureImportanceJSON, &result.AnomalyDetected, &anomalyScore,
			&result.Explanation, &result.OrganizationID, &result.RequestID, &result.PredictedAt,
			&result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(predictionInputJSON, &result.PredictionInput)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal prediction input: %w", err)
		}
		
		err = json.Unmarshal(probabilityScoresJSON, &result.ProbabilityScores)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal probability scores: %w", err)
		}
		
		err = json.Unmarshal(featureImportanceJSON, &result.FeatureImportance)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal feature importance: %w", err)
		}
		
		result.AnomalyScore = anomalyScore
		
		results = append(results, &result)
	}
	
	return results, totalCount, nil
}

// CreateAnalyticsJob creates a new analytics job
func (r *AnalyticsRepositoryImpl) CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error {
	query := `
		INSERT INTO analytics_jobs (id, name, type, description, status, job_type, parameters, 
			input_source, output_destination, start_time, end_time, progress, error_details, 
			organization_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
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
		errorDetails, job.OrganizationID, job.CreatedBy, job.CreatedAt, job.UpdatedAt)
	
	return err
}

// GetAnalyticsJob retrieves an analytics job by ID
func (r *AnalyticsRepositoryImpl) GetAnalyticsJob(ctx context.Context, id string) (*models.AnalyticsJob, error) {
	query := `
		SELECT id, name, type, description, status, job_type, parameters, 
			input_source, output_destination, start_time, end_time, progress, error_details, 
			organization_id, created_by, created_at, updated_at
		FROM analytics_jobs WHERE id = $1
	`
	
	var job models.AnalyticsJob
	var parametersJSON []byte
	var startTime, endTime *time.Time
	var errorDetails *string
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.Name, &job.Type, &job.Description, &job.Status, &job.JobType, &parametersJSON,
		&job.InputSource, &job.OutputDestination, &startTime, &endTime, &job.Progress, &errorDetails,
		&job.OrganizationID, &job.CreatedBy, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
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
func (r *AnalyticsRepositoryImpl) UpdateAnalyticsJob(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE analytics_jobs SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteAnalyticsJob deletes an analytics job
func (r *AnalyticsRepositoryImpl) DeleteAnalyticsJob(ctx context.Context, id string) error {
	query := "DELETE FROM analytics_jobs WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListAnalyticsJobs retrieves analytics jobs with filtering
func (r *AnalyticsRepositoryImpl) ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, limit, offset int) ([]*models.AnalyticsJob, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM analytics_jobs WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
	if jobType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("job_type = $%d", argIndex))
		countArgs = append(countArgs, *jobType)
		argIndex++
	}
	
	if status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		countArgs = append(countArgs, *status)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + strings.Join(whereConditions[1:], " AND ")
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := `
		SELECT id, name, type, description, status, job_type, parameters, 
			input_source, output_destination, start_time, end_time, progress, error_details, 
			organization_id, created_by, created_at, updated_at
		FROM analytics_jobs 
		WHERE organization_id = $1
	`
	
	args := []interface{}{orgID}
	argIndex = 2
	
	whereConditions = []string{} // Reset for main query
	if jobType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("job_type = $%d", argIndex))
		args = append(args, *jobType)
		argIndex++
	}
	
	if status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}
	
	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var jobs []*models.AnalyticsJob
	for rows.Next() {
		var job models.AnalyticsJob
		var parametersJSON []byte
		var startTime, endTime *time.Time
		var errorDetails *string
		
		err := rows.Scan(
			&job.ID, &job.Name, &job.Type, &job.Description, &job.Status, &job.JobType, &parametersJSON,
			&job.InputSource, &job.OutputDestination, &startTime, &endTime, &job.Progress, &errorDetails,
			&job.OrganizationID, &job.CreatedBy, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(parametersJSON, &job.Parameters)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}
		
		job.StartTime = startTime
		job.EndTime = endTime
		job.ErrorDetails = errorDetails
		
		jobs = append(jobs, &job)
	}
	
	return jobs, totalCount, nil
}

// CreateFleetAnalytics creates new fleet analytics
func (r *AnalyticsRepositoryImpl) CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error {
	query := `
		INSERT INTO fleet_analytics (id, organization_id, start_date, end_date, 
			total_vehicles, active_vehicles, total_distance, average_speed, total_fuel_consumption, 
			average_fuel_efficiency, total_trips, completed_trips, cancelled_trips, 
			average_trip_duration, idle_time_percentage, routing_efficiency, fuel_cost, 
			maintenance_cost, revenue, profit_margin, safety_score, accidents_count, 
			incidents_count, emergency_stops, geofence_violations, alerts_triggered, 
			data_points_collected, processing_time, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31)
	`
	
	_, err := r.db.Exec(ctx, query, analytics.ID, analytics.OrganizationID, analytics.StartDate,
		analytics.EndDate, analytics.TotalVehicles, analytics.ActiveVehicles, analytics.TotalDistance,
		analytics.AverageSpeed, analytics.TotalFuelConsumption, analytics.AverageFuelEfficiency,
		analytics.TotalTrips, analytics.CompletedTrips, analytics.CancelledTrips,
		analytics.AverageTripDuration, analytics.IdleTimePercentage, analytics.RoutingEfficiency,
		analytics.FuelCost, analytics.MaintenanceCost, analytics.Revenue, analytics.ProfitMargin,
		analytics.SafetyScore, analytics.AccidentsCount, analytics.IncidentsCount,
		analytics.EmergencyStops, analytics.GeofenceViolations, analytics.AlertsTriggered,
		analytics.DataPointsCollected, analytics.ProcessingTime, analytics.GeneratedAt,
		analytics.CreatedAt, analytics.UpdatedAt)
	
	return err
}

// GetFleetAnalytics retrieves fleet analytics by ID
func (r *AnalyticsRepositoryImpl) GetFleetAnalytics(ctx context.Context, id string) (*models.FleetAnalytics, error) {
	query := `
		SELECT id, organization_id, start_date, end_date, 
			total_vehicles, active_vehicles, total_distance, average_speed, total_fuel_consumption, 
			average_fuel_efficiency, total_trips, completed_trips, cancelled_trips, 
			average_trip_duration, idle_time_percentage, routing_efficiency, fuel_cost, 
			maintenance_cost, revenue, profit_margin, safety_score, accidents_count, 
			incidents_count, emergency_stops, geofence_violations, alerts_triggered, 
			data_points_collected, processing_time, generated_at, created_at, updated_at
		FROM fleet_analytics WHERE id = $1
	`
	
	var analytics models.FleetAnalytics
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&analytics.ID, &analytics.OrganizationID, &analytics.StartDate, &analytics.EndDate,
		&analytics.TotalVehicles, &analytics.ActiveVehicles, &analytics.TotalDistance,
		&analytics.AverageSpeed, &analytics.TotalFuelConsumption, &analytics.AverageFuelEfficiency,
		&analytics.TotalTrips, &analytics.CompletedTrips, &analytics.CancelledTrips,
		&analytics.AverageTripDuration, &analytics.IdleTimePercentage, &analytics.RoutingEfficiency,
		&analytics.FuelCost, &analytics.MaintenanceCost, &analytics.Revenue, &analytics.ProfitMargin,
		&analytics.SafetyScore, &analytics.AccidentsCount, &analytics.IncidentsCount,
		&analytics.EmergencyStops, &analytics.GeofenceViolations, &analytics.AlertsTriggered,
		&analytics.DataPointsCollected, &analytics.ProcessingTime, &analytics.GeneratedAt,
		&analytics.CreatedAt, &analytics.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	return &analytics, nil
}

// GetFleetAnalyticsByPeriod retrieves fleet analytics by organization and date period
func (r *AnalyticsRepositoryImpl) GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
	query := `
		SELECT id, organization_id, start_date, end_date, 
			total_vehicles, active_vehicles, total_distance, average_speed, total_fuel_consumption, 
			average_fuel_efficiency, total_trips, completed_trips, cancelled_trips, 
			average_trip_duration, idle_time_percentage, routing_efficiency, fuel_cost, 
			maintenance_cost, revenue, profit_margin, safety_score, accidents_count, 
			incidents_count, emergency_stops, geofence_violations, alerts_triggered, 
			data_points_collected, processing_time, generated_at, created_at, updated_at
		FROM fleet_analytics 
		WHERE organization_id = $1 AND start_date >= $2 AND end_date <= $3
	`
	
	var analytics models.FleetAnalytics
	
	err := r.db.QueryRow(ctx, query, orgID, startDate, endDate).Scan(
		&analytics.ID, &analytics.OrganizationID, &analytics.StartDate, &analytics.EndDate,
		&analytics.TotalVehicles, &analytics.ActiveVehicles, &analytics.TotalDistance,
		&analytics.AverageSpeed, &analytics.TotalFuelConsumption, &analytics.AverageFuelEfficiency,
		&analytics.TotalTrips, &analytics.CompletedTrips, &analytics.CancelledTrips,
		&analytics.AverageTripDuration, &analytics.IdleTimePercentage, &analytics.RoutingEfficiency,
		&analytics.FuelCost, &analytics.MaintenanceCost, &analytics.Revenue, &analytics.ProfitMargin,
		&analytics.SafetyScore, &analytics.AccidentsCount, &analytics.IncidentsCount,
		&analytics.EmergencyStops, &analytics.GeofenceViolations, &analytics.AlertsTriggered,
		&analytics.DataPointsCollected, &analytics.ProcessingTime, &analytics.GeneratedAt,
		&analytics.CreatedAt, &analytics.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	return &analytics, nil
}

// UpdateFleetAnalytics updates existing fleet analytics
func (r *AnalyticsRepositoryImpl) UpdateFleetAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE fleet_analytics SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteFleetAnalytics deletes fleet analytics
func (r *AnalyticsRepositoryImpl) DeleteFleetAnalytics(ctx context.Context, id string) error {
	query := "DELETE FROM fleet_analytics WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListFleetAnalytics retrieves fleet analytics with filtering
func (r *AnalyticsRepositoryImpl) ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*models.FleetAnalytics, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM fleet_analytics WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
	if startDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("start_date >= $%d", argIndex))
		countArgs = append(countArgs, *startDate)
		argIndex++
	}
	
	if endDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("end_date <= $%d", argIndex))
		countArgs = append(countArgs, *endDate)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + strings.Join(whereConditions[1:], " AND ")
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := `
		SELECT id, organization_id, start_date, end_date, 
			total_vehicles, active_vehicles, total_distance, average_speed, total_fuel_consumption, 
			average_fuel_efficiency, total_trips, completed_trips, cancelled_trips, 
			average_trip_duration, idle_time_percentage, routing_efficiency, fuel_cost, 
			maintenance_cost, revenue, profit_margin, safety_score, accidents_count, 
			incidents_count, emergency_stops, geofence_violations, alerts_triggered, 
			data_points_collected, processing_time, generated_at, created_at, updated_at
		FROM fleet_analytics 
		WHERE organization_id = $1
	`
	
	args := []interface{}{orgID}
	argIndex = 2
	
	whereConditions = []string{} // Reset for main query
	if startDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("start_date >= $%d", argIndex))
		args = append(args, *startDate)
		argIndex++
	}
	
	if endDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("end_date <= $%d", argIndex))
		args = append(args, *endDate)
		argIndex++
	}
	
	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	finalMainQuery += fmt.Sprintf(" ORDER BY start_date DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var analyticsList []*models.FleetAnalytics
	for rows.Next() {
		var analytics models.FleetAnalytics
		
		err := rows.Scan(
			&analytics.ID, &analytics.OrganizationID, &analytics.StartDate, &analytics.EndDate,
			&analytics.TotalVehicles, &analytics.ActiveVehicles, &analytics.TotalDistance,
			&analytics.AverageSpeed, &analytics.TotalFuelConsumption, &analytics.AverageFuelEfficiency,
			&analytics.TotalTrips, &analytics.CompletedTrips, &analytics.CancelledTrips,
			&analytics.AverageTripDuration, &analytics.IdleTimePercentage, &analytics.RoutingEfficiency,
			&analytics.FuelCost, &analytics.MaintenanceCost, &analytics.Revenue, &analytics.ProfitMargin,
			&analytics.SafetyScore, &analytics.AccidentsCount, &analytics.IncidentsCount,
			&analytics.EmergencyStops, &analytics.GeofenceViolations, &analytics.AlertsTriggered,
			&analytics.DataPointsCollected, &analytics.ProcessingTime, &analytics.GeneratedAt,
			&analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// CreateDriverBehaviorAnalytics creates new driver behavior analytics
func (r *AnalyticsRepositoryImpl) CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error {
	query := `
		INSERT INTO driver_behavior_analytics (id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 
			$18, $19, $20, $21, $22, $23)
	`
	
	improvementSuggestionsJSON, err := json.Marshal(analytics.ImprovementSuggestions)
	if err != nil {
		return fmt.Errorf("failed to marshal improvement suggestions: %w", err)
	}
	
	_, err = r.db.Exec(ctx, query, analytics.ID, analytics.DriverID, analytics.OrganizationID,
		analytics.AnalysisPeriodStart, analytics.AnalysisPeriodEnd, analytics.TotalDrivingHours,
		analytics.AggressiveAccelerationCount, analytics.HardBrakingCount, analytics.SharpTurnCount,
		analytics.OverSpeedingCount, analytics.IdlingTime, analytics.AverageSpeed,
		analytics.MaxSpeed, analytics.FuelEfficiency, analytics.SafetyScore, analytics.EfficiencyScore,
		analytics.ComfortScore, analytics.Ranking, improvementSuggestionsJSON,
		analytics.BehaviorTrend, analytics.GeneratedAt, analytics.CreatedAt, analytics.UpdatedAt)
	
	return err
}

// GetDriverBehaviorAnalytics retrieves driver behavior analytics by ID
func (r *AnalyticsRepositoryImpl) GetDriverBehaviorAnalytics(ctx context.Context, id string) (*models.DriverBehaviorAnalytics, error) {
	query := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics WHERE id = $1
	`
	
	var analytics models.DriverBehaviorAnalytics
	var improvementSuggestionsJSON []byte
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&analytics.ID, &analytics.DriverID, &analytics.OrganizationID, &analytics.AnalysisPeriodStart,
		&analytics.AnalysisPeriodEnd, &analytics.TotalDrivingHours, &analytics.AggressiveAccelerationCount,
		&analytics.HardBrakingCount, &analytics.SharpTurnCount, &analytics.OverSpeedingCount,
		&analytics.IdlingTime, &analytics.AverageSpeed, &analytics.MaxSpeed, &analytics.FuelEfficiency,
		&analytics.SafetyScore, &analytics.EfficiencyScore, &analytics.ComfortScore, &analytics.Ranking,
		&improvementSuggestionsJSON, &analytics.BehaviorTrend, &analytics.GeneratedAt,
		&analytics.CreatedAt, &analytics.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(improvementSuggestionsJSON, &analytics.ImprovementSuggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal improvement suggestions: %w", err)
	}
	
	return &analytics, nil
}

// GetDriverBehaviorAnalyticsByDriver retrieves driver behavior analytics by driver ID
func (r *AnalyticsRepositoryImpl) GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM driver_behavior_analytics WHERE driver_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, driverID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	query := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics 
		WHERE driver_id = $1
		ORDER BY analysis_period_start DESC 
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, driverID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var analyticsList []*models.DriverBehaviorAnalytics
	for rows.Next() {
		var analytics models.DriverBehaviorAnalytics
		var improvementSuggestionsJSON []byte
		
		err := rows.Scan(
			&analytics.ID, &analytics.DriverID, &analytics.OrganizationID, &analytics.AnalysisPeriodStart,
			&analytics.AnalysisPeriodEnd, &analytics.TotalDrivingHours, &analytics.AggressiveAccelerationCount,
			&analytics.HardBrakingCount, &analytics.SharpTurnCount, &analytics.OverSpeedingCount,
			&analytics.IdlingTime, &analytics.AverageSpeed, &analytics.MaxSpeed, &analytics.FuelEfficiency,
			&analytics.SafetyScore, &analytics.EfficiencyScore, &analytics.ComfortScore, &analytics.Ranking,
			&improvementSuggestionsJSON, &analytics.BehaviorTrend, &analytics.GeneratedAt,
			&analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(improvementSuggestionsJSON, &analytics.ImprovementSuggestions)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal improvement suggestions: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// GetDriverBehaviorAnalyticsByOrg retrieves driver behavior analytics by organization
func (r *AnalyticsRepositoryImpl) GetDriverBehaviorAnalyticsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM driver_behavior_analytics WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	query := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics 
		WHERE organization_id = $1
		ORDER BY analysis_period_start DESC 
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var analyticsList []*models.DriverBehaviorAnalytics
	for rows.Next() {
		var analytics models.DriverBehaviorAnalytics
		var improvementSuggestionsJSON []byte
		
		err := rows.Scan(
			&analytics.ID, &analytics.DriverID, &analytics.OrganizationID, &analytics.AnalysisPeriodStart,
			&analytics.AnalysisPeriodEnd, &analytics.TotalDrivingHours, &analytics.AggressiveAccelerationCount,
			&analytics.HardBrakingCount, &analytics.SharpTurnCount, &analytics.OverSpeedingCount,
			&analytics.IdlingTime, &analytics.AverageSpeed, &analytics.MaxSpeed, &analytics.FuelEfficiency,
			&analytics.SafetyScore, &analytics.EfficiencyScore, &analytics.ComfortScore, &analytics.Ranking,
			&improvementSuggestionsJSON, &analytics.BehaviorTrend, &analytics.GeneratedAt,
			&analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(improvementSuggestionsJSON, &analytics.ImprovementSuggestions)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal improvement suggestions: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// UpdateDriverBehaviorAnalytics updates existing driver behavior analytics
func (r *AnalyticsRepositoryImpl) UpdateDriverBehaviorAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE driver_behavior_analytics SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteDriverBehaviorAnalytics deletes driver behavior analytics
func (r *AnalyticsRepositoryImpl) DeleteDriverBehaviorAnalytics(ctx context.Context, id string) error {
	query := "DELETE FROM driver_behavior_analytics WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListDriverBehaviorAnalytics retrieves driver behavior analytics with filtering
func (r *AnalyticsRepositoryImpl) ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM driver_behavior_analytics WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
	if periodStart != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_start >= $%d", argIndex))
		countArgs = append(countArgs, *periodStart)
		argIndex++
	}
	
	if periodEnd != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_end <= $%d", argIndex))
		countArgs = append(countArgs, *periodEnd)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + strings.Join(whereConditions[1:], " AND ")
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics 
		WHERE organization_id = $1
	`
	
	args := []interface{}{orgID}
	argIndex = 2
	
	whereConditions = []string{} // Reset for main query
	if periodStart != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_start >= $%d", argIndex))
		args = append(args, *periodStart)
		argIndex++
	}
	
	if periodEnd != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_end <= $%d", argIndex))
		args = append(args, *periodEnd)
		argIndex++
	}
	
	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	finalMainQuery += fmt.Sprintf(" ORDER BY analysis_period_start DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var analyticsList []*models.DriverBehaviorAnalytics
	for rows.Next() {
		var analytics models.DriverBehaviorAnalytics
		var improvementSuggestionsJSON []byte
		
		err := rows.Scan(
			&analytics.ID, &analytics.DriverID, &analytics.OrganizationID, &analytics.AnalysisPeriodStart,
			&analytics.AnalysisPeriodEnd, &analytics.TotalDrivingHours, &analytics.AggressiveAccelerationCount,
			&analytics.HardBrakingCount, &analytics.SharpTurnCount, &analytics.OverSpeedingCount,
			&analytics.IdlingTime, &analytics.AverageSpeed, &analytics.MaxSpeed, &analytics.FuelEfficiency,
			&analytics.SafetyScore, &analytics.EfficiencyScore, &analytics.ComfortScore, &analytics.Ranking,
			&improvementSuggestionsJSON, &analytics.BehaviorTrend, &analytics.GeneratedAt,
			&analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(improvementSuggestionsJSON, &analytics.ImprovementSuggestions)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal improvement suggestions: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// CreatePredictiveMaintenanceAnalytics creates new predictive maintenance analytics
func (r *AnalyticsRepositoryImpl) CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error {
	query := `
		INSERT INTO predictive_maintenance_analytics (id, vehicle_id, organization_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	
	componentRisksJSON, err := json.Marshal(analytics.ComponentRisks)
	if err != nil {
		return fmt.Errorf("failed to marshal component risks: %w", err)
	}
	
	var failurePredictionJSON []byte
	if analytics.FailurePrediction != nil {
		failurePredictionJSON, err = json.Marshal(analytics.FailurePrediction)
		if err != nil {
			return fmt.Errorf("failed to marshal failure prediction: %w", err)
		}
	}
	
	maintenanceScheduleJSON, err := json.Marshal(analytics.MaintenanceSchedule)
	if err != nil {
		return fmt.Errorf("failed to marshal maintenance schedule: %w", err)
	}
	
	partsNeededJSON, err := json.Marshal(analytics.PartsNeeded)
	if err != nil {
		return fmt.Errorf("failed to marshal parts needed: %w", err)
	}
	
	recommendationsJSON, err := json.Marshal(analytics.Recommendations)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendations: %w", err)
	}
	
	_, err = r.db.Exec(ctx, query, analytics.ID, analytics.VehicleID, analytics.OrganizationID,
		analytics.AnalysisDate, analytics.NextServiceDue, analytics.ServicePriority,
		componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON,
		analytics.CostEstimation, analytics.UptimePrediction, analytics.ReliabilityScore,
		recommendationsJSON, analytics.GeneratedAt, analytics.CreatedAt, analytics.UpdatedAt)
	
	return err
}

// GetPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics by ID
func (r *AnalyticsRepositoryImpl) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string) (*models.PredictiveMaintenanceAnalytics, error) {
	query := `
		SELECT id, vehicle_id, organization_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics WHERE id = $1
	`
	
	var analytics models.PredictiveMaintenanceAnalytics
	var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&analytics.ID, &analytics.VehicleID, &analytics.OrganizationID, &analytics.AnalysisDate,
		&analytics.NextServiceDue, &analytics.ServicePriority, &componentRisksJSON,
		&failurePredictionJSON, &maintenanceScheduleJSON, &partsNeededJSON, &analytics.CostEstimation,
		&analytics.UptimePrediction, &analytics.ReliabilityScore, &recommendationsJSON,
		&analytics.GeneratedAt, &analytics.CreatedAt, &analytics.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(componentRisksJSON, &analytics.ComponentRisks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal component risks: %w", err)
	}
	
	if failurePredictionJSON != nil {
		var failurePrediction models.FailurePrediction
		err = json.Unmarshal(failurePredictionJSON, &failurePrediction)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal failure prediction: %w", err)
		}
		analytics.FailurePrediction = &failurePrediction
	}
	
	err = json.Unmarshal(maintenanceScheduleJSON, &analytics.MaintenanceSchedule)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal maintenance schedule: %w", err)
	}
	
	err = json.Unmarshal(partsNeededJSON, &analytics.PartsNeeded)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal parts needed: %w", err)
	}
	
	err = json.Unmarshal(recommendationsJSON, &analytics.Recommendations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal recommendations: %w", err)
	}
	
	return &analytics, nil
}

// GetPredictiveMaintenanceAnalyticsByVehicle retrieves predictive maintenance analytics by vehicle ID
func (r *AnalyticsRepositoryImpl) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM predictive_maintenance_analytics WHERE vehicle_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, vehicleID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	query := `
		SELECT id, vehicle_id, organization_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics 
		WHERE vehicle_id = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var analyticsList []*models.PredictiveMaintenanceAnalytics
	for rows.Next() {
		var analytics models.PredictiveMaintenanceAnalytics
		var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte
		
		err := rows.Scan(
			&analytics.ID, &analytics.VehicleID, &analytics.OrganizationID, &analytics.AnalysisDate,
			&analytics.NextServiceDue, &analytics.ServicePriority, &componentRisksJSON,
			&failurePredictionJSON, &maintenanceScheduleJSON, &partsNeededJSON, &analytics.CostEstimation,
			&analytics.UptimePrediction, &analytics.ReliabilityScore, &recommendationsJSON,
			&analytics.GeneratedAt, &analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(componentRisksJSON, &analytics.ComponentRisks)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal component risks: %w", err)
		}
		
		if failurePredictionJSON != nil {
			var failurePrediction models.FailurePrediction
			err = json.Unmarshal(failurePredictionJSON, &failurePrediction)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal failure prediction: %w", err)
			}
			analytics.FailurePrediction = &failurePrediction
		}
		
		err = json.Unmarshal(maintenanceScheduleJSON, &analytics.MaintenanceSchedule)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal maintenance schedule: %w", err)
		}
		
		err = json.Unmarshal(partsNeededJSON, &analytics.PartsNeeded)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal parts needed: %w", err)
		}
		
		err = json.Unmarshal(recommendationsJSON, &analytics.Recommendations)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal recommendations: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// UpdatePredictiveMaintenanceAnalytics updates existing predictive maintenance analytics
func (r *AnalyticsRepositoryImpl) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return errors.New("no updates provided")
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE predictive_maintenance_analytics SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeletePredictiveMaintenanceAnalytics deletes predictive maintenance analytics
func (r *AnalyticsRepositoryImpl) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string) error {
	query := "DELETE FROM predictive_maintenance_analytics WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ListPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics with filtering
func (r *AnalyticsRepositoryImpl) ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM predictive_maintenance_analytics WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
	if vehicleID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		countArgs = append(countArgs, *vehicleID)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + fmt.Sprintf("%s", whereConditions[1:])
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := `
		SELECT id, vehicle_id, organization_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics 
		WHERE organization_id = $1
	`
	
	args := []interface{}{orgID}
	argIndex = 2
	
	whereConditions = []string{} // Reset for main query
	if vehicleID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, *vehicleID)
		argIndex++
	}
	
	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + fmt.Sprintf("%s", whereConditions)
	}
	
	finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var analyticsList []*models.PredictiveMaintenanceAnalytics
	for rows.Next() {
		var analytics models.PredictiveMaintenanceAnalytics
		var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte
		
		err := rows.Scan(
			&analytics.ID, &analytics.VehicleID, &analytics.OrganizationID, &analytics.AnalysisDate,
			&analytics.NextServiceDue, &analytics.ServicePriority, &componentRisksJSON,
			&failurePredictionJSON, &maintenanceScheduleJSON, &partsNeededJSON, &analytics.CostEstimation,
			&analytics.UptimePrediction, &analytics.ReliabilityScore, &recommendationsJSON,
			&analytics.GeneratedAt, &analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		
		err = json.Unmarshal(componentRisksJSON, &analytics.ComponentRisks)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal component risks: %w", err)
		}
		
		if failurePredictionJSON != nil {
			var failurePrediction models.FailurePrediction
			err = json.Unmarshal(failurePredictionJSON, &failurePrediction)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal failure prediction: %w", err)
			}
			analytics.FailurePrediction = &failurePrediction
		}
		
		err = json.Unmarshal(maintenanceScheduleJSON, &analytics.MaintenanceSchedule)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal maintenance schedule: %w", err)
		}
		
		err = json.Unmarshal(partsNeededJSON, &analytics.PartsNeeded)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal parts needed: %w", err)
		}
		
		err = json.Unmarshal(recommendationsJSON, &analytics.Recommendations)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal recommendations: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// HealthCheck performs a health check on the database
func (r *AnalyticsRepositoryImpl) HealthCheck(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *AnalyticsRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
