package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const queryTimeout = 30 * time.Second

// AnalyticsRepositoryImpl implements the AnalyticsRepository interface
type AnalyticsRepositoryImpl struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewAnalyticsRepository creates a new instance of AnalyticsRepositoryImpl
func NewAnalyticsRepository(db *pgxpool.Pool, logger *zap.Logger) *AnalyticsRepositoryImpl {
	return &AnalyticsRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *AnalyticsRepositoryImpl) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, queryTimeout)
}

// CreatePredictiveModel creates a new predictive model
func (r *AnalyticsRepositoryImpl) CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		INSERT INTO predictive_models (id, organization_id, name, description, model_type, algorithm, input_features, output_variable, 
			training_data_start, training_data_end, accuracy, precision, recall, f1_score, 
			status, version, trained_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
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
	
	_, err = r.db.Exec(ctx, query, model.ID, model.OrganizationID, model.Name, model.Description, model.ModelType, model.Algorithm,
		inputFeaturesJSON, model.OutputVariable, model.TrainingDataStart, model.TrainingDataEnd,
		model.Accuracy, precision, recall, f1Score, model.Status, model.Version,
		model.TrainedAt, model.ExpiresAt, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create predictive model: %w", err)
	}
	return nil
}

// GetPredictiveModel retrieves a predictive model by ID
func (r *AnalyticsRepositoryImpl) GetPredictiveModel(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveModel, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, organization_id, name, description, model_type, algorithm, input_features, output_variable,
			training_data_start, training_data_end, accuracy, precision, recall, f1_score,
			status, version, trained_at, expires_at, created_at, updated_at
		FROM predictive_models WHERE id = $1 AND organization_id = $2
	`

	var model models.PredictiveModel
	var inputFeaturesJSON []byte
	var precision, recall, f1Score *float64

	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&model.ID, &model.OrganizationID, &model.Name, &model.Description, &model.ModelType, &model.Algorithm,
		&inputFeaturesJSON, &model.OutputVariable, &model.TrainingDataStart, &model.TrainingDataEnd,
		&model.Accuracy, &precision, &recall, &f1Score, &model.Status, &model.Version,
		&model.TrainedAt, &model.ExpiresAt, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get predictive model: %w", err)
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
func (r *AnalyticsRepositoryImpl) GetPredictiveModelByVersion(ctx context.Context, name, version string, orgID uuid.UUID) (*models.PredictiveModel, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, organization_id, name, description, model_type, algorithm, input_features, output_variable, 
			training_data_start, training_data_end, accuracy, precision, recall, f1_score, 
			status, version, trained_at, expires_at, created_at, updated_at
		FROM predictive_models WHERE name = $1 AND version = $2 AND organization_id = $3
	`
	
	var model models.PredictiveModel
	var inputFeaturesJSON []byte
	var precision, recall, f1Score *float64
	
	err := r.db.QueryRow(ctx, query, name, version, orgID).Scan(
		&model.ID, &model.OrganizationID, &model.Name, &model.Description, &model.ModelType, &model.Algorithm,
		&inputFeaturesJSON, &model.OutputVariable, &model.TrainingDataStart, &model.TrainingDataEnd,
		&model.Accuracy, &precision, &recall, &f1Score, &model.Status, &model.Version,
		&model.TrainedAt, &model.ExpiresAt, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("predictive model not found for name %s version %s: %w", name, version, err)
		}
		return nil, fmt.Errorf("failed to get predictive model by version: %w", err)
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
func (r *AnalyticsRepositoryImpl) UpdatePredictiveModel(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	args = append(args, id, orgID)
	query := fmt.Sprintf("UPDATE predictive_models SET %s WHERE id = $%d AND organization_id = $%d", 
		strings.Join(setParts, ", "), argIndex, argIndex+1)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update predictive model %s: %w", id, err)
	}
	return nil
}

// DeletePredictiveModel deletes a predictive model
func (r *AnalyticsRepositoryImpl) DeletePredictiveModel(ctx context.Context, id string, orgID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM predictive_models WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete predictive model %s: %w", id, err)
	}
	return nil
}

