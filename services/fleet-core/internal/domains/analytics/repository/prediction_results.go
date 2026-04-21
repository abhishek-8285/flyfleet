package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
)

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

	// Batch insert prediction results to avoid N+1 queries
	if len(results) > 0 {
		var args []interface{}
		var placeholders []string
		for i, result := range results {
			inputJSON, _ := json.Marshal(result.PredictionInput)
			probJSON, _ := json.Marshal(result.ProbabilityScores)
			featJSON, _ := json.Marshal(result.FeatureImportance)

			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*15+1, i*15+2, i*15+3, i*15+4, i*15+5, i*15+6, i*15+7, i*15+8, i*15+9, i*15+10, i*15+11, i*15+12, i*15+13, i*15+14, i*15+15))
			args = append(args,
				result.ID, result.ModelID, inputJSON, result.PredictedValue,
				result.ConfidenceScore, probJSON, featJSON, result.AnomalyDetected,
				result.AnomalyScore, result.Explanation, result.OrganizationID, result.RequestID, result.PredictedAt,
				result.CreatedAt, result.UpdatedAt)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO prediction_results (id, model_id, prediction_input, predicted_value, confidence_score, "+
				"probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, "+
				"organization_id, request_id, predicted_at, created_at, updated_at) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err = tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert prediction results: %w", err)
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
func (r *AnalyticsRepositoryImpl) GetPredictionResult(ctx context.Context, id string, tenantID uuid.UUID) (*models.PredictionResult, error) {
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

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
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
func (r *AnalyticsRepositoryImpl) GetPredictionResultsByModel(ctx context.Context, modelID string, tenantID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM prediction_results WHERE model_id = $1 AND organization_id = $2"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, modelID, tenantID).Scan(&totalCount)
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

	rows, err := r.executeWithObservability(ctx, query, modelID, tenantID, limit, offset)
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
func (r *AnalyticsRepositoryImpl) GetPredictionResultsByOrg(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM prediction_results WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&totalCount)
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

	rows, err := r.executeWithObservability(ctx, query, tenantID, limit, offset)
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
func (r *AnalyticsRepositoryImpl) UpdatePredictionResult(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
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

	args = append(args, id, tenantID)
	query := fmt.Sprintf("UPDATE prediction_results SET %s WHERE id = $%d AND organization_id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update prediction result %s: %w", id, err)
	}
	return nil
}

// DeletePredictionResult deletes a prediction result
func (r *AnalyticsRepositoryImpl) DeletePredictionResult(ctx context.Context, id string, tenantID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM prediction_results WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete prediction result %s: %w", id, err)
	}
	return nil
}

// ListPredictionResults retrieves prediction results with filtering
func (r *AnalyticsRepositoryImpl) ListPredictionResults(ctx context.Context, tenantID uuid.UUID, modelID *string, anomalyDetected *bool, cursor string, limit, offset int) ([]*models.PredictionResult, int, error) {
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
		SELECT id, model_id, prediction_input, predicted_value, confidence_score,
		       probability_scores, feature_importance, anomaly_detected, anomaly_score,
		       explanation, organization_id, request_id, predicted_at, created_at, updated_at
		FROM prediction_results
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM prediction_results WHERE organization_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2
	var conditions []string

	if modelID != nil && *modelID != "" {
		conditions = append(conditions, fmt.Sprintf("model_id = $%d", argIndex))
		args = append(args, *modelID)
		argIndex++
	}

	if anomalyDetected != nil {
		conditions = append(conditions, fmt.Sprintf("anomaly_detected = $%d", argIndex))
		args = append(args, *anomalyDetected)
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
		return nil, 0, fmt.Errorf("failed to count prediction results: %w", err)
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
		return nil, 0, fmt.Errorf("failed to query prediction results: %w", err)
	}
	defer rows.Close()

	var results []*models.PredictionResult
	for rows.Next() {
		var pr models.PredictionResult
		var predictionInputJSON, probabilityScoresJSON, featureImportanceJSON []byte

		err := rows.Scan(
			&pr.ID, &pr.ModelID, &predictionInputJSON, &pr.PredictedValue, &pr.ConfidenceScore,
			&probabilityScoresJSON, &featureImportanceJSON, &pr.AnomalyDetected, &pr.AnomalyScore,
			&pr.Explanation, &pr.OrganizationID, &pr.RequestID, &pr.PredictedAt, &pr.CreatedAt, &pr.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan prediction result: %w", err)
		}

		if err := json.Unmarshal(predictionInputJSON, &pr.PredictionInput); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal prediction input: %w", err)
		}
		if err := json.Unmarshal(probabilityScoresJSON, &pr.ProbabilityScores); err != nil {
			pr.ProbabilityScores = nil
		}
		if err := json.Unmarshal(featureImportanceJSON, &pr.FeatureImportance); err != nil {
			pr.FeatureImportance = nil
		}

		results = append(results, &pr)
	}

	return results, totalCount, nil
}
