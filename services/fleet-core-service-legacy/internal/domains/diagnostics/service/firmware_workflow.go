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

// FirmwareWorkflowService handles firmware lifecycle and rollout workflows
type FirmwareWorkflowService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewFirmwareWorkflowService creates a new firmware workflow service
func NewFirmwareWorkflowService(repo repository.Repository, logger *zap.Logger) *FirmwareWorkflowService {
	return &FirmwareWorkflowService{
		repo:   repo,
		logger: logger,
	}
}

// PublishFirmware publishes a firmware package
func (s *FirmwareWorkflowService) PublishFirmware(ctx context.Context, firmwareID, userID uuid.UUID) error {
	firmware, err := s.repo.GetFirmware(ctx, firmwareID)
	if err != nil {
		return fmt.Errorf("failed to get firmware: %w", err)
	}
	if firmware == nil {
		return ErrFirmwareNotFound
	}

	if firmware.Status != string(models.FirmwareDraft) {
		return fmt.Errorf("can only publish draft firmware, current status: %s", firmware.Status)
	}

	now := time.Now()
	firmware.Status = string(models.FirmwarePublished)
	firmware.ReleasedAt = &now

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      uuid.Nil, // Firmware is global
		EntityType: "firmware",
		EntityID:   firmwareID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": models.FirmwareDraft},
		NewValue:   map[string]interface{}{"status": models.FirmwarePublished},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     strPtr("Firmware published"),
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for firmware publish", zap.Error(err))
	}

	s.logger.Info("Firmware published",
		zap.String("firmware_id", firmwareID.String()),
		zap.String("version", firmware.Version),
	)

	return nil
}

// DeprecateFirmware deprecates a firmware package
func (s *FirmwareWorkflowService) DeprecateFirmware(ctx context.Context, firmwareID, userID uuid.UUID, reason string) error {
	firmware, err := s.repo.GetFirmware(ctx, firmwareID)
	if err != nil {
		return fmt.Errorf("failed to get firmware: %w", err)
	}
	if firmware == nil {
		return ErrFirmwareNotFound
	}

	if firmware.Status != string(models.FirmwarePublished) {
		return fmt.Errorf("can only deprecate published firmware, current status: %s", firmware.Status)
	}

	firmware.Status = string(models.FirmwareDeprecated)

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      uuid.Nil,
		EntityType: "firmware",
		EntityID:   firmwareID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": models.FirmwarePublished},
		NewValue:   map[string]interface{}{"status": models.FirmwareDeprecated},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     &reason,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for firmware deprecate", zap.Error(err))
	}

	s.logger.Info("Firmware deprecated",
		zap.String("firmware_id", firmwareID.String()),
		zap.String("version", firmware.Version),
		zap.String("reason", reason),
	)

	return nil
}

// StartRollout starts a firmware rollout
func (s *FirmwareWorkflowService) StartRollout(ctx context.Context, rolloutID, orgID, userID uuid.UUID) error {
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return fmt.Errorf("failed to get rollout: %w", err)
	}
	if rollout == nil {
		return ErrRolloutNotFound
	}

	if rollout.OrgID != orgID {
		return fmt.Errorf("rollout does not belong to organization")
	}

	if rollout.Status != string(models.RolloutScheduled) {
		return fmt.Errorf("can only start scheduled rollouts, current status: %s", rollout.Status)
	}

	// Verify firmware is published
	firmware, err := s.repo.GetFirmware(ctx, rollout.FirmwareID)
	if err != nil {
		return fmt.Errorf("failed to get firmware: %w", err)
	}
	if firmware == nil || firmware.Status != string(models.FirmwarePublished) {
		return ErrFirmwareNotPublished
	}

	rollout.Status = string(models.RolloutInProgress)
	startedAt := time.Now()
	rollout.StartedAt = &startedAt

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityType: "rollout",
		EntityID:   rolloutID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": models.RolloutScheduled},
		NewValue:   map[string]interface{}{"status": models.RolloutInProgress},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     strPtr("Rollout started"),
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for rollout start", zap.Error(err))
	}

	s.logger.Info("Rollout started",
		zap.String("rollout_id", rolloutID.String()),
		zap.Int("total_devices", rollout.TotalDevices),
	)

	return nil
}

