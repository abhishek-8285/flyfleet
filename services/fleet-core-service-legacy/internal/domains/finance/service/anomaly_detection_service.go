package service

import (
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	"go.uber.org/zap"
)

// AnomalyDetectionConfig holds anomaly detection configuration
type AnomalyDetectionConfig struct {
	MaxPricePerLiter      float64 // Maximum expected price per liter
	MinPricePerLiter      float64 // Minimum expected price per liter
	MaxLitersPerFill      float64 // Maximum liters per fill-up
	MaxDistancePerDay     float64 // Maximum expected distance per day
	PriceVarianceThreshold float64 // Price variance threshold for anomaly
}

// DefaultAnomalyDetectionConfig returns default anomaly detection configuration
func DefaultAnomalyDetectionConfig() *AnomalyDetectionConfig {
	return &AnomalyDetectionConfig{
		MaxPricePerLiter:       2.5,  // USD
		MinPricePerLiter:       0.3,  // USD
		MaxLitersPerFill:       300,  // Liters
		MaxDistancePerDay:      1000, // Kilometers
		PriceVarianceThreshold: 0.3,  // 30% variance
	}
}

// AnomalyDetectionService handles fuel and expense anomaly detection
type AnomalyDetectionService struct {
	config *AnomalyDetectionConfig
	logger *zap.Logger
}

// NewAnomalyDetectionService creates a new anomaly detection service
func NewAnomalyDetectionService(config *AnomalyDetectionConfig, logger *zap.Logger) *AnomalyDetectionService {
	if config == nil {
		config = DefaultAnomalyDetectionConfig()
	}
	return &AnomalyDetectionService{
		config: config,
		logger: logger,
	}
}

// AnomalyResult represents the result of anomaly detection
type AnomalyResult struct {
	IsAnomaly      bool     `json:"is_anomaly"`
	AnomalyType    string   `json:"anomaly_type,omitempty"`
	AnomalyReason  string   `json:"anomaly_reason"`
	Confidence     float64  `json:"confidence"` // 0-1 scale
	SuggestedAction string  `json:"suggested_action,omitempty"`
}

// DetectFuelTransactionAnomalies detects anomalies in a fuel transaction
func (s *AnomalyDetectionService) DetectFuelTransactionAnomalies(tx *models.FuelTransaction, previousTx *models.FuelTransaction) *AnomalyResult {
	result := &AnomalyResult{
		IsAnomaly:  false,
		Confidence: 0.0,
	}

	// Check 1: Price per liter anomaly
	pricePerLiter := tx.Amount / tx.Liters
	if pricePerLiter > s.config.MaxPricePerLiter || pricePerLiter < s.config.MinPricePerLiter {
		result.IsAnomaly = true
		result.AnomalyType = "PRICE_ANOMALY"
		result.AnomalyReason = fmt.Sprintf("Price per liter ($%.2f) outside expected range ($%.2f-$%.2f)",
			pricePerLiter, s.config.MinPricePerLiter, s.config.MaxPricePerLiter)
		result.Confidence = 0.9
		result.SuggestedAction = "Review receipt and verify station pricing"
		return result
	}

	// Check 2: Excessive liters
	if tx.Liters > s.config.MaxLitersPerFill {
		result.IsAnomaly = true
		result.AnomalyType = "VOLUME_ANOMALY"
		result.AnomalyReason = fmt.Sprintf("Fuel volume (%.1fL) exceeds maximum expected (%.1fL)",
			tx.Liters, s.config.MaxLitersPerFill)
		result.Confidence = 0.85
		result.SuggestedAction = "Verify vehicle tank capacity and actual fill amount"
		return result
	}

	// Check 3: Distance-based anomaly (if previous transaction available)
	if previousTx != nil && previousTx.OdometerReading != nil && tx.OdometerReading != nil {
		distanceDriven := *tx.OdometerReading - *previousTx.OdometerReading
		if distanceDriven < 0 {
			result.IsAnomaly = true
			result.AnomalyType = "ODOMETER_ANOMALY"
			result.AnomalyReason = "Odometer reading decreased (possible tampering)"
			result.Confidence = 0.95
			result.SuggestedAction = "Investigate odometer tampering"
			return result
		}

		// Calculate fuel efficiency
		if distanceDriven > 0 {
			efficiency := distanceDriven / tx.Liters // km per liter
			if efficiency < 2.0 || efficiency > 20.0 {
				result.IsAnomaly = true
				result.AnomalyType = "EFFICIENCY_ANOMALY"
				result.AnomalyReason = fmt.Sprintf("Fuel efficiency (%.1f km/L) outside normal range", efficiency)
				result.Confidence = 0.75
				result.SuggestedAction = "Check for fuel theft or vehicle mechanical issues"
			}
		}
	}

	// Check 4: Price variance from location average (if historical data available)
	// This would require historical price data per location

	return result
}

// DetectExpenseAnomalies detects anomalies in expense reports
func (s *AnomalyDetectionService) DetectExpenseAnomalies(report *models.ExpenseReport, policy *models.ExpensePolicy) *AnomalyResult {
	result := &AnomalyResult{
		IsAnomaly:  false,
		Confidence: 0.0,
	}

	// Check 1: Exceeds policy limit
	if policy != nil && policy.LimitAmount != nil {
		if report.TotalAmount > *policy.LimitAmount {
			result.IsAnomaly = true
			result.AnomalyType = "POLICY_LIMIT_EXCEEDED"
			result.AnomalyReason = fmt.Sprintf("Total amount ($%.2f) exceeds policy limit ($%.2f)",
				report.TotalAmount, *policy.LimitAmount)
			result.Confidence = 1.0
			result.SuggestedAction = "Requires manager approval"
			return result
		}
	}

	// Check 2: Unusual amount patterns (would require historical data)
	// Check 3: Duplicate expense detection (would require comparison with other reports)

	return result
}

// GetAnomalyTypes returns all supported anomaly types
func (s *AnomalyDetectionService) GetAnomalyTypes() []string {
	return []string{
		"PRICE_ANOMALY",
		"VOLUME_ANOMALY",
		"ODOMETER_ANOMALY",
		"EFFICIENCY_ANOMALY",
		"POLICY_LIMIT_EXCEEDED",
		"DUPLICATE_TRANSACTION",
		"TIMING_ANOMALY",
	}
}

// UpdateConfig updates the anomaly detection configuration
func (s *AnomalyDetectionService) UpdateConfig(config *AnomalyDetectionConfig) {
	if config != nil {
		s.config = config
	}
}
