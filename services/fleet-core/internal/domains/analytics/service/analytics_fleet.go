package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/analytics/models"
	"github.com/google/uuid"
)

// Fleet Analytics & Stats Methods
func (s *AnalyticsServiceImpl) CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error {
	if err := validateFleetAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Calculate KPIs if context/signals allow, otherwise use provided
	if analytics.AverageFuelEfficiency == 0 && analytics.TotalDistance > 0 && analytics.TotalFuelConsumption > 0 {
		analytics.AverageFuelEfficiency = s.calculateFuelEfficiency(analytics.TotalDistance, analytics.TotalFuelConsumption)
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

	return s.repo.CreateFleetAnalytics(ctx, analytics)
}

// GetFleetAnalytics retrieves fleet analytics by ID
func (s *AnalyticsServiceImpl) GetFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.FleetAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.FleetAnalytics

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{"id": id}
		s.logQuery(ctx, orgID, "get_fleet_analytics", params, time.Since(start), rowCount, false, err)
	}()

	if id == "" {
		err = errors.New("ID is required")
		return nil, err
	}

	res, err = s.repo.GetFleetAnalytics(ctx, id, orgID)
	return res, err
}

// GetFleetAnalyticsByPeriod retrieves fleet analytics by organization and date period
func (s *AnalyticsServiceImpl) GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
	start := time.Now()
	var err error
	var res *models.FleetAnalytics
	cacheHit := false

	defer func() {
		rowCount := 0
		if res != nil {
			rowCount = 1
		}
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		}
		s.logQuery(ctx, orgID, "get_fleet_analytics_period", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, err
	}
	if startDate.IsZero() || endDate.IsZero() {
		err = errors.New("start and end dates are required")
		return nil, err
	}
	if startDate.After(endDate) {
		err = errors.New("start date cannot be after end date")
		return nil, err
	}

	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "fleet_analytics_period", params)

		if found, _ := s.cache.Get(ctx, cacheKey, &res); found {
			cacheHit = true
			return res, nil
		}

		res, err = s.repo.GetFleetAnalyticsByPeriod(ctx, orgID, startDate, endDate)
		if err == nil && res != nil {
			_ = s.cache.Set(ctx, cacheKey, res, 0)
		}
		return res, err
	}

	res, err = s.repo.GetFleetAnalyticsByPeriod(ctx, orgID, startDate, endDate)
	return res, err
}

// UpdateFleetAnalytics updates existing fleet analytics
func (s *AnalyticsServiceImpl) UpdateFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateFleetAnalytics(ctx, id, orgID, updates)
}

// DeleteFleetAnalytics deletes fleet analytics
func (s *AnalyticsServiceImpl) DeleteFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteFleetAnalytics(ctx, id, orgID)
}

// ListFleetAnalytics retrieves fleet analytics with filtering, validation and caching
func (s *AnalyticsServiceImpl) ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, cursor string, limit, offset int) ([]*models.FleetAnalytics, int, error) {
	start := time.Now()
	var err error
	var data []*models.FleetAnalytics
	var total int
	cacheHit := false

	defer func() {
		rowCount := len(data)
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if startDate != nil {
			params["start_date"] = startDate.Format(time.RFC3339)
		}
		if endDate != nil {
			params["end_date"] = endDate.Format(time.RFC3339)
		}
		s.logQuery(ctx, orgID, "list_fleet_analytics", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
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

	if startDate != nil && endDate != nil {
		if err = validateQueryPeriod(*startDate, *endDate); err != nil {
			return nil, 0, err
		}
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}
		if startDate != nil {
			params["start_date"] = startDate.Format(time.RFC3339)
		}
		if endDate != nil {
			params["end_date"] = endDate.Format(time.RFC3339)
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "list_fleet_analytics", params)
		type listResult struct {
			Data  []*models.FleetAnalytics
			Total int
		}
		var cachedRes listResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cachedRes); found {
			cacheHit = true
			data = cachedRes.Data
			total = cachedRes.Total
			return data, total, nil
		}

		data, total, err = s.repo.ListFleetAnalytics(ctx, orgID, startDate, endDate, cursor, limit, offset)
		if err == nil {
			_ = s.cache.Set(ctx, cacheKey, listResult{Data: data, Total: total}, 0)
		}
		return data, total, err
	}

	data, total, err = s.repo.ListFleetAnalytics(ctx, orgID, startDate, endDate, cursor, limit, offset)
	return data, total, err
}

