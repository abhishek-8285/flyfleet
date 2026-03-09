package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance_v2/models"
	"github.com/google/uuid"
)

type DVIRRepository interface {
	CreateDVIR(ctx context.Context, report *models.DVIRReport) error
	CreateDVIRDefect(ctx context.Context, defect *models.DVIRDefect) error
	GetDVIRsByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.DVIRReport, error)
}

type dvirRepo struct {
	db *sql.DB
}

func NewDVIRRepository(db *sql.DB) DVIRRepository {
	return &dvirRepo{db: db}
}

func (r *dvirRepo) CreateDVIR(ctx context.Context, report *models.DVIRReport) error {
	query := `
		INSERT INTO dvir_reports (
			id, driver_id, vehicle_id, inspection_type, condition, notes, signature, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`
	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.DriverID, report.VehicleID, report.InspectionType, report.Condition, report.Notes, report.Signature, time.Now())
	return err
}

func (r *dvirRepo) CreateDVIRDefect(ctx context.Context, defect *models.DVIRDefect) error {
	query := `
		INSERT INTO dvir_defects (
			id, dvir_id, defect_type, description, resolved, resolved_by, resolved_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`
	_, err := r.db.ExecContext(ctx, query,
		defect.ID, defect.DVIRID, defect.DefectType, defect.Description, defect.Resolved, defect.ResolvedBy, defect.ResolvedAt)
	return err
}

func (r *dvirRepo) GetDVIRsByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.DVIRReport, error) {
	query := `SELECT id, driver_id, vehicle_id, inspection_type, condition, notes, signature, created_at FROM dvir_reports WHERE vehicle_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.DVIRReport
	for rows.Next() {
		var rep models.DVIRReport
		if err := rows.Scan(&rep.ID, &rep.DriverID, &rep.VehicleID, &rep.InspectionType, &rep.Condition, &rep.Notes, &rep.Signature, &rep.CreatedAt); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, nil
}
