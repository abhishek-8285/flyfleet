package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
)

// Driver Behavior & Predictive Maintenance Methods
func (s *AnalyticsServiceImpl) CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error {
	if err := validateDriverBehaviorAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Apply Safety Scoring Model from Specification
	harshEvents := analytics.HardBrakingCount + analytics.AggressiveAccelerationCount + analytics.SharpTurnCount
	score, risk := s.calculateSafetyScore(analytics.OverSpeedingCount, harshEvents, analytics.FatigueEventCount, analytics.ViolationCount)
	analytics.SafetyScore = score
	analytics.RiskLevel = risk

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreateDriverBehaviorAnalytics(ctx, analytics)
}

// GetDriverBehaviorAnalytics retrieves driver behavior analytics by ID
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalytics(ctx context.Context, id string, tenantID uuid.UUID) (*models.DriverBehaviorAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.DriverBehaviorAnalytics
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{
			"id": id,
		}
		s.logQuery(ctx, tenantID, "get_driver_behavior_analytics", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetDriverBehaviorAnalytics(ctx, id, tenantID)
	return res, err
}

// GetDriverBehaviorAnalyticsByDriver retrieves driver behavior analytics by driver ID
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.DriverBehaviorAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"driver_id": driverID.String(),
			"limit":     limit,
			"offset":    offset,
		}
		s.logQuery(ctx, tenantID, "get_driver_behavior_by_driver", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if driverID == uuid.Nil {
		err = errors.New("driver ID is required")
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	data, total, err = s.repo.GetDriverBehaviorAnalyticsByDriver(ctx, driverID, tenantID, limit, offset)
	return data, total, err
}

// GetDriverBehaviorAnalyticsByOrg retrieves driver behavior analytics by organization
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalyticsByOrg(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.DriverBehaviorAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		s.logQuery(ctx, tenantID, "get_driver_behavior_by_org", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if tenantID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	data, total, err = s.repo.GetDriverBehaviorAnalyticsByOrg(ctx, tenantID, limit, offset)
	return data, total, err
}

// UpdateDriverBehaviorAnalytics updates existing driver behavior analytics
func (s *AnalyticsServiceImpl) UpdateDriverBehaviorAnalytics(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateDriverBehaviorAnalytics(ctx, id, tenantID, updates)
}

// DeleteDriverBehaviorAnalytics deletes driver behavior analytics
func (s *AnalyticsServiceImpl) DeleteDriverBehaviorAnalytics(ctx context.Context, id string, tenantID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteDriverBehaviorAnalytics(ctx, id, tenantID)
}

// ListDriverBehaviorAnalytics retrieves driver behavior analytics with filtering, validation and caching
func (s *AnalyticsServiceImpl) ListDriverBehaviorAnalytics(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd *time.Time, cursor string, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.DriverBehaviorAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if periodStart != nil {
			params["period_start"] = periodStart.Format(time.RFC3339)
		}
		if periodEnd != nil {
			params["period_end"] = periodEnd.Format(time.RFC3339)
		}
		s.logQuery(ctx, tenantID, "list_driver_behavior", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if tenantID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	if periodStart != nil && periodEnd != nil {
		if err = validateQueryPeriod(*periodStart, *periodEnd); err != nil {
			return nil, 0, err
		}
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if periodStart != nil {
			params["period_start"] = periodStart.Format(time.RFC3339)
		}
		if periodEnd != nil {
			params["period_end"] = periodEnd.Format(time.RFC3339)
		}
		cacheKey := s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "list_driver_behavior", params)
		type listResult struct {
			Data  []*models.DriverBehaviorAnalytics
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListDriverBehaviorAnalytics(ctx, tenantID, periodStart, periodEnd, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListDriverBehaviorAnalytics(ctx, tenantID, periodStart, periodEnd, cursor, limit, offset)
	return data, total, err
}

// CreatePredictiveMaintenanceAnalytics creates new predictive maintenance analytics with validation
func (s *AnalyticsServiceImpl) CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error {
	if err := validatePredictiveMaintenanceAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreatePredictiveMaintenanceAnalytics(ctx, analytics)
}

// GetPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics by ID
func (s *AnalyticsServiceImpl) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string, tenantID uuid.UUID) (*models.PredictiveMaintenanceAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.PredictiveMaintenanceAnalytics
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{
			"id": id,
		}
		s.logQuery(ctx, tenantID, "get_predictive_maintenance_analytics", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetPredictiveMaintenanceAnalytics(ctx, id, tenantID)
	return res, err
}

// GetPredictiveMaintenanceAnalyticsByVehicle retrieves predictive maintenance analytics by vehicle ID
func (s *AnalyticsServiceImpl) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictiveMaintenanceAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"vehicle_id": vehicleID.String(),
			"limit":      limit,
			"offset":     offset,
		}
		s.logQuery(ctx, tenantID, "get_predictive_maintenance_by_vehicle", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if vehicleID == uuid.Nil {
		err = errors.New("vehicle ID is required")
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	data, total, err = s.repo.GetPredictiveMaintenanceAnalyticsByVehicle(ctx, vehicleID, tenantID, limit, offset)
	return data, total, err
}

// UpdatePredictiveMaintenanceAnalytics updates existing predictive maintenance analytics
func (s *AnalyticsServiceImpl) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, tenantID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictiveMaintenanceAnalytics(ctx, id, tenantID, updates)
}

// DeletePredictiveMaintenanceAnalytics deletes predictive maintenance analytics
func (s *AnalyticsServiceImpl) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string, tenantID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictiveMaintenanceAnalytics(ctx, id, tenantID)
}

// ListPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics with filtering, validation and caching
func (s *AnalyticsServiceImpl) ListPredictiveMaintenanceAnalytics(ctx context.Context, tenantID uuid.UUID, vehicleID *uuid.UUID, cursor string, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.PredictiveMaintenanceAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if vehicleID != nil {
			params["vehicle_id"] = vehicleID.String()
		}
		s.logQuery(ctx, tenantID, "list_predictive_maintenance", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if tenantID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > maxSyncRows {
		limit = maxSyncRows
	}
	if offset < 0 {
		offset = 0
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if vehicleID != nil {
			params["vehicle_id"] = vehicleID.String()
		}
		var cacheKey string
		if vehicleID != nil {
			cacheKey = s.cache.GenerateKey(ctx, tenantID.String(), "vehicle", vehicleID.String(), "list_predictive_maintenance", params)
		} else {
			cacheKey = s.cache.GenerateKey(ctx, tenantID.String(), "fleet", "", "list_predictive_maintenance", params)
		}
		type listResult struct {
			Data  []*models.PredictiveMaintenanceAnalytics
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListPredictiveMaintenanceAnalytics(ctx, tenantID, vehicleID, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListPredictiveMaintenanceAnalytics(ctx, tenantID, vehicleID, cursor, limit, offset)
	return data, total, err
}

// HealthCheck performs a health check on the service and its dependencies
