package service

import (
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/repository"
)

type HealthState string

const (
	HealthHealthy  HealthState = "HEALTHY"
	HealthWarning  HealthState = "WARNING"
	HealthCritical HealthState = "CRITICAL"
	HealthOffline  HealthState = "OFFLINE"
)

type DeviceHealthService struct {
	repo *repository.DiagnosticsRepository
}

func NewDeviceHealthService(repo *repository.DiagnosticsRepository) *DeviceHealthService {
	return &DeviceHealthService{repo: repo}
}

func (s *DeviceHealthService) AssessHealth(device *models.Device, telemetry *models.DeviceTelemetry) HealthState {
	if device.LastSeenAt == nil {
		return HealthOffline
	}

	delay := time.Since(*device.LastSeenAt)

	// rules config
	offlineThreshold := 24 * time.Hour
	criticalThreshold := 6 * time.Hour

	if delay > offlineThreshold {
		return HealthOffline
	}

	if delay > criticalThreshold {
		return HealthCritical
	}

	if telemetry != nil && telemetry.BatteryVoltage != nil && *telemetry.BatteryVoltage < 11.5 {
		return HealthWarning
	}

	return HealthHealthy
}
