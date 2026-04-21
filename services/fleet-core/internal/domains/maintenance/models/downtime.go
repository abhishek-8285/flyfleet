package models

import (
	"time"

	"github.com/google/uuid"
)

type VehicleDowntime struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	VehicleID       uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	WorkOrderID     *uuid.UUID `json:"work_order_id,omitempty" db:"work_order_id"`
	DowntimeStart   time.Time  `json:"downtime_start" db:"downtime_start"`
	DowntimeEnd     *time.Time `json:"downtime_end,omitempty" db:"downtime_end"`
	DowntimeMinutes int        `json:"downtime_minutes" db:"downtime_minutes"`
	Reason          string     `json:"reason" db:"reason"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type DowntimeMetrics struct {
	VehicleID        uuid.UUID `json:"vehicle_id"`
	MTTR             float64   `json:"mtr_minutes"` // Mean Time To Repair
	MTBF             float64   `json:"mtbf_days"`   // Mean Time Between Failures
	TotalDowntime    int       `json:"total_downtime_minutes"`
	MaintenanceCount int       `json:"maintenance_count"`
}
