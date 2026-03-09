package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/customer_v2/repository"
	"go.uber.org/zap"
)

type CustomerNotificationService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewCustomerNotificationService(repo repository.CustomerV2Repository, logger *zap.Logger) *CustomerNotificationService {
	return &CustomerNotificationService{
		repo:   repo,
		logger: logger,
	}
}

// NotifyShipmentEvent triggers user notifications (SMS, Email, WhatsApp, Push) based on shipment lifecycle events
func (s *CustomerNotificationService) NotifyShipmentEvent(ctx context.Context, shipment *models.Shipment, eventType string) error {
	s.logger.Info("Dispatching notification for shipment event", 
		zap.String("shipment_id", shipment.ID.String()),
		zap.String("event_type", eventType),
	)

	// Fetch customer details for contact info
	customer, err := s.repo.GetCustomerByID(ctx, shipment.CustomerID)
	if err != nil || customer == nil {
		s.logger.Error("Could not fetch customer for notification dispatch")
		return err
	}

	// Determine message template based on event
	var message string
	switch eventType {
	case "booking_confirmed":
		message = "Your booking request has been confirmed and a vehicle is being assigned."
	case "shipment_started":
		message = "Your shipment has started its journey."
	case "checkpoint_reached":
		message = "Your shipment has reached a checkpoint."
	case "shipment_delivered":
		message = "Your shipment has been securely delivered."
	case "invoice_generated":
		message = "An invoice has been generated for your recent shipment."
	default:
		message = "Update regarding your shipment."
	}

	// Stub: Send via multiple channels
	s.dispatchEmail(customer.ContactEmail, message)
	s.dispatchSMS(customer.ContactPhone, message)
	
	return nil
}

func (s *CustomerNotificationService) dispatchEmail(email *string, message string) {
	if email != nil {
		s.logger.Debug("Sending email notification", zap.String("email", *email), zap.String("message", message))
		// emailClient.Send(...)
	}
}

func (s *CustomerNotificationService) dispatchSMS(phone *string, message string) {
	if phone != nil {
		s.logger.Debug("Sending SMS notification", zap.String("phone", *phone), zap.String("message", message))
		// smsClient.Send(...)
	}
}
