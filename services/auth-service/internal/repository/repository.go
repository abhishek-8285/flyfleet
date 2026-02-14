package repository

import (
	"context"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type AuthRepository interface {
	// User operations
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateTheme(ctx context.Context, userID uuid.UUID, theme string) error

	// Organization operations
	GetOrganizationsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	CreateOrganization(ctx context.Context, org *models.Organization) error
	UpdateOrganization(ctx context.Context, org *models.Organization) error
	AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) error

	// Token operations
	SaveToken(ctx context.Context, token *models.AuthToken) error
	GetTokenByRefresh(ctx context.Context, refreshToken string) (*models.AuthToken, error)
	DeleteToken(ctx context.Context, accessToken string) error
	DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error
	GetTokensByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthToken, error)

	// MFA operations
	SaveMFACode(ctx context.Context, mfa *models.MFACode) error
	GetMFACode(ctx context.Context, userID uuid.UUID, code string) (*models.MFACode, error)
	DeleteMFACode(ctx context.Context, id uuid.UUID) error

	// Security & Audit operations
	LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	GetSecurityEvents(ctx context.Context, userID uuid.UUID) ([]models.SecurityEvent, error)
	GetAuditLogByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
	ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.AuditLog, int, error)
	GetAuditSettings(ctx context.Context, orgID uuid.UUID) (*models.AuditSettings, error)
	UpdateAuditSettings(ctx context.Context, settings *models.AuditSettings) error
	GetUserRoleInOrg(ctx context.Context, userID, orgID uuid.UUID) (string, error)
	IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error)
	ResetFailedLogins(ctx context.Context, userID uuid.UUID) error
	LockAccount(ctx context.Context, userID uuid.UUID, duration time.Duration) error

	// Admin operations
	ListOrganizations(ctx context.Context, page, size int) ([]models.Organization, int, error)
	ListUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.User, int, error)
	SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID) error

	// Operational - Vehicles
	ListVehicles(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Vehicle, int, error)
	CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error
	GetVehicleByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error)
	UpdateVehicle(ctx context.Context, vehicle *models.Vehicle) error
	DeleteVehicle(ctx context.Context, id uuid.UUID) error
	UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error

	// Operational - Devices
	ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error)
	CreateDevice(ctx context.Context, device *models.Device) error
	GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error)
	AssignDeviceToVehicle(ctx context.Context, deviceID, vehicleID uuid.UUID) error
	UnassignDevice(ctx context.Context, deviceID uuid.UUID) error
	SoftDeleteDevice(ctx context.Context, id uuid.UUID) error
	AssignDriverToVehicle(ctx context.Context, driverID, vehicleID uuid.UUID) error

	// Operational - Documents
	CreateDocument(ctx context.Context, doc *models.Document) error
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error)
	ListDocumentsByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error)
	UpdateDocument(ctx context.Context, doc *models.Document) error
	DeleteDocument(ctx context.Context, id uuid.UUID) error

	// Security Suite Extensions

	// Security Suite Extensions
	GetMFADashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.MFADashboardSummary, error)
	GetMFAAdoptionTrend(ctx context.Context, orgID uuid.UUID, days int) ([]map[string]any, error)
	GetMFAMethodDistribution(ctx context.Context, orgID uuid.UUID) (map[string]int, error)
	GetMFAEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.MFAEvent, int, error)
	GetTopFailureLocations(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error)
	GetTopFailureUsers(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error)
	GetLockedUsers(ctx context.Context, orgID uuid.UUID) ([]models.User, error)
	UnlockUser(ctx context.Context, userID uuid.UUID) error
	ResetMFA(ctx context.Context, userID uuid.UUID) error
	GetMFAPolicy(ctx context.Context, orgID uuid.UUID) (*models.MFAPolicy, error)
	UpdateMFAPolicy(ctx context.Context, policy *models.MFAPolicy, updaterID uuid.UUID, reason string) error
	AddMFAException(ctx context.Context, exc *models.MFAException) error
	GetMFAExceptionByID(ctx context.Context, id uuid.UUID) (*models.MFAException, error)
	GetMFAExceptions(ctx context.Context, orgID uuid.UUID) ([]models.MFAException, error)
	DeleteMFAException(ctx context.Context, id uuid.UUID) error
	GetMFAPolicyHistory(ctx context.Context, orgID uuid.UUID) ([]models.MFAPolicyHistory, error)
	RollbackMFAPolicy(ctx context.Context, historyID uuid.UUID) error
	GetMFAOrgOverride(ctx context.Context, orgID uuid.UUID) (*models.MFAOrgOverride, error)
	UpdateMFAOrgOverride(ctx context.Context, override *models.MFAOrgOverride) error
	
	// Internal - Log Events
	LogMFAEvent(ctx context.Context, event *models.MFAEvent) error
	LogMFAFailure(ctx context.Context, failure *models.MFAFailure) error
	CreateMFALockout(ctx context.Context, lockout *models.MFALockout) error

	// IP Rules
	ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error)
	GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error)
	CreateIPRule(ctx context.Context, rule *models.IPRule) error
	DeleteIPRule(ctx context.Context, id uuid.UUID) error

	// Access Reviews
	ListAccessReviews(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AccessReview, int, error)
	GetAccessReviewByID(ctx context.Context, id uuid.UUID) (*models.AccessReview, error)
	CreateAccessReview(ctx context.Context, review *models.AccessReview) error
	UpdateAccessReviewStatus(ctx context.Context, id uuid.UUID, status string) error
	ListReviewItems(ctx context.Context, reviewID uuid.UUID) ([]models.ReviewItem, error)
	UpdateReviewItem(ctx context.Context, itemID uuid.UUID, status string, reviewerID uuid.UUID, notes string) error

	// Sessions - Extended for Admin management
	ListAllSessions(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AuthToken, int, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessionsForOrg(ctx context.Context, orgID uuid.UUID) error

	// RBAC operations
	GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	CreateRole(ctx context.Context, role *models.Role) error
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListRolesByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Role, error)
	
	ListPermissions(ctx context.Context) ([]models.PermissionDefinition, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.PermissionDefinition, error)
	UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	
	// Production Readiness - RBAC Extensions
	CloneRole(ctx context.Context, sourceRoleID uuid.UUID, newRoleName, newRoleDescription string) (*models.Role, error)
	GetEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]models.PermissionDefinition, error)
	CountUsersWithRole(ctx context.Context, roleName string, orgID *uuid.UUID) (int, error)
	GetRoleByName(ctx context.Context, orgID uuid.UUID, name string) (*models.Role, error)
	CreatePermission(ctx context.Context, perm *models.PermissionDefinition) error
	GetPermission(ctx context.Context, module, action string) (*models.PermissionDefinition, error)

	// Security Production Expansion
	LogLoginEvent(ctx context.Context, event *models.LoginEvent) error
	ListLoginEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.LoginEvent, int, error)
	GetLoginEventByID(ctx context.Context, id uuid.UUID) (*models.LoginEvent, error)
	
	CreateThreatAlert(ctx context.Context, alert *models.ThreatAlert) error
	ListThreatAlerts(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.ThreatAlert, int, error)
	GetThreatAlertByID(ctx context.Context, id uuid.UUID) (*models.ThreatAlert, error)
	UpdateThreatAlertStatus(ctx context.Context, id uuid.UUID, status string) error
	
	ListMFAUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.MFAUser, int, error)
	GetMFAUser(ctx context.Context, userID uuid.UUID) (*models.MFAUser, error)
	UpdateMFAUserStatus(ctx context.Context, userID uuid.UUID, status string) error

	// Audit & Compliance Expansion
	ListEntityAuditHistory(ctx context.Context, entityType, entityID string) ([]models.EntityAuditHistory, error)
	CreateEvidenceExport(ctx context.Context, export *models.EvidenceExport) error
	ListEvidenceExports(ctx context.Context, orgID uuid.UUID) ([]models.EvidenceExport, error)
	GetEvidenceExportByID(ctx context.Context, id uuid.UUID) (*models.EvidenceExport, error)
	UpdateEvidenceExportStatus(ctx context.Context, id uuid.UUID, status, fileURL, errorMsg string) error
	GetAuditPolicy(ctx context.Context, orgID uuid.UUID) (*models.AuditPolicy, error)
	UpdateAuditPolicy(ctx context.Context, policy *models.AuditPolicy) error

	// Stats
	GetSystemStats(ctx context.Context) (*models.SystemStats, error)
	GetSecurityTelemetry(ctx context.Context) (map[string]any, error)
}

type postgresRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) AuthRepository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}