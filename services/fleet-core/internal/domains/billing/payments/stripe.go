package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

type StripeAdapter struct {
	apiKey string
}

func NewStripeAdapter(apiKey string) *StripeAdapter {
	stripe.Key = apiKey
	return &StripeAdapter{apiKey: apiKey}
}

func (s *StripeAdapter) CreatePaymentIntent(ctx context.Context, invoiceID uuid.UUID, amount float64, currency string) (string, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)), // Stripe uses cents
		Currency: stripe.String(currency),
	}
	params.AddMetadata("invoice_id", invoiceID.String())

	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe payment intent failed: %w", err)
	}

	return pi.ClientSecret, nil
}

func (s *StripeAdapter) RefundPayment(ctx context.Context, paymentID string, amount float64) error {
	// Logic to call Stripe Refund API
	return nil
}

func (s *StripeAdapter) GetStatus(ctx context.Context, paymentID string) (string, error) {
	// Logic to query Stripe PaymentIntent status
	return "SUCCEEDED", nil
}
