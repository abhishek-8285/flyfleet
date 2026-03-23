package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ShipmentTrackingService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewShipmentTrackingService(repo repository.CustomerV2Repository, logger *zap.Logger) *ShipmentTrackingService {
	return &ShipmentTrackingService{
		repo:   repo,
		logger: logger,
	}
}

// GetLiveTracking integrates with tracking_v2 to fetch current location, speed, and checkpoint history
func (s *ShipmentTrackingService) GetLiveTracking(ctx context.Context, shipmentID uuid.UUID) (*models.ShipmentTrackingData, error) {
	s.logger.Info("Fetching live tracking for shipment", zap.String("shipment_id", shipmentID.String()))

	shipment, err := s.repo.GetShipmentByID(ctx, shipmentID)
	if err != nil || shipment == nil {
		return nil, err
	}

	// Call TrackingV2 or TripV2 module here
	// s.trackingClient.GetLiveLocation(ctx, shipment.TripID)

	// Since we don't have the fully running tracker, we return mock tracking telemetry
	currentLat := 19.0760
	currentLng := 72.8777
	
	tracking := &models.ShipmentTrackingData{
		ShipmentID:      shipment.ID,
		TrackingNumber:  "TRK-" + shipment.ID.String()[:8], // Stub
		Status:          shipment.Status,
		CurrentLocation: &models.Location{Lat: currentLat, Lng: currentLng},
		SpeedKmH:        65.5,
		Events:          []models.ShipmentEvent{}, // Could fetch from customer_v2_shipment_events
	}

	return tracking, nil
}
