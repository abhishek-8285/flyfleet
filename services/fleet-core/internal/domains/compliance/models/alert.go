package models

import (
	"time"

	"github.com/google/uuid"
)

type AlertType string

const (
	AlertPanic            AlertType = "PANIC_ALERT"
	AlertDeviceTamper     AlertType = "DEVICE_TAMPER"
	AlertDeviceOffline    AlertType = "DEVICE_OFFLINE"
	AlertDriverViolation  AlertType = "DRIVER_VIOLATION"
	AlertDocumentExpiring AlertType = "DOCUMENT_EXPIRING"
	AlertDocumentExpired  AlertType = "DOCUMENT_EXPIRED"
)

type EntityType string

const (
	EntityVehicle  EntityType = "VEHICLE"
	EntityDriver   EntityType = "DRIVER"
	EntityDevice   EntityType = "DEVICE"
	EntityDocument EntityType = "DOCUMENT"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "LOW"
	AlertSeverityMedium   AlertSeverity = "MEDIUM"
	AlertSeverityHigh     AlertSeverity = "HIGH"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

type ComplianceAlert struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	OrganizationID uuid.UUID     `json:"organization_id" db:"organization_id"`
	Type           AlertType     `json:"alert_type" db:"alert_type"`
	EntityType     EntityType    `json:"entity_type" db:"entity_type"`
	EntityID       uuid.UUID     `json:"entity_id" db:"entity_id"`
	Severity       AlertSeverity `json:"severity" db:"severity"`
	Metadata       []byte        `json:"metadata" db:"metadata"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	ResolvedAt     *time.Time    `json:"resolved_at" db:"resolved_at"`
}