// CreateDriverBehaviorAnalytics creates new driver behavior analytics with validation
func (s *AnalyticsServiceImpl) GetVehicleDailyStats(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.VehicleDailyStats, string, error) {
	start := time.Now()
	var err error
	var res []*models.VehicleDailyStats
	var nextCursor string
	cacheHit := false
	rowCount := 0

	defer func() {
		params := map[string]interface{}{
			"vehicle_id": vehicleID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		s.logQuery(ctx, orgID, "get_vehicle_daily_stats", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if vehicleID == uuid.Nil {
		err = errors.New("vehicle ID is required")
		return nil, "", err
	}
	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"vehicle_id": vehicleID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "vehicle", vehicleID.String(), "vehicle_daily_stats", params)

		type cachedResult struct {
			Data   []*models.VehicleDailyStats `json:"data"`
			Cursor string                      `json:"cursor"`
		}
		var cached cachedResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cached); found {
			cacheHit = true
			res = cached.Data
			nextCursor = cached.Cursor
			rowCount = len(res)
			return res, nextCursor, nil
		}

		res, nextCursor, err = s.repo.GetVehicleDailyStats(ctx, vehicleID, orgID, startDate, endDate, cursor, limit)
		if err == nil {
			rowCount = len(res)
			_ = s.cache.Set(ctx, cacheKey, cachedResult{Data: res, Cursor: nextCursor}, 0)
		}
		return res, nextCursor, err
	}

	res, nextCursor, err = s.repo.GetVehicleDailyStats(ctx, vehicleID, orgID, startDate, endDate, cursor, limit)
	rowCount = len(res)
	return res, nextCursor, err
}

// GetDriverDailyStats retrieves daily stats for a driver with cursor pagination
func (s *AnalyticsServiceImpl) GetDriverDailyStats(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.DriverDailyStats, string, error) {
	start := time.Now()
	var err error
	var res []*models.DriverDailyStats
	var nextCursor string
	cacheHit := false
	rowCount := 0

	defer func() {
		params := map[string]interface{}{
			"driver_id":  driverID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		s.logQuery(ctx, orgID, "get_driver_daily_stats", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if driverID == uuid.Nil {
		err = errors.New("driver ID is required")
		return nil, "", err
	}
	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"driver_id":  driverID.String(),
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "driver", driverID.String(), "driver_daily_stats", params)

		type cachedResult struct {
			Data   []*models.DriverDailyStats `json:"data"`
			Cursor string                     `json:"cursor"`
		}
		var cached cachedResult
		if found, _ := s.cache.Get(ctx, cacheKey, &cached); found {
			cacheHit = true
			res = cached.Data
			nextCursor = cached.Cursor
			rowCount = len(res)
			return res, nextCursor, nil
		}

		res, nextCursor, err = s.repo.GetDriverDailyStats(ctx, driverID, orgID, startDate, endDate, cursor, limit)
		if err == nil {
			rowCount = len(res)
			_ = s.cache.Set(ctx, cacheKey, cachedResult{Data: res, Cursor: nextCursor}, 0)
		}
		return res, nextCursor, err
	}

	res, nextCursor, err = s.repo.GetDriverDailyStats(ctx, driverID, orgID, startDate, endDate, cursor, limit)
	rowCount = len(res)
	return res, nextCursor, err
}

// GetFleetHourlyStats retrieves pre-aggregated hourly stats for the fleet
func (s *AnalyticsServiceImpl) GetFleetHourlyStats(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.FleetHourlyStats, string, error) {
	start := time.Now()
	var err error
	var res []*models.FleetHourlyStats
	var nextCursor string
	cacheHit := false
	rowCount := 0

	defer func() {
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		s.logQuery(ctx, orgID, "get_fleet_hourly_stats", params, time.Since(start), rowCount, cacheHit, err)
	}()

	if orgID == uuid.Nil {
		err = errors.New("organization ID is required")
		return nil, "", err
	}
	if err = validateQueryPeriod(startDate, endDate); err != nil {
		return nil, "", err
	}

	if s.cache != nil {
		params := map[string]interface{}{
			"start_date": startDate.Format(time.RFC3339),
			"end_date":   endDate.Format(time.RFC3339),
			"cursor":     cursor,
			"limit":      limit,
		}
		cacheKey := s.cache.GenerateKey(ctx, orgID.String(), "fleet", "", "fleet_hourly_stats", params)

		type cachedResult struct {
			Data   []*models.FleetHourlyStats `json:"data"`
			Cursor string                     `json:"cursor"`
		}
		var cached cachedResult

		if found, _ := s.cache.Get(ctx, cacheKey, &cached); found {
			cacheHit = true
			res = cached.Data
			nextCursor = cached.Cursor
			rowCount = len(res)
			return res, nextCursor, nil
		}

		res, nextCursor, err = s.repo.GetFleetHourlyStats(ctx, orgID, startDate, endDate, cursor, limit)
		if err == nil {
			rowCount = len(res)
			_ = s.cache.Set(ctx, cacheKey, cachedResult{Data: res, Cursor: nextCursor}, 0)
		}
		return res, nextCursor, err
	}

	res, nextCursor, err = s.repo.GetFleetHourlyStats(ctx, orgID, startDate, endDate, cursor, limit)
	rowCount = len(res)
	return res, nextCursor, err
}

// InvalidateVehicleCache invalidates all analytics cache related to a vehicle
