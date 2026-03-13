package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/route/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteOptimizationRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRouteOptimizationRepository creates a new route optimization repository
func NewRouteOptimizationRepository(db *pgxpool.Pool) RouteOptimizationRepository {
	return &RouteOptimizationRepositoryImpl{db: db}
}

// CreateOptimizationRequest creates a new optimization request
func (r *RouteOptimizationRepositoryImpl) CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error {
	query := `
		INSERT INTO route_optimization_requests (
			id, organization_id, name, description, start_time, end_time,
			start_time_window, end_time_window, constraints, stops, vehicles,
			priorities, optimization_goals, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	constraintsJSON, _ := json.Marshal(request.Constraints)
	stopsJSON, _ := json.Marshal(request.Stops)
	vehiclesJSON, _ := json.Marshal(request.Vehicles)
	prioritiesJSON, _ := json.Marshal(request.Priorities)
	goalsJSON, _ := json.Marshal(request.OptimizationGoals)

	_, err := r.db.Exec(ctx, query,
		request.ID, request.OrganizationID, request.Name, request.Description,
		request.StartTime, request.EndTime, request.StartTimeWindow, request.EndTimeWindow,
		constraintsJSON, stopsJSON, vehiclesJSON, prioritiesJSON, goalsJSON,
		request.Status, request.CreatedAt, request.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create optimization request: %w", err)
	}
	return nil
}

// GetOptimizationRequest retrieves an optimization request by ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error) {
	query := `
		SELECT id, organization_id, name, description, start_time, end_time,
		       start_time_window, end_time_window, constraints, stops, vehicles,
		       priorities, optimization_goals, status, result, created_at, updated_at
		FROM route_optimization_requests
		WHERE id = $1
	`

	var request models.RouteOptimizationRequest
	var constraintsJSON, stopsJSON, vehiclesJSON, prioritiesJSON, goalsJSON, resultJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&request.ID, &request.OrganizationID, &request.Name, &request.Description,
		&request.StartTime, &request.EndTime, &request.StartTimeWindow, &request.EndTimeWindow,
		&constraintsJSON, &stopsJSON, &vehiclesJSON, &prioritiesJSON, &goalsJSON,
		&request.Status, &resultJSON, &request.CreatedAt, &request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get optimization request: %w", err)
	}

	if err := json.Unmarshal(constraintsJSON, &request.Constraints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal constraints: %w", err)
	}
	if err := json.Unmarshal(stopsJSON, &request.Stops); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stops: %w", err)
	}
	if err := json.Unmarshal(vehiclesJSON, &request.Vehicles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vehicles: %w", err)
	}
	if err := json.Unmarshal(prioritiesJSON, &request.Priorities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal priorities: %w", err)
	}
	if err := json.Unmarshal(goalsJSON, &request.OptimizationGoals); err != nil {
		return nil, fmt.Errorf("failed to unmarshal goals: %w", err)
	}
	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &request.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return &request, nil
}

// UpdateOptimizationRequest updates an optimization request
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		// Handle JSON fields specially
		if key == "constraints" || key == "stops" || key == "vehicles" || key == "priorities" || key == "optimization_goals" || key == "result" {
			if jsonBytes, ok := value.([]byte); ok {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
				args = append(args, jsonBytes)
				argIndex++
			} else if jsonBytes, err := json.Marshal(value); err == nil {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
				args = append(args, jsonBytes)
				argIndex++
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE route_optimization_requests SET %s WHERE id = $%d",
		fmt.Sprintf("%s", setParts), argIndex)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update optimization request: %w", err)
	}
	return nil
}

// DeleteOptimizationRequest deletes an optimization request
func (r *RouteOptimizationRepositoryImpl) DeleteOptimizationRequest(ctx context.Context, id string) error {
	query := "DELETE FROM route_optimization_requests WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete optimization request: %w", err)
	}
	return nil
}

// ListOptimizationRequests lists optimization requests for an organization
func (r *RouteOptimizationRepositoryImpl) ListOptimizationRequests(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error) {
	countQuery := "SELECT COUNT(*) FROM route_optimization_requests WHERE organization_id = $1"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count optimization requests: %w", err)
	}

	query := `
		SELECT id, organization_id, name, description, start_time, end_time,
		       start_time_window, end_time_window, constraints, stops, vehicles,
		       priorities, optimization_goals, status, result, created_at, updated_at
		FROM route_optimization_requests
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list optimization requests: %w", err)
	}
	defer rows.Close()

	var requests []*models.RouteOptimizationRequest
	for rows.Next() {
		var request models.RouteOptimizationRequest
		var constraintsJSON, stopsJSON, vehiclesJSON, prioritiesJSON, goalsJSON, resultJSON []byte

		err := rows.Scan(
			&request.ID, &request.OrganizationID, &request.Name, &request.Description,
			&request.StartTime, &request.EndTime, &request.StartTimeWindow, &request.EndTimeWindow,
			&constraintsJSON, &stopsJSON, &vehiclesJSON, &prioritiesJSON, &goalsJSON,
			&request.Status, &resultJSON, &request.CreatedAt, &request.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan optimization request: %w", err)
		}

		if err := json.Unmarshal(constraintsJSON, &request.Constraints); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal constraints: %w", err)
		}
		if err := json.Unmarshal(stopsJSON, &request.Stops); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal stops: %w", err)
		}
		if err := json.Unmarshal(vehiclesJSON, &request.Vehicles); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal vehicles: %w", err)
		}
		if err := json.Unmarshal(prioritiesJSON, &request.Priorities); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal priorities: %w", err)
		}
		if err := json.Unmarshal(goalsJSON, &request.OptimizationGoals); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal goals: %w", err)
		}
		if len(resultJSON) > 0 {
			if err := json.Unmarshal(resultJSON, &request.Result); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		requests = append(requests, &request)
	}

	return requests, total, nil
}

