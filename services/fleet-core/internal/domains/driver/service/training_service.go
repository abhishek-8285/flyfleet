package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"go.uber.org/zap"
)

type TrainingRepo interface {
	CreateModule(ctx context.Context, module *models.TrainingModule) error
	CreateRecord(ctx context.Context, record *models.TrainingRecord) error
	GetRecordsByDriverID(ctx context.Context, driverID string) ([]*models.TrainingRecord, error)
}

type TrainingService struct {
	repo   TrainingRepo
	logger *zap.Logger
}

func NewTrainingService(repo TrainingRepo, logger *zap.Logger) *TrainingService {
	return &TrainingService{
		repo:   repo,
		logger: logger,
	}
}

func (s *TrainingService) AssignTraining(ctx context.Context, driverID string, moduleID string) error {
	record := &models.TrainingRecord{
		DriverID: driverID,
		ModuleID: moduleID,
	}
	return s.repo.CreateRecord(ctx, record)
}

func (s *TrainingService) CompleteTraining(ctx context.Context, driverID string, moduleID string, certificateURL string) error {
	// Simple implementation for now: create a completed record
	// In a real system, we would update an existing pending record
	now := time.Now()
	record := &models.TrainingRecord{
		DriverID:       driverID,
		ModuleID:       moduleID,
		CompletedAt:    &now,
		CertificateURL: certificateURL,
	}
	return s.repo.CreateRecord(ctx, record)
}

func (s *TrainingService) GetTrainingHistory(ctx context.Context, driverID string) ([]*models.TrainingRecord, error) {
	return s.repo.GetRecordsByDriverID(ctx, driverID)
}
