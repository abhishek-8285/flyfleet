//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics_v2/metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_Validation(t *testing.T) {
	t.Run("Metrics Registration", func(t *testing.T) {
		// Record some metrics
		metrics.RecordDeviceOnlineStatus(true)
		metrics.RecordCommandFailure()
		
		// In a real test, we'd scrap /metrics endpoint.
		// For now we just verify the functions don't panic.
		assert.NotPanics(t, func() {
			metrics.RecordCommandFailure()
		})
	})
}
