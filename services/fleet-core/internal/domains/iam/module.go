package iam

import (
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"github.com/flyfleet/fleet-core/internal/domains/iam/email"
	"github.com/flyfleet/fleet-core/internal/domains/iam/handlers"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	redis2 "github.com/flyfleet/fleet-core/internal/domains/iam/redis"
	fcmiddleware "github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/permissions"
)

// Module wires together the IAM domain: repositories, services, and handlers.
// It exposes RegisterPublicRoutes and RegisterAuthRoutes to attach all 74 auth endpoints to a chi router.
type Module struct {
	// Handlers (exported for cross-domain route wiring)
	AuthHandler    *handlers.AuthHandler
	ProfileHandler *handlers.ProfileHandler
	SessionHandler *handlers.SessionHandler
	UserHandler    *handlers.UserHandler
	InviteHandler  *handlers.InviteHandler
	RBACHandler    *handlers.RBACHandler
	MFAHandler     *handlers.MFAHandler

	// Services (exposed for cross-domain calls)
	AuthService     *service.AuthService
	SessionService  *service.SessionService
	RBACService     *service.RBACService
	UserService     *service.UserService
	MFAService      *service.MFAService
	PasswordService *service.PasswordService
	InviteService   *service.InvitationService

	// Repositories (exposed for cross-domain calls)
	UserRepo    *repository.UserRepository
	RoleRepo    *repository.RoleRepository
	TenantRepo  *repository.TenantRepository
	SessionRepo *repository.SessionRepository
}

// NewModule creates all IAM repos, services, and handlers.
func NewModule(pool *pgxpool.Pool, rdb *redis.Client, jwtSecret string, iamCfg config.IAMConfig, logger *zap.Logger) *Module {
	// 1. Create repositories
	userRepo := repository.NewUserRepository(pool)
	tenantRepo := repository.NewTenantRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	passwordRepo := repository.NewPasswordHistoryRepository(pool)
	auditRepo := repository.NewAuditRepository(pool)
	invitationRepo := repository.NewInvitationRepository(pool)

	// 2. Create email service
	emailSvc := email.NewEmailService(iamCfg.Email, logger)

	// 3. Create JWT manager
	jwtManager := &jwtManagerImpl{secret: []byte(jwtSecret)}

	// 4. Create services
	secCfg := iamCfg.SecurityConfig()
	passwordSvc := service.NewPasswordService(userRepo, passwordRepo, sessionRepo, secCfg, logger)
	iamRedis := redis2.NewClient(rdb) // Wrap raw redis in IAM-specific client
	authSvc := service.NewAuthService(userRepo, tenantRepo, roleRepo, sessionRepo, passwordRepo, iamRedis, jwtManager, secCfg, logger, emailSvc, passwordSvc)
	sessionSvc := service.NewSessionService(sessionRepo, iamRedis, logger)
	rbacSvc := service.NewRBACService(roleRepo, logger)
	userSvc := service.NewUserService(userRepo, tenantRepo, roleRepo, auditRepo, logger)
	mfaSvc := service.NewMFAService(userRepo, iamRedis, logger)
	inviteSvc := service.NewInvitationService(invitationRepo, userRepo, tenantRepo, roleRepo, auditRepo, emailSvc, logger)

	// 5. Create handlers
	authHandler := handlers.NewAuthHandler(authSvc, passwordSvc, logger)
	profileHandler := handlers.NewProfileHandler(userRepo, logger)
	sessionHandler := handlers.NewSessionHandler(sessionSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)
	inviteHandler := handlers.NewInviteHandler(inviteSvc, logger)
	rbacHandler := handlers.NewRBACHandler(rbacSvc, logger)
	mfaHandler := handlers.NewMFAHandler(mfaSvc, logger)

	return &Module{
		AuthHandler:    authHandler,
		ProfileHandler: profileHandler,
		SessionHandler: sessionHandler,
		UserHandler:    userHandler,
		InviteHandler:  inviteHandler,
		RBACHandler:    rbacHandler,
		MFAHandler:     mfaHandler,
		AuthService:     authSvc,
		SessionService:  sessionSvc,
		RBACService:     rbacSvc,
		UserService:     userSvc,
		MFAService:      mfaSvc,
		PasswordService: passwordSvc,
		InviteService:   inviteSvc,
		UserRepo:        userRepo,
		RoleRepo:        roleRepo,
		TenantRepo:      tenantRepo,
		SessionRepo:     sessionRepo,
	}
}

// RegisterPublicRoutes attaches all public (no-auth) IAM endpoints to the given chi router.
// These endpoints do not require JWT middleware — the caller should NOT apply auth middleware.
// This method expects to be called within an r.Route("/api/v1", ...) block.
func (m *Module) RegisterPublicRoutes(r chi.Router, logger *zap.Logger) {
	r.Post("/auth/identify", m.AuthHandler.Identify)
	r.Post("/auth/login", m.AuthHandler.Login)
	r.Post("/auth/forgot-password", m.AuthHandler.ForgotPassword)
	r.Post("/auth/register", m.AuthHandler.Register)
	r.Post("/auth/mfa/verify", m.AuthHandler.VerifyMFA)
	r.Post("/auth/refresh", m.AuthHandler.Refresh)
	r.Post("/auth/reset-password", m.AuthHandler.ResetPassword)
	r.Post("/auth/verify-email", m.AuthHandler.VerifyEmail)
	r.Get("/auth/invite/{token}", m.InviteHandler.ValidateInviteToken)
	r.Post("/auth/invite/accept", m.InviteHandler.AcceptInvite)

	// Frontend compatibility aliases
	r.Post("/auth/resolve-tenants", m.AuthHandler.Identify)
}

