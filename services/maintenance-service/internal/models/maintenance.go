package models

import (
	"time"

	"github.com/google/uuid"
)

// MaintenanceSchedule represents a scheduled maintenance task
type MaintenanceSchedule struct {
	ID              string            `json:"id" db:"id"`
	OrgID           uuid.UUID         `json:"org_id" db:"org_id"`
	VehicleID       uuid.UUID         `json:"vehicle_id" db:"vehicle_id"`
	Type            string            `json:"type" db:"type"` // routine, preventive, corrective, emergency
	Name            string            `json:"name" db:"name"`
	Description     string            `json:"description" db:"description"`
	DueDate         time.Time         `json:"due_date" db:"due_date"`
	DueOdometer     *int              `json:"due_odometer,omitempty" db:"due_odometer"`
	DueEngineHours  *float64          `json:"due_engine_hours,omitempty" db:"due_engine_hours"`
	AssignedTo      *uuid.UUID        `json:"assigned_to,omitempty" db:"assigned_to"` // mechanic ID
	Priority        string            `json:"priority" db:"priority"` // low, medium, high, critical
	Status          string            `json:"status" db:"status"`     // scheduled, in_progress, completed, cancelled
	EstimatedCost   *float64          `json:"estimated_cost,omitempty" db:"estimated_cost"`
	ActualCost      *float64          `json:"actual_cost,omitempty" db:"actual_cost"`
	PartsRequired   []string          `json:"parts_required,omitempty" db:"parts_required"`
	ServiceItems    []ServiceItem     `json:"service_items,omitempty" db:"service_items"`
	Notes           string            `json:"notes" db:"notes"`
	Reminders       []Reminder        `json:"reminders,omitempty" db:"reminders"`
	Recurring       *RecurringPattern `json:"recurring,omitempty" db:"recurring"`
	CreatedBy       uuid.UUID         `json:"created_by" db:"created_by"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
	UpdatedBy       uuid.UUID         `json:"updated_by" db:"updated_by"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
	CompletedBy     *uuid.UUID        `json:"completed_by,omitempty" db:"completed_by"`
}

// MaintenanceLog represents a completed maintenance record
type MaintenanceLog struct {
	ID              string        `json:"id" db:"id"`
	OrgID           uuid.UUID     `json:"org_id" db:"org_id"`
	VehicleID       uuid.UUID     `json:"vehicle_id" db:"vehicle_id"`
	ScheduleID      *string       `json:"schedule_id,omitempty" db:"schedule_id"`
	Type            string        `json:"type" db:"type"` // routine, preventive, corrective, emergency
	Name            string        `json:"name" db:"name"`
	Description     string        `json:"description" db:"description"`
	StartDate       time.Time     `json:"start_date" db:"start_date"`
	EndDate         time.Time     `json:"end_date" db:"end_date"`
	DurationMinutes int           `json:"duration_minutes" db:"duration_minutes"`
	PerformedBy     uuid.UUID     `json:"performed_by" db:"performed_by"` // mechanic ID
	Cost            float64       `json:"cost" db:"cost"`
	Odometer        int           `json:"odometer" db:"odometer"`
	EngineHours     *float64      `json:"engine_hours,omitempty" db:"engine_hours"`
	PartsUsed       []PartUsed    `json:"parts_used,omitempty" db:"parts_used"`
	ServiceItems    []ServiceItem `json:"service_items,omitempty" db:"service_items"`
	Notes           string        `json:"notes" db:"notes"`
	Documents       []string      `json:"documents,omitempty" db:"documents"` // file URLs
	Status          string        `json:"status" db:"status"`                 // completed, partially_completed
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	CreatedBy       uuid.UUID     `json:"created_by" db:"created_by"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	UpdatedBy       uuid.UUID     `json:"updated_by" db:"updated_by"`
}

// ServiceItem represents a specific maintenance task within a schedule/log
type ServiceItem struct {
	ID          string  `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	Completed   bool    `json:"completed" db:"completed"`
	Notes       string  `json:"notes" db:"notes"`
	Cost        float64 `json:"cost" db:"cost"`
}

// PartUsed represents a part used in maintenance
type PartUsed struct {
	ID          string  `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	PartNumber  string  `json:"part_number" db:"part_number"`
	Quantity    int     `json:"quantity" db:"quantity"`
	UnitPrice   float64 `json:"unit_price" db:"unit_price"`
	TotalPrice  float64 `json:"total_price" db:"total_price"`
	Notes       string  `json:"notes" db:"notes"`
}

// Reminder represents maintenance reminders
type Reminder struct {
	ID          string    `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"` // email, sms, notification
	DaysBefore  int       `json:"days_before" db:"days_before"`
	Recipient   string    `json:"recipient" db:"recipient"` // email/phone
	SentAt      time.Time `json:"sent_at" db:"sent_at"`
	Status      string    `json:"status" db:"status"` // pending, sent, failed
}

