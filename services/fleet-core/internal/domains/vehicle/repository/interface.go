package repository

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/google/uuid"
)

// VehicleRepositoryInterface is the primary vehicle data access contract.
type VehicleRepositoryInterface interface {
	Create(ctx context.Context, vehicle *models.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error)
	Update(ctx context.Context, vehicle *models.Vehicle) error
	Delete(ctx context.Context, id, orgID uuid.UUID) error
	AssignDriver(ctx context.Context, vehicleID, orgID, driverID uuid.UUID) error
	GetStats(ctx context.Context) (*models.VehicleStats, error)

	// Enterprise extensions
	GetByRegistrationNumber(ctx context.Context, tenantID uuid.UUID, registrationNumber string) (*models.Vehicle, error)
	Search(ctx context.Context, params models.VehicleSearchParams) ([]*models.Vehicle, int64, error)
}

// ComplianceRepositoryInterface is the compliance sub-domain data access contract.
type ComplianceRepositoryInterface interface {
	Upsert(ctx context.Context, c *models.VehicleCompliance) error
	Get(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleCompliance, error)
	GetExpiringVehicles(ctx context.Context, tenantID uuid.UUID, withinDays int) ([]models.ComplianceExpiryAlert, error)
}

// FinancialRepositoryInterface is the financials sub-domain data access contract.
type FinancialRepositoryInterface interface {
	Upsert(ctx context.Context, f *models.VehicleFinancials) error
	Get(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleFinancials, error)
}

// MaintenanceScheduleRepositoryInterface is the maintenance schedule data access contract.
type MaintenanceScheduleRepositoryInterface interface {
	Upsert(ctx context.Context, s *models.MaintenanceSchedule) error
	Get(ctx context.Context, vehicleID uuid.UUID) (*models.MaintenanceSchedule, error)
}

// DocumentsRepositoryInterface is the immutable document store contract.
type DocumentsRepositoryInterface interface {
	Insert(ctx context.Context, doc *models.VehicleDocument) error
	List(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleDocument, error)
	Get(ctx context.Context, id uuid.UUID) (*models.VehicleDocument, error)
}

// EventsRepositoryInterface is the append-only event log contract.
type EventsRepositoryInterface interface {
	Append(ctx context.Context, event *models.VehicleEvent) error
	List(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.VehicleEvent, int64, error)
}
