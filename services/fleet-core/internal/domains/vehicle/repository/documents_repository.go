package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentsRepository is an append-only store. No Update or Delete methods exposed.
type DocumentsRepository struct {
	db *pgxpool.Pool
}

func NewDocumentsRepository(db *pgxpool.Pool) *DocumentsRepository {
	return &DocumentsRepository{db: db}
}

func (r *DocumentsRepository) Insert(ctx context.Context, doc *models.VehicleDocument) error {
	query := `
		INSERT INTO vehicle_documents (id, vehicle_id, document_type, file_url, version, hash, uploaded_at, uploaded_by)
		VALUES ($1, $2, $3, $4,
		  COALESCE((SELECT MAX(version) FROM vehicle_documents WHERE vehicle_id = $2 AND document_type = $3), 0) + 1,
		  $5, NOW(), $6)
		RETURNING version, uploaded_at`

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		doc.ID, doc.VehicleID, doc.DocumentType, doc.FileURL,
		doc.Hash, doc.UploadedBy,
	).Scan(&doc.Version, &doc.UploadedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return fmt.Errorf("failed to insert vehicle document: %w", err)
	}
	return nil
}

func (r *DocumentsRepository) List(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleDocument, error) {
	start := time.Now()
	rows, err := r.db.Query(ctx,
		`SELECT id, vehicle_id, document_type, file_url, version, hash, uploaded_at, uploaded_by
		 FROM vehicle_documents WHERE vehicle_id = $1 ORDER BY uploaded_at DESC`, vehicleID)
	metrics.RecordDBQuery(`SELECT id, vehicle_id, document_type, file_url, version, hash, uploaded_at, uploaded_by FROM vehicle_documents WHERE vehicle_id = $1 ORDER BY uploaded_at DESC`, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var docs []models.VehicleDocument
	for rows.Next() {
		d := models.VehicleDocument{}
		if err := rows.Scan(&d.ID, &d.VehicleID, &d.DocumentType, &d.FileURL, &d.Version, &d.Hash, &d.UploadedAt, &d.UploadedBy); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *DocumentsRepository) Get(ctx context.Context, id uuid.UUID) (*models.VehicleDocument, error) {
	d := &models.VehicleDocument{}
	start := time.Now()
	err := r.db.QueryRow(ctx,
		`SELECT id, vehicle_id, document_type, file_url, version, hash, uploaded_at, uploaded_by
		 FROM vehicle_documents WHERE id = $1`, id,
	).Scan(&d.ID, &d.VehicleID, &d.DocumentType, &d.FileURL, &d.Version, &d.Hash, &d.UploadedAt, &d.UploadedBy)
	metrics.RecordDBQuery(`SELECT id, vehicle_id, document_type, file_url, version, hash, uploaded_at, uploaded_by FROM vehicle_documents WHERE id = $1`, time.Since(start))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return d, nil
}
