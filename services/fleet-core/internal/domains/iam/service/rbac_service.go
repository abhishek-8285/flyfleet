package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrRoleNotFound   = errors.New("role_not_found")
	ErrRoleIsSystem   = errors.New("role_is_system")
	ErrPermissionNotFound = errors.New("permission_not_found")
)

// permissionLookup maps "resource:action" strings to their bit position.
// In production this is loaded from the permissions table at startup.
var permissionBitPosition = map[string]int{
	"vehicle:read":        0,
	"vehicle:write":       1,
	"vehicle:delete":      2,
	"driver:read":         3,
	"driver:write":        4,
	"driver:delete":       5,
	"trip:read":           6,
	"trip:write":          7,
	"trip:delete":         8,
	"maintenance:read":    9,
	"maintenance:write":   10,
	"maintenance:delete":  11,
	"report:read":         12,
	"report:write":        13,
	"alert:read":          14,
	"alert:write":         15,
	"alert:delete":        16,
	"setting:read":        17,
	"setting:write":       18,
	"user:read":           19,
	"user:write":          20,
	"user:delete":         21,
	"role:read":           22,
	"role:write":          23,
	"role:delete":         24,
	"billing:read":        25,
	"billing:write":       26,
	"integration:read":    27,
	"integration:write":   28,
	"audit:read":          29,
	"governance:read":     30,
	"governance:write":    31,
}

type RBACService struct {
	roleRepo *repository.RoleRepository
	log      *zap.Logger
}

func NewRBACService(roleRepo *repository.RoleRepository, log *zap.Logger) *RBACService {
	return &RBACService{
		roleRepo: roleRepo,
		log:      log,
	}
}

// ListRoles returns all roles for an organization.
func (s *RBACService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]models.Role, error) {
	return s.roleRepo.ListRoles(ctx, tenantID)
}

// GetRole returns a single role by ID.
func (s *RBACService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// CreateRole creates a new role with the given permissions.
func (s *RBACService) CreateRole(ctx context.Context, tenantID uuid.UUID, name, description string, permissionStrings []string, createdBy uuid.UUID) (*models.Role, error) {
	mask, err := s.permissionsToMask(permissionStrings)
	if err != nil {
		return nil, err
	}

	role := &models.Role{
		TenantID:          tenantID,
		Name:           name,
		Description:    description,
		PermissionMask: mask,
		CreatedBy:      createdBy,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	s.log.Info("role_created", zap.String("role_id", role.ID.String()), zap.String("name", name))
	return role, nil
}

// UpdateRole updates a role's name and description, scoped to the caller's org.
func (s *RBACService) UpdateRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, name, description string) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	if role.TenantID != tenantID {
		return ErrRoleNotFound // Don't reveal role existence in other orgs
	}
	if role.IsSystem {
		return ErrRoleIsSystem
	}

	return s.roleRepo.Update(ctx, id, name, description)
}

// UpdateRolePermissions updates a role's permission mask, scoped to the caller's org.
func (s *RBACService) UpdateRolePermissions(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, permissionStrings []string) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	if role.TenantID != tenantID {
		return ErrRoleNotFound
	}
	if role.IsSystem {
		return ErrRoleIsSystem
	}

	mask, err := s.permissionsToMask(permissionStrings)
	if err != nil {
		return err
	}

	return s.roleRepo.UpdatePermissionMask(ctx, id, mask)
}

// CloneRole creates a copy of an existing role with a new name, scoped to the caller's org.
func (s *RBACService) CloneRole(ctx context.Context, sourceID uuid.UUID, tenantID uuid.UUID, name, description string, createdBy uuid.UUID) (*models.Role, error) {
	source, err := s.roleRepo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, ErrRoleNotFound
	}
	if source.TenantID != tenantID {
		return nil, ErrRoleNotFound
	}

	return s.roleRepo.Clone(ctx, sourceID, name, description, createdBy)
}

// DeleteRole soft-deletes a role, scoped to the caller's org (system roles cannot be deleted).
func (s *RBACService) DeleteRole(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	if role.TenantID != tenantID {
		return ErrRoleNotFound
	}
	if role.IsSystem {
		return ErrRoleIsSystem
	}

	return s.roleRepo.Delete(ctx, id)
}

