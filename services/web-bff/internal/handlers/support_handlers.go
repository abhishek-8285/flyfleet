package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type SupportHandlers struct {
	cfg       *config.Config
	logger    *zap.Logger
	incidents map[string]map[string]interface{}
	mu        sync.RWMutex
}

func NewSupportHandlers(cfg *config.Config, logger *zap.Logger) *SupportHandlers {
	h := &SupportHandlers{
		cfg:       cfg,
		logger:    logger,
		incidents: make(map[string]map[string]interface{}),
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

	// Mock Data
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
