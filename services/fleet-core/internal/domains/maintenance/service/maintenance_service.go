package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/repository"
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

func (s *MaintenanceService) CreateTechnician(ctx context.Context, tech *models.Technician) error {
	return s.repo.CreateTechnician(ctx, tech)
}

func (s *MaintenanceService) AddTechnicianSkill(ctx context.Context, techID uuid.UUID, skill string, level int) error {
	return s.repo.AddSkill(ctx, &models.TechnicianSkill{
		TechnicianID: techID,
		SkillName:    skill,
		Level:        level,
	})
}

func (s *MaintenanceService) GetTechnician(ctx context.Context, id uuid.UUID) (*models.Technician, error) {
	return s.repo.GetTechnicianByID(ctx, id)
}

func (s *MaintenanceService) CreateSchedule(ctx context.Context, orgID uuid.UUID, req *models.CreateScheduleRequest, createdBy uuid.UUID) (*models.MaintenanceSchedule, error) {
	schedule := &models.MaintenanceSchedule{
		ID:             uuid.New().String(),
		OrgID:          orgID,
		VehicleID:      req.VehicleID,
		Type:           req.Type,
		Name:           req.Name,
		Description:    &req.Description,
		DueDate:        req.DueDate,
		DueOdometer:    req.DueOdometer,
		DueEngineHours: req.DueEngineHours,
		AssignedTo:     req.AssignedTo,
		Priority:       req.Priority,
		Status:         models.StatusScheduled,
		EstimatedCost:  req.EstimatedCost,
		PartsRequired:  req.PartsRequired,
		ServiceItems:   req.ServiceItems,
		Notes:          req.Notes,
		Reminders:      req.Reminders,
		Recurring:      req.Recurring,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *MaintenanceService) GetSchedule(ctx context.Context, id string, orgID uuid.UUID) (*models.MaintenanceSchedule, error) {
	return s.repo.GetScheduleByID(ctx, id)
}

func (s *MaintenanceService) ListSchedules(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceSchedule, int, error) {
	return s.repo.ListSchedules(ctx, filter)
}

func (s *MaintenanceService) UpdateSchedule(ctx context.Context, id string, orgID uuid.UUID, req *models.UpdateScheduleRequest) (*models.MaintenanceSchedule, error) {
	updates := make(map[string]interface{})
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Status != nil {
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

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.repo.UpdateSchedule(ctx, id, updates); err != nil {
			return nil, err
		}
	}
	return s.GetSchedule(ctx, id, orgID)
}

func (s *MaintenanceService) DeleteSchedule(ctx context.Context, id string, orgID uuid.UUID) error {
	return s.repo.DeleteSchedule(ctx, id)
}

func (s *MaintenanceService) CreateLog(ctx context.Context, orgID uuid.UUID, req *models.CreateLogRequest) (*models.MaintenanceLog, error) {
	log := &models.MaintenanceLog{
		OrgID:        orgID,
		VehicleID:    req.VehicleID,
		ScheduleID:   req.ScheduleID,
		Type:         req.Type,
		Name:         req.Name,
		Description:  &req.Description,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Status:       req.Status,
		Cost:         req.Cost,
		Odometer:     req.Odometer,
		EngineHours:  req.EngineHours,
		PerformedBy:  req.PerformedBy,
		PartsUsed:    req.PartsUsed,
		ServiceItems: req.ServiceItems,
		Notes:        req.Notes,
		Documents:    req.Documents,
	}
	if err := s.repo.CreateLog(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *MaintenanceService) GetLog(ctx context.Context, id string, orgID uuid.UUID) (*models.MaintenanceLog, error) {
	return s.repo.GetLogByID(ctx, id)
}

func (s *MaintenanceService) ListLogs(ctx context.Context, filter models.MaintenanceListFilter) ([]*models.MaintenanceLog, int, error) {
	return s.repo.ListLogs(ctx, filter)
}

func (s *MaintenanceService) UpdateLog(ctx context.Context, id string, orgID uuid.UUID, req *models.UpdateLogRequest) (*models.MaintenanceLog, error) {
	updates := make(map[string]interface{})
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Cost != nil {
		updates["cost"] = *req.Cost
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.repo.UpdateLog(ctx, id, updates); err != nil {
			return nil, err
		}
	}
	return s.GetLog(ctx, id, orgID)
}

func (s *MaintenanceService) DeleteLog(ctx context.Context, id string, orgID uuid.UUID) error {
	return s.repo.DeleteLog(ctx, id)
}

func (s *MaintenanceService) GetMaintenanceHistory(ctx context.Context, vehicleID, orgID uuid.UUID, limit int) ([]*models.MaintenanceLog, error) {
	return s.repo.GetMaintenanceHistory(ctx, vehicleID, limit)
}

func (s *MaintenanceService) GetOverdueSchedules(ctx context.Context, orgID uuid.UUID) ([]*models.MaintenanceSchedule, error) {
	return s.repo.GetOverdueSchedules(ctx, orgID)
}

func (s *MaintenanceService) GetUpcomingSchedules(ctx context.Context, orgID uuid.UUID, days int) ([]*models.MaintenanceSchedule, error) {
	return s.repo.GetUpcomingSchedules(ctx, orgID, days)
}

func (s *MaintenanceService) CreateWarranty(ctx context.Context, w *models.ComponentWarranty) error {
	return s.repo.CreateWarranty(ctx, w)
}

func (s *MaintenanceService) CheckWarrantyCoverage(ctx context.Context, vehicleID uuid.UUID, parts []string) ([]models.ComponentWarranty, error) {
	return s.repo.GetWarrantyCoverage(ctx, vehicleID, parts)
}

func (s *MaintenanceService) GetWarrantyExpiryAlerts(ctx context.Context, orgID uuid.UUID, days int) ([]models.ComponentWarranty, error) {
	return s.repo.GetWarrantyExpiryAlerts(ctx, orgID, days)
}

func (s *MaintenanceService) GetMaintenanceSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.MaintenanceSummary, error) {
	return s.repo.GetMaintenanceSummary(ctx, orgID, fromDate, toDate)
}
