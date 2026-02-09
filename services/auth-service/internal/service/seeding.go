package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *AuthService) EnsureSeedData(ctx context.Context) error {
	s.logger.Info("Starting RBAC Seeding...")

	// 1. Seed Permissions
	permissions := []models.PermissionDefinition{
		// Users
		{Module: "users", Action: "read", Description: "View users"},
		{Module: "users", Action: "create", Description: "Create users"},
		{Module: "users", Action: "update", Description: "Update users"},
		{Module: "users", Action: "delete", Description: "Delete users"},
		
		// Roles
		{Module: "roles", Action: "read", Description: "View roles"},
		{Module: "roles", Action: "create", Description: "Create roles"},
		{Module: "roles", Action: "update", Description: "Update roles"},
		{Module: "roles", Action: "delete", Description: "Delete roles"},
		
		// Vehicles
		{Module: "vehicles", Action: "read", Description: "View vehicles"},
		{Module: "vehicles", Action: "create", Description: "Create vehicles"},
		{Module: "vehicles", Action: "update", Description: "Update vehicles"},
		{Module: "vehicles", Action: "delete", Description: "Delete vehicles"},
		
		// Reports
		{Module: "reports", Action: "read", Description: "View reports"},
		{Module: "reports", Action: "export", Description: "Export reports"},
		
		// Settings
		{Module: "settings", Action: "read", Description: "View settings"},
		{Module: "settings", Action: "update", Description: "Update settings"},

		// Audit
		{Module: "audit", Action: "read", Description: "View audit logs"},
	}

	permMap := make(map[string]uuid.UUID)

	for _, p := range permissions {
		existing, err := s.repo.GetPermission(ctx, p.Module, p.Action)
		if err != nil { // likely not found or db error
			// try create
			p.ID = uuid.New()
			if err := s.repo.CreatePermission(ctx, &p); err != nil {
				s.logger.Error("Failed to seed permission", zap.String("module", p.Module), zap.String("action", p.Action), zap.Error(err))
				continue
			}
			permMap[fmt.Sprintf("%s:%s", p.Module, p.Action)] = p.ID
		} else {
			permMap[fmt.Sprintf("%s:%s", existing.Module, existing.Action)] = existing.ID
		}
	}

	// 2. Seed System Roles
	roles := []struct {
		Name        string
		Description string
		Permissions []string // "module:action"
	}{
		{
			Name:        "Super Admin",
			Description: "Full access to the entire system",
			Permissions: []string{"*"}, // special handler or all
		},
		{
			Name:        "Org Admin",
			Description: "Full access to organization resources",
			Permissions: []string{"*"}, // For now same as super admin regarding permissions list, logic handles scope
		},
		{
			Name:        "Member",
			Description: "Standard member access",
			Permissions: []string{
				"vehicles:read", "vehicles:create", "vehicles:update",
				"users:read",
				"reports:read", "reports:export",
				"audit:read",
			},
		},
		{
			Name:        "Viewer",
			Description: "Read-only access",
			Permissions: []string{
				"vehicles:read", "users:read", "roles:read", "reports:read", "settings:read", "audit:read",
			},
		},
	}

	// Helper to get all perm IDs
	getAllPermIDs := func() []uuid.UUID {
		var ids []uuid.UUID
		for _, id := range permMap {
			ids = append(ids, id)
		}
		return ids
	}

	for _, r := range roles {
		role, err := s.repo.GetRoleByName(ctx, uuid.Nil, r.Name)
		if err != nil {
			s.logger.Error("Failed to check role existence", zap.String("role", r.Name), zap.Error(err))
			continue
		}

		var roleID uuid.UUID
		if role == nil {
			// Create
			newRole, err := s.CreateRole(ctx, r.Name, r.Description, uuid.Nil, true, uuid.Nil) // System action
			if err != nil {
				s.logger.Error("Failed to create system role", zap.String("role", r.Name), zap.Error(err))
				continue
			}
			roleID = newRole.ID
			s.logger.Info("Seeded system role", zap.String("role", r.Name))
		} else {
			roleID = role.ID
		}

		// Assign Permissions
		var permIDs []uuid.UUID
		if len(r.Permissions) == 1 && r.Permissions[0] == "*" {
			permIDs = getAllPermIDs()
		} else {
			for _, key := range r.Permissions {
				if id, ok := permMap[key]; ok {
					permIDs = append(permIDs, id)
				}
			}
		}

		if err := s.UpdateRolePermissions(ctx, roleID, permIDs, uuid.Nil); err != nil { // System action
			s.logger.Error("Failed to update role permissions", zap.String("role", r.Name), zap.Error(err))
		}
	}

	s.logger.Info("RBAC Seeding Completed")
	return nil
}
