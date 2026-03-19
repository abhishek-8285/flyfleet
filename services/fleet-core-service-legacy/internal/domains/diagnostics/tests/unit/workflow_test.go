package unit

import (
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/workflows"
	"github.com/stretchr/testify/assert"
)

func TestDeviceWorkflowTransitions(t *testing.T) {
	engine := workflows.NewDeviceWorkflowEngine()

	t.Run("Valid Transition: ACTIVE to MAINTENANCE", func(t *testing.T) {
		err := engine.ValidateTransition("ACTIVE", "MAINTENANCE")
		assert.NoError(t, err)
	})

	t.Run("Valid Transition: ERROR to ACTIVE", func(t *testing.T) {
		err := engine.ValidateTransition("ERROR", "ACTIVE")
		assert.NoError(t, err)
	})

	t.Run("Invalid Transition: DECOMMISSIONED to ACTIVE", func(t *testing.T) {
		err := engine.ValidateTransition("DECOMMISSIONED", "ACTIVE")
		assert.Error(t, err)
	})

	t.Run("Allowed Transitions Check", func(t *testing.T) {
		// Verify some other negatives
		err := engine.ValidateTransition("ACTIVE", "ACTIVE") // Not explicitly allowed in map
		assert.Error(t, err)
	})
}

func TestFirmwareWorkflowValidation(t *testing.T) {
	engine := workflows.NewFirmwareWorkflowEngine()

	t.Run("Deployable: PUBLISHED", func(t *testing.T) {
		err := engine.ValidateDeployable("PUBLISHED")
		assert.NoError(t, err)
	})

	t.Run("Not Deployable: DRAFT", func(t *testing.T) {
		err := engine.ValidateDeployable("DRAFT")
		assert.Error(t, err)
	})

	t.Run("Not Deployable: DEPRECATED", func(t *testing.T) {
		err := engine.ValidateDeployable("DEPRECATED")
		assert.Error(t, err)
	})
}
