package repository

import (
	"context"
	"database/sql"
	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/google/uuid"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresOrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) OrganizationRepository {
	return &postgresOrganizationRepository{db: db}
}

func (r *postgresOrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	query := `INSERT INTO organizations (id, name, domain, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, org.ID, org.Name, org.TaxID, org.Status, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *postgresOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `SELECT id, name, domain, status, created_at, updated_at FROM organizations WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var org models.Organization
	err := row.Scan(&org.ID, &org.Name, &org.TaxID, &org.Status, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *postgresOrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	query := `UPDATE organizations SET name = $1, domain = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, org.Name, org.TaxID, org.Status, org.ID)
	return err
}

func (r *postgresOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
