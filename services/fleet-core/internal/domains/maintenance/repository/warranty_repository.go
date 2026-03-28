package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/google/uuid"
)

func (r *MaintenanceRepositoryImpl) CreateWarranty(ctx context.Context, w *models.ComponentWarranty) error {
	query := `
		INSERT INTO component_warranties (
			id, org_id, vehicle_id, part_number, warranty_provider,
			warranty_start, warranty_end, warranty_terms, claim_contact, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	w.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		w.ID, w.OrgID, w.VehicleID, w.PartNumber, w.WarrantyProvider,
		w.WarrantyStart, w.WarrantyEnd, w.WarrantyTerms, w.ClaimContact, w.CreatedAt,
	)
	return err
}

func (r *MaintenanceRepositoryImpl) CheckWarranty(ctx context.Context, vehicleID uuid.UUID, partNumber string) (*models.ComponentWarranty, error) {
	query := `
		SELECT id, org_id, vehicle_id, part_number, warranty_provider,
			   warranty_start, warranty_end, warranty_terms, claim_contact, created_at
		FROM component_warranties
		WHERE vehicle_id = $1 AND part_number = $2 AND warranty_end >= CURRENT_DATE
		LIMIT 1`

	var w models.ComponentWarranty
	err := r.db.QueryRow(ctx, query, vehicleID, partNumber).Scan(
		&w.ID, &w.OrgID, &w.VehicleID, &w.PartNumber, &w.WarrantyProvider,
		&w.WarrantyStart, &w.WarrantyEnd, &w.WarrantyTerms, &w.ClaimContact, &w.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *MaintenanceRepositoryImpl) GetWarrantyCoverage(ctx context.Context, vehicleID uuid.UUID, parts []string) ([]models.ComponentWarranty, error) {
	query := `
		SELECT id, org_id, vehicle_id, part_number, warranty_provider,
			   warranty_start, warranty_end, warranty_terms, claim_contact, created_at
		FROM component_warranties
		WHERE vehicle_id = $1 AND part_number = ANY($2) AND warranty_end >= CURRENT_DATE`

	rows, err := r.db.Query(ctx, query, vehicleID, parts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ComponentWarranty
	for rows.Next() {
		var w models.ComponentWarranty
		err := rows.Scan(
			&w.ID, &w.OrgID, &w.VehicleID, &w.PartNumber, &w.WarrantyProvider,
			&w.WarrantyStart, &w.WarrantyEnd, &w.WarrantyTerms, &w.ClaimContact, &w.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, w)
	}
	return results, nil
}

func (r *MaintenanceRepositoryImpl) GetWarrantyExpiryAlerts(ctx context.Context, orgID uuid.UUID, days int) ([]models.ComponentWarranty, error) {
	query := `
		SELECT id, org_id, vehicle_id, part_number, warranty_provider,
			   warranty_start, warranty_end, warranty_terms, claim_contact, created_at
		FROM component_warranties
		WHERE org_id = $1 AND warranty_end >= CURRENT_DATE AND warranty_end <= CURRENT_DATE + ($2 * interval '1 day')
		ORDER BY warranty_end ASC`

	rows, err := r.db.Query(ctx, query, orgID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ComponentWarranty
	for rows.Next() {
		var w models.ComponentWarranty
		err := rows.Scan(
			&w.ID, &w.OrgID, &w.VehicleID, &w.PartNumber, &w.WarrantyProvider,
			&w.WarrantyStart, &w.WarrantyEnd, &w.WarrantyTerms, &w.ClaimContact, &w.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, w)
	}
	return results, nil
}
