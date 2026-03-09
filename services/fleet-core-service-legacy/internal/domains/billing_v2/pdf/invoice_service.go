package pdf

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
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
	data.Notes = invoice.Notes
	data.TotalAmount = invoice.TotalAmount
	data.AmountPaid = invoice.AmountPaid
	data.BalanceDue = invoice.BalanceDue

	// Calculate totals
	var taxableValue, totalTax float64
	for _, item := range items {
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemTax := itemSubtotal * (item.TaxPercent / 100)
		taxableValue += itemSubtotal
		totalTax += itemTax
	}

	data.TaxableValue = taxableValue
	data.TotalTax = totalTax
	data.CGST = totalTax / 2
	data.SGST = totalTax / 2

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
	data.Terms = config.InvoiceTerms
	data.Notes = config.InvoiceNotes
	data.ShowHSN = config.ShowHSN
	data.ShowQRCode = config.ShowQRCode
	data.ShowPaymentInfo = config.ShowPaymentInfo

	// Map invoice
	data.InvoiceNumber = invoice.InvoiceNumber
	data.InvoiceDate = invoice.InvoiceDate.Format("02 Jan 2006")
	data.DueDate = invoice.DueDate.Format("02 Jan 2006")
	data.CustomerName = invoice.CustomerName
	data.TotalAmount = invoice.TotalAmount
	data.AmountPaid = invoice.AmountPaid
	data.BalanceDue = invoice.BalanceDue

	// Calculate totals
	var taxableValue, totalTax float64
	for _, item := range items {
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemTax := itemSubtotal * (item.TaxPercent / 100)
		taxableValue += itemSubtotal
		totalTax += itemTax
	}

	data.TaxableValue = taxableValue
	data.TotalTax = totalTax
	data.CGST = totalTax / 2
	data.SGST = totalTax / 2

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
