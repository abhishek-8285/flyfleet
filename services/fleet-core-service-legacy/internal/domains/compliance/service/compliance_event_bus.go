package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"go.uber.org/zap"
)

type ComplianceEvent struct {
	Event         string    `json:"event"`
	DriverID      string    `json:"driver_id,omitempty"`
	VehicleID     string    `json:"vehicle_id,omitempty"`
	ViolationType string    `json:"violation_type,omitempty"`
	AlertType     string    `json:"alert_type,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Metadata      any       `json:"metadata,omitempty"`
}

type ComplianceEventBus interface {
	Publish(ctx context.Context, event ComplianceEvent) error
}

type natsEventBusImpl struct {
	logger logger.Logger
}

func NewComplianceEventBus(logger logger.Logger) ComplianceEventBus {
	return &natsEventBusImpl{
		logger: logger,
	}
}

func (b *natsEventBusImpl) Publish(ctx context.Context, event ComplianceEvent) error {
	// In a real implementation:
	// b.natsConn.Publish(event.Event, payload)

	b.logger.Info("Compliance Event Published",
		zap.String("event", event.Event),
		zap.String("driver", event.DriverID),
		zap.String("vehicle", event.VehicleID),
		zap.Time("timestamp", event.Timestamp),
	)

	return nil
}
