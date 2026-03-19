package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceStatus string

const (
	DeviceStatusActive         DeviceStatus = "ACTIVE"
	DeviceStatusOffline        DeviceStatus = "OFFLINE"
	DeviceStatusTampered       DeviceStatus = "TAMPERED"
	DeviceStatusDecommissioned DeviceStatus = "DECOMMISSIONED"
)

type AIS140Device struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	VehicleID       uuid.UUID    `json:"vehicle_id" db:"vehicle_id"`
	DeviceIMEI      string       `json:"device_imei" db:"device_imei"`
	SIMNumber       string       `json:"sim_number" db:"sim_number"`
	DeviceModel     string       `json:"device_model" db:"device_model"`
	FirmwareVersion string       `json:"firmware_version" db:"firmware_version"`
	Status          DeviceStatus `json:"status" db:"status"`
	LastSeen        *time.Time   `json:"last_seen" db:"last_seen"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

type DeviceHealthLog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	DeviceID       uuid.UUID  `json:"device_id" db:"device_id"`
	Timestamp      time.Time  `json:"timestamp" db:"timestamp"`
	BatteryVoltage float64    `json:"battery_voltage" db:"battery_voltage"`
	GPSSignal      string     `json:"gps_signal" db:"gps_signal"`
	NetworkSignal  string     `json:"network_signal" db:"network_signal"`
	Temperature    float64    `json:"temperature" db:"temperature"`
	Status         string     `json:"status" db:"status"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type PanicAlertStatus string

const (
	PanicAlertActive     PanicAlertStatus = "ACTIVE"
	PanicAlertInProgress PanicAlertStatus = "IN_PROGRESS"
	PanicAlertResolved   PanicAlertStatus = "RESOLVED"
)

type PanicAlert struct {
	ID         uuid.UUID        `json:"id" db:"id"`
	VehicleID  uuid.UUID        `json:"vehicle_id" db:"vehicle_id"`
	DriverID   *uuid.UUID       `json:"driver_id" db:"driver_id"`
	Timestamp  time.Time        `json:"timestamp" db:"timestamp"`
	Latitude   float64          `json:"latitude" db:"latitude"`
	Longitude  float64          `json:"longitude" db:"longitude"`
	Status     PanicAlertStatus `json:"status" db:"status"`
	ResolvedAt *time.Time       `json:"resolved_at" db:"resolved_at"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
}

type PanicAlertAction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	AlertID   uuid.UUID `json:"alert_id" db:"alert_id"`
	Action    string    `json:"action" db:"action"`
	Actor     uuid.UUID `json:"actor" db:"actor"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
