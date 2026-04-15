package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/flyfleet/fleet-core/internal/pkg/pdfoptimizer"
	"github.com/flyfleet/fleet-core/internal/pkg/qrcode"
	"github.com/google/uuid"
)

func main() {
	// 1. Set Termux Chrome Path
	os.Setenv("CHROME_PATH", "/data/data/com.termux/files/usr/bin/chromium-browser")

	// 2. High-Fidelity Professional Data (e-invoice format)
	now := time.Now()
	invoice := &models.Invoice{
		ID:             uuid.New(),
		InvoiceNumber:  "INV-E2E-001",
		CustomerName:   "Ishvar International Textiles",
		CustomerEmail:  "accounts@ishvar.com",
		BillingAddress: "Sector 18, Gurugram, Haryana 122001",
		InvoiceDate:    now,
		DueDate:        now.AddDate(0, 0, 15),
		Status:         "issued",
		Currency:       "INR",
		TotalAmount:    147500.00,
		BalanceDue:     147500.00,
		Metadata: map[string]interface{}{
			"vehicle_no":      "DL-01-LH-8783",
			"lr_no":           "LR-MASTER-999",
			"eway_bill":       "751590248946",
			"place_of_supply": "Haryana (06)",
			"e_way_bill_date": "20 Mar 2026",
			"doc_date":        "19 Mar 2026",
			"reverse_charge":  false,
			"irn":             "ANVJ1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ123456",
			"ack_no":          "ACK-987654321",
			"ack_date":        "20 Mar 2026 10:30 AM",
		},
	}

	items := []models.InvoiceItem{
		{
			Description: "Full Truck Load Freight - Mumbai to Delhi",
			Quantity:    1,
			UnitPrice:   125000.00,
			TaxPercent:  18.0,
			Total:       147500.00,
		},
	}

	// 3. Build Extended Template Data
	// Compute totals per spec:
	// Subtotal = sum of line totals (amount including tax)
	// TaxableValue = sum of line pre-tax totals (qty * rate)
	// TotalTax = sum of tax amounts
	var subtotal, taxableValue, totalTax float64
	for _, it := range items {
		qty := float64(it.Quantity)
		rate := it.UnitPrice
		preTax := qty * rate
		tax := preTax * (it.TaxPercent / 100)
		lineTotal := preTax + tax

		subtotal += lineTotal
		taxableValue += preTax
		totalTax += tax
	}
	total := subtotal // Grand total equals subtotal (no extra charges)
	// Create tax description string for financial summary
	var taxDesc string
	if len(items) > 0 && totalTax > 0 {
		tr := items[0].TaxPercent
		if totalTax > 0 { // we used IGST logic later
			taxDesc = fmt.Sprintf("IGST @ %.2f%%", tr)
		} else {
			taxDesc = fmt.Sprintf("CGST @ %.2f%% + SGST @ %.2f%%", tr/2, tr/2)
		}
	}

	data := pdf.ProfessionalInvoiceTemplateData{
		// Issuer
		Company_name:   "FlyFleet Technologies",
		CompanyAddress: "Mumbai, Maharashtra, India - 400001",
		CompanyGSTIN:   "27AAAAA0000A1Z5",
		CompanyPAN:     "ABCDE1234F",
		CompanyEmail:   "support@flyfleet.com",
		CompanyPhone:   "+91 9876543210",
		CompanyWebsite: "https://flyfleet.com",
		// Invoice
		InvoiceNumber: invoice.InvoiceNumber,
		InvoiceDate:   invoice.InvoiceDate.Format("02 Jan 2006"),
		DueDate:       invoice.DueDate.Format("02 Jan 2006"),
		PlaceOfSupply: "Haryana (06)",
		ReverseCharge: "No",
		CopyType:      "Original Copy",
		PageNumber:    "Page 1 of 1",
		// Transporter
		Transporter:  "FlyFleet Logistics",
		VehicleNo:    "DL-01-LH-8783",
		DocNo:        "LR-MASTER-999",
		DocDate:      "19 Mar 2026",
		EWayBill:     "751590248946",
		EWayBillDate: "20 Mar 2026",
		// Party
		BillingName:     invoice.CustomerName,
		BillingGSTIN:    "06AABCI1234R1ZX",
		BillingMobile:   "+91 9988776655",
		BillingEmail:    invoice.CustomerEmail,
		BillingAddress:  invoice.BillingAddress,
		ShippingName:    invoice.CustomerName,
		ShippingGSTIN:   "06AABCI1234R1ZX",
		ShippingMobile:  "+91 9988776655",
		ShippingEmail:   invoice.CustomerEmail,
		ShippingAddress: invoice.BillingAddress,
		// Compliance
		IRN:     "ANVJ1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ123456",
		AckNo:   "ACK-987654321",
		AckDate: "20 Mar 2026 10:30 AM",
		// Items & Financials
		Items:          make([]pdf.InvoiceItemData, 0, len(items)),
		Subtotal:       taxableValue, // net before tax
		TotalDiscount:  0,
		TaxableValue:   taxableValue, // taxable amount
		CGST:           0,
		SGST:           0,
		IGST:           totalTax,
		CESS:           0,
		AdditionalCESS: 0,
		TotalTax:       totalTax,
		TotalAmount:    total, // grand total (147500)
		AmountInWords:  pdf.AmountToWords(total),
		TaxDescription: taxDesc,
		// Payment
		BankName:            "HDFC Bank",
		AccountName:         "FlyFleet Technologies Pvt Ltd",
		AccountNumber:       "1234567890123456",
		IFSC:                "HDFC0001234",
		Branch:              "Andheri East, Mumbai",
		UPI:                 "flyfleet@upi",
		PaymentQR:           "",
		AuthorizedSignatory: "Authorized Signatory",
		PrimaryColor:        "#2563eb",
	}

	// Generate QR codes (PNG base64 data URIs)
	upiLink := qrcode.GenerateUPILink(data.UPI, data.Company_name, fmt.Sprintf("%.2f", total), "Invoice "+invoice.InvoiceNumber)
	if pngB64, err := qrcode.GenerateBase64(upiLink, 100); err == nil {
		data.PaymentQR = pngB64
	}
	verifyURL := fmt.Sprintf("https://verify.flyfleet.com/invoice/%s", invoice.ID.String())
	if verB64, err := qrcode.GenerateBase64(verifyURL, 60); err == nil {
		data.VerificationQR = verB64
	}

	// Map items
	for i, it := range items {
		qty := float64(it.Quantity)
		unitPrice := it.UnitPrice
		lineTotal := qty * unitPrice
		taxAmt := lineTotal * (it.TaxPercent / 100)
		data.Items = append(data.Items, pdf.InvoiceItemData{
			SlNo:           i + 1,
			Description:    it.Description,
			HSNCode:        "996511",
			Quantity:       qty,
			Unit:           "Nos",
			UnitPrice:      unitPrice,
			Discount:       0,
			TaxPercent:     it.TaxPercent,
			TaxAmount:      taxAmt,
			CESS:           0,
			AdditionalCESS: 0,
			Total:          lineTotal + taxAmt,
		})
	}

	// 3. Generate HTML using the new template directly
	fmt.Println("🎨 Rendering E2E Invoice Layout...")
	html := pdf.GenerateProfessionalInvoiceHTML(invoice, items, data)

	// DEBUG: Write raw HTML to file for inspection
	debugHTML := "/data/data/com.termux/files/home/flyfleet/services/fleet-core/debug_invoice.html"
	if err := os.WriteFile(debugHTML, []byte(html), 0644); err != nil {
		fmt.Printf("❌ Failed to write debug HTML: %v\n", err)
	} else {
		fmt.Printf("✅ Debug HTML saved to: %s\n", debugHTML)
	}

	// Also print first 3000 chars to terminal
	fmt.Println("--- HTML PREVIEW (first 3000 chars) ---")
	fmt.Println(html[:min(3000, len(html))])
	fmt.Println("--- END PREVIEW ---")

	// 4. PDF generation
	gen := pdf.NewPDFGenerator(60 * time.Second)
	pdfData, err := gen.GenerateFromHTML(context.Background(), html, pdf.DefaultA4Portrait())
	if err != nil {
		log.Fatalf("❌ Generation Failed: %v", err)
	}

	// 5. Optimize
	fmt.Println("⚡ Optimizing PDF...")
	optimizedData, _ := pdfoptimizer.Optimize(pdfData)

	// 6. Save with timestamped name
	filename := fmt.Sprintf("invoice_e2e_%s.pdf", now.Format("20060102_150405"))
	os.WriteFile(filename, optimizedData, 0644)
	fmt.Printf("✅ SUCCESS: %d bytes (%.1f KB) - %s\n", len(optimizedData), float64(len(optimizedData))/1024.0, filename)
}

// amountToWords converts a float amount to words (simplified)
func amountToWords(amount float64) string {
	// Placeholder: integrate your existing words conversion
	return fmt.Sprintf("Rupees %.2f only", amount)
}