// ListPredictiveModels retrieves predictive models with filtering
func (r *AnalyticsRepositoryImpl) ListPredictiveModels(ctx context.Context, orgID uuid.UUID, modelType *string, status *string, cursor string, limit, offset int) ([]*models.PredictiveModel, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	// Count query
	countQuery := "SELECT COUNT(*) FROM predictive_models WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
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
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + strings.Join(whereConditions[1:], " AND ")
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	
	// Main query
	mainQuery := "SELECT id, organization_id, name, description, model_type, algorithm, input_features, output_variable, " +
		"training_data_start, training_data_end, accuracy, precision, recall, f1_score, " +
		"status, version, trained_at, expires_at, created_at, updated_at " +
		"FROM predictive_models WHERE organization_id = $1"
	
	args := []interface{}{orgID}
	argIndex = 2
	
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
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
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
			&model.ID, &model.OrganizationID, &model.Name, &model.Description, &model.ModelType, &model.Algorithm,
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

// CreatePredictionResults bulk inserts multiple prediction results
func (r *AnalyticsRepositoryImpl) CreatePredictionResults(ctx context.Context, results []*models.PredictionResult) error {
	if len(results) == 0 {
		return nil
	}
	
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	// Use a transaction for bulk insert
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO prediction_results (id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	for _, result := range results {
		inputJSON, _ := json.Marshal(result.PredictionInput)
		probJSON, _ := json.Marshal(result.ProbabilityScores)
		featJSON, _ := json.Marshal(result.FeatureImportance)
		
		_, err = tx.Exec(ctx, query, result.ID, result.ModelID, inputJSON, result.PredictedValue,
			result.ConfidenceScore, probJSON, featJSON, result.AnomalyDetected,
			result.AnomalyScore, result.Explanation, result.OrganizationID, result.RequestID, result.PredictedAt,
			result.CreatedAt, result.UpdatedAt)
		
		if err != nil {
			return fmt.Errorf("failed to insert result %s: %w", result.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreatePredictionResult creates a new prediction result
func (r *AnalyticsRepositoryImpl) CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	if err != nil {
		return fmt.Errorf("failed to create prediction result: %w", err)
	}
	return nil
}

// GetPredictionResult retrieves a prediction result by ID
func (r *AnalyticsRepositoryImpl) GetPredictionResult(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictionResult, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results WHERE id = $1 AND organization_id = $2
	`
	
	var result models.PredictionResult
	var predictionInputJSON, probabilityScoresJSON, featureImportanceJSON []byte
	var anomalyScore *float64
	
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&result.ID, &result.ModelID, &predictionInputJSON, &result.PredictedValue, &result.ConfidenceScore,
		&probabilityScoresJSON, &featureImportanceJSON, &result.AnomalyDetected, &anomalyScore,
		&result.Explanation, &result.OrganizationID, &result.RequestID, &result.PredictedAt,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("prediction result %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get prediction result: %w", err)
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
func (r *AnalyticsRepositoryImpl) GetPredictionResultsByModel(ctx context.Context, modelID string, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM prediction_results WHERE model_id = $1 AND organization_id = $2"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, modelID, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count prediction results by model: %w", err)
	}
	
	// Main query
	query := `
		SELECT id, model_id, prediction_input, predicted_value, confidence_score, 
			probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, 
			organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results 
		WHERE model_id = $1 AND organization_id = $2
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.executeWithObservability(ctx, query, modelID, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query prediction results by model: %w", err)
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
	
	rows, err := r.executeWithObservability(ctx, query, orgID, limit, offset)
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
func (r *AnalyticsRepositoryImpl) UpdatePredictionResult(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	args = append(args, id, orgID)
	query := fmt.Sprintf("UPDATE prediction_results SET %s WHERE id = $%d AND organization_id = $%d", 
		strings.Join(setParts, ", "), argIndex, argIndex+1)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update prediction result %s: %w", id, err)
	}
	return nil
}

// DeletePredictionResult deletes a prediction result
func (r *AnalyticsRepositoryImpl) DeletePredictionResult(ctx context.Context, id string, orgID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM prediction_results WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete prediction result %s: %w", id, err)
	}
	return nil
}

// ListPredictionResults retrieves prediction results with filtering
func (r *AnalyticsRepositoryImpl) ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, cursor string, limit, offset int) ([]*models.PredictionResult, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
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
func (r *AnalyticsRepositoryImpl) ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, cursor string, limit, offset int) ([]*models.AnalyticsJob, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
			result_key, organization_id, created_by, created_at, updated_at
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
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
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
			&job.ResultKey, &job.OrganizationID, &job.CreatedBy, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan analytics job row: %w", err)
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
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	if err != nil {
		return fmt.Errorf("failed to create fleet analytics: %w", err)
	}
	return nil
}

// GetFleetAnalytics retrieves fleet analytics by ID
func (r *AnalyticsRepositoryImpl) GetFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.FleetAnalytics, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, organization_id, start_date, end_date, 
			total_vehicles, active_vehicles, total_distance, average_speed, total_fuel_consumption, 
			average_fuel_efficiency, total_trips, completed_trips, cancelled_trips, 
			average_trip_duration, idle_time_percentage, routing_efficiency, fuel_cost, 
			maintenance_cost, revenue, profit_margin, safety_score, accidents_count, 
			incidents_count, emergency_stops, geofence_violations, alerts_triggered, 
			data_points_collected, processing_time, generated_at, created_at, updated_at
		FROM fleet_analytics WHERE id = $1 AND organization_id = $2
	`
	
	var analytics models.FleetAnalytics
	
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("fleet analytics %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get fleet analytics: %w", err)
	}
	
	return &analytics, nil
}

// GetFleetAnalyticsByPeriod retrieves fleet analytics by organization and date period
func (r *AnalyticsRepositoryImpl) GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("fleet analytics not found for period: %w", err)
		}
		return nil, fmt.Errorf("failed to get fleet analytics by period: %w", err)
	}
	
	return &analytics, nil
}

// UpdateFleetAnalytics updates existing fleet analytics
func (r *AnalyticsRepositoryImpl) UpdateFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	args = append(args, id, orgID)
	query := fmt.Sprintf("UPDATE fleet_analytics SET %s WHERE id = $%d AND organization_id = $%d", 
		strings.Join(setParts, ", "), argIndex, argIndex+1)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update fleet analytics %s: %w", id, err)
	}
	return nil
}

// DeleteFleetAnalytics deletes fleet analytics
func (r *AnalyticsRepositoryImpl) DeleteFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM fleet_analytics WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete fleet analytics %s: %w", id, err)
	}
	return nil
}

// ListFleetAnalytics retrieves fleet analytics with filtering
func (r *AnalyticsRepositoryImpl) ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, cursor string, limit, offset int) ([]*models.FleetAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
		return nil, 0, fmt.Errorf("failed to count fleet analytics: %w", err)
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
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(start_date, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY start_date DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY start_date DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query fleet analytics: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan fleet analytics row: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// CreateDriverBehaviorAnalytics creates new driver behavior analytics
func (r *AnalyticsRepositoryImpl) CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	if err != nil {
		return fmt.Errorf("failed to create driver behavior analytics: %w", err)
	}
	return nil
}

// GetDriverBehaviorAnalytics retrieves driver behavior analytics by ID
func (r *AnalyticsRepositoryImpl) GetDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.DriverBehaviorAnalytics, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics WHERE id = $1 AND organization_id = $2
	`
	
	var analytics models.DriverBehaviorAnalytics
	var improvementSuggestionsJSON []byte
	
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&analytics.ID, &analytics.DriverID, &analytics.OrganizationID, &analytics.AnalysisPeriodStart,
		&analytics.AnalysisPeriodEnd, &analytics.TotalDrivingHours, &analytics.AggressiveAccelerationCount,
		&analytics.HardBrakingCount, &analytics.SharpTurnCount, &analytics.OverSpeedingCount,
		&analytics.IdlingTime, &analytics.AverageSpeed, &analytics.MaxSpeed, &analytics.FuelEfficiency,
		&analytics.SafetyScore, &analytics.EfficiencyScore, &analytics.ComfortScore, &analytics.Ranking,
		&improvementSuggestionsJSON, &analytics.BehaviorTrend, &analytics.GeneratedAt,
		&analytics.CreatedAt, &analytics.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("driver behavior analytics %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get driver behavior analytics: %w", err)
	}
	
	err = json.Unmarshal(improvementSuggestionsJSON, &analytics.ImprovementSuggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal improvement suggestions: %w", err)
	}
	
	return &analytics, nil
}

// GetDriverBehaviorAnalyticsByDriver retrieves driver behavior analytics by driver ID
func (r *AnalyticsRepositoryImpl) GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	// Count query
	countQuery := "SELECT COUNT(*) FROM driver_behavior_analytics WHERE driver_id = $1 AND organization_id = $2"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, driverID, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count driver behavior analytics: %w", err)
	}
	
	// Main query
	query := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics 
		WHERE driver_id = $1 AND organization_id = $2
		ORDER BY analysis_period_start DESC 
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.executeWithObservability(ctx, query, driverID, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query driver behavior analytics: %w", err)
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
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	// Count query
	countQuery := "SELECT COUNT(*) FROM driver_behavior_analytics WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count driver behavior analytics by org: %w", err)
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
	
	rows, err := r.executeWithObservability(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query driver behavior analytics by org: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan driver behavior analytics row: %w", err)
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
func (r *AnalyticsRepositoryImpl) UpdateDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	args = append(args, id, orgID)
	query := fmt.Sprintf("UPDATE driver_behavior_analytics SET %s WHERE id = $%d AND organization_id = $%d", 
		strings.Join(setParts, ", "), argIndex, argIndex+1)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update driver behavior analytics %s: %w", id, err)
	}
	return nil
}

