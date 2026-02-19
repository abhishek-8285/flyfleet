package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// BackupStatus represents the status of a backup operation
type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "PENDING"
	BackupStatusInProgress BackupStatus = "IN_PROGRESS"
	BackupStatusCompleted  BackupStatus = "COMPLETED"
	BackupStatusFailed     BackupStatus = "FAILED"
)

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull        BackupType = "FULL"
	BackupTypeIncremental BackupType = "INCREMENTAL"
	BackupTypeSnapshot    BackupType = "SNAPSHOT"
)

// RestoreStatus represents the status of a restore operation
type RestoreStatus string

const (
	RestoreStatusPending    RestoreStatus = "PENDING"
	RestoreStatusInProgress RestoreStatus = "IN_PROGRESS"
	RestoreStatusCompleted  RestoreStatus = "COMPLETED"
	RestoreStatusFailed     RestoreStatus = "FAILED"
)

// JSONB represents a JSONB column in the database
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}

// Backup represents a system backup record
type Backup struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Type        BackupType   `json:"type" db:"type"`
	Status      BackupStatus `json:"status" db:"status"`
	S3Key       string       `json:"s3_key" db:"s3_key"`
	SizeBytes   int64        `json:"size_bytes" db:"size_bytes"`
	CreatedBy   uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	Metadata    JSONB        `json:"metadata,omitempty" db:"metadata"`
}

// RestoreJob represents a system restore operation
type RestoreJob struct {
	ID                uuid.UUID     `json:"id" db:"id"`
	BackupID          uuid.UUID     `json:"backup_id" db:"backup_id"`
	Status            RestoreStatus `json:"status" db:"status"`
	TargetEnvironment string        `json:"target_environment" db:"target_environment"`
	StartedAt         time.Time     `json:"started_at" db:"started_at"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
	PerformedBy       uuid.UUID     `json:"performed_by" db:"performed_by"`
	Logs              string        `json:"logs,omitempty" db:"logs"`
}

// StoragePoolStatus represents the status of a storage pool
type StoragePoolStatus string

const (
	StoragePoolStatusActive  StoragePoolStatus = "ACTIVE"
	StoragePoolStatusWarning StoragePoolStatus = "WARNING"
	StoragePoolStatusFull    StoragePoolStatus = "FULL"
	StoragePoolStatusOffline StoragePoolStatus = "OFFLINE"
)

// StoragePool represents a storage pool for backups
type StoragePool struct {
	ID             uuid.UUID         `json:"id" db:"id"`
	Name           string            `json:"name" db:"name"`
	Status         StoragePoolStatus `json:"status" db:"status"`
	TotalBytes     int64             `json:"total_bytes" db:"total_bytes"`
	UsedBytes      int64             `json:"used_bytes" db:"used_bytes"`
	AvailableBytes int64             `json:"available_bytes" db:"available_bytes"`
	Location       string            `json:"location" db:"location"`
	BackupCount    int               `json:"backup_count" db:"backup_count"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
}

// BackupVerification represents the result of a backup verification
type BackupVerification struct {
	ID           uuid.UUID `json:"id" db:"id"`
	BackupID     uuid.UUID `json:"backup_id" db:"backup_id"`
	Status       string    `json:"status" db:"status"`
	Checksum     string    `json:"checksum" db:"checksum"`
	VerifiedAt   time.Time `json:"verified_at" db:"verified_at"`
	VerifiedBy   uuid.UUID `json:"verified_by" db:"verified_by"`
	ErrorMessage string    `json:"error_message,omitempty" db:"error_message"`
}

// DRTest represents a disaster-recovery test execution.
type DRTest struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TestID          string     `json:"test_id" db:"test_id"`
	Name            string     `json:"name" db:"name"`
	Description     string     `json:"description" db:"description"`
	TestType        string     `json:"test_type" db:"test_type"`
	TargetEnv       string     `json:"target_env" db:"target_env"`
	Status          string     `json:"status" db:"status"`
	ProgressPercent int        `json:"progress_percent" db:"progress_percent"`
	InitiatedBy     string     `json:"initiated_by" db:"initiated_by"`
	SuccessCriteria string     `json:"success_criteria" db:"success_criteria"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt       *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type DRTestStep struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TestID      uuid.UUID  `json:"test_id" db:"test_id"`
	StepOrder   int        `json:"step_order" db:"step_order"`
	StepName    string     `json:"step_name" db:"step_name"`
	Status      string     `json:"status" db:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

type DRTestLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TestID    uuid.UUID `json:"test_id" db:"test_id"`
	Timestamp time.Time `json:"timestamp" db:"ts"`
	Level     string    `json:"level" db:"level"`
	Message   string    `json:"message" db:"message"`
}

type DRTestMetrics struct {
	TotalTests      int        `json:"total_tests"`
	SuccessRate     float64    `json:"success_rate"`
	AvgRecoveryTime string     `json:"avg_recovery_time"`
	NextScheduled   *time.Time `json:"next_scheduled,omitempty"`
}

type DRTestSettings struct {
	ScheduleEnabled    bool     `json:"schedule_enabled"`
	Frequency          string   `json:"frequency"`
	ScheduleDay        string   `json:"schedule_day"`
	ScheduleTime       string   `json:"schedule_time"`
	Intensity          string   `json:"intensity"`
	NotificationEmails []string `json:"notification_emails"`
}
