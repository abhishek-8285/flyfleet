# PDF Generation with Chrome Headless

## Overview

The billing module uses **Chrome Headless** for PDF generation, providing:
- ✅ Perfect page orientation (portrait/landscape)
- ✅ Full modern CSS support (Flexbox, Grid)
- ✅ Print media queries
- ✅ Custom fonts
- ✅ Consistent rendering across platforms

## Installation

### Prerequisites

Chrome or Chromium must be installed on the server.

#### Ubuntu/Debian
```bash
# Install Chromium
sudo apt-get update
sudo apt-get install -y chromium-browser

# Or Google Chrome
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo dpkg -i google-chrome-stable_current_amd64.deb
```

#### CentOS/RHEL
```bash
sudo yum install -y chromium
```

#### macOS
```bash
brew install --cask google-chrome
```

#### Docker
```dockerfile
FROM golang:1.21

# Install Chromium
RUN apt-get update && apt-get install -y \
    chromium \
    fonts-noto \
    && rm -rf /var/lib/apt/lists/*

ENV CHROME_PATH=/usr/bin/chromium
```

## Configuration

### Environment Variables

```bash
# Path to Chrome/Chromium binary (optional - auto-detected)
export CHROME_PATH=/usr/bin/google-chrome

# PDF generation timeout
export PDF_TIMEOUT=30s
```

### Code Usage

```go
import "github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/pdf"

// Create PDF generator
generator := pdf.NewPDFGenerator("", 30*time.Second)

// Generate from HTML
html := `<!DOCTYPE html><html>...</html>`
pdfData, err := generator.GenerateFromHTML(ctx, html, pdf.DefaultA4Portrait())

// Save to file
err = pdf.SaveToFile(pdfData, "/path/to/invoice.pdf")

// Get base64 encoded PDF
base64PDF := pdf.EncodeToBase64(pdfData)

// Generate invoice PDF
invoiceService := pdf.NewInvoiceService(generator)
pdfData, err := invoiceService.GeneratePDF(ctx, invoice, items)
```

## API Endpoint

### Download Invoice PDF

```
GET /api/v1/billing/invoices/{id}/pdf
```

**Response:**
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename=INV-2024-000001.pdf`

**Example:**
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8084/api/v1/billing/invoices/550e8400-e29b-41d4-a716-446655440000/pdf \
  -o invoice.pdf
```

## PDF Options

### Page Formats

```go
// A4 Portrait (default)
pdf.DefaultA4Portrait()

// A4 Landscape
pdf.DefaultA4Landscape()

// Custom options
options := pdf.PDFOptions{
    Format:       "A4",           // A4, Letter, Legal
    Orientation:  "portrait",     // portrait, landscape
    Scale:        1.0,            // 0.1 - 2.0
    PrintBackground: true,
    MarginTop:    0.4,            // inches
    MarginBottom: 0.4,
    MarginLeft:   0.4,
    MarginRight:  0.4,
    PreferCSSPageSize: true,
}
```

### CSS for Print

```css
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
    
    .page-break {
        page-break-before: always;
    }
}
```

## Invoice Template

The default invoice template includes:
- Company logo and details
- Customer billing address
- Invoice number and dates
- Status badge (PAID, PENDING, OVERDUE, PARTIAL)
- Line items table
- Tax calculation
- Totals section
- Notes section
- Footer

## Troubleshooting

### Chrome Not Found
```
Error: chrome/chromium not found
```
**Solution:** Install Chrome/Chromium or set `CHROME_PATH` environment variable.

### Permission Denied
```
Error: failed to create temp dir: permission denied
```
**Solution:** Ensure the application has write access to `/tmp`.

### Timeout
```
Error: context deadline exceeded
```
**Solution:** Increase timeout or optimize HTML template size.

### Missing Fonts
```
PDF shows boxes instead of characters
```
**Solution:** Install required fonts:
```bash
sudo apt-get install fonts-noto fonts-liberation
```

## Performance Tips

1. **Reuse Generator**: Create one `PDFGenerator` instance and reuse it
2. **Cache PDFs**: Store generated PDFs in S3/blob storage
3. **Async Generation**: Generate PDFs asynchronously for large batches
4. **Optimize HTML**: Minimize CSS and remove unused styles

## Security Considerations

1. **Sandbox Mode**: Chrome runs with `--no-sandbox` in containers
2. **Temp Files**: All temp files are cleaned up automatically
3. **Timeout**: Prevents hanging on complex templates
4. **Input Sanitization**: Sanitize HTML input to prevent XSS

## Alternatives

If Chrome Headless is not suitable:

1. **chromedp** - Go package for Chrome DevTools Protocol
   ```go
   import "github.com/chromedp/chromedp"
   ```

2. **wkhtmltopdf** - Deprecated but lighter weight
   ```bash
   sudo apt-get install wkhtmltopdf
   ```

3. **WeasyPrint** - Python-based, good CSS support
   ```bash
   pip install weasyprint
   ```

4. **PrinceXML** - Commercial, best print quality
   ```
   https://www.princexml.com/
   ```

## Testing

```bash
# Test PDF generation
go test ./internal/domains/billing_v2/pdf/... -v

# Test with sample invoice
go run cmd/test_pdf.go
```

## License

This PDF generation module is part of FlyFleet and follows the same license.