// DeleteDriverBehaviorAnalytics deletes driver behavior analytics
func (r *AnalyticsRepositoryImpl) DeleteDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM driver_behavior_analytics WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete driver behavior analytics %s: %w", id, err)
	}
	return nil
}

// ListDriverBehaviorAnalytics retrieves driver behavior analytics with filtering
func (r *AnalyticsRepositoryImpl) ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, cursor string, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
		return nil, 0, fmt.Errorf("failed to count driver behavior analytics: %w", err)
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
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(analysis_period_start, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY analysis_period_start DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY analysis_period_start DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query driver behavior analytics: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan driver behavior analytics row: %w", err)
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
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	if err != nil {
		return fmt.Errorf("failed to create predictive maintenance analytics: %w", err)
	}
	return nil
}

// GetPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics by ID
func (r *AnalyticsRepositoryImpl) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveMaintenanceAnalytics, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, vehicle_id, organization_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics WHERE id = $1 AND organization_id = $2
	`
	
	var analytics models.PredictiveMaintenanceAnalytics
	var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte
	
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&analytics.ID, &analytics.VehicleID, &analytics.OrganizationID, &analytics.AnalysisDate,
		&analytics.NextServiceDue, &analytics.ServicePriority, &componentRisksJSON,
		&failurePredictionJSON, &maintenanceScheduleJSON, &partsNeededJSON, &analytics.CostEstimation,
		&analytics.UptimePrediction, &analytics.ReliabilityScore, &recommendationsJSON,
		&analytics.GeneratedAt, &analytics.CreatedAt, &analytics.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("predictive maintenance analytics %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get predictive maintenance analytics: %w", err)
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
func (r *AnalyticsRepositoryImpl) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	// Count query
	countQuery := "SELECT COUNT(*) FROM predictive_maintenance_analytics WHERE vehicle_id = $1 AND organization_id = $2"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, vehicleID, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count predictive maintenance analytics: %w", err)
	}
	
	// Main query
	query := `
		SELECT id, vehicle_id, organization_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics 
		WHERE vehicle_id = $1 AND organization_id = $2
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4
	`
	
	rows, err := r.executeWithObservability(ctx, query, vehicleID, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query predictive maintenance analytics: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan predictive maintenance analytics row: %w", err)
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
func (r *AnalyticsRepositoryImpl) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
	
	args = append(args, id, orgID)
	query := fmt.Sprintf("UPDATE predictive_maintenance_analytics SET %s WHERE id = $%d AND organization_id = $%d", 
		strings.Join(setParts, ", "), argIndex, argIndex+1)
	
	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update predictive maintenance analytics %s: %w", id, err)
	}
	return nil
}

// DeletePredictiveMaintenanceAnalytics deletes predictive maintenance analytics
func (r *AnalyticsRepositoryImpl) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM predictive_maintenance_analytics WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete predictive maintenance analytics %s: %w", id, err)
	}
	return nil
}

// ListPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics with filtering
func (r *AnalyticsRepositoryImpl) ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, cursor string, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
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
		finalCountQuery += " AND " + strings.Join(whereConditions[1:], " AND ")
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count predictive maintenance analytics: %w", err)
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
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query predictive maintenance analytics: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan predictive maintenance analytics row: %w", err)
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
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *AnalyticsRepositoryImpl) Ping(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	return r.db.Ping(ctx)
}

// GetVehicleDailyStats retrieves pre-aggregated daily stats for a vehicle with cursor pagination
func (r *AnalyticsRepositoryImpl) GetVehicleDailyStats(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.VehicleDailyStats, string, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if limit <= 0 { limit = 30 }
	if limit > 100 { limit = 100 }

	query := `
		SELECT vehicle_id, organization_id, date, total_distance, total_fuel_consumed, 
			total_engine_hours, total_idle_hours, max_speed, avg_speed,
			harsh_braking_count, rapid_acceleration_count, sharp_turn_count, overspeeding_count, updated_at
		FROM vehicle_daily_stats
		WHERE vehicle_id = $1 AND organization_id = $2 AND date >= $3 AND date <= $4
	`
	params := []interface{}{vehicleID, orgID, startDate, endDate}
	
	if cursor != "" {
		query += " AND date > $5"
		params = append(params, cursor)
	}
	
	query += fmt.Sprintf(" ORDER BY date ASC LIMIT %d", limit)

	if err := r.checkQueryCost(ctx, query, params...); err != nil {
		return nil, "", fmt.Errorf("cost guard error: %w", err)
	}

	rows, err := r.executeWithObservability(ctx, query, params...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query vehicle daily stats: %w", err)
	}
	defer rows.Close()

	var stats []*models.VehicleDailyStats
	var lastDate string
	for rows.Next() {
		s := &models.VehicleDailyStats{}
		err := rows.Scan(
			&s.VehicleID, &s.OrganizationID, &s.Date, &s.TotalDistance, &s.TotalFuelConsumed,
			&s.TotalEngineHours, &s.TotalIdleHours, &s.MaxSpeed, &s.AvgSpeed,
			&s.HarshBrakingCount, &s.RapidAccelerationCount, &s.SharpTurnCount, &s.OverspeedingCount, &s.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan vehicle daily stats: %w", err)
		}
		stats = append(stats, s)
		lastDate = s.Date.Format("2006-01-02")
	}

	nextCursor := ""
	if len(stats) == limit {
		nextCursor = lastDate
	}

	return stats, nextCursor, nil
}

// GetDriverDailyStats retrieves pre-aggregated daily stats for a driver with cursor pagination
func (r *AnalyticsRepositoryImpl) GetDriverDailyStats(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.DriverDailyStats, string, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if limit <= 0 { limit = 30 }
	if limit > 100 { limit = 100 }

	query := `
		SELECT driver_id, organization_id, date, total_driving_hours, total_distance, 
			total_idle_hours, safety_score, eco_score, violation_count, updated_at
		FROM driver_daily_stats
		WHERE driver_id = $1 AND organization_id = $2 AND date >= $3 AND date <= $4
	`
	params := []interface{}{driverID, orgID, startDate, endDate}
	
	if cursor != "" {
		query += " AND date > $5"
		params = append(params, cursor)
	}
	
	query += fmt.Sprintf(" ORDER BY date ASC LIMIT %d", limit)

	if err := r.checkQueryCost(ctx, query, params...); err != nil {
		return nil, "", fmt.Errorf("cost guard error: %w", err)
	}

	rows, err := r.executeWithObservability(ctx, query, params...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query driver daily stats: %w", err)
	}
	defer rows.Close()

	var stats []*models.DriverDailyStats
	var lastDate string
	for rows.Next() {
		s := &models.DriverDailyStats{}
		err := rows.Scan(
			&s.DriverID, &s.OrganizationID, &s.Date, &s.TotalDrivingHours, &s.TotalDistance,
			&s.TotalIdleHours, &s.SafetyScore, &s.EcoScore, &s.ViolationCount, &s.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan driver daily stats: %w", err)
		}
		stats = append(stats, s)
		lastDate = s.Date.Format("2006-01-02")
	}

	nextCursor := ""
	if len(stats) == limit {
		nextCursor = lastDate
	}

	return stats, nextCursor, nil
}

// GetFleetHourlyStats retrieves pre-aggregated hourly stats for the whole fleet with cursor pagination
func (r *AnalyticsRepositoryImpl) GetFleetHourlyStats(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.FleetHourlyStats, string, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if limit <= 0 { limit = 24 }
	if limit > 168 { limit = 168 } // Max 1 week hourly at once

	query := `
		SELECT organization_id, timestamp, active_vehicles, total_distance, total_fuel_consumed, 
			average_utilization, updated_at
		FROM fleet_hourly_stats
		WHERE organization_id = $1 AND timestamp >= $2 AND timestamp <= $3
	`
	params := []interface{}{orgID, startDate, endDate}

	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339, cursor)
		if err == nil {
			query += " AND timestamp > $4"
			params = append(params, cursorTime)
		}
	}

	query += fmt.Sprintf(" ORDER BY timestamp ASC LIMIT %d", limit)

	if err := r.checkQueryCost(ctx, query, params...); err != nil {
		return nil, "", fmt.Errorf("cost guard error: %w", err)
	}

	rows, err := r.executeWithObservability(ctx, query, params...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query fleet hourly stats: %w", err)
	}
	defer rows.Close()

	var stats []*models.FleetHourlyStats
	var lastTs time.Time
	for rows.Next() {
		s := &models.FleetHourlyStats{}
		err := rows.Scan(
			&s.OrganizationID, &s.Timestamp, &s.ActiveVehicles, &s.TotalDistance, &s.TotalFuelConsumed,
			&s.AverageUtilization, &s.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan fleet hourly stats: %w", err)
		}
		stats = append(stats, s)
		lastTs = s.Timestamp
	}

	nextCursor := ""
	if len(stats) == limit {
		nextCursor = lastTs.Format(time.RFC3339)
	}

	return stats, nextCursor, nil
}

// CreateQueryLog inserts a new query audit log entry
func (r *AnalyticsRepositoryImpl) CreateQueryLog(ctx context.Context, log *models.AnalyticsQueryLog) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO analytics_query_logs (
			id, organization_id, user_id, query_type, parameters, 
			duration_ms, row_count, cache_hit, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.OrganizationID, log.UserID, log.QueryType, log.Parameters,
		log.DurationMs, log.RowCount, log.CacheHit, log.ErrorMessage, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create query log: %w", err)
	}

	return nil
}

// RefreshVehicleDailyStats performs a background aggregation update
func (r *AnalyticsRepositoryImpl) RefreshVehicleDailyStats(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	// High-fidelity aggregation from telemetry_points
	query := `
		INSERT INTO vehicle_daily_stats (
			vehicle_id, organization_id, date,
			total_distance, total_fuel_consumed, max_speed, avg_speed,
			harsh_braking_count, rapid_acceleration_count, sharp_turn_count, overspeeding_count,
			updated_at, last_refreshed_at
		)
		SELECT 
			vehicle_id, 
			tenant_id, 
			recorded_at::date,
			COALESCE(MAX(odometer) - MIN(odometer), 0),
			COALESCE(SUM(CASE WHEN speed > 0 THEN 0.1 ELSE 0 END), 0), -- Placeholder for complex fuel logic
			MAX(speed), 
			AVG(speed),
			COUNT(*) FILTER (WHERE event_type = 'HARSH_BRAKE'),
			COUNT(*) FILTER (WHERE event_type = 'RAPID_ACCEL'),
			COUNT(*) FILTER (WHERE event_type = 'SHARP_TURN'),
			COUNT(*) FILTER (WHERE event_type = 'OVER_SPEED'),
			NOW(),
			NOW()
		FROM telemetry_points
		WHERE recorded_at >= CURRENT_DATE - INTERVAL '2 days'
		GROUP BY vehicle_id, tenant_id, recorded_at::date
		ON CONFLICT (vehicle_id, date) DO UPDATE SET
			total_distance = EXCLUDED.total_distance,
			total_fuel_consumed = EXCLUDED.total_fuel_consumed,
			max_speed = EXCLUDED.max_speed,
			avg_speed = EXCLUDED.avg_speed,
			harsh_braking_count = EXCLUDED.harsh_braking_count,
			rapid_acceleration_count = EXCLUDED.rapid_acceleration_count,
			sharp_turn_count = EXCLUDED.sharp_turn_count,
			overspeeding_count = EXCLUDED.overspeeding_count,
			updated_at = NOW(),
			last_refreshed_at = NOW()
	`
	if err := r.checkQueryCost(ctx, query); err != nil {
		return fmt.Errorf("query too expensive: %w", err)
	}

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to aggregate vehicle daily stats: %w", err)
	}
	return nil
}

// RefreshDriverDailyStats performs a background aggregation update joining telemetry and trips
func (r *AnalyticsRepositoryImpl) RefreshDriverDailyStats(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO driver_daily_stats (
			driver_id, organization_id, date,
			total_driving_hours, total_distance, total_idle_hours,
			violation_count, updated_at, last_refreshed_at
		)
		SELECT 
			t.driver_id, 
			t.tenant_id, 
			tp.recorded_at::date,
			COUNT(*) * 1.0/360 as driving_hours, -- assuming 10s sampling
			COALESCE(MAX(tp.odometer) - MIN(tp.odometer), 0),
			COUNT(*) FILTER (WHERE NOT tp.ignition) * 1.0/360 as idle_hours,
			COUNT(*) FILTER (WHERE tp.event_type = 'VIOLATION'),
			NOW(),
			NOW()
		FROM telemetry_points tp
		JOIN trips t ON t.vehicle_id = tp.vehicle_id 
			AND tp.recorded_at BETWEEN t.start_time AND COALESCE(t.end_time, NOW())
		WHERE tp.recorded_at >= CURRENT_DATE - INTERVAL '2 days'
		GROUP BY t.driver_id, t.tenant_id, tp.recorded_at::date
		ON CONFLICT (driver_id, date) DO UPDATE SET
			total_driving_hours = EXCLUDED.total_driving_hours,
			total_distance = EXCLUDED.total_distance,
			total_idle_hours = EXCLUDED.total_idle_hours,
			violation_count = EXCLUDED.violation_count,
			updated_at = NOW(),
			last_refreshed_at = NOW()
	`
	if err := r.checkQueryCost(ctx, query); err != nil {
		return fmt.Errorf("query too expensive: %w", err)
	}

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to aggregate driver daily stats: %w", err)
	}
	return nil
}

