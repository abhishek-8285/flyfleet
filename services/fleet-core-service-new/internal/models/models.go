package models

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID           uuid.UUID `json:"id"`
	OrgID        uuid.UUID `json:"org_id"`
	LicensePlate string    `json:"license_plate"`
	Model        string    `json:"model"`
	Make         string    `json:"make"`
	Year         int32     `json:"year"`
	Status       string    `json:"status"`
	VIN          string    `json:"vin"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    uuid.UUID `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    uuid.UUID `json:"updated_by"`
}

type Address struct {
	Street    string `json:"street"`
	City      string `json:"city"`
	State     string `json:"state"`
	Pincode   string `json:"pincode"`
	Country   string `json:"country"`
}

type EmergencyContact struct {
	Name     string `json:"name"`
	Relation string `json:"relation"`
	Phone    string `json:"phone"`
}

type Driver struct {
	ID               uuid.UUID         `json:"id"`
	OrgID            uuid.UUID         `json:"org_id"`
	FirstName        string            `json:"first_name"`
	LastName         string            `json:"last_name"`
	Email            string            `json:"email"`
	Phone            string            `json:"phone"`
	LicenseNumber    string            `json:"license_number"`
	LicenseType      string            `json:"license_type"`
	LicenseExpiry    time.Time         `json:"license_expiry"`
	Status           string            `json:"status"`
	CurrentVehicleID *uuid.UUID        `json:"current_vehicle_id,omitempty"`
	Rating           float64           `json:"rating"`
	BehaviorScore    int               `json:"behavior_score"`
	Address          *Address          `json:"address,omitempty"`
	EmergencyContact *EmergencyContact `json:"emergency_contact,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	CreatedBy        uuid.UUID         `json:"created_by"`
	UpdatedAt        time.Time         `json:"updated_at"`
	UpdatedBy        uuid.UUID         `json:"updated_by"`
}

type WorkOrder struct {
	ID              uuid.UUID  `json:"id"`
	OrgID           uuid.UUID  `json:"org_id"`
	VehicleID       uuid.UUID  `json:"vehicle_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Type            string     `json:"type"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty"`
	CompletedDate   *time.Time `json:"completed_date,omitempty"`
	EstimatedCost   float64    `json:"estimated_cost"`
	ActualCost      float64    `json:"actual_cost"`
	VendorID        *uuid.UUID `json:"vendor_id,omitempty"`
	ApprovalStatus  string     `json:"approval_status"`
	CreatedAt       time.Time  `json:"created_at"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UpdatedBy       uuid.UUID  `json:"updated_by"`
}

type Inspection struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	VehicleID   uuid.UUID `json:"vehicle_id"`
	InspectorID uuid.UUID `json:"inspector_id"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
	Odometer    int64     `json:"odometer"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

type FuelLog struct {
	ID              uuid.UUID  `json:"id"`
	OrgID           uuid.UUID  `json:"org_id"`
	VehicleID       uuid.UUID  `json:"vehicle_id"`
	DriverID        uuid.UUID  `json:"driver_id"`
	StationName     string     `json:"station_name"`
	FuelType        string     `json:"fuel_type"`
	Quantity        float64    `json:"quantity"`
	UnitPrice       float64    `json:"unit_price"`
	TotalCost       float64    `json:"total_cost"`
	Odometer        float64    `json:"odometer"`
	TransactionDate time.Time  `json:"transaction_date"`
	Status          string     `json:"status"`
	ReceiptURL      *string    `json:"receipt_url,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UpdatedBy       uuid.UUID  `json:"updated_by"`
}

type Telemetry struct {
	VehicleID      uuid.UUID         `json:"vehicle_id"`
	OrgID          uuid.UUID         `json:"org_id"`
	DeviceID       string            `json:"device_id"`
	Latitude       float64           `json:"latitude"`
	Longitude      float64           `json:"longitude"`
	Speed          float64           `json:"speed"`
	FuelLevel      float64           `json:"fuel_level"`
	BatteryVoltage float64           `json:"battery_voltage"`
	CoolantTemp    float64           `json:"coolant_temp"`
	Odometer       int64             `json:"odometer"`
	Timestamp      time.Time         `json:"timestamp"`
	Metadata       map[string]string `json:"metadata"`
}
