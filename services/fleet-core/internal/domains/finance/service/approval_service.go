package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ApprovalService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewApprovalService(repo *repository.FinanceRepository, logger *zap.Logger) *ApprovalService {
	return &ApprovalService{repo: repo, logger: logger}
}

// InitiateApproval starts a new approval process for an entity
func (s *ApprovalService) InitiateApproval(ctx context.Context, entityID, entityType string, orgID string, amount float64) error {
	// 1. Find matching config
	config, err := s.repo.GetMatchingApprovalConfig(ctx, orgID, entityType, amount)
	if err != nil {
		return fmt.Errorf("failed to find approval config: %w", err)
	}

	if config == nil {
		s.logger.Info("No specific approval config found, falling back to basic approval",
			zap.String("entity_type", entityType), zap.Float64("amount", amount))
		return nil
	}

	// 2. Create approval instance
	instance := &models.ApprovalInstance{
		ID:                uuid.New().String(),
		EntityID:          entityID,
		EntityType:        entityType,
		ConfigID:          config.ID,
		CurrentLevelIndex: 0,
		Status:            "PENDING",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.CreateApprovalInstance(ctx, instance); err != nil {
		return fmt.Errorf("failed to create approval instance: %w", err)
	}

	s.logger.Info("Approval initiated", zap.String("instance_id", instance.ID), zap.String("entity_id", entityID))
	return nil
}

// ProcessAction records an approval or rejection
func (s *ApprovalService) ProcessAction(ctx context.Context, entityID, entityType, approverID, action, comments string) (string, error) {
	// 1. Get current instance
	instance, err := s.repo.GetApprovalInstanceEffectively(ctx, entityID, entityType)
	if err != nil {
		return "", err
	}
	if instance == nil {
		return "", fmt.Errorf("no active approval instance found for %s %s", entityType, entityID)
	}

	// 2. Get config
	config, err := s.repo.GetApprovalConfigByID(ctx, instance.ConfigID)
	if err != nil {
		return "", err
	}

	// 3. Record action
	approvalAction := &models.ApprovalAction{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		ApproverID: approverID,
		Action:     action,
		LevelIndex: instance.CurrentLevelIndex,
		Comments:   comments,
		ActionTime: time.Now(),
	}

	if err := s.repo.RecordApprovalAction(ctx, approvalAction); err != nil {
		return "", err
	}

	// 4. Update instance state
	newStatus := instance.Status
	if action == "REJECTED" {
		instance.Status = "REJECTED"
		newStatus = "REJECTED"
	} else if action == "APPROVED" {
		if instance.CurrentLevelIndex >= len(config.Levels)-1 {
			instance.Status = "APPROVED"
			newStatus = "APPROVED"
		} else {
			instance.CurrentLevelIndex++
			// Still PENDING at next level
		}
	}
	instance.UpdatedAt = time.Now()

	if err := s.repo.UpdateApprovalInstance(ctx, instance); err != nil {
		return "", err
	}

	return newStatus, nil
}
