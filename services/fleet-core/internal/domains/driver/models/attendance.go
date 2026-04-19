package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverAttendance represents a daily attendance record for a driver
type DriverAttendance struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	DriverID       uuid.UUID  `json:"driver_id" db:"driver_id"`
	OrgID          string     `json:"tenant_id" db:"tenant_id"`
	AttendanceDate time.Time  `json:"attendance_date" db:"attendance_date"`
	Status         string     `json:"status" db:"status"` // present, absent, late, half_day
	CheckInTime    *time.Time `json:"check_in_time" db:"check_in_time"`
	CheckOutTime   *time.Time `json:"check_out_time" db:"check_out_time"`
	Remarks        string     `json:"remarks" db:"remarks"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// DriverSchedule represents a recurring weekly schedule for a driver
type DriverSchedule struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DriverID  uuid.UUID `json:"driver_id" db:"driver_id"`
	OrgID     string    `json:"tenant_id" db:"tenant_id"`
	DayOfWeek int       `json:"day_of_week" db:"day_of_week"` // 0-6 (Sunday-Saturday)
	StartTime string    `json:"start_time" db:"start_time"`   // HH:MM format
	EndTime   string    `json:"end_time" db:"end_time"`       // HH:MM format
	IsWorking bool      `json:"is_working" db:"is_working"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
