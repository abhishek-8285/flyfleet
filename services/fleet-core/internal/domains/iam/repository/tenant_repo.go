package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepository struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

func (r *TenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]models.TenantInfo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.name, uo.role, ''::text,
			   COALESCE((SELECT COUNT(*) FROM vehicles v WHERE v.tenant_id = o.id), 0)
		FROM user_tenants uo
		JOIN tenants o ON o.id = uo.tenant_id
		WHERE uo.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []models.TenantInfo
	for rows.Next() {
		var o models.TenantInfo
		if err := rows.Scan(&o.ID, &o.Name, &o.Role, &o.ThemeColor, &o.VehicleCount); err != nil {
			return nil, err
		}
		tenants = append(tenants, o)
	}
	return tenants, nil
}

func (r *TenantRepository) GetUserTenantRole(ctx context.Context, userID, tenantID uuid.UUID) (string, uuid.UUID, error) {
	var role string
	var roleID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT uo.role, COALESCE(uo.role_id, '00000000-0000-0000-0000-000000000000'::uuid)
		FROM user_tenants uo
		WHERE uo.user_id = $1 AND uo.tenant_id = $2`,
		userID, tenantID,
	).Scan(&role, &roleID)
	if err != nil {
		return "", uuid.Nil, err
	}
	return role, roleID, nil
}

func (r *TenantRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var o models.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, ''::text, ''::text, created_at, updated_at
		FROM tenants WHERE id = $1`, id).Scan(
		&o.ID, &o.Name, &o.ThemeColor, &o.ThemeLogoURL, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *TenantRepository) LogTenantSwitch(ctx context.Context, userID, fromTenantID, toTenantID uuid.UUID, ip, userAgent string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tenant_switch_events (user_id, from_tenant_id, to_tenant_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, fromTenantID, toTenantID, ip, userAgent,
	)
	return err
}

func (r *TenantRepository) AddUserToTenant(ctx context.Context, userID, tenantID uuid.UUID, role string, roleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_tenants (user_id, tenant_id, role, role_id, is_active)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (user_id, tenant_id) DO UPDATE 
		SET role = EXCLUDED.role, role_id = EXCLUDED.role_id, is_active = true`,
		userID, tenantID, role, roleID,
	)
	return err
}
