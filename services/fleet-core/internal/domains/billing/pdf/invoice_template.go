package pdf

// InvoiceTemplate is the HTML template for invoices
const InvoiceTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice - {{.InvoiceNumber}}</title>
    <style>
        @page {
            size: A4;
            margin: 0.5in;
        }
        
        @media print {
            body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            font-size: 11pt;
            line-height: 1.5;
            color: #1a1a1a;
            background: #fff;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 3px solid #3b82f6;
        }
        
        .logo {
            font-size: 24pt;
            font-weight: bold;
            color: #3b82f6;
        }
        
        .invoice-details {
            text-align: right;
        }
        
        .invoice-title {
            font-size: 28pt;
            font-weight: bold;
            color: #1a1a1a;
            margin-bottom: 10px;
        }
        
        .invoice-number {
            font-size: 12pt;
            color: #666;
        }
        
        .addresses {
            display: flex;
            justify-content: space-between;
            margin-bottom: 30px;
            gap: 40px;
        }
        
        .address-box {
            flex: 1;
        }
        
        .address-label {
            font-size: 9pt;
            font-weight: 600;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 8px;
        }
        
        .address-content {
            font-size: 10pt;
            line-height: 1.6;
        }
        
        .invoice-meta {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 20px;
            margin-bottom: 30px;
            padding: 15px;
            background: #f8fafc;
            border-radius: 8px;
        }
        
        .meta-item {
            text-align: center;
        }
        
        .meta-label {
            font-size: 8pt;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        
        .meta-value {
            font-size: 11pt;
            font-weight: 600;
            color: #1a1a1a;
        }
        
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 30px;
        }
        
        .items-table th {
            background: #3b82f6;
            color: #fff;
            padding: 12px;
            text-align: left;
            font-size: 9pt;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .items-table td {
            padding: 12px;
            border-bottom: 1px solid #e2e8f0;
            font-size: 10pt;
        }
        
        .items-table tr:last-child td {
            border-bottom: 2px solid #3b82f6;
        }
        
        .text-right {
            text-align: right;
        }
        
        .text-center {
            text-align: center;
        }
        
        .totals {
            margin-left: auto;
            width: 300px;
        }
        
        .total-row {
            display: flex;
            justify-content: space-between;
            padding: 10px 15px;
            border-bottom: 1px solid #e2e8f0;
        }
        
        .total-row.grand-total {
            background: #3b82f6;
            color: #fff;
            font-size: 12pt;
            font-weight: bold;
            border-bottom: none;
            border-radius: 0 0 8px 8px;
        }
        
        .status-badge {
            display: inline-block;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 9pt;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .status-paid {
            background: #10b981;
            color: #fff;
        }
        
        .status-pending {
            background: #f59e0b;
            color: #fff;
        }
        
        .status-overdue {
            background: #ef4444;
            color: #fff;
        }
        
        .status-partial {
            background: #3b82f6;
            color: #fff;
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 2px solid #e2e8f0;
            font-size: 9pt;
            color: #666;
            text-align: center;
        }
        
        .notes {
            margin-top: 30px;
            padding: 15px;
            background: #f8fafc;
            border-radius: 8px;
            font-size: 9pt;
        }
        
        .notes-title {
            font-weight: 600;
            margin-bottom: 8px;
            color: #1a1a1a;
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <div class="header">
            <div class="logo">FlyFleet</div>
            <div class="invoice-details">
                <div class="invoice-title">INVOICE</div>
                <div class="invoice-number">{{.InvoiceNumber}}</div>
            </div>
        </div>
        
        <!-- Addresses -->
        <div class="addresses">
            <div class="address-box">
                <div class="address-label">Bill To</div>
                <div class="address-content">
                    <strong>{{.CustomerName}}</strong><br>
                    {{.CustomerAddress}}
                </div>
            </div>
            <div class="address-box">
                <div class="address-label">From</div>
                <div class="address-content">
                    <strong>FlyFleet Inc.</strong><br>
                    123 Fleet Street<br>
                    San Francisco, CA 94102<br>
                    support@flyfleet.com
                </div>
            </div>
        </div>
        
        <!-- Invoice Meta -->
        <div class="invoice-meta">
            <div class="meta-item">
                <div class="meta-label">Invoice Date</div>
                <div class="meta-value">{{.InvoiceDate}}</div>
            </div>
            <div class="meta-item">
                <div class="meta-label">Due Date</div>
                <div class="meta-value">{{.DueDate}}</div>
            </div>
            <div class="meta-item">
                <div class="meta-label">Status</div>
                <div class="meta-value">
                    <span class="status-badge status-{{.StatusClass}}">{{.Status}}</span>
                </div>
            </div>
        </div>
        
        <!-- Items Table -->
        <table class="items-table">
            <thead>
                <tr>
                    <th style="width: 50%;">Description</th>
                    <th class="text-center">Qty</th>
                    <th class="text-right">Unit Price</th>
                    <th class="text-right">Tax %</th>
                    <th class="text-right">Total</th>
                </tr>
            </thead>
            <tbody>
                {{.ItemsRows}}
            </tbody>
        </table>
        
        <!-- Totals -->
        <div class="totals">
            <div class="total-row">
                <span>Subtotal</span>
                <span>{{.Currency}} {{.Subtotal}}</span>
            </div>
            <div class="total-row">
                <span>Tax</span>
                <span>{{.Currency}} {{.TaxAmount}}</span>
            </div>
            <div class="total-row">
                <span>Paid</span>
                <span>{{.Currency}} {{.AmountPaid}}</span>
            </div>
            <div class="total-row grand-total">
                <span>Balance Due</span>
                <span>{{.Currency}} {{.BalanceDue}}</span>
            </div>
        </div>
        
        <!-- Notes -->
        {{if .Notes}}
        <div class="notes">
            <div class="notes-title">Notes</div>
            <div>{{.Notes}}</div>
        </div>
        {{end}}
        
        <!-- Footer -->
        <div class="footer">
            <p>Thank you for your business!</p>
            <p>For questions about this invoice, please contact support@flyfleet.com</p>
        </div>
    </div>
</body>
</html>`

// InvoiceTemplateData represents the data for invoice template
type InvoiceTemplateData struct {
	InvoiceNumber   string
	CustomerName    string
	CustomerAddress string
	InvoiceDate     string
	DueDate         string
	Status          string
	StatusClass     string
	Currency        string
	Subtotal        string
	TaxAmount       string
	AmountPaid      string
	BalanceDue      string
	Notes           string
	ItemsRows       string
}

// InvoiceService provides PDF generation for invoices
type InvoiceService struct {
	generator *PDFGenerator
}

// NewInvoiceService creates a new invoice PDF service
func NewInvoiceService(generator *PDFGenerator) *InvoiceService {
	return &InvoiceService{
		generator: generator,
	}
}
