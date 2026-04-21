//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	// Create mock invoice with 25 line items
	invoice := &models.Invoice{
		ID:            uuid.New(),
		TenantID:         uuid.New(),
		InvoiceNumber: "INV-2026-STRESS-001",
		CustomerName:  "Stress Test Customer Pvt Ltd",
		CustomerEmail: "billing@stresstest.com",
		BillingAddress: `456 Industrial Area, Phase II
Mumbai, Maharashtra 400001
India`,
		InvoiceDate:  time.Now(),
		DueDate:      time.Now().Add(30 * 24 * time.Hour),
		Status:       models.StatusPending,
		Currency:     "INR",
		TotalAmount:  0, // Will calculate
		AmountPaid:   0,
		BalanceDue:   0, // Will calculate
		Subtotal:     0, // Will calculate
		Notes:        "This is a pagination stress test invoice with 25 line items to verify CSS page-break-inside: avoid; rules work correctly.",
		Category:     "General",
		Metadata: map[string]interface{}{
			"vehicle_no":      "MH02AB1234",
			"lr_no":           "LR-2026-001",
			"e_way_bill":      "EWB-2026-STRESS-001",
			"place_of_supply": "Maharashtra (27)",
			"consignee_name":  "Stress Test Consignee",
		},
	}

	// Generate 25 distinct line items with alternating tax logic
	// Items 1,3,5... = IGST (inter-state simulation)
	// Items 2,4,6... = CGST/SGST (intra-state simulation)
	items := []models.InvoiceItem{
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Industrial Grade Steel Bearings - Premium Quality", Quantity: 150, UnitPrice: 1250.00, TaxPercent: 18.0, HSNCode: "8482", CustomFields: map[string]interface{}{"category": "Bearings", "remarks": "Heavy duty"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Hydraulic Pump Assembly - Model HP-5000", Quantity: 25, UnitPrice: 8500.00, TaxPercent: 28.0, HSNCode: "8413", CustomFields: map[string]interface{}{"category": "Pumps", "remarks": "Industrial"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Electronic Control Unit - ECU-2026", Quantity: 50, UnitPrice: 4200.00, TaxPercent: 18.0, HSNCode: "8537", CustomFields: map[string]interface{}{"category": "Electronics", "remarks": "Automotive grade"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Precision Machined Components Set", Quantity: 200, UnitPrice: 350.00, TaxPercent: 12.0, HSNCode: "7318", CustomFields: map[string]interface{}{"category": "Components", "remarks": "CNC machined"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "High Temperature Industrial Gasket Kit", Quantity: 500, UnitPrice: 125.00, TaxPercent: 18.0, HSNCode: "8484", CustomFields: map[string]interface{}{"category": "Sealing", "remarks": "Heat resistant"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Variable Frequency Drive - VFD-5HP", Quantity: 15, UnitPrice: 15000.00, TaxPercent: 28.0, HSNCode: "8504", CustomFields: map[string]interface{}{"category": "Drives", "remarks": "5 Horsepower"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Stainless Steel Fastener Assortment Box", Quantity: 100, UnitPrice: 450.00, TaxPercent: 18.0, HSNCode: "7318", CustomFields: map[string]interface{}{"category": "Fasteners", "remarks": "SS304 grade"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Pneumatic Cylinder - Double Acting 50mm Bore", Quantity: 40, UnitPrice: 2800.00, TaxPercent: 28.0, HSNCode: "8412", CustomFields: map[string]interface{}{"category": "Pneumatics", "remarks": "ISO 6431"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Digital Pressure Sensor with Display", Quantity: 75, UnitPrice: 1850.00, TaxPercent: 18.0, HSNCode: "9026", CustomFields: map[string]interface{}{"category": "Sensors", "remarks": "0-10 bar range"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Conveyor Belt Rubber - 500mm Width per Meter", Quantity: 300, UnitPrice: 890.00, TaxPercent: 28.0, HSNCode: "4010", CustomFields: map[string]interface{}{"category": "Conveyor", "remarks": "Heavy duty"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "PLC Module - 16 Input/Output Digital", Quantity: 20, UnitPrice: 12500.00, TaxPercent: 18.0, HSNCode: "8537", CustomFields: map[string]interface{}{"category": "Automation", "remarks": "Modbus RTU"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Industrial LED Flood Light 100W IP65", Quantity: 80, UnitPrice: 2200.00, TaxPercent: 12.0, HSNCode: "9405", CustomFields: map[string]interface{}{"category": "Lighting", "remarks": "Outdoor rated"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Gearbox Reduction Unit - 10:1 Ratio", Quantity: 30, UnitPrice: 9800.00, TaxPercent: 28.0, HSNCode: "8483", CustomFields: map[string]interface{}{"category": "Gears", "remarks": "Helical type"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Proximity Sensor Inductive M12 NPN", Quantity: 150, UnitPrice: 680.00, TaxPercent: 18.0, HSNCode: "8536", CustomFields: map[string]interface{}{"category": "Sensors", "remarks": "4mm sensing"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Motor Coupling Flexible - 25mm Bore", Quantity: 60, UnitPrice: 1450.00, TaxPercent: 28.0, HSNCode: "8483", CustomFields: map[string]interface{}{"category": "Couplings", "remarks": "Spider type"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Cable Tray Galvanized Steel 100x50mm", Quantity: 250, UnitPrice: 420.00, TaxPercent: 18.0, HSNCode: "7314", CustomFields: map[string]interface{}{"category": "Cabling", "remarks": "Per meter"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Solenoid Valve 3/2 Way 24VDC", Quantity: 100, UnitPrice: 1280.00, TaxPercent: 28.0, HSNCode: "8481", CustomFields: map[string]interface{}{"category": "Valves", "remarks": "NC type"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Timing Belt Pulley - 40 Teeth T5 Profile", Quantity: 80, UnitPrice: 520.00, TaxPercent: 18.0, HSNCode: "8483", CustomFields: map[string]interface{}{"category": "Power Transmission", "remarks": "Aluminum"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Linear Guide Rail - 500mm Length", Quantity: 40, UnitPrice: 3200.00, TaxPercent: 28.0, HSNCode: "8483", CustomFields: map[string]interface{}{"category": "Linear Motion", "remarks": "MGN15H"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Rotary Encoder Incremental 1024 PPR", Quantity: 35, UnitPrice: 4500.00, TaxPercent: 18.0, HSNCode: "8543", CustomFields: map[string]interface{}{"category": "Encoders", "remarks": "Quadrature output"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Air Compressor Filter Element Set", Quantity: 200, UnitPrice: 380.00, TaxPercent: 28.0, HSNCode: "8421", CustomFields: map[string]interface{}{"category": "Filters", "remarks": "5 micron"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Safety Light Curtain - 600mm Detection Zone", Quantity: 12, UnitPrice: 28000.00, TaxPercent: 18.0, HSNCode: "8536", CustomFields: map[string]interface{}{"category": "Safety", "remarks": "Type 4 PLe"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Ball Screw Assembly - 16mm Diameter 400mm Length", Quantity: 25, UnitPrice: 6800.00, TaxPercent: 28.0, HSNCode: "8483", CustomFields: map[string]interface{}{"category": "Linear Motion", "remarks": "C7 precision"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Industrial Ethernet Switch 8 Port Gigabit", Quantity: 18, UnitPrice: 8900.00, TaxPercent: 18.0, HSNCode: "8517", CustomFields: map[string]interface{}{"category": "Networking", "remarks": "DIN rail mount"}},
		{ID: uuid.New(), InvoiceID: invoice.ID, Description: "Servo Motor 400W with Brake 220V AC", Quantity: 10, UnitPrice: 32000.00, TaxPercent: 28.0, HSNCode: "8501", CustomFields: map[string]interface{}{"category": "Motors", "remarks": "With encoder"}},
	}

	// Calculate totals
	var subtotal, totalTax float64
	for _, item := range items {
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemTax := itemSubtotal * (item.TaxPercent / 100)
		subtotal += itemSubtotal
		totalTax += itemTax
	}

	invoice.Subtotal = subtotal
	invoice.TotalAmount = subtotal + totalTax
	invoice.BalanceDue = invoice.TotalAmount

	fmt.Printf("Generated invoice with %d line items\n", len(items))
	fmt.Printf("Subtotal: ₹%.2f\n", subtotal)
	fmt.Printf("Total Tax: ₹%.2f\n", totalTax)
	fmt.Printf("Grand Total: ₹%.2f\n", invoice.TotalAmount)

	// Create PDF generator with explicit Chromium path for Termux
	os.Setenv("CHROME_PATH", "/data/data/com.termux/files/usr/lib/chromium/chrome")
	generator := pdf.NewPDFGenerator(120 * time.Second)
	invoiceService := pdf.NewInvoiceService(generator)

	// Determine output directory
	outputDir := filepath.Join(os.Getenv("HOME"), "flyfleet", "test-results")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(outputDir, "invoice_pagination_stress_test.pdf")

	// Generate PDF
	fmt.Println("Generating PDF...")
	err := invoiceService.GeneratePDFAndSave(ctx, invoice, items, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "PDF generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: PDF saved to %s\n", outputPath)
}
