package repository

import (
	"context"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
)

// CloneRole creates a new role with the same permissions as the source role
func (r *postgresRepository) CloneRole(ctx context.Context, sourceRoleID uuid.UUID, newRoleName, newRoleDescription string) (*models.Role, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Get source role to check org and existence
	var sourceRole models.Role
	err = tx.QueryRow(ctx, `SELECT organization_id, is_system FROM roles WHERE id = $1`, sourceRoleID).Scan(&sourceRole.OrganizationID, &sourceRole.IsSystem)
	if err != nil {
		return nil, err
	}

	// 2. Create new role
	newRoleID := uuid.New()
	newRole := &models.Role{
		ID:             newRoleID,
		Name:           newRoleName,
		Description:    newRoleDescription,
		OrganizationID: sourceRole.OrganizationID,
		IsSystem:       false, // Cloned roles are never system roles
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO roles (id, name, description, organization_id, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, newRole.ID, newRole.Name, newRole.Description, newRole.OrganizationID, newRole.IsSystem, newRole.CreatedAt, newRole.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 3. Copy permissions
	_, err = tx.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT $1, permission_id FROM role_permissions WHERE role_id = $2
	`, newRole.ID, sourceRoleID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return newRole, nil
}

// GetEffectivePermissions returns all unique permissions for a user across all their assigned roles
func (r *postgresRepository) GetEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]models.PermissionDefinition, error) {
	query := `
		SELECT DISTINCT p.id, p.module, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.module, p.action ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.PermissionDefinition
	for rows.Next() {
		var perm models.PermissionDefinition
		if err := rows.Scan(&perm.ID, &perm.Module, &perm.Action, &perm.Description); err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

// CountUsersWithRole counts how many users have a specific role, optionally scoped to an org
func (r *postgresRepository) CountUsersWithRole(ctx context.Context, roleName string, orgID *uuid.UUID) (int, error) {
	var count int
	var err error
	
	if orgID != nil {
		// Count users with this role in specific org
		query := `
			SELECT COUNT(DISTINCT ur.user_id)
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name = $1 AND r.organization_id = $2
		`
		err = r.db.QueryRow(ctx, query, roleName, *orgID).Scan(&count)
	} else {
		// Global count (typically for system roles like SuperAdmin which might not be bound to orgs or checking globally)
		query := `
			SELECT COUNT(DISTINCT ur.user_id)
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name = $1
		`
		err = r.db.QueryRow(ctx, query, roleName).Scan(&count)
	}
	
	return count, err
}

func (r *postgresRepository) GetRoleByName(ctx context.Context, orgID uuid.UUID, name string) (*models.Role, error) {
	query := `SELECT id, name, description, organization_id, is_system, created_at, updated_at FROM roles WHERE organization_id = $1 AND name = $2`
	rows, err := r.db.Query(ctx, query, orgID, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.OrganizationID, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		return &role, nil
	}
	return nil, nil
}

func (r *postgresRepository) CreatePermission(ctx context.Context, perm *models.PermissionDefinition) error {
	query := `INSERT INTO permissions (id, module, action, description) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, perm.ID, perm.Module, perm.Action, perm.Description)
	return err
}

func (r *postgresRepository) GetPermission(ctx context.Context, module, action string) (*models.PermissionDefinition, error) {
	query := `SELECT id, module, action, description FROM permissions WHERE module = $1 AND action = $2`
	var p models.PermissionDefinition
	err := r.db.QueryRow(ctx, query, module, action).Scan(&p.ID, &p.Module, &p.Action, &p.Description)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
