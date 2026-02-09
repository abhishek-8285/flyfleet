package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockRepository implements AuthRepository for testing
type MockRepository struct {
	roles       map[uuid.UUID]*models.Role
	permissions map[uuid.UUID]*models.PermissionDefinition
	// Add other necessary fields
}

// Implement only necessary methods for RBAC testing
func (m *MockRepository) ListPermissions(ctx context.Context) ([]models.PermissionDefinition, error) {
	return nil, nil 
}
func (m *MockRepository) CreatePermission(ctx context.Context, perm *models.PermissionDefinition) error {
	m.permissions[perm.ID] = perm
	return nil
}
func (m *MockRepository) GetPermission(ctx context.Context, module, action string) (*models.PermissionDefinition, error) {
	return nil, nil // Mock: not found
}

func (m *MockRepository) GetRoleByName(ctx context.Context, orgID uuid.UUID, name string) (*models.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			if r.OrganizationID != nil && *r.OrganizationID == orgID {
				return r, nil
			}
			if r.OrganizationID == nil && orgID == uuid.Nil {
				return r, nil
			}
		}
	}
	return nil, nil
}

func (m *MockRepository) CreateRole(ctx context.Context, role *models.Role) error {
	m.roles[role.ID] = role
	return nil
}

func (m *MockRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	if role, ok := m.roles[id]; ok {
		return role, nil
	}
	return nil, errors.New("role not found")
}

func (m *MockRepository) UpdateRole(ctx context.Context, role *models.Role) error {
	m.roles[role.ID] = role
	return nil
}

func (m *MockRepository) UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return nil
}

