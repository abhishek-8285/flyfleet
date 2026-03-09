package alert_router

import (
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts_v2/models"
)

type AlertRouter interface {
	Route(priority models.AlertPriority) []models.Channel
}

type alertRouter struct {
	// Potentially configurable mapping in the future
}

func NewAlertRouter() AlertRouter {
	return &alertRouter{}
}

func (r *alertRouter) Route(priority models.AlertPriority) []models.Channel {
	switch priority {
	case models.PriorityCritical:
		// Send everywhere for critical alerts
		return []models.Channel{models.ChannelEmail, models.ChannelSMS, models.ChannelPush, models.ChannelInApp}
	case models.PriorityHigh:
		return []models.Channel{models.ChannelEmail, models.ChannelPush, models.ChannelInApp}
	case models.PriorityNormal:
		return []models.Channel{models.ChannelPush, models.ChannelInApp}
	case models.PriorityLow:
		return []models.Channel{models.ChannelInApp}
	default:
		return []models.Channel{models.ChannelInApp}
	}
}