// RefreshFleetHourlyStats performs a background aggregation update for organization-wide KPIs
func (r *AnalyticsRepositoryImpl) RefreshFleetHourlyStats(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO fleet_hourly_stats (
			organization_id, timestamp,
			active_vehicles, total_distance, average_utilization,
			updated_at, last_refreshed_at
		)
		SELECT 
			tenant_id, 
			date_trunc('hour', recorded_at),
			COUNT(DISTINCT vehicle_id),
			COALESCE(MAX(odometer) - MIN(odometer), 0),
			100.0 * COUNT(DISTINCT vehicle_id) / 100, -- Placeholder for capacity
			NOW(),
			NOW()
		FROM telemetry_points
		WHERE recorded_at >= NOW() - INTERVAL '48 hours'
		GROUP BY tenant_id, date_trunc('hour', recorded_at)
		ON CONFLICT (organization_id, timestamp) DO UPDATE SET
			active_vehicles = EXCLUDED.active_vehicles,
			total_distance = EXCLUDED.total_distance,
			average_utilization = EXCLUDED.average_utilization,
			updated_at = NOW(),
			last_refreshed_at = NOW()
	`
	if err := r.checkQueryCost(ctx, query); err != nil {
		return fmt.Errorf("query too expensive: %w", err)
	}

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to aggregate fleet hourly stats: %w", err)
	}
	return nil
}
// CreateFeatureVector bulk inserts feature store entries
func (r *AnalyticsRepositoryImpl) CreateFeatureVector(ctx context.Context, vector []*models.MLFeatureStore) error {
	if len(vector) == 0 {
		return nil
	}
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO ml_feature_store (organization_id, entity_type, entity_id, feature_name, feature_value, feature_timestamp)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	
	for _, f := range vector {
		_, err = tx.Exec(ctx, query, f.OrganizationID, f.EntityType, f.EntityID, f.FeatureName, f.FeatureValue, f.FeatureTimestamp)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetLatestFeatures retrieves the most recent features for an entity
func (r *AnalyticsRepositoryImpl) GetLatestFeatures(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) ([]*models.MLFeatureStore, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `SELECT organization_id, entity_type, entity_id, feature_name, feature_value, feature_timestamp
	          FROM ml_feature_store
	          WHERE organization_id = $1 AND entity_type = $2 AND entity_id = $3
	          ORDER BY feature_name, feature_timestamp DESC
	          LIMIT 100` // Fetch recent features
	
	rows, err := r.db.Query(ctx, query, orgID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []*models.MLFeatureStore
	for rows.Next() {
		f := &models.MLFeatureStore{}
		if err := rows.Scan(&f.OrganizationID, &f.EntityType, &f.EntityID, &f.FeatureName, &f.FeatureValue, &f.FeatureTimestamp); err != nil {
			return nil, err
		}
		features = append(features, f)
	}
	return features, nil
}

// PurgeAgedData deletes data older than the retention period
func (r *AnalyticsRepositoryImpl) PurgeAgedData(ctx context.Context, table string, retentionPeriod time.Duration) (int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	// Safety check for table name
	allowedTables := map[string]bool{
		"vehicle_daily_stats":   true,
		"driver_daily_stats":    true,
		"fleet_hourly_stats":    true,
		"prediction_results":    true,
		"analytics_query_logs":  true,
		"ml_feature_store":      true,
	}
	if !allowedTables[table] {
		return 0, fmt.Errorf("table %s not allowed for purging", table)
	}

	cutoff := time.Now().Add(-retentionPeriod)
	
	// Column name for timestamp varies by table
	timeColumn := "created_at"
	if table == "vehicle_daily_stats" || table == "driver_daily_stats" {
		timeColumn = "date"
	} else if table == "fleet_hourly_stats" {
		timeColumn = "timestamp"
	} else if table == "ml_feature_store" {
		timeColumn = "feature_timestamp"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s < $1", table, timeColumn)
	res, err := r.db.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected(), nil
}

// SetTenantContext sets the organization ID in the transaction for RLS with fail-safe check (Phase 9)
func (r *AnalyticsRepositoryImpl) SetTenantContext(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error {
	if orgID == uuid.Nil {
		return ErrInvalidTenant
	}
	_, err := tx.Exec(ctx, "SELECT set_config('app.org_id', $1, true)", orgID.String())
	return err
}

// SaveModelDeployment persists a model deployment strategy
func (r *AnalyticsRepositoryImpl) SaveModelDeployment(ctx context.Context, deployment *models.ModelDeployment) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO model_deployments (id, model_id, organization_id, deployment_strategy, traffic_split, is_active, deployed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			deployment_strategy = EXCLUDED.deployment_strategy,
			traffic_split = EXCLUDED.traffic_split,
			is_active = EXCLUDED.is_active,
			deployed_at = EXCLUDED.deployed_at
	`
	_, err := r.db.Exec(ctx, query, deployment.ID, deployment.ModelID, deployment.OrganizationID,
		deployment.DeploymentStrategy, deployment.TrafficSplit, deployment.IsActive, deployment.DeployedAt, deployment.CreatedAt)
	return err
}

