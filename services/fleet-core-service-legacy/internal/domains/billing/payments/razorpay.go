package payments

import (
	"context"

	"github.com/google/uuid"
)

type RazorpayAdapter struct {
	keyID     string
	keySecret string
}

func NewRazorpayAdapter(keyID, keySecret string) *RazorpayAdapter {
	return &RazorpayAdapter{
		keyID:     keyID,
		keySecret: keySecret,
	}
}

func (r *RazorpayAdapter) CreatePaymentIntent(ctx context.Context, invoiceID uuid.UUID, amount float64, currency string) (string, error) {
	// Logic to call Razorpay Order API
	return "pay_123_abc", nil
}

func (r *RazorpayAdapter) RefundPayment(ctx context.Context, paymentID string, amount float64) error {
	// Logic to call Razorpay Refund API
	return nil
}

func (r *RazorpayAdapter) GetStatus(ctx context.Context, paymentID string) (string, error) {
	// Logic to query Razorpay Payment status
	return "captured", nil
}
