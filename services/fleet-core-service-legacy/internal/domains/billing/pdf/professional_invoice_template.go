package pdf

import (
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
)

// ProfessionalInvoiceTemplateData represents all data for invoice template
type ProfessionalInvoiceTemplateData struct {
	// Company Info
	Company_name    string
	CompanyLogo     string // Base64 or URL
	CompanyAddress  string
	CompanyGSTIN    string
	CompanyPAN      string
	CompanyEmail    string
	CompanyPhone    string
	CompanyWebsite  string

	// Customer Info
	CustomerName    string
	CustomerAddress string
	CustomerGSTIN   string
	CustomerPAN     string
	CustomerEmail   string
	CustomerPhone   string

	// Invoice Details
	InvoiceNumber   string
	InvoiceDate     string
	DueDate         string
	PlaceOfSupply   string
	ReverseCharge   string

	// Items
	Items           []InvoiceItemData

	// Totals
	Subtotal        float64
	TaxableValue    float64
	CGST            float64
	SGST            float64
	IGST            float64
	TotalTax        float64
	TotalAmount     float64
	AmountPaid      float64
	BalanceDue      float64

	// Additional
	Notes           string
	Terms           []string
	PaymentInfo     PaymentInfo
	QRCode          string // UPI QR code data

	// Styling
	PrimaryColor    string
	Theme           string

	// Visibility
	ShowQRCode      bool
	ShowPaymentInfo bool
	ShowHSN         bool
}

// InvoiceItemData represents a line item
type InvoiceItemData struct {
	ID              int
	Description     string
	HSNCode         string
	Quantity        int
	Unit            string
	Rate            float64
	TaxPercent      float64
	TaxAmount       float64
	Total           float64
}

// PaymentInfo represents payment details
type PaymentInfo struct {
	BankName        string
	AccountName     string
	AccountNumber   string
	IFSC            string
	UPI             string
	Branch          string
}

