package models

import (
	"time"

	"github.com/google/uuid"
)

type DriverDutyStatus string

const (
	DutyStatusOffDuty DriverDutyStatus = "OFF_DUTY"
	DutyStatusDriving DriverDutyStatus = "DRIVING"
	DutyStatusOnDuty  DriverDutyStatus = "ON_DUTY"
	DutyStatusBreak   DriverDutyStatus = "BREAK"
	DutyStatusRest    DriverDutyStatus = "REST"
)

type BreakType string

const (
	BreakTypeRest           BreakType = "REST"
	BreakTypeMeal           BreakType = "MEAL"
	BreakTypeMandatoryBreak BreakType = "MANDATORY_BREAK"
)

// DriverDutyLog replaces the legacy FMCSA duty_status_logs
type DriverDutyLog struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	DriverID  uuid.UUID        `json:"driver_id" db:"driver_id"`
	VehicleID *uuid.UUID       `json:"vehicle_id" db:"vehicle_id"`
	Status    DriverDutyStatus `json:"status" db:"status"`
	StartTime time.Time        `json:"start_time" db:"start_time"`
	EndTime   *time.Time       `json:"end_time" db:"end_time"`
	Location  string           `json:"location" db:"location"`
	Notes     string           `json:"notes" db:"notes"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

// DriverBreakLog tracks mandatory resting periods
type DriverBreakLog struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	DriverID        uuid.UUID  `json:"driver_id" db:"driver_id"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time" db:"end_time"`
	DurationMinutes int        `json:"duration_minutes" db:"duration_minutes"`
	BreakType       BreakType  `json:"break_type" db:"break_type"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}
