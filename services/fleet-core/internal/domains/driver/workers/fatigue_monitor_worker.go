package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"go.uber.org/zap"
)

type DutyStatusProvider interface {
	GetActiveDriversDutyHours(ctx context.Context) ([]models.DriverDutyHours, error)
}

type FatigueMonitorWorker struct {
	provider  DutyStatusProvider
	publisher EventPublisher
	logger    *zap.Logger
}

func NewFatigueMonitorWorker(provider DutyStatusProvider, publisher EventPublisher, logger *zap.Logger) *FatigueMonitorWorker {
	return &FatigueMonitorWorker{
		provider:  provider,
		publisher: publisher,
		logger:    logger,
	}
}

func (w *FatigueMonitorWorker) Run(ctx context.Context) {
	w.logger.Info("starting fatigue monitor worker")
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.monitorFatigue(ctx)
		}
	}
}

func (w *FatigueMonitorWorker) monitorFatigue(ctx context.Context) {
	drivers, err := w.provider.GetActiveDriversDutyHours(ctx)
	if err != nil {
		w.logger.Error("failed to get active drivers duty hours", zap.Error(err))
		return
	}

	for _, d := range drivers {
		// Rules:
		// continuous driving > 5 hours
		// daily duty > 12 hours (implied by MTW act or similar rule)
		// weekly driving > 48 hours
		if d.ContinuousDrivingHours > 5.0 || d.WeeklyDrivingHours > 48.0 {
			w.publishFatigueAlert(ctx, d.DriverID, "HIGH", "Exceeded driving limits")
		}
	}
}

func (w *FatigueMonitorWorker) publishFatigueAlert(ctx context.Context, driverID string, severity string, reason string) {
	event := map[string]interface{}{
		"driver_id": driverID,
		"severity":  severity,
		"reason":    reason,
		"timestamp": time.Now(),
	}
	err := w.publisher.Publish(ctx, "driver.fatigue_alert", event)
	if err != nil {
		w.logger.Error("failed to publish fatigue alert", zap.Error(err), zap.String("driver_id", driverID))
	}
}
