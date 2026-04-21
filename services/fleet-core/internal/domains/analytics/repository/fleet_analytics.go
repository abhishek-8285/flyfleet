package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
)

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
func (r *AnalyticsRepositoryImpl) GetFleetAnalytics(ctx context.Context, id string, tenantID uuid.UUID) (*models.FleetAnalytics, error) {
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

	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
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
func (r *AnalyticsRepositoryImpl) GetFleetAnalyticsByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
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

	err := r.db.QueryRow(ctx, query, tenantID, startDate, endDate).Scan(
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
func (r *AnalyticsRepositoryImpl) UpdateFleetAnalytics(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE fleet_analytics SET %s WHERE id = $%d AND organization_id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update fleet analytics %s: %w", id, err)
	}
	return nil
}

// DeleteFleetAnalytics deletes fleet analytics
func (r *AnalyticsRepositoryImpl) DeleteFleetAnalytics(ctx context.Context, id string, tenantID uuid.UUID) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	query := "DELETE FROM fleet_analytics WHERE id = $1 AND organization_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete fleet analytics %s: %w", id, err)
	}
	return nil
}

// ListFleetAnalytics retrieves fleet analytics with filtering
func (r *AnalyticsRepositoryImpl) ListFleetAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate *time.Time, cursor string, limit, offset int) ([]*models.FleetAnalytics, int, error) {
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
	countQuery := `SELECT COUNT(*) FROM fleet_analytics WHERE organization_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2
	var conditions []string

	if startDate != nil && !startDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("start_date >= $%d", argIndex))
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil && !endDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("end_date <= $%d", argIndex))
		args = append(args, *endDate)
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
		return nil, 0, fmt.Errorf("failed to count fleet analytics: %w", err)
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
		return nil, 0, fmt.Errorf("failed to query fleet analytics: %w", err)
	}
	defer rows.Close()

	var analyticsList []*models.FleetAnalytics
	for rows.Next() {
		var a models.FleetAnalytics
		err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.StartDate, &a.EndDate,
			&a.TotalVehicles, &a.ActiveVehicles, &a.TotalDistance, &a.AverageSpeed, &a.TotalFuelConsumption,
			&a.AverageFuelEfficiency, &a.TotalTrips, &a.CompletedTrips, &a.CancelledTrips,
			&a.AverageTripDuration, &a.IdleTimePercentage, &a.RoutingEfficiency, &a.FuelCost,
			&a.MaintenanceCost, &a.Revenue, &a.ProfitMargin, &a.SafetyScore, &a.AccidentsCount,
			&a.IncidentsCount, &a.EmergencyStops, &a.GeofenceViolations, &a.AlertsTriggered,
			&a.DataPointsCollected, &a.ProcessingTime, &a.GeneratedAt,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan fleet analytics: %w", err)
		}
		analyticsList = append(analyticsList, &a)
	}

	return analyticsList, totalCount, nil
}
