package models

import (
	"time"

	"github.com/google/uuid"
)

type VehicleLifecycleMetrics struct {
	VehicleID                 uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	TotalMaintenanceCost      float64   `json:"total_maintenance_cost" db:"total_maintenance_cost"`
	MaintenanceEvents         int       `json:"maintenance_events" db:"maintenance_events"`
	DowntimeHours             float64   `json:"downtime_hours" db:"downtime_hours"`
	CostPerKm                 float64   `json:"cost_per_km" db:"cost_per_km"`
	ReplacementRecommendation string    `json:"replacement_recommendation" db:"replacement_recommendation"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

type CostPerVehicle struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	TotalCost float64   `json:"total_cost"`
}

type CostPerCategory struct {
	Category  string  `json:"category"`
	TotalCost float64 `json:"total_cost"`
}

type VendorCostComparison struct {
	VendorID   uuid.UUID `json:"vendor_id"`
	VendorName string    `json:"vendor_name"`
	TotalCost  float64   `json:"total_cost"`
	JobCount   int       `json:"job_count"`
	AvgCost    float64   `json:"avg_cost"`
}
