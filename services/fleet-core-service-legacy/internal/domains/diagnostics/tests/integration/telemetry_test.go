//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTelemetry_Ingestion(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	deviceID := uuid.New()

	// Initial device setup
	device := &models.Device{
		ID:         deviceID,
		OrgID:      orgID,
		DeviceIMEI: ptr("telemetry-test-imei"),
		Status:     ptr("ACTIVE"),
	}
	err := repo.CreateDevice(ctx, device)
	assert.NoError(t, err)

	t.Run("IngestTelemetry Updates Device", func(t *testing.T) {
		now := time.Now().UTC()
		telemetry := &models.DeviceTelemetry{
			GPSLat:         ptr(12.9716),
			GPSLon:         ptr(77.5946),
			BatteryVoltage: ptr(12.8),
			Timestamp:      ptr(now),
		}

		err := telemetryService.IngestTelemetry(ctx, deviceID, orgID, telemetry)
		assert.NoError(t, err)

		// Verify telemetry record
		tels, err := repo.GetTelemetry(ctx, deviceID, 1)
		assert.NoError(t, err)
		assert.Len(t, tels, 1)
		assert.Equal(t, 12.9716, *tels[0].GPSLat)

		// Verify device last_seen_at
		updatedDevice, err := repo.GetDevice(ctx, deviceID, orgID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedDevice.LastSeenAt)
		assert.WithinDuration(t, now, updatedDevice.LastSeenAt.UTC(), time.Second)
	})
}
