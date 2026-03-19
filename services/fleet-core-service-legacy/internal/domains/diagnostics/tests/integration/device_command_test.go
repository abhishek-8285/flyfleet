//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCommand_Lifecycle(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	deviceID := uuid.New()
	cmdID := uuid.New()

	// 1. Setup Device first to satisfy FK
	device := &models.Device{
		ID:         deviceID,
		OrgID:      orgID,
		DeviceIMEI: ptr("command-test-imei"),
		Status:     ptr("ACTIVE"),
	}
	err := repo.CreateDevice(ctx, device)
	assert.NoError(t, err)

	t.Run("Command State Transitions", func(t *testing.T) {
		cmd := &models.DeviceCommand{
			ID:          cmdID,
			DeviceID:    &deviceID,
			CommandType: ptr("REBOOT"),
			Status:      ptr("CREATED"),
		}

		err := repo.CreateCommand(ctx, cmd)
		assert.NoError(t, err)

		// Transition to DELIVERED
		err = commandService.UpdateCommandStatus(ctx, cmdID, service.CommandStateDelivered)
		assert.NoError(t, err)

		// Transition to APPLIED
		err = commandService.UpdateCommandStatus(ctx, cmdID, service.CommandStateApplied)
		assert.NoError(t, err)
	})
}
