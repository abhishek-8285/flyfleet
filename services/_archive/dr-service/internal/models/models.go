package models

import (
	"time"

	"github.com/google/uuid"
)

type Backup struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	BackupType      string     `json:"backup_type" db:"backup_type"` // FULL, INCREMENTAL
	Status          string     `json:"status" db:"status"`           // SCHEDULED, RUNNING, COMPLETED, FAILED, VERIFIED
	StartTime       *time.Time `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time" db:"end_time"`
	DurationSeconds *int       `json:"duration_seconds" db:"duration_seconds"`
	SizeBytes       *int64     `json:"size_bytes" db:"size_bytes"`
	Location        *string    `json:"location" db:"location"`
	RetentionDays   int        `json:"retention_days" db:"retention_days"`
	Checksum        *string    `json:"checksum" db:"checksum"`
	InitiatedBy     *string    `json:"initiated_by" db:"initiated_by"`
	Remarks         *string    `json:"remarks" db:"remarks"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type BackupLog struct {
	ID        int       `json:"id" db:"id"`
	BackupID  uuid.UUID `json:"backup_id" db:"backup_id"`
	Action    string    `json:"action" db:"action"`
	Result    string    `json:"result" db:"result"` // SUCCESS, FAILURE
	Actor     *string   `json:"actor" db:"actor"`
	Message   *string   `json:"message" db:"message"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type StoragePool struct {
	ID                int       `json:"id" db:"id"`
	Provider          string    `json:"provider" db:"provider"` // AWS, AZURE, ONPREM
	Bucket            string    `json:"bucket" db:"bucket"`
	Region            *string   `json:"region" db:"region"`
	Tier              string    `json:"tier" db:"tier"`
	EncryptionEnabled bool      `json:"encryption_enabled" db:"encryption_enabled"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type CreateBackupRequest struct {
	BackupType    string `json:"backup_type" binding:"required,oneof=FULL INCREMENTAL"`
	Location      string `json:"location" binding:"required"`
	RetentionDays int    `json:"retention_days"`
	Remarks       string `json:"remarks"`
	InitiatedBy   string `json:"initiated_by"`
}

type Restore struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	BackupID        uuid.UUID  `json:"backup_id" db:"backup_id"`
	RestoreType     string     `json:"restore_type" db:"restore_type"` // FULL, INCREMENTAL
	TargetEnv       string     `json:"target_env" db:"target_env"`     // PROD, STAGING, UAT
	Status          string     `json:"status" db:"status"`             // PENDING, RUNNING, COMPLETED, FAILED, CANCELLED
	StartTime       *time.Time `json:"start_time" db:"started_at"`
	CompletedTime   *time.Time `json:"completed_at" db:"completed_at"`
	InitiatedBy     *string    `json:"initiated_by" db:"initiated_by"`
	DurationSeconds *int       `json:"duration_seconds" db:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type RestoreLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RestoreID uuid.UUID `json:"restore_id" db:"restore_id"`
	Level     string    `json:"level" db:"level"` // INFO, WARN, ERROR
	Module    string    `json:"module" db:"module"`
	Message   *string   `json:"message" db:"message"`
	CreatedAt time.Time `json:"timestamp" db:"created_at"`
}

type CreateRestoreRequest struct {
	BackupID       string `json:"backup_id" binding:"required"`
	TargetEnv      string `json:"target_env" binding:"required,oneof=PROD STAGING UAT"`
	IntegrityCheck bool   `json:"integrity_check"`
	RequestedBy    string `json:"requested_by"`
}

// ========== DR Testing Models ==========

type DRTest struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TestID          string     `json:"test_id" db:"test_id"`
	Name            string     `json:"name" db:"name"`
	Description     *string    `json:"description" db:"description"`
	TestType        string     `json:"test_type" db:"test_type"`   // FAILOVER, BACKUP, RESTORE
	TargetEnv       string     `json:"target_env" db:"target_env"` // Environment name
	Status          string     `json:"status" db:"status"`         // SCHEDULED, RUNNING, PASSED, FAILED, ABORTED
	ProgressPercent int        `json:"progress_percent" db:"progress_percent"`
	InitiatedBy     *string    `json:"initiated_by" db:"initiated_by"`
	SuccessCriteria *string    `json:"success_criteria" db:"success_criteria"`
	ScheduledAt     *time.Time `json:"scheduled_at" db:"scheduled_at"`
	StartedAt       *time.Time `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type DRTestStep struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TestID      uuid.UUID  `json:"test_id" db:"test_id"`
	StepOrder   int        `json:"step_order" db:"step_order"`
	StepName    string     `json:"step_name" db:"step_name"`
	Status      string     `json:"status" db:"status"` // PENDING, RUNNING, PASSED, FAILED, SKIPPED
	StartedAt   *time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type DRTestLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TestID    uuid.UUID `json:"test_id" db:"test_id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Level     string    `json:"level" db:"level"` // INFO, WARN, ERROR, RUN
	Module    *string   `json:"module" db:"module"`
	Message   string    `json:"message" db:"message"`
}

type DRTestSettings struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	OrgID              *string   `json:"org_id" db:"org_id"`
	ScheduleEnabled    bool      `json:"schedule_enabled" db:"schedule_enabled"`
	Frequency          string    `json:"frequency" db:"frequency"`       // WEEKLY, MONTHLY
	ScheduleDay        string    `json:"schedule_day" db:"schedule_day"` // SATURDAY, etc.
	ScheduleTime       string    `json:"schedule_time" db:"schedule_time"`
	Intensity          string    `json:"intensity" db:"intensity"` // NORMAL, HIGH
	NotificationEmails []string  `json:"notification_emails" db:"-"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type CreateDRTestRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	TestType        string `json:"test_type" binding:"required,oneof=FAILOVER BACKUP RESTORE"`
	TargetEnv       string `json:"target_env" binding:"required"`
	SuccessCriteria string `json:"success_criteria"`
	ScheduledAt     string `json:"scheduled_at"` // ISO timestamp or empty for immediate
	InitiatedBy     string `json:"initiated_by"`
}

type UpdateDRTestSettingsRequest struct {
	ScheduleEnabled    bool     `json:"schedule_enabled"`
	Frequency          string   `json:"frequency"`
	ScheduleDay        string   `json:"schedule_day"`
	ScheduleTime       string   `json:"schedule_time"`
	Intensity          string   `json:"intensity"`
	NotificationEmails []string `json:"notification_emails"`
}

type DRTestMetrics struct {
	TotalTests      int     `json:"total_tests"`
	SuccessRate     float64 `json:"success_rate"`
	AvgRecoveryTime string  `json:"avg_recovery_time"`
	NextScheduled   string  `json:"next_scheduled"`
}
