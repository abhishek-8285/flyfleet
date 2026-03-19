package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/google/uuid"
)

type BookingService struct {
	repo repository.CustomerV2Repository
}

func NewBookingService(repo repository.CustomerV2Repository) *BookingService {
	return &BookingService{repo: repo}
}

func (s *BookingService) CreateBooking(ctx context.Context, b *models.BookingRequest) error {
	b.Status = "REQUESTED"
	return s.repo.CreateBookingRequest(ctx, b)
}

func (s *BookingService) GetBookingByID(ctx context.Context, id uuid.UUID) (*models.BookingRequest, error) {
	return s.repo.GetBookingRequestByID(ctx, id)
}
