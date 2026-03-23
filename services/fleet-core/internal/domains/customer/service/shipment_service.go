package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/google/uuid"
)

// TrackingServiceInterface represents the dependency on the tracking_v2 module
type TrackingServiceInterface interface {
	GetTelemetryForShipment(ctx context.Context, shipmentID uuid.UUID) (map[string]interface{}, error)
}

type ShipmentService struct {
	repo            repository.CustomerV2Repository
	trackingService TrackingServiceInterface
}

func NewShipmentService(repo repository.CustomerV2Repository, tracking TrackingServiceInterface) *ShipmentService {
	return &ShipmentService{
		repo:            repo,
		trackingService: tracking,
	}
}

func (s *ShipmentService) CreateShipment(ctx context.Context, sh *models.Shipment) error {
	sh.Status = "BOOKED"
	return s.repo.CreateShipment(ctx, sh)
}

func (s *ShipmentService) GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error) {
	return s.repo.GetShipmentByID(ctx, id)
}

func (s *ShipmentService) GetShipmentTracking(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	trackingData, err := s.repo.GetShipmentTracking(ctx, id)
	if err != nil {
		return nil, err
	}
	if trackingData == nil {
		return nil, nil
	}

	// Fetch telemetry from tracking_v2
	if s.trackingService != nil {
		telemetry, err := s.trackingService.GetTelemetryForShipment(ctx, id)
		if err == nil && telemetry != nil {
			trackingData["telemetry"] = telemetry
		}
	}

	return trackingData, nil
}
