package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type SupportHandlers struct {
	cfg              *config.Config
	logger           *zap.Logger
	incidents        map[string]map[string]interface{}
	mu               sync.RWMutex
	supportServiceURL string
}

func NewSupportHandlers(cfg *config.Config, logger *zap.Logger) *SupportHandlers {
	h := &SupportHandlers{
		cfg:              cfg,
		logger:           logger,
		incidents:        make(map[string]map[string]interface{}),
		supportServiceURL: "http://localhost:8082", // Will be configurable
	}
	h.seedIncidents()
	return h
}

func (h *SupportHandlers) seedIncidents() {
	h.mu.Lock()
	defer h.mu.Unlock()

	seeds := []map[string]interface{}{
		{
			"id":           "INC-4921",
			"priority":     "CRITICAL",
			"status":       "Open",
			"subject":      "Database latency spike US-East",
			"description":  "Persistent high latency detected in RDS cluster for the US-East region. Affecting 15% of write operations.",
			"assigned_to":  "Unassigned",
			"organization": "Global Fleet Ops",
			"created_at":   "2023-10-24 14:00:00",
			"reporter":     "System Monitor",
			"updates": []map[string]interface{}{
				{
					"time":    "14:05",
					"message": "Initial investigation started by Platform Team.",
					"user":    "System",
				},
				{
					"time":    "14:15",
					"message": "Scale-up procedure initiated for read replicas.",
					"user":    "Alice W.",
				},
			},
		},
		{
			"id":           "INC-4920",
			"priority":     "HIGH",
			"status":       "Investigating",
			"subject":      "Authentication failures for API v2",
			"assigned_to":  "Sarah J.",
			"organization": "Acme Logistics",
			"created_at":   "2023-10-24 13:45:00",
			"description":  "Users reporting 502/504 errors intermittently on v2 endpoints.",
			"updates":      []map[string]interface{}{},
		},
		{
			"id":           "T-9024",
			"priority":     "HIGH",
			"status":       "IN PROGRESS",
			"subject":      "Vehicle telemetry sync failure",
			"assigned_to":  "John D.",
			"organization": "Fleet Ops",
			"created_at":   "2023-10-24 13:30:00",
			"description":  "Missing heartbeats from units in the southern corridor.",
			"updates":      []map[string]interface{}{},
		},
		{
			"id":           "T-9023",
			"priority":     "MEDIUM",
			"status":       "PENDING",
			"subject":      "Driver app login issue",
			"assigned_to":  "Anna S.",
			"organization": "Driver Support",
			"created_at":   "2023-10-24 13:15:00",
			"description":  "New drivers unable to complete initial registration flow.",
			"updates":      []map[string]interface{}{},
		},
	}

	for _, s := range seeds {
		h.incidents[s["id"].(string)] = s
	}
}

// --- Dashboard APIs ---