// GetActiveModelDeployment retrieves the active deployment for a model
func (r *AnalyticsRepositoryImpl) GetActiveModelDeployment(ctx context.Context, modelID string, orgID uuid.UUID) (*models.ModelDeployment, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, model_id, organization_id, deployment_strategy, traffic_split, is_active, deployed_at, created_at
		FROM model_deployments
		WHERE model_id = $1 AND organization_id = $2 AND is_active = true
		LIMIT 1
	`
	var d models.ModelDeployment
	err := r.db.QueryRow(ctx, query, modelID, orgID).Scan(
		&d.ID, &d.ModelID, &d.OrganizationID, &d.DeploymentStrategy, &d.TrafficSplit, &d.IsActive, &d.DeployedAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// SaveDriverRiskEvent persists a risk event for coaching
func (r *AnalyticsRepositoryImpl) SaveDriverRiskEvent(ctx context.Context, event *models.DriverRiskEvent) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO driver_risk_events (id, driver_id, organization_id, risk_type, score_delta, metadata, event_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, event.ID, event.DriverID, event.OrganizationID,
		event.RiskType, event.ScoreDelta, event.Metadata, event.EventTime, event.CreatedAt)
	return err
}

// GetDriverRiskEvents retrieves risk events for a specific driver and period
func (r *AnalyticsRepositoryImpl) GetDriverRiskEvents(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.DriverRiskEvent, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, driver_id, organization_id, risk_type, score_delta, metadata, event_time, created_at
		FROM driver_risk_events
		WHERE driver_id = $1 AND organization_id = $2 AND event_time BETWEEN $3 AND $4
		ORDER BY event_time DESC
	`
	rows, err := r.db.Query(ctx, query, driverID, orgID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.DriverRiskEvent
	for rows.Next() {
		var e models.DriverRiskEvent
		if err := rows.Scan(&e.ID, &e.DriverID, &e.OrganizationID, &e.RiskType, &e.ScoreDelta, &e.Metadata, &e.EventTime, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

// GetAnalyticsQuota retrieves the current quota for an organization
func (r *AnalyticsRepositoryImpl) GetAnalyticsQuota(ctx context.Context, orgID uuid.UUID) (*models.AnalyticsQuota, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `SELECT organization_id, daily_limit, used_credits, last_reset_at FROM analytics_quotas WHERE organization_id = $1`
	var q models.AnalyticsQuota
	err := r.db.QueryRow(ctx, query, orgID).Scan(&q.OrganizationID, &q.DailyLimit, &q.UsedCredits, &q.LastResetAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// UpdateUsedCredits updates the used credits for an organization
func (r *AnalyticsRepositoryImpl) UpdateUsedCredits(ctx context.Context, orgID uuid.UUID, credits int64) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `UPDATE analytics_quotas SET used_credits = used_credits + $1, last_reset_at = NOW() WHERE organization_id = $2`
	_, err := r.db.Exec(ctx, query, credits, orgID)
	return err
}

// CreateVersionedFeatureVector bulk inserts feature vectors with versioning and lineage (Phase 11)
func (r *AnalyticsRepositoryImpl) CreateVersionedFeatureVector(ctx context.Context, vector []*models.MLFeatureStore) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	batch := `INSERT INTO ml_feature_store (organization_id, entity_type, entity_id, feature_name, feature_value, feature_version, feature_lineage, feature_source, feature_timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, f := range vector {
		_, err := r.db.Exec(ctx, batch, f.OrganizationID, f.EntityType, f.EntityID, f.FeatureName, f.FeatureValue, f.FeatureVersion, f.FeatureLineage, f.FeatureSource, f.FeatureTimestamp)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetFeaturesByVersion retrieves features for a specific version (Phase 10/11 Lineage)
func (r *AnalyticsRepositoryImpl) GetFeaturesByVersion(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID, version string) ([]*models.MLFeatureStore, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT organization_id, entity_type, entity_id, feature_name, feature_value, feature_version, feature_lineage, feature_source, feature_timestamp
		FROM ml_feature_store
		WHERE organization_id = $1 AND entity_type = $2 AND entity_id = $3 AND feature_version = $4
	`
	rows, err := r.db.Query(ctx, query, orgID, entityType, entityID, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []*models.MLFeatureStore
	for rows.Next() {
		var f models.MLFeatureStore
		if err := rows.Scan(&f.OrganizationID, &f.EntityType, &f.EntityID, &f.FeatureName, &f.FeatureValue, &f.FeatureVersion, &f.FeatureLineage, &f.FeatureSource, &f.FeatureTimestamp); err != nil {
			return nil, err
		}
		features = append(features, &f)
	}
	return features, nil
}

// SaveQueryPlan persists a query plan for slow query analysis (Phase 10)
func (r *AnalyticsRepositoryImpl) SaveQueryPlan(ctx context.Context, plan *models.AnalyticsQueryPlan) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO analytics_query_plans (id, organization_id, query_type, query_params, execution_duration_ms, query_plan, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, plan.ID, plan.OrganizationID, plan.QueryType, plan.QueryParams, plan.ExecutionDuration, plan.QueryPlan, plan.CreatedAt)
	return err
}

// CreateAccessLog records an analytics access event (Phase 11 Compliance)
func (r *AnalyticsRepositoryImpl) CreateAccessLog(ctx context.Context, log *models.AnalyticsAccessLog) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO analytics_access_logs (id, organization_id, user_id, resource_id, resource_type, access_context, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, log.ID, log.OrganizationID, log.UserID, log.ResourceID, log.ResourceType, log.AccessContext, log.IPAddress, log.UserAgent, log.CreatedAt)
	return err
}

// GetMaterializedReports retrieves pre-computed summaries (Phase 11)
func (r *AnalyticsRepositoryImpl) GetMaterializedReports(ctx context.Context, orgID uuid.UUID, reportType string, period string) ([]*models.MaterializedFleetReport, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, organization_id, report_type, metric_key, metric_value, period, generated_at
		FROM materialized_fleet_reports
		WHERE organization_id = $1 AND report_type = $2 AND period = $3
	`
	rows, err := r.db.Query(ctx, query, orgID, reportType, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.MaterializedFleetReport
	for rows.Next() {
		var rep models.MaterializedFleetReport
		if err := rows.Scan(&rep.ID, &rep.OrganizationID, &rep.ReportType, &rep.MetricKey, &rep.MetricValue, &rep.Period, &rep.GeneratedAt); err != nil {
			return nil, err
		}
		reports = append(reports, &rep)
	}
	return reports, nil
}

// GetCapacityStats retrieves table size and partition count metrics (Phase 11)
func (r *AnalyticsRepositoryImpl) GetCapacityStats(ctx context.Context) (map[string]int64, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `SELECT table_name, data_size FROM v_analytics_capacity_stats`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var name string
		var size int64
		if err := rows.Scan(&name, &size); err != nil {
			return nil, err
		}
		stats[name] = size
	}
	return stats, nil
}
