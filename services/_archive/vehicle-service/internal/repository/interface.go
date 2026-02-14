package repository

import (
	"context"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/google/uuid"
)

type VehicleRepositoryInterface interface {
	Create(ctx context.Context, vehicle *models.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error)
	Update(ctx context.Context, vehicle *models.Vehicle) error
	Delete(ctx context.Context, id uuid.UUID) error
	AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error
	
	// Location tracking methods can be added here when implemented
}