package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeviceMonitorWorker struct {
	repo            repository.ComplianceRepository
	alertDispatcher AlertDispatcher
	logger          logger.Logger
}

func NewDeviceMonitorWorker(repo repository.ComplianceRepository, alertDispatcher AlertDispatcher, logger logger.Logger) *DeviceMonitorWorker {
	return &DeviceMonitorWorker{
		repo:            repo,
		alertDispatcher: alertDispatcher,
		logger:          logger,
	}
}

func (w *DeviceMonitorWorker) Run(ctx context.Context) error {
	w.logger.Info("starting device monitor worker")

	// 1. Detect Offline Devices (> 5 minutes)
	offlineThreshold := time.Now().Add(-5 * time.Minute)
	devices, err := w.repo.GetOfflineDevices(ctx, offlineThreshold)
	if err != nil {
		w.logger.Error("failed to fetch offline devices", zap.Error(err))
		return err
	}

	for _, device := range devices {
		w.logger.Warn("Device offline detected", zap.String("device_id", device.ID.String()), zap.String("imei", device.DeviceIMEI))

		// Trigger Compliance Alert
		alert := &models.ComplianceAlert{
			ID:             uuid.New(),
			TenantID: uuid.Nil, // In real app, fetch from vehicle association
			Type:           models.AlertDeviceOffline,
			EntityType:     models.EntityDevice,
			EntityID:       device.ID,
			Severity:       models.AlertSeverityMedium,
		}

		if err := w.alertDispatcher.Dispatch(ctx, alert); err != nil {
			w.logger.Error("failed to dispatch offline alert", zap.Error(err), zap.String("device_id", device.ID.String()))
		}

		// Update device status to OFFLINE if it wasn't already
		if device.Status != models.DeviceStatusOffline {
			device.Status = models.DeviceStatusOffline
			w.repo.RegisterDevice(ctx, &device) // Re-using RegisterDevice as an UPSERT/Update status for now or add UpdateDevice
		}
	}

	return nil
}
