package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/google/uuid"
)

// VehicleRepositoryInterface is the primary vehicle data access contract.
type VehicleRepositoryInterface interface {
	Create(ctx context.Context, vehicle *models.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error)
	Update(ctx context.Context, vehicle *models.Vehicle) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
	AssignDriver(ctx context.Context, vehicleID, tenantID, driverID uuid.UUID) error
	UnassignDriver(ctx context.Context, vehicleID, tenantID uuid.UUID) error
	GetStats(ctx context.Context) (*models.VehicleStats, error)

	// Enterprise extensions
	GetByRegistrationNumber(ctx context.Context, tenantID uuid.UUID, registrationNumber string) (*models.Vehicle, error)
	Search(ctx context.Context, params models.VehicleSearchParams) ([]*models.Vehicle, int64, error)
	IsGPSAvailable(ctx context.Context, tenantID uuid.UUID, deviceID string, excludingVehicleID *uuid.UUID) (bool, string, error)
	GetAuditLogs(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleAuditLog, error)
}

// ComplianceRepositoryInterface is the compliance sub-domain data access contract.
type ComplianceRepositoryInterface interface {
	Upsert(ctx context.Context, c *models.VehicleCompliance) error
	Get(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleCompliance, error)
	GetExpiringVehicles(ctx context.Context, tenantID uuid.UUID, withinDays int) ([]models.ComplianceExpiryAlert, error)
}

// DocumentsRepositoryInterface is the document sub-domain data access contract.
type DocumentsRepositoryInterface interface {
	Insert(ctx context.Context, doc *models.VehicleDocument) error
	List(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleDocument, error)
	Get(ctx context.Context, id uuid.UUID) (*models.VehicleDocument, error)
}

// EventsRepositoryInterface tracks domain events for audit and history.
type EventsRepositoryInterface interface {
	Append(ctx context.Context, event *models.VehicleEvent) error
	List(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.VehicleEvent, int64, error)
}
