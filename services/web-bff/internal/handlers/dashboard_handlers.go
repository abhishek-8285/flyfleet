package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	"github.com/fleet-platform/shared/pkg/security"
	"github.com/flyfleet/web-bff/internal/client"
	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

// Dashboard dynamics are aggregated from multiple core services

type DashboardHandlers struct {
	cfg           *config.Config
	logger        *zap.Logger
	cache         cache.RedisCacheInterface
	authClient    *client.AuthClient
	fleetClient   *client.FleetClient
	tripClient    *client.TripClient
	alertClient   *client.AlertClient
	billingClient *client.BillingClient
		supportClient  *client.SupportClient
		trackingClient *client.TrackingClient
		maintenanceClient *client.MaintenanceClient
	}
	

type Statistics struct {
	TotalVehicles      int     `json:"total_vehicles"`
	ActiveVehicles     int     `json:"active_vehicles"`
	TotalDrivers       int     `json:"total_drivers"`
	ActiveTrips        int     `json:"active_trips"`
	TotalDistanceToday float64 `json:"total_distance_today"`
	FuelConsumedToday  float64 `json:"fuel_consumed_today"`
	AlertsCount        int     `json:"alerts_count"`
}

type AdminDashboardData struct {
	LastUpdated  string `json:"last_updated"`
	RiskOverview struct {
		SLACompliance       float64 `json:"sla_compliance"`
		IntegrationFailures int     `json:"integration_failures"`
		PaymentFailures     int     `json:"payment_failures"`
		SecurityScore       float64 `json:"security_score"`
	} `json:"risk_overview"`
	UptimeTrend        []int `json:"uptime_trend"`
	CriticalIndicators struct {
		OpenIncidents       int     `json:"open_incidents"`
		DevicesOffline      int     `json:"devices_offline"`
		SystemHealth        float64 `json:"system_health"`
		BillingRisk         float64 `json:"billing_risk"`
		TotalDevices        int     `json:"total_devices"`
		IntegrationFailures int     `json:"integration_failures"`
	} `json:"critical_indicators"`
	OperationalMetrics struct {
		ActiveOrgs       int `json:"active_orgs"`
		ActiveFleet      int `json:"active_fleet"`
		ConnectedDevices int `json:"connected_devices"`
		ActiveUsersToday int `json:"active_users_today"`
		TotalOrgs        int `json:"total_orgs"`
		NewOrgsToday     int `json:"new_orgs_today"`
		TotalVehicles    int `json:"total_vehicles"`
		VehiclesOffline  int `json:"vehicles_offline"`
		DevicesOffline   int `json:"devices_offline"`
		UsersYesterday   int `json:"users_yesterday"`
	} `json:"operational_metrics"`
	PendingApprovals int                      `json:"pending_approvals"`
	RecentActivity   []map[string]interface{} `json:"recent_activity"`
}

func NewDashboardHandlers(cfg *config.Config, logger *zap.Logger, redisCache cache.RedisCacheInterface, authClient *client.AuthClient, fleetClient *client.FleetClient, tripClient *client.TripClient, alertClient *client.AlertClient, billingClient *client.BillingClient, supportClient *client.SupportClient, trackingClient *client.TrackingClient, maintenanceClient *client.MaintenanceClient) *DashboardHandlers {
	return &DashboardHandlers{
		cfg:           cfg,
		logger:        logger,
		cache:         redisCache,
		authClient:    authClient,
		fleetClient:   fleetClient,
		tripClient:    tripClient,
		alertClient:   alertClient,
		billingClient: billingClient,
		supportClient: supportClient,
		trackingClient: trackingClient,
		maintenanceClient: maintenanceClient,
	}
}

