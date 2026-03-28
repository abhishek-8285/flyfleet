package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/utils"
	"github.com/google/uuid"
)

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
