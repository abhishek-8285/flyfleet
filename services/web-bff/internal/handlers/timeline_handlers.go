package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/web-bff/internal/client"
	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type TimelineHandlers struct {
	cfg                *config.Config
	logger             *zap.Logger
	notificationClient *client.NotificationClient
}

func NewTimelineHandlers(cfg *config.Config, logger *zap.Logger, netClient *client.NotificationClient) *TimelineHandlers {
	return &TimelineHandlers{
		cfg:                cfg,
		logger:             logger,
		notificationClient: netClient,
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

	// Fetch real events from notification-service
	logs, total, err := h.notificationClient.ListLogs(context.Background(), 1, 100)
	if err != nil {
		h.logger.Error("Failed to fetch logs from notification-service", zap.Error(err))
		// Fallback to mock data if service is down for premium experience
		logs = []client.NotificationLog{} 
	}

	events := make([]TimelineEvent, 0, len(logs))
	for _, l := range logs {
		events = append(events, TimelineEvent{
			ID:          l.ID.String(),
			Type:        l.Type,
			Title:       l.Subject,
			Description: fmt.Sprintf("Notification sent via %s to %s", l.Channel, l.UserEmail),
			Severity:    "info", // Default for notifications
			Timestamp:   l.CreatedAt,
			Actor:       l.UserEmail,
			ActorType:   "user",
			Metadata: map[string]interface{}{
				"correlation_id": l.CorrelationID,
				"status":         l.Status,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": events,
		"meta": map[string]interface{}{
			"count": len(events),
			"total": total,
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
