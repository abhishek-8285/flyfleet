package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/flyfleet/india-integration-service/internal/models"
	"github.com/flyfleet/india-integration-service/internal/repository"
	"github.com/google/uuid"
)

// IndiaIntegrationService defines the service interface for India integrations
type IndiaIntegrationService interface {
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
	
	HealthCheck(ctx context.Context) error
}

// IndiaIntegrationServiceImpl implements the IndiaIntegrationService interface
type IndiaIntegrationServiceImpl struct {
	repo repository.IndiaIntegrationRepository
}

// NewIndiaIntegrationService creates a new instance of IndiaIntegrationService
func NewIndiaIntegrationService(repo repository.IndiaIntegrationRepository) *IndiaIntegrationServiceImpl {
	return &IndiaIntegrationServiceImpl{
		repo: repo,
	}
}

// Validation functions
func validateFASTagTransaction(transaction *models.FASTagTransaction) error {
	if transaction.TagNumber == "" {
		return errors.New("tag number is required")
	}
	if transaction.VehicleID == uuid.Nil {
		return errors.New("vehicle ID is required")
	}
	if transaction.DriverID == uuid.Nil {
		return errors.New("driver ID is required")
	}
	if transaction.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if transaction.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if transaction.TransactionTime.IsZero() {
		return errors.New("transaction time is required")
	}
	if transaction.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

func validateEWayBill(bill *models.EWayBill) error {
	if bill.EWayBillNumber == "" {
		return errors.New("e-way bill number is required")
	}
	if bill.DocumentNumber == "" {
		return errors.New("document number is required")
	}
	if bill.DocumentDate.IsZero() {
		return errors.New("document date is required")
	}
	if bill.FromPlace == "" {
		return errors.New("from place is required")
	}
	if bill.FromPincode == "" {
		return errors.New("from pincode is required")
	}
	if bill.ToPlace == "" {
		return errors.New("to place is required")
	}
	if bill.ToPincode == "" {
		return errors.New("to pincode is required")
	}
	if bill.GoodsValue <= 0 {
		return errors.New("goods value must be greater than zero")
	}
	return nil
}

func validateVehicleVerification(verification *models.VehicleVerification) error {
	if verification.RegistrationNumber == "" {
		return errors.New("registration number is required")
	}
	if verification.VehicleID == uuid.Nil {
		return errors.New("vehicle ID is required")
	}
	if verification.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if verification.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

func validateDriverVerification(verification *models.DriverVerification) error {
	if verification.LicenseNumber == "" {
		return errors.New("license number is required")
	}
	if verification.DriverID == uuid.Nil {
		return errors.New("driver ID is required")
	}
	if verification.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if verification.Status == "" {
		return errors.New("status is required")
	}
	if verification.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func validateLocalPermit(permit *models.LocalPermit) error {
	if permit.PermitNumber == "" {
		return errors.New("permit number is required")
	}
	if permit.VehicleID == uuid.Nil {
		return errors.New("vehicle ID is required")
	}
	if permit.DriverID == uuid.Nil {
		return errors.New("driver ID is required")
	}
	if permit.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if permit.ValidFrom.IsZero() || permit.ValidTo.IsZero() {
		return errors.New("valid from and valid to dates are required")
	}
	if permit.ValidTo.Before(permit.ValidFrom) {
		return errors.New("valid to date cannot be before valid from date")
	}
	if permit.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

func validateGSTCompliance(compliance *models.GSTCompliance) error {
	if compliance.GSTIN == "" {
		return errors.New("GSTIN is required")
	}
	if compliance.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if compliance.Status == "" {
		return errors.New("status is required")
	}
	if compliance.LegalName == "" {
		return errors.New("legal name is required")
	}
	return nil
}

func validateIntegrationConfig(config *models.IntegrationConfig) error {
	if config.ServiceType == "" {
		return errors.New("service type is required")
	}
	if config.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if config.APIEndpoint == "" {
		return errors.New("API endpoint is required")
	}
	return nil
}

// CreateFASTagTransaction creates a new FASTag transaction with validation
func (s *IndiaIntegrationServiceImpl) CreateFASTagTransaction(ctx context.Context, transaction *models.FASTagTransaction) error {
	if err := validateFASTagTransaction(transaction); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if transaction.ID == "" {
		transaction.ID = uuid.New().String()
	}
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now()
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.UpdatedAt = time.Now()
	}

	return s.repo.CreateFASTagTransaction(ctx, transaction)
}

// GetFASTagTransaction retrieves a FASTag transaction by ID
func (s *IndiaIntegrationServiceImpl) GetFASTagTransaction(ctx context.Context, id string) (*models.FASTagTransaction, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetFASTagTransaction(ctx, id)
}

// UpdateFASTagTransaction updates an existing FASTag transaction with validation
func (s *IndiaIntegrationServiceImpl) UpdateFASTagTransaction(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	// Perform partial validation based on updates
	if amount, ok := updates["amount"]; ok {
		if amountFloat, ok := amount.(float64); !ok || amountFloat <= 0 {
			return errors.New("amount must be a positive number")
		}
	}

	return s.repo.UpdateFASTagTransaction(ctx, id, updates)
}

// DeleteFASTagTransaction deletes a FASTag transaction
func (s *IndiaIntegrationServiceImpl) DeleteFASTagTransaction(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteFASTagTransaction(ctx, id)
}

// ListFASTagTransactions retrieves FASTag transactions with filtering
func (s *IndiaIntegrationServiceImpl) ListFASTagTransactions(ctx context.Context, vehicleID *uuid.UUID, tagNumber *string, startDate, endDate *time.Time, limit, offset int) ([]*models.FASTagTransaction, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListFASTagTransactions(ctx, vehicleID, tagNumber, startDate, endDate, limit, offset)
}

// CreateEWayBill creates a new E-Way Bill with validation
func (s *IndiaIntegrationServiceImpl) CreateEWayBill(ctx context.Context, bill *models.EWayBill) error {
	if err := validateEWayBill(bill); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if bill.ID == "" {
		bill.ID = uuid.New().String()
	}
	if bill.CreatedAt.IsZero() {
		bill.CreatedAt = time.Now()
	}
	if bill.UpdatedAt == nil || bill.UpdatedAt.IsZero() {
		updatedAt := time.Now()
		bill.UpdatedAt = &updatedAt
	}

	return s.repo.CreateEWayBill(ctx, bill)
}

// GetEWayBill retrieves an E-Way Bill by ID
func (s *IndiaIntegrationServiceImpl) GetEWayBill(ctx context.Context, id string) (*models.EWayBill, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetEWayBill(ctx, id)
}

// GetEWayBillByNumber retrieves an E-Way Bill by its number
func (s *IndiaIntegrationServiceImpl) GetEWayBillByNumber(ctx context.Context, number string) (*models.EWayBill, error) {
	if number == "" {
		return nil, errors.New("e-way bill number is required")
	}

	return s.repo.GetEWayBillByNumber(ctx, number)
}

// UpdateEWayBill updates an existing E-Way Bill with validation
func (s *IndiaIntegrationServiceImpl) UpdateEWayBill(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	// Perform partial validation based on updates
	if goodsValue, ok := updates["goods_value"]; ok {
		if valueFloat, ok := goodsValue.(float64); !ok || valueFloat <= 0 {
			return errors.New("goods value must be a positive number")
		}
	}

	return s.repo.UpdateEWayBill(ctx, id, updates)
}

// DeleteEWayBill deletes an E-Way Bill
func (s *IndiaIntegrationServiceImpl) DeleteEWayBill(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteEWayBill(ctx, id)
}

// ListEWayBills retrieves E-Way Bills with filtering
func (s *IndiaIntegrationServiceImpl) ListEWayBills(ctx context.Context, vehicleNumber *string, transporterID *string, startDate, endDate *time.Time, limit, offset int) ([]*models.EWayBill, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListEWayBills(ctx, vehicleNumber, transporterID, startDate, endDate, limit, offset)
}

// CreateVehicleVerification creates a new vehicle verification with validation
func (s *IndiaIntegrationServiceImpl) CreateVehicleVerification(ctx context.Context, verification *models.VehicleVerification) error {
	if err := validateVehicleVerification(verification); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if verification.ID == "" {
		verification.ID = uuid.New().String()
	}
	if verification.VerificationDate.IsZero() {
		verification.VerificationDate = time.Now()
	}
	if verification.CreatedAt.IsZero() {
		verification.CreatedAt = time.Now()
	}
	if verification.UpdatedAt.IsZero() {
		verification.UpdatedAt = time.Now()
	}

	return s.repo.CreateVehicleVerification(ctx, verification)
}

// GetVehicleVerification retrieves a vehicle verification by ID
func (s *IndiaIntegrationServiceImpl) GetVehicleVerification(ctx context.Context, id string) (*models.VehicleVerification, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetVehicleVerification(ctx, id)
}

// GetVehicleVerificationByRegNumber retrieves a vehicle verification by registration number
func (s *IndiaIntegrationServiceImpl) GetVehicleVerificationByRegNumber(ctx context.Context, regNumber string) (*models.VehicleVerification, error) {
	if regNumber == "" {
		return nil, errors.New("registration number is required")
	}

	return s.repo.GetVehicleVerificationByRegNumber(ctx, regNumber)
}

// UpdateVehicleVerification updates an existing vehicle verification with validation
func (s *IndiaIntegrationServiceImpl) UpdateVehicleVerification(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateVehicleVerification(ctx, id, updates)
}

// DeleteVehicleVerification deletes a vehicle verification
func (s *IndiaIntegrationServiceImpl) DeleteVehicleVerification(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteVehicleVerification(ctx, id)
}

// ListVehicleVerifications retrieves vehicle verifications with filtering
func (s *IndiaIntegrationServiceImpl) ListVehicleVerifications(ctx context.Context, regNumber *string, status *string, limit, offset int) ([]*models.VehicleVerification, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListVehicleVerifications(ctx, regNumber, status, limit, offset)
}

// CreateDriverVerification creates a new driver verification with validation
func (s *IndiaIntegrationServiceImpl) CreateDriverVerification(ctx context.Context, verification *models.DriverVerification) error {
	if err := validateDriverVerification(verification); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if verification.ID == "" {
		verification.ID = uuid.New().String()
	}
	if verification.VerificationDate.IsZero() {
		verification.VerificationDate = time.Now()
	}
	if verification.CreatedAt.IsZero() {
		verification.CreatedAt = time.Now()
	}
	if verification.UpdatedAt.IsZero() {
		verification.UpdatedAt = time.Now()
	}

	return s.repo.CreateDriverVerification(ctx, verification)
}

// GetDriverVerification retrieves a driver verification by ID
func (s *IndiaIntegrationServiceImpl) GetDriverVerification(ctx context.Context, id string) (*models.DriverVerification, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetDriverVerification(ctx, id)
}

// GetDriverVerificationByLicenseNumber retrieves a driver verification by license number
func (s *IndiaIntegrationServiceImpl) GetDriverVerificationByLicenseNumber(ctx context.Context, licenseNumber string) (*models.DriverVerification, error) {
	if licenseNumber == "" {
		return nil, errors.New("license number is required")
	}

	return s.repo.GetDriverVerificationByLicenseNumber(ctx, licenseNumber)
}

// UpdateDriverVerification updates an existing driver verification with validation
func (s *IndiaIntegrationServiceImpl) UpdateDriverVerification(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateDriverVerification(ctx, id, updates)
}

// DeleteDriverVerification deletes a driver verification
func (s *IndiaIntegrationServiceImpl) DeleteDriverVerification(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteDriverVerification(ctx, id)
}

// ListDriverVerifications retrieves driver verifications with filtering
func (s *IndiaIntegrationServiceImpl) ListDriverVerifications(ctx context.Context, licenseNumber *string, status *string, limit, offset int) ([]*models.DriverVerification, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListDriverVerifications(ctx, licenseNumber, status, limit, offset)
}

// CreateLocalPermit creates a new local permit with validation
func (s *IndiaIntegrationServiceImpl) CreateLocalPermit(ctx context.Context, permit *models.LocalPermit) error {
	if err := validateLocalPermit(permit); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if permit.ID == "" {
		permit.ID = uuid.New().String()
	}
	if permit.CreatedAt.IsZero() {
		permit.CreatedAt = time.Now()
	}
	if permit.UpdatedAt.IsZero() {
		permit.UpdatedAt = time.Now()
	}

	return s.repo.CreateLocalPermit(ctx, permit)
}

// GetLocalPermit retrieves a local permit by ID
func (s *IndiaIntegrationServiceImpl) GetLocalPermit(ctx context.Context, id string) (*models.LocalPermit, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetLocalPermit(ctx, id)
}

// GetLocalPermitByNumber retrieves a local permit by permit number
func (s *IndiaIntegrationServiceImpl) GetLocalPermitByNumber(ctx context.Context, permitNumber string) (*models.LocalPermit, error) {
	if permitNumber == "" {
		return nil, errors.New("permit number is required")
	}

	return s.repo.GetLocalPermitByNumber(ctx, permitNumber)
}

// UpdateLocalPermit updates an existing local permit with validation
func (s *IndiaIntegrationServiceImpl) UpdateLocalPermit(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	// Perform partial validation based on updates
	if validFrom, ok := updates["valid_from"]; ok {
		if fromTime, ok := validFrom.(time.Time); ok {
			if validTo, ok := updates["valid_to"]; ok {
				if toTime, ok := validTo.(time.Time); ok && toTime.Before(fromTime) {
					return errors.New("valid to date cannot be before valid from date")
				}
			}
		}
	}

	return s.repo.UpdateLocalPermit(ctx, id, updates)
}

// DeleteLocalPermit deletes a local permit
func (s *IndiaIntegrationServiceImpl) DeleteLocalPermit(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteLocalPermit(ctx, id)
}

// ListLocalPermits retrieves local permits with filtering
func (s *IndiaIntegrationServiceImpl) ListLocalPermits(ctx context.Context, vehicleID *uuid.UUID, permitType *string, status *string, limit, offset int) ([]*models.LocalPermit, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListLocalPermits(ctx, vehicleID, permitType, status, limit, offset)
}

// CreateGSTCompliance creates a new GST compliance record with validation
func (s *IndiaIntegrationServiceImpl) CreateGSTCompliance(ctx context.Context, compliance *models.GSTCompliance) error {
	if err := validateGSTCompliance(compliance); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if compliance.ID == "" {
		compliance.ID = uuid.New().String()
	}
	if compliance.LastUpdated.IsZero() {
		compliance.LastUpdated = time.Now()
	}
	if compliance.CreatedAt.IsZero() {
		compliance.CreatedAt = time.Now()
	}
	if compliance.UpdatedAt.IsZero() {
		compliance.UpdatedAt = time.Now()
	}

	return s.repo.CreateGSTCompliance(ctx, compliance)
}

// GetGSTCompliance retrieves a GST compliance record by ID
func (s *IndiaIntegrationServiceImpl) GetGSTCompliance(ctx context.Context, id string) (*models.GSTCompliance, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetGSTCompliance(ctx, id)
}

// GetGSTComplianceByGSTIN retrieves a GST compliance record by GSTIN
func (s *IndiaIntegrationServiceImpl) GetGSTComplianceByGSTIN(ctx context.Context, gstin string) (*models.GSTCompliance, error) {
	if gstin == "" {
		return nil, errors.New("GSTIN is required")
	}

	return s.repo.GetGSTComplianceByGSTIN(ctx, gstin)
}

// UpdateGSTCompliance updates an existing GST compliance record with validation
func (s *IndiaIntegrationServiceImpl) UpdateGSTCompliance(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateGSTCompliance(ctx, id, updates)
}

// DeleteGSTCompliance deletes a GST compliance record
func (s *IndiaIntegrationServiceImpl) DeleteGSTCompliance(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteGSTCompliance(ctx, id)
}

// ListGSTCompliances retrieves GST compliance records with filtering
func (s *IndiaIntegrationServiceImpl) ListGSTCompliances(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]*models.GSTCompliance, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListGSTCompliances(ctx, orgID, status, limit, offset)
}

// CreateIntegrationConfig creates a new integration configuration with validation
func (s *IndiaIntegrationServiceImpl) CreateIntegrationConfig(ctx context.Context, config *models.IntegrationConfig) error {
	if err := validateIntegrationConfig(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	if config.UpdatedAt.IsZero() {
		config.UpdatedAt = time.Now()
	}

	return s.repo.CreateIntegrationConfig(ctx, config)
}

// GetIntegrationConfig retrieves an integration configuration by ID
func (s *IndiaIntegrationServiceImpl) GetIntegrationConfig(ctx context.Context, id string) (*models.IntegrationConfig, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetIntegrationConfig(ctx, id)
}

// GetIntegrationConfigByServiceType retrieves an integration configuration by service type and organization
func (s *IndiaIntegrationServiceImpl) GetIntegrationConfigByServiceType(ctx context.Context, orgID uuid.UUID, serviceType string) (*models.IntegrationConfig, error) {
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID is required")
	}
	if serviceType == "" {
		return nil, errors.New("service type is required")
	}

	return s.repo.GetIntegrationConfigByServiceType(ctx, orgID, serviceType)
}

// UpdateIntegrationConfig updates an existing integration configuration with validation
func (s *IndiaIntegrationServiceImpl) UpdateIntegrationConfig(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateIntegrationConfig(ctx, id, updates)
}

// DeleteIntegrationConfig deletes an integration configuration
func (s *IndiaIntegrationServiceImpl) DeleteIntegrationConfig(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteIntegrationConfig(ctx, id)
}

// ListIntegrationConfigs retrieves integration configurations with filtering
func (s *IndiaIntegrationServiceImpl) ListIntegrationConfigs(ctx context.Context, orgID uuid.UUID, serviceType *string, active *bool, limit, offset int) ([]*models.IntegrationConfig, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListIntegrationConfigs(ctx, orgID, serviceType, active, limit, offset)
}

// HealthCheck performs a health check on the service and its dependencies
func (s *IndiaIntegrationServiceImpl) HealthCheck(ctx context.Context) error {
	log.Println("Performing health check for India Integration Service")
	
	// Check repository/database health
	if err := s.repo.HealthCheck(ctx); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}
	
	return nil
}