package models

import (
	"time"

	"github.com/google/uuid"
)

// MaintenanceSchedule holds the next service forecast for a vehicle.
type MaintenanceSchedule struct {
	VehicleID           uuid.UUID  `json:"vehicle_id"             db:"vehicle_id"`
	NextServiceDate     *time.Time `json:"next_service_date"      db:"next_service_date"`
	NextServiceOdometer *float64   `json:"next_service_odometer"  db:"next_service_odometer"`
	EngineHours         *float64   `json:"engine_hours"           db:"engine_hours"`
	PredictiveScore     *float64   `json:"predictive_score"       db:"predictive_score"` // 0-100
	LastServiceDate     *time.Time `json:"last_service_date"      db:"last_service_date"`
	CreatedAt           time.Time  `json:"created_at"             db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"             db:"updated_at"`
}

// UpsertScheduleRequest is the request body for updating maintenance schedule.
type UpsertScheduleRequest struct {
	NextServiceDate     *time.Time `json:"next_service_date"`
	NextServiceOdometer *float64   `json:"next_service_odometer"`
	EngineHours         *float64   `json:"engine_hours"`
	PredictiveScore     *float64   `json:"predictive_score" validate:"omitempty,min=0,max=100"`
	LastServiceDate     *time.Time `json:"last_service_date"`
}
