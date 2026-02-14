package models

import (
	"time"

	"github.com/google/uuid"
)

type FuelLogStatus string

const (
	FuelLogStatusPending  FuelLogStatus = "PENDING"
	FuelLogStatusApproved FuelLogStatus = "APPROVED"
	FuelLogStatusRejected FuelLogStatus = "REJECTED"
)

type FuelLog struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	OrganizationID   uuid.UUID     `json:"organization_id" db:"organization_id"`
	VehicleID        uuid.UUID     `json:"vehicle_id" db:"vehicle_id"`
	DriverID         uuid.UUID     `json:"driver_id" db:"driver_id"`
	StationName      string        `json:"station_name" db:"station_name"`
	FuelType         string        `json:"fuel_type" db:"fuel_type"`
	Quantity         float64       `json:"quantity" db:"quantity"`
	UnitPrice        float64       `json:"unit_price" db:"unit_price"`
	TotalCost        float64       `json:"total_cost" db:"total_cost"`
	Odometer         float64       `json:"odometer" db:"odometer"`
	TransactionDate  time.Time     `json:"transaction_date" db:"transaction_date"`
	Status           FuelLogStatus `json:"status" db:"status"`
	ReceiptURL       *string       `json:"receipt_url,omitempty" db:"receipt_url"`
	Notes            *string       `json:"notes,omitempty" db:"notes"`
	CreatedBy        uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`

	// Joined data
	VehicleInfo *VehicleInfo `json:"vehicle,omitempty"`
	DriverInfo  *DriverInfo  `json:"driver,omitempty"`
}

type VehicleInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Plate     string    `json:"plate"`
	FleetCode string    `json:"fleet_code"`
}

type DriverInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CreateFuelLogRequest struct {
	VehicleID       uuid.UUID `json:"vehicle_id" binding:"required"`
	DriverID        uuid.UUID `json:"driver_id" binding:"required"`
	StationName     string    `json:"station_name" binding:"required"`
	FuelType        string    `json:"fuel_type" binding:"required"`
	Quantity        float64   `json:"quantity" binding:"required,gt=0"`
	UnitPrice       float64   `json:"unit_price" binding:"required,gt=0"`
	TotalCost       float64   `json:"total_cost" binding:"required,gt=0"`
	Odometer        float64   `json:"odometer" binding:"required"`
	TransactionDate time.Time `json:"transaction_date" binding:"required"`
	ReceiptURL      *string   `json:"receipt_url,omitempty"`
	Notes           *string   `json:"notes,omitempty"`
}

type UpdateFuelLogRequest struct {
	StationName     *string    `json:"station_name,omitempty"`
	FuelType        *string    `json:"fuel_type,omitempty"`
	Quantity        *float64   `json:"quantity,omitempty" binding:"omitempty,gt=0"`
	UnitPrice       *float64   `json:"unit_price,omitempty" binding:"omitempty,gt=0"`
	TotalCost       *float64   `json:"total_cost,omitempty" binding:"omitempty,gt=0"`
	Odometer        *float64   `json:"odometer,omitempty"`
	TransactionDate *time.Time `json:"transaction_date,omitempty"`
	ReceiptURL      *string    `json:"receipt_url,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
}

type ChangeFuelLogStatusRequest struct {
	Status FuelLogStatus `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	Notes  *string       `json:"notes,omitempty"`
}

type FuelLogFilter struct {
	OrganizationID uuid.UUID       `json:"organization_id"`
	VehicleID      *uuid.UUID      `json:"vehicle_id"`
	DriverID       *uuid.UUID      `json:"driver_id"`
	Status         []FuelLogStatus `json:"status"`
	StartDate      *time.Time      `json:"start_date"`
	EndDate        *time.Time      `json:"end_date"`
	Search         string          `json:"search"`
	Page           int             `json:"page"`
	PageSize       int             `json:"page_size"`
}

type FuelLogListResponse struct {
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Logs     []FuelLog `json:"logs"`
}

type FuelLogStats struct {
	TotalConsumption float64          `json:"total_consumption"`
	TotalCost        float64          `json:"total_cost"`
	AvgCostPerKm     float64          `json:"avg_cost_per_km"`
	AvgConsumption   float64          `json:"avg_consumption"`
	PendingApprovals int              `json:"pending_approvals"`
	ByStatus         map[string]int   `json:"by_status"`
	ByFuelType       map[string]float64 `json:"by_fuel_type"`
}
