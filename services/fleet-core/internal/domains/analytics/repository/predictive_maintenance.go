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

func (r *AnalyticsRepositoryImpl) CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		INSERT INTO predictive_maintenance_analytics (id, vehicle_id, tenant_id, analysis_date, 
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

	_, err = r.db.Exec(ctx, query, analytics.ID, analytics.VehicleID, analytics.TenantID,
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
func (r *AnalyticsRepositoryImpl) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string, tenantID uuid.UUID) (*models.PredictiveMaintenanceAnalytics, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, vehicle_id, tenant_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics WHERE id = $1 AND tenant_id = $2
	`

	var analytics models.PredictiveMaintenanceAnalytics
	var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&analytics.ID, &analytics.VehicleID, &analytics.TenantID, &analytics.AnalysisDate,
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
func (r *AnalyticsRepositoryImpl) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	// Count query
	countQuery := "SELECT COUNT(*) FROM predictive_maintenance_analytics WHERE vehicle_id = $1 AND tenant_id = $2"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, vehicleID, tenantID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count predictive maintenance analytics: %w", err)
	}

	// Main query
	query := `
		SELECT id, vehicle_id, tenant_id, analysis_date, 
			next_service_due, service_priority, component_risks, failure_prediction, 
			maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, 
			reliability_score, recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics 
		WHERE vehicle_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4
	`

	rows, err := r.executeWithObservability(ctx, query, vehicleID, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query predictive maintenance analytics: %w", err)
	}
	defer rows.Close()

	var analyticsList []*models.PredictiveMaintenanceAnalytics
	for rows.Next() {
		var analytics models.PredictiveMaintenanceAnalytics
		var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte

		err := rows.Scan(
			&analytics.ID, &analytics.VehicleID, &analytics.TenantID, &analytics.AnalysisDate,
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
func (r *AnalyticsRepositoryImpl) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE predictive_maintenance_analytics SET %s WHERE id = $%d AND tenant_id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update predictive maintenance analytics %s: %w", id, err)
	}
	return nil
}

// DeletePredictiveMaintenanceAnalytics deletes predictive maintenance analytics
func (r *AnalyticsRepositoryImpl) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string, tenantID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM predictive_maintenance_analytics WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete predictive maintenance analytics %s: %w", id, err)
	}
	return nil
}

// ListPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics with filtering
func (r *AnalyticsRepositoryImpl) ListPredictiveMaintenanceAnalytics(ctx context.Context, tenantID uuid.UUID, vehicleID *uuid.UUID, cursor string, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
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

	// This method lists predictive maintenance analytics for an organization, optionally filtering by vehicle.
	// The cursor is based on analysis_date (or created_at).
	baseQuery := `
		SELECT id, vehicle_id, tenant_id, analysis_date, next_service_due,
		       service_priority, component_risks, failure_prediction, maintenance_schedule,
		       parts_needed, cost_estimation, uptime_prediction, reliability_score,
		       recommendations, generated_at, created_at, updated_at
		FROM predictive_maintenance_analytics
		WHERE tenant_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM predictive_maintenance_analytics WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2
	var conditions []string

	if vehicleID != nil {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, *vehicleID)
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
		return nil, 0, fmt.Errorf("failed to count predictive maintenance analytics: %w", err)
	}

	// List with pagination and cursor
	query := baseQuery + whereClause
	if cursor != "" {
		query += " AND analysis_date < $" + fmt.Sprintf("%d", argIndex)
		args = append(args, cursor)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY analysis_date DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query predictive maintenance analytics: %w", err)
	}
	defer rows.Close()

	var analyticsList []*models.PredictiveMaintenanceAnalytics
	for rows.Next() {
		var a models.PredictiveMaintenanceAnalytics
		var componentRisksJSON, failurePredictionJSON, maintenanceScheduleJSON, partsNeededJSON, recommendationsJSON []byte

		err := rows.Scan(
			&a.ID, &a.VehicleID, &a.TenantID, &a.AnalysisDate, &a.NextServiceDue,
			&a.ServicePriority, &componentRisksJSON, &failurePredictionJSON, &maintenanceScheduleJSON,
			&partsNeededJSON, &a.CostEstimation, &a.UptimePrediction, &a.ReliabilityScore,
			&recommendationsJSON, &a.GeneratedAt, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan predictive maintenance analytics: %w", err)
		}

		if err := json.Unmarshal(componentRisksJSON, &a.ComponentRisks); err != nil {
			a.ComponentRisks = nil
		}
		if err := json.Unmarshal(failurePredictionJSON, &a.FailurePrediction); err != nil {
			a.FailurePrediction = nil
		}
		if err := json.Unmarshal(maintenanceScheduleJSON, &a.MaintenanceSchedule); err != nil {
			a.MaintenanceSchedule = nil
		}
		if err := json.Unmarshal(partsNeededJSON, &a.PartsNeeded); err != nil {
			a.PartsNeeded = nil
		}
		if err := json.Unmarshal(recommendationsJSON, &a.Recommendations); err != nil {
			a.Recommendations = nil
		}

		analyticsList = append(analyticsList, &a)
	}

	return analyticsList, totalCount, nil
}
