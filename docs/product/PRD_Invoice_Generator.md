# PRD: Invoice Generator — FlyFleet

> **Status:** Draft  
> **Branch:** `feature/invoice-generation`  
> **Created:** 2026-04-03  
> **Author:** PM Toolkit  

---

## 1. Problem Statement

FlyFleet has a **complete billing database schema** (15 migrations), a **polished React frontend** (5 pages), and a **sophisticated Svelte template designer** — but **zero backend implementation**. The frontend calls API endpoints that return 404/500. Users cannot:

- Generate or view invoices in the dashboard
- Download PDF invoices
- Create invoices via the UI
- Record payments
- Auto-generate monthly invoices

**Current state:** Frontend exists, database exists, **the bridge between them does not**.

---

## 2. Goals & Success Metrics

### Goals
| # | Goal | Metric |
|---|------|--------|
| G1 | Generate PDF invoices from HTML templates | PDF generated < 2s, < 5MB |
| G2 | CRUD invoices via REST API | All 9 billingService methods return 200 |
| G3 | Create invoices from dashboard UI | Form submission → DB → PDF in < 3s |
| G4 | View/download invoices from portal | Customer sees their invoices with PDF link |
| G5 | Auto-generate monthly invoices (async) | Cron job runs 1st of month, creates invoices |

### Success Metrics
- **Time to first invoice:** < 5 min from dashboard
- **PDF generation latency:** p95 < 2s
- **Invoice accuracy:** 100% (line items, tax, totals match)
- **API coverage:** 9/9 endpoints implemented

---

## 3. Scope

### In Scope (MVP)
| Feature | Priority | Effort |
|---------|----------|--------|
| HTTP handlers for billing endpoints | P0 | M |
| Invoice repository (DB access layer) | P0 | M |
| PDF generation engine (chromedp) | P0 | L |
| Create/Edit invoice from dashboard | P0 | M |
| Invoice detail view with PDF download | P0 | S |
| Customer portal invoice list | P0 | S |
| Invoice numbering/sequencing | P0 | S |
| Tax calculation (flat rate from settings) | P0 | S |

### Out of Scope (MVP)
| Feature | Deferred To |
|---------|-------------|
| gRPC BillingService implementation | Phase 2 |
| Payment gateway integration (Stripe/Razorpay) | Phase 2 |
| Email delivery of invoices | Phase 2 |
| Monthly auto-invoicing async worker | Phase 2 |
| Dunning/retry logic | Phase 2 |
| Accounting exports | Phase 2 |
| Credit notes | Phase 2 |
| Usage-based billing | Phase 2 |
| Multi-currency support | Phase 3 |
| Advanced tax engine (region-based) | Phase 3 |

---

## 4. Technical Architecture

### 4.1 Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (React)                      │
│  BillingDashboard → InvoiceForm → InvoiceDetail          │
│  billingService.ts → HTTP calls to /api/v1/admin/billing │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP
                         ▼
┌─────────────────────────────────────────────────────────┐
│              Go Monolith (fleet-core :8084)              │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ HTTP Handlers│→ │  Service     │→ │  Repository   │  │
│  │ (billing/)   │  │  Layer       │  │  (invoices)   │  │
│  └──────────────┘  └──────────────┘  └───────┬───────┘  │
│         │                                     │          │
│         ▼                                     ▼          │
│  ┌──────────────┐                    ┌───────────────┐  │
│  │ PDF Generator│                    │  PostgreSQL   │  │
│  │  (chromedp)  │                    │   fleet_db    │  │
│  └──────────────┘                    └───────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 4.2 File Structure (New Files)

```
services/fleet-core/internal/
├── domains/billing/
│   ├── handler.go          # HTTP handlers (chi router)
│   ├── service.go          # Business logic layer
│   ├── repository.go       # DB access (invoices, items, settings)
│   ├── pdf_generator.go    # chromedp PDF rendering
│   ├── tax_calculator.go   # Flat tax rate from settings
│   ├── sequence.go         # Invoice numbering (INV-2026-001)
│   └── models.go           # Go structs matching DB schema
├── platform/billing/
│   └── routes.go           # Route registration to chi router
```

### 4.3 API Contract

| Method | Endpoint | Handler | Response |
|--------|----------|---------|----------|
| `GET` | `/api/v1/admin/billing/invoices` | `ListInvoices` | `{data: Invoice[], total: int}` |
| `GET` | `/api/v1/admin/billing/invoices/:id` | `GetInvoice` | `Invoice` |
| `POST` | `/api/v1/admin/billing/invoices` | `CreateInvoice` | `Invoice` |
| `PUT` | `/api/v1/admin/billing/invoices/:id` | `UpdateInvoice` | `Invoice` |
| `DELETE` | `/api/v1/admin/billing/invoices/:id` | `DeleteInvoice` | `204` |
| `POST` | `/api/v1/admin/billing/invoices/:id/payments` | `RecordPayment` | `{amount, balance_due}` |
| `GET` | `/api/v1/admin/billing-settings` | `GetSettings` | `BillingSettings` |
| `PUT` | `/api/v1/admin/billing-settings` | `UpdateSettings` | `BillingSettings` |
| `GET` | `/api/v1/admin/billing-metrics` | `GetMetrics` | `BillingMetrics` |
| `POST` | `/api/v1/invoice/generate-pdf` | `GeneratePDF` | `application/pdf` |

