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
	IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error)
	ResetFailedLogins(ctx context.Context, userID uuid.UUID) error
	LockAccount(ctx context.Context, userID uuid.UUID, duration time.Duration) error

	// Admin operations
	ListOrganizations(ctx context.Context, page, size int) ([]models.Organization, int, error)
	ListUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.User, int, error)
	SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID) error

	// Operational - Devices
	ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error)
	CreateDevice(ctx context.Context, device *models.Device) error
	SoftDeleteDevice(ctx context.Context, id uuid.UUID) error

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
	UpdateMFAPolicy(ctx context.Context, policy *models.MFAPolicy) error
	AddMFAException(ctx context.Context, exc *models.MFAException) error
	GetMFAExceptions(ctx context.Context, orgID uuid.UUID) ([]models.MFAException, error)
	DeleteMFAException(ctx context.Context, id uuid.UUID) error
	
	// Internal - Log Events
	LogMFAEvent(ctx context.Context, event *models.MFAEvent) error
	LogMFAFailure(ctx context.Context, failure *models.MFAFailure) error
	CreateMFALockout(ctx context.Context, lockout *models.MFALockout) error
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