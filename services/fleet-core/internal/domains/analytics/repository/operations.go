package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *AnalyticsRepositoryImpl) GetVehicleDailyStats(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.VehicleDailyStats, string, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

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

	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

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

	if limit <= 0 {
		limit = 24
	}
	if limit > 168 {
		limit = 168
	} // Max 1 week hourly at once

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

	// Batch insert to avoid N+1 queries
	if len(vector) > 0 {
		var args []interface{}
		var placeholders []string
		for i, f := range vector {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
			args = append(args, f.OrganizationID, f.EntityType, f.EntityID, f.FeatureName, f.FeatureValue, f.FeatureTimestamp)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO ml_feature_store (organization_id, entity_type, entity_id, feature_name, feature_value, feature_timestamp) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err = tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert feature vectors: %w", err)
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
		"vehicle_daily_stats":  true,
		"driver_daily_stats":   true,
		"fleet_hourly_stats":   true,
		"prediction_results":   true,
		"analytics_query_logs": true,
		"ml_feature_store":     true,
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

	// Batch insert to avoid N+1 queries
	if len(vector) > 0 {
		var args []interface{}
		var placeholders []string
		for i, f := range vector {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
			args = append(args, f.OrganizationID, f.EntityType, f.EntityID, f.FeatureName, f.FeatureValue, f.FeatureVersion, f.FeatureLineage, f.FeatureSource, f.FeatureTimestamp)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO ml_feature_store (organization_id, entity_type, entity_id, feature_name, feature_value, feature_version, feature_lineage, feature_source, feature_timestamp) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err := r.db.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert versioned feature vectors: %w", err)
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

	// Query sizes for key analytics tables
	query := `
		SELECT
			relname as table_name,
			pg_total_relation_size(quote_ident(relname)::regclass) as size_bytes
		FROM pg_class
		WHERE relname IN (
			'vehicle_daily_stats',
			'driver_daily_stats',
			'fleet_hourly_stats',
			'analytics_query_logs',
			'ml_feature_store',
			'prediction_results',
			'model_deployments',
			'driver_risk_events',
			'materialized_fleet_reports'
		)
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query capacity stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var table string
		var size int64
		if err := rows.Scan(&table, &size); err != nil {
			return nil, fmt.Errorf("failed to scan capacity stat: %w", err)
		}
		stats[table] = size
	}

	return stats, nil
}
