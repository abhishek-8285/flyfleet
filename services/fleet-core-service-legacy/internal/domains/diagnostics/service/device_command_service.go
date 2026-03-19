package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	"github.com/google/uuid"
)

type CommandType string

const (
	CommandReboot         CommandType = "REBOOT"
	CommandResetConfig    CommandType = "RESET_CONFIG"
	CommandSyncTime       CommandType = "SYNC_TIME"
	CommandCaptureLogs    CommandType = "CAPTURE_LOGS"
	CommandUpdateFirmware CommandType = "UPDATE_FIRMWARE"
)

type CommandState string

const (
	CommandStateCreated   CommandState = "CREATED"
	CommandStateQueued    CommandState = "QUEUED"
	CommandStateDelivered CommandState = "DELIVERED"
	CommandStateAcked     CommandState = "ACKED"
	CommandStateApplied   CommandState = "APPLIED"
	CommandStateFailed    CommandState = "FAILED"
)

type DeviceCommandService struct {
	repo *repository.DiagnosticsRepository
}

func NewDeviceCommandService(repo *repository.DiagnosticsRepository) *DeviceCommandService {
	return &DeviceCommandService{repo: repo}
}

func (s *DeviceCommandService) UpdateCommandStatus(ctx context.Context, commandID uuid.UUID, state CommandState) error {
	err := s.repo.UpdateCommandStatus(ctx, commandID, string(state))
	if err != nil {
		return fmt.Errorf("failed to transition command: %w", err) // Context wrapping per prompt requirement
	}
	return nil
}
