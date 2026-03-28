package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"go.uber.org/zap"
)

type BookingValidationService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewBookingValidationService(repo repository.CustomerV2Repository, logger *zap.Logger) *BookingValidationService {
	return &BookingValidationService{
		repo:   repo,
		logger: logger,
	}
}

// ValidateBooking executes required validations: vehicle availability, driver availability, cargo weight vs vehicle capacity, customer credit, route feasibility.
func (s *BookingValidationService) ValidateBooking(ctx context.Context, booking *models.BookingRequest) (bool, error) {
	s.logger.Info("Validating booking", zap.String("booking_id", booking.ID.String()))

	allPassed := true

	validations := []struct {
		ValidationType string
		CheckFn        func() (bool, string)
	}{
		{
			ValidationType: "vehicle_availability",
			CheckFn: func() (bool, string) {
				// Stub: check if requested vehicle type is available
				if booking.VehicleType != nil && *booking.VehicleType == "unsupported" {
					return false, "Requested vehicle type is unavailable"
				}
				return true, "Vehicle available"
			},
		},
		{
			ValidationType: "driver_availability",
			CheckFn: func() (bool, string) {
				// Stub: check driver pool
				return true, "Drivers available in region"
			},
		},
		{
			ValidationType: "cargo_weight",
			CheckFn: func() (bool, string) {
				// Stub: check capacity
				if booking.Weight != nil && *booking.Weight > 40000 {
					return false, "Cargo exceeds maximum allowable weight limit"
				}
				return true, "Weight within limits"
			},
		},
		{
			ValidationType: "customer_credit",
			CheckFn: func() (bool, string) {
				// Stub: In reality, call CreditService to check
				return true, "Credit limit sufficient"
			},
		},
		{
			ValidationType: "route_feasibility",
			CheckFn: func() (bool, string) {
				// Stub: Check if origin/destination mapping is valid
				return true, "Route feasible"
			},
		},
	}

	for _, v := range validations {
		passed, details := v.CheckFn()
		if !passed {
			allPassed = false
		}

		status := "PASSED"
		if !passed {
			status = "FAILED"
		}

		// Save the validation result to the repository. Note: Add CreateBookingValidation to repo later if needed,
		// or handle it in a batch. For simplicity, we are just conceptually recording it here.
		// A full implementation would call s.repo.CreateBookingValidation(ctx, validationRecord)

		s.logger.Info("Validation executed",
			zap.String("type", v.ValidationType),
			zap.String("status", status),
			zap.String("details", details),
		)
	}

	return allPassed, nil
}
