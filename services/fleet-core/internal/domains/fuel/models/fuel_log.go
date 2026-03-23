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
	ID               uuid.UUID     `json:"id" db:"id" validate:"required,uuid"`
	OrganizationID   uuid.UUID     `json:"organization_id" db:"organization_id" validate:"required,uuid"`
	VehicleID        uuid.UUID     `json:"vehicle_id" db:"vehicle_id" validate:"required,uuid"`
	DriverID         uuid.UUID     `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	StationName      string        `json:"station_name" db:"station_name" validate:"required,max=255"`
	FuelType         string        `json:"fuel_type" db:"fuel_type" validate:"required,max=50"`
	Quantity         float64       `json:"quantity" db:"quantity" validate:"required,gt=0"`
	UnitPrice        float64       `json:"unit_price" db:"unit_price" validate:"required,gt=0"`
	TotalCost        float64       `json:"total_cost" db:"total_cost" validate:"required,gt=0"`
	Odometer         float64       `json:"odometer" db:"odometer" validate:"required,gte=0"`
	TransactionDate  time.Time     `json:"transaction_date" db:"transaction_date" validate:"required"`
	Status           FuelLogStatus `json:"status" db:"status" validate:"required,oneof=PENDING APPROVED REJECTED"`
	GSTNumber        *string       `json:"gst_number,omitempty" db:"gst_number" validate:"omitempty,len=15"`
	GSTRate          *float64      `json:"gst_rate,omitempty" db:"gst_rate"`
	GSTAmount        *float64      `json:"gst_amount,omitempty" db:"gst_amount"`
	InvoiceNumber    *string       `json:"invoice_number,omitempty" db:"invoice_number"`
	VendorGSTIN      *string       `json:"vendor_gstin,omitempty" db:"vendor_gstin" validate:"omitempty,len=15"`
	ReceiptURL       *string       `json:"receipt_url,omitempty" db:"receipt_url" validate:"omitempty,url"`
	Notes            *string       `json:"notes,omitempty" db:"notes" validate:"omitempty,max=1000"`
	AnomalyFlag      bool          `json:"anomaly_flag" db:"anomaly_flag"`
	AnomalyReason    *string       `json:"anomaly_reason,omitempty" db:"anomaly_reason"`
	CreatedBy        uuid.UUID     `json:"created_by" db:"created_by" validate:"required,uuid"`
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
	VehicleID       uuid.UUID `json:"vehicle_id" validate:"required,uuid"`
	DriverID        uuid.UUID `json:"driver_id" validate:"required,uuid"`
	StationName     string    `json:"station_name" validate:"required,max=255"`
	FuelType        string    `json:"fuel_type" validate:"required,max=50"`
	Quantity        float64   `json:"quantity" validate:"required,gt=0"`
	UnitPrice       float64   `json:"unit_price" validate:"required,gt=0"`
	TotalCost       float64   `json:"total_cost" validate:"required,gt=0"`
	Odometer        float64   `json:"odometer" validate:"required,gte=0"`
	TransactionDate time.Time `json:"transaction_date" validate:"required"`
	GSTNumber       *string   `json:"gst_number,omitempty" validate:"omitempty,len=15"`
	GSTRate         *float64  `json:"gst_rate,omitempty"`
	InvoiceNumber   *string   `json:"invoice_number,omitempty"`
	VendorGSTIN     *string   `json:"vendor_gstin,omitempty" validate:"omitempty,len=15"`
	ReceiptURL      *string   `json:"receipt_url,omitempty" validate:"omitempty,url"`
	Notes           *string   `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type UpdateFuelLogRequest struct {
	StationName     *string    `json:"station_name,omitempty" validate:"omitempty,max=255"`
	FuelType        *string    `json:"fuel_type,omitempty" validate:"omitempty,max=50"`
	Quantity        *float64   `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice       *float64   `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	TotalCost       *float64   `json:"total_cost,omitempty" validate:"omitempty,gt=0"`
	Odometer        *float64   `json:"odometer,omitempty" validate:"omitempty,gte=0"`
	TransactionDate *time.Time `json:"transaction_date,omitempty"`
	GSTNumber       *string    `json:"gst_number,omitempty" validate:"omitempty,len=15"`
	GSTRate         *float64   `json:"gst_rate,omitempty"`
	InvoiceNumber   *string    `json:"invoice_number,omitempty"`
	VendorGSTIN     *string    `json:"vendor_gstin,omitempty" validate:"omitempty,len=15"`
	ReceiptURL      *string    `json:"receipt_url,omitempty" validate:"omitempty,url"`
	Notes           *string    `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type ChangeFuelLogStatusRequest struct {
	Status FuelLogStatus `json:"status" validate:"required,oneof=APPROVED REJECTED"`
	Notes  *string       `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type FuelLogFilter struct {
	OrganizationID uuid.UUID       `json:"organization_id" validate:"required,uuid"`
	VehicleID      *uuid.UUID      `json:"vehicle_id" validate:"omitempty,uuid"`
	DriverID       *uuid.UUID      `json:"driver_id" validate:"omitempty,uuid"`
	Status         []FuelLogStatus `json:"status"`
	StartDate      *time.Time      `json:"start_date"`
	EndDate        *time.Time      `json:"end_date"`
	Search         string          `json:"search" validate:"omitempty,max=255"`
	Page           int             `json:"page" validate:"gte=1"`
	PageSize       int             `json:"page_size" validate:"gte=1,lte=100"`
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
