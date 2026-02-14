package service

import (
	"context"

	"github.com/flyfleet/device-diagnostics-service/internal/models"
	"github.com/flyfleet/device-diagnostics-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeviceDiagnosticsService interface {
	// Dashboard
	GetSummary(ctx context.Context, orgID uuid.UUID) (*models.DeviceSummary, error)
	ListDevices(ctx context.Context, orgID uuid.UUID, status string, limit, offset int) ([]*models.Device, error)

	// Device details
	GetDevice(ctx context.Context, id, orgID uuid.UUID) (*models.Device, error)
	GetDeviceSensors(ctx context.Context, deviceID, orgID uuid.UUID) ([]models.SensorReading, error)
	GetDeviceLogs(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DiagnosticsLog, error)

	// Telemetry
	GetDeviceTelemetry(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceTelemetry, error)

	// Commands
	SendCommand(ctx context.Context, deviceID, orgID, userID uuid.UUID, command, category string, params map[string]interface{}) (*models.DeviceCommand, error)
	GetCommandHistory(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error)

	// Firmware
	ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error)
	CreateFirmware(ctx context.Context, version, hwTarget, fileURL string, fileSize int, checksum, notes *string, isCritical bool) (*models.FirmwarePackage, error)
	ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error)
	CreateRollout(ctx context.Context, firmwareID, orgID, userID uuid.UUID, name, deviceGroup string, totalDevices int) (*models.FirmwareRollout, error)
}

type service struct {
	repo   repository.Repository
	logger *zap.Logger
}

func NewDeviceDiagnosticsService(repo repository.Repository, logger *zap.Logger) DeviceDiagnosticsService {
	return &service{repo: repo, logger: logger}
}

func (s *service) GetSummary(ctx context.Context, orgID uuid.UUID) (*models.DeviceSummary, error) {
	return s.repo.GetDeviceSummary(ctx, orgID)
}

func (s *service) ListDevices(ctx context.Context, orgID uuid.UUID, status string, limit, offset int) ([]*models.Device, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.ListDevices(ctx, orgID, status, limit, offset)
}

func (s *service) GetDevice(ctx context.Context, id, orgID uuid.UUID) (*models.Device, error) {
	return s.repo.GetDevice(ctx, id, orgID)
}

func (s *service) GetDeviceSensors(ctx context.Context, deviceID, orgID uuid.UUID) ([]models.SensorReading, error) {
	device, err := s.repo.GetDevice(ctx, deviceID, orgID)
	if err != nil {
		return nil, err
	}

	sensors := []models.SensorReading{}

	// Build sensor readings from device data
	if device.BatteryLevel != nil {
		status := "normal"
		if *device.BatteryLevel < 20 {
			status = "warning"
		}
		if *device.BatteryLevel < 10 {
			status = "critical"
		}
		sensors = append(sensors, models.SensorReading{
			Name:   "Battery",
			Value:  *device.BatteryLevel,
			Unit:   "%",
			Status: status,
		})
	}

	if device.SignalStrength != nil {
		status := "normal"
		if *device.SignalStrength < -80 {
			status = "warning"
		}
		if *device.SignalStrength < -90 {
			status = "critical"
		}
		sensors = append(sensors, models.SensorReading{
			Name:   "Signal Strength",
			Value:  *device.SignalStrength,
			Unit:   "dBm",
			Status: status,
		})
	}

	// Add ignition status (mock for now - would come from latest telemetry)
	sensors = append(sensors, models.SensorReading{
		Name:   "Ignition",
		Value:  "ON",
		Status: "normal",
	})

	return sensors, nil
}

func (s *service) GetDeviceLogs(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DiagnosticsLog, error) {
	if limit == 0 {
		limit = 100
	}
	return s.repo.ListDiagnosticsLogs(ctx, deviceID, limit, offset)
}

func (s *service) GetDeviceTelemetry(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceTelemetry, error) {
	if limit == 0 {
		limit = 100
	}
	return s.repo.ListTelemetry(ctx, deviceID, limit, offset)
}

func (s *service) SendCommand(ctx context.Context, deviceID, orgID, userID uuid.UUID, command, category string, params map[string]interface{}) (*models.DeviceCommand, error) {
	cmd := &models.DeviceCommand{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		OrgID:     orgID,
		Command:   command,
		Category:  category,
		Params:    params,
		Status:    "pending",
		CreatedBy: &userID,
	}

	if err := s.repo.CreateCommand(ctx, cmd); err != nil {
		return nil, err
	}

	s.logger.Info("Command queued",
		zap.String("device_id", deviceID.String()),
		zap.String("command", command),
		zap.String("command_id", cmd.ID.String()),
	)

	// TODO: In production, send command to device via MQTT/WebSocket
	// For now, mark as sent
	cmd.Status = "sent"

	return cmd, nil
}

func (s *service) GetCommandHistory(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.ListCommands(ctx, deviceID, limit, offset)
}

func (s *service) ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.ListFirmware(ctx, limit, offset)
}

func (s *service) CreateFirmware(ctx context.Context, version, hwTarget, fileURL string, fileSize int, checksum, notes *string, isCritical bool) (*models.FirmwarePackage, error) {
	fw := &models.FirmwarePackage{
		ID:             uuid.New(),
		Version:        version,
		HardwareTarget: hwTarget,
		FileURL:        fileURL,
		FileSize:       &fileSize,
		Checksum:       checksum,
		ReleaseNotes:   notes,
		Status:         "draft",
		IsCritical:     isCritical,
	}

	if err := s.repo.CreateFirmware(ctx, fw); err != nil {
		return nil, err
	}

	return fw, nil
}

func (s *service) ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.ListRollouts(ctx, orgID, limit, offset)
}

func (s *service) CreateRollout(ctx context.Context, firmwareID, orgID, userID uuid.UUID, name, deviceGroup string, totalDevices int) (*models.FirmwareRollout, error) {
	rollout := &models.FirmwareRollout{
		ID:              uuid.New(),
		FirmwareID:      firmwareID,
		OrgID:           orgID,
		Name:            &name,
		DeviceGroup:     &deviceGroup,
		TotalDevices:    totalDevices,
		Status:          "scheduled",
		RolloutStrategy: "gradual",
		MaxConcurrent:   10,
		CreatedBy:       &userID,
	}

	if err := s.repo.CreateRollout(ctx, rollout); err != nil {
		return nil, err
	}

	s.logger.Info("Rollout created",
		zap.String("rollout_id", rollout.ID.String()),
		zap.String("firmware_id", firmwareID.String()),
		zap.Int("total_devices", totalDevices),
	)

	return rollout, nil
}
