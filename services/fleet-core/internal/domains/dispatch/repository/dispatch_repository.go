package repository

import (
	"context"

	"github.com/google/uuid"
)

type DispatchRepository interface {
	SaveAssignment(ctx context.Context, assignment interface{}) error
	GetActiveTripsForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]interface{}, error)
}

type dispatchRepository struct {
	// db connection
}

func NewDispatchRepository() DispatchRepository {
	return &dispatchRepository{}
}

func (r *dispatchRepository) SaveAssignment(ctx context.Context, assignment interface{}) error {
	return nil
}

func (r *dispatchRepository) GetActiveTripsForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]interface{}, error) {
	return nil, nil
}
