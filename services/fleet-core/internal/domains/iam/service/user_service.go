package service

import (
	"context"
	"errors"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidStatus = errors.New("invalid_user_status")
)

type UserService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	roleRepo   *repository.RoleRepository
	auditRepo *repository.AuditRepository
	log      *zap.Logger
}

func NewUserService(
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	roleRepo *repository.RoleRepository,
	auditRepo *repository.AuditRepository,
	log *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		roleRepo:   roleRepo,
		auditRepo: auditRepo,
		log:       log,
	}
}

// ListUsers returns a paginated list of users with optional filtering.
func (s *UserService) ListUsers(ctx context.Context, params models.ListUsersParams) (*models.ListUsersResponse, error) {
	resp, err := s.userRepo.ListUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	// Enrich each user with their role names
	for i := range resp.Users {
		roles, err := s.roleRepo.GetUserRoles(ctx, resp.Users[i].ID, uuid.Nil)
		if err == nil && len(roles) > 0 {
			roleNames := make([]string, 0, len(roles))
			for _, r := range roles {
				roleNames = append(roleNames, r.Name)
			}
			resp.Users[i].Roles = roleNames
		} else {
			resp.Users[i].Roles = []string{}
		}
	}

	return resp, nil
}

// GetUser returns a single user by ID.
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUserStatus changes a user's status (active, disabled, locked).
func (s *UserService) UpdateUserStatus(ctx context.Context, userID, tenantID uuid.UUID, status string, actorID uuid.UUID) error {
	var userStatus models.UserStatus
	switch status {
	case "active":
		userStatus = models.UserStatusActive
	case "disabled":
		userStatus = models.UserStatusDisabled
	case "locked":
		userStatus = models.UserStatusLocked
	default:
		return ErrInvalidStatus
	}

	if err := s.userRepo.UpdateStatus(ctx, userID, tenantID, userStatus); err != nil {
		return err
	}

	// Log the audit event
	s.auditRepo.Log(ctx, models.AuditLog{
		UserID:   actorID,
		Action:   "user.status_changed",
		Resource: "user",
		Data: map[string]interface{}{
			"target_user_id": userID.String(),
			"new_status":     status,
		},
	})

	s.log.Info("user_status_changed",
		zap.String("user_id", userID.String()),
		zap.String("status", status),
		zap.String("actor_id", actorID.String()),
	)
	return nil
}

// DeleteUser soft-deletes a user by disabling their account.
func (s *UserService) DeleteUser(ctx context.Context, userID, tenantID uuid.UUID, actorID uuid.UUID) error {
	if err := s.userRepo.SoftDelete(ctx, userID, tenantID); err != nil {
		return err
	}

	s.auditRepo.Log(ctx, models.AuditLog{
		UserID:   actorID,
		Action:   "user.deleted",
		Resource: "user",
		Data: map[string]interface{}{
			"target_user_id": userID.String(),
		},
	})

	s.log.Info("user_deleted",
		zap.String("user_id", userID.String()),
		zap.String("actor_id", actorID.String()),
	)
	return nil
}

// GetUserAuditLog returns the audit log entries for a specific user.
func (s *UserService) GetUserAuditLog(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.auditRepo.GetUserAuditLog(ctx, userID, limit)
}

// UpdateUserPreferences updates user UI preferences (sidebar state, language).
func (s *UserService) UpdateUserPreferences(ctx context.Context, userID, tenantID uuid.UUID, sidebarCollapsed *bool, language *string, actorID uuid.UUID) error {
	if err := s.userRepo.UpdatePreferences(ctx, userID, tenantID, sidebarCollapsed, language); err != nil {
		return err
	}

	s.auditRepo.Log(ctx, models.AuditLog{
		UserID:   actorID,
		Action:   "user.preferences_updated",
		Resource: "user",
		Data: map[string]interface{}{
			"target_user_id":    userID.String(),
			"sidebar_collapsed": sidebarCollapsed,
			"language":          language,
		},
	})

	s.log.Info("user_preferences_updated",
		zap.String("user_id", userID.String()),
		zap.Bool("sidebar_collapsed", sidebarCollapsed != nil && *sidebarCollapsed),
		zap.String("actor_id", actorID.String()),
	)
	return nil
}

// CreateUser handles direct admin-led user creation with a temporary password.
func (s *UserService) CreateUser(ctx context.Context, tenantID uuid.UUID, req models.CreateUserRequest, actorID uuid.UUID) (*models.User, error) {
	// Hash the temporary password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:         req.Email,
		Name:          req.Name,
		PasswordHash:  string(hash),
		Status:        models.UserStatusActive,
		EmailVerified: true, // Admin created users are trusted
		MustResetPassword: true,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Assign roles
	var primaryRoleID uuid.UUID
	for i, roleID := range req.RoleIDs {
		if i == 0 {
			primaryRoleID = roleID
		}
		if err := s.roleRepo.AssignRole(ctx, user.ID, roleID, tenantID, uuid.Nil); err != nil {
			s.log.Error("failed to assign role during direct user creation",
				zap.String("role_id", roleID.String()),
				zap.Error(err),
			)
		}
	}

	// Link user to tenant for visibility and login
	if err := s.tenantRepo.AddUserToTenant(ctx, user.ID, tenantID, "Member", primaryRoleID); err != nil {
		s.log.Error("failed to link user to tenant", zap.Error(err))
	}

	// Log audit
	s.auditRepo.Log(ctx, models.AuditLog{
		UserID:   actorID,
		Action:   "user.created_direct",
		Resource: "user",
		Data: map[string]interface{}{
			"target_user_id": user.ID.String(),
			"email":          user.Email,
			"tenant_id":      tenantID.String(),
		},
	})

	s.log.Info("user_created_direct",
		zap.String("user_id", user.ID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.String("actor_id", actorID.String()),
	)

	return user, nil
}