// CreateOptimizationResult creates a new optimization result
func (r *RouteOptimizationRepositoryImpl) CreateOptimizationResult(ctx context.Context, result *models.RouteOptimizationResult) error {
	query := `
		INSERT INTO route_optimization_results (
			id, request_id, optimization_time, total_distance, total_time,
			total_cost, total_vehicles_used, total_stops, routes, performance_metrics,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	routesJSON, _ := json.Marshal(result.Routes)
	metricsJSON, _ := json.Marshal(result.PerformanceMetrics)

	_, err := r.db.Exec(ctx, query,
		result.ID, result.RequestID, result.OptimizationTime,
		result.TotalDistance, result.TotalTime, result.TotalCost,
		result.TotalVehiclesUsed, result.TotalStops, routesJSON, metricsJSON,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create optimization result: %w", err)
	}
	return nil
}

// GetOptimizationResult retrieves an optimization result by ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error) {
	query := `
		SELECT id, request_id, optimization_time, total_distance, total_time,
		       total_cost, total_vehicles_used, total_stops, routes, performance_metrics,
		       created_at
		FROM route_optimization_results
		WHERE id = $1
	`

	var result models.RouteOptimizationResult
	var routesJSON, metricsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.RequestID, &result.OptimizationTime,
		&result.TotalDistance, &result.TotalTime, &result.TotalCost,
		&result.TotalVehiclesUsed, &result.TotalStops, &routesJSON, &metricsJSON,
		&result.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get optimization result: %w", err)
	}

	if err := json.Unmarshal(routesJSON, &result.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}
	if err := json.Unmarshal(metricsJSON, &result.PerformanceMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &result, nil
}

// GetOptimizationResultByRequestID retrieves an optimization result by request ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error) {
	query := `
		SELECT id, request_id, optimization_time, total_distance, total_time,
		       total_cost, total_vehicles_used, total_stops, routes, performance_metrics,
		       created_at
		FROM route_optimization_results
		WHERE request_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var result models.RouteOptimizationResult
	var routesJSON, metricsJSON []byte

	err := r.db.QueryRow(ctx, query, requestID).Scan(
		&result.ID, &result.RequestID, &result.OptimizationTime,
		&result.TotalDistance, &result.TotalTime, &result.TotalCost,
		&result.TotalVehiclesUsed, &result.TotalStops, &routesJSON, &metricsJSON,
		&result.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get optimization result: %w", err)
	}

	if err := json.Unmarshal(routesJSON, &result.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}
	if err := json.Unmarshal(metricsJSON, &result.PerformanceMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &result, nil
}

// UpdateOptimizationResult updates an optimization result
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationResult(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if jsonBytes, err := json.Marshal(value); err == nil {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, jsonBytes)
			argIndex++
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE route_optimization_results SET %s WHERE id = $%d",
		fmt.Sprintf("%s", setParts), argIndex)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update optimization result: %w", err)
	}
	return nil
}

