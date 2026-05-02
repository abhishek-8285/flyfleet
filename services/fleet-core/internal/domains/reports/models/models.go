package models

import (
	"time"

	"github.com/google/uuid"
)

type ScheduledReport struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	TenantID uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	Name           string         `json:"name" db:"name"`
	ReportType     string         `json:"report_type" db:"report_type"`
	Format         string         `json:"format" db:"format"`
	Frequency      string         `json:"frequency" db:"frequency"`
	CronExpression string         `json:"cron_expression" db:"cron_expression"`
	DeliveryTime   string         `json:"delivery_time" db:"delivery_time"`
	Timezone       string         `json:"timezone" db:"timezone"`
	Filters        map[string]any `json:"filters" db:"filters"`
	Status         string         `json:"status" db:"status"`
	CreatedBy      uuid.UUID      `json:"created_by" db:"created_by"`
	UpdatedBy      uuid.UUID      `json:"updated_by" db:"updated_by"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`

	Recipients []string `json:"recipients,omitempty"` // Joined from report_recipients
}

type ReportExecution struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ReportID        uuid.UUID  `json:"report_id" db:"report_id"`
	StartedAt       *time.Time `json:"started_at" db:"started_at"`
	FinishedAt      *time.Time `json:"finished_at" db:"finished_at"`
	DurationSeconds *int       `json:"duration_seconds" db:"duration_seconds"`
	Status          string     `json:"status" db:"status"`
	ErrorMessage    *string    `json:"error_message" db:"error_message"`
	FileURL         *string    `json:"file_url" db:"file_url"`
	FileSize        *int64     `json:"file_size" db:"file_size"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy       uuid.UUID  `json:"updated_by" db:"updated_by"`
}

type GlobalSettings struct {
	ID                int       `json:"id" db:"id"`
	MaxConcurrentJobs int       `json:"max_concurrent_jobs" db:"max_concurrent_jobs"`
	MaxRetryAttempts  int       `json:"max_retry_attempts" db:"max_retry_attempts"`
	RetentionDays     int       `json:"retention_days" db:"retention_days"`
	SMTPHost          *string   `json:"smtp_host" db:"smtp_host"`
	SMTPPort          *int      `json:"smtp_port" db:"smtp_port"`
	SMTPUser          *string   `json:"smtp_user" db:"smtp_user"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by" db:"updated_by"`
}
