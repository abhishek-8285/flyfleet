package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaxService struct {
	repo   *repository.TaxRepository
	logger *zap.Logger
}

func NewTaxService(repo *repository.TaxRepository, logger *zap.Logger) *TaxService {
	return &TaxService{
		repo:   repo,
		logger: logger,
	}
}

// CalculateInvoiceTaxes computes the tax breakdown for an invoice based on region
func (s *TaxService) CalculateInvoiceTaxes(ctx context.Context, invoice *models.Invoice, countryCode string, stateCode *string) ([]models.InvoiceTaxLine, error) {
	region, err := s.repo.GetRegionByCountry(ctx, countryCode, stateCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax region: %w", err)
	}
	if region == nil {
		return nil, nil // No specific tax region found, assuming 0 tax or fallback
	}

	rates, err := s.repo.GetTaxRatesByRegion(ctx, region.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax rates: %w", err)
	}

	var taxLines []models.InvoiceTaxLine
	for _, rate := range rates {
		line := models.InvoiceTaxLine{
			ID:         uuid.New(),
			InvoiceID:  invoice.ID,
			TaxRateID:  rate.ID,
			Name:       rate.Name,
			Percentage: rate.Percentage,
			Amount:     invoice.Subtotal * (rate.Percentage / 100),
		}
		taxLines = append(taxLines, line)
	}

	return taxLines, nil
}

type SequenceService struct {
	repo   *repository.SequenceRepository
	logger *zap.Logger
}

func NewSequenceService(repo *repository.SequenceRepository, logger *zap.Logger) *SequenceService {
	return &SequenceService{
		repo:   repo,
		logger: logger,
	}
}

// GetNextInvoiceNumber increments and returns the next sequential invoice number
func (s *SequenceService) GetNextInvoiceNumber(ctx context.Context, orgID uuid.UUID) (string, error) {
	return s.repo.GetNextInvoiceNumber(ctx, orgID)
}
