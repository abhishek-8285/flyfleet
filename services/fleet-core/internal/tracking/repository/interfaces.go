package repository

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

// VehicleRepositoryer defines the interface for vehicle data access.
type VehicleRepositoryer interface {
	GetOrgVehicles(ctx context.Context, tenantID uuid.UUID) ([]VehicleRecord, error)
	GetDriverInfo(ctx context.Context, driverID uuid.UUID) (*models.DriverInfo, error)
	GetDriversBatch(ctx context.Context, driverIDs []uuid.UUID) (map[uuid.UUID]*models.DriverInfo, error)
}

// LiveCacher defines the interface for the Redis live location cache.
type LiveCacher interface {
	SetLiveLocation(ctx context.Context, vehicleID uuid.UUID, loc *models.LiveLocation) error
	GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*models.LiveLocation, error)
	GetBatchLiveLocations(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.LiveLocation, error)
	PublishPositionUpdate(ctx context.Context, tenantID uuid.UUID, msg []byte) error
	GetGeofenceState(ctx context.Context, vehicleID, geofenceID string) (bool, error)
	SetGeofenceState(ctx context.Context, vehicleID, geofenceID string, inside bool) error
}

// TelemetryRepositoryer defines the interface for telemetry data access.
type TelemetryRepositoryer interface {
	Write(ctx context.Context, event *models.TelemetryEvent) error
	GetLatestByVehicle(ctx context.Context, vehicleID uuid.UUID) (*models.TelemetryEvent, error)
	GetBatchLatest(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.TelemetryEvent, error)
	GetHistoryBucket(ctx context.Context, vehicleID uuid.UUID, start, end time.Time, bucketInterval string) ([]models.TelemetryEvent, error)
}

// GeofenceRepositoryer defines the interface for geofence data access.
type GeofenceRepositoryer interface {
	GetOrgGeofences(ctx context.Context, tenantID uuid.UUID) ([]models.Geofence, error)
	CreateGeofence(ctx context.Context, g *models.Geofence) error
	CheckPointInGeofences(ctx context.Context, tenantID uuid.UUID, lat, lng float64) ([]models.Geofence, error)
	LogGeofenceEvent(ctx context.Context, event *models.GeofenceEvent) error
}

// TenantChecker verifies that a vehicle belongs to an organization.
type TenantChecker interface {
	VehicleBelongsToOrg(ctx context.Context, tenantID, vehicleID uuid.UUID) (bool, error)
}
