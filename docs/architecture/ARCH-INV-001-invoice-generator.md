# Architecture Specification: Invoice Generator

**Document ID:** ARCH-INV-001  
**Version:** 1.0.0  
**Status:** Proposed  
**Date:** 2026-04-03  
**Branch:** `feature/invoice-generation`  
**Author:** Senior Architect  
**Reviewers:** TBD  

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Context](#2-system-context)
3. [Current State Analysis](#3-current-state-analysis)
4. [Architecture Decision](#4-architecture-decision)
5. [Target Architecture](#5-target-architecture)
6. [Data Model](#6-data-model)
7. [API Specification](#7-api-specification)
8. [Component Design](#8-component-design)
9. [Implementation Plan](#9-implementation-plan)
10. [Quality Assurance](#10-quality-assurance)
11. [Deployment Strategy](#11-deployment-strategy)
12. [Risk Register](#12-risk-register)
13. [Future Phases](#13-future-phases)
14. [Appendices](#14-appendices)

---

## 1. Executive Summary

### 1.1 Problem Statement

FlyFleet has a **complete billing database schema** (15 migrations, 15+ tables), a **polished React frontend** (5 pages with TypeScript types and service layer), and **proto definitions** — but **zero backend implementation**. The frontend calls REST API endpoints that return 404/500 errors.

### 1.2 Solution Overview

Build a complete invoice generation and management system as a **new domain module** within the existing Go monolith (`services/fleet-core`), following the established layered architecture pattern. The system will expose 11 REST endpoints covering invoice CRUD, payment recording, metrics aggregation, and PDF generation.

### 1.3 Scope

| In Scope | Out of Scope (Deferred) |
|----------|------------------------|
| Invoice CRUD operations | Monthly auto-invoicing cron |
| PDF generation (chromedp) | Email delivery (SendGrid/SES) |
| Payment recording (manual) | Payment gateway (Razorpay/Stripe) |
| Billing settings management | gRPC BillingService |
| Metrics aggregation | Dunning management |
| Template-based PDF rendering | Multi-currency FX |
| Invoice numbering sequencing | Region-based tax engine (CGST/SGST/IGST) |
| Soft delete | Credit notes/refunds |
| | Accounting exports (QuickBooks/Tally) |
| | e-Invoice IRN integration |

### 1.4 Key Metrics

- **11 API endpoints** (9 admin, 1 PDF, 1 portal)
- **~15 Go source files** across 4 layers
- **15+ database tables** (existing, no schema changes needed for MVP)
- **Target latency:** <200ms for CRUD, <2s for PDF generation
- **Target coverage:** ≥80% unit test coverage

---

## 2. System Context

### 2.1 C4 Context Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        FlyFleet Platform                            │
│                                                                     │
│  ┌──────────────┐         ┌──────────────────────────────┐          │
│  │   React      │         │     Go Monolith (:8084)      │          │
│  │  Dashboard   │◄───────►│  ┌────────────────────────┐  │          │
│  │              │  REST   │  │   Billing Domain       │  │          │
│  │  - Billing   │  JSON   │  │   (NEW)                │  │          │
│  │    Dashboard │         │  └────────────────────────┘  │          │
│  │  - Invoice   │         │  ┌────────────────────────┐  │          │
│  │    Form      │         │  │   PDF Engine           │  │          │
│  │  - Invoice   │         │  │   (chromedp)           │  │          │
│  │    Detail    │         │  └────────────────────────┘  │          │
│  │  - Settings  │         └──────────┬───────────────────┘          │
│  └──────────────┘                    │                               │
│                                      │                               │
│                           ┌──────────▼───────────────────┐          │
│                           │     PostgreSQL + TimescaleDB │          │
│                           │  - invoices                  │          │
│                           │  - invoice_items             │          │
│                           │  - billing_settings          │          │
│                           │  - billing_invoice_sequences │          │
│                           │  - billing_invoice_templates │          │
│                           │  - billing_invoice_documents │          │
│                           └──────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Existing Infrastructure Inventory

| Component | Status | Location | Notes |
|-----------|--------|----------|-------|
| **Database Schema** | ✅ Complete | `infrastructure/migrations/023-055` | 15 migrations, 15+ tables |
| **Frontend (React)** | ✅ Complete | `apps/dashboard-web/src/` | 5 pages, billingService.ts, types |
| **Proto Definitions** | ⚠️ Minimal | `proto/billing/v1/billing.proto` | 3 RPCs, all unimplemented |
| **Go Monolith** | ⚠️ Skeleton | `services/fleet-core/internal/domains/billing/` | Billing domain exists |
| **Common Package** | ✅ Partial | `packages/common/go/pkg/` | Base service, middleware, DB wrapper |
| **HTML Templates** | ⚠️ Partial | `bill_maintenance.html`, `bill_transport.html` | 2 static files, not DB-seeded |
| **PDF Libraries** | ✅ Available | `go.mod` dependencies | chromedp v0.14.2, gofpdf v0.9.0 |
| **Tenant Middleware** | ✅ Complete | `packages/common/go/pkg/middleware/tenant.go` | Org ID extraction |
| **Service Base** | ✅ Complete | `packages/common/go/pkg/service/base.go` | Logger, DB, Viper, Telemetry |

### 2.3 Dependency Analysis

```
Direct Dependencies (billing domain):
├── github.com/go-chi/chi/v5          # HTTP router
├── github.com/go-playground/validator/v10  # DTO validation
├── github.com/jackc/pgx/v5           # PostgreSQL driver
├── github.com/chromedp/chromedp      # PDF generation
├── go.uber.org/zap                   # Structured logging
└── github.com/fleet-platform/shared  # Internal common package

No circular dependencies. Coupling score: 0/100 (excellent).
```

---

## 3. Current State Analysis

### 3.1 Database Schema (Existing)

#### Core Tables (Migration 023)

```sql
-- invoices (core)
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    customer_name TEXT NOT NULL,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    currency TEXT NOT NULL DEFAULT 'USD',
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    amount_paid DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    balance_due DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- invoice_items
CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    tax_percent DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
    total DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- billing_settings
CREATE TABLE billing_settings (
    org_id UUID PRIMARY KEY,
    default_tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
    base_currency TEXT NOT NULL DEFAULT 'USD',
    grace_period_days INTEGER NOT NULL DEFAULT 14,
    auto_invoicing_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Runtime Additions (Migration 036)

```sql
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_number TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_email TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS billing_address TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS subtotal DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS tax_amount DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS amount_remaining DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS notes TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_invoices_org_deleted ON invoices (org_id, is_deleted);
```

#### Tax & Sequencing (Migration 052)

```sql
-- Invoice sequencing table (gapless sequential numbering per org)
CREATE TABLE billing_invoice_sequences (
    org_id UUID PRIMARY KEY REFERENCES organizations(id),
    prefix VARCHAR(20) NOT NULL DEFAULT 'INV-',
    next_value BIGINT NOT NULL DEFAULT 1,
    suffix VARCHAR(20) DEFAULT '',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tax regions and rates (system-wide, no RLS)
CREATE TABLE billing_tax_regions (...);
CREATE TABLE billing_tax_rates (...);
CREATE TABLE billing_invoice_tax_lines (...);
```

#### Schema ERD (MVP Tables Highlighted)

```
┌─────────────────────────┐     ┌─────────────────────────┐
│    **invoices** ★       │     │   invoice_items         │
├─────────────────────────┤     ├─────────────────────────┤
│ id (UUID, PK)           │◄────│ invoice_id (UUID, FK)   │
│ org_id (UUID)           │     │ id (UUID, PK)           │
│ customer_name (TEXT)    │     │ description (TEXT)      │
│ customer_email (TEXT)   │     │ quantity (INT)          │
│ billing_address (TEXT)  │     │ unit_price (DECIMAL)    │
│ invoice_number (TEXT)   │     │ tax_percent (DECIMAL)   │
│ invoice_date (DATE)     │     │ total (DECIMAL)         │
│ due_date (DATE)         │     │ org_id (UUID)           │
│ status (TEXT)           │     │ is_deleted (BOOL)       │
│ currency (TEXT)         │     └─────────────────────────┘
│ subtotal (DECIMAL)      │
│ tax_amount (DECIMAL)    │     ┌─────────────────────────┐
│ total_amount (DECIMAL)  │     │ **billing_settings** ★  │
│ amount_paid (DECIMAL)   │     ├─────────────────────────┤
│ balance_due (DECIMAL)   │     │ org_id (UUID, PK)       │
│ amount_remaining (DEC)  │     │ default_tax_rate (DEC)  │
│ notes (TEXT)            │     │ base_currency (TEXT)    │
│ is_deleted (BOOL)       │     │ grace_period_days (INT) │
│ deleted_at (TIMESTAMPTZ)│     │ auto_invoicing (BOOL)   │
│ created_at (TIMESTAMPTZ)│     └─────────────────────────┘
│ updated_at (TIMESTAMPTZ)│
└─────────────────────────┘     ┌─────────────────────────┐
                                │ **billing_invoice_      │
┌─────────────────────────┐     │    sequences** ★        │
│ billing_invoice_        │     ├─────────────────────────┤
│   templates             │     │ org_id (UUID, PK)       │
├─────────────────────────┤     │ prefix (VARCHAR)        │
│ id (UUID, PK)           │     │ next_value (BIGINT)     │
│ org_id (UUID)           │     │ suffix (VARCHAR)        │
│ name (TEXT)             │     │ updated_at (TIMESTAMPTZ)│
│ html_content (TEXT)     │     └─────────────────────────┘
│ is_default (BOOL)       │
└─────────────────────────┘     ┌─────────────────────────┐
                                │ billing_invoice_        │
┌─────────────────────────┐     │    documents            │
│ billing_invoice_        │     ├─────────────────────────┤
│   documents             │     │ id (UUID, PK)           │
├─────────────────────────┤     │ invoice_id (UUID, FK)   │
│ id (UUID, PK)           │     │ file_path (TEXT)        │
│ invoice_id (UUID, FK)   │     │ file_size (INT)         │
│ file_path (TEXT)        │     │ mime_type (TEXT)        │
│ file_size (INT)         │     │ generated_at (TIMESTAMPTZ)│
│ generated_at (TIMESTAMPTZ)│   └─────────────────────────┘
└─────────────────────────┘
```

★ = **MVP scope** | Other tables exist but are deferred to Phase 2/3

### 3.2 Frontend Contract

#### TypeScript Types (Existing)

```typescript
// apps/dashboard-web/src/types/billing.ts

export type InvoiceStatus = "PAID" | "PENDING" | "OVERDUE" | "PARTIAL" | "DRAFT" | "SENT" | "VIEWED";

export interface InvoiceItem {
  id: string;
  invoice_id: string;
  description: string;
  quantity: number;
  unit_price: number;
  tax_percent: number;
  total: number;
  created_at?: string;
  updated_at?: string;
}

export interface Invoice {
  id: string;
  org_id: string;
  customer_name: string;
  customer_email?: string;
  billing_address?: string;
  invoice_number?: string;
  invoice_date: string;
  due_date: string;
  status: InvoiceStatus;
  currency: string;
  total_amount: number;
  amount_paid: number;
  balance_due: number;
  subtotal?: number;
  tax_amount?: number;
  notes?: string;
  items?: InvoiceItem[];
  created_at: string;
  updated_at: string;
}

export interface BillingSettings {
  org_id: string;
  default_tax_rate: number;
  base_currency: string;
  grace_period_days: number;
  auto_invoicing_enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface BillingMetrics {
  total_invoiced_mtd: number;
  total_invoiced_trend?: string;
  total_invoiced_trend_up?: boolean;
  pending_payment: number;
  overdue_amount: number;
  paid_count: number;
  open_count: number;
  subtitle_overdue?: string;
  subtitle_paid?: string;
}
```

#### Frontend Service Layer (Existing)

```typescript
// apps/dashboard-web/src/services/billingService.ts

const billingService = {
  getInvoices: async (page = 1, size = 10) => {
    return apiClient.get<any, { data: any[]; total: number; page: number; size: number }>('/billing/invoices', {
      params: { page, size },
    });
  },
  getInvoice: async (id: string) => {
    return apiClient.get<any, Invoice>(`/billing/invoices/${id}`);
  },
  createInvoice: async (invoice: Partial<Invoice>) => {
    return apiClient.post<any, Invoice>('/billing/invoices', invoice);
  },
  updateInvoice: async (id: string, invoice: Partial<Invoice>) => {
    return apiClient.put<any, Invoice>(`/billing/invoices/${id}`, invoice);
  },
  deleteInvoice: async (id: string) => {
    return apiClient.delete(`/billing/invoices/${id}`);
  },
  recordPayment: async (id: string, amount: number) => {
    return apiClient.post(`/billing/invoices/${id}/payments`, { amount });
  },
  downloadPDF: async (id: string): Promise<Blob> => {
    const response = await apiClient.get(`/billing/invoices/${id}/pdf`, {
      responseType: 'blob',
    });
    return response as unknown as Blob;
  },
  getSettings: async () => {
    return apiClient.get<any, BillingSettings>('/billing/settings');
  },
  updateSettings: async (settings: Partial<BillingSettings>) => {
    return apiClient.put('/billing/settings', settings);
  },
  getMetrics: async () => {
    return apiClient.get<any, BillingMetrics>('/billing/metrics');
  },
};
```

#### API Client Interceptor Behavior

The `apiClient` (`apps/dashboard-web/src/services/apiClient.ts`) provides:
- **Base URL:** `/api/v1` (env-aware)
- **Request interceptor:** Attaches `Authorization: Bearer {token}`, `X-Org-ID`, trace ID, idempotency `X-Request-ID` on mutations
- **Response interceptor:** Unwraps `.data` automatically — **all service methods return the response body directly, not AxiosResponse**
- **Error handling:** 401 token refresh, 403 redirect, 429 rate limiting, 503 maintenance

**Critical implication:** Backend responses must follow the `{ data: ..., total: ..., page: ..., size: ... }` envelope format for list endpoints, and return the entity directly for single-resource endpoints.

### 3.3 Existing Go Patterns

#### Service Base Pattern (packages/common/go/pkg/service/base.go)

```go
type Base struct {
    Name    string
    Logger  logger.Logger
    DB      database.Database
    Config  *viper.Viper
    cleanup func()
}

func NewBase(serviceName string, defaultCfg Config) (*Base, error)
func (b *Base) ServeHealth(port string) *http.Server
func (b *Base) RunWithContext(ctx context.Context, runners ...func(context.Context) error) error
```

#### Tenant Middleware (packages/common/go/pkg/middleware/tenant.go)

```go
func TenantMiddleware(next http.Handler) http.Handler          // Permissive
func StrictTenantMiddleware(next http.Handler) http.Handler    // Strict (403 without org_id)
func GetOrgID(ctx context.Context) string
func MustGetOrgID(ctx context.Context) string
```

Org ID extraction order:
1. `X-Organization-ID` header (set by BFF)
2. `X-Org-ID` header (legacy)
3. JWT context (already in context from AuthMiddleware)

#### Monolith Config (services/fleet-core/internal/config/config.go)

```go
type Config struct {
    Port        string
    DatabaseURL string
    NatsURL     string
    JWTSecret   string
    LogLevel    string
    // Module toggles
    Auth     ServiceConfig
    Alerts   ServiceConfig
    Tracking ServiceConfig
}
```

**Gap:** No `Billing ServiceConfig` entry yet — needs to be added.

### 3.4 Gap Analysis

| Layer | Exists | Gap | Action |
|-------|--------|-----|--------|
| Database schema | ✅ | — | No changes needed |
| Frontend types | ✅ | — | No changes needed |
| Frontend service | ✅ | — | No changes needed |
| API client | ✅ | — | No changes needed |
| Go monolith entry point | ❌ | No main.go | Create |
| Go monolith config | ⚠️ | No billing module toggle | Add to Config struct |
| Common DB wrapper | ✅ | — | Use as-is |
| Common middleware | ✅ | — | Use as-is |
| Billing domain (Go) | ❌ | Zero code | Create all layers |
| PDF engine (Go) | ❌ | chromedp unused | Create platform/pdf/ |
| Route registration | ❌ | No routes.go | Create platform/billing/ |
| HTML templates | ⚠️ | 2 static files | Seed into DB or use as-is |

---

## 4. Architecture Decision

### 4.1 Decision Record

| Decision | Choice | Rationale | Alternatives Considered |
|----------|--------|-----------|------------------------|
| **Architecture style** | Layered monolith | Matches existing fleet-core pattern, team of 1-2 devs, simple deployment | Microservices (overkill), Clean Architecture (too abstract for MVP) |
| **Service location** | `services/fleet-core` | Only active Go service; archived services are git-ignored | New billing-service (duplicate infra, more ops burden) |
| **PDF engine** | chromedp (primary) + gofpdf (fallback) | chromedp = pixel-perfect HTML→PDF from existing templates; gofpdf already in go.mod | wkhtmltopdf (deprecated), Puppeteer (Node.js dependency) |
| **Tax calculation** | Flat rate from billing_settings | 95% of FlyFleet customers are India-based with single GST rate | Region-based tax engine (deferred to Phase 3) |
| **Invoice numbering** | Table-based sequence (billing_invoice_sequences) | Atomic UPDATE + RETURNING, no race conditions, per-org isolation | PostgreSQL SEQUENCE (can't be org-scoped easily) |
| **Payment recording** | Manual entry only (MVP) | Payment gateway integration is a separate epic | Razorpay/Stripe integration (deferred to Phase 2) |
| **API protocol** | REST-only | Frontend already uses REST; apiClient interceptors configured | gRPC (adds complexity, no frontend benefit) |
| **PDF generation** | Synchronous | PDF generation is fast (<2s); async queue adds complexity | Async worker with NATS (deferred to Phase 2) |
| **PDF caching** | Cache on first generation | Store in billing_invoice_documents, serve cached copy on subsequent downloads | Always regenerate (wasteful), CDN cache (overkill) |
| **Soft delete** | is_deleted + deleted_at | Existing pattern in migration 036; all queries filter is_deleted=false | Hard delete (loses audit trail), archive table (complex) |

### 4.2 Principles

1. **Follow existing conventions** — Match the trip/domain pattern for file structure, naming, error handling
2. **Keep files focused** — Target <300 lines per file; split if larger
3. **Multi-tenant by default** — Every query scoped by org_id; RLS enforced at DB level
4. **Defensive programming** — Validate all DTOs, return structured errors, never panic in handlers
5. **Testable layers** — Repository interfaces, service with mockable dependencies, handler integration tests

---

## 5. Target Architecture

### 5.1 Module Structure

```
services/fleet-core/
├── go.mod                          # Add chromedp, gofpdf, validator, chi
├── go.sum
├── cmd/
│   └── main.go                     # Entry point — ADD billing module toggle
├── internal/
│   ├── config/
│   │   └── config.go               # Existing — ADD billing module toggle
│   ├── domains/
│   │   └── billing/                # NEW DOMAIN MODULE
│   │       ├── models/
│   │       │   ├── invoice.go      # Domain structs + DTOs
│   │       │   ├── settings.go     # Settings + Metrics structs
│   │       │   └── errors.go       # Domain errors
│   │       ├── repository/
│   │       │   ├── invoice_repo.go     # CRUD for invoices + items
│   │       │   ├── settings_repo.go    # Settings CRUD
│   │       │   ├── metrics_repo.go     # Metrics aggregation
│   │       │   ├── sequence_repo.go    # Invoice numbering
│   │       │   └── template_repo.go    # HTML template fetching
│   │       ├── service/
│   │       │   ├── invoice_service.go  # Business logic
│   │       │   ├── tax_service.go      # Tax calculation
│   │       │   └── sequence_service.go # Number generation
│   │       └── handlers/
│   │           ├── invoice_handler.go      # 9 admin endpoints
│   │           ├── pdf_handler.go          # PDF generation
│   │           └── invoice_handler_test.go # Unit tests
│   └── platform/
│       ├── billing/
│       │   └── routes.go               # Route registration to chi router
│       └── pdf/
│           └── engine.go               # PDF generation engine (chromedp)
└── templates/
    ├── invoice_default.html            # Default invoice HTML template
    └── invoice_minimal.html            # Minimal template (fallback)
```

### 5.2 Component Interaction Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Request Flow                                       │
└─────────────────────────────────────────────────────────────────────────────┘

Client (React Dashboard)
    │
    │ POST /api/v1/admin/billing/invoices
    │ Authorization: Bearer <token>
    │ X-Organization-ID: <org-uuid>
    │ Content-Type: application/json
    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  chi Router (platform/billing/routes.go)                                 │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ StrictTenantMiddleware                                             │  │
│  │  → Extract org_id from X-Organization-ID header                    │  │
│  │  → Inject into context                                             │  │
│  │  → 403 if missing                                                  │  │
│  └────────────────────────────────────────────────────────────────────┘  │
    │
    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Handler (domains/billing/handlers/invoice_handler.go)                   │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 1. Decode JSON body into CreateInvoiceRequest DTO                  │  │
│  │ 2. Validate DTO (go-playground/validator)                          │  │
│  │ 3. Extract org_id from context (MustGetOrgID)                      │  │
│  │ 4. Call service.CreateInvoice(ctx, orgID, &dto)                    │  │
│  │ 5. Return 201 with Invoice JSON                                    │  │
│  └────────────────────────────────────────────────────────────────────┘  │
    │
    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Service (domains/billing/service/invoice_service.go)                    │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 1. Get next invoice number from sequence_service                   │  │
│  │ 2. Get default_tax_rate from settings_repo                         │  │
│  │ 3. Calculate tax via tax_service                                   │  │
│  │ 4. Calculate subtotal, tax_amount, total_amount, balance_due       │  │
│  │ 5. BEGIN transaction via repo                                      │  │
│  │ 6. INSERT invoice (with invoice_number, org_id, ...)               │  │
│  │ 7. INSERT invoice_items (for each line item)                       │  │
│  │ 8. COMMIT                                                          │  │
│  │ 9. Return Invoice                                                  │  │
│  └────────────────────────────────────────────────────────────────────┘  │
    │
    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  Repository (domains/billing/repository/invoice_repo.go)                 │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ 1. db.WithTenant(ctx, orgID, func(tx) error { ... })               │  │
│  │    → Sets SET LOCAL app.current_org_id = '<org-uuid>'              │  │
│  │ 2. tx.QueryRow(ctx, "INSERT INTO invoices ... RETURNING id")       │  │
│  │ 3. tx.Exec(ctx, "INSERT INTO invoice_items ...") × N               │  │
│  │ 4. tx.Commit()                                                     │  │
│  └────────────────────────────────────────────────────────────────────┘  │
    │
    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  PostgreSQL (fleet_db)                                                   │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ RLS Policy: USING (org_id = get_current_org_id())                  │  │
│  │ CHECK constraints: balance_due >= 0, total_amount > 0              │  │
│  │ FK: invoice_items.invoice_id → invoices.id ON DELETE CASCADE       │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

### 5.3 PDF Generation Flow

```
Client POST /api/v1/invoice/generate-pdf {invoice_id}
    │
    ▼
Handler (pdf_handler.go)
    │ 1. Decode request, validate invoice_id
    │ 2. Call invoice_service.GetInvoice(ctx, orgID, invoiceID)
    │ 3. Check if PDF already cached in billing_invoice_documents
    │    → If cached: return cached PDF bytes
    │    → If not: proceed to generation
    ▼
PDF Engine (platform/pdf/engine.go)
    │ 1. Fetch HTML template from billing_invoice_templates (or file)
    │ 2. Execute Go html/template with invoice data
    │ 3. chromedp — launch headless Chrome (reuses browser context)
    │ 4. Navigate to data: URI with rendered HTML
    │ 5. Emulate print media, A4 size, capture PDF bytes
    │ 6. Save PDF to billing_invoice_documents table
    │ 7. Close chromedp tab (not browser — reuse)
    ▼
Handler
    │ Return PDF as application/pdf (binary download)
```

### 5.4 Route Registration

```go
// internal/platform/billing/routes.go

func RegisterBillingRoutes(router *chi.Mux, db database.Database, logger *zap.Logger) {
    // Initialize repositories
    invoiceRepo := repository.NewInvoiceRepository(db)
    settingsRepo := repository.NewSettingsRepository(db)
    metricsRepo := repository.NewMetricsRepository(db)
    sequenceRepo := repository.NewSequenceRepository(db)
    templateRepo := repository.NewTemplateRepository(db)

    // Initialize services
    taxService := service.NewTaxService(settingsRepo)
    sequenceService := service.NewSequenceService(sequenceRepo)
    invoiceService := service.NewInvoiceService(invoiceRepo, taxService, sequenceService, settingsRepo, metricsRepo)

    // Initialize handlers
    invoiceHandler := handlers.NewInvoiceHandler(invoiceService, logger)
    pdfHandler := handlers.NewPDFHandler(invoiceService, pdfEngine, logger)

    // Admin routes (require strict tenant isolation)
    router.Route("/api/v1/admin", func(r chi.Router) {
        r.Use(middleware.StrictTenantMiddleware)

        r.Get("/billing/invoices", invoiceHandler.ListInvoices)
        r.Get("/billing/invoices/{id}", invoiceHandler.GetInvoice)
        r.Post("/billing/invoices", invoiceHandler.CreateInvoice)
        r.Put("/billing/invoices/{id}", invoiceHandler.UpdateInvoice)
        r.Delete("/billing/invoices/{id}", invoiceHandler.DeleteInvoice)
        r.Post("/billing/invoices/{id}/payments", invoiceHandler.RecordPayment)
        r.Get("/billing/settings", invoiceHandler.GetSettings)
        r.Put("/billing/settings", invoiceHandler.UpdateSettings)
        r.Get("/billing/metrics", invoiceHandler.GetMetrics)
    })

    // PDF route (shared endpoint)
    router.Post("/api/v1/invoice/generate-pdf", pdfHandler.GeneratePDF)

    // Portal route (customer-facing, permissive tenant)
    router.Route("/api/v1/portal", func(r chi.Router) {
        r.Use(middleware.TenantMiddleware)
        r.Get("/invoices", invoiceHandler.ListCustomerInvoices)
    })
}
```

---

## 6. Data Model

### 6.1 Go Domain Models

#### models/invoice.go

```go
package models

import "time"

// Invoice represents a billing invoice in the system.
type Invoice struct {
    ID             string     `json:"id" db:"id"`
    OrgID          string     `json:"org_id" db:"org_id"`
    CustomerName   string     `json:"customer_name" db:"customer_name"`
    CustomerEmail  *string    `json:"customer_email,omitempty" db:"customer_email"`
    BillingAddress *string    `json:"billing_address,omitempty" db:"billing_address"`
    InvoiceNumber  string     `json:"invoice_number" db:"invoice_number"`
    InvoiceDate    time.Time  `json:"invoice_date" db:"invoice_date"`
    DueDate        time.Time  `json:"due_date" db:"due_date"`
    Status         string     `json:"status" db:"status"` // lowercase in DB
    Currency       string     `json:"currency" db:"currency"`
    Subtotal       float64    `json:"subtotal" db:"subtotal"`
    TaxAmount      float64    `json:"tax_amount" db:"tax_amount"`
    TotalAmount    float64    `json:"total_amount" db:"total_amount"`
    AmountPaid     float64    `json:"amount_paid" db:"amount_paid"`
    BalanceDue     float64    `json:"balance_due" db:"balance_due"`
    AmountRemaining float64   `json:"amount_remaining" db:"amount_remaining"`
    Notes          *string    `json:"notes,omitempty" db:"notes"`
    IsDeleted      bool       `json:"-" db:"is_deleted"`
    DeletedAt      *time.Time `json:"-" db:"deleted_at"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
    Items          []InvoiceItem `json:"items,omitempty"`
}

// InvoiceStatus constants — lowercase in DB, uppercase in API responses
const (
    StatusDraft    = "draft"
    StatusSent     = "sent"
    StatusViewed   = "viewed"
    StatusPaid     = "paid"
    StatusPending  = "pending"
    StatusOverdue  = "overdue"
    StatusPartial  = "partial"
)

// InvoiceItem represents a line item on an invoice.
type InvoiceItem struct {
    ID          string    `json:"id" db:"id"`
    InvoiceID   string    `json:"invoice_id" db:"invoice_id"`
    OrgID       string    `json:"org_id" db:"org_id"`
    Description string    `json:"description" db:"description"`
    Quantity    int       `json:"quantity" db:"quantity"`
    UnitPrice   float64   `json:"unit_price" db:"unit_price"`
    TaxPercent  float64   `json:"tax_percent" db:"tax_percent"`
    Total       float64   `json:"total" db:"total"`
    IsDeleted   bool      `json:"-" db:"is_deleted"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

#### DTOs (Create/Update Requests)

```go
// CreateInvoiceRequest is the DTO for POST /billing/invoices
type CreateInvoiceRequest struct {
    CustomerName   string              `json:"customer_name" validate:"required,min=2,max=200"`
    CustomerEmail  *string             `json:"customer_email,omitempty" validate:"omitempty,email"`
    BillingAddress *string             `json:"billing_address,omitempty" validate:"omitempty,max=500"`
    InvoiceDate    string              `json:"invoice_date" validate:"required,datetime=2006-01-02"`
    DueDate        string              `json:"due_date" validate:"required,datetime=2006-01-02"`
    Currency       string              `json:"currency" validate:"required,iso4217"`
    Notes          *string             `json:"notes,omitempty" validate:"omitempty,max=1000"`
    Items          []CreateInvoiceItem `json:"items" validate:"required,min=1,dive"`
}

type CreateInvoiceItem struct {
    Description string  `json:"description" validate:"required,min=2,max=500"`
    Quantity    int     `json:"quantity" validate:"required,min=1"`
    UnitPrice   float64 `json:"unit_price" validate:"required,min=0"`
    TaxPercent  float64 `json:"tax_percent" validate:"gte=0,lte=100"`
}

// UpdateInvoiceRequest is the DTO for PUT /billing/invoices/:id
type UpdateInvoiceRequest struct {
    CustomerName   *string              `json:"customer_name,omitempty" validate:"omitempty,min=2,max=200"`
    CustomerEmail  *string              `json:"customer_email,omitempty" validate:"omitempty,email"`
    BillingAddress *string              `json:"billing_address,omitempty" validate:"omitempty,max=500"`
    DueDate        *string              `json:"due_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
    Notes          *string              `json:"notes,omitempty" validate:"omitempty,max=1000"`
    Items          *[]CreateInvoiceItem `json:"items,omitempty" validate:"omitempty,min=1,dive"`
}

// RecordPaymentRequest is the DTO for POST /billing/invoices/:id/payments
type RecordPaymentRequest struct {
    Amount float64 `json:"amount" validate:"required,gt=0"`
}
```

#### models/settings.go

```go
package models

import "time"

// BillingSettings represents the billing configuration for an organization.
type BillingSettings struct {
    OrgID                string    `json:"org_id" db:"org_id"`
    DefaultTaxRate       float64   `json:"default_tax_rate" db:"default_tax_rate"`
    BaseCurrency         string    `json:"base_currency" db:"base_currency"`
    GracePeriodDays      int       `json:"grace_period_days" db:"grace_period_days"`
    AutoInvoicingEnabled bool      `json:"auto_invoicing_enabled" db:"auto_invoicing_enabled"`
    CreatedAt            time.Time `json:"created_at" db:"created_at"`
    UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateBillingSettingsRequest is the DTO for PUT /billing/settings
type UpdateBillingSettingsRequest struct {
    DefaultTaxRate       *float64 `json:"default_tax_rate,omitempty" validate:"omitempty,gte=0,lte=100"`
    BaseCurrency         *string  `json:"base_currency,omitempty" validate:"omitempty,iso4217"`
    GracePeriodDays      *int     `json:"grace_period_days,omitempty" validate:"omitempty,gte=0"`
    AutoInvoicingEnabled *bool    `json:"auto_invoicing_enabled,omitempty"`
}

// BillingMetrics represents aggregated billing statistics.
type BillingMetrics struct {
    TotalInvoicedMTD    float64 `json:"total_invoiced_mtd" db:"total_invoiced_mtd"`
    TotalInvoicedTrend  *string `json:"total_invoiced_trend,omitempty"`
    TotalInvoicedTrendUp *bool  `json:"total_invoiced_trend_up,omitempty"`
    PendingPayment      float64 `json:"pending_payment" db:"pending_payment"`
    OverdueAmount       float64 `json:"overdue_amount" db:"overdue_amount"`
    PaidCount           int     `json:"paid_count" db:"paid_count"`
    OpenCount           int     `json:"open_count" db:"open_count"`
    SubtitleOverdue     *string `json:"subtitle_overdue,omitempty"`
    SubtitlePaid        *string `json:"subtitle_paid,omitempty"`
}
```

#### models/errors.go

```go
package models

import "fmt"

// Domain error types
var (
    ErrInvoiceNotFound     = &DomainError{Code: "invoice_not_found", Message: "Invoice not found"}
    ErrSettingsNotFound    = &DomainError{Code: "settings_not_found", Message: "Billing settings not found for organization"}
    ErrInvalidInvoiceDate  = &DomainError{Code: "invalid_invoice_date", Message: "Invoice date must be valid"}
    ErrInvalidDueDate      = &DomainError{Code: "invalid_due_date", Message: "Due date must be after invoice date"}
    ErrEmptyInvoiceItems   = &DomainError{Code: "empty_invoice_items", Message: "Invoice must have at least one line item"}
    ErrPaymentExceedsBalance = &DomainError{Code: "payment_exceeds_balance", Message: "Payment amount exceeds balance due"}
    ErrInvoiceNotEditable  = &DomainError{Code: "invoice_not_editable", Message: "Invoice cannot be modified after being sent or paid"}
)

// DomainError represents a business logic error
type DomainError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
```

### 6.2 Database Query Patterns

#### Invoice CRUD (invoice_repo.go)

```sql
-- CreateInvoice (within transaction)
INSERT INTO invoices (
    org_id, customer_name, customer_email, billing_address,
    invoice_number, invoice_date, due_date, status, currency,
    subtotal, tax_amount, total_amount, amount_paid, balance_due,
    amount_remaining, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING id, created_at, updated_at;

-- Insert items (batch)
INSERT INTO invoice_items (
    invoice_id, org_id, description, quantity, unit_price, tax_percent, total
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- GetInvoice (with items)
SELECT i.*, 
       json_agg(
           json_build_object(
               'id', it.id,
               'invoice_id', it.invoice_id,
               'description', it.description,
               'quantity', it.quantity,
               'unit_price', it.unit_price,
               'tax_percent', it.tax_percent,
               'total', it.total
           ) ORDER BY it.created_at
       ) FILTER (WHERE it.id IS NOT NULL) as items
FROM invoices i
LEFT JOIN invoice_items it ON i.id = it.invoice_id AND it.is_deleted = false
WHERE i.id = $1 AND i.org_id = $2 AND i.is_deleted = false
GROUP BY i.id;

-- ListInvoices (paginated, org-scoped, soft-delete filtered)
SELECT id, org_id, customer_name, invoice_number, invoice_date, due_date,
       status, currency, total_amount, amount_paid, balance_due,
       created_at, updated_at
FROM invoices
WHERE org_id = $1 AND is_deleted = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- Count for pagination
SELECT COUNT(*) FROM invoices
WHERE org_id = $1 AND is_deleted = false;

-- Soft delete
UPDATE invoices 
SET is_deleted = true, deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2;

-- Record payment
UPDATE invoices
SET amount_paid = amount_paid + $1,
    balance_due = GREATEST(total_amount - (amount_paid + $1), 0),
    amount_remaining = GREATEST(total_amount - (amount_paid + $1), 0),
    status = CASE 
        WHEN (amount_paid + $1) >= total_amount THEN 'paid'
        WHEN (amount_paid + $1) > 0 THEN 'partial'
        ELSE status
    END,
    updated_at = NOW()
WHERE id = $2 AND org_id = $3 AND is_deleted = false
RETURNING amount_paid, balance_due;
```

#### Metrics Aggregation (metrics_repo.go)

```sql
-- Total invoiced this month
SELECT COALESCE(SUM(total_amount), 0)
FROM invoices
WHERE org_id = $1
  AND invoice_date >= date_trunc('month', CURRENT_DATE)
  AND is_deleted = false;

-- Pending payments (sent or viewed)
SELECT COALESCE(SUM(balance_due), 0)
FROM invoices
WHERE org_id = $1 AND status IN ('sent', 'viewed') AND is_deleted = false;

-- Overdue amount
SELECT COALESCE(SUM(balance_due), 0)
FROM invoices
WHERE org_id = $1 AND status = 'overdue' AND is_deleted = false;

-- Paid count
SELECT COUNT(*) FROM invoices
WHERE org_id = $1 AND status = 'paid' AND is_deleted = false;

-- Open count (sent + viewed)
SELECT COUNT(*) FROM invoices
WHERE org_id = $1 AND status IN ('sent', 'viewed') AND is_deleted = false;
```

#### Invoice Sequencing (sequence_repo.go)

```sql
-- Get next invoice number (atomic, per-org)
UPDATE billing_invoice_sequences
SET next_value = next_value + 1,
    updated_at = NOW()
WHERE org_id = $1
RETURNING prefix, next_value - 1, suffix;

-- Initialize sequence for new org (upsert)
INSERT INTO billing_invoice_sequences (org_id, prefix, next_value, suffix)
VALUES ($1, 'INV-', 1, '')
ON CONFLICT (org_id) DO NOTHING;
```

---

## 7. API Specification

### 7.1 Endpoint Summary

| # | Method | Endpoint | Handler | Auth | Status |
|---|--------|----------|---------|------|--------|
| 1 | `GET` | `/api/v1/admin/billing/invoices` | `ListInvoices` | Strict Tenant | 200 |
| 2 | `GET` | `/api/v1/admin/billing/invoices/:id` | `GetInvoice` | Strict Tenant | 200 |
| 3 | `POST` | `/api/v1/admin/billing/invoices` | `CreateInvoice` | Strict Tenant | 201 |
| 4 | `PUT` | `/api/v1/admin/billing/invoices/:id` | `UpdateInvoice` | Strict Tenant | 200 |
| 5 | `DELETE` | `/api/v1/admin/billing/invoices/:id` | `DeleteInvoice` | Strict Tenant | 204 |
| 6 | `POST` | `/api/v1/admin/billing/invoices/:id/payments` | `RecordPayment` | Strict Tenant | 200 |
| 7 | `GET` | `/api/v1/admin/billing/settings` | `GetSettings` | Strict Tenant | 200 |
| 8 | `PUT` | `/api/v1/admin/billing/settings` | `UpdateSettings` | Strict Tenant | 200 |
| 9 | `GET` | `/api/v1/admin/billing/metrics` | `GetMetrics` | Strict Tenant | 200 |
| 10 | `POST` | `/api/v1/invoice/generate-pdf` | `GeneratePDF` | Auth (permissive) | 200 |
| 11 | `GET` | `/api/v1/portal/invoices` | `ListCustomerInvoices` | Permissive Tenant | 200 |

### 7.2 Detailed Endpoint Specifications

#### 7.2.1 List Invoices

```
GET /api/v1/admin/billing/invoices?page=1&size=10
```

**Headers:**
```
Authorization: Bearer <token>
X-Organization-ID: <org-uuid>
```

**Query Parameters:**
| Param | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| page | int | No | 1 | Page number (1-indexed) |
| size | int | No | 10 | Items per page (max 100) |

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "org_id": "org-uuid",
      "customer_name": "Acme Corp",
      "invoice_number": "INV-2026-001",
      "invoice_date": "2026-04-01",
      "due_date": "2026-04-15",
      "status": "SENT",
      "currency": "INR",
      "total_amount": 11800.00,
      "amount_paid": 0.00,
      "balance_due": 11800.00,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "size": 10
}
```

**Status Mapping:** DB stores lowercase (`sent`), API returns uppercase (`SENT`) to match frontend `InvoiceStatus` type.

#### 7.2.2 Get Invoice

```
GET /api/v1/admin/billing/invoices/:id
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "org_id": "org-uuid",
  "customer_name": "Acme Corp",
  "customer_email": "billing@acme.com",
  "billing_address": "123 Business St, Mumbai",
  "invoice_number": "INV-2026-001",
  "invoice_date": "2026-04-01",
  "due_date": "2026-04-15",
  "status": "SENT",
  "currency": "INR",
  "subtotal": 10000.00,
  "tax_amount": 1800.00,
  "total_amount": 11800.00,
  "amount_paid": 0.00,
  "balance_due": 11800.00,
  "notes": "Payment due within 14 days",
  "items": [
    {
      "id": "item-uuid",
      "invoice_id": "550e8400-e29b-41d4-a716-446655440000",
      "description": "Fleet Management - Monthly Subscription",
      "quantity": 1,
      "unit_price": 10000.00,
      "tax_percent": 18.00,
      "total": 10000.00
    }
  ],
  "created_at": "2026-04-01T10:00:00Z",
  "updated_at": "2026-04-01T10:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` — Invoice not found or belongs to different org
- `403 Forbidden` — Missing org_id in context

#### 7.2.3 Create Invoice

```
POST /api/v1/admin/billing/invoices
```

**Request Body:**
```json
{
  "customer_name": "Acme Corp",
  "customer_email": "billing@acme.com",
  "billing_address": "123 Business St, Mumbai",
  "invoice_date": "2026-04-01",
  "due_date": "2026-04-15",
  "currency": "INR",
  "notes": "Payment due within 14 days",
  "items": [
    {
      "description": "Fleet Management - Monthly Subscription",
      "quantity": 1,
      "unit_price": 10000.00,
      "tax_percent": 18.00
    }
  ]
}
```

**Validation Rules:**
| Field | Rule |
|-------|------|
| customer_name | Required, 2-200 chars |
| customer_email | Optional, valid email format |
| billing_address | Optional, max 500 chars |
| invoice_date | Required, ISO 8601 date (YYYY-MM-DD) |
| due_date | Required, must be ≥ invoice_date |
| currency | Required, valid ISO 4217 code |
| notes | Optional, max 1000 chars |
| items | Required, min 1 item |
| items[].description | Required, 2-500 chars |
| items[].quantity | Required, ≥ 1 |
| items[].unit_price | Required, ≥ 0 |
| items[].tax_percent | Required, 0-100 |

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "invoice_number": "INV-2026-001",
  "customer_name": "Acme Corp",
  "invoice_date": "2026-04-01",
  "due_date": "2026-04-15",
  "status": "DRAFT",
  "currency": "INR",
  "subtotal": 10000.00,
  "tax_amount": 1800.00,
  "total_amount": 11800.00,
  "amount_paid": 0.00,
  "balance_due": 11800.00,
  "items": [...],
  "created_at": "2026-04-01T10:00:00Z",
  "updated_at": "2026-04-01T10:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request` — Validation failure
- `403 Forbidden` — Missing org_id
- `409 Conflict` — Invoice number collision (extremely rare)

#### 7.2.4 Update Invoice

```
PUT /api/v1/admin/billing/invoices/:id
```

**Request Body:** (all fields optional)
```json
{
  "customer_name": "Acme Corp Updated",
  "due_date": "2026-04-30",
  "notes": "Extended payment terms"
}
```

**Business Rules:**
- Cannot update invoice if status is `sent`, `viewed`, or `paid`
- Cannot update items if invoice has been sent
- Partial update — only provided fields are modified

**Response (200 OK):** Updated Invoice object

**Error Responses:**
- `404 Not Found` — Invoice not found
- `409 Conflict` — Invoice not editable
- `400 Bad Request` — Validation failure

#### 7.2.5 Delete Invoice

```
DELETE /api/v1/admin/billing/invoices/:id
```

**Behavior:** Soft delete — sets `is_deleted = true`, `deleted_at = NOW()`

**Response:** `204 No Content`

**Error Responses:**
- `404 Not Found` — Invoice not found
- `409 Conflict` — Invoice has payments recorded

#### 7.2.6 Record Payment

```
POST /api/v1/admin/billing/invoices/:id/payments
```

**Request Body:**
```json
{
  "amount": 5900.00
}
```

**Validation:**
- `amount` > 0
- `amount` ≤ `balance_due`

**Business Logic:**
```
new_amount_paid = amount_paid + payment_amount
new_balance_due = MAX(total_amount - new_amount_paid, 0)

IF new_amount_paid >= total_amount:
    status = "paid"
ELSE IF new_amount_paid > 0:
    status = "partial"
ELSE:
    status = unchanged
```

**Response (200 OK):**
```json
{
  "amount_paid": 5900.00,
  "balance_due": 5900.00
}
```

#### 7.2.7 Get Settings

```
GET /api/v1/admin/billing/settings
```

**Response (200 OK):**
```json
{
  "org_id": "org-uuid",
  "default_tax_rate": 18.00,
  "base_currency": "INR",
  "grace_period_days": 14,
  "auto_invoicing_enabled": false,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**Note:** If no settings exist for org, auto-create with defaults.

#### 7.2.8 Update Settings

```
PUT /api/v1/admin/billing/settings
```

**Request Body:** (all fields optional)
```json
{
  "default_tax_rate": 18.00,
  "base_currency": "INR",
  "grace_period_days": 30,
  "auto_invoicing_enabled": true
}
```

**Response (200 OK):** Updated BillingSettings

#### 7.2.9 Get Metrics

```
GET /api/v1/admin/billing/metrics
```

**Response (200 OK):**
```json
{
  "total_invoiced_mtd": 150000.00,
  "total_invoiced_trend": "+12.5%",
  "total_invoiced_trend_up": true,
  "pending_payment": 45000.00,
  "overdue_amount": 12000.00,
  "paid_count": 28,
  "open_count": 14,
  "subtitle_overdue": "3 invoices",
  "subtitle_paid": "This month"
}
```

#### 7.2.10 Generate PDF

```
POST /api/v1/invoice/generate-pdf
```

**Request Body:**
```json
{
  "invoice_id": "550e8400-e29b-41d4-a716-446655440000",
  "template_id": "default"
}
```

**Response (200 OK):**
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="INV-2026-001.pdf"
Content-Length: 45678

<binary PDF data>
```

**Performance:** <2s for first generation, <200ms for cached PDFs

#### 7.2.11 List Customer Invoices (Portal)

```
GET /api/v1/portal/invoices?page=1&size=10
```

**Auth:** Permissive tenant middleware (org_id from header or JWT)

**Response:** Same envelope format as admin ListInvoices, but filtered to customer's org only.

---

## 8. Component Design

### 8.1 Repository Layer

#### Interface Design

```go
// repository/interfaces.go (optional — for testability)

type InvoiceRepository interface {
    Create(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) (*models.Invoice, error)
    GetByID(ctx context.Context, id, orgID string) (*models.Invoice, error)
    List(ctx context.Context, orgID string, page, size int) ([]models.Invoice, int, error)
    Update(ctx context.Context, invoice *models.Invoice) (*models.Invoice, error)
    SoftDelete(ctx context.Context, id, orgID string) error
    RecordPayment(ctx context.Context, id, orgID string, amount float64) (float64, float64, error)
}

type SettingsRepository interface {
    GetOrCreate(ctx context.Context, orgID string) (*models.BillingSettings, error)
    Update(ctx context.Context, settings *models.BillingSettings) (*models.BillingSettings, error)
    GetDefaultTaxRate(ctx context.Context, orgID string) (float64, error)
}

type MetricsRepository interface {
    GetMetrics(ctx context.Context, orgID string) (*models.BillingMetrics, error)
}

type SequenceRepository interface {
    NextInvoiceNumber(ctx context.Context, orgID string) (string, error)
}

type TemplateRepository interface {
    GetDefault(ctx context.Context) (string, error) // Returns HTML content
    GetByID(ctx context.Context, id string) (string, error)
}
```

#### Transaction Pattern

```go
// Using the common DB wrapper's WithTenant pattern
func (r *InvoiceRepository) Create(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) (*models.Invoice, error) {
    var createdInvoice *models.Invoice
    
    err := r.db.WithTenant(ctx, invoice.OrgID, func(tx database.Tx) error {
        // Insert invoice
        err := tx.QueryRow(ctx, `
            INSERT INTO invoices (...) VALUES (...) RETURNING id, created_at, updated_at
        `, args...).Scan(&invoice.ID, &invoice.CreatedAt, &invoice.UpdatedAt)
        if err != nil {
            return err
        }
        
        // Insert items
        for _, item := range items {
            item.InvoiceID = invoice.ID
            item.OrgID = invoice.OrgID
            _, err := tx.Exec(ctx, `
                INSERT INTO invoice_items (...) VALUES (...)
            `, itemArgs...)
            if err != nil {
                return err
            }
        }
        
        createdInvoice = invoice
        createdInvoice.Items = items
        return nil
    })
    
    return createdInvoice, err
}
```

### 8.2 Service Layer

#### Invoice Service

```go
// service/invoice_service.go

type InvoiceService struct {
    repo            InvoiceRepository
    taxService      *TaxService
    sequenceService *SequenceService
    settingsRepo    SettingsRepository
    metricsRepo     MetricsRepository
}

func NewInvoiceService(repo InvoiceRepository, taxService *TaxService, sequenceService *SequenceService, settingsRepo SettingsRepository, metricsRepo MetricsRepository) *InvoiceService {
    return &InvoiceService{
        repo: repo,
        taxService: taxService,
        sequenceService: sequenceService,
        settingsRepo: settingsRepo,
        metricsRepo: metricsRepo,
    }
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, orgID string, req *models.CreateInvoiceRequest) (*models.Invoice, error) {
    // 1. Validate due date >= invoice date
    invoiceDate, _ := time.Parse("2006-01-02", req.InvoiceDate)
    dueDate, _ := time.Parse("2006-01-02", req.DueDate)
    if dueDate.Before(invoiceDate) {
        return nil, models.ErrInvalidDueDate
    }
    
    if len(req.Items) == 0 {
        return nil, models.ErrEmptyInvoiceItems
    }
    
    // 2. Get invoice number
    invoiceNumber, err := s.sequenceService.NextInvoiceNumber(ctx, orgID)
    if err != nil {
        return nil, fmt.Errorf("failed to generate invoice number: %w", err)
    }
    
    // 3. Get tax rate
    taxRate, err := s.settingsRepo.GetDefaultTaxRate(ctx, orgID)
    if err != nil {
        taxRate = 0 // Default to no tax if settings missing
    }
    
    // 4. Build invoice
    invoice := &models.Invoice{
        OrgID:         orgID,
        CustomerName:  req.CustomerName,
        CustomerEmail: req.CustomerEmail,
        BillingAddress: req.BillingAddress,
        InvoiceNumber: invoiceNumber,
        InvoiceDate:   invoiceDate,
        DueDate:       dueDate,
        Status:        models.StatusDraft,
        Currency:      req.Currency,
        Notes:         req.Notes,
    }
    
    // 5. Calculate totals
    items := make([]models.InvoiceItem, len(req.Items))
    subtotal := 0.0
    for i, itemReq := range req.Items {
        itemTotal := float64(itemReq.Quantity) * itemReq.UnitPrice
        items[i] = models.InvoiceItem{
            Description: itemReq.Description,
            Quantity:    itemReq.Quantity,
            UnitPrice:   itemReq.UnitPrice,
            TaxPercent:  itemReq.TaxPercent,
            Total:       itemTotal,
        }
        subtotal += itemTotal
    }
    
    taxAmount := s.taxService.CalculateTax(subtotal, taxRate)
    totalAmount := subtotal + taxAmount
    
    invoice.Subtotal = subtotal
    invoice.TaxAmount = taxAmount
    invoice.TotalAmount = totalAmount
    invoice.AmountPaid = 0
    invoice.BalanceDue = totalAmount
    invoice.AmountRemaining = totalAmount
    
    // 6. Persist
    return s.repo.Create(ctx, invoice, items)
}
```

#### Tax Service

```go
// service/tax_service.go

type TaxService struct {
    settingsRepo SettingsRepository
}

func NewTaxService(settingsRepo SettingsRepository) *TaxService {
    return &TaxService{settingsRepo: settingsRepo}
}

func (s *TaxService) CalculateTax(subtotal float64, taxRate float64) float64 {
    return subtotal * (taxRate / 100.0)
}

// RoundTo2 rounds to 2 decimal places (standard for currency)
func RoundTo2(val float64) float64 {
    return math.Round(val*100) / 100
}
```

#### Sequence Service

```go
// service/sequence_service.go

type SequenceService struct {
    repo SequenceRepository
}

func NewSequenceService(repo SequenceRepository) *SequenceService {
    return &SequenceService{repo: repo}
}

func (s *SequenceService) NextInvoiceNumber(ctx context.Context, orgID string) (string, error) {
    return s.repo.NextInvoiceNumber(ctx, orgID)
}
```

### 8.3 Handler Layer

#### Invoice Handler

```go
// handlers/invoice_handler.go

type InvoiceHandler struct {
    service *service.InvoiceService
    logger  *zap.Logger
}

func NewInvoiceHandler(service *service.InvoiceService, logger *zap.Logger) *InvoiceHandler {
    return &InvoiceHandler{service: service, logger: logger}
}

func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
    orgID := middleware.MustGetOrgID(r.Context())
    
    var req models.CreateInvoiceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
        return
    }
    
    if err := validator.New().Struct(req); err != nil {
        h.respondValidationError(w, err)
        return
    }
    
    invoice, err := h.service.CreateInvoice(r.Context(), orgID, &req)
    if err != nil {
        h.logger.Error("Failed to create invoice", zap.Error(err))
        h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create invoice")
        return
    }
    
    h.respondJSON(w, http.StatusCreated, invoice)
}

// Helper: standardized error response
func (h *InvoiceHandler) respondError(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{
        "code":    code,
        "message": message,
    })
}

// Helper: standardized success response
func (h *InvoiceHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}
```

### 8.4 PDF Engine

#### platform/pdf/engine.go

```go
package pdf

import (
    "bytes"
    "context"
    "fmt"
    "html/template"
    "sync"
    "time"

    "github.com/chromedp/chromedp"
    "go.uber.org/zap"
)

// PDFEngine generates PDFs from HTML templates using chromedp
type PDFEngine struct {
    logger    *zap.Logger
    browserCtx context.Context
    cancel     context.CancelFunc
    mu         sync.Mutex // Protects browserCtx access
}

// NewPDFEngine initializes the PDF generation engine
func NewPDFEngine(logger *zap.Logger) (*PDFEngine, error) {
    // Create a shared browser context (reuses Chrome process)
    ctx, cancel := chromedp.NewContext(context.Background())
    
    return &PDFEngine{
        logger:    logger,
        browserCtx: ctx,
        cancel:    cancel,
    }, nil
}

// GenerateFromTemplate renders an HTML template with data and returns PDF bytes
func (e *PDFEngine) GenerateFromTemplate(ctx context.Context, tmplContent string, data interface{}) ([]byte, error) {
    // 1. Execute Go template
    tmpl, err := template.New("invoice").Parse(tmplContent)
    if err != nil {
        return nil, fmt.Errorf("failed to parse template: %w", err)
    }
    
    var htmlBuf bytes.Buffer
    if err := tmpl.Execute(&htmlBuf, data); err != nil {
        return nil, fmt.Errorf("failed to execute template: %w", err)
    }
    
    // 2. Generate PDF from HTML
    return e.GenerateFromHTML(ctx, htmlBuf.Bytes())
}

// GenerateFromHTML converts HTML bytes to PDF
func (e *PDFEngine) GenerateFromHTML(ctx context.Context, htmlContent []byte) ([]byte, error) {
    e.mu.Lock()
    browserCtx := e.browserCtx
    e.mu.Unlock()
    
    // Create a new tab for this PDF generation
    tabCtx, tabCancel := chromedp.NewContext(browserCtx)
    defer tabCancel()
    
    var pdfBuf []byte
    
    // Navigate to data: URI with HTML content
    dataURI := fmt.Sprintf("data:text/html;charset=utf-8,%s", htmlContent)
    
    err := chromedp.Run(tabCtx,
        chromedp.Navigate(dataURI),
        chromedp.EmulateMedia(chromedp.MediaPrint),
        chromedp.ActionFunc(func(ctx context.Context) error {
            var err error
            pdfBuf, _, err = chromedp.PrintToPDF(ctx,
                chromedp.PrintToPaperWidth(8.27),  // A4 width in inches
                chromedp.PrintToPaperHeight(11.69), // A4 height in inches
                chromedp.PrintToPDFLandscape(false),
            )
            return err
        }),
    )
    
    if err != nil {
        return nil, fmt.Errorf("chromedp PDF generation failed: %w", err)
    }
    
    return pdfBuf, nil
}

// Close shuts down the browser context
func (e *PDFEngine) Close() {
    e.cancel()
}
```

### 8.5 Response Envelope

The frontend `apiClient` interceptor unwraps `.data` automatically. Backend responses must follow this format:

**List endpoints:**
```json
{
  "data": [...],
  "total": 42,
  "page": 1,
  "size": 10
}
```

**Single resource endpoints:**
```json
{
  "id": "...",
  "customer_name": "...",
  ...
}
```

**Error responses:**
```json
{
  "code": "validation_error",
  "message": "Request body validation failed",
  "details": [
    {"field": "customer_name", "message": "is required"}
  ]
}
```

---

## 9. Implementation Plan

### 9.1 Phase Breakdown

```
Phase 1: Foundation (Day 1-2)
├── models/invoice.go        — Domain structs + DTOs
├── models/settings.go       — Settings + Metrics structs
├── models/errors.go         — Domain errors
├── repository/invoice_repo.go — CRUD operations
├── repository/settings_repo.go — Settings CRUD
├── repository/metrics_repo.go — Metrics aggregation
├── repository/sequence_repo.go — Invoice numbering
└── repository/template_repo.go — Template fetching

Phase 2: Service Layer (Day 2-3)
├── service/invoice_service.go — Business logic
├── service/tax_service.go     — Tax calculation
└── service/sequence_service.go — Number generation

Phase 3: HTTP Handlers (Day 3-4)
├── handlers/invoice_handler.go — 9 admin endpoints
├── handlers/pdf_handler.go    — PDF generation
└── handlers/invoice_handler_test.go — Unit tests

Phase 4: PDF Engine (Day 4-5)
├── platform/pdf/engine.go     — chromedp PDF generator
└── platform/billing/routes.go — Route registration

Phase 5: Frontend Integration (Day 5-6)
├── billingService.ts fixes    — API path corrections (if needed)
├── InvoiceForm.tsx wiring     — Connect create/edit
├── InvoiceDetail.tsx wiring   — PDF download
└── BillingDashboard.tsx wiring — Real metrics

Phase 6: Testing & Polish (Day 6-7)
├── Integration tests          — HTTP endpoint tests
├── E2E test                   — Create → View → Download PDF
└── Load test                  — 100 concurrent PDF generations
```

### 9.2 File Creation Order

| Order | File | Lines (est.) | Dependencies |
|-------|------|--------------|--------------|
| 1 | `models/errors.go` | 40 | None |
| 2 | `models/invoice.go` | 120 | errors.go |
| 3 | `models/settings.go` | 60 | errors.go |
| 4 | `repository/sequence_repo.go` | 80 | models |
| 5 | `repository/settings_repo.go` | 100 | models |
| 6 | `repository/invoice_repo.go` | 200 | models |
| 7 | `repository/metrics_repo.go` | 80 | models |
| 8 | `repository/template_repo.go` | 60 | models |
| 9 | `service/tax_service.go` | 30 | models, repository |
| 10 | `service/sequence_service.go` | 30 | repository |
| 11 | `service/invoice_service.go` | 180 | models, all repos, services |
| 12 | `handlers/invoice_handler.go` | 250 | service, middleware |
| 13 | `platform/pdf/engine.go` | 120 | chromedp |
| 14 | `handlers/pdf_handler.go` | 100 | pdf engine, service |
| 15 | `platform/billing/routes.go` | 60 | all handlers, repos |
| 16 | `handlers/invoice_handler_test.go` | 200 | handlers, mocks |
| 17 | `cmd/main.go` | 80 | routes, config |
| 18 | `templates/invoice_default.html` | 150 | — |

**Total:** ~1,960 lines across 18 files

### 9.3 go.mod Additions

```go
require (
    github.com/go-chi/chi/v5 v5.2.1
    github.com/go-playground/validator/v10 v10.25.0
    github.com/chromedp/chromedp v0.14.2
    github.com/go-fpdf/fpdf v0.9.0  // Fallback PDF engine
)
```

---

## 10. Quality Assurance

### 10.1 Testing Strategy

| Layer | Test Type | Tool | Coverage Target |
|-------|-----------|------|-----------------|
| Models | Unit | `go test` | N/A (structs only) |
| Repository | Unit (mock DB) | `sqlmock` | ≥80% |
| Service | Unit (mock repos) | `go test` | ≥80% |
| Handler | Integration | `httptest` | All endpoints |
| PDF Engine | Unit | `chromedp` | Happy path + error |
| E2E | End-to-end | Playwright | Create → View → PDF |
| Load | Performance | `k6` or `wrk` | 100 concurrent PDF |

### 10.2 Test Examples

#### Repository Unit Test

```go
func TestInvoiceRepository_Create(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()
    
    repo := NewInvoiceRepository(&mockDB{db: db})
    
    invoice := &models.Invoice{...}
    items := []models.InvoiceItem{...}
    
    mock.ExpectBegin()
    mock.ExpectQuery("INSERT INTO invoices").WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow("uuid-1", time.Now(), time.Now()))
    mock.ExpectExec("INSERT INTO invoice_items").WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectCommit()
    
    result, err := repo.Create(context.Background(), invoice, items)
    assert.NoError(t, err)
    assert.Equal(t, "uuid-1", result.ID)
}
```

#### Service Unit Test

```go
func TestInvoiceService_CreateInvoice_DueDateBeforeInvoiceDate(t *testing.T) {
    svc := NewInvoiceService(mockRepo, mockTax, mockSeq, mockSettings, mockMetrics)
    
    req := &models.CreateInvoiceRequest{
        InvoiceDate: "2026-04-15",
        DueDate:     "2026-04-01", // Before invoice date
        Items:       []models.CreateInvoiceItem{{Description: "Test", Quantity: 1, UnitPrice: 100}},
    }
    
    result, err := svc.CreateInvoice(context.Background(), "org-1", req)
    assert.Nil(t, result)
    assert.Equal(t, models.ErrInvalidDueDate, err)
}
```

#### Handler Integration Test

```go
func TestInvoiceHandler_CreateInvoice(t *testing.T) {
    handler := NewInvoiceHandler(mockService, zap.NewNop())
    
    body := `{"customer_name":"Test","invoice_date":"2026-04-01","due_date":"2026-04-15","currency":"INR","items":[{"description":"Test","quantity":1,"unit_price":100}]}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/billing/invoices", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Organization-ID", "org-1")
    rec := httptest.NewRecorder()
    
    handler.CreateInvoice(rec, req)
    
    assert.Equal(t, http.StatusCreated, rec.Code)
    var resp models.Invoice
    json.NewDecoder(rec.Body).Decode(&resp)
    assert.Equal(t, "Test", resp.CustomerName)
}
```

### 10.3 Acceptance Criteria Checklist

#### Backend
- [ ] All 11 API endpoints return correct responses (200/201/204)
- [ ] Invoice creation generates auto-incremented invoice_number (INV-2026-001)
- [ ] Tax calculation matches billing_settings.default_tax_rate
- [ ] Soft delete sets is_deleted = true (not physical delete)
- [ ] Record payment updates amount_paid and recalculates balance_due
- [ ] PDF generation returns valid PDF (Content-Type: application/pdf, >0 bytes)
- [ ] PDF contains: invoice number, customer, line items, subtotal, tax, total, due date
- [ ] Metrics endpoint returns accurate aggregations

#### Frontend
- [ ] InvoiceForm.tsx successfully creates invoices via API
- [ ] InvoiceDetail.tsx shows invoice with line items + PDF download
- [ ] BillingDashboard.tsx shows real metrics (not mock data)
- [ ] InvoicesList.tsx (portal) shows customer's invoices with PDF links

#### Quality
- [ ] Unit tests for repository layer (≥80% coverage)
- [ ] Unit tests for service layer (≥80% coverage)
- [ ] Integration tests for all 11 endpoints
- [ ] No goroutine leaks in chromedp (proper context cancellation)
- [ ] All database constraints respected (FK, CHECK, NOT NULL)

---

## 11. Deployment Strategy

### 11.1 Infrastructure Requirements

| Resource | Requirement | Notes |
|----------|-------------|-------|
| **PostgreSQL** | Existing | No schema changes needed |
| **Chrome/Chromium** | Required for chromedp | Docker image: `chromedp/headless-shell` |
| **Memory** | +200MB per Chrome process | Shared browser context reduces overhead |
| **CPU** | Moderate | PDF generation is CPU-bound |
| **Disk** | Minimal | PDFs stored in DB, not filesystem |

### 11.2 Docker Configuration

```dockerfile
# services/fleet-core/Dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o monolith ./cmd/main.go

# Runtime stage with Chrome
FROM chromedp/headless-shell:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/monolith /usr/local/bin/
COPY --from=builder /app/templates /app/templates

EXPOSE 8084

CMD ["/usr/local/bin/monolith"]
```

### 11.3 Environment Variables

```bash
# Billing Module
BILLING_ENABLED=true
BILLING_DEFAULT_TAX_RATE=18.0
BILLING_BASE_CURRENCY=INR
BILLING_GRACE_PERIOD_DAYS=14

# PDF Engine
PDF_ENGINE=chromedp
PDF_ENGINE_FALLBACK=gofpdf
PDF_TIMEOUT=30s
```

### 11.4 Migration Strategy

**No new migrations required** — all 15 billing migrations already exist.

**Pre-deployment checklist:**
- [ ] Verify migration 052 has been applied (billing_invoice_sequences table)
- [ ] Verify migration 036 has been applied (invoice runtime columns)
- [ ] Seed default HTML template into `billing_invoice_templates` (or use file-based)
- [ ] Initialize `billing_invoice_sequences` for existing orgs

---

## 12. Risk Register

| ID | Risk | Impact | Probability | Mitigation | Owner |
|----|------|--------|-------------|------------|-------|
| R1 | chromedp requires Chrome binary on server | High | Medium | Use `chromedp/headless-shell` Docker image; fallback to gofpdf | DevOps |
| R2 | PDF generation slow at scale (>5s) | Medium | Low | Cache PDFs in `billing_invoice_documents`; async queue for bulk generation (Phase 2) | Backend |
| R3 | Template rendering bugs (Go html/template) | Medium | Medium | Unit test templates with sample data before deploy; use data: URI approach | Backend |
| R4 | Invoice sequence race condition | High | Low | Use table-based sequence with UPDATE ... RETURNING (atomic); row-level lock | Backend |
| R5 | Frontend/backend status mismatch | Medium | Medium | Standardize: lowercase in DB, uppercase in API responses, convert in service layer | Full-stack |
| R6 | Large file bloat (existing pattern) | Low | High | Keep billing files focused; target <300 lines per file; split if larger | Architect |
| R7 | Multi-tenant data leak | Critical | Low | StrictTenantMiddleware on all admin routes; RLS at DB level; org_id on every query | Security |
| R8 | chromedp goroutine leak | Medium | Medium | Proper context cancellation in defer; timeout on all chromedp operations | Backend |
| R9 | Frontend API path mismatch | Medium | High | Verify all paths match frontend billingService.ts expectations before deploy | Full-stack |

---

## 13. Future Phases

### Phase 2: Automation (Week 4-5)

| Feature | Description | Complexity |
|---------|-------------|------------|
| Monthly auto-invoicing | Cron job generates invoices for active subscriptions | Medium |
| Email delivery | SendGrid/SES integration for invoice delivery | Medium |
| Payment gateway | Razorpay (India) / Stripe (global) integration | High |
| gRPC BillingService | Implement proto/billing/v1/billing.proto RPCs | Medium |
| Dunning management | Retry failed payments with exponential backoff | Medium |

### Phase 3: Advanced (Week 6-8)

| Feature | Description | Complexity |
|---------|-------------|------------|
| Multi-currency | FX rates, currency conversion at invoice time | High |
| Region-based tax | CGST/SGST/IGST split for India; VAT for EU | High |
| Credit notes | Refunds, adjustments, negative invoices | Medium |
| Accounting exports | QuickBooks, Tally, Xero export formats | Medium |
| Usage-based billing | TimescaleDB aggregation for metered billing | High |
| Approval workflows | Multi-level invoice approval before sending | Medium |
| e-Invoice IRN | India GST e-Invoice integration via NIC API | High |

---

## 14. Appendices

### Appendix A: Existing Frontend Routes

```typescript
// apps/dashboard-web/src/App.tsx

/admin/billing              → BillingDashboard
/admin/billing/new          → InvoiceForm (create mode)
/admin/billing/invoices/:id → InvoiceDetail (view mode)
/admin/billing/invoices/:id/edit → InvoiceForm (edit mode)
/admin/billing/settings     → BillingSettingsPage
```

### Appendix B: Existing HTML Templates

| File | Purpose | Status |
|------|---------|--------|
| `bill_maintenance.html` | Maintenance billing template | Static, needs DB seeding |
| `bill_transport.html` | Transport billing template | Static, needs DB seeding |

**Recommendation:** Use these as the basis for `templates/invoice_default.html` and `templates/invoice_minimal.html`. Seed into `billing_invoice_templates` table on first run.

### Appendix C: Proto-to-REST Mapping

The existing `proto/billing/v1/billing.proto` defines gRPC RPCs that are **not** implemented in this MVP. The REST endpoints provide equivalent functionality:

| gRPC RPC | REST Equivalent |
|----------|-----------------|
| `GetMetrics` | `GET /api/v1/admin/billing/metrics` |
| `ListInvoices` | `GET /api/v1/admin/billing/invoices` |
| `GetSubscription` | Not in MVP scope |

### Appendix D: Status State Machine

```
┌───────┐     ┌──────┐     ┌────────┐     ┌──────┐
│ DRAFT │────►│ SENT │────►│ VIEWED │────►│ PAID │
└───────┘     └──────┘     └────────┘     └──────┘
     │            │            │
     │            ▼            ▼
     │       ┌─────────┐  ┌─────────┐
     └──────►│ OVERDUE │  │ OVERDUE │
             └─────────┘  └─────────┘
                  │            │
                  ▼            ▼
             ┌──────┐     ┌──────┐
             │ PAID │     │PARTIAL│
             └──────┘     └──────┘
                               │
                               ▼
                          ┌──────┐
                          │ PAID │
                          └──────┘
```

**MVP Transitions:**
- Create → `DRAFT`
- Send (manual) → `SENT`
- Customer views → `VIEWED` (deferred to Phase 2 — requires portal tracking)
- Due date passes, unpaid → `OVERDUE` (cron job, Phase 2)
- Full payment → `PAID`
- Partial payment → `PARTIAL`

### Appendix E: Glossary

| Term | Definition |
|------|------------|
| **Org ID** | UUID identifying the tenant/organization |
| **RLS** | Row Level Security — PostgreSQL feature for multi-tenant isolation |
| **chromedp** | Go library for controlling headless Chrome via DevTools Protocol |
| **gofpdf** | Pure Go PDF generation library (lightweight fallback) |
| **DTO** | Data Transfer Object — request/response structs |
| **MVP** | Minimum Viable Product — Phase 1 scope |
| **BFF** | Backend-for-Frontend — API aggregator layer |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-04-03 | Senior Architect | Initial version |

---

## Approvals

| Role | Name | Status | Date |
|------|------|--------|------|
| Author | Senior Architect | ✅ Approved | 2026-04-03 |
| Tech Lead | TBD | ⏳ Pending | — |
| Security | TBD | ⏳ Pending | — |
| Product | TBD | ⏳ Pending | — |

---

*This document is a living specification. Update it as decisions are made during implementation.*
