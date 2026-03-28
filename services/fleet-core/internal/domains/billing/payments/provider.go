package payments

import (
	"context"
	"github.com/google/uuid"
)

// PaymentProvider defines the interface for external payment gateways
type PaymentProvider interface {
	// CreatePaymentIntent initializes a session/intent for a specific amount
	CreatePaymentIntent(ctx context.Context, invoiceID uuid.UUID, amount float64, currency string) (string, error)

	// RefundPayment handles partial or full refunds
	RefundPayment(ctx context.Context, paymentID string, amount float64) error

	// GetStatus retrieves the current status of a payment from the provider
	GetStatus(ctx context.Context, paymentID string) (string, error)
}