// GetPermissionMatrix returns all roles with their permission breakdown.
func (s *RBACService) GetPermissionMatrix(ctx context.Context, tenantID uuid.UUID) ([]models.RolePermissionMatrixEntry, error) {
	roles, err := s.roleRepo.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	allPerms, err := s.roleRepo.GetAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	// Build a lookup from bit position to permission
	bitToPerm := make(map[int]models.Permission)
	for _, p := range allPerms {
		bitToPerm[p.BitPosition] = p
	}

	entries := make([]models.RolePermissionMatrixEntry, 0, len(roles))
	for _, role := range roles {
		entry := models.RolePermissionMatrixEntry{
			ID:          role.ID,
			Name:        role.Name,
			IsSystem:    role.IsSystem,
			Permissions: make(map[string]map[string]bool),
		}

		// For each bit set in the mask, find the corresponding permission
		for bit := 0; bit < 63; bit++ {
			if role.PermissionMask&(1<<bit) != 0 {
				if perm, ok := bitToPerm[bit]; ok {
					if entry.Permissions[perm.Resource] == nil {
						entry.Permissions[perm.Resource] = make(map[string]bool)
					}
					entry.Permissions[perm.Resource][perm.Action] = true
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// AssignRole assigns a role to a user within an organization.
func (s *RBACService) AssignRole(ctx context.Context, userID, roleID, tenantID, assignedBy uuid.UUID) error {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	return s.roleRepo.AssignRole(ctx, userID, roleID, tenantID, assignedBy)
}

// BulkAssignRoles assigns one or more roles to one or more users.
// Returns the count of successful assignments and any per-user errors.
type BulkAssignResult struct {
	Assigned int                         `json:"assigned_count"`
	Failed   []BulkAssignFailure         `json:"failed"`
}

type BulkAssignFailure struct {
	UserID string `json:"user_id"`
	Error  string `json:"error"`
}

func (s *RBACService) BulkAssignRoles(ctx context.Context, userIDs, roleIDs []uuid.UUID, tenantID, assignedBy uuid.UUID) *BulkAssignResult {
	result := &BulkAssignResult{}

	for _, uid := range userIDs {
		for _, rid := range roleIDs {
			if err := s.AssignRole(ctx, uid, rid, tenantID, assignedBy); err != nil {
				result.Failed = append(result.Failed, BulkAssignFailure{
					UserID: uid.String(),
					Error:  err.Error(),
				})
			} else {
				result.Assigned++
			}
		}
	}

	s.log.Info("bulk_role_assignments_completed",
		zap.Int("assigned", result.Assigned),
		zap.Int("failed", len(result.Failed)),
		zap.Int("user_count", len(userIDs)),
		zap.Int("role_count", len(roleIDs)),
	)

	return result
}

// UnassignRole removes a role from a user within an organization.
func (s *RBACService) UnassignRole(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	return s.roleRepo.UnassignRole(ctx, userID, roleID, tenantID)
}

// GetUserRoles returns all roles assigned to a user in an organization.
func (s *RBACService) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]models.Role, error) {
	return s.roleRepo.GetUserRoles(ctx, userID, tenantID)
}

// GetUserPermissions returns the effective permissions for a user in an organization
// by OR-ing all their role masks.
func (s *RBACService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	mask, err := s.roleRepo.ComputeEffectivePermissionMask(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	allPerms, err := s.roleRepo.GetAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	granted := make([]string, 0, len(allPerms))
	for _, p := range allPerms {
		if mask&(1<<p.BitPosition) != 0 {
			granted = append(granted, fmt.Sprintf("%s:%s", p.Resource, p.Action))
		}
	}

	return granted, nil
}

// ListAllPermissions returns all available permissions in the system.
func (s *RBACService) ListAllPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.roleRepo.GetAllPermissions(ctx)
}

// permissionsToMask converts a slice of "resource:action" strings to a permission bitmask.
func (s *RBACService) permissionsToMask(permissions []string) (int64, error) {
	var mask int64
	for _, perm := range permissions {
		parts := strings.SplitN(perm, ":", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid permission format: %s (expected resource:action)", perm)
		}
		key := strings.ToLower(parts[0]) + ":" + strings.ToLower(parts[1])
		if bit, ok := permissionBitPosition[key]; ok {
			mask |= 1 << bit
		} else {
			// Unknown permission — skip but log
			s.log.Warn("unknown_permission", zap.String("permission", perm))
		}
	}
	return mask, nil
}
