package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/flyfleet/fleet-core/internal/pkg/pdfoptimizer"
	"github.com/google/uuid"
)

func generateFor(category string, filename string) {
	os.Setenv("CHROME_PATH", "/data/data/com.termux/files/usr/bin/chromium-browser")

	invoice := &models.Invoice{
		ID:            uuid.New(),
		OrgID:         uuid.New(),
		InvoiceNumber: "INV-" + category + "-001",
		Category:      category,
		CustomerName:  "Industrial " + category + " Client",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 7),
		Status:        models.StatusPending,
		Currency:      "INR",
		TotalAmount:   56000.00,
		Metadata: map[string]interface{}{
			"vehicle_no":     "DL-01-LH-8783",
			"consignee_name": "Sample Consignee Ltd",
		},
	}

	items := []models.InvoiceItem{
		{
			Description: category + " Premium High-Value Supply",
			Quantity:    1,
			UnitPrice:   50000.00,
			TaxPercent:  12.0,
			HSNCode:     "7113", // Jewelry HSN
			CustomFields: map[string]interface{}{
				"purity":       "22 Karat",
				"gross_weight": "12.50g",
				"net_weight":   "11.20g",
				"stone_chg":    "Rs. 4500",
				"batch_no":     "B-99281",
				"expiry_date":  "21-Dec-2027",
				"mfd_date":     "21-Mar-2026",
			},
		},
	}

	generator := pdf.NewPDFGenerator(60 * time.Second)
	html := pdf.GenerateInvoiceHTML(invoice, items)

	pdfData, err := generator.GenerateFromHTML(context.Background(), html, pdf.DefaultA4Portrait())
	if err != nil {
		fmt.Printf("❌ Failed to generate %s: %v\n", category, err)
		return
	}

	optimized, _ := pdfoptimizer.Optimize(pdfData)
	os.WriteFile(filename, optimized, 0644)
	fmt.Printf("✅ Success: %s created (%d bytes)\n", filename, len(optimized))
}

func main() {
	generateFor("JEWELLER", "jeweller_invoice.pdf")
	generateFor("BAKERY", "bakery_invoice.pdf")
}