---

## 5. Implementation Plan

### Phase 1: Foundation (Week 1)
| Task | File | Description |
|------|------|-------------|
| 1.1 | `models.go` | Define Go structs: Invoice, InvoiceItem, BillingSettings, BillingMetrics |
| 1.2 | `repository.go` | CRUD operations for invoices, invoice_items, billing_settings |
| 1.3 | `sequence.go` | Invoice numbering: `INV-{YEAR}-{SEQ}` with atomic counter |
| 1.4 | `tax_calculator.go` | Calculate tax from billing_settings.default_tax_rate |
| 1.5 | `routes.go` | Register all 9 billing routes on chi router |

### Phase 2: HTTP Handlers (Week 1-2)
| Task | File | Description |
|------|------|-------------|
| 2.1 | `handler.go` | `ListInvoices` — paginated list with org_id filter |
| 2.2 | `handler.go` | `GetInvoice` — single invoice with line items |
| 2.3 | `handler.go` | `CreateInvoice` — validate, sequence, insert, return |
| 2.4 | `handler.go` | `UpdateInvoice` — partial update, audit log |
| 2.5 | `handler.go` | `DeleteInvoice` — soft delete (is_deleted = true) |
| 2.6 | `handler.go` | `RecordPayment` — update amount_paid, recalc balance |
| 2.7 | `handler.go` | `GetSettings` / `UpdateSettings` |
| 2.8 | `handler.go` | `GetMetrics` — aggregate: total_invoiced_mtd, pending, overdue |

### Phase 3: PDF Generation (Week 2)
| Task | File | Description |
|------|------|-------------|
| 3.1 | `pdf_generator.go` | chromedp setup: render HTML template → PDF |
| 3.2 | `pdf_generator.go` | Fetch HTML template from `billing_invoice_templates` table |
| 3.3 | `pdf_generator.go` | Inject invoice data into Go html/template |
| 3.4 | `pdf_generator.go` | Render with chromedp headless Chrome, return PDF bytes |
| 3.5 | `handler.go` | `GeneratePDF` endpoint — returns `application/pdf` |
| 3.6 | Frontend | Wire up "Download PDF" button in InvoiceDetail.tsx |

### Phase 4: Frontend Integration (Week 2-3)
| Task | File | Description |
|------|------|-------------|
| 4.1 | `billingService.ts` | Fix API base path (remove double `/api/v1`) |
| 4.2 | `InvoiceForm.tsx` | Connect to `POST /api/v1/admin/billing/invoices` |
| 4.3 | `InvoiceDetail.tsx` | Connect PDF download to `/api/v1/invoice/generate-pdf` |
| 4.4 | `BillingDashboard.tsx` | Connect to real metrics, remove mock data |
| 4.5 | `InvoicesList.tsx` | Connect portal to `/api/v1/portal/invoices` |

### Phase 5: Testing & Polish (Week 3)
| Task | Description |
|------|-------------|
| 5.1 | Unit tests for repository (invoices CRUD) |
| 5.2 | Unit tests for service layer (tax calc, sequencing) |
| 5.3 | Integration tests for HTTP handlers |
| 5.4 | PDF visual validation (GST-compliant layout) |
| 5.5 | E2E test: Create invoice → View → Download PDF |
| 5.6 | Load test: 100 concurrent PDF generations |

---

## 6. Database Operations

### 6.1 Invoice Creation Flow
```sql
BEGIN;
  -- 1. Get next sequence number
  SELECT nextval('invoice_seq_2026') INTO seq;
  
  -- 2. Insert invoice
  INSERT INTO invoices (invoice_number, org_id, customer_name, ...)
  VALUES (format('INV-2026-%03d', seq), ...);
  
  -- 3. Insert line items
  INSERT INTO invoice_items (invoice_id, description, quantity, unit_price, ...)
  VALUES ...;
  
  -- 4. Calculate totals
  UPDATE invoices SET 
    subtotal = SUM(items.total),
    tax_amount = subtotal * tax_rate / 100,
    total_amount = subtotal + tax_amount
  WHERE id = new_invoice_id;
COMMIT;
```

### 6.2 PDF Generation Flow
```
1. Fetch invoice + items from DB
2. Fetch HTML template from billing_invoice_templates
3. Execute Go html/template with invoice data
4. Launch chromedp headless browser
5. Navigate to data: URI with rendered HTML
6. Emulate print media, A4 size
7. Capture PDF bytes
8. Save to billing_invoice_documents table
9. Return PDF to client
```