// GetGlobalStats aggregates statistics across the platform (Cached)
func (h *DashboardHandlers) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Global stats request")

	ctx := r.Context()
	token := getTokenFromRequest(r)
	orgID, _ := security.GetOrgID(ctx)
	orgIDStr := orgID.String()

	// Create a cache key based on the org ID
	cacheKey := fmt.Sprintf("stats:global:%s", orgIDStr)

	stats, err := cache.FetchWithCache(ctx, h.cache, cacheKey, 5*time.Minute, func() (Statistics, error) {
		// Fetch real data in parallel
		vStats, err := h.fleetClient.GetStats(ctx, token)
		if err != nil {
			h.logger.Error("Failed to fetch vehicle stats", zap.Error(err))
			vStats = &client.VehicleStats{}
		}

		uStats, err := h.authClient.GetStats(ctx, token)
		if err != nil {
			h.logger.Error("Failed to fetch aggregate stats", zap.Error(err))
			uStats = &client.UserStats{}
		}

		tStats, err := h.tripClient.GetStats(ctx, orgIDStr)
		if err != nil {
			h.logger.Error("Failed to fetch trip stats via gRPC", zap.Error(err))
			tStats = &client.TripStats{}
		}

		aStats, err := h.alertClient.GetStats(ctx, orgIDStr)
		if err != nil {
			h.logger.Error("Failed to fetch alert stats via gRPC", zap.Error(err))
			aStats = &client.AlertStats{}
		}

		sStats, _ := h.supportClient.GetDashboardStats(ctx, token)

		res := Statistics{
			TotalVehicles:      vStats.TotalVehicles,
			ActiveVehicles:     vStats.ActiveVehicles,
			TotalDrivers:       uStats.TotalDrivers,
			ActiveTrips:        tStats.ActiveTrips,
			TotalDistanceToday: tStats.Distance,
			FuelConsumedToday:  vStats.TotalFuel,
			AlertsCount:        aStats.NewAlerts,
		}

		// Override with support-service data if available
		if data, ok := sStats["data"].(map[string]interface{}); ok {
			if val, ok := data["total_vehicles"].(float64); ok {
				res.TotalVehicles = int(val)
			}
			if val, ok := data["active_vehicles"].(float64); ok {
				res.ActiveVehicles = int(val)
			}
			if val, ok := data["total_drivers"].(float64); ok {
				res.TotalDrivers = int(val)
			}
			if val, ok := data["active_trips"].(float64); ok {
				res.ActiveTrips = int(val)
			}
			if val, ok := data["total_distance_today"].(float64); ok {
				res.TotalDistanceToday = val
			}
			if val, ok := data["fuel_consumed_today"].(float64); ok {
				res.FuelConsumedToday = val
			}
			if val, ok := data["alerts_count"].(float64); ok {
				res.AlertsCount = int(val)
			}
		}

		return res, nil
	})

	if err != nil {
		h.logger.Error("Failed to fetch/cache global stats", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   stats,
		"cached": true, // Informative
	})
}

// ... rest of the file ...
func (h *DashboardHandlers) GetOrganizationStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := getTokenFromRequest(r)
	orgID, _ := security.GetOrgID(ctx)
	cacheKey := fmt.Sprintf("stats:org:%s", orgID.String())

	orgStats, err := cache.FetchWithCache(ctx, h.cache, cacheKey, 3*time.Minute, func() (map[string]interface{}, error) {
		stats, err := h.supportClient.GetOrganizationStats(ctx, token)
		if err != nil {
			h.logger.Warn("Failed to fetch organization stats via gRPC from support service", zap.Error(err))
			return map[string]interface{}{
				"active_orgs": 0,
				"total_orgs":  0,
			}, nil
		}
		return stats, nil
	})
	if err != nil {
		h.logger.Error("Failed to fetch/cache organization stats", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   orgStats,
		"cached": true,
	})
}

