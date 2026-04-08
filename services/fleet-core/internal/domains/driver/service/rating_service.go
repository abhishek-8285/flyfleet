package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"go.uber.org/zap"
)

type RatingRepo interface {
	CreateRating(ctx context.Context, rating *models.Rating) error
	GetAverageRating(ctx context.Context, driverID string) (float64, error)
}

type RatingService struct {
	repo   RatingRepo
	logger *zap.Logger
}

func NewRatingService(repo RatingRepo, logger *zap.Logger) *RatingService {
	return &RatingService{
		repo:   repo,
		logger: logger,
	}
}

func (s *RatingService) SubmitRating(ctx context.Context, rating *models.Rating) error {
	if rating.OverallRating == 0 {
		rating.OverallRating = float64(rating.PunctualityScore+rating.BehaviorScore+rating.DrivingScore) / 3.0
	}
	return s.repo.CreateRating(ctx, rating)
}

func (s *RatingService) GetDriverRating(ctx context.Context, driverID string) (float64, error) {
	return s.repo.GetAverageRating(ctx, driverID)
}