func (m *MockRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// Stub other methods to satisfy interface
func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) { return nil, nil }
func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error { return nil }
func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) { return nil, nil }
func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error { return nil }
func (m *MockRepository) UpdateTheme(ctx context.Context, userID uuid.UUID, theme string) error { return nil }
func (m *MockRepository) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockRepository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error) { return 0, nil }
func (m *MockRepository) LockAccount(ctx context.Context, userID uuid.UUID, duration time.Duration) error { return nil }
func (m *MockRepository) GetOrganizationsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) { return nil, nil }
func (m *MockRepository) GetUserRoleInOrg(ctx context.Context, userID, orgID uuid.UUID) (string, error) { return "", nil }
func (m *MockRepository) SaveToken(ctx context.Context, token *models.AuthToken) error { return nil }
func (m *MockRepository) GetTokenByRefresh(ctx context.Context, refreshToken string) (*models.AuthToken, error) { return nil, nil }
func (m *MockRepository) GetTokensByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthToken, error) { return nil, nil }
func (m *MockRepository) DeleteToken(ctx context.Context, accessToken string) error { return nil }
func (m *MockRepository) DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockRepository) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error { return nil }
func (m *MockRepository) GetSecurityEvents(ctx context.Context, userID uuid.UUID) ([]models.SecurityEvent, error) { return nil, nil }
func (m *MockRepository) SaveMFACode(ctx context.Context, code *models.MFACode) error { return nil }
func (m *MockRepository) GetMFACode(ctx context.Context, userID uuid.UUID, code string) (*models.MFACode, error) { return nil, nil }
func (m *MockRepository) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.AuditLog, int, error) { return nil, 0, nil }
func (m *MockRepository) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) { return nil, nil }
func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error { return nil }
func (m *MockRepository) GetAuditSettings(ctx context.Context, orgID uuid.UUID) (*models.AuditSettings, error) { return nil, nil }
func (m *MockRepository) UpdateAuditSettings(ctx context.Context, settings *models.AuditSettings) error { return nil }
func (m *MockRepository) CreateOrganization(ctx context.Context, org *models.Organization) error { return nil }
func (m *MockRepository) UpdateOrganization(ctx context.Context, org *models.Organization) error { return nil }
func (m *MockRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) { return nil, nil }
func (m *MockRepository) ListOrganizations(ctx context.Context, page, size int) ([]models.Organization, int, error) { return nil, 0, nil }
func (m *MockRepository) ListUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.User, int, error) { return nil, 0, nil }
func (m *MockRepository) SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) SoftDeleteUser(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) ListVehicles(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Vehicle, int, error) { return nil, 0, nil }
func (m *MockRepository) CreateVehicle(ctx context.Context, v *models.Vehicle) error { return nil }
func (m *MockRepository) GetVehicleByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) { return nil, nil }
func (m *MockRepository) UpdateVehicle(ctx context.Context, v *models.Vehicle) error { return nil }
func (m *MockRepository) UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
func (m *MockRepository) DeleteVehicle(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error) { return nil, 0, nil }
func (m *MockRepository) CreateDevice(ctx context.Context, d *models.Device) error { return nil }
func (m *MockRepository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error) { return nil, nil }
func (m *MockRepository) AssignDeviceToVehicle(ctx context.Context, deviceID, vehicleID uuid.UUID) error { return nil }
func (m *MockRepository) UnassignDevice(ctx context.Context, deviceID uuid.UUID) error { return nil }
func (m *MockRepository) SoftDeleteDevice(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) AssignDriverToVehicle(ctx context.Context, driverID, vehicleID uuid.UUID) error { return nil }
func (m *MockRepository) CreateDocument(ctx context.Context, d *models.Document) error { return nil }
func (m *MockRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error) { return nil, nil }
func (m *MockRepository) ListDocumentsByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error) { return nil, nil }
func (m *MockRepository) UpdateDocument(ctx context.Context, d *models.Document) error { return nil }
func (m *MockRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) error { return nil }
func (m *MockRepository) DeleteMFACode(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) GetMFADashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.MFADashboardSummary, error) { return nil, nil }
func (m *MockRepository) GetMFAAdoptionTrend(ctx context.Context, orgID uuid.UUID, days int) ([]map[string]any, error) { return nil, nil }
func (m *MockRepository) GetMFAMethodDistribution(ctx context.Context, orgID uuid.UUID) (map[string]int, error) { return nil, nil }
func (m *MockRepository) GetMFAEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.MFAEvent, int, error) { return nil, 0, nil }
func (m *MockRepository) GetTopFailureLocations(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) { return nil, nil }
func (m *MockRepository) GetTopFailureUsers(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) { return nil, nil }
func (m *MockRepository) GetLockedUsers(ctx context.Context, orgID uuid.UUID) ([]models.User, error) { return nil, nil }
func (m *MockRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockRepository) ResetMFA(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockRepository) GetMFAPolicy(ctx context.Context, orgID uuid.UUID) (*models.MFAPolicy, error) { return nil, nil }
func (m *MockRepository) UpdateMFAPolicy(ctx context.Context, p *models.MFAPolicy, uid uuid.UUID, r string) error { return nil }
func (m *MockRepository) GetMFAPolicyHistory(ctx context.Context, orgID uuid.UUID) ([]models.MFAPolicyHistory, error) { return nil, nil }
func (m *MockRepository) RollbackMFAPolicy(ctx context.Context, historyID uuid.UUID) error { return nil }
func (m *MockRepository) GetMFAOrgOverride(ctx context.Context, orgID uuid.UUID) (*models.MFAOrgOverride, error) { return nil, nil }
func (m *MockRepository) UpdateMFAOrgOverride(ctx context.Context, o *models.MFAOrgOverride) error { return nil }
func (m *MockRepository) AddMFAException(ctx context.Context, exc *models.MFAException) error { return nil }
func (m *MockRepository) GetMFAExceptionByID(ctx context.Context, id uuid.UUID) (*models.MFAException, error) { return nil, nil }
func (m *MockRepository) GetMFAExceptions(ctx context.Context, orgID uuid.UUID) ([]models.MFAException, error) { return nil, nil }
func (m *MockRepository) DeleteMFAException(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockRepository) LogMFAEvent(ctx context.Context, e *models.MFAEvent) error { return nil }
func (m *MockRepository) LogMFAFailure(ctx context.Context, f *models.MFAFailure) error { return nil }
func (m *MockRepository) CreateMFALockout(ctx context.Context, l *models.MFALockout) error { return nil }
func (m *MockRepository) ListRolesByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Role, error) { return nil, nil }
func (m *MockRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.PermissionDefinition, error) { return nil, nil }
func (m *MockRepository) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }
func (m *MockRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error { return nil }
func (m *MockRepository) CloneRole(ctx context.Context, sourceRoleID uuid.UUID, newRoleName, newRoleDescription string) (*models.Role, error) { return nil, nil }
func (m *MockRepository) GetEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]models.PermissionDefinition, error) { return nil, nil }
func (m *MockRepository) CountUsersWithRole(ctx context.Context, roleName string, orgID *uuid.UUID) (int, error) { return 0, nil }
func (m *MockRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) { return nil, nil }

