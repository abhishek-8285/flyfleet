package service

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type CustomerService struct {
	repo       *repository.CustomerRepository
	logger     *zap.Logger
	jwtSecret  []byte // In a real app, inject this from config
}

func NewCustomerService(repo *repository.CustomerRepository, logger *zap.Logger, jwtSecret string) *CustomerService {
	return &CustomerService{
		repo:      repo,
		logger:    logger,
		jwtSecret: []byte(jwtSecret),
	}
}

// Authenticate validates API Key and returns a JWT
func (s *CustomerService) Authenticate(ctx context.Context, apiKey string) (string, *models.Customer, error) {
	customer, err := s.repo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		s.logger.Warn("authentication failed: invalid api key", zap.Error(err))
		return "", nil, errors.New("invalid api key")
	}

	if customer.Status != "active" {
		return "", nil, errors.New("customer account is not active")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": customer.ID,
		"org": customer.OrgID,
		"role": "customer",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(), // 1 hour expiry
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, customer, nil
}

// TrackShipment retrieves shipment details details
func (s *CustomerService) TrackShipment(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	return s.repo.GetShipmentByTrackingNumber(ctx, trackingNumber)
}

// CreateBooking checks logic and submits a booking
func (s *CustomerService) CreateBooking(ctx context.Context, customerID string, req *models.BookingRequestCreate) (*models.BookingRequest, error) {
	booking := &models.BookingRequest{
		CustomerID:          customerID,
		RequestType:         req.RequestType,
		Origin:              req.Origin,
		Destination:         req.Destination,
		RequestedPickupTime: req.RequestedPickupTime,
		CargoDetails:        req.CargoDetails,
		Status:              "pending",
		Notes:               req.Notes,
	}

	if err := s.repo.CreateBookingRequest(ctx, booking); err != nil {
		return nil, err
	}
	
	s.logger.Info("booking request created", 
		zap.String("customer_id", customerID),
		zap.String("booking_id", booking.ID))

	return booking, nil
}

// GetCustomerInvoices retrieves invoices
func (s *CustomerService) GetCustomerInvoices(ctx context.Context, customerID string) ([]*models.Invoice, error) {
	return s.repo.GetInvoices(ctx, customerID)
}

// Internal: CreateCustomer (for admin use)
func (s *CustomerService) CreateCustomer(ctx context.Context, c *models.Customer) error {
	// In reality, generate API Key here
	return s.repo.Create(ctx, c)
}
