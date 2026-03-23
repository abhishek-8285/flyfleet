package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/google/uuid"
)

// IndiaIntegrationRepository defines the interface for India integration data access
type IndiaIntegrationRepository interface {
	// FASTag Transaction operations
	CreateFASTagTransaction(ctx context.Context, transaction *models.FASTagTransaction) error
	GetFASTagTransaction(ctx context.Context, id string) (*models.FASTagTransaction, error)
	UpdateFASTagTransaction(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteFASTagTransaction(ctx context.Context, id string) error
	ListFASTagTransactions(ctx context.Context, vehicleID *uuid.UUID, tagNumber *string, startDate, endDate *time.Time, limit, offset int) ([]*models.FASTagTransaction, int, error)
	
	// E-Way Bill operations
	CreateEWayBill(ctx context.Context, bill *models.EWayBill) error
	GetEWayBill(ctx context.Context, id string) (*models.EWayBill, error)
	GetEWayBillByNumber(ctx context.Context, number string) (*models.EWayBill, error)
	UpdateEWayBill(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteEWayBill(ctx context.Context, id string) error
	ListEWayBills(ctx context.Context, vehicleNumber *string, transporterID *string, startDate, endDate *time.Time, limit, offset int) ([]*models.EWayBill, int, error)
	
	// Vehicle Verification operations
	CreateVehicleVerification(ctx context.Context, verification *models.VehicleVerification) error
	GetVehicleVerification(ctx context.Context, id string) (*models.VehicleVerification, error)
	GetVehicleVerificationByRegNumber(ctx context.Context, regNumber string) (*models.VehicleVerification, error)
	UpdateVehicleVerification(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteVehicleVerification(ctx context.Context, id string) error
	ListVehicleVerifications(ctx context.Context, regNumber *string, status *string, limit, offset int) ([]*models.VehicleVerification, int, error)
	
	// Driver Verification operations
	CreateDriverVerification(ctx context.Context, verification *models.DriverVerification) error
	GetDriverVerification(ctx context.Context, id string) (*models.DriverVerification, error)
	GetDriverVerificationByLicenseNumber(ctx context.Context, licenseNumber string) (*models.DriverVerification, error)
	UpdateDriverVerification(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteDriverVerification(ctx context.Context, id string) error
	ListDriverVerifications(ctx context.Context, licenseNumber *string, status *string, limit, offset int) ([]*models.DriverVerification, int, error)
	
	// Local Permit operations
	CreateLocalPermit(ctx context.Context, permit *models.LocalPermit) error
	GetLocalPermit(ctx context.Context, id string) (*models.LocalPermit, error)
	GetLocalPermitByNumber(ctx context.Context, permitNumber string) (*models.LocalPermit, error)
	UpdateLocalPermit(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteLocalPermit(ctx context.Context, id string) error
	ListLocalPermits(ctx context.Context, vehicleID *uuid.UUID, permitType *string, status *string, limit, offset int) ([]*models.LocalPermit, int, error)
	
	// GST Compliance operations
	CreateGSTCompliance(ctx context.Context, compliance *models.GSTCompliance) error
	GetGSTCompliance(ctx context.Context, id string) (*models.GSTCompliance, error)
	GetGSTComplianceByGSTIN(ctx context.Context, gstin string) (*models.GSTCompliance, error)
	UpdateGSTCompliance(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteGSTCompliance(ctx context.Context, id string) error
	ListGSTCompliances(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]*models.GSTCompliance, int, error)
	
	// Integration Config operations
	CreateIntegrationConfig(ctx context.Context, config *models.IntegrationConfig) error
	GetIntegrationConfig(ctx context.Context, id string) (*models.IntegrationConfig, error)
	GetIntegrationConfigByServiceType(ctx context.Context, orgID uuid.UUID, serviceType string) (*models.IntegrationConfig, error)
	UpdateIntegrationConfig(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteIntegrationConfig(ctx context.Context, id string) error
	ListIntegrationConfigs(ctx context.Context, orgID uuid.UUID, serviceType *string, active *bool, limit, offset int) ([]*models.IntegrationConfig, int, error)
	
	// Bulk operations
	BulkCreateVehicleVerification(ctx context.Context, verifications []*models.VehicleVerification) error
	BulkCreateDriverVerification(ctx context.Context, verifications []*models.DriverVerification) error

	HealthCheck(ctx context.Context) error
	Ping(ctx context.Context) error
	CreateIntegrationLog(ctx context.Context, log *models.IntegrationLog) error

	// Compliance scan operations
	GetExpiringVehicleVerifications(ctx context.Context, days int) ([]*models.VehicleVerification, error)
	GetExpiringDriverVerifications(ctx context.Context, days int) ([]*models.DriverVerification, error)
	GetExpiringLocalPermits(ctx context.Context, days int) ([]*models.LocalPermit, error)
	GetNonCompliantGSTRecords(ctx context.Context) ([]*models.GSTCompliance, error)
}