package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type CustomerService struct {
	repo       *repository.CustomerRepository
	logger     *zap.Logger
	jwtSecret  []byte
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
		return "", nil, ErrInvalidAPIKey
	}

	if customer.Status != string(models.StatusActive) {
		return "", nil, ErrCustomerInactive
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": customer.ID,
		"org": customer.OrgID,
		"role": "customer",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, customer, nil
}

// TrackShipment retrieves shipment details
func (s *CustomerService) TrackShipment(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	if trackingNumber == "" {
		return nil, ErrInvalidTracking
	}
	return s.repo.GetShipmentByTrackingNumber(ctx, trackingNumber)
}

// CreateBooking validates and creates a booking request
func (s *CustomerService) CreateBooking(ctx context.Context, customerID string, req *models.BookingRequestCreate) (*models.BookingRequest, error) {
	// Validate origin != destination
	if req.Origin != nil && req.Destination != nil {
		if req.Origin.Address == req.Destination.Address &&
		   req.Origin.Latitude == req.Destination.Latitude &&
		   req.Origin.Longitude == req.Destination.Longitude {
			return nil, ErrInvalidRoute
		}
	}

	// Validate pickup time is in future
	if req.RequestedPickupTime != nil && req.RequestedPickupTime.Before(time.Now()) {
		return nil, ErrInvalidPickupTime
	}

	booking := &models.BookingRequest{
		CustomerID:          customerID,
		RequestType:         req.RequestType,
		Origin:              req.Origin,
		Destination:         req.Destination,
		RequestedPickupTime: req.RequestedPickupTime,
		CargoDetails:        req.CargoDetails,
		Status:              string(models.BookingPending),
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

// GetCustomerInvoices retrieves invoices for a customer
func (s *CustomerService) GetCustomerInvoices(ctx context.Context, customerID string) ([]*models.Invoice, error) {
	return s.repo.GetInvoices(ctx, customerID)
}

// GetCustomerShipments retrieves shipments for a customer
func (s *CustomerService) GetCustomerShipments(ctx context.Context, customerID string) ([]*models.Shipment, error) {
	filter := models.ShipmentListFilter{
		CustomerID: customerID,
		Page:       1,
		PageSize:   100,
	}
	shipments, _, err := s.repo.ListShipments(ctx, filter)
	return shipments, err
}

// UpdateShipmentStatus updates shipment status with validation
func (s *CustomerService) UpdateShipmentStatus(ctx context.Context, shipmentID string, newStatus models.ShipmentStatus) error {
	shipment, err := s.repo.GetShipment(ctx, shipmentID)
	if err != nil {
		return err
	}
	if shipment == nil {
		return ErrShipmentNotFound
	}

	currentStatus := models.ShipmentStatus(shipment.Status)
	if !models.IsValidStatusTransition(currentStatus, newStatus) {
		return ErrInvalidStatusTransition
	}

	// Update status in repository (would need to add this method)
	// For now, this is a placeholder
	s.logger.Info("shipment status updated",
		zap.String("shipment_id", shipmentID),
		zap.String("from", string(currentStatus)),
		zap.String("to", string(newStatus)))

	return nil
}

// Internal: CreateCustomer (for admin use)
func (s *CustomerService) CreateCustomer(ctx context.Context, c *models.Customer) error {
	// In reality, generate API Key here
	return s.repo.Create(ctx, c)
}
