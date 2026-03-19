//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"
	"github.com/stretchr/testify/assert"
)

func TestDeviceHealth_Assessment(t *testing.T) {
	t.Run("Offline Detection", func(t *testing.T) {
		oldTime := time.Now().Add(-48 * time.Hour) // More than 24h
		device := &models.Device{
			LastSeenAt: &oldTime,
		}
		health := healthService.AssessHealth(device, nil)
		assert.Equal(t, service.HealthOffline, health)
	})

	t.Run("Low Battery Detection", func(t *testing.T) {
		now := time.Now()
		device := &models.Device{
			LastSeenAt: &now,
		}
		// Telemetry with low voltage
		telemetry := &models.DeviceTelemetry{
			BatteryVoltage: ptr(10.5),
		}
		health := healthService.AssessHealth(device, telemetry)
		assert.Equal(t, service.HealthWarning, health)
	})

	t.Run("Healthy State", func(t *testing.T) {
		now := time.Now()
		device := &models.Device{
			LastSeenAt: &now,
		}
		telemetry := &models.DeviceTelemetry{
			BatteryVoltage: ptr(12.6),
		}
		health := healthService.AssessHealth(device, telemetry)
		assert.Equal(t, service.HealthHealthy, health)
	})
}
