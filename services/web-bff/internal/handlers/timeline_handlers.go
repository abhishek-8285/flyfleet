package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type TimelineHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewTimelineHandlers(cfg *config.Config, logger *zap.Logger) *TimelineHandlers {
	return &TimelineHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// TimelineEvent represents an activity stream item
type TimelineEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`       // "alert", "trip", "config", "login", "device"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`   // "info", "warning", "critical", "success"
	Timestamp   time.Time              `json:"timestamp"`
	Actor       string                 `json:"actor"`      // Name of user or system component
	ActorType   string                 `json:"actor_type"` // "user", "system"
	Metadata    map[string]interface{} `json:"metadata"`
}

// GetTimeline returns the activity feed
func (h *TimelineHandlers) GetTimeline(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Timeline request received")

	// Mock Data
	events := []TimelineEvent{
		{
			ID:          "evt_1",
			Type:        "alert",
			Title:       "Vehicle Breakdown Reported",
			Description: "Vehicle Truck-892 reported engine failure on I-95 Southbound. Driver John Doe initiated SOS signal.",
			Severity:    "critical",
			Timestamp:   time.Now().Add(-2 * time.Minute),
			Actor:       "System",
			ActorType:   "system",
			Metadata: map[string]interface{}{
				"vehicle_id": "Truck-892",
				"location":   "Lat: 40.7128, Long: -74.0060",
			},
		},
		{
			ID:          "evt_2",
			Type:        "trip",
			Title:       "Delivery Route #4022 Initiated",
			Description: "Trip started from Distribution Center A towards Boston Metro Area.",
			Severity:    "success",
			Timestamp:   time.Now().Add(-15 * time.Minute),
			Actor:       "Mike Ross",
			ActorType:   "user",
			Metadata: map[string]interface{}{
				"trip_id": "T-4022",
			},
		},
		{
			ID:          "evt_3",
			Type:        "config",
			Title:       "Speed Limit Policy Update",
			Description: "Global speed threshold updated from 75 MPH to 70 MPH for Heavy Duty Fleet.",
			Severity:    "info",
			Timestamp:   time.Now().Add(-45 * time.Minute),
			Actor:       "Alex Morgan (Admin)",
			ActorType:   "user",
			Metadata: nil,
		},
		{
			ID:          "evt_4",
			Type:        "login",
			Title:       "Successful Login",
			Description: "User accessed the dashboard via Web Client from IP 192.168.1.42.",
			Severity:    "info",
			Timestamp:   time.Now().Add(-1 * time.Hour),
			Actor:       "Sarah Connor",
			ActorType:   "user",
			Metadata: nil,
		},
		{
			ID:          "evt_5",
			Type:        "device",
			Title:       "Telematics Signal Weak",
			Description: "Intermittent signal loss detected on Unit #221 in Sector 4. Last ping latency > 500ms.",
			Severity:    "warning",
			Timestamp:   time.Now().Add(-2 * time.Hour),
			Actor:       "ECU-882910",
			ActorType:   "system",
			Metadata: map[string]interface{}{
				"device_id": "Unit-221",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": events,
		"meta": map[string]interface{}{
			"count": len(events),
			"total": 4285,
		},
	})
}

// GetTimelineStats returns high-level metrics for the timeline dashboard
func (h *TimelineHandlers) GetTimelineStats(w http.ResponseWriter, r *http.Request) {
	stats := []map[string]interface{}{
		{
			"label": "Daily Events Count",
			"value": "4,285",
			"trend": "+12%",
			"status": "success",
			"icon": "event_note",
		},
		{
			"label": "Critical Alerts (24h)",
			"value": "12",
			"trend": "Requires immediate action",
			"status": "danger",
			"icon": "crisis_alert",
		},
		{
			"label": "Active Trips",
			"value": "842",
			"trend": "Currently on route",
			"status": "success",
			"icon": "local_shipping",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *TimelineHandlers) GetTimelineFilters(w http.ResponseWriter, r *http.Request) {
	// Simple stub for filter options
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"available_filters": ["date_range", "event_type", "severity", "entity", "actor"]}`))
}
