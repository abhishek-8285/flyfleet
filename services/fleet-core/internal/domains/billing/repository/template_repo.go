package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepository struct {
	db *pgxpool.Pool
}

func NewTemplateRepository(db *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) GetLatestByName(ctx context.Context, name string) (*models.InvoiceTemplate, error) {
	var t models.InvoiceTemplate
	err := r.db.QueryRow(ctx, `
		SELECT id, name, version, html_content, is_default, created_at
		FROM billing_invoice_templates
		WHERE name = $1
		ORDER BY version DESC
		LIMIT 1
	`, name).Scan(&t.ID, &t.Name, &t.Version, &t.HTMLContent, &t.IsDefault, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepository) GetDefault(ctx context.Context) (*models.InvoiceTemplate, error) {
	var t models.InvoiceTemplate
	err := r.db.QueryRow(ctx, `
		SELECT id, name, version, html_content, is_default, created_at
		FROM billing_invoice_templates
		WHERE is_default = true
		ORDER BY version DESC
		LIMIT 1
	`).Scan(&t.ID, &t.Name, &t.Version, &t.HTMLContent, &t.IsDefault, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepository) Create(ctx context.Context, t *models.InvoiceTemplate) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO billing_invoice_templates (name, version, html_content, is_default)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, t.Name, t.Version, t.HTMLContent, t.IsDefault).Scan(&t.ID, &t.CreatedAt)
}
