package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"go.uber.org/zap"
)

// TripFinancialService handles financial calculations for trips
type TripFinancialService struct {
	logger *zap.Logger
	config *FinancialConfig
}

// FinancialConfig holds configuration for financial calculations
type FinancialConfig struct {
	BaseRatePerKM         float64 `json:"base_rate_per_km"`
	TollReimbursementRate float64 `json:"toll_reimbursement_rate"`
	FuelPricePerLiter     float64 `json:"fuel_price_per_liter"`
	DriverAllowancePerDay float64 `json:"driver_allowance_per_day"`
	DriverAllowancePerKM  float64 `json:"driver_allowance_per_km"`
	SLAPenaltyPerHour     float64 `json:"sla_penalty_per_hour"`
	GSTRate               float64 `json:"gst_rate"` // Percentage
	TDSRate               float64 `json:"tds_rate"` // Percentage
	TDSThreshold          float64 `json:"tds_threshold"`
	MinFuelEfficiency     float64 `json:"min_fuel_efficiency"` // km/l
}

// NewTripFinancialService creates a new financial service
func NewTripFinancialService(logger *zap.Logger, config *FinancialConfig) *TripFinancialService {
	// Set defaults if config is nil
	if config == nil {
		config = &FinancialConfig{
			BaseRatePerKM:         50.0,
			TollReimbursementRate: 1.0,
			FuelPricePerLiter:     95.0,
			DriverAllowancePerDay: 500.0,
			DriverAllowancePerKM:  2.0,
			SLAPenaltyPerHour:     200.0,
			GSTRate:               18.0,
			TDSRate:               2.0,
			TDSThreshold:          30000.0,
			MinFuelEfficiency:     6.0,
		}
	}

	return &TripFinancialService{
		logger: logger,
		config: config,
	}
}

// CalculateFinancials calculates all financial aspects of a completed trip
func (s *TripFinancialService) CalculateFinancials(ctx context.Context, trip *models.Trip, actualDistance, fuelUsed float64, plannedDuration, actualDuration int) (*models.TripFinancials, error) {
	if trip == nil {
		return nil, fmt.Errorf("trip is nil")
	}

	financials := &models.TripFinancials{
		TripID:       trip.ID,
		CalculatedAt: time.Now().UTC(),
	}

	// Calculate base fare
	baseFare := actualDistance * s.config.BaseRatePerKM
	financials.BillableDistance = actualDistance

	// Calculate toll cost (simplified - would integrate with FASTag in production)
	tollCost := actualDistance * s.config.TollReimbursementRate * 0.1 // Assume 10% of distance is toll roads
	financials.TollCost = tollCost

	// Calculate fuel cost
	fuelCost := fuelUsed * s.config.FuelPricePerLiter
	financials.FuelCost = fuelCost

	// Calculate driver allowance
	tripDays := float64(actualDuration) / (24 * 60) // Convert minutes to days
	if tripDays < 1 {
		tripDays = 1
	}
	driverAllowance := (s.config.DriverAllowancePerDay * tripDays) + (actualDistance * s.config.DriverAllowancePerKM)
	financials.DriverAllowance = driverAllowance

	// Calculate SLA penalty if delayed
	var slaPenalty float64
	if actualDuration > plannedDuration {
		delayHours := float64(actualDuration-plannedDuration) / 60.0
		slaPenalty = delayHours * s.config.SLAPenaltyPerHour
	}
	financials.SLAPenalty = slaPenalty

	// Calculate subtotal
	subtotal := baseFare + tollCost + fuelCost + driverAllowance - slaPenalty
	financials.Subtotal = subtotal

	// Determine if interstate (affects GST type)
	isInterstate := trip.GSTClassification == "interstate" || trip.GSTClassification == "IGST"

	// Apply GST
	gstAmount := subtotal * (s.config.GSTRate / 100.0)
	financials.GSTAmount = gstAmount
	financials.GSTRate = s.config.GSTRate

	// Split GST based on place of supply
	if isInterstate {
		financials.IGSTAmount = gstAmount
		financials.PlaceOfSupply = "Interstate"
	} else {
		financials.CGSTAmount = gstAmount / 2
		financials.SGSTAmount = gstAmount / 2
		financials.PlaceOfSupply = "Intrastate"
	}

	// Calculate total
	totalAmount := subtotal + gstAmount
	financials.TotalAmount = totalAmount

	// Check TDS applicability
	tdsApplicable := totalAmount > s.config.TDSThreshold
	financials.TDSApplicable = tdsApplicable
	if tdsApplicable {
		financials.TDSAmount = totalAmount * (s.config.TDSRate / 100.0)
	}

	// Determine if reverse charge applies (B2B vs B2C)
	financials.ReverseCharge = s.isReverseChargeApplicable(trip)

	s.logger.Info("Trip financials calculated",
		zap.String("trip_id", trip.ID),
		zap.Float64("total_amount", totalAmount),
		zap.Float64("gst_amount", gstAmount),
		zap.Bool("tds_applicable", tdsApplicable))

	return financials, nil
}

