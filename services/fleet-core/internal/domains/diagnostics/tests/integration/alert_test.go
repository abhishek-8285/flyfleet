//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAlertSystem_Triggers(t *testing.T) {
	ctx := context.Background()
	deviceID := uuid.New()

	t.Run("Create and Resolve Alert", func(t *testing.T) {
		err := alertService.CreateAlert(ctx, deviceID, service.AlertDeviceOffline, service.SeverityCritical, "Device has been offline for > 24h")
		assert.NoError(t, err)

		// Resolve it
		err = alertService.ResolveAlert(ctx, 1) // Mock/Placeholder ID
		assert.NoError(t, err)
	})
}
