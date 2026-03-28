package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	"github.com/google/uuid"
)

type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "INFO"
	SeverityWarning  AlertSeverity = "WARNING"
	SeverityCritical AlertSeverity = "CRITICAL"
)

type AlertType string

const (
	AlertDeviceOffline        AlertType = "DEVICE_OFFLINE"
	AlertLowBattery           AlertType = "LOW_BATTERY"
	AlertSignalLost           AlertType = "SIGNAL_LOST"
	AlertTelemetryDelay       AlertType = "TELEMETRY_DELAY"
	AlertFirmwareUpdateFailed AlertType = "FIRMWARE_UPDATE_FAILED"
	AlertDeviceOverheating    AlertType = "DEVICE_OVERHEATING"
)

type DeviceAlertService struct {
	repo *repository.DiagnosticsRepository
	// In a complete implementation, this would also take an EventPublisher
}

func NewDeviceAlertService(repo *repository.DiagnosticsRepository) *DeviceAlertService {
	return &DeviceAlertService{
		repo: repo,
	}
}

// CreateAlert stores a new alert for a device and emits an event.
func (s *DeviceAlertService) CreateAlert(ctx context.Context, deviceID uuid.UUID, alertType AlertType, severity AlertSeverity, msg string) error {
	// Normally this would persist to the DB:
	// INSERT INTO device_alerts (device_id, alert_type, severity, alert_message) ...

	// Emit event
	fmt.Printf("[ALERT CREATED] Device %s: %s (%s) - %s\n", deviceID.String(), alertType, severity, msg)
	return nil
}

// ResolveAlert marks an active alert as resolved.
func (s *DeviceAlertService) ResolveAlert(ctx context.Context, alertID int64) error {
	// UPDATE device_alerts SET resolved_at = NOW() WHERE id = $1
	fmt.Printf("[ALERT RESOLVED] Alert ID %d resolved at %s\n", alertID, time.Now().Format(time.RFC3339))
	return nil
}
