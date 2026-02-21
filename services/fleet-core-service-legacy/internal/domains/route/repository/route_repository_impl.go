package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/route/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RouteOptimizationRepositoryImpl implements the RouteOptimizationRepository interface
type RouteOptimizationRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRouteOptimizationRepository creates a new instance of RouteOptimizationRepositoryImpl
func NewRouteOptimizationRepository(db *pgxpool.Pool) *RouteOptimizationRepositoryImpl {
	return &RouteOptimizationRepositoryImpl{
		db: db,
	}
}

// CreateOptimizationRequest creates a new route optimization request
func (r *RouteOptimizationRepositoryImpl) CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error {
	query := `
		INSERT INTO route_optimization_requests (
			id, organization_id, name, description, start_time, end_time, 
			start_time_window, end_time_window, constraints, stops, vehicles, 
			priorities, optimization_goals, created_at, updated_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	stopsJSON, err := json.Marshal(request.Stops)
	if err != nil {
		return fmt.Errorf("failed to marshal stops: %w", err)
	}

	vehiclesJSON, err := json.Marshal(request.Vehicles)
	if err != nil {
		return fmt.Errorf("failed to marshal vehicles: %w", err)
	}

	constraintsJSON, err := json.Marshal(request.Constraints)
	if err != nil {
		return fmt.Errorf("failed to marshal constraints: %w", err)
	}

	prioritiesJSON, err := json.Marshal(request.Priorities)
	if err != nil {
		return fmt.Errorf("failed to marshal priorities: %w", err)
	}

	goalsJSON, err := json.Marshal(request.OptimizationGoals)
	if err != nil {
		return fmt.Errorf("failed to marshal optimization goals: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		request.ID,
		request.OrganizationID,
		request.Name,
		request.Description,
		request.StartTime,
		request.EndTime,
		request.StartTimeWindow,
		request.EndTimeWindow,
		constraintsJSON,
		stopsJSON,
		vehiclesJSON,
		prioritiesJSON,
		goalsJSON,
		request.CreatedAt,
		request.UpdatedAt,
		request.Status,
	)

	return err
}

// GetOptimizationRequest retrieves a route optimization request by ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error) {
	query := `
		SELECT id, organization_id, name, description, start_time, end_time,
		       start_time_window, end_time_window, constraints, stops, vehicles,
		       priorities, optimization_goals, created_at, updated_at, status, result
		FROM route_optimization_requests
		WHERE id = $1`

	var request models.RouteOptimizationRequest
	var constraintsJSON, stopsJSON, vehiclesJSON, prioritiesJSON, goalsJSON []byte
	var resultJSON []byte

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&request.ID,
		&request.OrganizationID,
		&request.Name,
		&request.Description,
		&request.StartTime,
		&request.EndTime,
		&request.StartTimeWindow,
		&request.EndTimeWindow,
		&constraintsJSON,
		&stopsJSON,
		&vehiclesJSON,
		&prioritiesJSON,
		&goalsJSON,
		&request.CreatedAt,
		&request.UpdatedAt,
		&request.Status,
		&resultJSON,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("route optimization request not found")
		}
		return nil, fmt.Errorf("failed to get route optimization request: %w", err)
	}

	// Unmarshal JSON fields
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
		return nil, fmt.Errorf("failed to unmarshal optimization goals: %w", err)
	}

	if resultJSON != nil {
		var result models.RouteOptimizationResult
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		request.Result = &result
	}

	return &request, nil
}

// UpdateOptimizationRequest updates an existing route optimization request
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	setClause += ", updated_at = $%d"
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE route_optimization_requests SET %s WHERE id = $%d", setClause, argIndex)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteOptimizationRequest deletes a route optimization request
func (r *RouteOptimizationRepositoryImpl) DeleteOptimizationRequest(ctx context.Context, id string) error {
	query := "DELETE FROM route_optimization_requests WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete route optimization request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("route optimization request not found")
	}

	return nil
}

// ListOptimizationRequests retrieves route optimization requests for an organization
func (r *RouteOptimizationRepositoryImpl) ListOptimizationRequests(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error) {
	countQuery := "SELECT COUNT(*) FROM route_optimization_requests WHERE organization_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count route optimization requests: %w", err)
	}

	query := `
		SELECT id, organization_id, name, description, start_time, end_time,
		       start_time_window, end_time_window, constraints, stops, vehicles,
		       priorities, optimization_goals, created_at, updated_at, status, result
		FROM route_optimization_requests
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list route optimization requests: %w", err)
	}
	defer rows.Close()

	var requests []*models.RouteOptimizationRequest
	for rows.Next() {
		var request models.RouteOptimizationRequest
		var constraintsJSON, stopsJSON, vehiclesJSON, prioritiesJSON, goalsJSON []byte
		var resultJSON []byte

		err := rows.Scan(
			&request.ID,
			&request.OrganizationID,
			&request.Name,
			&request.Description,
			&request.StartTime,
			&request.EndTime,
			&request.StartTimeWindow,
			&request.EndTimeWindow,
			&constraintsJSON,
			&stopsJSON,
			&vehiclesJSON,
			&prioritiesJSON,
			&goalsJSON,
			&request.CreatedAt,
			&request.UpdatedAt,
			&request.Status,
			&resultJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan route optimization request: %w", err)
		}

		// Unmarshal JSON fields
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
			return nil, 0, fmt.Errorf("failed to unmarshal optimization goals: %w", err)
		}

		if resultJSON != nil {
			var result models.RouteOptimizationResult
			if err := json.Unmarshal(resultJSON, &result); err != nil {
				log.Printf("Warning: failed to unmarshal result: %v", err)
			} else {
				request.Result = &result
			}
		}

		requests = append(requests, &request)
	}

	return requests, totalCount, nil
}

