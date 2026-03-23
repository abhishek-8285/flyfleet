package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ShipmentAssignmentService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewShipmentAssignmentService(repo repository.CustomerV2Repository, logger *zap.Logger) *ShipmentAssignmentService {
	return &ShipmentAssignmentService{
		repo:   repo,
		logger: logger,
	}
}

// AssignShipment handles the Booking -> Shipment conversion, vehicle/driver assignment and event publishing
func (s *ShipmentAssignmentService) AssignShipment(ctx context.Context, bookingID uuid.UUID) (*models.Shipment, error) {
	s.logger.Info("Assigning resources to booking", zap.String("booking_id", bookingID.String()))

	booking, err := s.repo.GetBookingRequestByID(ctx, bookingID)
	if err != nil {
		s.logger.Error("Failed to fetch booking for assignment", zap.Error(err))
		return nil, err
	}

	// Stub: In reality, we'd interact with trip_v2 to assign a vehicle, driver and create a trip
	// trip, err := s.tripClient.CreateTrip(ctx, &tripProto.CreateTripReq{...})
	mockTripID := uuid.New()

	// Create Shipment
	shipment := &models.Shipment{
		ID:          uuid.New(),
		CustomerID:  booking.CustomerID,
		BookingID:   &booking.ID,
		TripID:      &mockTripID,
		Status:      "ASSIGNED",
		Origin:      booking.PickupLocation,
		Destination: booking.DropLocation,
	}

	if err := s.repo.CreateShipment(ctx, shipment); err != nil {
		s.logger.Error("Failed to create shipment from booking", zap.Error(err))
		return nil, err
	}

	// Update booking status
	booking.Status = "CONFIRMED"
	// _ = s.repo.UpdateBooking(ctx, booking) 

	// Publish event logic (e.g., NATS/Kafka)
	// s.eventBus.Publish("shipment.assigned", shipment)

	s.logger.Info("Shipment assigned successfully", zap.String("shipment_id", shipment.ID.String()))
	return shipment, nil
}