// RecurringPattern represents recurring maintenance schedule
type RecurringPattern struct {
	Frequency   string `json:"frequency" db:"frequency"`     // daily, weekly, monthly, yearly
	Interval    int    `json:"interval" db:"interval"`       // every X frequency units
	DayOfWeek   *int   `json:"day_of_week,omitempty" db:"day_of_week"`     // 0-6 (Sunday-Saturday)
	DayOfMonth  *int   `json:"day_of_month,omitempty" db:"day_of_month"`   // 1-31
	Month       *int   `json:"month,omitempty" db:"month"`                 // 1-12
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	MaxOccurrences *int    `json:"max_occurrences,omitempty" db:"max_occurrences"`
}

// CreateScheduleRequest represents the request body for creating a maintenance schedule
type CreateScheduleRequest struct {
	VehicleID      uuid.UUID         `json:"vehicle_id" validate:"required"`
	Type           string            `json:"type" validate:"required,oneof=routine preventive corrective emergency"`
	Name           string            `json:"name" validate:"required,max=100"`
	Description    string            `json:"description" validate:"max=500"`
	DueDate        time.Time         `json:"due_date" validate:"required"`
	DueOdometer    *int              `json:"due_odometer,omitempty" validate:"omitempty,min=0"`
	DueEngineHours *float64          `json:"due_engine_hours,omitempty" validate:"omitempty,min=0"`
	AssignedTo     *uuid.UUID        `json:"assigned_to,omitempty"`
	Priority       string            `json:"priority" validate:"required,oneof=low medium high critical"`
	EstimatedCost  *float64          `json:"estimated_cost,omitempty" validate:"omitempty,min=0"`
	PartsRequired  []string          `json:"parts_required,omitempty"`
	ServiceItems   []ServiceItem     `json:"service_items,omitempty" validate:"dive"`
	Notes          string            `json:"notes" validate:"max=1000"`
	Reminders      []Reminder        `json:"reminders,omitempty" validate:"dive"`
	Recurring      *RecurringPattern `json:"recurring,omitempty"`
}

// UpdateScheduleRequest represents the request body for updating a maintenance schedule
type UpdateScheduleRequest struct {
	Type           *string            `json:"type,omitempty" validate:"omitempty,oneof=routine preventive corrective emergency"`
	Name           *string            `json:"name,omitempty" validate:"omitempty,max=100"`
	Description    *string            `json:"description,omitempty" validate:"omitempty,max=500"`
	DueDate        *time.Time         `json:"due_date,omitempty"`
	DueOdometer    *int               `json:"due_odometer,omitempty" validate:"omitempty,min=0"`
	DueEngineHours *float64           `json:"due_engine_hours,omitempty" validate:"omitempty,min=0"`
	AssignedTo     *uuid.UUID         `json:"assigned_to,omitempty"`
	Priority       *string            `json:"priority,omitempty" validate:"omitempty,oneof=low medium high critical"`
	Status         *string            `json:"status,omitempty" validate:"omitempty,oneof=scheduled in_progress completed cancelled"`
	EstimatedCost  *float64           `json:"estimated_cost,omitempty" validate:"omitempty,min=0"`
	ActualCost     *float64           `json:"actual_cost,omitempty" validate:"omitempty,min=0"`
	PartsRequired  *[]string          `json:"parts_required,omitempty"`
	ServiceItems   *[]ServiceItem     `json:"service_items,omitempty" validate:"omitempty,dive"`
	Notes          *string            `json:"notes,omitempty" validate:"omitempty,max=1000"`
	Reminders      *[]Reminder        `json:"reminders,omitempty" validate:"omitempty,dive"`
	Recurring      *RecurringPattern  `json:"recurring,omitempty"`
	CompletedAt    *time.Time         `json:"completed_at,omitempty"`
	CompletedBy    *uuid.UUID         `json:"completed_by,omitempty"`
}

