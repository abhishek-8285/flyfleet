package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
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
	GetLatestDeviceTelemetry(ctx context.Context, deviceID uuid.UUID) (*models.DeviceTelemetry, error)

	// Commands
	SendCommand(ctx context.Context, deviceID, orgID, userID uuid.UUID, command, category string, params map[string]interface{}) (*models.DeviceCommand, error)
	GetCommandHistory(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error)
	TransitionCommand(ctx context.Context, deviceID, orgID, commandID uuid.UUID, nextStatus string, result map[string]interface{}, errorMessage *string) (*models.DeviceCommand, error)
	ListDeviceTwins(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DeviceTwin, error)
	GetDeviceTwin(ctx context.Context, deviceID, orgID uuid.UUID) (*models.DeviceTwin, error)
	UpdateDesiredState(ctx context.Context, deviceID, orgID uuid.UUID, desiredState map[string]interface{}, commandID *uuid.UUID, commandStatus *string) error
	UpdateReportedState(ctx context.Context, deviceID, orgID uuid.UUID, reportedState map[string]interface{}, lastSeen *time.Time) error

	// Firmware
	ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error)
	CreateFirmware(ctx context.Context, version, hwTarget, fileURL string, fileSize int, checksum, notes *string, isCritical bool) (*models.FirmwarePackage, error)
	ListRollouts(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.FirmwareRollout, error)
	GetRollout(ctx context.Context, rolloutID, orgID uuid.UUID) (*models.FirmwareRollout, error)
	ListRolloutEvents(ctx context.Context, rolloutID, orgID uuid.UUID, eventType string, limit, offset int) ([]*models.FirmwareRolloutEvent, error)
	AddRolloutNote(ctx context.Context, rolloutID, orgID uuid.UUID, actorUserID *uuid.UUID, message string, metadata map[string]interface{}) (*models.FirmwareRolloutEvent, error)
	CreateRollout(ctx context.Context, firmwareID, orgID, userID uuid.UUID, name, deviceGroup string, totalDevices int) (*models.FirmwareRollout, error)
	TransitionRollout(ctx context.Context, rolloutID, orgID uuid.UUID, nextStatus string, actorUserID *uuid.UUID, reason *string) (*models.FirmwareRollout, error)
}

type service struct {
	repo   repository.Repository
	logger *zap.Logger
}

var allowedCommandTransitions = map[string]map[string]bool{
	"created":   {"queued": true},
	"queued":    {"delivered": true, "failed": true},
	"delivered": {"acked": true, "timed_out": true},
	"acked":     {"applied": true, "failed": true},
	"applied":   {},
	"failed":    {},
	"timed_out": {},
}

var allowedRolloutTransitions = map[string]map[string]bool{
	"scheduled":   {"in_progress": true, "cancelled": true},
	"in_progress": {"paused": true, "completed": true, "cancelled": true},
	"paused":      {"in_progress": true, "cancelled": true},
	"completed":   {},
	"cancelled":   {},
}

func normalizeCommandStatus(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "pending":
		return "queued"
	case "sent":
		return "delivered"
	case "acknowledged":
		return "acked"
	case "executed", "completed":
		return "applied"
	default:
		return s
	}
}

