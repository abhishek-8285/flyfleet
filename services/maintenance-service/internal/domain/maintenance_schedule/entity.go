package maintenance_schedule

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Status string
type Priority string
type MaintenanceType string

const (
	StatusScheduled  Status = "scheduled"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusOverdue    Status = "overdue"

	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"

	TypePreventive MaintenanceType = "preventive"
	TypeCorrective MaintenanceType = "corrective"
	TypeInspection MaintenanceType = "inspection"
)

type MaintenanceSchedule struct {
	ID            uuid.UUID       `json:"id"`
	OrgID         uuid.UUID       `json:"org_id"`
	VehicleID     uuid.UUID       `json:"vehicle_id" validate:"required"`
	Type          MaintenanceType `json:"type" validate:"required,oneof=preventive corrective inspection"`
	Name          string          `json:"name" validate:"required"`
	Description   string          `json:"description"`
	DueDate       time.Time       `json:"due_date" validate:"required"`
	DueOdometer   *int            `json:"due_odometer,omitempty"`
	Status        Status          `json:"status"`
	Priority      Priority        `json:"priority" validate:"required,oneof=low medium high critical"`
	AssignedTo    *uuid.UUID      `json:"assigned_to,omitempty"`
	EstimatedCost float64         `json:"estimated_cost"`
	ActualCost    *float64        `json:"actual_cost,omitempty"`
	PartsRequired []interface{}   `json:"parts_required"`
	Notes         string          `json:"notes"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	CompletedBy   *uuid.UUID      `json:"completed_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, schedule *MaintenanceSchedule) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*MaintenanceSchedule, error)
	Update(ctx context.Context, schedule *MaintenanceSchedule) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*MaintenanceSchedule, int64, error)
}

type Service interface {
	Create(ctx context.Context, orgID uuid.UUID, schedule *MaintenanceSchedule) (*MaintenanceSchedule, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*MaintenanceSchedule, error)
	Update(ctx context.Context, orgID, id uuid.UUID, schedule *MaintenanceSchedule) (*MaintenanceSchedule, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*MaintenanceSchedule, int64, error)
	MarkComplete(ctx context.Context, orgID, id uuid.UUID, actualCost float64, notes string) (*MaintenanceSchedule, error)
}

