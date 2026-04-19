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

// InspectionRepository handles maintenance inspection database operations
type InspectionRepository struct {
	db *pgxpool.Pool
}

// NewInspectionRepository creates a new inspection repository
func NewInspectionRepository(db *pgxpool.Pool) *InspectionRepository {
	return &InspectionRepository{db: db}
}

// --- Templates ---

// CreateTemplate inserts a new inspection template
func (r *InspectionRepository) CreateTemplate(ctx context.Context, template *models.InspectionTemplate) error {
	query := `
		INSERT INTO inspection_templates (
			tenant_id, name, description, applicable_vehicle_types, checklist_items, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		template.OrgID, template.Name, template.Description,
		template.ApplicableVehicleTypes, template.ChecklistItems, template.IsActive,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
}

// GetTemplateByID retrieves a template by ID
func (r *InspectionRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.InspectionTemplate, error) {
	query := `
		SELECT id, tenant_id, name, description, applicable_vehicle_types,
			   checklist_items, is_active, created_at, updated_at
		FROM inspection_templates WHERE id = $1
	`
	t := &models.InspectionTemplate{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.OrgID, &t.Name, &t.Description, &t.ApplicableVehicleTypes,
		&t.ChecklistItems, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// ListTemplates retrieves templates for an organization
func (r *InspectionRepository) ListTemplates(ctx context.Context, orgID uuid.UUID, vehicleType string) ([]models.InspectionTemplate, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	if vehicleType != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(applicable_vehicle_types)", argIdx))
		args = append(args, vehicleType)
	}

	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, applicable_vehicle_types,
			   checklist_items, is_active, created_at, updated_at
		FROM inspection_templates WHERE %s AND is_active = true
		ORDER BY name ASC
	`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.InspectionTemplate
	for rows.Next() {
		var t models.InspectionTemplate
		if err := rows.Scan(
			&t.ID, &t.OrgID, &t.Name, &t.Description, &t.ApplicableVehicleTypes,
			&t.ChecklistItems, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

// --- Inspections ---

// CreateInspection inserts a new vehicle inspection report
func (r *InspectionRepository) CreateInspection(ctx context.Context, inspection *models.VehicleInspection) error {
	query := `
		INSERT INTO vehicle_inspections (
			tenant_id, vehicle_id, template_id, inspector_id,
			status, results, notes, odometer_at_inspection
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		inspection.OrgID, inspection.VehicleID, inspection.TemplateID,
		inspection.InspectorID, inspection.Status, inspection.Results,
		inspection.Notes, inspection.OdometerAtInspection,
	).Scan(&inspection.ID, &inspection.CreatedAt, &inspection.UpdatedAt)
}

// GetInspectionByID retrieves an inspection report by ID
func (r *InspectionRepository) GetInspectionByID(ctx context.Context, id uuid.UUID) (*models.VehicleInspection, error) {
	query := `
		SELECT id, tenant_id, vehicle_id, template_id, inspector_id,
			   status, results, notes, odometer_at_inspection, created_at, updated_at
		FROM vehicle_inspections WHERE id = $1
	`
	i := &models.VehicleInspection{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&i.ID, &i.OrgID, &i.VehicleID, &i.TemplateID, &i.InspectorID,
		&i.Status, &i.Results, &i.Notes, &i.OdometerAtInspection, &i.CreatedAt, &i.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return i, err
}

// ListByVehicle retrieves inspections for a specific vehicle
func (r *InspectionRepository) ListByVehicle(ctx context.Context, vehicleID uuid.UUID, limit int) ([]models.VehicleInspection, error) {
	query := `
		SELECT id, tenant_id, vehicle_id, template_id, inspector_id,
			   status, results, notes, odometer_at_inspection, created_at, updated_at
		FROM vehicle_inspections WHERE vehicle_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inspections []models.VehicleInspection
	for rows.Next() {
		var i models.VehicleInspection
		if err := rows.Scan(
			&i.ID, &i.OrgID, &i.VehicleID, &i.TemplateID, &i.InspectorID,
			&i.Status, &i.Results, &i.Notes, &i.OdometerAtInspection, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		inspections = append(inspections, i)
	}
	return inspections, nil
}
