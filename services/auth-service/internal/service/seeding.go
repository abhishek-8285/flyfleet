package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) EnsureSeedData(ctx context.Context) error {
	s.logger.Info("Starting RBAC Seeding...")

	// 1. Seed Permissions
	permissions := []models.PermissionDefinition{
		// Users
		{Module: "user", Action: "read", Description: "View users"},
		{Module: "user", Action: "create", Description: "Create users"},
		{Module: "user", Action: "update", Description: "Update users"},
		{Module: "user", Action: "delete", Description: "Delete users"},
		
		// Roles
		{Module: "roles", Action: "read", Description: "View roles"},
		{Module: "roles", Action: "create", Description: "Create roles"},
		{Module: "roles", Action: "update", Description: "Update roles"},
		{Module: "roles", Action: "delete", Description: "Delete roles"},

		// Organizations
		{Module: "org", Action: "read", Description: "View organizations"},
		{Module: "org", Action: "create", Description: "Create organizations"},
		{Module: "org", Action: "update", Description: "Update organizations"},
		{Module: "org", Action: "delete", Description: "Delete organizations"},
		
		// Vehicles
		{Module: "vehicle", Action: "read", Description: "View vehicles"},
		{Module: "vehicle", Action: "create", Description: "Create vehicles"},
		{Module: "vehicle", Action: "update", Description: "Update vehicles"},
		{Module: "vehicle", Action: "delete", Description: "Delete vehicles"},
		
		// Trips & Drivers
		{Module: "trip", Action: "read", Description: "View trips"},
		{Module: "driver", Action: "read", Description: "View drivers"},
		{Module: "device", Action: "read", Description: "View devices"},
		
		// Reports & Documents
		{Module: "reports", Action: "read", Description: "View reports"},
		{Module: "reports", Action: "export", Description: "Export reports"},
		{Module: "documents", Action: "read", Description: "View documents"},
		
		// Settings
		{Module: "settings", Action: "read", Description: "View settings"},
		{Module: "settings", Action: "update", Description: "Update settings"},

		// Audit
		{Module: "audit_logs", Action: "read", Description: "View audit logs"},

		// Dashboard
		{Module: "dashboard", Action: "read", Description: "View dashboard"},
		
		// Modules from Sidebar
		{Module: "fuel", Action: "read", Description: "Fuel management"},
		{Module: "governance", Action: "read", Description: "Governance"},
		{Module: "disaster_recovery", Action: "read", Description: "DR"},
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

	// 3. Seed Default Organization
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	org, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		s.logger.Info("Creating default organization...")
		org = &models.Organization{
			ID:           orgID,
			Name:         "FlyFleet Global",
			Domain:       "test.com",
			DefaultTheme: "light",
			Status:       "active",
		}
		if err := s.repo.CreateOrganization(ctx, org); err != nil {
			s.logger.Error("Failed to seed default organization", zap.Error(err))
		}
	}

	// 4. Seed Super Admin User
	adminEmail := "admin@test.com"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	admin, err := s.repo.GetUserByEmail(ctx, adminEmail)
	if err != nil || admin == nil {
		s.logger.Info("Creating default admin user...")
		admin = &models.User{
			ID:           uuid.New(),
			Email:        adminEmail,
			Name:         "Admin User",
			PasswordHash: string(hashedPassword),
			Status:       "active",
			Verified:     true,
		}
		if err := s.repo.CreateUser(ctx, admin); err != nil {
			s.logger.Error("Failed to seed admin user", zap.Error(err))
		} else {
			// Link to Org
			if err := s.repo.AddUserToOrganization(ctx, admin.ID, orgID, "super_admin"); err != nil {
				s.logger.Error("Failed to link admin to organization", zap.Error(err))
			}
			
			// Assign Role
			superAdminRole, _ := s.repo.GetRoleByName(ctx, uuid.Nil, "Super Admin")
			if superAdminRole != nil {
				if err := s.repo.AssignRoleToUser(ctx, admin.ID, superAdminRole.ID, uuid.Nil); err != nil {
					s.logger.Error("Failed to assign Super Admin role", zap.Error(err))
				}
				// Also update the primary RoleID on the user struct
				admin.RoleID = &superAdminRole.ID
				s.repo.UpdateUser(ctx, admin)
			}
		}
	} else if admin.PasswordHash == "hash" {
		s.logger.Info("Updating default admin user password...")
		admin.PasswordHash = string(hashedPassword)
		if err := s.repo.UpdateUser(ctx, admin); err != nil {
			s.logger.Error("Failed to update admin password", zap.Error(err))
		}
	}

	// Always ensure Admin is linked to Org and has Role
	if admin != nil {
		// Link to Org if not already linked
		_, err := s.repo.GetUserRoleInOrg(ctx, admin.ID, orgID)
		if err != nil { // Likely "sql: no rows in result set"
			s.logger.Info("Linking admin to organization...")
			if err := s.repo.AddUserToOrganization(ctx, admin.ID, orgID, "super_admin"); err != nil {
				s.logger.Error("Failed to link admin to organization", zap.Error(err))
			}
		}

		// Assign Super Admin Role if not assigned
		superAdminRole, _ := s.repo.GetRoleByName(ctx, uuid.Nil, "Super Admin")
		if superAdminRole != nil {
			// Check if already assigned
			userRoles, _ := s.repo.GetUserRoles(ctx, admin.ID)
			hasRole := false
			for _, r := range userRoles {
				if r.ID == superAdminRole.ID {
					hasRole = true
					break
				}
			}
			if !hasRole {
				s.logger.Info("Assigning Super Admin role to admin...")
				if err := s.repo.AssignRoleToUser(ctx, admin.ID, superAdminRole.ID, uuid.Nil); err != nil {
					s.logger.Error("Failed to assign Super Admin role", zap.Error(err))
				}
			}
		}
	}

	s.logger.Info("RBAC Seeding Completed")
	return nil
}
