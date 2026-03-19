package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type QuoteService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewQuoteService(repo repository.CustomerV2Repository, logger *zap.Logger) *QuoteService {
	return &QuoteService{
		repo:   repo,
		logger: logger,
	}
}

// GenerateQuote computes a pricing quote for a given booking request based on distance, weight, type, contract rate, fuel surcharge
func (s *QuoteService) GenerateQuote(ctx context.Context, booking *models.BookingRequest) (*models.BookingQuote, error) {
	s.logger.Info("Generating quote for booking", zap.String("booking_id", booking.ID.String()))

	// Mock pricing logic
	var baseRate float64 = 5000.0 // Fixed base rate
	var perKmRate float64 = 20.0  // Rate per km
	var mockDistanceKm float64 = 150.0 // Mock distance for quote calculation

	// Calculate cost
	distanceCost := mockDistanceKm * perKmRate
	
	weightCost := 0.0
	if booking.Weight != nil {
		weightCost = *booking.Weight * 2.5 // $2.5 per kg
	}

	fuelSurcharge := (baseRate + distanceCost + weightCost) * 0.15 // 15% fuel surcharge
	
	totalAmount := baseRate + distanceCost + weightCost + fuelSurcharge

	// Create quote record
	validUntil := time.Now().Add(48 * time.Hour)
	quote := &models.BookingQuote{
		ID:          uuid.New(),
		BookingID:   booking.ID,
		QuoteAmount: totalAmount,
		Currency:    "INR",
		ValidUntil:  &validUntil,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}

	// Persist quote
	// err := s.repo.CreateBookingQuote(ctx, quote) // Assume repo method exists or will exist

	s.logger.Info("Quote generated successfully", zap.Float64("amount", quote.QuoteAmount))

	return quote, nil
}
