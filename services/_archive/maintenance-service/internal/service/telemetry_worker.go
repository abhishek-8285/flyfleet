package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/maintenance-service/internal/repository"
)

type TelemetryWorker struct {
	repo   repository.MaintenanceRepository
	nc     *nats.Conn
	logger *zap.Logger
}

func NewTelemetryWorker(repo repository.MaintenanceRepository, nc *nats.Conn, logger *zap.Logger) *TelemetryWorker {
	return &TelemetryWorker{
		repo:   repo,
		nc:     nc,
		logger: logger,
	}
}

type TelemetryMessage struct {
	VehicleID string   `json:"vehicle_id"`
	Odometer  *float64 `json:"odometer"`
	Timestamp int64    `json:"timestamp"`
}

func (w *TelemetryWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting telemetry worker")

	sub, err := w.nc.Subscribe("fleet.telemetry.gps", func(msg *nats.Msg) {
		var payload TelemetryMessage
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			w.logger.Error("Failed to unmarshal telemetry message", zap.Error(err))
			return
		}

		if payload.Odometer == nil {
			return
		}

		vehicleID, err := uuid.Parse(payload.VehicleID)
		if err != nil {
			w.logger.Error("Invalid vehicle ID in telemetry", zap.String("vehicle_id", payload.VehicleID))
			return
		}

		w.processOdometer(ctx, vehicleID, int(*payload.Odometer))
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to telemetry: %w", err)
	}

	w.logger.Info("Subscribed to fleet.telemetry.gps")

	go func() {
		<-ctx.Done()
		w.logger.Info("Stopping telemetry worker")
		sub.Unsubscribe()
	}()

	return nil
}

func (w *TelemetryWorker) processOdometer(ctx context.Context, vehicleID uuid.UUID, odometer int) {
	schedules, err := w.repo.GetDueByOdometer(ctx, vehicleID, odometer)
	if err != nil {
		w.logger.Error("Failed to query due schedules", zap.Error(err))
		return
	}

	for _, s := range schedules {
		w.logger.Warn("Maintenance due by odometer!",
			zap.String("vehicle_id", vehicleID.String()),
			zap.String("schedule_id", s.ID),
			zap.String("name", s.Name),
			zap.Int("due_odometer", *s.DueOdometer),
			zap.Int("current_odometer", odometer),
		)

		// In a real system, we might trigger a notification or update status.
		// For now, let's mark it as a priority if it hits exactly or exceeds.
		// We could also publish to fleet.alerts.maintenance
		alertData, _ := json.Marshal(map[string]interface{}{
			"type":        "maintenance_due",
			"vehicle_id":  vehicleID.String(),
			"schedule_id": s.ID,
			"name":        s.Name,
			"odometer":    odometer,
			"priority":    s.Priority,
			"timestamp":   time.Now().Unix(),
		})
		w.nc.Publish("fleet.alerts.maintenance", alertData)
	}
}
