package handlers

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	"github.com/google/uuid"
)

// ─── Service interfaces (for mocking in tests) ───
// Real services satisfy these interfaces implicitly.

type authService interface {
	Identify(ctx context.Context, email string) (*models.IdentifyResponse, error)
	Login(ctx context.Context, req models.LoginRequest, ip, userAgent string) (*models.AuthTokensResponse, *models.MFARequiredResponse, error)
	VerifyMFA(ctx context.Context, req models.MFAVerifyRequest, ip, userAgent string) (*models.AuthTokensResponse, error)
	Refresh(ctx context.Context, refreshToken, ip, ua string) (*models.AuthTokensResponse, error)
	Register(ctx context.Context, email, password, name string, orgID uuid.UUID) (*models.AuthTokensResponse, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPasswordWithToken(ctx context.Context, token, newPassword string) error
	Logout(ctx context.Context, userID uuid.UUID, tokenJTI string) error
	SwitchOrg(ctx context.Context, userID uuid.UUID, targetOrgID uuid.UUID, ip, ua string) (*models.OrgSwitchResponse, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.OrgInfo, error)
	VerifyEmail(ctx context.Context, token string) error
}

type passwordService interface {
	Validate(ctx context.Context, userID uuid.UUID, password string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	GetPolicy() models.PasswordPolicyResponse
}

type sessionService interface {
	ListSessions(ctx context.Context, userID, orgID uuid.UUID) ([]models.Session, error)
	RevokeSession(ctx context.Context, userID, orgID uuid.UUID, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
}

type rbacService interface {
	ListRoles(ctx context.Context, orgID uuid.UUID) ([]models.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)
	CreateRole(ctx context.Context, orgID uuid.UUID, name, description string, perms []string, createdBy uuid.UUID) (*models.Role, error)
	UpdateRole(ctx context.Context, id, orgID uuid.UUID, name, description string) error
	UpdateRolePermissions(ctx context.Context, id, orgID uuid.UUID, perms []string) error
	CloneRole(ctx context.Context, sourceID, orgID uuid.UUID, name, description string, createdBy uuid.UUID) (*models.Role, error)
	DeleteRole(ctx context.Context, id, orgID uuid.UUID) error
	GetPermissionMatrix(ctx context.Context, orgID uuid.UUID) ([]models.RolePermissionMatrixEntry, error)
	AssignRole(ctx context.Context, userID, roleID, orgID, assignedBy uuid.UUID) error
	BulkAssignRoles(ctx context.Context, userIDs, roleIDs []uuid.UUID, orgID, assignedBy uuid.UUID) *service.BulkAssignResult
	UnassignRole(ctx context.Context, userID, roleID, orgID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID, orgID uuid.UUID) ([]models.Role, error)
	GetUserPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
	ListAllPermissions(ctx context.Context) ([]models.Permission, error)
}

type userService interface {
	ListUsers(ctx context.Context, params models.ListUsersParams) (*models.ListUsersResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string, actorID uuid.UUID) error
	DeleteUser(ctx context.Context, userID uuid.UUID, actorID uuid.UUID) error
	GetUserAuditLog(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error)
}

type invitationService interface {
	CreateInvite(ctx context.Context, orgID, invitedBy uuid.UUID, email string, roleIDs []uuid.UUID, message string) (*models.Invitation, error)
	AcceptInvite(ctx context.Context, token, name, password string) (*models.User, error)
	ResendInvite(ctx context.Context, inviteID, actorID uuid.UUID) error
	RevokeInvite(ctx context.Context, inviteID, actorID uuid.UUID) error
	ValidateInviteToken(ctx context.Context, token string) (*models.Invitation, error)
}

type mfaService interface {
	SetupTOTP(ctx context.Context, userID uuid.UUID) (string, string, error)
	VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) error
	GenerateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
	DisableMFA(ctx context.Context, userID uuid.UUID, password string) error
}

type userRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name, phone, avatarURL string) error
	GetOrganizations(ctx context.Context, id uuid.UUID) ([]models.OrgInfo, error)
}

// Compile-time: real services satisfy interfaces
var _ authService = (*service.AuthService)(nil)
var _ passwordService = (*service.PasswordService)(nil)
var _ sessionService = (*service.SessionService)(nil)
var _ rbacService = (*service.RBACService)(nil)
var _ userService = (*service.UserService)(nil)
var _ invitationService = (*service.InvitationService)(nil)
var _ mfaService = (*service.MFAService)(nil)
