package models

import (
	"time"

	"github.com/google/uuid"
)

type VehicleSpecs struct {
	VehicleID             uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	FuelType              string    `json:"fuel_type" db:"fuel_type"`
	TankCapacityLiters    float64   `json:"tank_capacity_liters" db:"tank_capacity_liters"`
	EfficiencyExpectedKmL float64   `json:"efficiency_expected_km_l" db:"efficiency_expected_km_l"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}
