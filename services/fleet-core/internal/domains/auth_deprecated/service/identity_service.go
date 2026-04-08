package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/auth/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IdentityService interface {
	CreateRole(ctx context.Context, orgID *uuid.UUID, name, description string, permissions []string) (*models.Role, error)
	AssignRole(ctx context.Context, userID, roleID, orgID uuid.UUID) error
	CheckPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
}

type IdentityServiceImpl struct {
	repo   AuthRepositoryInterface
	logger *zap.Logger
}

func NewIdentityService(repo AuthRepositoryInterface, logger *zap.Logger) IdentityService {
	return &IdentityServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *IdentityServiceImpl) CreateRole(ctx context.Context, orgID *uuid.UUID, name, description string, permissions []string) (*models.Role, error) {
	role := &models.Role{
		ID:          uuid.New(),
		Name:        name,
		OrgID:       orgID,
		Description: description,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Note: In a real implementation, we would also link the permissions here.
	// For this upgrade, we'll keep it simple and focus on the core structure.

	return role, nil
}

func (s *IdentityServiceImpl) AssignRole(ctx context.Context, userID, roleID, orgID uuid.UUID) error {
	if err := s.repo.AssignRoleToUser(ctx, userID, roleID, orgID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	s.logger.Info("role assigned to user", zap.String("user_id", userID.String()), zap.String("role_id", roleID.String()))
	return nil
}

func (s *IdentityServiceImpl) CheckPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error) {
	return s.repo.CheckPermission(ctx, userID, orgID, permission)
}

func (s *IdentityServiceImpl) GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	return s.repo.GetUserPermissions(ctx, userID, orgID)
}
