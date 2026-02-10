package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

// Dashboard constants
const (
	MockTotalVehicles      = 1248
	MockActiveVehicles     = 892
	MockTotalDrivers       = 156
	MockActiveTrips        = 342
	MockTotalDistanceToday = 12450.5
	MockFuelConsumedToday  = 1850.2
	MockAlertsCount        = 5
)

type DashboardHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

type Statistics struct {
	TotalVehicles      int `json:"total_vehicles"`
	ActiveVehicles     int `json:"active_vehicles"`
	TotalDrivers       int `json:"total_drivers"`
	ActiveTrips        int `json:"active_trips"`
	TotalDistanceToday float64 `json:"total_distance_today"`
	FuelConsumedToday  float64 `json:"fuel_consumed_today"`
	AlertsCount        int `json:"alerts_count"`
}

func NewDashboardHandlers(cfg *config.Config, logger *zap.Logger) *DashboardHandlers {
	return &DashboardHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// GetGlobalStats aggregates statistics across the platform
func (h *DashboardHandlers) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Global stats request")

	// In a complete implementation, this would aggregate data from:
	// 1. auth-service (total organizations, total users)
	// 2. vehicle-service (total vehicles, active trips)
	
	// Aggregated "Premium Mock" data aligned with FE Statistics interface
	stats := Statistics{
		TotalVehicles:      MockTotalVehicles,
		ActiveVehicles:     MockActiveVehicles,
		TotalDrivers:       MockTotalDrivers,
		ActiveTrips:        MockActiveTrips,
		TotalDistanceToday: MockTotalDistanceToday,
		FuelConsumedToday:  MockFuelConsumedToday,
		AlertsCount:        MockAlertsCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetOrganizationStats returns statistics for organization management
func (h *DashboardHandlers) GetOrganizationStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Organization stats request")

	// Mocking high-level tenant statistics
	stats := map[string]interface{}{
		"total_orgs":    12,
		"active_orgs":   9,
		"pending_orgs":  2,
		"inactive_orgs": 1,
		"total_users":   156,
		"avg_org_size":  13,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": stats,
	})
}
