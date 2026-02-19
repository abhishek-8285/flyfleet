package maintenance_schedule

import (
	"time"

	"github.com/flyfleet/maintenance-service/internal/domain/maintenance_schedule"
	"github.com/google/uuid"
)

type CreateRequest struct {
	VehicleID     uuid.UUID                          `json:"vehicle_id" validate:"required"`
	Type          maintenance_schedule.MaintenanceType `json:"type" validate:"required,oneof=preventive corrective inspection"`
	Name          string                             `json:"name" validate:"required"`
	Description   string                             `json:"description"`
	DueDate       time.Time                          `json:"due_date" validate:"required"`
	DueOdometer   *int                               `json:"due_odometer,omitempty"`
	Priority      maintenance_schedule.Priority      `json:"priority" validate:"required,oneof=low medium high critical"`
	AssignedTo    *uuid.UUID                         `json:"assigned_to,omitempty"`
	EstimatedCost float64                            `json:"estimated_cost"`
	PartsRequired []interface{}                      `json:"parts_required"`
	Notes         string                             `json:"notes"`
}

func (r CreateRequest) ToDomain() *maintenance_schedule.MaintenanceSchedule {
	return &maintenance_schedule.MaintenanceSchedule{
		VehicleID:     r.VehicleID,
		Type:          r.Type,
		Name:          r.Name,
		Description:   r.Description,
		DueDate:       r.DueDate,
		DueOdometer:   r.DueOdometer,
		Priority:      r.Priority,
		AssignedTo:    r.AssignedTo,
		EstimatedCost: r.EstimatedCost,
		PartsRequired: r.PartsRequired,
		Notes:         r.Notes,
	}
}

type UpdateRequest struct {
	Type          maintenance_schedule.MaintenanceType `json:"type" validate:"required,oneof=preventive corrective inspection"`
	Name          string                             `json:"name" validate:"required"`
	Description   string                             `json:"description"`
	DueDate       time.Time                          `json:"due_date" validate:"required"`
	DueOdometer   *int                               `json:"due_odometer,omitempty"`
	Priority      maintenance_schedule.Priority      `json:"priority" validate:"required,oneof=low medium high critical"`
	AssignedTo    *uuid.UUID                         `json:"assigned_to,omitempty"`
	EstimatedCost float64                            `json:"estimated_cost"`
	PartsRequired []interface{}                      `json:"parts_required"`
	Notes         string                             `json:"notes"`
}

func (r UpdateRequest) ToDomain() *maintenance_schedule.MaintenanceSchedule {
	return &maintenance_schedule.MaintenanceSchedule{
		Type:          r.Type,
		Name:          r.Name,
		Description:   r.Description,
		DueDate:       r.DueDate,
		DueOdometer:   r.DueOdometer,
		Priority:      r.Priority,
		AssignedTo:    r.AssignedTo,
		EstimatedCost: r.EstimatedCost,
		PartsRequired: r.PartsRequired,
		Notes:         r.Notes,
	}
}

type CompleteRequest struct {
	ActualCost float64 `json:"actual_cost" validate:"required,min=0"`
	Notes      string  `json:"notes"`
}

