package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type SupportOrgHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewSupportOrgHandlers(cfg *config.Config, logger *zap.Logger) *SupportOrgHandlers {
	return &SupportOrgHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// 1. ORGANIZATION BASIC INFO
func (h *SupportOrgHandlers) GetOrgSummary(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org summary", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"orgId":       orgID,
		"name":        "Acme Logistics Inc.",
		"plan":        "Enterprise Gold",
		"renewalDate": "2024-11-15",
		"status":      "ACTIVE",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 2. DEVICE STATISTICS
func (h *SupportOrgHandlers) GetOrgDeviceStats(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org device stats", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"total":   2450,
		"online":  2100,
		"offline": 350,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 3. INTEGRATION HEALTH
func (h *SupportOrgHandlers) GetOrgIntegrationHealth(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org integration health", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"successRate": 98.5,
		"latencyMs":   45,
		"errors24h":   12,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 4. FLEET OVERVIEW MAP DATA
func (h *SupportOrgHandlers) GetOrgFleetLocations(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org fleet locations", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"regions": []map[string]interface{}{
			{
				"name":          "North America",
				"activePercent": 98,
			},
			{
				"name":          "Europe",
				"activePercent": 85,
			},
			{
				"name":          "APAC",
				"activePercent": 92,
			},
		},
		"criticalDevices": 5,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 5. DIAGNOSTIC EVENTS
func (h *SupportOrgHandlers) GetOrgRecentDiagnostics(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org recent diagnostics", zap.String("orgId", orgID))

	res := []map[string]interface{}{
		{
			"deviceId": "DEV-2938",
			"issue":    "GPS Signal Lost",
			"firmware": "v4.2.1",
			"status":   "INVESTIGATING",
			"time":     time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		},
		{
			"deviceId": "DEV-9912",
			"issue":    "Engine Overheat",
			"firmware": "v4.1.9",
			"status":   "CRITICAL",
			"time":     time.Now().Add(-42 * time.Minute).Format(time.RFC3339),
		},
		{
			"deviceId": "DEV-1022",
			"issue":    "Battery Low",
			"firmware": "v4.2.1",
			"status":   "RESOLVED",
			"time":     time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 6. VIEW ALL LOGS
func (h *SupportOrgHandlers) GetOrgDiagnosticLogs(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org diagnostic logs", zap.String("orgId", orgID))

	// Mocking query params filter
	res := map[string]interface{}{
		"logs": []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"deviceId":  "DEV-2938",
				"type":      "NETWORK",
				"message":   "Lost connection to gateway-01",
			},
		},
		"total": 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 7. ACTIVE ALERTS
func (h *SupportOrgHandlers) GetOrgAlerts(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org alerts", zap.String("orgId", orgID))

	res := []map[string]interface{}{
		{
			"alertId":  "INC-9921",
			"title":    "High Latency Detected",
			"severity": "CRITICAL",
			"timeAgo":  "14m",
			"desc":     "US-East Region experiencing 400ms+ latency.",
		},
		{
			"alertId":  "WRN-4421",
			"title":    "API Rate Limit Approaching",
			"severity": "WARNING",
			"timeAgo":  "32m",
			"desc":     "90% of hourly quota consumed.",
		},
		{
			"alertId":  "INF-1102",
			"title":    "Firmware Update Pending",
			"severity": "INFO",
			"timeAgo":  "2h",
			"desc":     "240 devices scheduled for update at 02:00 UTC.",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 8. ORG CONTACTS
func (h *SupportOrgHandlers) GetOrgContacts(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org contacts", zap.String("orgId", orgID))

	res := []map[string]interface{}{
		{
			"name":  "Jane Doe",
			"role":  "Admin",
			"email": "jane@acme.com",
			"initials": "JD",
		},
		{
			"name":  "Robert King",
			"role":  "Tech Lead",
			"email": "rob@acme.com",
			"initials": "RK",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 9. FULL DIRECTORY
func (h *SupportOrgHandlers) GetOrgUsers(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org users", zap.String("orgId", orgID))

	res := []map[string]interface{}{
		{
			"userId": "USR-101",
			"name":   "Jane Doe",
			"email":  "jane@acme.com",
			"role":   "Admin",
		},
		{
			"userId": "USR-102",
			"name":   "Robert King",
			"email":  "rob@acme.com",
			"role":   "Tech Lead",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 10. IMPERSONATION
func (h *SupportOrgHandlers) ImpersonateOrg(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Impersonating org", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"sessionToken": "abc123mocktoken",
		"mode":         "READ_ONLY",
		"targetOrg":    orgID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 11. TROUBLESHOOT CONFIGS
func (h *SupportOrgHandlers) GetOrgConfigDiagnostics(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org config diagnostics", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"issues": []string{},
		"recommendations": []string{
			"Optimize gateway selection for APAC region",
			"Enable advanced compression for DEV-prefix devices",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 12. ACTIVITY LOGS TAB
func (h *SupportOrgHandlers) GetOrgActivity(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org activity logs", zap.String("orgId", orgID))

	res := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"user":      "Jane Doe",
			"action":    "Modified Alert Policy: High Latency",
		},
		{
			"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"user":      "Robert King",
			"action":    "Deactivated Device: DEV-4412",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 13. TECHNICAL CONFIG TAB
func (h *SupportOrgHandlers) GetOrgConfig(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Getting org technical config", zap.String("orgId", orgID))

	res := map[string]interface{}{
		"gateway":     "aws-us-east-1",
		"environment": "production",
		"version":     "v4.2.0",
		"features": []string{
			"real-time-tracking",
			"advanced-diagnostics",
			"fleet-impersonation",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 14. DEVICE SEARCH (MAP FILTER)
func (h *SupportOrgHandlers) SearchOrgDevices(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	query := r.URL.Query().Get("q")
	h.logger.Info("Searching org devices", zap.String("orgId", orgID), zap.String("query", query))

	res := []map[string]interface{}{
		{
			"deviceId": "DEV-2938",
			"status":   "ONLINE",
			"location": "New York, US",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 15. ALERT ACTIONS
func (h *SupportOrgHandlers) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("alertId")
	h.logger.Info("Acknowledging alert", zap.String("alertId", alertID))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Alert %s acknowledged"}`, alertID)))
}

// 16. EXPORT DATA
func (h *SupportOrgHandlers) ExportOrgData(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	h.logger.Info("Exporting org data", zap.String("orgId", orgID))

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=org_%s_export.csv", orgID))
	w.Write([]byte("Org ID,Name,Type\n"))
	w.Write([]byte(fmt.Sprintf("%s,Acme Logistics Inc.,SUMMARY\n", orgID)))
}
