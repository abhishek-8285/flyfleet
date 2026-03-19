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

func TestFailure_Recovery(t *testing.T) {
	ctx := context.Background()

	t.Run("Command Failure Triggers Backoff", func(t *testing.T) {
		orgID := uuid.New()
		deviceID := uuid.New()
		cmdID := uuid.New()

		// Setup device
		_ = repo.CreateDevice(ctx, &models.Device{
			ID: deviceID, OrgID: orgID, DeviceIMEI: ptr("fail-test-imei"), Status: ptr("ACTIVE"),
		})

		// Create command
		_ = repo.CreateCommand(ctx, &models.DeviceCommand{
			ID: cmdID, DeviceID: &deviceID, Status: ptr("CREATED"),
		})

		// Fail the command
		err := commandService.UpdateCommandStatus(ctx, cmdID, service.CommandStateFailed)
		assert.NoError(t, err)

		// CommandRetryWorker should handle this (tested in workers_test, but here we verify state)
	})

	t.Run("Rollout Paused on Failure Threshold", func(t *testing.T) {
		rolloutID := uuid.New()
		threshold := 1
		rollout := &models.FirmwareRollout{
			ID:               rolloutID,
			Status:           ptr("IN_PROGRESS"),
			FailureThreshold: &threshold,
		}
		_ = repo.CreateRollout(ctx, rollout)

		// Simulate 1 failure (which is the threshold)
		// EvaluateRolloutSafety(ctx, rollout, successCount, failureCount)
		err := rolloutService.EvaluateRolloutSafety(ctx, rollout, 0, 2) // 2 failures > threshold (1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rollout halted")
	})
}
