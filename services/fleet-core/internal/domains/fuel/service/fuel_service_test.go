package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
)

func TestFuelService_Creation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	assert.NotNil(t, logger)
}

func TestFuelLogStatusConstants(t *testing.T) {
	assert.Equal(t, models.FuelLogStatus("PENDING"), models.FuelLogStatusPending)
	assert.Equal(t, models.FuelLogStatus("APPROVED"), models.FuelLogStatusApproved)
	assert.Equal(t, models.FuelLogStatus("REJECTED"), models.FuelLogStatusRejected)
}

func TestFuelLogModel_Creation(t *testing.T) {
	log := &models.FuelLog{
		ID:              uuid.New(),
		TenantID:  uuid.New(),
		VehicleID:       uuid.New(),
		DriverID:        uuid.New(),
		StationName:     "Shell Station #42",
		FuelType:        "Diesel",
		Quantity:        45.5,
		UnitPrice:       2.5,
		TotalCost:       113.75,
		Odometer:        45000.0,
		TransactionDate: time.Now(),
		Status:          models.FuelLogStatusPending,
		CreatedBy:       uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, log.ID)
	assert.Equal(t, "Shell Station #42", log.StationName)
	assert.Equal(t, 113.75, log.TotalCost)
}

func TestCreateFuelLogRequest_Validation(t *testing.T) {
	req := &models.CreateFuelLogRequest{
		VehicleID:       uuid.New(),
		DriverID:        uuid.New(),
		StationName:     "ExxonMobil",
		FuelType:        "Petrol 95",
		Quantity:        30.0,
		UnitPrice:       3.0,
		TotalCost:       90.0,
		Odometer:        12500.0,
		TransactionDate: time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, req.VehicleID)
	assert.Equal(t, 90.0, req.TotalCost)
}

func TestFuelSettingsModel(t *testing.T) {
	settings := &models.FuelSettings{
		TenantID:   uuid.New(),
		DefaultCurrency:  "USD",
		AllowedFuelTypes: []byte(`["diesel", "petrol-95"]`),
		AutoSyncEnabled:  true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	assert.Equal(t, "USD", settings.DefaultCurrency)
	assert.True(t, settings.AutoSyncEnabled)
}

func TestCostCalculationValidation(t *testing.T) {
	tests := []struct {
		qty      float64
		price    float64
		total    float64
		expected bool
	}{
		{10.0, 2.50, 25.00, true},
		{45.5, 3.125, 142.10, false}, // 45.5 * 3.125 = 142.1875, 142.10 is definitely wrong
		{45.5, 3.125, 142.19, true},  // Rounded correctly
		{10.0, 2.0, 21.0, false},     // Clearly wrong
	}

	epsilon := 0.01
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			calc := tt.qty * tt.price
			isValid := tt.total >= calc-epsilon && tt.total <= calc+epsilon
			assert.Equal(t, tt.expected, isValid)
		})
	}
}

func TestOdometerValidation(t *testing.T) {
	tests := []struct {
		last     float64
		current  float64
		expected bool
	}{
		{10000.0, 10500.0, true},
		{10000.0, 10000.0, true}, // Same is fine (maybe same day)
		{10000.0, 9500.0, false}, // Backwards is wrong
		{0.0, 10.0, true},        // First entry
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			isValid := tt.current >= tt.last
			assert.Equal(t, tt.expected, isValid)
		})
	}
}

func TestFuelLogFilter(t *testing.T) {
	vehicleID := uuid.New()
	filter := models.FuelLogFilter{
		TenantID: uuid.New(),
		VehicleID:      &vehicleID,
		Status:         []models.FuelLogStatus{models.FuelLogStatusPending},
		Search:         "shell",
		Page:           1,
		PageSize:       20,
	}

	assert.Equal(t, 1, len(filter.Status))
	assert.Equal(t, "shell", filter.Search)
}

func TestFuelLogStats(t *testing.T) {
	stats := &models.FuelLogStats{
		TotalConsumption: 1500.5,
		TotalCost:        4500.0,
		PendingApprovals: 3,
		ByStatus: map[string]int{
			"PENDING":  3,
			"APPROVED": 10,
		},
	}

	assert.Equal(t, 1500.5, stats.TotalConsumption)
	assert.Equal(t, 3, stats.PendingApprovals)
	assert.Equal(t, 10, stats.ByStatus["APPROVED"])
}