// DeleteOptimizationResult deletes an optimization result
func (r *RouteOptimizationRepositoryImpl) DeleteOptimizationResult(ctx context.Context, id string) error {
	query := "DELETE FROM route_optimization_results WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete optimization result: %w", err)
	}
	return nil
}

// ListOptimizationResults lists optimization results for a request
func (r *RouteOptimizationRepositoryImpl) ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error) {
	countQuery := "SELECT COUNT(*) FROM route_optimization_results WHERE request_id = $1"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, requestID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count optimization results: %w", err)
	}

	query := `
		SELECT id, request_id, optimization_time, total_distance, total_time,
		       total_cost, total_vehicles_used, total_stops, routes, performance_metrics,
		       created_at
		FROM route_optimization_results
		WHERE request_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, requestID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list optimization results: %w", err)
	}
	defer rows.Close()

	var results []*models.RouteOptimizationResult
	for rows.Next() {
		var result models.RouteOptimizationResult
		var routesJSON, metricsJSON []byte

		err := rows.Scan(
			&result.ID, &result.RequestID, &result.OptimizationTime,
			&result.TotalDistance, &result.TotalTime, &result.TotalCost,
			&result.TotalVehiclesUsed, &result.TotalStops, &routesJSON, &metricsJSON,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan optimization result: %w", err)
		}

		if err := json.Unmarshal(routesJSON, &result.Routes); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal routes: %w", err)
		}
		if err := json.Unmarshal(metricsJSON, &result.PerformanceMetrics); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal metrics: %w", err)
		}

		results = append(results, &result)
	}

	return results, total, nil
}

// CreateOptimizationJob creates a new optimization job
func (r *RouteOptimizationRepositoryImpl) CreateOptimizationJob(ctx context.Context, job *models.OptimizationJob) error {
	query := `
		INSERT INTO route_optimization_jobs (
			id, request_id, status, created_at, started_at, completed_at, error, result_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		job.ID, job.RequestID, job.Status, job.CreatedAt,
		job.StartedAt, job.CompletedAt, job.Error, job.ResultID,
	)

	if err != nil {
		return fmt.Errorf("failed to create optimization job: %w", err)
	}
	return nil
}

// GetOptimizationJob retrieves an optimization job by ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error) {
	query := `
		SELECT id, request_id, status, created_at, started_at, completed_at, error, result_id
		FROM route_optimization_jobs
		WHERE id = $1
	`

	var job models.OptimizationJob
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.RequestID, &job.Status, &job.CreatedAt,
		&job.StartedAt, &job.CompletedAt, &job.Error, &job.ResultID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get optimization job: %w", err)
	}

	return &job, nil
}

// UpdateOptimizationJob updates an optimization job
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationJob(ctx context.Context, id string, updates map[string]interface{}) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE route_optimization_jobs SET %s WHERE id = $%d",
		fmt.Sprintf("%s", setParts), argIndex)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update optimization job: %w", err)
	}
	return nil
}

