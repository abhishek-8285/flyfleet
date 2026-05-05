package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver"
	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BehaviorScoreService handles driver behavior score management
type BehaviorScoreService struct {
	repo   Repository
	logger *zap.Logger
}

// NewBehaviorScoreService creates a new behavior score service
func NewBehaviorScoreService(repo Repository, logger *zap.Logger) *BehaviorScoreService {
	return &BehaviorScoreService{
		repo:   repo,
		logger: logger,
	}
}

// UpdateBehaviorScore updates driver behavior score with validation and audit logging
func (s *BehaviorScoreService) UpdateBehaviorScore(ctx context.Context, driverID, tenantID, userID uuid.UUID, userEmail string, scoreChange int, reason, eventType string) (*models.Driver, error) {
	// Validate score change range
	if scoreChange < -20 || scoreChange > 20 {
		return nil, ErrInvalidScoreChange
	}

	// Get current driver
	driver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}
	if driver == nil {
		return nil, ErrDriverNotFound
	}

	if driver.TenantID != tenantID.String() {
		return nil, fmt.Errorf("driver does not belong to organization")
	}

	oldScore := driver.BehaviorScore
	newScore := oldScore + scoreChange

	// Ensure score stays within bounds
	if newScore < 0 {
		newScore = 0
		scoreChange = -oldScore // Adjust actual change
	} else if newScore > 100 {
		newScore = 100
		scoreChange = 100 - oldScore // Adjust actual change
	}

	// Update score in repository
	if err := s.repo.UpdateBehaviorScore(ctx, driverID.String(), scoreChange); err != nil {
		return nil, fmt.Errorf("failed to update behavior score: %w", err)
	}

	// Create audit log
	if err := s.repo.LogBehaviorScoreChange(ctx, driverID, tenantID, oldScore, newScore, reason, eventType, userID, userEmail); err != nil {
		s.logger.Warn("Failed to create audit log for behavior score change", zap.Error(err))
	}

	// Create behavior score update record
	scoreUpdate := &models.BehaviorScoreUpdate{
		ID:          uuid.New().String(),
		DriverID:    driverID.String(),
		ScoreChange: scoreChange,
		Reason:      reason,
		EventType:   eventType,
		CreatedBy:   userID.String(),
		CreatedAt:   time.Now(),
	}
	if err := s.repo.CreateBehaviorScoreUpdate(ctx, scoreUpdate); err != nil {
		s.logger.Warn("Failed to create behavior score update record", zap.Error(err))
	}

	// Get updated driver
	updatedDriver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get updated driver: %w", err)
	}

	s.logger.Info("Driver behavior score updated",
		zap.String("driver_id", driverID.String()),
		zap.Int("old_score", oldScore),
		zap.Int("new_score", newScore),
		zap.Int("score_change", scoreChange),
		zap.String("reason", reason),
		zap.String("event_type", eventType),
	)

	return updatedDriver, nil
}

// GetBehaviorScoreHistory retrieves behavior score change history for a driver
func (s *BehaviorScoreService) GetBehaviorScoreHistory(ctx context.Context, driverID string, limit, offset int) ([]*models.BehaviorScoreUpdate, error) {
	// This would need a new repository method to fetch score updates
	// For now, return empty slice
	return []*models.BehaviorScoreUpdate{}, nil
}

func (s *BehaviorScoreService) CalculateBehaviorScore(ctx context.Context, driverID string, periodStart, periodEnd time.Time) (int, error) {
	driver, err := s.repo.GetByID(ctx, driverID)
	if err != nil {
		return 0, fmt.Errorf("failed to get driver: %w", err)
	}
	if driver == nil {
		return 0, ErrDriverNotFound
	}

	// Get behavior records for the period
	behaviorRecords, err := s.repo.GetBehaviorRecords(ctx, driverID, periodStart, periodEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to get behavior records: %w", err)
	}

	if len(behaviorRecords) == 0 {
		return driver.BehaviorScore, nil
	}

	// Calculate average score from records
	totalScore := 0.0
	for _, record := range behaviorRecords {
		totalScore += record.Score
	}
	averageScore := int(totalScore / float64(len(behaviorRecords)))

	// Ensure score is within bounds
	if averageScore < 0 {
		averageScore = 0
	} else if averageScore > 100 {
		averageScore = 100
	}

	return averageScore, nil
}