// CreateOptimizationResult creates a new route optimization result
func (r *RouteOptimizationRepositoryImpl) CreateOptimizationResult(ctx context.Context, result *models.RouteOptimizationResult) error {
	query := `
		INSERT INTO route_optimization_results (
			id, request_id, optimization_time, total_distance, total_time, 
			total_cost, total_vehicles_used, total_stops, routes, performance_metrics
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	routesJSON, err := json.Marshal(result.Routes)
	if err != nil {
		return fmt.Errorf("failed to marshal routes: %w", err)
	}

	metricsJSON, err := json.Marshal(result.PerformanceMetrics)
	if err != nil {
		return fmt.Errorf("failed to marshal performance metrics: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		result.ID,
		result.RequestID,
		result.OptimizationTime,
		result.TotalDistance,
		result.TotalTime,
		result.TotalCost,
		result.TotalVehiclesUsed,
		result.TotalStops,
		routesJSON,
		metricsJSON,
	)

	return err
}

// GetOptimizationResult retrieves a route optimization result by ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error) {
	query := `
		SELECT id, request_id, optimization_time, total_distance, total_time,
		       total_cost, total_vehicles_used, total_stops, routes, performance_metrics
		FROM route_optimization_results
		WHERE id = $1`

	var result models.RouteOptimizationResult
	var routesJSON, metricsJSON []byte

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&result.ID,
		&result.RequestID,
		&result.OptimizationTime,
		&result.TotalDistance,
		&result.TotalTime,
		&result.TotalCost,
		&result.TotalVehiclesUsed,
		&result.TotalStops,
		&routesJSON,
		&metricsJSON,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("route optimization result not found")
		}
		return nil, fmt.Errorf("failed to get route optimization result: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(routesJSON, &result.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}

	if err := json.Unmarshal(metricsJSON, &result.PerformanceMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal performance metrics: %w", err)
	}

	return &result, nil
}

// GetOptimizationResultByRequestID retrieves a route optimization result by request ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error) {
	query := `
		SELECT id, request_id, optimization_time, total_distance, total_time,
		       total_cost, total_vehicles_used, total_stops, routes, performance_metrics
		FROM route_optimization_results
		WHERE request_id = $1`

	var result models.RouteOptimizationResult
	var routesJSON, metricsJSON []byte

	row := r.db.QueryRow(ctx, query, requestID)
	err := row.Scan(
		&result.ID,
		&result.RequestID,
		&result.OptimizationTime,
		&result.TotalDistance,
		&result.TotalTime,
		&result.TotalCost,
		&result.TotalVehiclesUsed,
		&result.TotalStops,
		&routesJSON,
		&metricsJSON,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("route optimization result not found")
		}
		return nil, fmt.Errorf("failed to get route optimization result by request ID: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(routesJSON, &result.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}

	if err := json.Unmarshal(metricsJSON, &result.PerformanceMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal performance metrics: %w", err)
	}

	return &result, nil
}

// UpdateOptimizationResult updates an existing route optimization result
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationResult(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE route_optimization_results SET %s WHERE id = $%d", setClause, argIndex)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteOptimizationResult deletes a route optimization result
func (r *RouteOptimizationRepositoryImpl) DeleteOptimizationResult(ctx context.Context, id string) error {
	query := "DELETE FROM route_optimization_results WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete route optimization result: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("route optimization result not found")
	}

	return nil
}

// ListOptimizationResults retrieves route optimization results for a request
func (r *RouteOptimizationRepositoryImpl) ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error) {
	countQuery := "SELECT COUNT(*) FROM route_optimization_results WHERE request_id = $1"
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, requestID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count route optimization results: %w", err)
	}

	query := `
		SELECT id, request_id, optimization_time, total_distance, total_time,
		       total_cost, total_vehicles_used, total_stops, routes, performance_metrics
		FROM route_optimization_results
		WHERE request_id = $1
		ORDER BY optimization_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, requestID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list route optimization results: %w", err)
	}
	defer rows.Close()

	var results []*models.RouteOptimizationResult
	for rows.Next() {
		var result models.RouteOptimizationResult
		var routesJSON, metricsJSON []byte

		err := rows.Scan(
			&result.ID,
			&result.RequestID,
			&result.OptimizationTime,
			&result.TotalDistance,
			&result.TotalTime,
			&result.TotalCost,
			&result.TotalVehiclesUsed,
			&result.TotalStops,
			&routesJSON,
			&metricsJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan route optimization result: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(routesJSON, &result.Routes); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal routes: %w", err)
		}

		if err := json.Unmarshal(metricsJSON, &result.PerformanceMetrics); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal performance metrics: %w", err)
		}

		results = append(results, &result)
	}

	return results, totalCount, nil
}

