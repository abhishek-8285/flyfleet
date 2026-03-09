package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer_v2/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ETAPredictionService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewETAPredictionService(repo repository.CustomerV2Repository, logger *zap.Logger) *ETAPredictionService {
	return &ETAPredictionService{
		repo:   repo,
		logger: logger,
	}
}

// PredictETA uses vehicle speed, distance remaining, traffic and historical data to predict ETA
func (s *ETAPredictionService) PredictETA(ctx context.Context, shipmentID uuid.UUID) (*time.Time, error) {
	s.logger.Info("Predicting ETA for shipment", zap.String("shipment_id", shipmentID.String()))

	// Mock prediction logic:
	// distance = 500km
	// speed = average 50 km/h
	// time = 10 hours from now
	
	eta := time.Now().Add(10 * time.Hour)

	return &eta, nil
}
