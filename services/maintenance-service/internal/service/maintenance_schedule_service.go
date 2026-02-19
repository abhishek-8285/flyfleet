package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/maintenance-service/internal/domain"
	"github.com/flyfleet/maintenance-service/internal/domain/maintenance_schedule"
	"github.com/google/uuid"
)

type MaintenanceScheduleService struct {
	repo   maintenance_schedule.Repository
	events domain.EventPublisher
}

func NewMaintenanceScheduleService(repo maintenance_schedule.Repository, events domain.EventPublisher) *MaintenanceScheduleService {
	return &MaintenanceScheduleService{
		repo:   repo,
		events: events,
	}
}

func (s *MaintenanceScheduleService) Create(ctx context.Context, orgID uuid.UUID, schedule *maintenance_schedule.MaintenanceSchedule) (*maintenance_schedule.MaintenanceSchedule, error) {
	schedule.OrgID = orgID
	schedule.Status = maintenance_schedule.StatusScheduled
	
	if err := s.repo.Create(ctx, schedule); err != nil {
		return nil, err
	}

	// Publish Event
	_ = s.events.Publish(ctx, "maintenance_schedule.created", schedule)

	return schedule, nil
}

func (s *MaintenanceScheduleService) GetByID(ctx context.Context, orgID, id uuid.UUID) (*maintenance_schedule.MaintenanceSchedule, error) {
	schedule, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, fmt.Errorf("maintenance schedule not found")
	}
	return schedule, nil
}

func (s *MaintenanceScheduleService) Update(ctx context.Context, orgID, id uuid.UUID, update *maintenance_schedule.MaintenanceSchedule) (*maintenance_schedule.MaintenanceSchedule, error) {
	schedule, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	schedule.Type = update.Type
	schedule.Name = update.Name
	schedule.Description = update.Description
	schedule.DueDate = update.DueDate
	schedule.DueOdometer = update.DueOdometer
	schedule.Priority = update.Priority
	schedule.AssignedTo = update.AssignedTo
	schedule.EstimatedCost = update.EstimatedCost
	schedule.Notes = update.Notes

	if err := s.repo.Update(ctx, schedule); err != nil {
		return nil, err
	}

	_ = s.events.Publish(ctx, "maintenance_schedule.updated", schedule)

	return schedule, nil
}

func (s *MaintenanceScheduleService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, orgID, id); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "maintenance_schedule.deleted", map[string]interface{}{"id": id})
	return nil
}

func (s *MaintenanceScheduleService) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*maintenance_schedule.MaintenanceSchedule, int64, error) {
	return s.repo.List(ctx, orgID, limit, offset)
}

func (s *MaintenanceScheduleService) MarkComplete(ctx context.Context, orgID, id uuid.UUID, actualCost float64, notes string) (*maintenance_schedule.MaintenanceSchedule, error) {
	schedule, err := s.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	schedule.Status = maintenance_schedule.StatusCompleted
	schedule.ActualCost = &actualCost
	schedule.CompletedAt = &now
	if n := notes; n != "" {
		schedule.Notes = fmt.Sprintf("%s\nCompletion Notes: %s", schedule.Notes, n)
	}

	if err := s.repo.Update(ctx, schedule); err != nil {
		return nil, err
	}

	_ = s.events.Publish(ctx, "maintenance_schedule.completed", schedule)

	return schedule, nil
}

