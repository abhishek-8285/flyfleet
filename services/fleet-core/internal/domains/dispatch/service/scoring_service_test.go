package service

import (
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/stretchr/testify/assert"
)

func TestScoringService_Calculate(t *testing.T) {
	svc := &ScoringService{}

	tests := []struct {
		name         string
		distKm       float64
		vehicle      *models.Vehicle
		driver       *models.Driver
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:   "close vehicle high rating driver",
			distKm: 2.0,
			vehicle: &models.Vehicle{
				Status:          "ACTIVE",
				CapacityTonnage: 1.5,
			},
			driver: &models.Driver{
				Rating: 4.8,
			},
			expectedMin: 85.0,
			expectedMax: 95.0,
		},
		{
			name:   "far vehicle low rating driver",
			distKm: 18.0,
			vehicle: &models.Vehicle{
				Status:          "ACTIVE",
				CapacityTonnage: 3.0,
			},
			driver: &models.Driver{
				Rating: 2.0,
			},
			expectedMin: 40.0,
			expectedMax: 55.0,
		},
		{
			name:   "zero distance max rating",
			distKm: 0.0,
			vehicle: &models.Vehicle{
				Status: "ACTIVE",
			},
			driver: &models.Driver{
				Rating: 5.0,
			},
			expectedMin: 99.0,
			expectedMax: 101.0,
		},
		{
			name:   "zero distance zero rating",
			distKm: 0.0,
			vehicle: &models.Vehicle{
				Status: "ACTIVE",
			},
			driver: &models.Driver{
				Rating: 0.0,
			},
			expectedMin: 69.0,
			expectedMax: 71.0,
		},
		{
			name:   "distance 20km scores zero proximity",
			distKm: 20.0,
			vehicle: &models.Vehicle{
				Status: "ACTIVE",
			},
			driver: &models.Driver{
				Rating: 3.0,
			},
			expectedMin: 48.0,
			expectedMax: 52.0,
		},
		{
			name:   "distance beyond 20km proximity clamped to 0",
			distKm: 25.0,
			vehicle: &models.Vehicle{
				Status: "ACTIVE",
			},
			driver: &models.Driver{
				Rating: 3.0,
			},
			expectedMin: 38.0,
			expectedMax: 42.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.Calculate(tt.distKm, tt.vehicle, tt.driver)
			assert.GreaterOrEqual(t, score, tt.expectedMin,
				"score %.1f below expected minimum %.1f", score, tt.expectedMin)
			assert.LessOrEqual(t, score, tt.expectedMax,
				"score %.1f above expected maximum %.1f", score, tt.expectedMax)
		})
	}
}

func TestScoringService_Explain(t *testing.T) {
	svc := &ScoringService{}

	t.Run("close vehicle with high rating", func(t *testing.T) {
		reasons := svc.Explain(3.0, &models.Vehicle{}, &models.Driver{Rating: 4.5})
		assert.Contains(t, reasons, "Closest available vehicle")
		assert.Contains(t, reasons, "High driver rating (4.5)")
	})

	t.Run("medium distance no rating bonus", func(t *testing.T) {
		reasons := svc.Explain(12.0, &models.Vehicle{}, &models.Driver{Rating: 2.0})
		assert.Contains(t, reasons, "Within reasonable distance")
		assert.NotContains(t, reasons, "High driver rating")
	})

	t.Run("far distance low rating", func(t *testing.T) {
		reasons := svc.Explain(30.0, &models.Vehicle{}, &models.Driver{Rating: 1.0})
		assert.Empty(t, reasons)
	})
}

func TestScoringService_WeightedFormula(t *testing.T) {
	svc := &ScoringService{}

	// Verify the 40/30/20/10 weighting explicitly
	score := svc.Calculate(10.0, &models.Vehicle{}, &models.Driver{Rating: 3.0})
	// proximity = max(0, 100 - 10*5) = 50
	// fit = 100
	// driver = 3.0 * 20 = 60
	// score = 50*0.4 + 100*0.3 + 60*0.2 + 10 = 20 + 30 + 12 + 10 = 72
	assert.InDelta(t, 72.0, score, 0.1,
		"weighted formula: 40%% proximity + 30%% fit + 20%% driver + 10%% base")
}
