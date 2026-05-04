package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/document/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (
			id, tenant_id, entity_type, entity_id, document_type,
			file_name, file_path, file_size_bytes, mime_type,
			expiry_date, status, remarks, metadata, uploaded_by,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW()
		)
	`
	_, err := r.db.Exec(ctx, query,
		doc.ID, doc.TenantID, doc.EntityType, doc.EntityID, doc.DocumentType,
		doc.FileName, doc.FilePath, doc.FileSize, doc.MimeType,
		doc.ExpiryDate, doc.Status, doc.Remarks, doc.Metadata, doc.UploadedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	return nil
}

func (r *DocumentRepository) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.Document, error) {
	query := `
		SELECT id, tenant_id, COALESCE(entity_type, '') as entity_type, entity_id, COALESCE(document_type, '') as document_type,
		       COALESCE(file_name, '') as file_name, COALESCE(file_path, '') as file_path, COALESCE(file_size_bytes, 0) as file_size_bytes, COALESCE(mime_type, '') as mime_type,
		       expiry_date, COALESCE(status, 'PENDING') as status, remarks, COALESCE(metadata, '{}') as metadata, uploaded_by,
		       created_at, updated_at
		FROM documents
		WHERE entity_type = $1 AND entity_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.EntityType, &doc.EntityID, &doc.DocumentType,
			&doc.FileName, &doc.FilePath, &doc.FileSize, &doc.MimeType,
			&doc.ExpiryDate, &doc.Status, &doc.Remarks, &doc.Metadata, &doc.UploadedBy,
			&doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}
	return docs, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE documents SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