func (h *SupportHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard stats request")

	// Proxy to support service
	proxyURL := fmt.Sprintf("%s/api/v1/support/dashboard/summary", h.supportServiceURL)
	resp, err := http.Get(proxyURL)
	if err != nil {
		h.logger.Error("Failed to proxy to support service", zap.Error(err))
		// Fallback to mock data
		h.sendMockDashboardStats(w)
		return
	}
	defer resp.Body.Close()

	// Forward response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *SupportHandlers) sendMockDashboardStats(w http.ResponseWriter) {
	// Mock Data as fallback
	h.mu.RLock()
	openCount := 0
	for _, inc := range h.incidents {
		if inc["status"] != "RESOLVED" {
			openCount++
		}
	}
	h.mu.RUnlock()

	stats := map[string]interface{}{
		"open_tickets": map[string]interface{}{
			"count": openCount,
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


func (h *SupportHandlers) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard recent activity request")
	
	// Proxy to support service
	proxyURL := fmt.Sprintf("%s/api/v1/support/dashboard/recent-activity", h.supportServiceURL)
	resp, err := http.Get(proxyURL)
	if err != nil {
		h.logger.Error("Failed to proxy to support service for recent activity", zap.Error(err))
		// Fallback to mock data
		h.sendMockRecentActivity(w)
		return
	}
	defer resp.Body.Close()

	// Forward response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *SupportHandlers) sendMockRecentActivity(w http.ResponseWriter) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Use actual incidents data instead of hardcoded mock data
	activity := make([]map[string]interface{}, 0, len(h.incidents))
	for _, incident := range h.incidents {
		// Only include tickets that are not RESOLVED for recent activity
		if incident["status"] != "RESOLVED" {
			activity = append(activity, map[string]interface{}{
				"id":           incident["id"],
				"subject":      incident["subject"],
				"reporter":     incident["organization"],
				"priority":     incident["priority"],
				"assigned_to":  incident["assigned_to"],
				"last_updated": "10 mins ago", // TODO: Calculate real timestamp
				"status":       incident["status"],
			})
		}
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
	
	// Proxy to support service
	proxyURL := fmt.Sprintf("%s/api/v1/support/incidents", h.supportServiceURL)
	resp, err := http.Get(proxyURL)
	if err != nil {
		h.logger.Error("Failed to proxy to support service for incidents", zap.Error(err))
		// Fallback to mock data
		h.sendMockIncidents(w)
		return
	}
	defer resp.Body.Close()

	// Forward response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *SupportHandlers) sendMockIncidents(w http.ResponseWriter) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	incidentsList := make([]map[string]interface{}, 0, len(h.incidents))
	for _, inc := range h.incidents {
		incidentsList = append(incidentsList, inc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incidentsList)
}

func (h *SupportHandlers) GetIncidentDetail(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support incident detail request")
	id := r.URL.Query().Get("id")

	h.mu.RLock()
	incident, ok := h.incidents[id]
	h.mu.RUnlock()

	if !ok {
		// Fallback for demo: if ID not found, return the first one
		h.mu.RLock()
		for _, v := range h.incidents {
			incident = v
			break
		}
		h.mu.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

func (h *SupportHandlers) CreateIncident(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support create incident request")

	var body struct {
		Subject      string `json:"subject"`
		Description  string `json:"description"`
		Priority     string `json:"priority"`
		Organization string `json:"organization"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	newID := fmt.Sprintf("INC-%d", 5000+len(h.incidents))
	newIncident := map[string]interface{}{
		"id":           newID,
		"priority":     body.Priority,
		"status":       "Open",
		"subject":      body.Subject,
		"description":  body.Description,
		"assigned_to":  "Unassigned",
		"organization": body.Organization,
		"created_at":   time.Now().Format("2006-01-02 15:04:05"),
		"reporter":     "Abhishek (Support)",
		"updates":      []map[string]interface{}{},
	}

	h.incidents[newID] = newIncident

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newIncident)
}

func (h *SupportHandlers) PostComment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support post comment request")
	id := r.URL.Query().Get("id")

	var body struct {
		Message string `json:"message"`
		User    string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	incident, ok := h.incidents[id]
	if !ok {
		// Fallback for demo: if ID not found, use INC-4921
		incident = h.incidents["INC-4921"]
	}

	newUpdate := map[string]interface{}{
		"time":    "Just now",
		"message": body.Message,
		"user":    body.User,
	}

	updates, _ := incident["updates"].([]map[string]interface{})
	updates = append(updates, newUpdate)
	incident["updates"] = updates

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

func (h *SupportHandlers) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support update incident request")
	id := r.URL.Query().Get("id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	incident, ok := h.incidents[id]
	if !ok {
		// Fallback for demo
		incident = h.incidents["INC-4921"]
	}

	for k, v := range updates {
		if k == "status" || k == "assigned_to" || k == "priority" {
			incident[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

// --- New Support Command Center APIs ---

func (h *SupportHandlers) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support auth me request")
	// Mock Data for the current support agent
	profile := map[string]interface{}{
		"id":        "1476a987-76ad-4963-b64a-01a71eb739ca",
		"email":     "alex.morgan@fleet.com",
		"name":      "Alex Morgan",
		"role":      "Support Lead",
		"avatar":    "https://lh3.googleusercontent.com/aida-public/AB6AXuB4VhzOFW3Dq3K6lVra-IfzYBB25STyIEccTv5dhkHOP-UG8PC324sKwt0X1gtpfVRy_03a0VGH-UM_jfzD8OF3GLXMtXl60OV-x50rfLua-I3ee-Nk_-1tnyB5ojs_4_Rf4e7prPUNNcbTTTaOdgf-iF7L4WRC12e2czLr8OoPSWOUU72a9RqffeHmgMAAjNmh6zath8vPdZFF2cTHTDZsV0HhJd5EvSuqdR--8rd_9Jhb4JpqD_gbm_xI5Hha6JYZPAXuKlTUBm4",
		"status":    "Active",
		"last_login": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *SupportHandlers) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard metrics request")

	// Proxy to support service summary
	proxyURL := fmt.Sprintf("%s/api/v1/support/dashboard/summary", h.supportServiceURL)
	resp, err := http.Get(proxyURL)
	if err != nil {
		h.logger.Error("Failed to proxy to support service metrics", zap.Error(err))
		// Fallback to mock data matching the UI requirement
		metrics := map[string]interface{}{
			"openTickets":        42,
			"slaBreaches24h":     3,
			"activeDeviceErrors": 18,
			"avgResolutionTime":  "2.4h",
			"trends": map[string]interface{}{
				"openTicketsChange":    -12,
				"resolutionTimeChange": -15,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
		return
	}
	defer resp.Body.Close()

	// In a real scenario, we might transform the support-service response to match the dashboard UI exactly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *SupportHandlers) GetDashboardTrends(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard trends request")
	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "7d"
	}

	// Mock data for the trend chart
	trends := map[string]interface{}{
		"range": rangeParam,
		"points": []map[string]interface{}{
			{"day": "Mon", "count": 21},
			{"day": "Tue", "count": 35},
			{"day": "Wed", "count": 18},
			{"day": "Thu", "count": 42},
			{"day": "Fri", "count": 28},
			{"day": "Sat", "count": 15},
			{"day": "Sun", "count": 10},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trends)
}

func (h *SupportHandlers) GetTopErrors(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support top errors request")
	// Mock data for top system errors
	errors := []map[string]interface{}{
		{
			"errorCode": "E-504",
			"title":     "Connection Timeout",
			"count":     45,
			"severity":  "CRITICAL",
		},
		{
			"errorCode": "E-102",
			"title":     "GPS Signal Lost",
			"count":     28,
			"severity":  "WARNING",
		},
		{
			"errorCode": "E-401",
			"title":     "Auth Failure",
			"count":     16,
			"severity":  "MEDIUM",
		},
		{
			"errorCode": "E-309",
			"title":     "Sensor Malfunction",
			"count":     9,
			"severity":  "LOW",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errors)
}

func (h *SupportHandlers) GetRecentTickets(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support recent tickets request")
	// Proxy to support service recent activity
	proxyURL := fmt.Sprintf("%s/api/v1/support/dashboard/recent-activity", h.supportServiceURL)
	resp, err := http.Get(proxyURL)
	if err != nil {
		h.logger.Error("Failed to proxy to support service recent activity", zap.Error(err))
		// Fallback to mock data matching the UI table
		tickets := []map[string]interface{}{
			{
				"ticketId":   "T-9024",
				"subject":    "Vehicle telemetry sync failure",
				"reporter":   "Fleet Ops",
				"priority":   "HIGH",
				"assignedTo": "John D.",
				"updatedAt":  time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
				"status":     "IN_PROGRESS",
			},
			{
				"ticketId":   "T-9023",
				"subject":    "Driver app login issue",
				"reporter":   "Driver Support",
				"priority":   "MEDIUM",
				"assignedTo": "Anna S.",
				"updatedAt":  time.Now().Add(-24 * time.Minute).Format(time.RFC3339),
				"status":     "PENDING",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tickets": tickets,
			"total":   len(tickets),
		})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *SupportHandlers) SearchTickets(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support search tickets request", zap.String("query", r.URL.Query().Get("q")))
	// In a real scenario, this would proxy to ListTickets with filters
	h.GetRecentTickets(w, r)
}

func (h *SupportHandlers) GetTicketFilters(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support ticket filters request")
	filters := map[string]interface{}{
		"priorities": []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"},
		"statuses":   []string{"OPEN", "IN_PROGRESS", "RESOLVED", "PENDING"},
		"agents":     []string{"John D.", "Anna S.", "Mike R.", "Sarah J."},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filters)
}

func (h *SupportHandlers) ExportTickets(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support export tickets request")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}
	// Mock implementation for export
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=tickets.%s", format))
	w.Write([]byte("Ticket ID,Subject,Reporter,Priority,Assigned To,Last Updated,Status\n"))
	w.Write([]byte("T-9024,Vehicle telemetry sync failure,Fleet Ops,HIGH,John D.,10 mins ago,IN_PROGRESS\n"))
}

func (h *SupportHandlers) GetLiveDashboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Support dashboard live request")
	// For "Live Data" badge and auto-refresh parity
	status := map[string]interface{}{
		"status":      "connected",
		"last_update": time.Now().Format(time.RFC3339),
		"agent_count": 12,
		"system_load": "normal",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
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

// --- User Console APIs (Support Module) ---

func (h *SupportHandlers) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	h.logger.Info("Searching users", zap.String("query", query))
	// Mock search result
	users := []map[string]interface{}{
		{
			"userId": "USR-882910",
			"name":   "Sarah Jenkins",
			"email":  "sarah.j@fleetcorp.com",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *SupportHandlers) GetUserFullProfile(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting user profile", zap.String("userId", userId))
	profile := map[string]interface{}{
		"userId":      userId,
		"name":        "Sarah Jenkins",
		"email":       "sarah.j@fleetcorp.com",
		"role":        "Regional Manager",
		"region":      "EMEA",
		"status":      "ACTIVE",
		"verified":    true,
		"memberSince": "2021-02-12",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *SupportHandlers) GetUserSecurity(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting user security info", zap.String("userId", userId))
	security := map[string]interface{}{
		"mfaEnabled":         true,
		"mfaMethods":         []string{"AUTH_APP", "SMS"},
		"lastPasswordChange": "2025-01-27",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(security)
}

func (h *SupportHandlers) RevokeUserSession(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	sessionId := r.PathValue("sessionId")
	h.logger.Info("Revoking session", zap.String("userId", userId), zap.String("sessionId", sessionId))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Session revoked successfully"}`))
}

func (h *SupportHandlers) GetUserLoginHistory(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting login history", zap.String("userId", userId))
	history := []map[string]interface{}{
		{
			"time":      time.Now().Format(time.RFC3339),
			"eventType": "LOGIN",
			"ip":        "82.11.45.22",
			"location":  "London, UK",
			"riskScore": 10,
			"result":    "SUCCESS",
		},
		{
			"time":      time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"eventType": "LOGIN",
			"ip":        "192.158.1.1",
			"location":  "Moscow, RU",
			"riskScore": 90,
			"result":    "FAILED",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *SupportHandlers) ExportUserLoginHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Write([]byte("Time,EventType,IP,Location,RiskScore,Result\n"))
	w.Write([]byte("2026-02-10T14:32Z,LOGIN,82.11.45.22,London,10,SUCCESS\n"))
}

func (h *SupportHandlers) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting user roles", zap.String("userId", userId))
	roles := map[string]interface{}{
		"roles": []string{"Regional Manager", "Fleet Operator"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func (h *SupportHandlers) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting user permissions", zap.String("userId", userId))
	permissions := map[string]interface{}{
		"permissions": []string{
			"VIEW_FLEET",
			"APPROVE_MAINTENANCE",
			"GENERATE_REPORTS",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

func (h *SupportHandlers) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Resetting password", zap.String("userId", userId))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Password reset link sent"}`))
}

func (h *SupportHandlers) UnlockUserAccount(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Unlocking account", zap.String("userId", userId))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Account unlocked"}`))
}

func (h *SupportHandlers) ResetUserMFA(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Resetting MFA", zap.String("userId", userId))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "MFA factors cleared"}`))
}

func (h *SupportHandlers) ForceUserLogout(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Force logging out user", zap.String("userId", userId))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "All sessions terminated"}`))
}

func (h *SupportHandlers) GetUserNotes(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting user notes", zap.String("userId", userId))
	notes := []map[string]interface{}{
		{
			"note":  "User reported login issue",
			"agent": "Alex Morgan",
			"time":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *SupportHandlers) AddUserNote(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	var body struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	h.logger.Info("Adding user note", zap.String("userId", userId), zap.String("note", body.Note))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Note added successfully"}`))
}

func (h *SupportHandlers) GetUserAuditLog(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	h.logger.Info("Getting user audit log", zap.String("userId", userId))
	logs := []map[string]interface{}{
		{
			"action": "LOGIN",
			"time":   time.Now().Format(time.RFC3339),
			"details": "Successful login from Chrome 118",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