func normalizeRolloutStatus(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizeRolloutEventType(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "":
		return "", nil
	case "note":
		return "rollout.note", nil
	case "transition":
		return "rollout.transition", nil
	case "created":
		return "rollout.created", nil
	case "rollout.note", "rollout.transition", "rollout.created":
		return s, nil
	default:
		return "", fmt.Errorf("invalid event_type: %s", raw)
	}
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

	// Get ignition status from latest telemetry
	ignitionStatus := "unknown"
	ignitionValue := "OFF"
	latestTelemetry, err := s.repo.GetLatestTelemetry(ctx, deviceID)
	if err == nil && latestTelemetry != nil && latestTelemetry.Payload != nil {
		// Try to get ignition status from payload
		if ign, ok := latestTelemetry.Payload["ignition"].(bool); ok {
			if ign {
				ignitionValue = "ON"
			}
		}
		if ign, ok := latestTelemetry.Payload["ignition_status"].(string); ok {
			ignitionValue = ign
		}
	}
	if ignitionValue == "ON" {
		ignitionStatus = "normal"
	}

	sensors = append(sensors, models.SensorReading{
		Name:   "Ignition",
		Value:  ignitionValue,
		Status: ignitionStatus,
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

func (s *service) GetLatestDeviceTelemetry(ctx context.Context, deviceID uuid.UUID) (*models.DeviceTelemetry, error) {
	return s.repo.GetLatestTelemetry(ctx, deviceID)
}

func (s *service) SendCommand(ctx context.Context, deviceID, orgID, userID uuid.UUID, command, category string, params map[string]interface{}) (*models.DeviceCommand, error) {
	// Validate command
	command = strings.TrimSpace(command)
	category = strings.TrimSpace(category)
	if command == "" {
		return nil, NewValidationError("command", "command is required")
	}
	if category == "" {
		return nil, NewValidationError("category", "category is required")
	}

	// Validate device exists
	device, err := s.repo.GetDevice(ctx, deviceID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	if device == nil {
		return nil, ErrDeviceNotFound
	}

	// Validate device is online
	if device.Status != string(models.StatusActive) {
		return nil, fmt.Errorf("cannot send command to device with status: %s", device.Status)
	}

	cmd := &models.DeviceCommand{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		OrgID:     orgID,
		Command:   command,
		Category:  category,
		Params:    params,
		Status:    string(models.CommandQueued),
		CreatedBy: &userID,
	}

	if err := s.repo.CreateCommand(ctx, cmd); err != nil {
		return nil, err
	}

	commandStatus := "queued"
	commandDesired := map[string]interface{}{
		"last_command": map[string]interface{}{
			"id":       cmd.ID.String(),
			"command":  command,
			"category": category,
			"params":   params,
			"status":   commandStatus,
			"created":  cmd.CreatedAt.UTC().Format(time.RFC3339),
		},
	}
	if err := s.repo.UpsertDesiredState(ctx, deviceID, orgID, commandDesired, &cmd.ID, &commandStatus); err != nil {
		s.logger.Warn("Failed to upsert desired twin state on command create", zap.Error(err), zap.String("command_id", cmd.ID.String()))
	}

	s.logger.Info("Command queued",
		zap.String("device_id", deviceID.String()),
		zap.String("command", command),
		zap.String("command_id", cmd.ID.String()),
	)

	// Persist immediate dispatch transition so API clients can track command lifecycle.
	dispatchResult := map[string]interface{}{
		"dispatch_channel": "device-gateway",
		"queued_at":        cmd.CreatedAt.UTC().Format(time.RFC3339),
	}
	if err := s.repo.UpdateCommandStatus(ctx, cmd.ID, "delivered", dispatchResult, nil); err != nil {
		s.logger.Warn("Failed to transition command to sent status", zap.Error(err), zap.String("command_id", cmd.ID.String()))
	} else {
		sentAt := time.Now().UTC()
		cmd.Status = "delivered"
		cmd.SentAt = &sentAt
		cmd.Result = dispatchResult

		sentStatus := "delivered"
		commandDesired["last_command"].(map[string]interface{})["status"] = sentStatus
		commandDesired["last_command"].(map[string]interface{})["sent_at"] = sentAt.Format(time.RFC3339)
		if err := s.repo.UpsertDesiredState(ctx, deviceID, orgID, commandDesired, &cmd.ID, &sentStatus); err != nil {
			s.logger.Warn("Failed to upsert desired twin state on command dispatch", zap.Error(err), zap.String("command_id", cmd.ID.String()))
		}
	}

	return cmd, nil
}

func (s *service) GetCommandHistory(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*models.DeviceCommand, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.ListCommands(ctx, deviceID, limit, offset)
}

func (s *service) TransitionCommand(ctx context.Context, deviceID, orgID, commandID uuid.UUID, nextStatus string, result map[string]interface{}, errorMessage *string) (*models.DeviceCommand, error) {
	nextStatus = strings.TrimSpace(nextStatus)
	if nextStatus == "" {
		return nil, NewValidationError("status", "status is required")
	}

	next := normalizeCommandStatus(nextStatus)
	if _, ok := allowedCommandTransitions[next]; !ok {
		return nil, fmt.Errorf("invalid command status: %s", nextStatus)
	}

	cmd, err := s.repo.GetCommand(ctx, commandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command: %w", err)
	}
	if cmd == nil {
		return nil, ErrCommandNotFound
	}
	if cmd.DeviceID != deviceID {
		return nil, fmt.Errorf("command does not belong to device")
	}
	if cmd.OrgID != orgID {
		return nil, fmt.Errorf("command does not belong to organization")
	}

	current := normalizeCommandStatus(cmd.Status)
	if !allowedCommandTransitions[current][next] {
		return nil, fmt.Errorf("invalid transition from %s to %s", current, next)
	}

	if err := s.repo.UpdateCommandStatus(ctx, commandID, next, result, errorMessage); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetCommand(ctx, commandID)
	if err != nil {
		return nil, err
	}

	desired := map[string]interface{}{
		"last_command": map[string]interface{}{
			"id":       commandID.String(),
			"command":  updated.Command,
			"category": updated.Category,
			"status":   next,
		},
	}
	if err := s.repo.UpsertDesiredState(ctx, updated.DeviceID, updated.OrgID, desired, &commandID, &next); err != nil {
		s.logger.Warn("Failed to update twin desired state from command transition", zap.Error(err), zap.String("command_id", commandID.String()))
	}

	logType := "info"
	if next == "failed" || next == "timed_out" {
		logType = "warning"
	}
	logMessage := fmt.Sprintf("Command %s transitioned from %s to %s", updated.Command, current, next)
	logMeta := map[string]interface{}{
		"command_id":     commandID.String(),
		"device_id":      updated.DeviceID.String(),
		"previous_state": current,
		"next_state":     next,
		"result":         result,
	}
	if errorMessage != nil {
		logMeta["error_message"] = *errorMessage
	}
	if err := s.repo.CreateDiagnosticsLog(ctx, &models.DiagnosticsLog{
		ID:       uuid.New(),
		DeviceID: updated.DeviceID,
		OrgID:    updated.OrgID,
		LogType:  logType,
		Source:   "command-lifecycle",
		Message:  logMessage,
		Metadata: logMeta,
	}); err != nil {
		s.logger.Warn("Failed to write command lifecycle diagnostics log", zap.Error(err), zap.String("command_id", commandID.String()))
	}

	return updated, nil
}

func (s *service) ListDeviceTwins(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DeviceTwin, error) {
	if limit == 0 {
		limit = 50
	}
	twins, err := s.repo.ListDeviceTwins(ctx, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, twin := range twins {
		s.reconcileTwinDrift(twin)
	}
	return twins, nil
}

func (s *service) GetDeviceTwin(ctx context.Context, deviceID, orgID uuid.UUID) (*models.DeviceTwin, error) {
	twin, err := s.repo.GetDeviceTwin(ctx, deviceID, orgID)
	if err != nil {
		return nil, err
	}
	s.reconcileTwinDrift(twin)
	return twin, nil
}

func (s *service) UpdateDesiredState(ctx context.Context, deviceID, orgID uuid.UUID, desiredState map[string]interface{}, commandID *uuid.UUID, commandStatus *string) error {
	if len(desiredState) == 0 {
		return fmt.Errorf("desired state is required")
	}
	return s.repo.UpsertDesiredState(ctx, deviceID, orgID, desiredState, commandID, commandStatus)
}

func (s *service) UpdateReportedState(ctx context.Context, deviceID, orgID uuid.UUID, reportedState map[string]interface{}, lastSeen *time.Time) error {
	if len(reportedState) == 0 {
		return fmt.Errorf("reported state is required")
	}
	return s.repo.UpsertReportedState(ctx, deviceID, orgID, reportedState, lastSeen)
}

func (s *service) reconcileTwinDrift(twin *models.DeviceTwin) {
	if twin == nil {
		return
	}
	if len(twin.DesiredState) == 0 || len(twin.ReportedState) == 0 {
		twin.DriftStatus = "unknown"
		reason := "desired or reported state missing"
		twin.DriftReason = &reason
		return
	}

	desiredJSON, _ := json.Marshal(twin.DesiredState)
	reportedJSON, _ := json.Marshal(twin.ReportedState)
	if string(desiredJSON) == string(reportedJSON) {
		twin.DriftStatus = "in_sync"
		twin.DriftReason = nil
		return
	}
	twin.DriftStatus = "drifted"
	reason := "desired state differs from reported state"
	twin.DriftReason = &reason
}

func (s *service) ListFirmware(ctx context.Context, limit, offset int) ([]*models.FirmwarePackage, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.ListFirmware(ctx, limit, offset)
}

func (s *service) CreateFirmware(ctx context.Context, version, hwTarget, fileURL string, fileSize int, checksum, notes *string, isCritical bool) (*models.FirmwarePackage, error) {
	// Validate inputs
	version = strings.TrimSpace(version)
	hwTarget = strings.TrimSpace(hwTarget)
	fileURL = strings.TrimSpace(fileURL)

	if version == "" {
		return nil, NewValidationError("version", "version is required")
	}
	if hwTarget == "" {
		return nil, NewValidationError("hardware_target", "hardware_target is required")
	}
	if fileURL == "" {
		return nil, NewValidationError("file_url", "file_url is required")
	}
	if fileSize <= 0 {
		return nil, NewValidationError("file_size", "file_size must be positive")
	}

	// Check if version already exists for this hardware target
	// _, _ = s.repo.GetFirmware(ctx, uuid.Nil) // Would need a GetFirmwareByVersion method
	// For now, validation is done in repository via unique constraint

	fw := &models.FirmwarePackage{
		ID:             uuid.New(),
		Version:        version,
		HardwareTarget: hwTarget,
		FileURL:        fileURL,
		FileSize:       &fileSize,
		Checksum:       checksum,
		ReleaseNotes:   notes,
		Status:         string(models.FirmwareDraft),
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

func (s *service) GetRollout(ctx context.Context, rolloutID, orgID uuid.UUID) (*models.FirmwareRollout, error) {
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return nil, err
	}
	if rollout.OrgID != orgID {
		return nil, fmt.Errorf("rollout does not belong to organization")
	}
	return rollout, nil
}

func (s *service) ListRolloutEvents(ctx context.Context, rolloutID, orgID uuid.UUID, eventType string, limit, offset int) ([]*models.FirmwareRolloutEvent, error) {
	if limit == 0 {
		limit = 50
	}
	normalizedEventType, err := normalizeRolloutEventType(eventType)
	if err != nil {
		return nil, err
	}
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return nil, err
	}
	if rollout.OrgID != orgID {
		return nil, fmt.Errorf("rollout does not belong to organization")
	}
	return s.repo.ListRolloutEvents(ctx, rolloutID, orgID, normalizedEventType, limit, offset)
}

func (s *service) AddRolloutNote(ctx context.Context, rolloutID, orgID uuid.UUID, actorUserID *uuid.UUID, message string, metadata map[string]interface{}) (*models.FirmwareRolloutEvent, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return nil, fmt.Errorf("message is required")
	}
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return nil, err
	}
	if rollout.OrgID != orgID {
		return nil, fmt.Errorf("rollout does not belong to organization")
	}

	event := &models.FirmwareRolloutEvent{
		ID:        uuid.New(),
		RolloutID: rolloutID,
		OrgID:     orgID,
		EventType: "rollout.note",
		Message:   trimmed,
		Metadata:  metadata,
	}
	if actorUserID != nil && *actorUserID != uuid.Nil {
		event.ActorUserID = actorUserID
	}
	if err := s.repo.CreateRolloutEvent(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *service) CreateRollout(ctx context.Context, firmwareID, orgID, userID uuid.UUID, name, deviceGroup string, totalDevices int) (*models.FirmwareRollout, error) {
	// Validate inputs
	name = strings.TrimSpace(name)
	deviceGroup = strings.TrimSpace(deviceGroup)

	if name == "" {
		return nil, NewValidationError("name", "name is required")
	}
	if deviceGroup == "" {
		return nil, NewValidationError("device_group", "device_group is required")
	}
	if totalDevices <= 0 {
		return nil, NewValidationError("total_devices", "total_devices must be positive")
	}

	// Validate firmware exists and is published
	firmware, err := s.repo.GetFirmware(ctx, firmwareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get firmware: %w", err)
	}
	if firmware == nil {
		return nil, ErrFirmwareNotFound
	}
	if firmware.Status != string(models.FirmwarePublished) {
		return nil, ErrFirmwareNotPublished
	}

	rollout := &models.FirmwareRollout{
		ID:              uuid.New(),
		FirmwareID:      firmwareID,
		OrgID:           orgID,
		Name:            &name,
		DeviceGroup:     &deviceGroup,
		TotalDevices:    totalDevices,
		Status:          string(models.RolloutScheduled),
		RolloutStrategy: "gradual",
		MaxConcurrent:   10,
		CreatedBy:       &userID,
	}

	if err := s.repo.CreateRollout(ctx, rollout); err != nil {
		return nil, err
	}

	rolloutName := name
	createEvent := &models.FirmwareRolloutEvent{
		ID:        uuid.New(),
		RolloutID: rollout.ID,
		OrgID:     orgID,
		EventType: "rollout.created",
		ToStatus:  strPtr("scheduled"),
		Message:   "Rollout created",
		Metadata: map[string]interface{}{
			"name":          rolloutName,
			"firmware_id":   firmwareID.String(),
			"device_group":  deviceGroup,
			"total_devices": totalDevices,
		},
	}
	if userID != uuid.Nil {
		createEvent.ActorUserID = &userID
	}
	if err := s.repo.CreateRolloutEvent(ctx, createEvent); err != nil {
		s.logger.Warn("Failed to write rollout create event", zap.Error(err), zap.String("rollout_id", rollout.ID.String()))
	}

	s.logger.Info("Rollout created",
		zap.String("rollout_id", rollout.ID.String()),
		zap.String("firmware_id", firmwareID.String()),
		zap.Int("total_devices", totalDevices),
	)

	return rollout, nil
}

func (s *service) TransitionRollout(ctx context.Context, rolloutID, orgID uuid.UUID, nextStatus string, actorUserID *uuid.UUID, reason *string) (*models.FirmwareRollout, error) {
	next := normalizeRolloutStatus(nextStatus)
	if _, ok := allowedRolloutTransitions[next]; !ok {
		return nil, fmt.Errorf("invalid rollout status: %s", nextStatus)
	}

	currentRollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return nil, err
	}
	if currentRollout.OrgID != orgID {
		return nil, fmt.Errorf("rollout does not belong to organization")
	}

	current := normalizeRolloutStatus(currentRollout.Status)
	if !allowedRolloutTransitions[current][next] {
		return nil, fmt.Errorf("invalid rollout transition from %s to %s", current, next)
	}

	if err := s.repo.UpdateRolloutStatus(ctx, rolloutID, next); err != nil {
		return nil, err
	}

	transitionEvent := &models.FirmwareRolloutEvent{
		ID:         uuid.New(),
		RolloutID:  rolloutID,
		OrgID:      orgID,
		EventType:  "rollout.transition",
		FromStatus: strPtr(current),
		ToStatus:   strPtr(next),
		Message:    fmt.Sprintf("Rollout transitioned from %s to %s", current, next),
		Metadata: map[string]interface{}{
			"from_status": current,
			"to_status":   next,
		},
	}
	if actorUserID != nil && *actorUserID != uuid.Nil {
		transitionEvent.ActorUserID = actorUserID
	}
	if reason != nil {
		trimmed := strings.TrimSpace(*reason)
		if trimmed != "" {
			transitionEvent.Metadata["reason"] = trimmed
			transitionEvent.Message = fmt.Sprintf("Rollout transitioned from %s to %s (%s)", current, next, trimmed)
		}
	}
	if err := s.repo.CreateRolloutEvent(ctx, transitionEvent); err != nil {
		s.logger.Warn("Failed to write rollout transition event", zap.Error(err), zap.String("rollout_id", rolloutID.String()))
	}

	updated, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
