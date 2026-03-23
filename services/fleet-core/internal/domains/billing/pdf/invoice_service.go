package pdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/pkg/qrcode"
)

// GenerateInvoiceHTML generates HTML from invoice data using professional template
func GenerateInvoiceHTML(invoice *models.Invoice, items []models.InvoiceItem) string {
	// Use professional GST-compliant template
	data := DefaultProfessionalInvoiceTemplateData()

	// Map invoice to template data
	data.InvoiceNumber = invoice.InvoiceNumber
	data.InvoiceDate = invoice.InvoiceDate.Format("02 Jan 2006")
	data.DueDate = invoice.DueDate.Format("02 Jan 2006")
	data.CustomerName = invoice.CustomerName
	data.Consignee = invoice.CustomerName // Default
	if cName, ok := invoice.Metadata["consignee_name"].(string); ok {
		data.Consignee = cName
	}
	data.Notes = invoice.Notes
	data.TotalAmount = invoice.TotalAmount
	data.AmountPaid = invoice.AmountPaid
	data.BalanceDue = invoice.BalanceDue
	data.AmountInWords = AmountToWords(invoice.TotalAmount)

	// Map Logistics Metadata
	if vehicleNo, ok := invoice.Metadata["vehicle_no"].(string); ok {
		data.VehicleNo = vehicleNo
	}
	if lrNo, ok := invoice.Metadata["lr_no"].(string); ok {
		data.LRNo = lrNo
	}
	if ewayBill, ok := invoice.Metadata["e_way_bill"].(string); ok {
		data.EWayBill = ewayBill
	}
	if pos, ok := invoice.Metadata["place_of_supply"].(string); ok {
		data.PlaceOfSupply = pos
	}
	data.Transporter = "FlyFleet Logistics"
	data.Industry = strings.ToLower(invoice.Category)

	// 🛠 INDUSTRY DATA MAPPER: 100% Exhaustive for all 12 industries
	switch data.Industry {
	case "jeweller":
		data.CustomHeaders = []string{"Purity", "Gross Weight", "Net Weight", "Stone Chg"}
	case "bakery":
		data.CustomHeaders = []string{"Batch No", "Expiry Date", "Mfd Date"}
	case "automobile":
		data.CustomHeaders = []string{"Chassis No", "Engine No", "Warranty"}
	case "freelancer":
		data.CustomHeaders = []string{"Service Type", "Hours", "Project Code"}
	case "hotel":
		data.CustomHeaders = []string{"Room No", "Check-In", "Check-Out"}
	case "retail":
		data.CustomHeaders = []string{"Brand", "Size", "Color"}
	case "electrical":
		data.CustomHeaders = []string{"Model No", "Warranty", "Voltage"}
	case "furniture":
		data.CustomHeaders = []string{"Material", "Dimensions", "Wood Type"}
	case "garments":
		data.CustomHeaders = []string{"Fabric", "Size", "Design No"}
	case "stationery":
		data.CustomHeaders = []string{"Brand", "GSM/Unit"}
	case "transporter":
		data.CustomHeaders = []string{"Vehicle No", "LR No", "E-Way Bill"}
	case "general":
		data.CustomHeaders = []string{"Category", "Remarks"}
	}

	// 🖼 LOGO OPTIMIZATION: Use the 7KB optimized logo instead of the 2.5MB one
	logoData, _ := os.ReadFile("../../logo_tiny.jpg")
	if len(logoData) > 0 {
		data.CompanyLogo = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(logoData)
	}

	// Generate Light-Weight Vector QR Codes (SVG)
	upiLink := qrcode.GenerateUPILink("flyfleet@upi", data.Company_name, fmt.Sprintf("%.2f", invoice.TotalAmount), "Invoice "+invoice.InvoiceNumber)
	qrSVG, err := qrcode.GenerateSVG(upiLink, 100)
	if err == nil {
		data.QRCode = qrSVG
		data.ShowQRCode = true
	}

	verifyURL := fmt.Sprintf("https://verify.flyfleet.com/invoice/%s", invoice.ID.String())
	if verifySVG, err := qrcode.GenerateSVG(verifyURL, 60); err == nil {
		data.VerificationQR = verifySVG
	}

	// Calculate totals and Map Items
	var taxableValue, totalTax float64
	for i, item := range items {
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemTax := itemSubtotal * (item.TaxPercent / 100)
		taxableValue += itemSubtotal
		totalTax += itemTax

		// Map to template item
		itemData := InvoiceItemData{
			SlNo:        i + 1,
			Description: item.Description,
			HSNCode:     item.HSNCode,
			Unit:        "Nos",
			Quantity:    float64(item.Quantity),
			UnitPrice:   item.UnitPrice,
			Discount:    0,
			TaxPercent:  item.TaxPercent,
			TaxAmount:   itemTax,
			CESS:        0,
			AdditionalCESS: 0,
			Total:       itemSubtotal + itemTax,
		}

		// INDUSTRY DATA FILL: Map metadata to custom columns
		for _, header := range data.CustomHeaders {
			val := "-"
			key := strings.ReplaceAll(strings.ToLower(header), " ", "_")
			if itemVal, ok := item.CustomFields[key].(string); ok {
				val = itemVal
			}
			itemData.CustomValues = append(itemData.CustomValues, val)
		}
		data.Items = append(data.Items, itemData)
	}

	data.TaxableValue = taxableValue
	data.TotalTax = totalTax

	// 🛡️ DYNAMIC GST ENGINE: Intra-state (27) vs Inter-state
	companyState := "27" // Maharashtra
	customerState := ""
	if len(data.PlaceOfSupply) >= 3 {
		// Extract "06" from "Haryana (06)"
		customerState = data.PlaceOfSupply[len(data.PlaceOfSupply)-3 : len(data.PlaceOfSupply)-1]
	}

	if customerState != "" && customerState != companyState {
		// INTER-STATE: Apply IGST
		data.IGST = totalTax
		data.CGST = 0
		data.SGST = 0
	} else {
		// INTRA-STATE: Split into CGST + SGST
		data.IGST = 0
		data.CGST = totalTax / 2
		data.SGST = totalTax / 2
	}

	return GenerateProfessionalInvoiceHTML(invoice, items, data)
}