// GenerateProfessionalInvoiceHTML generates a professional GST-compliant invoice
func GenerateProfessionalInvoiceHTML(invoice *models.Invoice, items []models.InvoiceItem, data ProfessionalInvoiceTemplateData) string {
	// Build items rows
	var itemsRows strings.Builder
	for i, item := range items {
		taxAmount := float64(item.Quantity) * item.UnitPrice * (item.TaxPercent / 100)
		total := float64(item.Quantity) * item.UnitPrice + taxAmount

		itemsRows.WriteString(fmt.Sprintf(`
			<tr class="item-row">
				<td class="text-center">%d</td>
				<td class="text-left">
					<div class="item-desc">%s</div>
					<div class="item-hsn">HSN: %s</div>
				</td>
				<td class="text-center">%d</td>
				<td class="text-center">%s</td>
				<td class="text-right">₹%.2f</td>
				<td class="text-right">%.1f%%</td>
				<td class="text-right">₹%.2f</td>
				<td class="text-right">₹%.2f</td>
			</tr>`,
			i+1,
			item.Description,
			getHSNCode(item), // You can add HSN field to model
			item.Quantity,
			"Nos",
			item.UnitPrice,
			item.TaxPercent,
			taxAmount,
			total,
		))
	}

	// Determine tax type (CGST+SGST or IGST)
	isIntraState := true // Check based on state codes
	var taxBreakupRows string
	if isIntraState {
		taxBreakupRows = fmt.Sprintf(`
			<tr><td class="text-right">CGST</td><td class="text-right">₹%.2f</td></tr>
			<tr><td class="text-right">SGST</td><td class="text-right">₹%.2f</td></tr>`,
			data.CGST, data.SGST)
	} else {
		taxBreakupRows = fmt.Sprintf(`
			<tr><td class="text-right">IGST</td><td class="text-right">₹%.2f</td></tr>`,
			data.IGST)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Tax Invoice - %s</title>
	<style>
		@page {
			size: A4;
			margin: 0.5in;
		}
		
		@media print {
			body {
				-webkit-print-color-adjust: exact;
				print-color-adjust: exact;
			}
			.no-print {
				display: none;
			}
		}
		
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			font-size: 10pt;
			line-height: 1.4;
			color: #1a1a1a;
			background: #fff;
		}
		
		.invoice-container {
			max-width: 210mm;
			margin: 0 auto;
			border: 1px solid #e2e8f0;
			box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
		}
		
		/* Header */
		.invoice-header {
			display: flex;
			justify-content: space-between;
			align-items: flex-start;
			padding: 30px;
			background: linear-gradient(135deg, %s 0%%, %s 100%%);
			color: white;
			border-bottom: 4px solid rgba(255, 255, 255, 0.2);
		}
		
		.company-info {
			flex: 1;
		}
		
		.company-name {
			font-size: 28pt;
			font-weight: bold;
			margin-bottom: 10px;
		}
		
		.company-details {
			font-size: 9pt;
			line-height: 1.6;
			opacity: 0.95;
		}
		
		.invoice-title-box {
			text-align: right;
		}
		
		.invoice-title {
			font-size: 32pt;
			font-weight: bold;
			letter-spacing: 2px;
			margin-bottom: 10px;
		}
		
		.invoice-number {
			font-size: 12pt;
			opacity: 0.9;
		}
		
		/* Main Content */
		.invoice-body {
			padding: 30px;
		}
		
		/* Customer & Invoice Info Grid */
		.info-grid {
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 30px;
			margin-bottom: 30px;
		}
		
		.info-box {
			background: #f8fafc;
			border: 1px solid #e2e8f0;
			border-radius: 8px;
			padding: 20px;
		}
		
		.info-box-title {
			font-size: 9pt;
			font-weight: 600;
			color: #64748b;
			text-transform: uppercase;
			margin-bottom: 12px;
			padding-bottom: 8px;
			border-bottom: 2px solid %s;
		}
		
		.customer-name {
			font-size: 14pt;
			font-weight: bold;
			color: #1a1a1a;
			margin-bottom: 8px;
		}
		
		.customer-details {
			font-size: 9pt;
			line-height: 1.6;
			color: #475569;
		}
		
		.info-row {
			display: flex;
			justify-content: space-between;
			padding: 8px 0;
			border-bottom: 1px solid #e2e8f0;
		}
		
		.info-label {
			color: #64748b;
			font-weight: 500;
		}
		
		.info-value {
			color: #1a1a1a;
			font-weight: 600;
		}
		
		/* Items Table */
		.items-table {
			width: 100%%;
			border-collapse: collapse;
			margin-bottom: 25px;
			font-size: 9pt;
		}
		
		.items-table thead {
			background: %s;
			color: white;
		}
		
		.items-table th {
			padding: 12px 8px;
			text-align: center;
			font-weight: 600;
			text-transform: uppercase;
			font-size: 8pt;
			letter-spacing: 0.5px;
		}
		
		.items-table th.text-left {
			text-align: left;
		}
		
		.items-table td {
			padding: 12px 8px;
			border-bottom: 1px solid #e2e8f0;
		}
		
		.items-table td.text-center {
			text-align: center;
		}
		
		.items-table td.text-right {
			text-align: right;
		}
		
		.items-table td.text-left {
			text-align: left;
		}
		
		.item-desc {
			font-weight: 500;
			color: #1a1a1a;
		}
		
		.item-hsn {
			font-size: 8pt;
			color: #64748b;
			margin-top: 4px;
		}
		
		.items-table tbody tr:hover {
			background: #f8fafc;
		}
		
		/* Totals Section */
		.totals-section {
			display: grid;
			grid-template-columns: 1fr 300px;
			gap: 30px;
			margin-top: 25px;
		}
		
		.notes-section {
			background: #f8fafc;
			border: 1px solid #e2e8f0;
			border-radius: 8px;
			padding: 15px;
		}
		
		.notes-title {
			font-weight: 600;
			color: #1a1a1a;
			margin-bottom: 8px;
		}
		
		.notes-content {
			font-size: 9pt;
			color: #475569;
			line-height: 1.6;
		}
		
		.totals-box {
			background: #f8fafc;
			border: 1px solid #e2e8f0;
			border-radius: 8px;
			overflow: hidden;
		}
		
		.total-row {
			display: flex;
			justify-content: space-between;
			padding: 10px 15px;
			border-bottom: 1px solid #e2e8f0;
			font-size: 9pt;
		}
		
		.total-row:last-child {
			border-bottom: none;
		}
		
		.total-row.grand-total {
			background: %s;
			color: white;
			font-size: 12pt;
			font-weight: bold;
			padding: 15px;
		}
		
		/* Payment Info */
		.payment-info {
			margin-top: 25px;
			background: #f0f9ff;
			border: 1px solid #bae6fd;
			border-radius: 8px;
			padding: 20px;
		}
		
		.payment-info-title {
			font-weight: 600;
			color: #0369a1;
			margin-bottom: 12px;
			font-size: 10pt;
		}
		
		.payment-details {
			display: grid;
			grid-template-columns: repeat(3, 1fr);
			gap: 15px;
			font-size: 9pt;
		}
		
		.payment-detail-item {
			display: flex;
			flex-direction: column;
		}
		
		.payment-label {
			color: #0369a1;
			font-weight: 500;
			margin-bottom: 4px;
		}
		
		.payment-value {
			color: #1a1a1a;
			font-family: monospace;
		}
		
		/* Footer */
		.invoice-footer {
			background: #f8fafc;
			border-top: 2px solid #e2e8f0;
			padding: 20px 30px;
			text-align: center;
			font-size: 9pt;
			color: #64748b;
		}
		
		.footer-text {
			margin-bottom: 8px;
		}
		
		/* Status Badge */
		.status-badge {
			display: inline-block;
			padding: 6px 16px;
			border-radius: 20px;
			font-size: 9pt;
			font-weight: 700;
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
		
		.status-paid {
			background: #10b981;
			color: white;
		}
		
		.status-pending {
			background: #f59e0b;
			color: white;
		}
		
		.status-overdue {
			background: #ef4444;
			color: white;
		}
		
		.status-partial {
			background: #3b82f6;
			color: white;
		}
		
		/* QR Code */
		.qr-code {
			text-align: center;
			padding: 15px;
			background: white;
			border: 2px solid #e2e8f0;
			border-radius: 8px;
			margin-top: 15px;
		}
		
		.qr-code img {
			width: 120px;
			height: 120px;
		}
		
		.qr-code-text {
			font-size: 8pt;
			color: #64748b;
			margin-top: 8px;
		}
		
		/* Utility Classes */
		.text-center { text-align: center; }
		.text-right { text-align: right; }
		.text-left { text-align: left; }
		.font-bold { font-weight: bold; }
		.mt-4 { margin-top: 1rem; }
	</style>
</head>
<body>
	<div class="invoice-container">
		<!-- Header -->
		<div class="invoice-header">
			<div class="company-info">
				<div class="company-name">%s</div>
				<div class="company-details">
					%s<br>
					GSTIN: %s | PAN: %s<br>
					Email: %s | Phone: %s<br>
					%s
				</div>
			</div>
			<div class="invoice-title-box">
				<div class="invoice-title">TAX INVOICE</div>
				<div class="invoice-number">%s</div>
				<div style="margin-top: 15px;">
					<span class="status-badge status-%s">%s</span>
				</div>
			</div>
		</div>
		
		<!-- Body -->
		<div class="invoice-body">
			<!-- Customer & Invoice Info -->
			<div class="info-grid">
				<div class="info-box">
					<div class="info-box-title">Bill To</div>
					<div class="customer-name">%s</div>
					<div class="customer-details">
						%s<br>
						GSTIN: %s<br>
						PAN: %s<br>
						Email: %s<br>
						Phone: %s
					</div>
				</div>
				
				<div class="info-box">
					<div class="info-box-title">Invoice Details</div>
					<div class="info-row">
						<span class="info-label">Invoice Date</span>
						<span class="info-value">%s</span>
					</div>
					<div class="info-row">
						<span class="info-label">Due Date</span>
						<span class="info-value">%s</span>
					</div>
					<div class="info-row">
						<span class="info-label">Place of Supply</span>
						<span class="info-value">%s</span>
					</div>
					<div class="info-row">
						<span class="info-label">Reverse Charge</span>
						<span class="info-value">%s</span>
					</div>
				</div>
			</div>
			
			<!-- Items Table -->
			<table class="items-table">
				<thead>
					<tr>
						<th class="text-center" style="width: 5%%;">#</th>
						<th class="text-left" style="width: 35%%;">Description</th>
						<th class="text-center" style="width: 8%%;">Qty</th>
						<th class="text-center" style="width: 8%%;">Unit</th>
						<th class="text-right" style="width: 11%%;">Rate</th>
						<th class="text-right" style="width: 10%%;">GST%%</th>
						<th class="text-right" style="width: 11%%;">Tax Amt</th>
						<th class="text-right" style="width: 12%%;">Total</th>
					</tr>
				</thead>
				<tbody>
					%s
				</tbody>
			</table>
			
			<!-- Totals & Notes -->
			<div class="totals-section">
				<div class="notes-section">
					<div class="notes-title">Notes</div>
					<div class="notes-content">%s</div>
					
					<div class="mt-4">
						<div class="notes-title">Terms & Conditions</div>
						<div class="notes-content">
							<ol style="padding-left: 20px;">
								%s
							</ol>
						</div>
					</div>
					
					<!-- QR Code for UPI Payment -->
					<div class="qr-code">
						<img src="%s" alt="UPI QR Code" />
						<div class="qr-code-text">Scan to Pay via UPI</div>
					</div>
				</div>
				
				<div class="totals-box">
					<div class="total-row">
						<span>Taxable Value</span>
						<span>₹%.2f</span>
					</div>
					%s
					<div class="total-row">
						<span>Total Tax</span>
						<span>₹%.2f</span>
					</div>
					<div class="total-row" style="background: #f1f5f9; font-weight: 600;">
						<span>Grand Total</span>
						<span>₹%.2f</span>
					</div>
					<div class="total-row">
						<span>Amount Paid</span>
						<span>₹%.2f</span>
					</div>
					<div class="total-row grand-total">
						<span>Balance Due</span>
						<span>₹%.2f</span>
					</div>
				</div>
			</div>
			
			<!-- Payment Information -->
			<div class="payment-info">
				<div class="payment-info-title">🏦 Bank Details for Payment</div>
				<div class="payment-details">
					<div class="payment-detail-item">
						<span class="payment-label">Bank Name</span>
						<span class="payment-value">%s</span>
					</div>
					<div class="payment-detail-item">
						<span class="payment-label">Account Name</span>
						<span class="payment-value">%s</span>
					</div>
					<div class="payment-detail-item">
						<span class="payment-label">Account Number</span>
						<span class="payment-value">%s</span>
					</div>
					<div class="payment-detail-item">
						<span class="payment-label">IFSC Code</span>
						<span class="payment-value">%s</span>
					</div>
					<div class="payment-detail-item">
						<span class="payment-label">Branch</span>
						<span class="payment-value">%s</span>
					</div>
					<div class="payment-detail-item">
						<span class="payment-label">UPI ID</span>
						<span class="payment-value">%s</span>
					</div>
				</div>
			</div>
		</div>
		
		<!-- Footer -->
		<div class="invoice-footer">
			<div class="footer-text">
				<strong>%s</strong>
			</div>
			<div class="footer-text">
				%s | Email: %s | Website: %s
			</div>
			<div class="footer-text" style="margin-top: 10px; font-size: 8pt;">
				This is a computer-generated invoice. No signature required.
			</div>
		</div>
	</div>
</body>
</html>`,
		// Template variables
		data.InvoiceNumber,
		data.PrimaryColor,
		darkenColor(data.PrimaryColor, 20),
		data.PrimaryColor,
		data.PrimaryColor,
		data.PrimaryColor,
		data.Company_name,
		data.CompanyAddress,
		data.CompanyGSTIN,
		data.CompanyPAN,
		data.CompanyEmail,
		data.CompanyPhone,
		data.CompanyWebsite,
		data.InvoiceNumber,
		strings.ToLower(string(invoice.Status)),
		string(invoice.Status),
		data.CustomerName,
		data.CustomerAddress,
		data.CustomerGSTIN,
		data.CustomerPAN,
		data.CustomerEmail,
		data.CustomerPhone,
		data.InvoiceDate,
		data.DueDate,
		data.PlaceOfSupply,
		data.ReverseCharge,
		itemsRows.String(),
		data.Notes,
		getTermsHTML(data.Terms),
		data.QRCode,
		data.TaxableValue,
		taxBreakupRows,
		data.TotalTax,
		data.TotalAmount,
		data.AmountPaid,
		data.BalanceDue,
		data.PaymentInfo.BankName,
		data.PaymentInfo.AccountName,
		data.PaymentInfo.AccountNumber,
		data.PaymentInfo.IFSC,
		data.PaymentInfo.Branch,
		data.PaymentInfo.UPI,
		data.Company_name,
		data.CompanyAddress,
		data.CompanyEmail,
		data.CompanyWebsite,
	)

	return html
}

// Helper functions

func getHSNCode(item models.InvoiceItem) string {
	// You can add HSNCode field to InvoiceItem model
	// For now, return default
	return "99831"
}

func getTermsHTML(terms []string) string {
	if len(terms) == 0 {
		terms = []string{
			"Payment due within 30 days from invoice date",
			"Please make cheque/transfer payable to the company",
			"All disputes subject to local jurisdiction",
			"Interest @18%% p.a. on delayed payments",
		}
	}

	var html strings.Builder
	for _, term := range terms {
		html.WriteString(fmt.Sprintf("<li>%s</li>", term))
	}
	return html.String()
}

// darkenColor darkens a hex color by percentage
func darkenColor(hex string, percent int) string {
	// Simple implementation - in production use proper color library
	if hex == "" {
		hex = "#3b82f6"
	}
	return hex // Return as-is for now
}

// DefaultProfessionalInvoiceTemplateData returns default template data
func DefaultProfessionalInvoiceTemplateData() ProfessionalInvoiceTemplateData {
	return ProfessionalInvoiceTemplateData{
		Company_name:  "FlyFleet Technologies Pvt Ltd",
		PrimaryColor:  "#3b82f6",
		Theme:         "modern",
		Terms:         []string{},
		PaymentInfo: PaymentInfo{
			BankName:    "HDFC Bank",
			AccountName: "FlyFleet Technologies Pvt Ltd",
			AccountNumber: "123456789012",
			IFSC:        "HDFC0001234",
			Branch:      "Koramangala, Bangalore",
			UPI:         "flyfleet@hdfcbank",
		},
	}
}
