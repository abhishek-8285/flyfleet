package unit

import (
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

func TestDeviceStatusValidation(t *testing.T) {
	validate := validator.New()
	orgID := uuid.New()
	deviceID := uuid.New()

	t.Run("Valid Active Device", func(t *testing.T) {
		device := models.Device{
			ID:         deviceID,
			OrgID:      orgID,
			DeviceIMEI: ptr("123456789012345"),
			Status:     ptr("ACTIVE"),
		}

		err := validate.Struct(device)
		assert.NoError(t, err)
	})

	t.Run("Invalid Device Status", func(t *testing.T) {
		device := models.Device{
			ID:         deviceID,
			OrgID:      orgID,
			DeviceIMEI: ptr("123456789012345"),
			Status:     ptr("INVALID_STATUS"),
		}

		err := validate.Struct(device)
		assert.Error(t, err)
	})
}

func TestFirmwareStatusValidation(t *testing.T) {
	validate := validator.New()
	fwID := uuid.New()

	t.Run("Valid Published Firmware", func(t *testing.T) {
		fw := models.FirmwarePackage{
			ID:              fwID,
			FirmwareVersion: ptr("v1.0.0"),
			Status:          ptr("PUBLISHED"),
		}

		err := validate.Struct(fw)
		assert.NoError(t, err)
	})

	t.Run("Invalid Firmware Status", func(t *testing.T) {
		fw := models.FirmwarePackage{
			ID:              fwID,
			FirmwareVersion: ptr("v1.0.0"),
			Status:          ptr("RAW"),
		}

		err := validate.Struct(fw)
		assert.Error(t, err)
	})
}

func TestCommandStatusValidation(t *testing.T) {
	validate := validator.New()
	cmdID := uuid.New()
	deviceID := uuid.New()

	t.Run("Valid Created Command", func(t *testing.T) {
		cmd := models.DeviceCommand{
			ID:       cmdID,
			DeviceID: &deviceID,
			Status:   ptr("CREATED"),
		}

		err := validate.Struct(cmd)
		assert.NoError(t, err)
	})

	t.Run("Invalid Command Status", func(t *testing.T) {
		cmd := models.DeviceCommand{
			ID:       cmdID,
			DeviceID: &deviceID,
			Status:   ptr("RETRYING"),
		}

		err := validate.Struct(cmd)
		assert.Error(t, err)
	})
}

func TestRolloutStatusValidation(t *testing.T) {
	validate := validator.New()
	rolloutID := uuid.New()

	t.Run("Valid Scheduled Rollout", func(t *testing.T) {
		rollout := models.FirmwareRollout{
			ID:     rolloutID,
			Status: ptr("SCHEDULED"),
		}

		err := validate.Struct(rollout)
		assert.NoError(t, err)
	})

	t.Run("Invalid Rollout Status", func(t *testing.T) {
		rollout := models.FirmwareRollout{
			ID:     rolloutID,
			Status: ptr("EXECUTING"),
		}

		err := validate.Struct(rollout)
		assert.Error(t, err)
	})
}
