package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OnboardingRepo interface {
	CreateStep(ctx context.Context, step *models.OnboardingStep) error
	UpdateStep(ctx context.Context, step *models.OnboardingStep) error
	GetStepsByDriverID(ctx context.Context, driverID uuid.UUID) ([]*models.OnboardingStep, error)
}

type OnboardingService struct {
	repo   OnboardingRepo
	logger *zap.Logger
}

func NewOnboardingService(repo OnboardingRepo, logger *zap.Logger) *OnboardingService {
	return &OnboardingService{
		repo:   repo,
		logger: logger,
	}
}

func (s *OnboardingService) StartOnboarding(ctx context.Context, driverID uuid.UUID) error {
	initialStep := &models.OnboardingStep{
		DriverID:    driverID,
		Step:        "APPLICATION",
		Status:      "COMPLETED",
		Remarks:     "Application submitted",
		CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
	}

	if err := s.repo.CreateStep(ctx, initialStep); err != nil {
		return fmt.Errorf("failed to start onboarding: %w", err)
	}

	// Create next step: DOCUMENT_VERIFICATION
	nextStep := &models.OnboardingStep{
		DriverID: driverID,
		Step:     "DOCUMENT_VERIFICATION",
		Status:   "PENDING",
	}

	return s.repo.CreateStep(ctx, nextStep)
}

func (s *OnboardingService) AdvanceOnboarding(ctx context.Context, driverID uuid.UUID, stepName string, status string, remarks string) error {
	steps, err := s.repo.GetStepsByDriverID(ctx, driverID)
	if err != nil {
		return err
	}

	var currentStep *models.OnboardingStep
	for _, step := range steps {
		if step.Step == stepName && step.Status == "PENDING" {
			currentStep = step
			break
		}
	}

	if currentStep == nil {
		return fmt.Errorf("step %s not found or already completed for driver %s", stepName, driverID)
	}

	currentStep.Status = status
	currentStep.Remarks = remarks
	if status == "COMPLETED" {
		t := time.Now()
		currentStep.CompletedAt = &t
	}

	if err := s.repo.UpdateStep(ctx, currentStep); err != nil {
		return err
	}

	if status == "COMPLETED" {
		return s.createNextStep(ctx, driverID, stepName)
	}

	return nil
}

func (s *OnboardingService) createNextStep(ctx context.Context, driverID uuid.UUID, currentStepName string) error {
	var nextStepName string
	switch currentStepName {
	case "DOCUMENT_VERIFICATION":
		nextStepName = "POLICE_VERIFICATION"
	case "POLICE_VERIFICATION":
		nextStepName = "MEDICAL_CHECK"
	case "MEDICAL_CHECK":
		nextStepName = "TRAINING"
	case "TRAINING":
		nextStepName = "FINAL_APPROVAL"
	default:
		return nil // End of workflow
	}

	nextStep := &models.OnboardingStep{
		DriverID: driverID,
		Step:     nextStepName,
		Status:   "PENDING",
	}

	return s.repo.CreateStep(ctx, nextStep)
}

func (s *OnboardingService) GetOnboardingStatus(ctx context.Context, driverID uuid.UUID) ([]*models.OnboardingStep, error) {
	return s.repo.GetStepsByDriverID(ctx, driverID)
}
