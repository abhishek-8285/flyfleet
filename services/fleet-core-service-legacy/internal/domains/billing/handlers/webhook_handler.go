package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// StripeWebhook handles event notifications from Stripe
func (h *BillingHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.errorJSON(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// In production, verify signature here using stripe.Webhook.ConstructEvent

	var event struct {
		Type string `json:"type"`
		Data struct {
			Object map[string]interface{} `json:"object"`
		} `json:"data"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		h.errorJSON(w, "Invalid Payload", http.StatusBadRequest)
		return
	}

	if event.Type == "payment_intent.succeeded" {
		obj := event.Data.Object
		meta, _ := obj["metadata"].(map[string]interface{})
		invoiceIDStr, _ := meta["invoice_id"].(string)
		amountFloat, _ := obj["amount"].(float64)

		if invoiceIDStr != "" {
			invoiceID := uuid.MustParse(invoiceIDStr)
			// Trigger payment recording (Divide by 100 for cents-to-dollars)
			err := h.service.RecordPayment(r.Context(), invoiceID, amountFloat/100)
			if err != nil {
				h.logger.Error("Failed to record stripe payment", zap.Error(err), zap.String("invoice", invoiceIDStr))
				h.errorJSON(w, "Internal Error", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// RazorpayWebhook handles event notifications from Razorpay
func (h *BillingHandler) RazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	// Logic to verify Razorpay signature and handle 'payment.captured'
	h.json(w, map[string]string{"status": "ok"}, http.StatusOK)
}
