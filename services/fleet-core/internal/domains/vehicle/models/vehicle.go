package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	MinVehicleYear = 1900
	MaxVehicleYear = 2030
	MinCapacity    = 1
	MinOdometer    = 0

	// Status ENUM values
	StatusActive         = "ACTIVE"
	StatusMaintenance    = "MAINTENANCE"
	StatusSuspended      = "SUSPENDED"
	StatusDecommissioned = "DECOMMISSIONED"

	// Ownership type ENUM values
	OwnershipOwned    = "OWNED"
	OwnershipLeased   = "LEASED"
	OwnershipFinanced = "FINANCED"
	OwnershipVendor   = "VENDOR"
)

// Vehicle is the enterprise aggregate root for the vehicle domain.
type Vehicle struct {
	ID                 uuid.UUID  `json:"id"                   db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"            db:"tenant_id"`
	RegistrationNumber string     `json:"registration_number"  db:"registration_number"`
	ChassisNumber      *string    `json:"chassis_number"       db:"chassis_number"` // AES-256 encrypted
	EngineNumber       *string    `json:"engine_number"        db:"engine_number"`  // AES-256 encrypted
	VINNumber          *string    `json:"vin_number"           db:"vin_number"`
	VehicleType        string     `json:"vehicle_type"         db:"vehicle_type"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"     db:"capacity_tonnage"`
	FuelType           string     `json:"fuel_type"            db:"fuel_type"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"   db:"fuel_tank_capacity"`
	OwnershipType      string     `json:"ownership_type"       db:"ownership_type"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"    db:"parent_vehicle_id"`
	BranchID           *uuid.UUID `json:"branch_id"            db:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"       db:"cost_center_id"`
	Status             string     `json:"status"               db:"status"`
	GPSDeviceID        *string    `json:"gps_device_id"        db:"gps_device_id"`
	ImmobilizerEnabled bool       `json:"immobilizer_enabled"  db:"immobilizer_enabled"`
	AIS140Compliant    bool       `json:"ais140_compliant"     db:"ais140_compliant"`
	// Legacy / convenience fields
	Make  *string `json:"make"                 db:"make"`
	Model *string `json:"model"                db:"model"`
	Year  *int    `json:"year"                 db:"year"`
	Color *string `json:"color"                db:"color"`
	Odometer           *int64     `json:"odometer"             db:"odometer"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"    db:"current_driver_id"`
	SpeedLimit         float64    `json:"speed_limit"          db:"speed_limit"`


	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Version   int        `json:"version"    db:"version"`
}

// Backward-compat aliases used by existing code / tests.
func (v *Vehicle) OrgID() uuid.UUID     { return v.TenantID }
func (v *Vehicle) LicensePlate() string { return v.RegistrationNumber }

// VehicleLocation is the real-time location snapshot for a vehicle.
type VehicleLocation struct {
	ID        uuid.UUID `json:"id"         db:"id"`
	VehicleID uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	Latitude  float64   `json:"latitude"   db:"latitude"`
	Longitude float64   `json:"longitude"  db:"longitude"`
	Speed     float64   `json:"speed"      db:"speed"`
	Heading   float64   `json:"heading"    db:"heading"`
	Accuracy  float64   `json:"accuracy"   db:"accuracy"`
	Timestamp time.Time `json:"timestamp"  db:"timestamp"`
}

// CreateVehicleRequest is the body for POST /v1/vehicles.
type CreateVehicleRequest struct {
	RegistrationNumber string     `json:"registration_number" validate:"required"`
	VehicleType        string     `json:"vehicle_type"        validate:"required,oneof=TRUCK BUS VAN CAR TRAILER TANKER TIPPER CRANE OTHER"`
	FuelType           string     `json:"fuel_type"           validate:"required,oneof=PETROL DIESEL CNG ELECTRIC HYBRID LPG"`
	OwnershipType      string     `json:"ownership_type"      validate:"required,oneof=OWNED LEASED FINANCED VENDOR"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	ImmobilizerEnabled bool       `json:"immobilizer_enabled"`
	AIS140Compliant    bool       `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year" validate:"omitempty,min=1900,max=2030"`
	Color              *string    `json:"color"`
	Odometer           *int64     `json:"odometer"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"`
	SpeedLimit         float64    `json:"speed_limit"`
	// Legacy fields (still accepted from old clients)
	LicensePlate *string `json:"license_plate"`
}

// UpdateVehicleRequest is the body for PUT /v1/vehicles/{id}.
type UpdateVehicleRequest struct {
	VehicleType        *string    `json:"vehicle_type"      validate:"omitempty,oneof=TRUCK BUS VAN CAR TRAILER TANKER TIPPER CRANE OTHER"`
	FuelType           *string    `json:"fuel_type"         validate:"omitempty,oneof=PETROL DIESEL CNG ELECTRIC HYBRID LPG"`
	OwnershipType      *string    `json:"ownership_type"    validate:"omitempty,oneof=OWNED LEASED FINANCED VENDOR"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	ImmobilizerEnabled *bool      `json:"immobilizer_enabled"`
	AIS140Compliant    *bool      `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year"    validate:"omitempty,min=1900,max=2030"`
	Color              *string    `json:"color"`
	// Legacy fields
	Status    *string `json:"status"   validate:"omitempty,oneof=ACTIVE MAINTENANCE SUSPENDED DECOMMISSIONED active inactive maintenance retired"`
	Odometer           *int64     `json:"odometer" validate:"omitempty,min=0"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"`
	SpeedLimit         *float64   `json:"speed_limit"`
	Insurance          *string    `json:"insurance"`
}

// VehicleListResponse is the paginated list response.
type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

// AssignDriverRequest — body for POST /vehicles/{id}/assign-driver.
type AssignDriverRequest struct {
	DriverID uuid.UUID `json:"driver_id" validate:"required"`
}

// VehicleStats is the aggregated stats response.
type VehicleStats struct {
	TotalVehicles    int     `json:"total_vehicles"`
	ActiveVehicles   int     `json:"active_vehicles"`
	MaintenanceCount int     `json:"maintenance_count"`
	SuspendedCount   int     `json:"suspended_count"`
	TotalDistance    float64 `json:"total_distance"`
	TotalFuel        float64 `json:"total_fuel"`
}

// VehicleSearchParams holds query params for GET /vehicles/search.
type VehicleSearchParams struct {
	TenantID    *uuid.UUID
	Query       string
	Status      string
	VehicleType string
	FuelType    string
	BranchID    *uuid.UUID
	Page        int
	Limit       int
}