// CreateLogRequest represents the request body for creating a maintenance log
type CreateLogRequest struct {
	VehicleID    uuid.UUID     `json:"vehicle_id" validate:"required"`
	ScheduleID   *string       `json:"schedule_id,omitempty"`
	Type         string        `json:"type" validate:"required,oneof=routine preventive corrective emergency"`
	Name         string        `json:"name" validate:"required,max=100"`
	Description  string        `json:"description" validate:"max=500"`
	StartDate    time.Time     `json:"start_date" validate:"required"`
	EndDate      time.Time     `json:"end_date" validate:"required"`
	PerformedBy  uuid.UUID     `json:"performed_by" validate:"required"`
	Cost         float64       `json:"cost" validate:"min=0"`
	Odometer     int           `json:"odometer" validate:"min=0"`
	EngineHours  *float64      `json:"engine_hours,omitempty" validate:"omitempty,min=0"`
	PartsUsed    []PartUsed    `json:"parts_used,omitempty" validate:"dive"`
	ServiceItems []ServiceItem `json:"service_items,omitempty" validate:"dive"`
	Notes        string        `json:"notes" validate:"max=1000"`
	Documents    []string      `json:"documents,omitempty"`
	Status       string        `json:"status" validate:"required,oneof=completed partially_completed"`
}

// UpdateLogRequest represents the request body for updating a maintenance log
type UpdateLogRequest struct {
	Type         *string        `json:"type,omitempty" validate:"omitempty,oneof=routine preventive corrective emergency"`
	Name         *string        `json:"name,omitempty" validate:"omitempty,max=100"`
	Description  *string        `json:"description,omitempty" validate:"omitempty,max=500"`
	StartDate    *time.Time     `json:"start_date,omitempty"`
	EndDate      *time.Time     `json:"end_date,omitempty"`
	PerformedBy  *uuid.UUID     `json:"performed_by,omitempty"`
	Cost         *float64       `json:"cost,omitempty" validate:"omitempty,min=0"`
	Odometer     *int           `json:"odometer,omitempty" validate:"omitempty,min=0"`
	EngineHours  *float64       `json:"engine_hours,omitempty" validate:"omitempty,min=0"`
	PartsUsed    *[]PartUsed    `json:"parts_used,omitempty" validate:"omitempty,dive"`
	ServiceItems *[]ServiceItem `json:"service_items,omitempty" validate:"omitempty,dive"`
	Notes        *string        `json:"notes,omitempty" validate:"omitempty,max=1000"`
	Documents    *[]string      `json:"documents,omitempty"`
	Status       *string        `json:"status,omitempty" validate:"omitempty,oneof=completed partially_completed"`
}

// MaintenanceListFilter represents filters for listing maintenance records
type MaintenanceListFilter struct {
	OrgID         uuid.UUID  `json:"org_id" form:"org_id"`
	VehicleID     *uuid.UUID `json:"vehicle_id" form:"vehicle_id"`
	Type          []string   `json:"type" form:"type"`
	Status        []string   `json:"status" form:"status"`
	Priority      []string   `json:"priority" form:"priority"`
	AssignedTo    *uuid.UUID `json:"assigned_to" form:"assigned_to"`
	FromDate      time.Time  `json:"from_date" form:"from_date"`
	ToDate        time.Time  `json:"to_date" form:"to_date"`
	Page          int        `json:"page" form:"page"`
	PageSize      int        `json:"page_size" form:"page_size"`
	OrderBy       string     `json:"order_by" form:"order_by"`
	OrderDir      string     `json:"order_dir" form:"order_dir"`
}

// MaintenanceSummary represents maintenance statistics
type MaintenanceSummary struct {
	TotalSchedules     int                `json:"total_schedules"`
	Scheduled          int                `json:"scheduled"`
	InProgress         int                `json:"in_progress"`
	Completed          int                `json:"completed"`
	Cancelled          int                `json:"cancelled"`
	Overdue            int                `json:"overdue"`
	Upcoming           int                `json:"upcoming"`
	ByType             map[string]int     `json:"by_type"`
	ByPriority         map[string]int     `json:"by_priority"`
	ByStatus           map[string]int     `json:"by_status"`
	TotalCost          float64            `json:"total_cost"`
	AverageCost        float64            `json:"average_cost"`
	MaintenanceHistory []MaintenanceLog   `json:"maintenance_history"`
}

// MaintenanceType represents different types of maintenance
const (
	MaintenanceTypeRoutine      = "routine"
	MaintenanceTypePreventive   = "preventive"
	MaintenanceTypeCorrective   = "corrective"
	MaintenanceTypeEmergency    = "emergency"
)

// MaintenancePriority represents priority levels
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// MaintenanceStatus represents maintenance statuses
const (
	StatusScheduled   = "scheduled"
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
	StatusPartiallyCompleted = "partially_completed"
)

// ReminderType represents reminder types
const (
	ReminderTypeEmail = "email"
	ReminderTypeSMS   = "sms"
	ReminderTypeNotification = "notification"
)

// ReminderStatus represents reminder statuses
const (
	ReminderStatusPending = "pending"
	ReminderStatusSent    = "sent"
	ReminderStatusFailed  = "failed"
)

// RecurringFrequency represents recurring frequencies
const (
	FrequencyDaily   = "daily"
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
	FrequencyYearly  = "yearly"
)