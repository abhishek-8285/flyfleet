package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Define a proper mock for AuthRepository since MockRepository is a struct not using testify mock
type RepositoryMock struct {
	mock.Mock
}

// Implement only used methods
func (m *RepositoryMock) ListUsers(ctx context.Context, filter models.UserFilter) ([]models.User, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.User), args.Int(1), args.Error(2)
}

// Stub other methods from interface
func (m *RepositoryMock) GetUserByEmail(ctx context.Context, email string) (*models.User, error) { return nil, nil }
func (m *RepositoryMock) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) { return nil, nil }
func (m *RepositoryMock) CreateUser(ctx context.Context, user *models.User) error { return nil }
func (m *RepositoryMock) UpdateUser(ctx context.Context, user *models.User) error { return nil }
func (m *RepositoryMock) UpdateUserPassword(ctx context.Context, user *models.User) error { return nil }
func (m *RepositoryMock) UpdateUserVerified(ctx context.Context, userID uuid.UUID, verified bool) error { return nil }
func (m *RepositoryMock) UpdateTheme(ctx context.Context, userID uuid.UUID, theme string) error { return nil }
func (m *RepositoryMock) GetOrganizationsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) { return nil, nil }
func (m *RepositoryMock) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) { return nil, nil }
func (m *RepositoryMock) CreateOrganization(ctx context.Context, org *models.Organization) error { return nil }
func (m *RepositoryMock) UpdateOrganization(ctx context.Context, org *models.Organization) error { return nil }
func (m *RepositoryMock) AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) error { return nil }
func (m *RepositoryMock) SaveToken(ctx context.Context, token *models.AuthToken) error { return nil }
func (m *RepositoryMock) GetTokenByRefresh(ctx context.Context, refreshToken string) (*models.AuthToken, error) { return nil, nil }
func (m *RepositoryMock) DeleteToken(ctx context.Context, accessToken string) error { return nil }
func (m *RepositoryMock) DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *RepositoryMock) GetTokensByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthToken, error) { return nil, nil }
func (m *RepositoryMock) GetTokenByAccessToken(ctx context.Context, accessToken string) (*models.AuthToken, error) { return nil, nil }
func (m *RepositoryMock) UpdateTokenLastActive(ctx context.Context, accessToken string) error { return nil }
func (m *RepositoryMock) CleanupExpiredTokens(ctx context.Context) error { return nil }
func (m *RepositoryMock) SaveMFACode(ctx context.Context, mfa *models.MFACode) error { return nil }
func (m *RepositoryMock) GetMFACode(ctx context.Context, userID uuid.UUID, code string) (*models.MFACode, error) { return nil, nil }
func (m *RepositoryMock) DeleteMFACode(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error { return nil }
func (m *RepositoryMock) CreateAuditLog(ctx context.Context, log *models.AuditLog) error { return nil }
func (m *RepositoryMock) GetSecurityEvents(ctx context.Context, userID uuid.UUID) ([]models.SecurityEvent, error) { return nil, nil }
func (m *RepositoryMock) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) { return nil, nil }
func (m *RepositoryMock) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.AuditLog, int, error) { return nil, 0, nil }
func (m *RepositoryMock) GetAuditSettings(ctx context.Context, orgID uuid.UUID) (*models.AuditSettings, error) { return nil, nil }
func (m *RepositoryMock) UpdateAuditSettings(ctx context.Context, settings *models.AuditSettings) error { return nil }
func (m *RepositoryMock) GetUserRoleInOrg(ctx context.Context, userID, orgID uuid.UUID) (string, error) { return "", nil }
func (m *RepositoryMock) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error) { return 0, nil }
func (m *RepositoryMock) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *RepositoryMock) LockAccount(ctx context.Context, userID uuid.UUID, duration time.Duration) error { return nil }
func (m *RepositoryMock) ListOrganizations(ctx context.Context, page, size int) ([]models.Organization, int, error) { return nil, 0, nil }
func (m *RepositoryMock) SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) SoftDeleteUser(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) ListVehicles(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Vehicle, int, error) { return nil, 0, nil }
func (m *RepositoryMock) CreateVehicle(ctx context.Context, vehicle *models.Vehicle) error { return nil }
func (m *RepositoryMock) GetVehicleByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) { return nil, nil }
func (m *RepositoryMock) UpdateVehicle(ctx context.Context, vehicle *models.Vehicle) error { return nil }
func (m *RepositoryMock) DeleteVehicle(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
func (m *RepositoryMock) ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error) { return nil, 0, nil }
func (m *RepositoryMock) CreateDevice(ctx context.Context, device *models.Device) error { return nil }
func (m *RepositoryMock) GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error) { return nil, nil }
func (m *RepositoryMock) AssignDeviceToVehicle(ctx context.Context, deviceID, vehicleID uuid.UUID) error { return nil }
func (m *RepositoryMock) UnassignDevice(ctx context.Context, deviceID uuid.UUID) error { return nil }
func (m *RepositoryMock) SoftDeleteDevice(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) AssignDriverToVehicle(ctx context.Context, driverID, vehicleID uuid.UUID) error { return nil }
func (m *RepositoryMock) CreateDocument(ctx context.Context, doc *models.Document) error { return nil }
func (m *RepositoryMock) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error) { return nil, nil }
func (m *RepositoryMock) ListDocumentsByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error) { return nil, nil }
func (m *RepositoryMock) ListDocuments(ctx context.Context, orgID uuid.UUID, filter *models.DocumentFilter) ([]models.Document, int, error) { return nil, 0, nil }
func (m *RepositoryMock) UpdateDocument(ctx context.Context, doc *models.Document) error { return nil }
func (m *RepositoryMock) DeleteDocument(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) GetMFADashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.MFADashboardSummary, error) { return nil, nil }
func (m *RepositoryMock) GetMFAAdoptionTrend(ctx context.Context, orgID uuid.UUID, days int) ([]map[string]any, error) { return nil, nil }
func (m *RepositoryMock) GetMFAMethodDistribution(ctx context.Context, orgID uuid.UUID) (map[string]int, error) { return nil, nil }
func (m *RepositoryMock) GetMFAEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.MFAEvent, int, error) { return nil, 0, nil }
func (m *RepositoryMock) GetTopFailureLocations(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) { return nil, nil }
func (m *RepositoryMock) GetTopFailureUsers(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) { return nil, nil }
func (m *RepositoryMock) GetLockedUsers(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.User, int, error) { return nil, 0, nil }
func (m *RepositoryMock) UnlockUser(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *RepositoryMock) ResetMFA(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *RepositoryMock) GetMFAPolicy(ctx context.Context, orgID uuid.UUID) (*models.MFAPolicy, error) { return nil, nil }
func (m *RepositoryMock) UpdateMFAPolicy(ctx context.Context, policy *models.MFAPolicy, updaterID uuid.UUID, reason string) error { return nil }
func (m *RepositoryMock) AddMFAException(ctx context.Context, exc *models.MFAException) error { return nil }
func (m *RepositoryMock) GetMFAExceptionByID(ctx context.Context, id uuid.UUID) (*models.MFAException, error) { return nil, nil }
func (m *RepositoryMock) GetMFAExceptions(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.MFAException, int, error) { return nil, 0, nil }
func (m *RepositoryMock) DeleteMFAException(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) GetMFAPolicyHistory(ctx context.Context, orgID uuid.UUID) ([]models.MFAPolicyHistory, error) { return nil, nil }
func (m *RepositoryMock) RollbackMFAPolicy(ctx context.Context, historyID uuid.UUID) error { return nil }
func (m *RepositoryMock) GetMFAOrgOverride(ctx context.Context, orgID uuid.UUID) (*models.MFAOrgOverride, error) { return nil, nil }
func (m *RepositoryMock) UpdateMFAOrgOverride(ctx context.Context, override *models.MFAOrgOverride) error { return nil }
func (m *RepositoryMock) LogMFAEvent(ctx context.Context, event *models.MFAEvent) error { return nil }
func (m *RepositoryMock) LogMFAFailure(ctx context.Context, failure *models.MFAFailure) error { return nil }
func (m *RepositoryMock) CreateMFALockout(ctx context.Context, lockout *models.MFALockout) error { return nil }
func (m *RepositoryMock) ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error) { return nil, nil }
func (m *RepositoryMock) GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error) { return nil, nil }
func (m *RepositoryMock) CreateIPRule(ctx context.Context, rule *models.IPRule) error { return nil }
func (m *RepositoryMock) DeleteIPRule(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) ListAccessReviews(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AccessReview, int, error) { return nil, 0, nil }
func (m *RepositoryMock) GetAccessReviewByID(ctx context.Context, id uuid.UUID) (*models.AccessReview, error) { return nil, nil }
func (m *RepositoryMock) CreateAccessReview(ctx context.Context, review *models.AccessReview) error { return nil }
func (m *RepositoryMock) UpdateAccessReviewStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
func (m *RepositoryMock) ListReviewItems(ctx context.Context, reviewID uuid.UUID) ([]models.ReviewItem, error) { return nil, nil }
func (m *RepositoryMock) UpdateReviewItem(ctx context.Context, itemID uuid.UUID, status string, reviewerID uuid.UUID, notes string) error { return nil }
func (m *RepositoryMock) ListAllSessions(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AuthToken, int, error) { return nil, 0, nil }
func (m *RepositoryMock) RevokeSession(ctx context.Context, sessionID uuid.UUID) error { return nil }
func (m *RepositoryMock) RevokeAllSessionsForOrg(ctx context.Context, orgID uuid.UUID) error { return nil }
func (m *RepositoryMock) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) { return nil, nil }
func (m *RepositoryMock) CreateRole(ctx context.Context, role *models.Role) error { return nil }
func (m *RepositoryMock) UpdateRole(ctx context.Context, role *models.Role) error { return nil }
func (m *RepositoryMock) DeleteRole(ctx context.Context, id uuid.UUID) error { return nil }
func (m *RepositoryMock) ListRolesByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Role, error) { return nil, nil }
func (m *RepositoryMock) ListPermissions(ctx context.Context) ([]models.PermissionDefinition, error) { return nil, nil }
func (m *RepositoryMock) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.PermissionDefinition, error) { return nil, nil }
func (m *RepositoryMock) UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error { return nil }
func (m *RepositoryMock) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }
func (m *RepositoryMock) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error { return nil }
func (m *RepositoryMock) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) { return nil, nil }
func (m *RepositoryMock) CloneRole(ctx context.Context, sourceRoleID uuid.UUID, newRoleName, newRoleDescription string) (*models.Role, error) { return nil, nil }
func (m *RepositoryMock) GetEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]models.PermissionDefinition, error) { return nil, nil }
func (m *RepositoryMock) CountUsersWithRole(ctx context.Context, roleName string, orgID *uuid.UUID) (int, error) { return 0, nil }
func (m *RepositoryMock) GetRoleByName(ctx context.Context, orgID uuid.UUID, name string) (*models.Role, error) { return nil, nil }
func (m *RepositoryMock) CreatePermission(ctx context.Context, perm *models.PermissionDefinition) error { return nil }
func (m *RepositoryMock) GetPermission(ctx context.Context, module, action string) (*models.PermissionDefinition, error) { return nil, nil }
func (m *RepositoryMock) LogLoginEvent(ctx context.Context, event *models.LoginEvent) error { return nil }
func (m *RepositoryMock) ListLoginEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.LoginEvent, int, error) { return nil, 0, nil }
func (m *RepositoryMock) GetLoginEventByID(ctx context.Context, id uuid.UUID) (*models.LoginEvent, error) { return nil, nil }
func (m *RepositoryMock) CreateThreatAlert(ctx context.Context, alert *models.ThreatAlert) error { return nil }
func (m *RepositoryMock) ListThreatAlerts(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.ThreatAlert, int, error) { return nil, 0, nil }
func (m *RepositoryMock) GetThreatAlertByID(ctx context.Context, id uuid.UUID) (*models.ThreatAlert, error) { return nil, nil }
func (m *RepositoryMock) UpdateThreatAlertStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
func (m *RepositoryMock) ListMFAUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.MFAUser, int, error) { return nil, 0, nil }
func (m *RepositoryMock) GetMFAUser(ctx context.Context, userID uuid.UUID) (*models.MFAUser, error) { return nil, nil }
func (m *RepositoryMock) UpdateMFAUserStatus(ctx context.Context, userID uuid.UUID, status string) error { return nil }
func (m *RepositoryMock) ListEntityAuditHistory(ctx context.Context, entityType, entityID string) ([]models.EntityAuditHistory, error) { return nil, nil }
func (m *RepositoryMock) CreateEvidenceExport(ctx context.Context, export *models.EvidenceExport) error { return nil }
func (m *RepositoryMock) ListEvidenceExports(ctx context.Context, orgID uuid.UUID) ([]models.EvidenceExport, error) { return nil, nil }
func (m *RepositoryMock) GetEvidenceExportByID(ctx context.Context, id uuid.UUID) (*models.EvidenceExport, error) { return nil, nil }
func (m *RepositoryMock) UpdateEvidenceExportStatus(ctx context.Context, id uuid.UUID, status, fileURL, errorMsg string) error { return nil }
func (m *RepositoryMock) GetAuditPolicy(ctx context.Context, orgID uuid.UUID) (*models.AuditPolicy, error) { return nil, nil }
func (m *RepositoryMock) UpdateAuditPolicy(ctx context.Context, policy *models.AuditPolicy) error { return nil }
func (m *RepositoryMock) GetSystemStats(ctx context.Context) (*models.SystemStats, error) { return nil, nil }
func (m *RepositoryMock) GetSecurityTelemetry(ctx context.Context) (map[string]any, error) { return nil, nil }
func (m *RepositoryMock) GetSecuritySettings(ctx context.Context, orgID uuid.UUID) (*models.SecuritySettings, error) { return nil, nil }
func (m *RepositoryMock) UpdateSecuritySettings(ctx context.Context, settings *models.SecuritySettings) error { return nil }
func (m *RepositoryMock) GetSecurityMetrics(ctx context.Context, orgID uuid.UUID) (*models.SecurityMetrics, error) { return nil, nil }
func (m *RepositoryMock) GetVulnerabilityReport(ctx context.Context, orgID uuid.UUID) (*models.VulnerabilityReport, error) { return nil, nil }
func (m *RepositoryMock) GetComplianceReport(ctx context.Context, orgID uuid.UUID, reportType string, startDate, endDate time.Time) (*models.ComplianceReport, error) { return nil, nil }
func (m *RepositoryMock) ListSecurityEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.SecurityEvent, int, error) { return nil, 0, nil }


