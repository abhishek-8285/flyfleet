package models

import "time"

// DeviceRegistry holds the IoT device control-plane record.
// This drives firmware OTA tracking, SIM failover detection, and health monitoring.
type DeviceRegistry struct {
	DeviceID        string    `json:"device_id"`
	VehicleID       string    `json:"vehicle_id"`
	FirmwareVersion string    `json:"firmware_version"`
	SimICCID        string    `json:"sim_iccid"`
	Carrier         string    `json:"carrier"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
	Status          string    `json:"status"` // active | offline | decommissioned
	CertificateHash string    `json:"certificate_hash"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
