package repository

import (
	"context"
	"errors"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) ListRoles(ctx context.Context, orgID uuid.UUID) ([]models.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, name, description, is_system, permission_mask,
			   created_by, created_at, updated_at
		FROM roles WHERE org_id = $1 AND deleted_at IS NULL ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.OrgID, &role.Name, &role.Description,
			&role.IsSystem, &role.PermissionMask, &role.CreatedBy, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.pool.QueryRow(ctx, `
		SELECT id, org_id, name, description, is_system, permission_mask,
			   created_by, created_at, updated_at
		FROM roles WHERE id = $1 AND deleted_at IS NULL`, id).Scan(
		&role.ID, &role.OrgID, &role.Name, &role.Description, &role.IsSystem,
		&role.PermissionMask, &role.CreatedBy, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO roles (org_id, name, description, is_system, permission_mask, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		role.OrgID, role.Name, role.Description, role.IsSystem, role.PermissionMask, role.CreatedBy,
	).Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
}

func (r *RoleRepository) Update(ctx context.Context, id uuid.UUID, name, description string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE roles SET name = $1, description = $2, updated_at = NOW() WHERE id = $3",
		name, description, id,
	)
	return err
}

func (r *RoleRepository) UpdatePermissionMask(ctx context.Context, id uuid.UUID, mask int64) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE roles SET permission_mask = $1, updated_at = NOW() WHERE id = $2",
		mask, id,
	)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE roles SET deleted_at = NOW() WHERE id = $1 AND is_system = false",
		id,
	)
	return err
}

func (r *RoleRepository) Clone(ctx context.Context, sourceID uuid.UUID, name, description string, createdBy uuid.UUID) (*models.Role, error) {
	var source models.Role
	err := r.pool.QueryRow(ctx, `
		SELECT org_id, permission_mask FROM roles WHERE id = $1`, sourceID).Scan(
		&source.OrgID, &source.PermissionMask,
	)
	if err != nil {
		return nil, err
	}

	clone := &models.Role{
		OrgID:          source.OrgID,
		Name:           name,
		Description:    description,
		PermissionMask: source.PermissionMask,
		CreatedBy:      createdBy,
	}

	if err := r.Create(ctx, clone); err != nil {
		return nil, err
	}
	return clone, nil
}

func (r *RoleRepository) GetAllPermissions(ctx context.Context) ([]models.Permission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, resource, action, description, bit_position
		FROM permissions ORDER BY bit_position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action, &p.Description, &p.BitPosition); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *RoleRepository) GetUserRoles(ctx context.Context, userID, orgID uuid.UUID) ([]models.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.org_id, r.name, r.description, r.is_system, r.permission_mask,
			   r.created_by, r.created_at, r.updated_at
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.org_id = $2`, userID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.OrgID, &role.Name, &role.Description,
			&role.IsSystem, &role.PermissionMask, &role.CreatedBy, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) AssignRole(ctx context.Context, userID, roleID, orgID, assignedBy uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, org_id, assigned_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id, org_id) DO NOTHING`,
		userID, roleID, orgID, assignedBy,
	)
	return err
}

func (r *RoleRepository) UnassignRole(ctx context.Context, userID, roleID, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND org_id = $3`,
		userID, roleID, orgID,
	)
	return err
}

func (r *RoleRepository) ComputeEffectivePermissionMask(ctx context.Context, userID, orgID uuid.UUID) (int64, error) {
	var mask int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(r.permission_mask), 0)
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.org_id = $2`, userID, orgID).Scan(&mask)
	if err != nil {
		return 0, err
	}
	// For multiple roles, we need to OR all masks
	rows, err := r.pool.Query(ctx, `
		SELECT r.permission_mask
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.org_id = $2`, userID, orgID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	mask = 0
	for rows.Next() {
		var m int64
		if err := rows.Scan(&m); err != nil {
			return 0, err
		}
		mask |= m
	}
	return mask, nil
}
