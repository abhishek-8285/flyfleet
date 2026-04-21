package models

import "time"

// VehicleAlert represents an alert generated for a vehicle
type VehicleAlert struct {
	ID            string                 `json:"id" db:"id"`
	VehicleID     string                 `json:"vehicle_id" db:"vehicle_id"`
	TenantID         string                 `json:"tenant_id" db:"tenant_id"`
	Type          string                 `json:"alert_type" db:"alert_type"`
	Severity      string                 `json:"severity" db:"severity"`
	Description   string                 `json:"description" db:"description"`
	Latitude      float64                `json:"latitude" db:"latitude"`
	Longitude     float64                `json:"longitude" db:"longitude"`
	Timestamp     time.Time              `json:"alert_timestamp" db:"alert_timestamp"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	Processed     bool                   `json:"processed" db:"processed"`
	ProcessedAt   *time.Time             `json:"processed_at" db:"processed_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}