// IP Rules
func (m *MockRepository) ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error) { return nil, nil }
func (m *MockRepository) GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error) { return nil, nil }
func (m *MockRepository) CreateIPRule(ctx context.Context, rule *models.IPRule) error { return nil }
func (m *MockRepository) DeleteIPRule(ctx context.Context, id uuid.UUID) error { return nil }

// Access Reviews
func (m *MockRepository) ListAccessReviews(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AccessReview, int, error) { return nil, 0, nil }
func (m *MockRepository) GetAccessReviewByID(ctx context.Context, id uuid.UUID) (*models.AccessReview, error) { return nil, nil }
func (m *MockRepository) CreateAccessReview(ctx context.Context, review *models.AccessReview) error { return nil }
func (m *MockRepository) UpdateAccessReviewStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
func (m *MockRepository) ListReviewItems(ctx context.Context, reviewID uuid.UUID) ([]models.ReviewItem, error) { return nil, nil }
func (m *MockRepository) UpdateReviewItem(ctx context.Context, itemID uuid.UUID, status string, reviewerID uuid.UUID, notes string) error { return nil }

// Sessions - Extended
func (m *MockRepository) ListAllSessions(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AuthToken, int, error) { return nil, 0, nil }
func (m *MockRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error { return nil }
func (m *MockRepository) RevokeAllSessionsForOrg(ctx context.Context, orgID uuid.UUID) error { return nil }


func TestRBAC_UniqueRoleName(t *testing.T) {
	repo := &MockRepository{
		roles: make(map[uuid.UUID]*models.Role),
	}
	logger := zap.NewNop()
	svc := NewAuthService(repo, "", "", "", "", "", logger)

	orgID := uuid.New()

	// 1. Create Role
	_, err := svc.CreateRole(context.Background(), "Manager", "Role Desc", orgID, false, uuid.Nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// 2. Try Duplicate
	_, err = svc.CreateRole(context.Background(), "Manager", "Another Desc", orgID, false, uuid.Nil)
	if err == nil {
		t.Error("Expected error for duplicate role name, got nil")
	} else if err.Error() != "role with this name already exists" {
		t.Errorf("Expected 'role with this name already exists', got '%v'", err)
	}

	// 3. Update to duplicate
	role2, _ := svc.CreateRole(context.Background(), "Driver", "Role Desc", orgID, false, uuid.Nil)
	_, err = svc.UpdateRole(context.Background(), role2.ID, "Manager", "Updated Desc", uuid.Nil)
	if err == nil {
		t.Error("Expected error for updating to duplicate name, got nil")
	}
}

func TestRBAC_SystemRoleLockdown(t *testing.T) {
	repo := &MockRepository{
		roles: make(map[uuid.UUID]*models.Role),
	}
	logger := zap.NewNop()
	svc := NewAuthService(repo, "", "", "", "", "", logger)

	// Create System Role via Repo (bypassing service check if any)
	sysRole := &models.Role{
		ID:             uuid.New(),
		Name:           "SystemAdmin",
		IsSystem:       true,
		OrganizationID: nil,
	}
	repo.CreateRole(context.Background(), sysRole)

	// 1. Try Update Name
	_, err := svc.UpdateRole(context.Background(), sysRole.ID, "HackAdmin", "desc", uuid.Nil)
	if err == nil {
		t.Error("Expected error when updating system role, got nil")
	}

	// 2. Try Update Permissions
	err = svc.UpdateRolePermissions(context.Background(), sysRole.ID, []uuid.UUID{uuid.New()}, uuid.Nil)
	if err == nil {
		t.Error("Expected error when updating system role permissions, got nil")
	} else if err.Error() != "cannot modify permissions of system roles" {
		t.Errorf("Expected 'cannot modify permissions of system roles', got '%v'", err)
	}
}
