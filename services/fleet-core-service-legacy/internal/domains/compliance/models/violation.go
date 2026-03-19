package models

import (
	"time"

	"github.com/google/uuid"
)

type ViolationType string

const (
	ViolationDrivingLimitExceeded ViolationType = "DRIVING_LIMIT_EXCEEDED"
	ViolationDutyLimitExceeded    ViolationType = "DUTY_LIMIT_EXCEEDED"
	ViolationBreakRequired        ViolationType = "BREAK_REQUIRED"
	ViolationWeeklyLimitExceeded  ViolationType = "WEEKLY_LIMIT_EXCEEDED"
)

type ValidationSeverity string

const (
	SeverityLow      ValidationSeverity = "LOW"
	SeverityMedium   ValidationSeverity = "MEDIUM"
	SeverityHigh     ValidationSeverity = "HIGH"
	SeverityCritical ValidationSeverity = "CRITICAL"
)

type DriverComplianceViolation struct {
	ID            uuid.UUID          `json:"id" db:"id"`
	DriverID      uuid.UUID          `json:"driver_id" db:"driver_id"`
	VehicleID     *uuid.UUID         `json:"vehicle_id" db:"vehicle_id"`
	ViolationType ViolationType      `json:"violation_type" db:"violation_type"`
	Timestamp     time.Time          `json:"timestamp" db:"timestamp"`
	Location      string             `json:"location" db:"location"`
	Severity      ValidationSeverity `json:"severity" db:"severity"`
	Resolved      bool               `json:"resolved" db:"resolved"`
	Escalated     bool               `json:"escalated" db:"escalated"`
	Metadata      []byte             `json:"metadata" db:"metadata"`
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
}