---

## 7. Dependencies

| Dependency | Purpose | Status |
|------------|---------|--------|
| `github.com/chromedp/chromedp` | PDF rendering | Not in go.mod |
| `github.com/go-chi/chi/v5` | HTTP router | ✅ Already in go.mod |
| `github.com/jackc/pgx/v5` | PostgreSQL driver | ✅ Already in go.mod |
| `github.com/google/uuid` | UUID generation | ✅ Already in go.mod |
| Headless Chrome binary | chromedp runtime | Must be installed on server |

---

## 8. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| chromedp requires Chrome binary on server | High | Use `chromedp/headless-shell` Docker image, or fallback to `gofpdf` |
| PDF generation is slow (>5s) | Medium | Cache PDF in `billing_invoice_documents`, async generation for bulk |
| Template rendering bugs (Go html/template) | Medium | Validate templates with unit tests before deploy |
| Invoice sequence race conditions | High | Use DB sequence (`CREATE SEQUENCE`) not application counter |
| Frontend/backend status mismatch (PAID vs paid) | Medium | Standardize on lowercase DB, uppercase UI, convert in service layer |

---

## 9. Open Questions

| # | Question | Owner |
|---|----------|-------|
| Q1 | Should we use chromedp or gofpdf for PDF? (chromedp = pixel-perfect HTML→PDF, gofpdf = lighter, no Chrome dependency) | Engineering |
| Q2 | Should invoice PDFs be cached or generated on-demand? | Engineering |
| Q3 | Do we need GST-compliant invoices for India? (CGST/SGST split, HSN codes, place of supply) | Product |
| Q4 | Should we implement the gRPC service or stick with REST for now? | Engineering |
| Q5 | Who gets email notifications when invoices are generated? | Product |

---

## 10. Acceptance Criteria

- [ ] `GET /api/v1/admin/billing/invoices` returns paginated list
- [ ] `POST /api/v1/admin/billing/invoices` creates invoice with auto-generated invoice_number
- [ ] `GET /api/v1/admin/billing/invoices/:id` returns invoice with line items
- [ ] `PUT /api/v1/admin/billing/invoices/:id` updates invoice fields
- [ ] `DELETE /api/v1/admin/billing/invoices/:id` soft-deletes invoice
- [ ] `POST /api/v1/admin/billing/invoices/:id/payments` records payment, updates balance
- [ ] `POST /api/v1/invoice/generate-pdf` returns valid PDF (Content-Type: application/pdf)
- [ ] `GET /api/v1/admin/billing-settings` returns settings
- [ ] `PUT /api/v1/admin/billing-settings` updates settings
- [ ] `GET /api/v1/admin/billing-metrics` returns aggregated metrics
- [ ] InvoiceForm.tsx successfully creates invoices via the API
- [ ] InvoiceDetail.tsx shows invoice with line items and PDF download
- [ ] BillingDashboard.tsx shows real metrics (not mock data)
- [ ] PDF contains: invoice number, customer name, line items, subtotal, tax, total, due date
- [ ] All database constraints are respected (FK, CHECK, NOT NULL)

---

## 11. Future Phases

### Phase 2: Automation
- [ ] Monthly auto-invoicing cron job
- [ ] Email delivery (SendGrid/SES)
- [ ] Payment gateway (Razorpay for India, Stripe for global)
- [ ] gRPC BillingService implementation
- [ ] Dunning management (retry failed payments)

### Phase 3: Advanced
- [ ] Multi-currency with FX rates
- [ ] Region-based tax engine
- [ ] Credit notes and refunds
- [ ] Accounting exports (QuickBooks, Tally)
- [ ] Usage-based billing (TimescaleDB aggregation)
- [ ] Invoice approval workflows
- [ ] e-Invoice IRN integration (India GST)

---

## Appendix A: Existing Database Tables

| Table | Purpose | Row Count |
|-------|---------|-----------|
| `invoices` | Core invoice records | 1 (test) |
| `invoice_items` | Line items | 3 (test) |
| `billing_settings` | Org billing config | 0 |
| `billing_plans` | Subscription plans | 3 |
| `billing_subscriptions` | Active subscriptions | 0 |
| `billing_invoice_templates` | HTML templates | 2 (seeded) |
| `billing_invoice_documents` | Stored PDFs | 0 |
| `billing_tax_rates` | Tax rates by region | 0 |
| `billing_credit_notes` | Credit/refund notes | 0 |
| `billing_customer_balances` | Running balances | 0 |
| `billing_ledger_entries` | Double-entry ledger | 0 |
| `invoice_permissions` | RBAC for invoices | 4 |
| `invoice_approvals` | Approval workflow | 0 |
| `invoice_audit_log` | Audit trail | 0 |
| `invoice_sequence` | Numbering counter | 0 |
