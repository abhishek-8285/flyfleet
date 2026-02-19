package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
)

// OfferTrip creates a trip offer for a driver
func (s *TripService) OfferTrip(ctx context.Context, tripID, driverID string) (*models.TripOffer, error) {
	// Check if trip exists and is assignable
	trip, err := s.Get(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	if trip.Status != "planned" && trip.Status != "offered" {
		return nil, fmt.Errorf("trip cannot be offered in current status: %s", trip.Status)
	}

	expiresAt := time.Now().Add(15 * time.Minute) // 15 min expiry
	
	offer := &models.TripOffer{
		TripID:    tripID,
		DriverID:  driverID,
		Status:    "OFFERED",
		ExpiresAt: &expiresAt,
	}

	if err := s.repo.CreateOffer(ctx, offer); err != nil {
		s.logger.Error("failed to create trip offer", 
			zap.String("trip_id", tripID), 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to create trip offer: %w", err)
	}
	
	// Update trip status to reflect pending offer
	// s.repo.Update(ctx, tripID, map[string]interface{}{"status": "offered"})

	s.publishEvent("trip.offered", offer)
	return offer, nil
}

// RespondToOffer handles driver acceptance or rejection of a trip
func (s *TripService) RespondToOffer(ctx context.Context, offerID, driverID, response, reason string) error {
	offer, err := s.repo.GetOffer(ctx, offerID)
	if err != nil {
		return err
	}

	if offer.DriverID != driverID {
		return fmt.Errorf("unauthorized to respond to this offer")
	}

	if offer.Status != "OFFERED" {
		return fmt.Errorf("offer is no longer valid (status: %s)", offer.Status)
	}

	if offer.ExpiresAt != nil && offer.ExpiresAt.Before(time.Now()) {
		s.repo.RespondToOffer(ctx, offerID, "EXPIRED", "Auto-expired")
		return fmt.Errorf("offer has expired")
	}

	if response == "ACCEPTED" {
		// Update offer status
		if err := s.repo.RespondToOffer(ctx, offerID, "ACCEPTED", ""); err != nil {
			return err
		}
		
		// Assign trip to driver
		updates := map[string]interface{}{
			"driver_id": driverID,
			"status":    "assigned", // or planned
		}
		if err := s.repo.Update(ctx, offer.TripID, updates); err != nil {
			// Rollback offer status if possible, or log error
			s.logger.Error("failed to assign trip after acceptance", zap.Error(err))
			return fmt.Errorf("failed to assign trip: %w", err)
		}
		
		s.publishEvent("trip.accepted", map[string]string{"trip_id": offer.TripID, "driver_id": driverID})
		
	} else if response == "REJECTED" {
		if err := s.repo.RespondToOffer(ctx, offerID, "REJECTED", reason); err != nil {
			return err
		}
		s.publishEvent("trip.rejected", map[string]string{"trip_id": offer.TripID, "driver_id": driverID, "reason": reason})
	} else {
		return fmt.Errorf("invalid response: %s", response)
	}

	return nil
}

// GetDriverOffers retrieves active offers for a driver
func (s *TripService) GetDriverOffers(ctx context.Context, driverID string) ([]*models.TripOffer, error) {
	return s.repo.GetOffersForDriver(ctx, driverID)
}
