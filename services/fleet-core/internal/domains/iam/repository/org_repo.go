package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrgRepository struct {
	pool *pgxpool.Pool
}

func NewOrgRepository(pool *pgxpool.Pool) *OrgRepository {
	return &OrgRepository{pool: pool}
}

func (r *OrgRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.OrgInfo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.name, uo.role, ''::text,
			   COALESCE((SELECT COUNT(*) FROM vehicles v WHERE v.tenant_id = o.id), 0)
		FROM user_organizations uo
		JOIN organizations o ON o.id = uo.tenant_id
		WHERE uo.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.OrgInfo
	for rows.Next() {
		var o models.OrgInfo
		if err := rows.Scan(&o.ID, &o.Name, &o.Role, &o.ThemeColor, &o.VehicleCount); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, nil
}

func (r *OrgRepository) GetUserOrgRole(ctx context.Context, userID, orgID uuid.UUID) (string, uuid.UUID, error) {
	var role string
	var roleID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT uo.role, COALESCE(uo.role_id, '00000000-0000-0000-0000-000000000000'::uuid)
		FROM user_organizations uo
		WHERE uo.user_id = $1 AND uo.tenant_id = $2`,
		userID, orgID,
	).Scan(&role, &roleID)
	if err != nil {
		return "", uuid.Nil, err
	}
	return role, roleID, nil
}

func (r *OrgRepository) GetOrgByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var o models.Organization
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, ''::text, ''::text, created_at, updated_at
		FROM organizations WHERE id = $1`, id).Scan(
		&o.ID, &o.Name, &o.ThemeColor, &o.ThemeLogoURL, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrgRepository) LogOrgSwitch(ctx context.Context, userID, fromOrgID, toOrgID uuid.UUID, ip, userAgent string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org_switch_events (user_id, from_org_id, to_org_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, fromOrgID, toOrgID, ip, userAgent,
	)
	return err
}
