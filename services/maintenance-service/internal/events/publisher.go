package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/maintenance-service/internal/domain"
	"github.com/flyfleet/maintenance-service/internal/middleware"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NATSPublisher struct {
	nc     *nats.Conn
	logger *zap.Logger
}

func NewNATSPublisher(url string, logger *zap.Logger) (*NATSPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSPublisher{
		nc:     nc,
		logger: logger,
	}, nil
}

func (p *NATSPublisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	orgID, _ := middleware.GetOrgID(ctx)
	
	event := domain.Event{
		ID:        uuid.New(),
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    "maintenance-service",
		Payload:   payload,
		Metadata: domain.Metadata{
			OrgID: orgID,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.nc.Publish(eventType, data); err != nil {
		p.logger.Error("failed to publish event", zap.String("type", eventType), zap.Error(err))
		return err
	}

	p.logger.Info("event published", zap.String("type", eventType), zap.String("id", event.ID.String()))
	return nil
}

func (p *NATSPublisher) Close() {
	p.nc.Close()
}

