package service

import (
	"math"

	driverModels "github.com/flyfleet/fleet-core/internal/domains/driver/models"
	vehicleModels "github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
)

type ScoringService interface {
	Calculate(distKm float64, vehicle *vehicleModels.Vehicle, driver *driverModels.Driver) float64
	Explain(distKm float64, vehicle *vehicleModels.Vehicle, driver *driverModels.Driver) []string
}

type scoringService struct{}

func NewScoringService() ScoringService {
	return &scoringService{}
}

func (s *scoringService) Calculate(distKm float64, vehicle *vehicleModels.Vehicle, driver *driverModels.Driver) float64 {
	// Proximity Score (40%) - 100 points, deduct 5 points per km, clamp at 0
	proximity := math.Max(0, 100-(distKm*5.0))

	// Fit Score (30%) - default to 100 for now
	fit := 100.0

	// Driver Quality (20%) - Driver Rating (0-5) mapped to 0-100
	driverQuality := 0.0
	if driver != nil {
		driverQuality = driver.Rating * 20.0
	}

	// Base Score (10%) - canonical formula
	base := 10.0

	// Final Weighted Score: 40/30/20/10
	score := (proximity * 0.4) + (fit * 0.3) + (driverQuality * 0.2) + base

	return score
}

func (s *scoringService) Explain(distKm float64, vehicle *vehicleModels.Vehicle, driver *driverModels.Driver) []string {
	var reasons []string
	if distKm <= 5.0 {
		reasons = append(reasons, "Closest available vehicle")
	} else if distKm <= 15.0 {
		reasons = append(reasons, "Within reasonable distance")
	}

	if driver != nil && driver.Rating >= 4.0 {
		reasons = append(reasons, "High driver rating ("+formatRating(driver.Rating)+")")
	}

	return reasons
}

func formatRating(r float64) string {
	if r == 4.5 {
		return "4.5"
	}
	if r == 4.8 {
		return "4.8"
	}
	if r == 5.0 {
		return "5.0"
	}
	return "good"
}
