package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
	Type        string                 `json:"type"` // "alert", "trip", "config", "login", "device"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"` // "info", "warning", "critical", "success"
	Timestamp   time.Time              `json:"timestamp"`
	Actor       string                 `json:"actor"`      // Name of user or system component
	ActorType   string                 `json:"actor_type"` // "user", "system"
	Metadata    map[string]interface{} `json:"metadata"`
}

// GetTimeline returns the activity feed
func (h *TimelineHandlers) GetTimeline(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Timeline request received")

	// Fetch real events from notification-service
	logs, total, err := h.notificationClient.ListLogs(r.Context(), 1, 100)
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
	logs, _, err := h.notificationClient.ListLogs(r.Context(), 1, 500)
	if err != nil {
		h.logger.Warn("Failed to fetch logs for timeline stats", zap.Error(err))
		logs = []client.NotificationLog{}
	}

	now := time.Now().UTC()
	last24h := now.Add(-24 * time.Hour)
	last7d := now.Add(-7 * 24 * time.Hour)
	recentEvents := 0
	critical24h := 0
	weeklyEvents := 0

	for _, l := range logs {
		createdAt := l.CreatedAt.UTC()
		if createdAt.After(last24h) {
			recentEvents++
		}
		if createdAt.After(last7d) {
			weeklyEvents++
		}
		if l.Status == "FAILED" || l.Status == "failed" {
			critical24h++
		}
	}

	stats := []map[string]interface{}{
		{
			"label":  "Events (24h)",
			"value":  fmt.Sprintf("%d", recentEvents),
			"trend":  fmt.Sprintf("%d in 7d", weeklyEvents),
			"status": "success",
			"icon":   "event_note",
		},
		{
			"label":  "Failed Notifications (24h)",
			"value":  fmt.Sprintf("%d", critical24h),
			"trend":  "Delivery reliability monitor",
			"status": "danger",
			"icon":   "crisis_alert",
		},
		{
			"label":  "Total Events",
			"value":  fmt.Sprintf("%d", len(logs)),
			"trend":  "Across all channels",
			"status": "info",
			"icon":   "timeline",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *TimelineHandlers) GetTimelineFilters(w http.ResponseWriter, r *http.Request) {
	logs, _, err := h.notificationClient.ListLogs(r.Context(), 1, 500)
	if err != nil {
		h.logger.Warn("Failed to fetch logs for timeline filters", zap.Error(err))
		logs = []client.NotificationLog{}
	}

	typeSet := map[string]struct{}{}
	statusSet := map[string]struct{}{}
	channelSet := map[string]struct{}{}
	for _, l := range logs {
		if l.Type != "" {
			typeSet[l.Type] = struct{}{}
		}
		if l.Status != "" {
			statusSet[l.Status] = struct{}{}
		}
		if l.Channel != "" {
			channelSet[l.Channel] = struct{}{}
		}
	}

	toSortedSlice := func(set map[string]struct{}) []string {
		out := make([]string, 0, len(set))
		for k := range set {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available_filters": []string{"date_range", "event_type", "status", "channel", "actor"},
		"options": map[string]interface{}{
			"event_type": toSortedSlice(typeSet),
			"status":     toSortedSlice(statusSet),
			"channel":    toSortedSlice(channelSet),
		},
	})
}
