package service

import (
	"regexp"
	"strings"
)

// TaxCalculationService handles GST calculations and validation for Indian operations
type TaxCalculationService struct{}

func NewTaxCalculationService() *TaxCalculationService {
	return &TaxCalculationService{}
}

// GSTResult holds the components of GST
type GSTResult struct {
	TotalAmount float64
	GSTAmount   float64
	CGST        float64
	SGST        float64
	IGST        float64
	BaseAmount  float64
}

// CalculateGST calculates GST components based on amount and rate.
// CGST/SGST/IGST depends on whether it's intra-state or inter-state.
// Defaulting to 50/50 split for CGST/SGST if no specific flag is provided for IGST.
func (s *TaxCalculationService) CalculateGST(amount float64, rate float64, isInterState bool) GSTResult {
	gstAmount := (amount * rate) / 100
	baseAmount := amount

	result := GSTResult{
		TotalAmount: amount + gstAmount,
		GSTAmount:   gstAmount,
		BaseAmount:  baseAmount,
	}

	if isInterState {
		result.IGST = gstAmount
	} else {
		result.CGST = gstAmount / 2
		result.SGST = gstAmount / 2
	}

	return result
}

// ValidateGSTIN validates the format of an Indian GSTIN
// Format: 2 digits (state code), 10 chars (PAN), 1 char (entity code), 1 char ('Z'), 1 char (check digit)
func (s *TaxCalculationService) ValidateGSTIN(gstin string) bool {
	gstin = strings.ToUpper(strings.TrimSpace(gstin))
	if len(gstin) != 15 {
		return false
	}

	// Basic regex for GSTIN
	regex := `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`
	match, _ := regexp.MatchString(regex, gstin)
	return match
}
