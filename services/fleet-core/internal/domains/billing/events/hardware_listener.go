package events

import (
	"context"
	"encoding/json"

	"github.com/flyfleet/fleet-core/internal/domains/billing/service"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type HardwareListener struct {
	nc        *nats.Conn
	hwService *service.HardwareService
	logger    *zap.Logger
}

func NewHardwareListener(nc *nats.Conn, hwService *service.HardwareService, logger *zap.Logger) *HardwareListener {
	return &HardwareListener{
		nc:        nc,
		hwService: hwService,
		logger:    logger,
	}
}

func (l *HardwareListener) Start() error {
	_, err := l.nc.Subscribe("fleet.device.registered", func(m *nats.Msg) {
		var event struct {
			DeviceID       string `json:"device_id"`
			OrganizationID string `json:"organization_id"`
			VehicleID      string `json:"vehicle_id"`
			Timestamp      int64  `json:"timestamp"`
		}

		if err := json.Unmarshal(m.Data, &event); err != nil {
			l.logger.Error("Failed to unmarshal device registered event", zap.Error(err))
			return
		}

		// Trigger hardware activation in billing
		if err := l.hwService.ProcessActivation(context.Background(), event.DeviceID); err != nil {
			l.logger.Error("Failed to process hardware activation", zap.Error(err), zap.String("device_id", event.DeviceID))
			return
		}

		l.logger.Info("Successfully processed hardware activation for device", zap.String("device_id", event.DeviceID))
	})

	if err != nil {
		return err
	}

	l.logger.Info("Hardware listener started listening to fleet.device.registered")
	return nil
}
