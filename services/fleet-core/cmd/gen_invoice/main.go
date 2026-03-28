package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/pdf"
	"github.com/google/uuid"
)

func main() {
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Create a dummy invoice (Matching what we seeded)
	invoice := &models.Invoice{
		ID:            uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		OrgID:         orgID,
		InvoiceNumber: "INV-2026-001",
		CustomerName:  "Reliance Logistics",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 15),
		Status:        "issued",
		Currency:      "INR",
		TotalAmount:   59000.00,
		AmountPaid:    0.00,
		BalanceDue:    59000.00,
		Notes:         "Transport services for March 2026. This is a computer-generated invoice.",
	}

	// Create dummy items
	items := []models.InvoiceItem{
		{
			Description: "Full Truck Load (FTL) - Mumbai to Pune",
			Quantity:    2,
			UnitPrice:   25000.00,
			TaxPercent:  18.0,
			Total:       59000.00,
		},
	}

	// Generate HTML using the professional template
	html := pdf.GenerateInvoiceHTML(invoice, items)

	// Save to file
	filename := "generated_invoice.html"
	err := os.WriteFile(filename, []byte(html), 0644)
	if err != nil {
		log.Fatalf("Failed to save invoice: %v", err)
	}

	fmt.Printf("✅ Invoice generated successfully: %s\n", filename)

	// Also print the absolute path
	pwd, _ := os.Getwd()
	fmt.Printf("Absolute path: %s/%s\n", pwd, filename)
}
