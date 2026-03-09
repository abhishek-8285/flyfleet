package models

import (
	"time"

	"github.com/google/uuid"
)

type ComplianceDashboardMetrics struct {
	OrganizationID         uuid.UUID `json:"organization_id" db:"organization_id"`
	DriversActive          int       `json:"drivers_active" db:"drivers_active"`
	DriversOverLimit       int       `json:"drivers_over_limit" db:"drivers_over_limit"`
	VehiclesWithExpiredDocs int       `json:"vehicles_with_expired_docs" db:"vehicles_with_expired_docs"`
	PanicAlertsToday       int       `json:"panic_alerts_today" db:"panic_alerts_today"`
	OfflineDevices         int       `json:"offline_devices" db:"offline_devices"`
	WeeklyViolations       int       `json:"weekly_violations" db:"weekly_violations"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}
