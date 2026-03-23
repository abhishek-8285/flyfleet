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

func TestRepository_DeviceOperations(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	deviceID := uuid.New()

	t.Run("CreateDevice", func(t *testing.T) {
		device := &models.Device{
			ID:         deviceID,
			OrgID:      orgID,
			DeviceIMEI: ptr("123456789012345"),
			Status:     ptr("ACTIVE"),
		}

		err := repo.CreateDevice(ctx, device)
		assert.NoError(t, err)

		// Verify
		fetched, err := repo.GetDevice(ctx, deviceID, orgID)
		assert.NoError(t, err)
		assert.Equal(t, device.DeviceIMEI, fetched.DeviceIMEI)
	})

	t.Run("InsertTelemetry", func(t *testing.T) {
		telemetry := &models.DeviceTelemetry{
			DeviceID:       &deviceID,
			Timestamp:      ptr(time.Now()),
			BatteryVoltage: ptr(12.5),
			Speed:          ptr(60.0),
		}

		err := repo.InsertTelemetry(ctx, telemetry)
		assert.NoError(t, err)
		assert.NotZero(t, telemetry.ID)
	})

	t.Run("CreateCommand", func(t *testing.T) {
		cmdID := uuid.New()
		cmd := &models.DeviceCommand{
			ID:          cmdID,
			DeviceID:    &deviceID,
			CommandType: ptr("REBOOT"),
			Status:      ptr("CREATED"),
		}

		err := repo.CreateCommand(ctx, cmd)
		assert.NoError(t, err)
	})

	t.Run("CreateFirmware", func(t *testing.T) {
		fwID := uuid.New()
		fw := &models.FirmwarePackage{
			ID:              fwID,
			FirmwareVersion: ptr("v1.0.0"),
			Status:          ptr("DRAFT"),
		}

		err := repo.CreateFirmware(ctx, fw)
		assert.NoError(t, err)
	})

	t.Run("CreateRollout", func(t *testing.T) {
		rolloutID := uuid.New()
		rollout := &models.FirmwareRollout{
			ID:     rolloutID,
			Status: ptr("SCHEDULED"),
		}

		err := repo.CreateRollout(ctx, rollout)
		assert.NoError(t, err)
	})
}
