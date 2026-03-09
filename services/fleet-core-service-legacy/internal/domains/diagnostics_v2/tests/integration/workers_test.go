//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics_v2/workers"
	"github.com/google/uuid"
)

func TestWorkers_Lifecycle(t *testing.T) {
	ctx := context.Background()

	t.Run("Command Retry Worker", func(t *testing.T) {
		// Setup a failed command
		deviceID := uuid.New()
		cmdID := uuid.New()
		
		device := &models.Device{
			ID: deviceID,
			OrgID: uuid.New(),
			DeviceIMEI: ptr("worker-test-imei"),
			Status: ptr("ACTIVE"),
		}
		_ = repo.CreateDevice(ctx, device)

		cmd := &models.DeviceCommand{
			ID: cmdID,
			DeviceID: &deviceID,
			Status: ptr("FAILED"),
		}
		_ = repo.CreateCommand(ctx, cmd)

		worker := workers.NewCommandRetryWorker(commandService)
		worker.Process(ctx)
	})

	t.Run("Device Health Worker", func(t *testing.T) {
		worker := workers.NewDeviceHealthWorker(healthService)
		worker.Process(ctx)
	})
}
