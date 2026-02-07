package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/maintenance-service/internal/models"
	"github.com/flyfleet/maintenance-service/internal/repository"
)

type MaintenanceService struct {
	repo   *repository.MaintenanceRepositoryImpl
	logger *zap.Logger
}

func NewMaintenanceService(
	repo *repository.MaintenanceRepositoryImpl,
	logger *zap.Logger,
) *MaintenanceService {
	return &MaintenanceService{
		repo:   repo,
		logger: logger,
	}
}

// CreateSchedule creates a new maintenance schedule
func (s *MaintenanceService) CreateSchedule(ctx context.Context, orgID uuid.UUID, req *models.CreateScheduleRequest, createdBy uuid.UUID) (*models.MaintenanceSchedule, error) {
	// Validate maintenance type
	validTypes := map[string]bool{
		models.MaintenanceTypeRoutine:    true,
		models.MaintenanceTypePreventive: true,
		models.MaintenanceTypeCorrective: true,
		models.MaintenanceTypeEmergency:  true,
	}

	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid maintenance type: %s", req.Type)
	}

	// Validate priority
	validPriorities := map[string]bool{
		models.PriorityLow:      true,
		models.PriorityMedium:   true,
		models.PriorityHigh:     true,
		models.PriorityCritical: true,
	}

	if !validPriorities[req.Priority] {
		return nil, fmt.Errorf("invalid priority: %s", req.Priority)
	}

	schedule := &models.MaintenanceSchedule{
		ID:            uuid.New().String(),
		OrgID:         orgID,
		VehicleID:     req.VehicleID,
		Type:          req.Type,
		Name:          req.Name,
		Description:   req.Description,
		DueDate:       req.DueDate,
		DueOdometer:   req.DueOdometer,
		DueEngineHours: req.DueEngineHours,
		AssignedTo:    req.AssignedTo,
		Priority:      req.Priority,
		Status:        models.StatusScheduled,
		EstimatedCost: req.EstimatedCost,
		PartsRequired: req.PartsRequired,
		ServiceItems:  req.ServiceItems,
		Notes:         req.Notes,
		Reminders:     req.Reminders,
		Recurring:     req.Recurring,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
		s.logger.Error("failed to create maintenance schedule", zap.Error(err))
		return nil, fmt.Errorf("failed to create maintenance schedule: %w", err)
	}

	s.logger.Info("maintenance schedule created", zap.String("schedule_id", schedule.ID))
	return schedule, nil
}

// GetSchedule retrieves a maintenance schedule by ID
func (s *MaintenanceService) GetSchedule(ctx context.Context, id string) (*models.MaintenanceSchedule, error) {
	schedule, err := s.repo.GetScheduleByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		return nil, fmt.Errorf("maintenance schedule not found: %w", err)
	}
	return schedule, nil
}

// ListSchedules retrieves maintenance schedules with filtering and pagination
func (s *MaintenanceService) ListSchedules(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceSchedule, int, error) {
	schedules, totalCount, err := s.repo.ListSchedules(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list maintenance schedules", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list maintenance schedules: %w", err)
	}
	return schedules, totalCount, nil
}