// GetBehaviorReport generates a comprehensive behavior report for a driver
func (s *BehaviorScoreService) GetBehaviorReport(ctx context.Context, driverID string, periodType string, startDate, endDate *time.Time) (*models.DriverBehavior, error) {
	// Set default date range if not provided
	if startDate == nil {
		now := time.Now()
		switch periodType {
		case "daily":
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			startDate = &start
		case "weekly":
			start := now.AddDate(0, 0, -7)
			startDate = &start
		case "monthly":
			start := now.AddDate(0, -1, 0)
			startDate = &start
		default:
			start := now.AddDate(0, -1, 0)
			startDate = &start
		}
	}

	if endDate == nil {
		now := time.Now()
		endDate = &now
	}

	// Get behavior summary
	summary, err := s.repo.GetBehaviorSummary(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get behavior summary: %w", err)
	}

	if summary == nil {
		return nil, fmt.Errorf("no behavior data found for driver")
	}

	// Get incidents
	incidents, err := s.repo.GetIncidents(ctx, driverID, models.BehaviorReportRequest{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}

	// Build comprehensive report
	report := &models.DriverBehavior{
		ID:                 uuid.New().String(),
		DriverID:           driverID,
		PeriodStart:        *startDate,
		PeriodEnd:          *endDate,
		Score:              summary.Score,
		SafetyScore:        summary.SafetyScore,
		EfficiencyScore:    summary.EfficiencyScore,
		ComplianceScore:    summary.ComplianceScore,
		HarshAccelerations: summary.HarshAccelerations,
		HarshBrakings:      summary.HarshBrakings,
		SpeedingIncidents:  summary.SpeedingIncidents,
		RiskLevel:          summary.RiskLevel,
	}

	// Calculate risk factors
	riskFactors := []string{}
	if summary.SpeedingIncidents > 5 {
		riskFactors = append(riskFactors, "frequent_speeding")
	}
	if summary.HarshBrakings > 10 {
		riskFactors = append(riskFactors, "harsh_braking")
	}
	if summary.HarshAccelerations > 10 {
		riskFactors = append(riskFactors, "harsh_acceleration")
	}
	report.RiskFactors = riskFactors

	// Calculate improvement areas
	improvementAreas := []string{}
	if summary.SafetyScore < 70 {
		improvementAreas = append(improvementAreas, "safety practices")
	}
	if summary.EfficiencyScore < 70 {
		improvementAreas = append(improvementAreas, "fuel efficiency")
	}
	if summary.ComplianceScore < 70 {
		improvementAreas = append(improvementAreas, "regulatory compliance")
	}
	report.ImprovementAreas = improvementAreas

	_ = incidents // Use incidents data for detailed report

	return report, nil
}

// AssessRiskLevel assesses driver risk level based on behavior metrics
func (s *BehaviorScoreService) AssessRiskLevel(score float64, speedingIncidents, harshBrakings, harshAccelerations int) string {
	// Calculate risk score
	riskScore := 0.0

	// Factor 1: Overall behavior score (weight: 40%)
	riskScore += (100 - score) * 0.4

	// Factor 2: Speeding incidents (weight: 30%)
	if speedingIncidents > 10 {
		riskScore += 30
	} else if speedingIncidents > 5 {
		riskScore += 20
	} else if speedingIncidents > 2 {
		riskScore += 10
	}

	// Factor 3: Harsh braking (weight: 15%)
	if harshBrakings > 20 {
		riskScore += 15
	} else if harshBrakings > 10 {
		riskScore += 10
	} else if harshBrakings > 5 {
		riskScore += 5
	}

	// Factor 4: Harsh acceleration (weight: 15%)
	if harshAccelerations > 20 {
		riskScore += 15
	} else if harshAccelerations > 10 {
		riskScore += 10
	} else if harshAccelerations > 5 {
		riskScore += 5
	}

	// Determine risk level
	if riskScore >= 75 {
		return "critical"
	} else if riskScore >= 50 {
		return "high"
	} else if riskScore >= 25 {
		return "medium"
	}
	return "low"
}

// ProcessBehaviorEvent handles raw telemetry events and updates driver scores
func (s *BehaviorScoreService) ProcessBehaviorEvent(ctx context.Context, event *models.BehaviorEvent) error {
	// 1. Calculate penalty
	penalty := 0
	switch event.EventType {
	case "speeding":
		penalty = models.PenaltySpeeding
	case "harsh_braking":
		penalty = models.PenaltyHarshBrake
	case "harsh_acceleration":
		penalty = models.PenaltyHarshAcceleration
	case "phone_usage":
		penalty = models.PenaltyPhoneUsage
	case "seatbelt_violation":
		penalty = models.PenaltySeatbelt
	}

	// Increment Prometheus metrics
	driver.GlobalMetrics.BehaviorEventsTotal.WithLabelValues(event.EventType).Inc()

	if penalty == 0 {
		return nil
	}

	// 2. Update score
	driver, err := s.repo.GetByID(ctx, event.DriverID)
	if err != nil {
		return err
	}
	if driver == nil {
		return ErrDriverNotFound
	}

	tenantID, _ := uuid.Parse(driver.TenantID)
	// Using ActionDriverBehaviorScoreUpdated as a placeholder for event processing
	_, err = s.UpdateBehaviorScore(ctx, uuid.MustParse(event.DriverID), tenantID, uuid.Nil, "system", -penalty, fmt.Sprintf("Automated penalty for %s", event.EventType), event.EventType)
	return err
}
