package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MultiShipmentService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewMultiShipmentService(repo repository.CustomerV2Repository, logger *zap.Logger) *MultiShipmentService {
	return &MultiShipmentService{
		repo:   repo,
		logger: logger,
	}
}

// SplitBooking IntoShipments handles splitting a massive cargo booking into multiple individual vehicle shipments
func (s *MultiShipmentService) SplitBookingIntoShipments(ctx context.Context, bookingID uuid.UUID, numVehicles int) ([]*models.Shipment, error) {
	s.logger.Info("Splitting booking into shipments", zap.String("booking_id", bookingID.String()), zap.Int("num_vehicles", numVehicles))

	booking, err := s.repo.GetBookingRequestByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	var shipments []*models.Shipment

	for i := 1; i <= numVehicles; i++ {
		// Calculate split cargo logic (e.g., total weight / numVehicles)
		// ...

		shipment := &models.Shipment{
			ID:          uuid.New(),
			CustomerID:  booking.CustomerID,
			BookingID:   &booking.ID,
			Status:      "BOOKED",
			Origin:      booking.PickupLocation,
			Destination: booking.DropLocation,
		}

		if err := s.repo.CreateShipment(ctx, shipment); err != nil {
			s.logger.Error("Failed creating child shipment", zap.Error(err))
			continue
		}

		shipments = append(shipments, shipment)

		// Link in the linkage table
		// link := &models.BookingShipment{ BookingID: booking.ID, ShipmentID: shipment.ID, SequenceNumber: i }
		// _ = s.repo.CreateBookingShipmentLink(ctx, link)
	}

	s.logger.Info("Created split shipments", zap.Int("count", len(shipments)))
	return shipments, nil
}
