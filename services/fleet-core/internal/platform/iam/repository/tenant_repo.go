package repository

import (
	"context"
	"database/sql"
	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresTenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) TenantRepository {
	return &postgresTenantRepository{db: db}
}

func (r *postgresTenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `INSERT INTO tenants (id, name, tax_id, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, tenant.ID, tenant.Name, tenant.TaxID, tenant.Status, tenant.CreatedAt, tenant.UpdatedAt)
	return err
}

func (r *postgresTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `SELECT id, name, tax_id, status, created_at, updated_at FROM tenants WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var tenant models.Tenant
	err := row.Scan(&tenant.ID, &tenant.Name, &tenant.TaxID, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *postgresTenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	query := `UPDATE tenants SET name = $1, tax_id = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, tenant.Name, tenant.TaxID, tenant.Status, tenant.ID)
	return err
}

func (r *postgresTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
