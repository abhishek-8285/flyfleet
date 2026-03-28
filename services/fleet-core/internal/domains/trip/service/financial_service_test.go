//go:build integration
// +build integration

package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTripFinancialService_CalculateFinancials_Success(t *testing.T) {
	logger := zap.NewNop()
	config := &FinancialConfig{
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

	service := NewTripFinancialService(logger, config)
	ctx := context.Background()

	trip := &models.Trip{
		ID:                "trip-123",
		GSTClassification: "intrastate",
	}

	financials, err := service.CalculateFinancials(ctx, trip, 500.0, 50.0, 480, 500)

	assert.NoError(t, err)
	assert.NotNil(t, financials)
	assert.Equal(t, 500.0, financials.BillableDistance)
	assert.Greater(t, financials.TotalAmount, 0.0)
	assert.Greater(t, financials.GSTAmount, 0.0)
}

func TestTripFinancialService_CalculateFinancials_Interstate(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)
	ctx := context.Background()

	trip := &models.Trip{
		ID:                "trip-123",
		GSTClassification: "interstate",
	}

	financials, err := service.CalculateFinancials(ctx, trip, 500.0, 50.0, 480, 500)

	assert.NoError(t, err)
	assert.NotNil(t, financials)
	assert.Greater(t, financials.IGSTAmount, 0.0)
	assert.Equal(t, 0.0, financials.CGSTAmount)
	assert.Equal(t, 0.0, financials.SGSTAmount)
	assert.Equal(t, "Interstate", financials.PlaceOfSupply)
}

func TestTripFinancialService_CalculateFinancials_Intrastate(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)
	ctx := context.Background()

	trip := &models.Trip{
		ID:                "trip-123",
		GSTClassification: "intrastate",
	}

	financials, err := service.CalculateFinancials(ctx, trip, 500.0, 50.0, 480, 500)

	assert.NoError(t, err)
	assert.NotNil(t, financials)
	assert.Greater(t, financials.CGSTAmount, 0.0)
	assert.Greater(t, financials.SGSTAmount, 0.0)
	assert.Equal(t, 0.0, financials.IGSTAmount)
	assert.Equal(t, "Intrastate", financials.PlaceOfSupply)
}

func TestTripFinancialService_CalculateFinancials_WithDelay(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)
	ctx := context.Background()

	trip := &models.Trip{
		ID: "trip-123",
	}

	// 2 hours delay (120 minutes)
	financials, err := service.CalculateFinancials(ctx, trip, 500.0, 50.0, 480, 600)

	assert.NoError(t, err)
	assert.NotNil(t, financials)
	assert.Greater(t, financials.SLAPenalty, 0.0)
}

func TestTripFinancialService_CalculateFinancials_TDS(t *testing.T) {
	logger := zap.NewNop()
	config := &FinancialConfig{
		BaseRatePerKM: 100.0,
		TDSThreshold:  30000.0,
		TDSRate:       2.0,
	}
	service := NewTripFinancialService(logger, config)
	ctx := context.Background()

	trip := &models.Trip{
		ID: "trip-123",
	}

	// High value trip (> 30,000 INR)
	financials, err := service.CalculateFinancials(ctx, trip, 1000.0, 100.0, 480, 500)

	assert.NoError(t, err)
	assert.NotNil(t, financials)
	assert.True(t, financials.TDSApplicable)
	assert.Greater(t, financials.TDSAmount, 0.0)
}

func TestTripFinancialService_CalculateFuelEfficiency(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)

	efficiency, isEfficient := service.CalculateFuelEfficiency(300.0, 50.0)

	assert.Equal(t, 6.0, efficiency)
	assert.True(t, isEfficient)

	// Low efficiency
	efficiency2, isEfficient2 := service.CalculateFuelEfficiency(200.0, 50.0)

	assert.Equal(t, 4.0, efficiency2)
	assert.False(t, isEfficient2)

	// Zero fuel
	efficiency3, isEfficient3 := service.CalculateFuelEfficiency(100.0, 0.0)

	assert.Equal(t, 0.0, efficiency3)
	assert.False(t, isEfficient3)
}

func TestTripFinancialService_ValidateEWayBill(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)

	// Interstate > 50,000
	required, reason := service.ValidateEWayBill(60000.0, true)
	assert.True(t, required)
	assert.Contains(t, reason, "interstate")

	// Interstate < 50,000
	_, _ = service.ValidateEWayBill(40000.0, true)

	// Intrastate > 50,000
	required3, reason3 := service.ValidateEWayBill(60000.0, false)
	assert.True(t, required3)
	assert.Contains(t, reason3, "intrastate")

	// Intrastate < 50,000
	_, _ = service.ValidateEWayBill(40000.0, false)
}

func TestTripFinancialService_GenerateInvoiceData(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)
	ctx := context.Background()

	trip := &models.Trip{
		ID:                "trip-123",
		GSTClassification: "interstate",
	}

	financials, _ := service.CalculateFinancials(ctx, trip, 500.0, 50.0, 480, 500)
	invoiceData := service.GenerateInvoiceData(financials, "27ABCDE1234F1Z5")

	assert.NotNil(t, invoiceData)
	assert.Equal(t, "trip-123", invoiceData["trip_id"])
	assert.Equal(t, "27ABCDE1234F1Z5", invoiceData["customer_gstin"])
	assert.Contains(t, invoiceData, "total_amount")
	// For interstate, IGST is used instead of GST
	assert.Contains(t, invoiceData, "igst_amount")
	assert.Contains(t, invoiceData, "tds_amount")
}

func TestTripFinancialService_DefaultConfig(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)

	assert.NotNil(t, service.config)
	assert.Greater(t, service.config.BaseRatePerKM, 0.0)
	assert.Greater(t, service.config.GSTRate, 0.0)
}

func TestTripFinancialService_IsReverseChargeApplicable(t *testing.T) {
	logger := zap.NewNop()
	service := NewTripFinancialService(logger, nil)
	ctx := context.Background()

	// Interstate
	trip1 := &models.Trip{
		ID:                "trip-123",
		GSTClassification: "interstate",
	}
	financials1, _ := service.CalculateFinancials(ctx, trip1, 500.0, 50.0, 480, 500)
	assert.True(t, financials1.ReverseCharge)

	// Intrastate
	trip2 := &models.Trip{
		ID:                "trip-456",
		GSTClassification: "intrastate",
	}
	financials2, _ := service.CalculateFinancials(ctx, trip2, 500.0, 50.0, 480, 500)
	assert.False(t, financials2.ReverseCharge)
}