// PauseRollout pauses a firmware rollout
func (s *FirmwareWorkflowService) PauseRollout(ctx context.Context, rolloutID, orgID, userID uuid.UUID, reason string) error {
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return fmt.Errorf("failed to get rollout: %w", err)
	}
	if rollout == nil {
		return ErrRolloutNotFound
	}

	if rollout.Status != string(models.RolloutInProgress) {
		return fmt.Errorf("can only pause in-progress rollouts, current status: %s", rollout.Status)
	}

	rollout.Status = string(models.RolloutPaused)

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityType: "rollout",
		EntityID:   rolloutID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": models.RolloutInProgress},
		NewValue:   map[string]interface{}{"status": models.RolloutPaused},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     &reason,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for rollout pause", zap.Error(err))
	}

	s.logger.Info("Rollout paused",
		zap.String("rollout_id", rolloutID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// CompleteRollout completes a firmware rollout
func (s *FirmwareWorkflowService) CompleteRollout(ctx context.Context, rolloutID, orgID, userID uuid.UUID) error {
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return fmt.Errorf("failed to get rollout: %w", err)
	}
	if rollout == nil {
		return ErrRolloutNotFound
	}

	if rollout.Status != string(models.RolloutInProgress) && rollout.Status != string(models.RolloutPaused) {
		return fmt.Errorf("can only complete in-progress or paused rollouts, current status: %s", rollout.Status)
	}

	rollout.Status = string(models.RolloutCompleted)
	completedAt := time.Now()
	rollout.CompletedAt = &completedAt

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityType: "rollout",
		EntityID:   rolloutID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": rollout.Status},
		NewValue:   map[string]interface{}{"status": models.RolloutCompleted},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     strPtr("Rollout completed"),
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for rollout complete", zap.Error(err))
	}

	s.logger.Info("Rollout completed",
		zap.String("rollout_id", rolloutID.String()),
		zap.Int("completed_devices", rollout.CompletedDevices),
		zap.Int("failed_devices", rollout.FailedDevices),
	)

	return nil
}

// CancelRollout cancels a firmware rollout
func (s *FirmwareWorkflowService) CancelRollout(ctx context.Context, rolloutID, orgID, userID uuid.UUID, reason string) error {
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return fmt.Errorf("failed to get rollout: %w", err)
	}
	if rollout == nil {
		return ErrRolloutNotFound
	}

	if rollout.Status == string(models.RolloutCompleted) || rollout.Status == string(models.RolloutCancelled) {
		return fmt.Errorf("cannot cancel completed or cancelled rollout")
	}

	rollout.Status = string(models.RolloutCancelled)

	// Create audit log
	audit := &models.DiagnosticsAuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		EntityType: "rollout",
		EntityID:   rolloutID,
		Action:     "status_changed",
		OldValue:   map[string]interface{}{"status": rollout.Status},
		NewValue:   map[string]interface{}{"status": models.RolloutCancelled},
		UserID:     &userID,
		UserType:   "admin",
		Reason:     &reason,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, audit); err != nil {
		s.logger.Warn("Failed to create audit log for rollout cancel", zap.Error(err))
	}

	s.logger.Info("Rollout cancelled",
		zap.String("rollout_id", rolloutID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// UpdateRolloutProgress updates rollout progress
func (s *FirmwareWorkflowService) UpdateRolloutProgress(ctx context.Context, rolloutID, orgID uuid.UUID, completedDevices, failedDevices int) error {
	rollout, err := s.repo.GetRollout(ctx, rolloutID)
	if err != nil {
		return fmt.Errorf("failed to get rollout: %w", err)
	}
	if rollout == nil {
		return ErrRolloutNotFound
	}

	if rollout.OrgID != orgID {
		return fmt.Errorf("rollout does not belong to organization")
	}

	// Validate progress
	if completedDevices+failedDevices > rollout.TotalDevices {
		return fmt.Errorf("completed + failed devices cannot exceed total devices")
	}

	// Auto-complete if all devices are processed
	if completedDevices+failedDevices == rollout.TotalDevices && rollout.Status == string(models.RolloutInProgress) {
		return s.CompleteRollout(ctx, rolloutID, orgID, uuid.Nil)
	}

	return s.repo.UpdateRolloutProgress(ctx, rolloutID, completedDevices, failedDevices, rollout.Status)
}

func strPtr(s string) *string {
	return &s
}