func TestListUsers(t *testing.T) {
	mockRepo := new(RepositoryMock)
	logger := zap.NewNop()
	svc := &AuthService{
		repo:   mockRepo,
		logger: logger,
	}

	orgID := uuid.New()
	filter := models.UserFilter{
		OrgID: orgID,
		Page:  1,
		Limit: 10,
	}

	expectedUsers := []models.User{
		{ID: uuid.New(), Email: "user1@example.com", OrgID: orgID},
		{ID: uuid.New(), Email: "user2@example.com", OrgID: orgID},
	}

	mockRepo.On("ListUsers", mock.Anything, filter).Return(expectedUsers, 15, nil)

	response, err := svc.ListUsers(context.Background(), filter)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response.Data))
	assert.Equal(t, 15, response.Meta.Total)
	assert.Equal(t, 2, response.Meta.TotalPages)
	assert.True(t, response.Meta.HasNext)
	assert.False(t, response.Meta.HasPrev)

	mockRepo.AssertExpectations(t)
}

func TestListUsers_PaginationCalculations(t *testing.T) {
	mockRepo := new(RepositoryMock)
	logger := zap.NewNop()
	svc := &AuthService{
		repo:   mockRepo,
		logger: logger,
	}

	orgID := uuid.New()
	
	testCases := []struct {
		name       string
		page       int
		limit      int
		total      int
		totalPages int
		hasNext    bool
		hasPrev    bool
	}{
		{"First Page", 1, 10, 25, 3, true, false},
		{"Middle Page", 2, 10, 25, 3, true, true},
		{"Last Page", 3, 10, 25, 3, false, true},
		{"Single Page", 1, 10, 5, 1, false, false},
		{"Empty Result", 1, 10, 0, 0, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter := models.UserFilter{OrgID: orgID, Page: tc.page, Limit: tc.limit}
			mockRepo.On("ListUsers", mock.Anything, filter).Return([]models.User{}, tc.total, nil).Once()

			resp, err := svc.ListUsers(context.Background(), filter)
			assert.NoError(t, err)
			assert.Equal(t, tc.total, resp.Meta.Total)
			assert.Equal(t, tc.totalPages, resp.Meta.TotalPages)
			assert.Equal(t, tc.hasNext, resp.Meta.HasNext)
			assert.Equal(t, tc.hasPrev, resp.Meta.HasPrev)
		})
	}
}