// CreateOptimizationJob creates a new optimization job
func (r *RouteOptimizationRepositoryImpl) CreateOptimizationJob(ctx context.Context, job *models.OptimizationJob) error {
	query := `
		INSERT INTO optimization_jobs (
			id, request_id, status, created_at, started_at, completed_at, error, result_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		job.ID,
		job.RequestID,
		job.Status,
		job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
		job.Error,
		job.ResultID,
	)

	return err
}

// GetOptimizationJob retrieves an optimization job by ID
func (r *RouteOptimizationRepositoryImpl) GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error) {
	query := `
		SELECT id, request_id, status, created_at, started_at, completed_at, error, result_id
		FROM optimization_jobs
		WHERE id = $1`

	var job models.OptimizationJob

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&job.ID,
		&job.RequestID,
		&job.Status,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.Error,
		&job.ResultID,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("optimization job not found")
		}
		return nil, fmt.Errorf("failed to get optimization job: %w", err)
	}

	return &job, nil
}

// UpdateOptimizationJob updates an existing optimization job
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationJob(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE optimization_jobs SET %s WHERE id = $%d", setClause, argIndex)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// UpdateOptimizationJobStatus updates the status of an optimization job
func (r *RouteOptimizationRepositoryImpl) UpdateOptimizationJobStatus(ctx context.Context, id, status string, completedAt *time.Time, error *string) error {
	query := "UPDATE optimization_jobs SET status = $1, completed_at = $2, error = $3 WHERE id = $4"
	_, err := r.db.Exec(ctx, query, status, completedAt, error, id)
	return err
}

// ListOptimizationJobs retrieves optimization jobs for a request
func (r *RouteOptimizationRepositoryImpl) ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error) {
	args := []interface{}{requestID}
	argIndex := 2

	whereClause := "WHERE request_id = $1"
	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM optimization_jobs %s", whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count optimization jobs: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, request_id, status, created_at, started_at, completed_at, error, result_id
		FROM optimization_jobs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

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
			&job.ID,
			&job.RequestID,
			&job.Status,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.Error,
			&job.ResultID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan optimization job: %w", err)
		}

		jobs = append(jobs, &job)
	}

	return jobs, totalCount, nil
}

// HealthCheck performs a health check on the database
func (r *RouteOptimizationRepositoryImpl) HealthCheck(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *RouteOptimizationRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
