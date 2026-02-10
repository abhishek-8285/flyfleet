package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type IncidentHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewIncidentHandlers(cfg *config.Config, logger *zap.Logger) *IncidentHandlers {
	return &IncidentHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// 1. INCIDENT METRICS
func (h *IncidentHandlers) GetIncidentMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting incident metrics")

	res := map[string]interface{}{
		"total":        342,
		"newToday":     12,
		"criticalOpen": 5,
		"highPriority": 18,
		"inProgress":   42,
		"resolved24h":  28,
		"mttr":         "4h 12m",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 2. INCIDENT TREND DATA
func (h *IncidentHandlers) GetIncidentTrends(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting incident trends")
	
	// Mock daily count for last 30 days
	points := []map[string]interface{}{}
	now := time.Now()
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		points = append(points, map[string]interface{}{
			"date":  date,
			"count": 10 + (i % 5)*2 + (i % 3), // some fluctuation
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"points": points,
	})
}

// 3. INCIDENT FEED
func (h *IncidentHandlers) GetIncidentFeed(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting incident feed")

	res := []map[string]interface{}{
		{
			"incidentId": "INC-4921",
			"subject":    "Database latency spike in US-East region",
			"priority":   "CRITICAL",
			"timeAgo":    "2m",
                        "status":     "OPEN",
		},
		{
			"incidentId": "INC-4920",
			"subject":    "User reported API 500 errors on login",
			"priority":   "HIGH",
			"timeAgo":    "15m",
                        "status":     "INVESTIGATING",
		},
		{
			"incidentId": "INC-4918",
			"subject":    "Resolved: Payment gateway timeout",
			"priority":   "MEDIUM",
			"timeAgo":    "1h",
                        "status":     "CLOSED",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 4. INCIDENT LIST (TABLE) + 5. SEARCH
func (h *IncidentHandlers) GetIncidents(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("q")
	priority := r.URL.Query().Get("priority")
	
	h.logger.Info("Getting incidents list", zap.String("search", searchQuery), zap.String("priority", priority))

	items := []map[string]interface{}{
		{
			"incidentId":   "INC-4921",
			"priority":     "CRITICAL",
			"status":       "OPEN",
			"subject":      "Database latency spike US-East",
			"assignedTo":   "Unassigned",
			"organization": "Global Fleet Ops",
			"createdAt":    time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
                        "assigneeInitials": "UN",
		},
		{
			"incidentId":   "INC-4920",
			"priority":     "HIGH",
			"status":       "INVESTIGATING",
			"subject":      "Authentication failures for API v2",
			"assignedTo":   "Sarah J.",
			"organization": "Acme Logistics",
			"createdAt":    time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
                        "assigneeInitials": "JS",
		},
		{
			"incidentId":   "INC-4919",
			"priority":     "MEDIUM",
			"status":       "INVESTIGATING",
			"subject":      "Dashboard widget loading error",
			"assignedTo":   "Mike K.",
			"organization": "Red Arrow Transport",
			"createdAt":    time.Now().Add(-45 * time.Minute).Format(time.RFC3339),
                        "assigneeInitials": "MK",
		},
                {
			"incidentId":   "INC-4915",
			"priority":     "LOW",
			"status":       "RESOLVED",
			"subject":      "Typo in email notification template",
			"assignedTo":   "Alex L.",
			"organization": "Swift Delivery Co",
			"createdAt":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
                        "assigneeInitials": "AL",
		},
                {
			"incidentId":   "INC-4912",
			"priority":     "HIGH",
			"status":       "RESOLVED",
			"subject":      "GPS sync failure on Vehicle #240",
			"assignedTo":   "Robert T.",
			"organization": "Global Fleet Ops",
			"createdAt":    time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
                        "assigneeInitials": "RT",
		},
	}
        
        // Filter mock data if search query is present
        if searchQuery != "" {
             // simplified filter for mock
             h.logger.Debug("Mock filtering for search", zap.String("q", searchQuery))
        }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"total": 120,
	})
}

// 6. CREATE INCIDENT
func (h *IncidentHandlers) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	h.logger.Info("Creating incident", zap.Any("incident", body))

	res := map[string]interface{}{
		"incidentId": "INC-5001",
		"status":     "OPEN",
		"createdAt":  time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// 7. UPDATE INCIDENT
func (h *IncidentHandlers) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	h.logger.Info("Updating incident", zap.String("id", id), zap.Any("updates", body))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Incident updated"}`))
}

// 8. ASSIGN INCIDENT
func (h *IncidentHandlers) AssignIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		AgentId string `json:"agentId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	h.logger.Info("Assigning incident", zap.String("id", id), zap.String("agent", body.AgentId))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Incident assigned"}`))
}

// 9. CHANGE STATUS
func (h *IncidentHandlers) ChangeIncidentStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	h.logger.Info("Changing incident status", zap.String("id", id), zap.String("status", body.Status))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Status updated"}`))
}

// 10. INCIDENT DETAILS
func (h *IncidentHandlers) GetIncidentDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.logger.Info("Getting incident detail", zap.String("id", id))

	res := map[string]interface{}{
		"incidentId":   id,
		"priority":     "CRITICAL",
		"status":       "OPEN",
		"subject":      "Database latency spike US-East",
		"description":  "Persistent high latency detected in RDS cluster for the US-East region. Affecting 15% of write operations.",
		"organization": "Global Fleet Ops",
		"createdAt":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		"assignedTo":   "Sarah J.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// 11. INCIDENT COMMENTS
func (h *IncidentHandlers) GetIncidentComments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.logger.Info("Getting incident comments", zap.String("id", id))

	res := []map[string]interface{}{
		{
			"commentId": "COM-1",
			"user":      "Alex Morgan",
			"text":      "Investigating the RDS metrics now.",
			"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *IncidentHandlers) PostIncidentComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	h.logger.Info("Posting incident comment", zap.String("id", id))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Comment posted"}`))
}

// 12. EXPORT INCIDENTS
func (h *IncidentHandlers) ExportIncidents(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Exporting incidents")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=incidents_export.csv")
	w.Write([]byte("Incident ID,Priority,Status,Subject,Organization,Created\n"))
	w.Write([]byte("INC-4921,CRITICAL,OPEN,Database latency,Global Fleet Ops,2025-02-10\n"))
}

// 13. FILTER META
func (h *IncidentHandlers) GetIncidentFilters(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting incident filter metadata")

	res := map[string]interface{}{
		"priorities": []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"},
		"statuses":   []string{"OPEN", "INVESTIGATING", "RESOLVED", "CLOSED"},
		"orgs": []map[string]string{
			{"id": "ORG-1", "name": "Global Fleet Ops"},
			{"id": "ORG-2", "name": "Acme Logistics"},
		},
		"agents": []map[string]string{
			{"id": "USR-1", "name": "Sarah J."},
			{"id": "USR-2", "name": "Mike K."},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
