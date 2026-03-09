package service

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestReceiptOCRService_ExtractReceipt(t *testing.T) {
	svc := NewReceiptOCRService(nil, zap.NewNop())

	result, err := svc.ExtractReceipt(context.Background(), "https://cdn.example.com/r.png?vendor=hpcl&total=1234.56&gst=188.40&x=27AAACB2230M1ZT")
	if err != nil {
		t.Fatalf("ExtractReceipt returned error: %v", err)
	}
	if result.VendorGSTIN != "27AAACB2230M1ZT" {
		t.Fatalf("VendorGSTIN = %q, want %q", result.VendorGSTIN, "27AAACB2230M1ZT")
	}
	if result.TotalAmount != 1234.56 {
		t.Fatalf("TotalAmount = %v, want 1234.56", result.TotalAmount)
	}
	if result.GSTAmount != 188.40 {
		t.Fatalf("GSTAmount = %v, want 188.40", result.GSTAmount)
	}
	if !result.Valid {
		t.Fatal("expected receipt to be valid")
	}
}
