# Invoice Generator — Implementation Tasks

> **Branch:** `feature/invoice-generation`  
> **Parent PRD:** `docs/product/PRD_Invoice_Generator.md`

---

## Phase 1: Foundation (Models + Repository + Routes)

### Task 1.1: Define Go Models
**File:** `services/fleet-core/internal/domains/billing/models.go`

```go
package billing

type Invoice struct {
    ID              string          `json:"id"`
    OrgID           string          `json:"org_id"`
    CustomerName    string          `json:"customer_name"`
    CustomerEmail   *string         `json:"customer_email,omitempty"`
    BillingAddress  *string         `json:"billing_address,omitempty"`
    InvoiceNumber   string          `json:"invoice_number"`
    InvoiceDate     string          `json:"invoice_date"` // YYYY-MM-DD
    DueDate         string          `json:"due_date"`
    Status          string          `json:"status"` // draft, sent, viewed, paid, overdue, void, refunded
    Currency        string          `json:"currency"` // 3-letter ISO
    Subtotal        float64         `json:"subtotal"`
    TaxAmount       float64         `json:"tax_amount"`
    TotalAmount     float64         `json:"total_amount"`
    AmountPaid      float64         `json:"amount_paid"`
    BalanceDue      float64         `json:"balance_due"`
    AmountRemaining float64         `json:"amount_remaining"`
    Notes           *string         `json:"notes,omitempty"`
    CustomFields    json.RawMessage `json:"custom_fields,omitempty"`
    IsDeleted       bool            `json:"is_deleted"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
    Items           []InvoiceItem   `json:"items,omitempty"`
}

type InvoiceItem struct {
    ID          string  `json:"id"`
    InvoiceID   string  `json:"invoice_id"`
    Description string  `json:"description"`
    Quantity    int     `json:"quantity"`
    UnitPrice   float64 `json:"unit_price"`
    TaxPercent  float64 `json:"tax_percent"`
    Total       float64 `json:"total"`
}

type BillingSettings struct {
    OrgID                 string  `json:"org_id"`
    DefaultTaxRate        float64 `json:"default_tax_rate"`
    BaseCurrency          string  `json:"base_currency"`
    GracePeriodDays       int     `json:"grace_period_days"`
    AutoInvoicingEnabled  bool    `json:"auto_invoicing_enabled"`
}

type BillingMetrics struct {
    TotalInvoicedMTD float64 `json:"total_invoiced_mtd"`
    PendingPayment   float64 `json:"pending_payment"`
    OverdueAmount    float64 `json:"overdue_amount"`
    PaidCount        int     `json:"paid_count"`
    OpenCount        int     `json:"open_count"`
}
```

### Task 1.2: Invoice Repository
**File:** `services/fleet-core/internal/domains/billing/repository.go`

Methods:
- `ListInvoices(ctx, orgID, page, size) ([]Invoice, int, error)`
- `GetInvoice(ctx, id) (*Invoice, error)`
- `CreateInvoice(ctx, invoice) (*Invoice, error)`
- `UpdateInvoice(ctx, id, updates) (*Invoice, error)`
- `DeleteInvoice(ctx, id) error` — soft delete
- `RecordPayment(ctx, invoiceID, amount) (*Invoice, error)`
- `GetSettings(ctx, orgID) (*BillingSettings, error)`
- `UpdateSettings(ctx, orgID, settings) (*BillingSettings, error)`
- `GetMetrics(ctx, orgID) (*BillingMetrics, error)`

### Task 1.3: Invoice Number Sequencing
**File:** `services/fleet-core/internal/domains/billing/sequence.go`

```
Format: INV-{YEAR}-{SEQ}
Example: INV-2026-001, INV-2026-002

Implementation: Use PostgreSQL sequence
  CREATE SEQUENCE IF NOT EXISTS invoice_seq_2026 START 1;
  SELECT nextval('invoice_seq_2026') → 1, 2, 3...
```

### Task 1.4: Tax Calculator
**File:** `services/fleet-core/internal/domains/billing/tax_calculator.go`

```
subtotal = SUM(item.quantity * item.unit_price)
tax_amount = subtotal * (settings.default_tax_rate / 100)
total_amount = subtotal + tax_amount

For each item:
  item.total = item.quantity * item.unit_price
  item.tax_percent = settings.default_tax_rate
