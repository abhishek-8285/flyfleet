package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/flyfleet/maintenance-service/internal/models"
)

// MaintenanceRepository defines the interface for maintenance data operations
type MaintenanceRepository interface {
	// Schedule operations
	CreateSchedule(ctx context.Context, schedule *models.MaintenanceSchedule) error
	GetScheduleByID(ctx context.Context, id string) (*models.MaintenanceSchedule, error)
	GetSchedulesByVehicle(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceSchedule, error)
	GetSchedulesByOrg(ctx context.Context, orgID uuid.UUID, filters *ScheduleFilters) ([]*models.MaintenanceSchedule, error)
	UpdateSchedule(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteSchedule(ctx context.Context, id string) error
	
	// Log operations
	CreateLog(ctx context.Context, log *models.MaintenanceLog) error
	GetLogByID(ctx context.Context, id string) (*models.MaintenanceLog, error)
	UpdateLog(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteLog(ctx context.Context, id string) error
	
	// Query operations
	GetMaintenanceHistory(ctx context.Context, vehicleID uuid.UUID, limit int) ([]*models.MaintenanceLog, error)
	GetOverdueSchedules(ctx context.Context, orgID uuid.UUID) ([]*models.MaintenanceSchedule, error)
	GetUpcomingSchedules(ctx context.Context, orgID uuid.UUID, days int) ([]*models.MaintenanceSchedule, error)
	GetMaintenanceSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.MaintenanceSummary, error)
}

// ScheduleFilters represents filters for querying schedules
type ScheduleFilters struct {
	Status    *string
	Type      *string
	Priority  *string
	FromDate  *time.Time
	ToDate    *time.Time
	Overdue   *bool
	Upcoming  *bool
	Page      int
	PageSize  int
}