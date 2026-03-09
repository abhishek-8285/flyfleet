package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeviceWorkflowService handles device lifecycle and status workflows
type DeviceWorkflowService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewDeviceWorkflowService creates a new device workflow service
func NewDeviceWorkflowService(repo repository.Repository, logger *zap.Logger) *DeviceWorkflowService {
	return &DeviceWorkflowService{
		repo:   repo,
		logger: logger,
	}
}

// UpdateDeviceStatus updates device status with validation
func (s *DeviceWorkflowService) UpdateDeviceStatus(ctx context.Context, deviceID, orgID, userID uuid.UUID, newStatus models.DeviceStatus, reason string) error {
	device, err := s.repo.GetDevice(ctx, deviceID, orgID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	currentStatus := models.DeviceStatus(device.Status)
	if !IsValidDeviceStatusTransition(currentStatus, newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
	}

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityType: "device",
		EntityID:   deviceID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": currentStatus},
		NewValue:   map[string]interface{}{"status": newStatus},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     &reason,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for device status change", zap.Error(err))
	}

	s.logger.Info("Device status updated",
		zap.String("device_id", deviceID.String()),
		zap.String("from", string(currentStatus)),
		zap.String("to", string(newStatus)),
		zap.String("reason", reason),
	)

	return nil
}

// IsValidDeviceStatusTransition checks if a device status transition is valid
func IsValidDeviceStatusTransition(from, to models.DeviceStatus) bool {
	validTransitions := map[models.DeviceStatus][]models.DeviceStatus{
		models.StatusActive:      {models.StatusInactive, models.StatusError, models.StatusMaintenance},
		models.StatusInactive:    {models.StatusActive, models.StatusMaintenance},
		models.StatusError:       {models.StatusActive, models.StatusMaintenance},
		models.StatusMaintenance: {models.StatusActive, models.StatusInactive},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// MarkDeviceOffline marks a device as offline based on heartbeat timeout
func (s *DeviceWorkflowService) MarkDeviceOffline(ctx context.Context, deviceID uuid.UUID, timeoutMinutes int) error {
	device, err := s.repo.GetDevice(ctx, deviceID, uuid.Nil)
	if err != nil {
		return err
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	if device.LastHeartbeat == nil {
		return nil // No heartbeat recorded yet
	}

	timeSinceHeartbeat := time.Since(*device.LastHeartbeat)
	if timeSinceHeartbeat > time.Duration(timeoutMinutes)*time.Minute {
		// Device is offline
		if device.Status == string(models.StatusActive) {
			s.logger.Info("Device marked as offline due to heartbeat timeout",
				zap.String("device_id", deviceID.String()),
				zap.Duration("time_since_heartbeat", timeSinceHeartbeat),
			)
		}
	}

	return nil
}

// GetDeviceHealthStatus returns device health status based on multiple factors
func (s *DeviceWorkflowService) GetDeviceHealthStatus(ctx context.Context, deviceID uuid.UUID) (*DeviceHealthStatus, error) {
	device, err := s.repo.GetDevice(ctx, deviceID, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrDeviceNotFound
	}

	status := &DeviceHealthStatus{
		DeviceID:       deviceID,
		OverallStatus:  "healthy",
		LastHeartbeat:  device.LastHeartbeat,
		BatteryLevel:   device.BatteryLevel,
		SignalStrength: device.SignalStrength,
	}

	// Check battery
	if device.BatteryLevel != nil {
		if *device.BatteryLevel < 10 {
			status.BatteryStatus = "critical"
			status.OverallStatus = "critical"
		} else if *device.BatteryLevel < 20 {
			status.BatteryStatus = "warning"
			if status.OverallStatus == "healthy" {
				status.OverallStatus = "warning"
			}
		} else {
			status.BatteryStatus = "healthy"
		}
	}

	// Check signal
	if device.SignalStrength != nil {
		if *device.SignalStrength < -90 {
			status.SignalStatus = "critical"
			status.OverallStatus = "critical"
		} else if *device.SignalStrength < -80 {
			status.SignalStatus = "warning"
			if status.OverallStatus == "healthy" {
				status.OverallStatus = "warning"
			}
		} else {
			status.SignalStatus = "healthy"
		}
	}

	// Check heartbeat
	if device.LastHeartbeat != nil {
		timeSinceHeartbeat := time.Since(*device.LastHeartbeat)
		if timeSinceHeartbeat > 30*time.Minute {
			status.HeartbeatStatus = "offline"
			status.OverallStatus = "critical"
		} else if timeSinceHeartbeat > 5*time.Minute {
			status.HeartbeatStatus = "warning"
			if status.OverallStatus == "healthy" {
				status.OverallStatus = "warning"
			}
		} else {
			status.HeartbeatStatus = "online"
		}
	}

	return status, nil
}

// DeviceHealthStatus represents comprehensive device health
type DeviceHealthStatus struct {
	DeviceID       uuid.UUID  `json:"device_id"`
	OverallStatus  string     `json:"overall_status"` // healthy, warning, critical
	LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty"`
	HeartbeatStatus string    `json:"heartbeat_status"` // online, warning, offline
	BatteryLevel   *int       `json:"battery_level,omitempty"`
	BatteryStatus  string     `json:"battery_status"` // healthy, warning, critical
	SignalStrength *int       `json:"signal_strength,omitempty"`
	SignalStatus   string     `json:"signal_status"` // healthy, warning, critical
}