// GenerateInvoiceHTMLWithConfig generates HTML with custom billing configuration
func GenerateInvoiceHTMLWithConfig(invoice *models.Invoice, items []models.InvoiceItem, config *models.BillingConfiguration) string {
	// Use config for template data
	data := DefaultProfessionalInvoiceTemplateData()

	// Override with custom config
	data.Company_name = config.CompanyName
	data.CompanyAddress = config.CompanyAddress
	data.CompanyGSTIN = config.CompanyGSTIN
	data.CompanyPAN = config.CompanyPAN
	data.CompanyEmail = config.CompanyEmail
	data.CompanyPhone = config.CompanyPhone
	data.CompanyWebsite = config.CompanyWebsite
	data.CompanyLogo = config.CompanyLogoURL
	data.PrimaryColor = config.PrimaryColor
	data.PaymentInfo.BankName = config.BankName
	data.PaymentInfo.AccountName = config.BankAccountName
	data.PaymentInfo.AccountNumber = config.BankAccountNumber
	data.PaymentInfo.IFSC = config.BankIFSC
	data.PaymentInfo.Branch = config.BankBranch
	data.PaymentInfo.UPI = config.BankUPI
	data.Terms = strings.Join(config.InvoiceTerms, "\n")
	data.Notes = config.InvoiceNotes
	data.ShowHSN = config.ShowHSN
	data.ShowQRCode = config.ShowQRCode
	data.ShowPaymentInfo = config.ShowPaymentInfo

	// Map invoice
	data.InvoiceNumber = invoice.InvoiceNumber
	data.InvoiceDate = invoice.InvoiceDate.Format("02 Jan 2006")
	data.DueDate = invoice.DueDate.Format("02 Jan 2006")
	data.CustomerName = invoice.CustomerName
	data.Consignee = invoice.CustomerName
	if cName, ok := invoice.Metadata["consignee_name"].(string); ok {
		data.Consignee = cName
	}
	data.TotalAmount = invoice.TotalAmount
	data.AmountPaid = invoice.AmountPaid
	data.BalanceDue = invoice.BalanceDue
	data.AmountInWords = AmountToWords(invoice.TotalAmount)

	// Map Logistics Metadata
	if vehicleNo, ok := invoice.Metadata["vehicle_no"].(string); ok {
		data.VehicleNo = vehicleNo
	}
	if lrNo, ok := invoice.Metadata["lr_no"].(string); ok {
		data.LRNo = lrNo
	}
	if ewayBill, ok := invoice.Metadata["e_way_bill"].(string); ok {
		data.EWayBill = ewayBill
	}
	if pos, ok := invoice.Metadata["place_of_supply"].(string); ok {
		data.PlaceOfSupply = pos
	}
	data.Transporter = "FlyFleet Logistics"
	data.Industry = strings.ToLower(invoice.Category)

	// 🛠 INDUSTRY DATA MAPPER: 100% Exhaustive for all 12 industries
	switch data.Industry {
	case "jeweller":
		data.CustomHeaders = []string{"Purity", "Gross Weight", "Net Weight", "Stone Chg"}
	case "bakery":
		data.CustomHeaders = []string{"Batch No", "Expiry Date", "Mfd Date"}
	case "automobile":
		data.CustomHeaders = []string{"Chassis No", "Engine No", "Warranty"}
	case "freelancer":
		data.CustomHeaders = []string{"Service Type", "Hours", "Project Code"}
	case "hotel":
		data.CustomHeaders = []string{"Room No", "Check-In", "Check-Out"}
	case "retail":
		data.CustomHeaders = []string{"Brand", "Size", "Color"}
	case "electrical":
		data.CustomHeaders = []string{"Model No", "Warranty", "Voltage"}
	case "furniture":
		data.CustomHeaders = []string{"Material", "Dimensions", "Wood Type"}
	case "garments":
		data.CustomHeaders = []string{"Fabric", "Size", "Design No"}
	case "stationery":
		data.CustomHeaders = []string{"Brand", "GSM/Unit"}
	case "transporter":
		data.CustomHeaders = []string{"Vehicle No", "LR No", "E-Way Bill"}
	case "general":
		data.CustomHeaders = []string{"Category", "Remarks"}
	}

	// 🖼 LOGO OPTIMIZATION: Use the 7KB optimized logo instead of the 2.5MB one
	logoData, _ := os.ReadFile("../../logo_tiny.jpg")
	if len(logoData) > 0 {
		data.CompanyLogo = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(logoData)
	}

	// Generate Light-Weight Vector QR Codes (SVG)
	upiLink := qrcode.GenerateUPILink("flyfleet@upi", data.Company_name, fmt.Sprintf("%.2f", invoice.TotalAmount), "Invoice "+invoice.InvoiceNumber)
	qrSVG, err := qrcode.GenerateSVG(upiLink, 100)
	if err == nil {
		data.QRCode = qrSVG
		data.ShowQRCode = true
	}

	verifyURL := fmt.Sprintf("https://verify.flyfleet.com/invoice/%s", invoice.ID.String())
	if verifySVG, err := qrcode.GenerateSVG(verifyURL, 60); err == nil {
		data.VerificationQR = verifySVG
	}

	// Calculate totals and Map Items
	var taxableValue, totalTax float64
	for i, item := range items {
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemTax := itemSubtotal * (item.TaxPercent / 100)
		taxableValue += itemSubtotal
		totalTax += itemTax

		// Map to template item
		itemData := InvoiceItemData{
			SlNo:        i + 1,
			Description: item.Description,
			HSNCode:     item.HSNCode,
			Unit:        "Nos",
			Quantity:    float64(item.Quantity),
			UnitPrice:   item.UnitPrice,
			Discount:    0,
			TaxPercent:  item.TaxPercent,
			TaxAmount:   itemTax,
			CESS:        0,
			AdditionalCESS: 0,
			Total:       itemSubtotal + itemTax,
		}

		// INDUSTRY DATA FILL: Map metadata to custom columns
		for _, header := range data.CustomHeaders {
			val := "-"
			key := strings.ReplaceAll(strings.ToLower(header), " ", "_")
			if itemVal, ok := item.CustomFields[key].(string); ok {
				val = itemVal
			}
			itemData.CustomValues = append(itemData.CustomValues, val)
		}
		data.Items = append(data.Items, itemData)
	}

	data.TaxableValue = taxableValue
	data.TotalTax = totalTax

	// 🛡️ DYNAMIC GST ENGINE: Intra-state (27) vs Inter-state
	companyState := "27" // Maharashtra
	customerState := ""
	if len(data.PlaceOfSupply) >= 3 {
		// Extract "06" from "Haryana (06)"
		customerState = data.PlaceOfSupply[len(data.PlaceOfSupply)-3 : len(data.PlaceOfSupply)-1]
	}

	if customerState != "" && customerState != companyState {
		// INTER-STATE: Apply IGST
		data.IGST = totalTax
		data.CGST = 0
		data.SGST = 0
	} else {
		// INTRA-STATE: Split into CGST + SGST
		data.IGST = 0
		data.CGST = totalTax / 2
		data.SGST = totalTax / 2
	}

	return GenerateProfessionalInvoiceHTML(invoice, items, data)
}

// GenerateInvoicePDF generates a PDF for an invoice
func (s *InvoiceService) GeneratePDF(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) ([]byte, error) {
	html := GenerateInvoiceHTML(invoice, items)
	return s.generator.GenerateFromHTML(ctx, html, DefaultA4Portrait())
}

// GeneratePDFWithConfig generates PDF with custom billing configuration
func (s *InvoiceService) GeneratePDFWithConfig(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem, config *models.BillingConfiguration) ([]byte, error) {
	html := GenerateInvoiceHTMLWithConfig(invoice, items, config)
	return s.generator.GenerateFromHTML(ctx, html, DefaultA4Portrait())
}

// GeneratePDFAndSave generates PDF and saves to file
func (s *InvoiceService) GeneratePDFAndSave(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem, filename string) error {
	pdfData, err := s.GeneratePDF(ctx, invoice, items)
	if err != nil {
		return err
	}
	return SaveToFile(pdfData, filename)
}

// GeneratePDFBase64 generates PDF and returns base64 encoded string
func (s *InvoiceService) GeneratePDFBase64(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) (string, error) {
	pdfData, err := s.GeneratePDF(ctx, invoice, items)
	if err != nil {
		return "", err
	}
	return EncodeToBase64(pdfData), nil
}
