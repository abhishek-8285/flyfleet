package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
)

type DowntimeRepository interface {
	RecordDowntimeStart(ctx context.Context, downtime *models.VehicleDowntime) error
	RecordDowntimeEnd(ctx context.Context, id uuid.UUID, endTime time.Time) error
	GetDowntimeByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]*models.VehicleDowntime, error)
}

func (r *MaintenanceRepositoryImpl) RecordDowntimeStart(ctx context.Context, d *models.VehicleDowntime) error {
	query := `
		INSERT INTO vehicle_downtime (id, organization_id, vehicle_id, work_order_id, downtime_start, reason)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.DowntimeStart.IsZero() {
		d.DowntimeStart = time.Now()
	}

	_, err := r.db.Exec(ctx, query, d.ID, d.OrganizationID, d.VehicleID, d.WorkOrderID, d.DowntimeStart, d.Reason)
	return err
}

func (r *MaintenanceRepositoryImpl) RecordDowntimeEnd(ctx context.Context, id uuid.UUID, endTime time.Time) error {
	query := `
		UPDATE vehicle_downtime 
		SET downtime_end = $1, 
		    downtime_minutes = EXTRACT(EPOCH FROM ($1 - downtime_start))/60
		WHERE id = $2`
	_, err := r.db.Exec(ctx, query, endTime, id)
	return err
}
