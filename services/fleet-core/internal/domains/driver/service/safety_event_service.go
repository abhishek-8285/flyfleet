package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"go.uber.org/zap"
)

type SafetyRepo interface {
	CreateSafetyEvent(ctx context.Context, event *models.SafetyEvent) error
}

type SafetyEventService struct {
	repo   SafetyRepo
	logger *zap.Logger
}

func NewSafetyEventService(repo SafetyRepo, logger *zap.Logger) *SafetyEventService {
	return &SafetyEventService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SafetyEventService) RecordSafetyEvent(ctx context.Context, event *models.SafetyEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return s.repo.CreateSafetyEvent(ctx, event)
}

func (s *SafetyEventService) DetectFatigue(ctx context.Context, driverID string, continuousDrivingHours float64) error {
	if continuousDrivingHours > 5.0 {
		event := &models.SafetyEvent{
			DriverID:  driverID,
			EventType: "FATIGUE",
			Severity:  "HIGH",
			Timestamp: time.Now(),
		}
		return s.RecordSafetyEvent(ctx, event)
	}
	return nil
}
