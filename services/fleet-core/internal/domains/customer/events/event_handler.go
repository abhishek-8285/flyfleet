package events

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/customer/repository"
	"go.uber.org/zap"
)

type EventHandler struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewEventHandler(repo repository.CustomerV2Repository, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		repo:   repo,
		logger: logger,
	}
}

// HandleTripEvent processes trip_v2 logistics events from the event bus
func (h *EventHandler) HandleTripEvent(ctx context.Context, payload map[string]interface{}) error {
	eventType, ok := payload["event_type"].(string)
	if !ok {
		return nil // Ignore invalid payload
	}

	h.logger.Info("Received logistics event in customer_v2 module", zap.String("event_type", eventType))

	switch eventType {
	case "trip.started":
		// Sync shipment state
		h.logger.Debug("Handling trip.started event")
	case "trip.completed":
		// Mark shipments as delivered
		h.logger.Debug("Handling trip.completed event")
	case "vehicle.location.updated":
		// Push tracking updates
		h.logger.Debug("Handling vehicle.location.updated event")
	case "shipment.delivered":
		h.logger.Debug("Handling shipment.delivered event")
	}

	return nil
}