```

### Task 1.5: Route Registration
**File:** `services/fleet-core/internal/platform/billing/routes.go`

```go
func RegisterBillingRoutes(r *chi.Mux, handler *billing.Handler) {
    r.Route("/api/v1/admin", func(r chi.Router) {
        r.Get("/billing/invoices", handler.ListInvoices)
        r.Get("/billing/invoices/{id}", handler.GetInvoice)
        r.Post("/billing/invoices", handler.CreateInvoice)
        r.Put("/billing/invoices/{id}", handler.UpdateInvoice)
        r.Delete("/billing/invoices/{id}", handler.DeleteInvoice)
        r.Post("/billing/invoices/{id}/payments", handler.RecordPayment)
        r.Get("/billing-settings", handler.GetSettings)
        r.Put("/billing-settings", handler.UpdateSettings)
        r.Get("/billing-metrics", handler.GetMetrics)
    })
    r.Post("/api/v1/invoice/generate-pdf", handler.GeneratePDF)
}
```

---

## Phase 2: HTTP Handlers

### Task 2.1-2.8: Handler Implementation
**File:** `services/fleet-core/internal/domains/billing/handler.go`

Each handler:
1. Parse request (query params, path params, JSON body)
2. Validate input (required fields, formats)
3. Call service/repository
4. Return JSON response with appropriate status code
5. Handle errors (400 bad request, 404 not found, 500 server error)

Key validations:
- `CreateInvoice`: customer_name required, dates valid, currency 3-letter
- `UpdateInvoice`: only allow updating mutable fields (not invoice_number)
- `RecordPayment`: amount > 0, amount <= balance_due
- `ListInvoices`: default page=1, size=10, max size=100

---

## Phase 3: PDF Generation

### Task 3.1-3.6: chromedp PDF Engine
**File:** `services/fleet-core/internal/domains/billing/pdf_generator.go`

```go
func GeneratePDF(ctx context.Context, invoiceID string, db *pgxpool.Pool) ([]byte, error) {
    // 1. Fetch invoice + items
    // 2. Fetch HTML template from billing_invoice_templates
    // 3. Render Go html/template with invoice data
    // 4. Use chromedp to convert HTML → PDF
    // 5. Save PDF to billing_invoice_documents
    // 6. Return PDF bytes
}
```

**Dependencies to add:**
```bash
cd services/fleet-core
go get github.com/chromedp/chromedp
```

**Fallback option:** If chromedp is too heavy, use `github.com/jung-kurt/gofpdf` for basic PDF generation.

---

## Phase 4: Frontend Integration

### Task 4.1: Fix API Base Path
**File:** `apps/dashboard-web/src/services/billingService.ts`

Current: `/api/v1/admin` (correct)
Check: No double `/api/v1/api/v1` prefix bug

### Task 4.2-4.5: Wire Up Components
- `InvoiceForm.tsx` → `POST /api/v1/admin/billing/invoices`
- `InvoiceDetail.tsx` → PDF download button
- `BillingDashboard.tsx` → Real metrics
- `InvoicesList.tsx` → Portal invoice list

---

## Phase 5: Testing

### Test Checklist
- [ ] Unit: repository.ListInvoices returns paginated results
- [ ] Unit: repository.CreateInvoice generates correct invoice_number
- [ ] Unit: tax_calculator.CalculateTax applies correct rate
- [ ] Unit: sequence.NextSequence returns sequential numbers
- [ ] Integration: POST /api/v1/admin/billing/invoices creates record in DB
- [ ] Integration: GET /api/v1/admin/billing/invoices returns 200
- [ ] Integration: DELETE soft-deletes (is_deleted = true)
- [ ] E2E: Create invoice in UI → appears in list → download PDF
- [ ] Load: 100 concurrent PDF generations complete in < 10s

---

## Quick Start Commands

```bash
# From flyfleet root
cd services/fleet-core

# Add chromedp dependency
go get github.com/chromedp/chromedp

# Create billing domain files
mkdir -p internal/domains/billing
touch internal/domains/billing/{models,repository,handler,pdf_generator,tax_calculator,sequence}.go

# Create routes file
mkdir -p internal/platform/billing
touch internal/platform/billing/routes.go

# Run tests
go test ./internal/domains/billing/... -v

# Start server
export PORT=8084
export DATABASE_URL="postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
go run cmd/fleet-core/main.go
```