// UpdateSchedule modifies an existing maintenance schedule
func (s *MaintenanceService) UpdateSchedule(ctx context.Context, id string, req *models.UpdateScheduleRequest) (*models.MaintenanceSchedule, error) {
	// Validate that schedule exists
	_, err := s.GetSchedule(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	// Validate and add updates
	if req.Type != nil {
		validTypes := map[string]bool{
			models.MaintenanceTypeRoutine:    true,
			models.MaintenanceTypePreventive: true,
			models.MaintenanceTypeCorrective: true,
			models.MaintenanceTypeEmergency:  true,
		}
		if !validTypes[*req.Type] {
			return nil, fmt.Errorf("invalid maintenance type: %s", *req.Type)
		}
		updates["type"] = *req.Type
	}

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}

	if req.DueOdometer != nil {
		updates["due_odometer"] = *req.DueOdometer
	}

	if req.DueEngineHours != nil {
		updates["due_engine_hours"] = *req.DueEngineHours
	}

	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
	}

	if req.Priority != nil {
		validPriorities := map[string]bool{
			models.PriorityLow:      true,
			models.PriorityMedium:   true,
			models.PriorityHigh:     true,
			models.PriorityCritical: true,
		}
		if !validPriorities[*req.Priority] {
			return nil, fmt.Errorf("invalid priority: %s", *req.Priority)
		}
		updates["priority"] = *req.Priority
	}

	if req.Status != nil {
		validStatuses := map[string]bool{
			models.StatusScheduled:   true,
			models.StatusInProgress:  true,
			models.StatusCompleted:   true,
			models.StatusCancelled:   true,
		}
		if !validStatuses[*req.Status] {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		updates["status"] = *req.Status
	}

	if req.EstimatedCost != nil {
		updates["estimated_cost"] = *req.EstimatedCost
	}

	if req.ActualCost != nil {
		updates["actual_cost"] = *req.ActualCost
	}

	if req.PartsRequired != nil {
		updates["parts_required"] = *req.PartsRequired
	}

	if req.ServiceItems != nil {
		updates["service_items"] = *req.ServiceItems
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if req.Reminders != nil {
		updates["reminders"] = *req.Reminders
	}

	if req.Recurring != nil {
		updates["recurring"] = *req.Recurring
	}

	if req.CompletedAt != nil {
		updates["completed_at"] = *req.CompletedAt
	}

	if req.CompletedBy != nil {
		updates["completed_by"] = *req.CompletedBy
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := s.repo.UpdateSchedule(ctx, id, updates); err != nil {
			s.logger.Error("failed to update maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
			return nil, fmt.Errorf("failed to update maintenance schedule: %w", err)
		}
	}

	return s.GetSchedule(ctx, id)
}

// DeleteSchedule removes a maintenance schedule
func (s *MaintenanceService) DeleteSchedule(ctx context.Context, id string) error {
	// Check if schedule exists
	_, err := s.GetSchedule(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteSchedule(ctx, id); err != nil {
		s.logger.Error("failed to delete maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete maintenance schedule: %w", err)
	}

	s.logger.Info("maintenance schedule deleted", zap.String("schedule_id", id))
	return nil
}

// CreateLog creates a new maintenance log
func (s *MaintenanceService) CreateLog(ctx context.Context, orgID uuid.UUID, req *models.CreateLogRequest) (*models.MaintenanceLog, error) {
	// Validate maintenance type
	validTypes := map[string]bool{
		models.MaintenanceTypeRoutine:    true,
		models.MaintenanceTypePreventive: true,
		models.MaintenanceTypeCorrective: true,
		models.MaintenanceTypeEmergency:  true,
	}

	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid maintenance type: %s", req.Type)
	}

	// Validate status
	validStatuses := map[string]bool{
		models.StatusCompleted:           true,
		models.StatusPartiallyCompleted:  true,
	}

	if !validStatuses[req.Status] {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	log := &models.MaintenanceLog{
		ID:            uuid.New().String(),
		OrgID:         orgID,
		VehicleID:     req.VehicleID,
		ScheduleID:    req.ScheduleID,
		Type:          req.Type,
		Name:          req.Name,
		Description:   req.Description,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		PerformedBy:   req.PerformedBy,
		Cost:          req.Cost,
		Odometer:      req.Odometer,
		EngineHours:   req.EngineHours,
		PartsUsed:     req.PartsUsed,
		ServiceItems:  req.ServiceItems,
		Notes:         req.Notes,
		Documents:     req.Documents,
		Status:        req.Status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		s.logger.Error("failed to create maintenance log", zap.Error(err))
		return nil, fmt.Errorf("failed to create maintenance log: %w", err)
	}

	s.logger.Info("maintenance log created", zap.String("log_id", log.ID))
	return log, nil
}

// GetLog retrieves a maintenance log by ID
func (s *MaintenanceService) GetLog(ctx context.Context, id string) (*models.MaintenanceLog, error) {
	log, err := s.repo.GetLogByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get maintenance log", zap.String("log_id", id), zap.Error(err))
		return nil, fmt.Errorf("maintenance log not found: %w", err)
	}
	return log, nil
}

// ListLogs retrieves maintenance logs with filtering and pagination
func (s *MaintenanceService) ListLogs(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceLog, int, error) {
	logs, totalCount, err := s.repo.ListLogs(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list maintenance logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list maintenance logs: %w", err)
	}
	return logs, totalCount, nil
}

// UpdateLog modifies an existing maintenance log
func (s *MaintenanceService) UpdateLog(ctx context.Context, id string, req *models.UpdateLogRequest) (*models.MaintenanceLog, error) {
	// Validate that log exists
	_, err := s.GetLog(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	// Validate and add updates
	if req.Type != nil {
		validTypes := map[string]bool{
			models.MaintenanceTypeRoutine:    true,
			models.MaintenanceTypePreventive: true,
			models.MaintenanceTypeCorrective: true,
			models.MaintenanceTypeEmergency:  true,
		}
		if !validTypes[*req.Type] {
			return nil, fmt.Errorf("invalid maintenance type: %s", *req.Type)
		}
		updates["type"] = *req.Type
	}

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}

	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}

	if req.PerformedBy != nil {
		updates["performed_by"] = *req.PerformedBy
	}

	if req.Cost != nil {
		updates["cost"] = *req.Cost
	}

	if req.Odometer != nil {
		updates["odometer"] = *req.Odometer
	}

	if req.EngineHours != nil {
		updates["engine_hours"] = *req.EngineHours
	}

	if req.PartsUsed != nil {
		updates["parts_used"] = *req.PartsUsed
	}

	if req.ServiceItems != nil {
		updates["service_items"] = *req.ServiceItems
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if req.Documents != nil {
		updates["documents"] = *req.Documents
	}

	if req.Status != nil {
		validStatuses := map[string]bool{
			models.StatusCompleted:           true,
			models.StatusPartiallyCompleted:  true,
		}
		if !validStatuses[*req.Status] {
			return nil, fmt.Errorf("invalid status: %s", *req.Status)
		}
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := s.repo.UpdateLog(ctx, id, updates); err != nil {
			s.logger.Error("failed to update maintenance log", zap.String("log_id", id), zap.Error(err))
			return nil, fmt.Errorf("failed to update maintenance log: %w", err)
		}
	}

	return s.GetLog(ctx, id)
}

// DeleteLog removes a maintenance log
func (s *MaintenanceService) DeleteLog(ctx context.Context, id string) error {
	// Check if log exists
	_, err := s.GetLog(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteLog(ctx, id); err != nil {
		s.logger.Error("failed to delete maintenance log", zap.String("log_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete maintenance log: %w", err)
	}

	s.logger.Info("maintenance log deleted", zap.String("log_id", id))
	return nil
}

// GetMaintenanceHistory retrieves maintenance history for a vehicle
func (s *MaintenanceService) GetMaintenanceHistory(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceLog, error) {
	logs, err := s.repo.GetMaintenanceHistory(ctx, vehicleID, limit)
	if err != nil {
		s.logger.Error("failed to get maintenance history", zap.String("vehicle_id", vehicleID.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance history: %w", err)
	}
	return logs, nil
}

// GetOverdueSchedules retrieves overdue maintenance schedules
func (s *MaintenanceService) GetOverdueSchedules(ctx context.Context, orgID uuid.UUID) ([]*models.MaintenanceSchedule, error) {
	schedules, err := s.repo.GetOverdueSchedules(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get overdue schedules", zap.Error(err))
		return nil, fmt.Errorf("failed to get overdue schedules: %w", err)
	}
	return schedules, nil
}

// GetUpcomingSchedules retrieves upcoming maintenance schedules
func (s *MaintenanceService) GetUpcomingSchedules(ctx context.Context, orgID uuid.UUID, days int) ([]*models.MaintenanceSchedule, error) {
	schedules, err := s.repo.GetUpcomingSchedules(ctx, orgID, days)
	if err != nil {
		s.logger.Error("failed to get upcoming schedules", zap.Error(err))
		return nil, fmt.Errorf("failed to get upcoming schedules: %w", err)
	}
	return schedules, nil
}

// GetMaintenanceSummary retrieves maintenance statistics
func (s *MaintenanceService) GetMaintenanceSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.MaintenanceSummary, error) {
	summary, err := s.repo.GetMaintenanceSummary(ctx, orgID, fromDate, toDate)
	if err != nil {
		s.logger.Error("failed to get maintenance summary", zap.Error(err))
		return nil, fmt.Errorf("failed to get maintenance summary: %w", err)
	}
	return summary, nil
}
