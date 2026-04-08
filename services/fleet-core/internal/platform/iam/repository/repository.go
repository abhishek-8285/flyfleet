package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IAMRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Role, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error)
}

type iamRepository struct {
	db *pgxpool.Pool
}

func NewIAMRepository(db *pgxpool.Pool) IAMRepository {
	return &iamRepository{db: db}
}

func (r *iamRepository) CreateUser(ctx context.Context, user *models.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Insert into users
	queryUser := `INSERT INTO users (id, email, password_hash, name, status, created_at)
	              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(ctx, queryUser, user.ID, user.Email, user.PasswordHash, user.FullName, user.Status, user.CreatedAt)
	if err != nil {
		return err
	}

	// 2. Insert into user_organizations
	queryLink := `INSERT INTO user_organizations (id, user_id, org_id, role, status, created_at)
	              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(ctx, queryLink, uuid.New(), user.ID, user.OrganizationID, "admin", "active", time.Now())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *iamRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT u.id, u.email, u.password_hash, u.name, u.status, u.created_at, uo.org_id
		FROM users u
		LEFT JOIN user_organizations uo ON u.id = uo.user_id
		WHERE u.email = $1
	`
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Status, &user.CreatedAt, &user.OrganizationID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *iamRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT u.id, u.email, u.password_hash, u.name, u.status, u.created_at, uo.org_id
		FROM users u
		LEFT JOIN user_organizations uo ON u.id = uo.user_id
		WHERE u.id = $1
	`
	err := r.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Status, &user.CreatedAt, &user.OrganizationID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *iamRepository) GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Role, error) {
	query := `SELECT id, name, organization_id, description FROM roles WHERE organization_id = $1 OR organization_id IS NULL`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.OrganizationID, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *iamRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, userID, roleID)
	return err
}

func (r *iamRepository) GetPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.resource, p.action
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		p := &models.Permission{}
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}
