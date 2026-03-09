package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/google/uuid"
)

type EmergencyAnalyticsService struct {
	repo repository.EmergencyRepository
}

func NewEmergencyAnalyticsService(repo repository.EmergencyRepository) *EmergencyAnalyticsService {
	return &EmergencyAnalyticsService{repo: repo}
}

func (s *EmergencyAnalyticsService) GetIncidentAnalytics(ctx context.Context, orgID uuid.UUID) (*models.IncidentAnalyticsResponse, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)

	todayMetrics, err := s.repo.GetIncidentMetrics(ctx, orgID, todayStart, now)
	if err != nil {
		return nil, err
	}

	weekMetrics, err := s.repo.GetIncidentMetrics(ctx, orgID, weekStart, now)
	if err != nil {
		return nil, err
	}

	typeDist, err := s.repo.GetIncidentTypeDistribution(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &models.IncidentAnalyticsResponse{
		IncidentsToday:        todayMetrics.TotalIncidents,
		IncidentsThisWeek:     weekMetrics.TotalIncidents,
		IncidentsByType:       typeDist,
		AverageResponseTime:   weekMetrics.AvgResponseTimeMin,
		AverageResolutionTime: weekMetrics.AvgResolutionTimeMin,
	}, nil
}

func (s *EmergencyAnalyticsService) GetSafetyDashboard(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	daily, err := s.repo.GetIncidentsByDay(ctx, orgID, 30)
	if err != nil {
		return nil, err
	}

	dist, err := s.repo.GetIncidentTypeDistribution(ctx, orgID)
	if err != nil {
		return nil, err
	}

	scores, err := s.repo.GetVehicleSafetyScores(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Also get general metrics from existing repo method
	// Using a default range for "today" or "recent"
	// but for dashboard we'll stick to the specific analytics for now.

	return map[string]interface{}{
		"daily_trends":      daily,
		"type_distribution": dist,
		"vehicle_scores":    scores,
	}, nil
}
