package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	MinVehicleYear = 1900
	MaxVehicleYear = 2030
	MinCapacity = 1
	MinOdometer = 0
)

type Vehicle struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrgID        uuid.UUID  `json:"org_id" db:"org_id"`
	LicensePlate string     `json:"registration_number" db:"license_plate"`
	Model        string     `json:"model" db:"model"`
	Make         string     `json:"make" db:"make"`
	Year         int        `json:"year" db:"year"`
	Color        *string    `json:"color" db:"color"`
	Status       string     `json:"status" db:"status"` // active, inactive, maintenance, retired
	VehicleType  string     `json:"type" db:"vehicle_type"`
	Capacity     *int       `json:"capacity" db:"capacity"`
	FuelType     *string    `json:"fuel_type" db:"fuel_type"` // petrol, diesel, electric, hybrid
	VIN          string     `json:"vin" db:"vin"`
	Registration *string    `json:"registration" db:"registration"`
	Insurance    *string    `json:"insurance" db:"insurance"`
	LastService  *time.Time `json:"last_service" db:"last_service"`
	Odometer     *int64     `json:"odometer" db:"odometer"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type VehicleLocation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	VehicleID uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	Speed     float64   `json:"speed" db:"speed"`
	Heading   float64   `json:"heading" db:"heading"`
	Accuracy  float64   `json:"accuracy" db:"accuracy"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type CreateVehicleRequest struct {
	LicensePlate string `json:"registration_number" validate:"required"`
	Model        string `json:"model" validate:"required"`
	Make         string `json:"make" validate:"required"`
	Year         int    `json:"year" validate:"required,min=1900,max=2030"`
	Color        string `json:"color"`
	Capacity     int    `json:"capacity" validate:"min=1"`
	FuelType     string `json:"fuel_type" validate:"required,oneof=petrol diesel electric hybrid"`
	VIN          string `json:"vin"`
	Registration string `json:"registration"`
	Insurance    string `json:"insurance"`
}

type UpdateVehicleRequest struct {
	Model       *string `json:"model"`
	Make        *string `json:"make"`
	Year        *int    `json:"year" validate:"min=1900,max=2030"`
	Color       *string `json:"color"`
	Status      *string `json:"status" validate:"oneof=active inactive maintenance retired"`
	Capacity    *int    `json:"capacity" validate:"min=1"`
	FuelType    *string `json:"fuel_type" validate:"oneof=petrol diesel electric hybrid"`
	Registration *string `json:"registration"`
	Insurance   *string `json:"insurance"`
	Odometer    *int64  `json:"odometer" validate:"min=0"`
}

type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type AssignDriverRequest struct {
	DriverID uuid.UUID `json:"driver_id" validate:"required"`
}

type VehicleStats struct {
	TotalVehicles  int     `json:"total_vehicles"`
	ActiveVehicles int     `json:"active_vehicles"`
	TotalDistance  float64 `json:"total_distance"`
	TotalFuel      float64 `json:"total_fuel"`
}