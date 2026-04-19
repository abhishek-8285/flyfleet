package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
)

// VendorRepository handles maintenance vendor database operations
type VendorRepository struct {
	db *pgxpool.Pool
}

// NewVendorRepository creates a new vendor repository
func NewVendorRepository(db *pgxpool.Pool) *VendorRepository {
	return &VendorRepository{db: db}
}

// Create inserts a new vendor
func (r *VendorRepository) Create(ctx context.Context, vendor *models.MaintenanceVendor) error {
	query := `
		INSERT INTO maintenance_vendors (
			tenant_id, name, specialization, rating,
			contact_info, payment_terms, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		vendor.OrgID, vendor.Name, vendor.Specialization, vendor.Rating,
		vendor.ContactInfo, vendor.PaymentTerms, vendor.Status,
	).Scan(&vendor.ID, &vendor.CreatedAt, &vendor.UpdatedAt)
}

// GetByID retrieves a vendor by ID
func (r *VendorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MaintenanceVendor, error) {
	query := `
		SELECT id, tenant_id, name, specialization, rating,
			   contact_info, payment_terms, status, created_at, updated_at
		FROM maintenance_vendors WHERE id = $1
	`
	v := &models.MaintenanceVendor{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.OrgID, &v.Name, &v.Specialization, &v.Rating,
		&v.ContactInfo, &v.PaymentTerms, &v.Status, &v.CreatedAt, &v.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return v, err
}

// Update updates a vendor
func (r *VendorRepository) Update(ctx context.Context, vendor *models.MaintenanceVendor) error {
	query := `
		UPDATE maintenance_vendors SET
			name = $2, specialization = $3, rating = $4,
			contact_info = $5, payment_terms = $6, status = $7
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		vendor.ID, vendor.Name, vendor.Specialization, vendor.Rating,
		vendor.ContactInfo, vendor.PaymentTerms, vendor.Status,
	).Scan(&vendor.UpdatedAt)
}

// Delete deletes a vendor
func (r *VendorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM maintenance_vendors WHERE id = $1", id)
	return err
}

// List retrieves vendors with filters
func (r *VendorRepository) List(ctx context.Context, orgID uuid.UUID, search string, specialization string) ([]models.MaintenanceVendor, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if specialization != "" {
		conditions = append(conditions, fmt.Sprintf("specialization = $%d", argIdx))
		args = append(args, specialization)
	}

	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, specialization, rating,
			   contact_info, payment_terms, status, created_at, updated_at
		FROM maintenance_vendors WHERE %s
		ORDER BY name ASC
	`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendors []models.MaintenanceVendor
	for rows.Next() {
		var v models.MaintenanceVendor
		if err := rows.Scan(
			&v.ID, &v.OrgID, &v.Name, &v.Specialization, &v.Rating,
			&v.ContactInfo, &v.PaymentTerms, &v.Status, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		vendors = append(vendors, v)
	}
	return vendors, nil
}
