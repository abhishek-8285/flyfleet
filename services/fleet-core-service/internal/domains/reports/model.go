package reports

import (
	"time"

	"github.com/google/uuid"
)

type ReportType string

const (
	ReportTypeVehicleUsage ReportType = "VEHICLE_USAGE"
	ReportTypeDriverSafety ReportType = "DRIVER_SAFETY"
	ReportTypeFuelEfficiency ReportType = "FUEL_EFFICIENCY"
	ReportTypeMaintenanceCost ReportType = "MAINTENANCE_COST"
)

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "PENDING"
	ReportStatusProcessing ReportStatus = "PROCESSING"
	ReportStatusCompleted  ReportStatus = "COMPLETED"
	ReportStatusFailed     ReportStatus = "FAILED"
)

type Report struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	Type      ReportType   `json:"type" db:"type"`
	Status    ReportStatus `json:"status" db:"status"`
	Parameters map[string]interface{} `json:"parameters" db:"parameters"`
	FileURL   string       `json:"file_url,omitempty" db:"file_url"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

type CreateReportRequest struct {
	Type       ReportType             `json:"type" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}