// UpdateOptimizationJobStatus updates an optimization job status
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationJobStatus(ctx context.Context, id, status string, completedAt *time.Time, errMsg *string) error {
	query := `
		UPDATE route_optimization_jobs
		SET status = $1, completed_at = $2, error = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, status, completedAt, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update optimization job status: %w", err)
	}
	return nil
}

// ListOptimizationJobs lists optimization jobs for a request
func (r *RouteOptimizationRepositoryImpl) ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error) {
	countQuery := "SELECT COUNT(*) FROM route_optimization_jobs WHERE request_id = $1"
	countArgs := []interface{}{requestID}
	argIndex := 2

	if status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		countArgs = append(countArgs, status)
		argIndex++
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count optimization jobs: %w", err)
	}

	query := `
		SELECT id, request_id, status, created_at, started_at, completed_at, error, result_id
		FROM route_optimization_jobs
		WHERE request_id = $1
	`
	args := []interface{}{requestID}
	argIndex = 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list optimization jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.OptimizationJob
	for rows.Next() {
		var job models.OptimizationJob
		err := rows.Scan(
			&job.ID, &job.RequestID, &job.Status, &job.CreatedAt,
			&job.StartedAt, &job.CompletedAt, &job.Error, &job.ResultID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan optimization job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, total, nil
}

// ListRoadRestrictions lists road restrictions for a set of road IDs
func (r *RouteOptimizationRepositoryImpl) ListRoadRestrictions(ctx context.Context, roadIDs []string) ([]*models.RoadRestriction, error) {
	if len(roadIDs) == 0 {
		return []*models.RoadRestriction{}, nil
	}

	query := `
		SELECT id, road_id, restriction_type, allowed_vehicle_classes, permit_required, max_weight, max_height
		FROM road_restrictions
		WHERE road_id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, roadIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list road restrictions: %w", err)
	}
	defer rows.Close()

	var restrictions []*models.RoadRestriction
	for rows.Next() {
		var rst models.RoadRestriction
		var allowedClassesJSON []byte
		err := rows.Scan(&rst.ID, &rst.RoadID, &rst.RestrictionType, &allowedClassesJSON, &rst.PermitRequired, &rst.MaxWeight, &rst.MaxHeight)
		if err != nil {
			return nil, fmt.Errorf("failed to scan road restriction: %w", err)
		}
		if len(allowedClassesJSON) > 0 {
			json.Unmarshal(allowedClassesJSON, &rst.AllowedVehicleClasses)
		}
		restrictions = append(restrictions, &rst)
	}

	return restrictions, nil
}

// ListRoadTimeRestrictions lists time-based road restrictions
func (r *RouteOptimizationRepositoryImpl) ListRoadTimeRestrictions(ctx context.Context, roadIDs []string) ([]*models.RoadTimeRestriction, error) {
	if len(roadIDs) == 0 {
		return []*models.RoadTimeRestriction{}, nil
	}

	query := `
		SELECT id, road_id, start_time, end_time, allowed_vehicle_classes
		FROM road_time_restrictions
		WHERE road_id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, roadIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list road time restrictions: %w", err)
	}
	defer rows.Close()

	var restrictions []*models.RoadTimeRestriction
	for rows.Next() {
		var rst models.RoadTimeRestriction
		var allowedClassesJSON []byte
		err := rows.Scan(&rst.ID, &rst.RoadID, &rst.StartTime, &rst.EndTime, &allowedClassesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan road time restriction: %w", err)
		}
		if len(allowedClassesJSON) > 0 {
			json.Unmarshal(allowedClassesJSON, &rst.AllowedVehicleClasses)
		}
		restrictions = append(restrictions, &rst)
	}

	return restrictions, nil
}

// GetVehiclePermits retrieves all permits for a specific vehicle
func (r *RouteOptimizationRepositoryImpl) GetVehiclePermits(ctx context.Context, vehicleID string) ([]*models.VehiclePermit, error) {
	query := `
		SELECT id, vehicle_id, permit_type, valid_states, expiry_date
		FROM vehicle_permits
		WHERE vehicle_id = $1
	`

	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle permits: %w", err)
	}
	defer rows.Close()

	var permits []*models.VehiclePermit
	for rows.Next() {
		var pmt models.VehiclePermit
		var validStatesJSON []byte
		err := rows.Scan(&pmt.ID, &pmt.VehicleID, &pmt.PermitType, &validStatesJSON, &pmt.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle permit: %w", err)
		}
		if len(validStatesJSON) > 0 {
			json.Unmarshal(validStatesJSON, &pmt.ValidStates)
		}
		permits = append(permits, &pmt)
	}

	return permits, nil
}

// HealthCheck performs a health check on the database
func (r *RouteOptimizationRepositoryImpl) HealthCheck(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// Ping pings the database
func (r *RouteOptimizationRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
