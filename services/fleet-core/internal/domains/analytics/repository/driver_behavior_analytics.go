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
	"github.com/flyfleet/fleet-core/internal/domains/analytics/utils"
	"github.com/google/uuid"
)

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
