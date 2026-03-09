package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FraudSignal is a detection signal that raised the risk score.
type FraudSignal string

const (
	SignalMultiVehicleCard     FraudSignal = "MULTI_VEHICLE_CARD"
	SignalRepeatedStation      FraudSignal = "REPEATED_STATION"
	SignalNighttimeFueling     FraudSignal = "NIGHTTIME_FUELING"
	SignalSuddenSpike          FraudSignal = "SUDDEN_FUEL_SPIKE"
	SignalExceedsTankCapacity  FraudSignal = "EXCEEDS_TANK_CAPACITY"
	SignalHighFrequency        FraudSignal = "HIGH_FREQUENCY_FILLS"
)

// FraudDetectionService scores fuel log events for fraud risk.
type FraudDetectionService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewFraudDetectionService(repo *repository.FinanceRepository, logger *zap.Logger) *FraudDetectionService {
	return &FraudDetectionService{repo: repo, logger: logger}
}

// ScoreFuelLog computes a fraud score [0–100] for a fuel log with detected signals.
func (s *FraudDetectionService) ScoreFuelLog(ctx context.Context, logID uuid.UUID, orgID uuid.UUID) (float64, models.FraudRiskLevel, []FraudSignal, error) {
	score := 0.0
	var signals []FraudSignal

	// Signal 1: Nighttime fueling (22:00 – 05:00)
	hour := time.Now().Hour()
	if hour >= 22 || hour < 5 {
		score += 20
		signals = append(signals, SignalNighttimeFueling)
	}

	// For full implementation, these would query the DB:
	// Signal 2: High frequency fills (> 2 within 24h for same vehicle)
	// Signal 3: Sudden fuel spike (> 20% more than tank capacity)
	// Signal 4: Same card → multiple vehicles
	// Signal 5: Same station, consecutive days

	// Stub signals for demonstration
	_ = logID
	_ = orgID

	level := riskLevel(score)
	return score, level, signals, nil
}

// CreateFraudIncident persists a detected fraud incident.
func (s *FraudDetectionService) CreateFraudIncident(ctx context.Context, entityType string, entityID, orgID uuid.UUID, score float64, level models.FraudRiskLevel, signals []FraudSignal) (*models.FraudIncident, error) {
	sigStrs := make([]string, len(signals))
	for i, sig := range signals {
		sigStrs[i] = string(sig)
	}

	incident := &models.FraudIncident{
		ID:         uuid.New(),
		EntityType: entityType,
		EntityID:   entityID,
		OrgID:      orgID,
		FraudScore: score,
		RiskLevel:  level,
		Signals:    sigStrs,
		Status:     "OPEN",
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateFraudIncident(ctx, incident); err != nil {
		return nil, err
	}

	return incident, nil
}

func riskLevel(score float64) models.FraudRiskLevel {
	switch {
	case score >= 75:
		return models.FraudRiskCritical
	case score >= 50:
		return models.FraudRiskHigh
	case score >= 25:
		return models.FraudRiskMedium
	default:
		return models.FraudRiskLow
	}
}
