package repository

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, d *models.InvoiceDocument) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO billing_invoice_documents (invoice_id, template_id, template_version, snapshot, file_url, storage_key, checksum)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, generated_at
	`, d.InvoiceID, d.TemplateID, d.TemplateVersion, d.Snapshot, d.FileURL, d.StorageKey, d.Checksum).Scan(&d.ID, &d.GeneratedAt)
}

func (r *DocumentRepository) GetByInvoiceID(ctx context.Context, invoiceID string) ([]models.InvoiceDocument, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, invoice_id, template_id, template_version, snapshot, file_url, storage_key, checksum, generated_at
		FROM billing_invoice_documents
		WHERE invoice_id = $1
		ORDER BY generated_at DESC
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.InvoiceDocument
	for rows.Next() {
		var d models.InvoiceDocument
		if err := rows.Scan(&d.ID, &d.InvoiceID, &d.TemplateID, &d.TemplateVersion, &d.Snapshot, &d.FileURL, &d.StorageKey, &d.Checksum, &d.GeneratedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}
