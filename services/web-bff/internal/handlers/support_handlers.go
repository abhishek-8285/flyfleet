package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type SupportHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewSupportHandlers(cfg *config.Config, logger *zap.Logger) *SupportHandlers {
	return &SupportHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// --- Dashboard APIs ---

func (h *SupportHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard stats request")

	// Mock Data
	stats := map[string]interface{}{
		"open_tickets": map[string]interface{}{
			"count": 42,
			"trend": -12,
		},
		"sla_breaches": map[string]interface{}{
			"count":    3,
			"priority": "Critical",
		},
		"active_errors": map[string]interface{}{
			"count": 18,
			"scope": "Across fleet",
		},
		"avg_resolution": map[string]interface{}{
			"time":  "2.4h",
			"trend": 15, // 15% faster
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *SupportHandlers) GetDashboardTrends(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard trends request")
	// Mock Data for chart
	trends := []int{20, 45, 30, 60, 40, 75, 50}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trends)
}

func (h *SupportHandlers) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard recent activity request")
	// Mock Data for table
	activity := []map[string]interface{}{
		{
			"id":           "T-9024",
			"subject":      "Vehicle telemetry sync failure",
			"reporter":     "Fleet Ops",
			"priority":     "High",
			"assigned_to":  "John D.",
			"last_updated": "10 mins ago",
			"status":       "IN PROGRESS",
		},
		{
			"id":           "T-9023",
			"subject":      "Driver app login issue",
			"reporter":     "Driver Support",
			"priority":     "Medium",
			"assigned_to":  "Anna S.",
			"last_updated": "24 mins ago",
			"status":       "PENDING",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}

// --- User Console APIs ---

func (h *SupportHandlers) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support user profile request")
	// Mock Data
	profile := map[string]interface{}{
		"id":          "USR-882910",
		"name":        "Sarah Jenkins",
		"email":       "sarah.j@fleetcorp.com",
		"role":        "Regional Manager",
		"region":      "EMEA",
		"status":      "Active (Verified)",
		"member_since": "Feb 12, 2021",
		"mfa_status":  "Enforced",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *SupportHandlers) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support user sessions request")
	// Mock Data
	sessions := []map[string]interface{}{
		{
			"device":   "MacBook Pro",
			"os":       "macOS",
			"location": "London, UK",
			"ip":       "82.11.45.22",
			"browser":  "Chrome 118",
			"current":  true,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// --- Org Console APIs ---

func (h *SupportHandlers) GetOrgOverview(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support org overview request")
	// Mock Data
	org := map[string]interface{}{
		"name":           "Acme Logistics Inc.",
		"plan":           "Enterprise Gold",
		"renewal":        "Nov 15, 2024",
		"total_devices":  2450,
		"online_devices": 2100,
		"offline_devices": 350,
		"api_health":     98.5,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// --- Incident APIs ---

func (h *SupportHandlers) GetIncidents(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support incidents request")
	// Mock Data
	incidents := []map[string]interface{}{
		{
			"id":           "INC-4921",
			"priority":     "CRITICAL",
			"status":       "Open",
			"subject":      "Database latency spike US-East",
			"assigned_to":  "Unassigned",
			"organization": "Global Fleet Ops",
		},
		{
			"id":           "INC-4920",
			"priority":     "HIGH",
			"status":       "Investigating",
			"subject":      "Authentication failures for API v2",
			"assigned_to":  "Sarah J.",
			"organization": "Acme Logistics",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incidents)
}

// --- Logs APIs ---

func (h *SupportHandlers) GetSystemLogs(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support system logs request")
	// Mock Data
	logs := []map[string]interface{}{
		{
			"timestamp": "2023-10-24 14:32:01.452",
			"severity":  "ERROR",
			"module":    "AuthService",
			"message":   "Failed to validate token signature: key expired",
			"trace_id":  "tid_8f92a1",
		},
		{
			"timestamp": "2023-10-24 14:31:58.102",
			"severity":  "WARN",
			"module":    "RateLimiter",
			"message":   "User 4921 approaching API limit (95%)",
			"trace_id":  "tid_7b12c4",
		},
		{
			"timestamp": "2023-10-24 14:31:55.004",
			"severity":  "INFO",
			"module":    "FleetManager",
			"message":   "Vehicle status updated: V-8821 -> Active",
			"trace_id":  "tid_6d88e2",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