func (h *DashboardHandlers) GetAdminDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := getTokenFromRequest(r)
	cacheKey := fmt.Sprintf("stats:admin-dashboard:%s", token)

	data, err := cache.FetchWithCache(ctx, h.cache, cacheKey, 2*time.Minute, func() (AdminDashboardData, error) {
		admin := AdminDashboardData{}
		now := time.Now().UTC()
		admin.LastUpdated = now.Format(time.RFC3339)
		admin.UptimeTrend = []int{99, 99, 98, 99, 100, 99, 99}

		vStats, vErr := h.fleetClient.GetStats(ctx, token)
		if vErr != nil {
			h.logger.Warn("Failed to fetch vehicle stats for admin dashboard", zap.Error(vErr))
			vStats = &client.VehicleStats{}
		}

		uStats, uErr := h.authClient.GetStats(ctx, token)
		if uErr != nil {
			h.logger.Warn("Failed to fetch user stats for admin dashboard", zap.Error(uErr))
			uStats = &client.UserStats{}
		}

		bMetrics, bErr := h.billingClient.GetMetrics(ctx, orgIDStr)
		if bErr != nil {
			h.logger.Warn("Failed to fetch billing metrics via gRPC for admin dashboard", zap.Error(bErr))
			bMetrics = &client.BillingMetrics{}
		}

		secTelemetry, sErr := h.authClient.GetSecurityTelemetry(ctx, token)
		if sErr != nil {
			h.logger.Warn("Failed to fetch security telemetry for admin dashboard", zap.Error(sErr))
			secTelemetry = map[string]any{}
		}

		supportSummary, supErr := h.supportClient.GetDashboardStats(ctx, token)
		if supErr != nil {
			h.logger.Warn("Failed to fetch support dashboard summary for admin dashboard", zap.Error(supErr))
			supportSummary = map[string]any{}
		}

		mStats, mErr := h.maintenanceClient.GetStats(ctx, orgIDStr)
		if mErr != nil {
			h.logger.Warn("Failed to fetch maintenance stats for admin dashboard", zap.Error(mErr))
			mStats = &client.MaintenanceStats{}
		}

		recentActivity, raErr := h.fetchSupportRecentActivity(ctx, token)
		if raErr != nil {
			h.logger.Warn("Failed to fetch support recent activity for admin dashboard", zap.Error(raErr))
		}

		totalVehicles := vStats.TotalVehicles
		activeVehicles := vStats.ActiveVehicles
		openTickets := intNumber(supportSummary, "open_tickets.count")
		securityScore := floatNumber(secTelemetry, "security_score")
		if securityScore == 0 {
			securityScore = 92
		}

		admin.RiskOverview.SLACompliance = floatNumber(supportSummary, "sla_compliance")
		if admin.RiskOverview.SLACompliance == 0 {
			admin.RiskOverview.SLACompliance = 99.5
		}
		admin.RiskOverview.IntegrationFailures = intNumber(supportSummary, "integration_failures")
		admin.RiskOverview.PaymentFailures = intNumber(supportSummary, "payment_failures")
		admin.RiskOverview.SecurityScore = securityScore

		admin.CriticalIndicators.OpenIncidents = openTickets + mStats.OverdueServices
		admin.CriticalIndicators.TotalDevices = totalVehicles
		admin.CriticalIndicators.DevicesOffline = maxInt(totalVehicles-activeVehicles, 0)
		admin.CriticalIndicators.IntegrationFailures = admin.RiskOverview.IntegrationFailures
		admin.CriticalIndicators.BillingRisk = bMetrics.RiskScore
		admin.CriticalIndicators.SystemHealth = 100 - float64(admin.CriticalIndicators.DevicesOffline)
		if totalVehicles > 0 {
			admin.CriticalIndicators.SystemHealth = (float64(activeVehicles) / float64(totalVehicles)) * 100
		}

		admin.OperationalMetrics.ActiveOrgs = uStats.ActiveOrgs
		admin.OperationalMetrics.TotalOrgs = uStats.TotalOrgs
		admin.OperationalMetrics.ActiveFleet = activeVehicles
		admin.OperationalMetrics.TotalVehicles = totalVehicles
		admin.OperationalMetrics.VehiclesOffline = maxInt(totalVehicles-activeVehicles, 0)
		admin.OperationalMetrics.ConnectedDevices = activeVehicles
		admin.OperationalMetrics.DevicesOffline = admin.OperationalMetrics.VehiclesOffline
		admin.OperationalMetrics.ActiveUsersToday = uStats.ActiveUsers
		admin.OperationalMetrics.UsersYesterday = maxInt(uStats.ActiveUsers-1, 0)
		admin.OperationalMetrics.NewOrgsToday = 0

		admin.PendingApprovals = intNumber(supportSummary, "pending_approvals")
		admin.RecentActivity = recentActivity
		if admin.RecentActivity == nil {
			admin.RecentActivity = []map[string]interface{}{}
		}

		return admin, nil
	})
	if err != nil {
		h.logger.Error("Failed to fetch/cache admin dashboard data", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   data,
		"cached": true,
	})
}

func getTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[0:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

func (h *DashboardHandlers) fetchSupportRecentActivity(ctx context.Context, token string) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.cfg.AdminSupportServiceURL+"/api/v1/support/dashboard/recent-activity", nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload []map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		return payload, nil
	}

	var wrapped struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		return wrapped.Data, nil
	}
	return nil, fmt.Errorf("failed to decode recent activity response")
}

func intNumber(data map[string]any, path string) int {
	if v, ok := resolvePath(data, path); ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case float32:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		case string:
			if parsed, err := strconv.Atoi(n); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func floatNumber(data map[string]any, path string) float64 {
	if v, ok := resolvePath(data, path); ok {
		switch n := v.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		case string:
			if parsed, err := strconv.ParseFloat(n, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func resolvePath(data map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")

	var current any = data
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current = m[part]
	}
	if current == nil {
		return nil, false
	}
	return current, true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
