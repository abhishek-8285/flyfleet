package platform_compat

import (
	"encoding/json"

	"github.com/flyfleet/fleet-core/internal/domains/platform_compat/service"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type WebsocketGateway struct {
	nc     *nats.Conn
	hub    *service.Hub
	logger *zap.Logger
}

func NewWebsocketGateway(nc *nats.Conn, hub *service.Hub, logger *zap.Logger) *WebsocketGateway {
	return &WebsocketGateway{
		nc:     nc,
		hub:    hub,
		logger: logger,
	}
}

func (g *WebsocketGateway) Start() error {
	if g.nc == nil {
		g.logger.Warn("NATS connection is nil, skipping WebSocket Gateway start")
		return nil
	}

	js, err := g.nc.JetStream()
	if err != nil {
		return err
	}

	_, err = js.Subscribe("telemetry.gps", func(msg *nats.Msg) {
		var event struct {
			OrgID string `json:"org_id"`
		}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			g.logger.Debug("Failed to unmarshal telemetry event for websocket gateway", zap.Error(err))
			return
		}

		if event.OrgID != "" {
			orgID, parseErr := uuid.Parse(event.OrgID)
			if parseErr == nil {
				// Forward the full JSON message directly to WebSocket clients of that Org
				g.hub.BroadcastToOrg(orgID, json.RawMessage(msg.Data))
			}
		}
	}, nats.DeliverNew())

	return err
}