// isReverseChargeApplicable determines if reverse charge mechanism applies
func (s *TripFinancialService) isReverseChargeApplicable(trip *models.Trip) bool {
	// Simplified logic - in production, would check:
	// 1. Customer type (registered vs unregistered)
	// 2. Service type
	// 3. Place of supply
	// 4. GST classification

	if trip.GSTClassification == "interstate" {
		return true
	}

	return false
}

// CalculateFuelEfficiency calculates fuel efficiency metrics
func (s *TripFinancialService) CalculateFuelEfficiency(distance, fuelUsed float64) (float64, bool) {
	if fuelUsed <= 0 {
		return 0, false
	}

	efficiency := distance / fuelUsed
	isEfficient := efficiency >= s.config.MinFuelEfficiency

	return efficiency, isEfficient
}

// ValidateEWayBill validates E-way bill requirements
func (s *TripFinancialService) ValidateEWayBill(cargoValue float64, isInterstate bool) (bool, string) {
	// E-way bill required for:
	// - Interstate movement with value > 50,000 INR
	// - Intrastate movement with value > threshold (varies by state, typically 50,000-200,000 INR)

	intrastateThreshold := 50000.0

	if isInterstate && cargoValue > 50000.0 {
		return true, "E-way bill required for interstate movement"
	}

	if !isInterstate && cargoValue > intrastateThreshold {
		return true, "E-way bill required for intrastate movement"
	}

	return false, "E-way bill not required"
}

// GenerateInvoiceData prepares data for invoice generation
func (s *TripFinancialService) GenerateInvoiceData(financials *models.TripFinancials, customerGSTIN string) map[string]interface{} {
	invoiceData := map[string]interface{}{
		"trip_id":            financials.TripID,
		"invoice_date":       time.Now().UTC().Format("2006-01-02"),
		"billable_distance":  financials.BillableDistance,
		"base_fare":          financials.BillableDistance * s.config.BaseRatePerKM,
		"toll_charges":       financials.TollCost,
		"fuel_charges":       financials.FuelCost,
		"driver_allowance":   financials.DriverAllowance,
		"sla_penalty":        financials.SLAPenalty,
		"subtotal":           financials.Subtotal,
		"cgst_rate":          s.config.GSTRate / 2,
		"sgst_rate":          s.config.GSTRate / 2,
		"igst_rate":          0.0,
		"cgst_amount":        financials.CGSTAmount,
		"sgst_amount":        financials.SGSTAmount,
		"igst_amount":        financials.IGSTAmount,
		"total_amount":       financials.TotalAmount,
		"tds_applicable":     financials.TDSApplicable,
		"tds_amount":         financials.TDSAmount,
		"net_payable":        financials.TotalAmount - financials.TDSAmount,
		"customer_gstin":     customerGSTIN,
		"reverse_charge":     financials.ReverseCharge,
		"place_of_supply":    financials.PlaceOfSupply,
		"eway_bill_no":       financials.EWayBillNo,
	}

	return invoiceData
}
