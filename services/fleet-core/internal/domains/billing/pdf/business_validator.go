package pdf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
)

// BusinessValidator handles industrial and legal compliance rules
type BusinessValidator struct {
	CompanyStateCode string // e.g., "27"
}

// ValidationResult carries the audit results
type ValidationResult struct {
	IsValid bool
	Errors  []string
}

func NewBusinessValidator(companyState string) *BusinessValidator {
	return &BusinessValidator{
		CompanyStateCode: companyState,
	}
}

// ValidateInvoice performs a 4-tier legal and math audit
func (v *BusinessValidator) ValidateInvoice(invoice *models.Invoice, items []models.InvoiceItem) ValidationResult {
	result := ValidationResult{IsValid: true}

	// 1. Legal Threshold: E-Way Bill Check
	if invoice.TotalAmount > 50000 {
		eway, ok := invoice.Metadata["eway_bill"].(string)
		if !ok || len(strings.TrimSpace(eway)) < 12 {
			result.Errors = append(result.Errors, "LEGAL ERROR: E-Way Bill is mandatory for amounts over ₹50,000")
		}
	}

	// 2. GSTIN Format Check
	gstinRegex := regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
	// We assume GSTIN is in metadata or customer record
	if gstin, ok := invoice.Metadata["customer_gstin"].(string); ok {
		if !gstinRegex.MatchString(gstin) {
			result.Errors = append(result.Errors, fmt.Sprintf("FORMAT ERROR: Invalid Customer GSTIN format '%s'", gstin))
		}
	}

	// 3. Tax Type Consistency Audit
	pos, _ := invoice.Metadata["place_of_supply"].(string)
	if pos != "" {
		customerState := ""
		if len(pos) >= 2 {
			customerState = pos[len(pos)-3 : len(pos)-1] // Extracts "09" from "UP (09)"
		}

		// Simplified check: If states differ, IGST should be > 0
		if customerState != "" && customerState != v.CompanyStateCode {
			// This should be INTER-STATE
			for _, item := range items {
				if item.TaxPercent > 0 && invoice.Metadata["tax_type"] == "CGST" {
					result.Errors = append(result.Errors, "COMPLIANCE ERROR: Inter-state supply detected. Must use IGST instead of CGST/SGST")
					break
				}
			}
		}
	}

	// 4. Mathematical Integrity Check
	var calculatedTotal float64
	for _, item := range items {
		sub := float64(item.Quantity) * item.UnitPrice
		tax := sub * (item.TaxPercent / 100)
		calculatedTotal += (sub + tax)
	}

	// Allow for tiny rounding differences (0.01)
	if diff := calculatedTotal - invoice.TotalAmount; diff > 0.05 || diff < -0.05 {
		result.Errors = append(result.Errors, fmt.Sprintf("MATH ERROR: Invoice total (₹%.2f) does not match calculated total (₹%.2f)", invoice.TotalAmount, calculatedTotal))
	}

	if len(result.Errors) > 0 {
		result.IsValid = false
	}

	return result
}
