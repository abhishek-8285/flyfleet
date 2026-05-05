package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInviteNotFound    = errors.New("invite_not_found")
	ErrInviteExpired     = errors.New("invite_expired")
	ErrInviteAlreadyUsed = errors.New("invite_already_used")
	ErrInviteRevoked     = errors.New("invite_revoked")
	ErrDuplicateInvite   = errors.New("duplicate_invite")
)

type Inviter interface {
	SendInvitation(ctx context.Context, to, token, message, tenantName, inviterName string) error
}

type InvitationService struct {
	invitationRepo *repository.InvitationRepository
	userRepo       *repository.UserRepository
	tenantRepo     *repository.TenantRepository
	roleRepo       *repository.RoleRepository
	auditRepo      *repository.AuditRepository
	emailer        Inviter
	log            *zap.Logger
}

func NewInvitationService(
	invitationRepo *repository.InvitationRepository,
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	roleRepo *repository.RoleRepository,
	auditRepo *repository.AuditRepository,
	emailer Inviter,
	log *zap.Logger,
) *InvitationService {
	return &InvitationService{
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		roleRepo:       roleRepo,
		auditRepo:      auditRepo,
		emailer:        emailer,
		log:            log,
	}
}

// CreateInvite creates a new invitation and sends the email.
func (s *InvitationService) CreateInvite(ctx context.Context, tenantID, invitedBy uuid.UUID, email string, roleIDs []uuid.UUID, message string) (*models.Invitation, error) {
	// Check for existing pending invite for this email/org
	exists, err := s.invitationRepo.ExistsForEmail(ctx, tenantID, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateInvite
	}

	// Get tenant info for email
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	inv := &models.Invitation{
		TenantID:     tenantID,
		Email:     email,
		InvitedBy: invitedBy,
		RoleIDs:   roleIDs,
		Message:   message,
		Status:    models.InvitationPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 day default
	}

	if err := s.invitationRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Get inviter name for email
	inviter, err := s.userRepo.GetByID(ctx, invitedBy)
	inviterName := ""
	if err == nil && inviter != nil {
		inviterName = inviter.Name
	}

	// Send invitation email
	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.emailer.SendInvitation(emailCtx, email, inv.Token, message, tenant.Name, inviterName); err != nil {
			s.log.Error("failed to send invitation email",
				zap.String("email", email),
				zap.Error(err),
			)
		}
	}()

	s.log.Info("invite_created",
		zap.String("invite_id", inv.ID.String()),
		zap.String("email", email),
		zap.String("tenant_id", tenantID.String()),
	)

	return inv, nil
}

// AcceptInvite validates a token, creates the user account, and assigns roles.
func (s *InvitationService) AcceptInvite(ctx context.Context, token, name, password string) (*models.User, error) {
	inv, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInviteNotFound
	}

	// Validate status
	switch inv.Status {
	case models.InvitationPending:
		// OK
	case models.InvitationAccepted:
		return nil, ErrInviteAlreadyUsed
	case models.InvitationRevoked:
		return nil, ErrInviteRevoked
	case models.InvitationExpired:
		return nil, ErrInviteExpired
	}

	// Check expiry
	if time.Now().After(inv.ExpiresAt) {
		s.invitationRepo.Revoke(ctx, inv.ID)
		return nil, ErrInviteExpired
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:         inv.Email,
		Name:          name,
		PasswordHash:  string(hash),
		Status:        models.UserStatusActive,
		EmailVerified: true,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Assign roles
	var primaryRoleID uuid.UUID
	for i, roleID := range inv.RoleIDs {
		if i == 0 {
			primaryRoleID = roleID
		}
		if err := s.roleRepo.AssignRole(ctx, user.ID, roleID, inv.TenantID, uuid.Nil); err != nil {
			s.log.Error("failed to assign role during invite acceptance",
				zap.String("role_id", roleID.String()),
				zap.Error(err),
			)
		}
	}

	// Link user to tenant for visibility and login
	if err := s.tenantRepo.AddUserToTenant(ctx, user.ID, inv.TenantID, "Member", primaryRoleID); err != nil {
		s.log.Error("failed to link user to tenant during invite accept", zap.Error(err))
	}

	// Mark invite as accepted
	s.invitationRepo.Accept(ctx, inv.ID, user.ID)

	s.log.Info("invite_accepted",
		zap.String("invite_id", inv.ID.String()),
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	return user, nil
}

// ResendInvite extends the expiration and re-sends the invitation email.
func (s *InvitationService) ResendInvite(ctx context.Context, inviteID, actorID uuid.UUID) error {
	inv, err := s.invitationRepo.GetByToken(ctx, "")
	if err != nil {
		return err
	}

	// We need to look up by ID — GetByToken uses token so we use a workaround.
	// In production, add a GetByID method to the repo. For now, use ListPending.
	invites, err := s.invitationRepo.ListPending(ctx, inv.TenantID)
	if err != nil {
		return err
	}

	var found *models.Invitation
	for i := range invites {
		if invites[i].ID == inviteID {
			found = &invites[i]
			break
		}
	}
	if found == nil {
		return ErrInviteNotFound
	}

	newExpiry := time.Now().Add(7 * 24 * time.Hour)
	if err := s.invitationRepo.Resend(ctx, inviteID, newExpiry); err != nil {
		return err
	}

	// Re-send email
	tenant, err := s.tenantRepo.GetTenantByID(ctx, found.TenantID)
	tenantName := ""
	if err == nil && tenant != nil {
		tenantName = tenant.Name
	}

	inviter, err := s.userRepo.GetByID(ctx, actorID)
	inviterName := ""
	if err == nil && inviter != nil {
		inviterName = inviter.Name
	}

	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.emailer.SendInvitation(emailCtx, found.Email, found.Token, found.Message, tenantName, inviterName); err != nil {
			s.log.Error("failed to resend invitation email", zap.Error(err))
		}
	}()

	s.log.Info("invite_resent",
		zap.String("invite_id", inviteID.String()),
		zap.Time("new_expires_at", newExpiry),
	)

	return nil
}

// RevokeInvite revokes a pending invitation.
func (s *InvitationService) RevokeInvite(ctx context.Context, inviteID, actorID uuid.UUID) error {
	if err := s.invitationRepo.Revoke(ctx, inviteID); err != nil {
		return err
	}

	s.auditRepo.Log(ctx, models.AuditLog{
		UserID:   actorID,
		Action:   "invite.revoked",
		Resource: "invitation",
		Data: map[string]interface{}{
			"invite_id": inviteID.String(),
		},
	})

	s.log.Info("invite_revoked", zap.String("invite_id", inviteID.String()))
	return nil
}

// ValidateInviteToken checks if an invitation token is valid and returns invite details.
func (s *InvitationService) ValidateInviteToken(ctx context.Context, token string) (*models.Invitation, error) {
	inv, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInviteNotFound
	}

	if inv.Status != models.InvitationPending {
		return nil, fmt.Errorf("invite status is %s: %w", inv.Status, ErrInviteAlreadyUsed)
	}

	if time.Now().After(inv.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	return inv, nil
}
