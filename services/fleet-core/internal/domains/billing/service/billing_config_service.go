package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BillingConfigService handles billing configuration business logic
type BillingConfigService struct {
	repo   *repository.BillingConfigRepository
	logger *zap.Logger
}

// NewBillingConfigService creates a new service
func NewBillingConfigService(repo *repository.BillingConfigRepository, logger *zap.Logger) *BillingConfigService {
	return &BillingConfigService{
		repo:   repo,
		logger: logger,
	}
}

// GetConfig gets billing configuration for an organization
func (s *BillingConfigService) GetConfig(ctx context.Context, tenantID uuid.UUID) (*models.BillingConfiguration, error) {
	config, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Return default if not found
	if config == nil {
		config = models.DefaultBillingConfig(tenantID)
	}

	return config, nil
}

// GetOrCreateConfig gets or creates billing configuration
func (s *BillingConfigService) GetOrCreateConfig(ctx context.Context, tenantID, userID uuid.UUID) (*models.BillingConfiguration, error) {
	return s.repo.GetOrCreate(ctx, tenantID, userID)
}

// UpdateConfig updates billing configuration
func (s *BillingConfigService) UpdateConfig(ctx context.Context, tenantID, userID uuid.UUID, req *models.BillingConfigRequest) (*models.BillingConfiguration, error) {
	// Get existing or create new
	config, err := s.repo.GetOrCreate(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	config.CompanyName = req.CompanyName
	config.CompanyAddress = req.CompanyAddress
	config.CompanyGSTIN = req.CompanyGSTIN
	config.CompanyPAN = req.CompanyPAN
	config.CompanyEmail = req.CompanyEmail
	config.CompanyPhone = req.CompanyPhone
	config.CompanyWebsite = req.CompanyWebsite
	config.CompanyLogoURL = req.CompanyLogoURL
	config.PrimaryColor = req.PrimaryColor
	config.SecondaryColor = req.SecondaryColor
	config.Theme = req.Theme
	config.BankName = req.BankName
	config.BankAccountName = req.BankAccountName
	config.BankAccountNumber = req.BankAccountNumber
	config.BankIFSC = req.BankIFSC
	config.BankBranch = req.BankBranch
	config.BankUPI = req.BankUPI
	config.InvoicePrefix = req.InvoicePrefix
	config.InvoiceNotes = req.InvoiceNotes
	config.DefaultCurrency = req.DefaultCurrency
	config.DefaultTaxCalculation = req.DefaultTaxCalculation
	config.DefaultPaymentTermsDays = req.DefaultPaymentTermsDays
	config.UpdatedBy = userID

	// Handle boolean pointers
	if req.ShowHSN != nil {
		config.ShowHSN = *req.ShowHSN
	}
	if req.ShowQRCode != nil {
		config.ShowQRCode = *req.ShowQRCode
	}
	if req.ShowPaymentInfo != nil {
		config.ShowPaymentInfo = *req.ShowPaymentInfo
	}

	// Handle terms
	if req.InvoiceTerms != nil {
		config.InvoiceTerms = req.InvoiceTerms
	}

	// Save
	if err := s.repo.Update(ctx, config); err != nil {
		return nil, err
	}

	s.logger.Info("Billing configuration updated",
		zap.String("tenant_id", tenantID.String()),
		zap.String("company_name", config.CompanyName),
	)

	return config, nil
}

// DeleteConfig soft-deletes billing configuration
func (s *BillingConfigService) DeleteConfig(ctx context.Context, tenantID uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID)
}

// ValidateConfig validates billing configuration
func (s *BillingConfigService) ValidateConfig(config *models.BillingConfigRequest) error {
	// Add custom validation logic here
	if config.CompanyGSTIN != "" && len(config.CompanyGSTIN) != 15 {
		return &ValidationError{Field: "company_gstin", Message: "GSTIN must be 15 characters"}
	}

	if config.CompanyPAN != "" && len(config.CompanyPAN) != 10 {
		return &ValidationError{Field: "company_pan", Message: "PAN must be 10 characters"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
