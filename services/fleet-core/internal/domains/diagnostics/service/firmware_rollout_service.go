package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/repository"
	"github.com/google/uuid"
)

type RolloutState string

const (
	RolloutStateScheduled  RolloutState = "SCHEDULED"
	RolloutStateInProgress RolloutState = "IN_PROGRESS"
	RolloutStateCompleted  RolloutState = "COMPLETED"
	RolloutStateFailed     RolloutState = "FAILED"
	RolloutStateCancelled  RolloutState = "CANCELLED"
)

type FirmwareRolloutService struct {
	repo *repository.DiagnosticsRepository
}

func NewFirmwareRolloutService(repo *repository.DiagnosticsRepository) *FirmwareRolloutService {
	return &FirmwareRolloutService{repo: repo}
}

func (s *FirmwareRolloutService) StartRollout(ctx context.Context, rolloutID uuid.UUID) error {
	now := time.Now()
	ev := &models.FirmwareRolloutEvent{
		RolloutID: &rolloutID,
		EventType: ptrStr("rollout.started"),
		CreatedAt: &now,
	}

	if err := s.repo.UpdateRolloutStatus(ctx, rolloutID, string(RolloutStateInProgress)); err != nil {
		return fmt.Errorf("failed to start rollout: %w", err)
	}

	if err := s.repo.InsertRolloutEvent(ctx, ev); err != nil {
		return fmt.Errorf("failed to log rollout event: %w", err)
	}

	return nil
}

func (s *FirmwareRolloutService) CancelRollout(ctx context.Context, rolloutID uuid.UUID) error {
	now := time.Now()
	ev := &models.FirmwareRolloutEvent{
		RolloutID: &rolloutID,
		EventType: ptrStr("rollout.cancelled"),
		CreatedAt: &now,
	}

	if err := s.repo.UpdateRolloutStatus(ctx, rolloutID, string(RolloutStateCancelled)); err != nil {
		return fmt.Errorf("failed to cancel rollout: %w", err)
	}

	if err := s.repo.InsertRolloutEvent(ctx, ev); err != nil {
		return fmt.Errorf("failed to log rollout event: %w", err)
	}

	return nil
}

func (s *FirmwareRolloutService) EvaluateRolloutSafety(ctx context.Context, rollout *models.FirmwareRollout, successCount int, failureCount int) error {
	if rollout.FailureThreshold != nil {
		totalAttempted := successCount + failureCount
		if totalAttempted > 0 {
			rate := float64(failureCount) / float64(totalAttempted) * 100.0
			if rate > float64(*rollout.FailureThreshold) {
				s.PauseRollout(ctx, rollout.ID)

				if rollout.RollbackEnabled != nil && *rollout.RollbackEnabled {
					s.TriggerRollback(ctx, rollout.ID)
				}

				return fmt.Errorf("rollout halted: failure rate %.2f%% exceeded threshold %d%%", rate, *rollout.FailureThreshold)
			}
		}
	}
	return nil
}

func (s *FirmwareRolloutService) PauseRollout(ctx context.Context, rolloutID uuid.UUID) error {
	// e.g. status transition to PAUSED/FAILED
	return nil
}

func (s *FirmwareRolloutService) TriggerRollback(ctx context.Context, rolloutID uuid.UUID) error {
	// Revert firmware config across twins for affected devices
	return nil
}

func ptrStr(v string) *string {
	return &v
}
