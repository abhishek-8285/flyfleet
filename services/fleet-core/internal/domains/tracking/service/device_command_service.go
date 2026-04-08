package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/repository"
	"github.com/flyfleet/fleet-core/internal/platform/metrics"
	"go.uber.org/zap"
)

// defaultCommandTTL is how long a command stays pending before auto-expiry.
const defaultCommandTTL = 24 * time.Hour

// DeviceCommandService manages the outbound device command channel.
type DeviceCommandService struct {
	repo   repository.DeviceRegistryRepository
	logger *zap.Logger
}

// NewDeviceCommandService creates a new command service.
func NewDeviceCommandService(repo repository.DeviceRegistryRepository, logger *zap.Logger) *DeviceCommandService {
	return &DeviceCommandService{repo: repo, logger: logger}
}

// SendCommand enqueues a command for the device.
// Returns the generated command ID.
func (s *DeviceCommandService) SendCommand(ctx context.Context, deviceID, cmdType string, params map[string]any) (string, error) {
	// Validate command type
	switch cmdType {
	case models.CmdPing, models.CmdReboot, models.CmdImmobilize, models.CmdConfigUpdate:
		// ok
	default:
		return "", fmt.Errorf("unknown command type: %s", cmdType)
	}

	// Verify device exists
	device, err := s.repo.GetByDeviceID(ctx, deviceID)
	if err != nil || device == nil {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("marshal params: %w", err)
	}

	commandID, err := s.repo.CreateCommand(ctx, deviceID, cmdType, paramsJSON, defaultCommandTTL)
	if err != nil {
		s.logger.Error("failed to create device command",
			zap.String("device_id", deviceID),
			zap.String("type", cmdType),
			zap.Error(err),
		)
		return "", err
	}

	s.logger.Info("device command enqueued",
		zap.String("device_id", deviceID),
		zap.String("command_id", commandID),
		zap.String("type", cmdType),
	)
	// Record command sent metric
	metrics.RecordCommandSent(cmdType, "success")
	return commandID, nil
}

// GetPendingCommands retrieves all unexpired pending commands for a device.
func (s *DeviceCommandService) GetPendingCommands(ctx context.Context, deviceID string) ([]models.DeviceCommand, error) {
	return s.repo.GetPendingCommands(ctx, deviceID)
}

// AckCommand marks a command as executed with the device's result payload.
func (s *DeviceCommandService) AckCommand(ctx context.Context, commandID, result string) error {
	return s.repo.AckCommand(ctx, commandID, result)
}