// RegisterAuthRoutes attaches all authenticated IAM endpoints to the given chi router.
// The caller MUST apply JWT middleware before calling this method (e.g., via r.Use or r.Group).
// RBAC permission groups use fcmiddleware.RequirePermissions with fleet-core permission constants.
// This method expects to be called within an r.Route("/api/v1", ...) block.
func (m *Module) RegisterAuthRoutes(r chi.Router, logger *zap.Logger) {
	// Auth
	r.Post("/auth/logout", m.AuthHandler.Logout)
	r.Post("/auth/change-password", m.AuthHandler.ChangePassword)
	r.Get("/auth/password-policy", m.AuthHandler.GetPolicy)
	r.Post("/auth/tenants/{tenantId}/switch", m.AuthHandler.SwitchTenant)
	r.Get("/auth/tenants", m.AuthHandler.ListTenants)
	r.Post("/auth/reset-password-forced", m.AuthHandler.ResetPasswordForced)

	// Profile
	r.Get("/auth/profile", m.ProfileHandler.GetProfile)
	r.Put("/auth/profile", m.ProfileHandler.UpdateProfile)

	// Sessions
	r.Get("/auth/sessions", m.SessionHandler.ListSessions)
	r.Delete("/auth/sessions/{id}", m.SessionHandler.RevokeSession)
	r.Delete("/auth/sessions/all", m.SessionHandler.RevokeAllSessions)

	// MFA
	r.Post("/auth/mfa/setup/totp", m.MFAHandler.SetupTOTP)
	r.Post("/auth/mfa/setup/totp/verify", m.MFAHandler.VerifyTOTP)
	r.Post("/auth/mfa/disable", m.MFAHandler.DisableMFA)
	r.Get("/auth/mfa/backup-codes", m.MFAHandler.GetBackupCodes)

	// Admin - User management (requires user:read)
	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermReadUsers))
		r.Get("/admin/users", m.UserHandler.ListUsers)
		r.Get("/admin/users/{id}", m.UserHandler.GetUser)
		r.Get("/admin/users/{id}/audit-log", m.UserHandler.GetUserAuditLog)
	})

	// Admin - User status update (requires user:update)
	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermUpdateUsers))
		r.Put("/admin/users/{id}/status", m.UserHandler.UpdateStatus)
		r.Put("/admin/users/{id}", m.UserHandler.UpdatePreferences)
	})

	// Admin - User delete (requires user:delete)
	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermDeleteUsers))
		r.Delete("/admin/users/{id}", m.UserHandler.DeleteUser)
	})

	// Admin - Invitations (requires user:create)
	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermCreateUsers))
		r.Post("/admin/users/invite", m.InviteHandler.CreateInvite)
		r.Post("/admin/users/create", m.UserHandler.CreateUser)
		r.Post("/admin/users/invite/{id}/resend", m.InviteHandler.ResendInvite)
		r.Delete("/admin/users/invite/{id}", m.InviteHandler.RevokeInvite)
	})

	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermReadRoles))
		r.Get("/admin/rbac/roles", m.RBACHandler.ListRoles)

		r.Get("/admin/rbac/matrix", m.RBACHandler.GetPermissionMatrix)
		r.Get("/admin/rbac/permissions", m.RBACHandler.ListAllPermissions)
		r.Get("/admin/rbac/users/{id}/roles", m.RBACHandler.GetUserRoles)
		r.Get("/admin/rbac/users/{id}/permissions", m.RBACHandler.GetUserPermissions)
	})

	// Admin - RBAC writes (requires roles:update)
	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermUpdateRoles))
		r.Post("/admin/rbac/roles", m.RBACHandler.CreateRole)
		r.Put("/admin/rbac/roles/{id}", m.RBACHandler.UpdateRole)
		r.Post("/admin/rbac/roles/{id}/clone", m.RBACHandler.CloneRole)
		r.Put("/admin/rbac/roles/{id}/permissions", m.RBACHandler.UpdateRolePermissions)
		r.Post("/admin/rbac/users/assign", m.RBACHandler.AssignRole)
		r.Post("/admin/rbac/users/assign/bulk", m.RBACHandler.BulkAssignRole)
		r.Delete("/admin/rbac/users/unassign", m.RBACHandler.UnassignRole)
	})

	// Admin - RBAC deletes (requires roles:delete)
	r.Group(func(r chi.Router) {
		r.Use(fcmiddleware.RequirePermissions(logger, permissions.PermDeleteRoles))
		r.Delete("/admin/rbac/roles/{id}", m.RBACHandler.DeleteRole)
	})
}

// ── JWT Manager Implementation ──

// jwtManagerImpl implements service.JWTManager using HMAC-SHA256.
type jwtManagerImpl struct {
	secret []byte
}

func (m *jwtManagerImpl) GenerateToken(claims jwt.MapClaims, expiration time.Duration) (string, error) {
	now := time.Now()
	claims["exp"] = now.Add(expiration).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *jwtManagerImpl) VerifyToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
}

func (m *jwtManagerImpl) ExtractClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := m.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// Compile-time interface check
var _ service.JWTManager = (*jwtManagerImpl)(nil)
