package models

import (
	"time"

	"github.com/google/uuid"
)

// VehicleFinancials holds acquisition cost, depreciation, and running cost data.
type VehicleFinancials struct {
	VehicleID            uuid.UUID  `json:"vehicle_id"             db:"vehicle_id"`
	AcquisitionCost      *float64   `json:"acquisition_cost"       db:"acquisition_cost"`
	DepreciationMethod   *string    `json:"depreciation_method"    db:"depreciation_method"` // STRAIGHT_LINE, DECLINING_BALANCE
	DepreciationRate     *float64   `json:"depreciation_rate"      db:"depreciation_rate"`
	ResidualValue        *float64   `json:"residual_value"         db:"residual_value"`
	EMIAmount            *float64   `json:"emi_amount"             db:"emi_amount"`
	LeaseEndDate         *time.Time `json:"lease_end_date"         db:"lease_end_date"`
	InsurancePremium     *float64   `json:"insurance_premium"      db:"insurance_premium"`
	TotalMaintenanceCost float64    `json:"total_maintenance_cost" db:"total_maintenance_cost"`
	TotalFuelCost        float64    `json:"total_fuel_cost"        db:"total_fuel_cost"`
	TotalTollCost        float64    `json:"total_toll_cost"        db:"total_toll_cost"`
	TCOCached            *float64   `json:"tco_cached"             db:"tco_cached"`
	TCOCalculatedAt      *time.Time `json:"tco_calculated_at"      db:"tco_calculated_at"`
	CreatedAt            time.Time  `json:"created_at"             db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"             db:"updated_at"`
}

// UpsertFinancialsRequest is the request body for setting vehicle financial data.
type UpsertFinancialsRequest struct {
	AcquisitionCost    *float64   `json:"acquisition_cost"`
	DepreciationMethod *string    `json:"depreciation_method" validate:"omitempty,oneof=STRAIGHT_LINE DECLINING_BALANCE UNITS_OF_PRODUCTION"`
	DepreciationRate   *float64   `json:"depreciation_rate"`
	ResidualValue      *float64   `json:"residual_value"`
	EMIAmount          *float64   `json:"emi_amount"`
	LeaseEndDate       *time.Time `json:"lease_end_date"`
	InsurancePremium   *float64   `json:"insurance_premium"`
}
