package reports

import (
	"time"

	"github.com/google/uuid"
)

type ReportType string

const (
	ReportTypeVehicleUtilization ReportType = "VEHICLE_UTILIZATION"
	ReportTypeTripSummary        ReportType = "TRIP_SUMMARY"
	ReportTypeIdleTime           ReportType = "IDLE_TIME"
	ReportTypeDriverActivity     ReportType = "DRIVER_ACTIVITY"
	ReportTypeFuelCost           ReportType = "FUEL_COST"
	ReportTypeTripProfitability  ReportType = "TRIP_PROFITABILITY"
	ReportTypeInvoiceSummary     ReportType = "INVOICE_SUMMARY"
	ReportTypeExpense            ReportType = "EXPENSE_REPORT"
	ReportTypeHOSViolation       ReportType = "HOS_VIOLATION"
	ReportTypeDriverDutyLogs     ReportType = "DRIVER_DUTY_LOGS"
	ReportTypeELDExport          ReportType = "ELD_EXPORT"
	ReportTypeDriverPerformance  ReportType = "DRIVER_PERFORMANCE"
	ReportTypeRouteEfficiency    ReportType = "ROUTE_EFFICIENCY"
	ReportTypeFuelAnomaly        ReportType = "FUEL_ANOMALY"
	ReportTypeVehicleDowntime    ReportType = "VEHICLE_DOWNTIME"
)

type ReportStatus string

const (
	ReportStatusCreated       ReportStatus = "CREATED"
	ReportStatusQueued        ReportStatus = "QUEUED"
	ReportStatusRunning       ReportStatus = "RUNNING"
	ReportStatusQuerying      ReportStatus = "QUERYING"        // executing aggregation queries
	ReportStatusStreamingRows ReportStatus = "STREAMING_ROWS"  // cursor fetching + row streaming
	ReportStatusGenerating    ReportStatus = "GENERATING"      // writing CSV output
	ReportStatusCompressing   ReportStatus = "COMPRESSING"     // gzip compression
	ReportStatusUploading     ReportStatus = "UPLOADING"       // uploading to object storage
	ReportStatusDelivering    ReportStatus = "DELIVERING"      // sending email/webhook
	ReportStatusCompleted     ReportStatus = "COMPLETED"
	ReportStatusFailed        ReportStatus = "FAILED"
	ReportStatusCancelled     ReportStatus = "CANCELLED"
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "QUEUED"
	JobStatusRunning    JobStatus = "RUNNING"
	JobStatusGenerating JobStatus = "GENERATING"
	JobStatusUploading  JobStatus = "UPLOADING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
)

type Report struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	TenantID     uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Type         ReportType             `json:"report_type" db:"report_type"`
	Status       ReportStatus           `json:"status" db:"status"`
	Parameters   map[string]interface{} `json:"parameters" db:"parameters"`
	RequestedBy  uuid.UUID              `json:"requested_by" db:"requested_by"`
	RequestedAt  time.Time              `json:"requested_at" db:"requested_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Progress     int                    `json:"progress" db:"progress"`
	FileURL      string                 `json:"file_url,omitempty" db:"file_url"`
	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	Checksum     string                 `json:"checksum,omitempty" db:"checksum"`
}

type ReportJob struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ReportID  uuid.UUID `json:"report_id" db:"report_id"`
	JobStatus JobStatus `json:"job_status" db:"job_status"`
	WorkerID  string    `json:"worker_id,omitempty" db:"worker_id"`
	Attempts  int       `json:"attempts" db:"attempts"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ReportTemplate struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	QueryTemplate  string                 `json:"query_template" db:"query_template"`
	AllowedFilters map[string]interface{} `json:"allowed_filters" db:"allowed_filters"`
	OutputFormat   string                 `json:"output_format" db:"output_format"`
	Version        int                    `json:"version" db:"version"`
}

type ScheduledReport struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	TenantID         uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	ReportTemplateID uuid.UUID              `json:"report_template_id" db:"report_template_id"`
	CronExpression   string                 `json:"cron_expression" db:"cron_expression"`
	Parameters       map[string]interface{} `json:"parameters" db:"parameters"`
	Enabled          bool                   `json:"enabled" db:"enabled"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

type ReportAuditLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ReportID  uuid.UUID `json:"report_id,omitempty" db:"report_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	IPAddress string    `json:"ip_address,omitempty" db:"ip_address"`
}

// API Models

type CreateReportRequest struct {
	ReportType ReportType             `json:"report_type" validate:"required"`
	Format     string                 `json:"format" validate:"required"`
	From       string                 `json:"from" validate:"required"`
	To         string                 `json:"to" validate:"required"`
	Filters    map[string]interface{} `json:"filters"`
}

type CreateReportResponse struct {
	ReportID uuid.UUID    `json:"report_id"`
	Status   ReportStatus `json:"status"`
}

type ReportStatusResponse struct {
	Status   ReportStatus `json:"status"`
	Progress int          `json:"progress"`
}
