# Custom Billing Configuration - Per Organization

## Overview

Each organization using FlyFleet can **customize their invoices** with:
- Company branding (logo, colors)
- Company details (GSTIN, PAN, address)
- Bank account details
- Custom terms & conditions
- Feature toggles (HSN, QR code, payment info)

---

## API Endpoints

### Get Billing Configuration

```http
GET /api/v1/billing/billing-config
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "uuid",
  "org_id": "uuid",
  "company_name": "Your Company Pvt Ltd",
  "company_gstin": "29ABCDE1234F1Z5",
  "company_pan": "ABCDE1234F",
  "primary_color": "#3b82f6",
  "bank_name": "HDFC Bank",
  "bank_account_number": "123456789012",
  "bank_ifsc": "HDFC0001234",
  "invoice_prefix": "INV",
  "show_hsn": true,
  "show_qr_code": true,
  ...
}
```

### Update Billing Configuration

```http
PUT /api/v1/billing/billing-config
Authorization: Bearer <token>
Content-Type: application/json

{
  "company_name": "Your Company Pvt Ltd",
  "company_address": "123 Street, City, State",
  "company_gstin": "29ABCDE1234F1Z5",
  "company_pan": "ABCDE1234F",
  "company_email": "billing@company.com",
  "company_phone": "+91-80-1234-5678",
  "company_website": "www.company.com",
  "company_logo_url": "data:image/png;base64,...",
  
  "primary_color": "#3b82f6",
  "secondary_color": "#1e40af",
  "theme": "modern",
  
  "bank_name": "HDFC Bank",
  "bank_account_name": "Your Company Pvt Ltd",
  "bank_account_number": "123456789012",
  "bank_ifsc": "HDFC0001234",
  "bank_branch": "Koramangala",
  "bank_upi": "company@oksbi",
  
  "invoice_prefix": "INV",
  "invoice_terms": [
    "Payment due within 30 days",
    "Interest @18% on delayed payments"
  ],
  "invoice_notes": "Thank you for your business!",
  
  "show_hsn": true,
  "show_qr_code": true,
  "show_payment_info": true,
  
  "default_currency": "INR",
  "default_payment_terms_days": 30
}
```

### Delete Billing Configuration

```http
DELETE /api/v1/billing/billing-config
Authorization: Bearer <token>
```

---

## Database Schema

```sql
billing_configurations (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL UNIQUE,  -- One config per org
    
    -- Company
    company_name VARCHAR(255),
    company_address TEXT,
    company_gstin VARCHAR(20),
    company_pan VARCHAR(10),
    company_email VARCHAR(255),
    company_phone VARCHAR(20),
    company_website VARCHAR(255),
    company_logo_url TEXT,
    
    -- Branding
    primary_color VARCHAR(7),
    secondary_color VARCHAR(7),
    theme VARCHAR(50),
    
    -- Bank
    bank_name VARCHAR(255),
    bank_account_name VARCHAR(255),
    bank_account_number VARCHAR(50),
    bank_ifsc VARCHAR(20),
    bank_branch VARCHAR(255),
    bank_upi VARCHAR(255),
    
    -- Invoice
    invoice_prefix VARCHAR(20),
    invoice_terms TEXT[],
    invoice_notes TEXT,
    show_hsn BOOLEAN,
    show_qr_code BOOLEAN,
    show_payment_info BOOLEAN,
    
    -- Defaults
    default_currency VARCHAR(10),
    default_tax_calculation VARCHAR(20),
    default_payment_terms_days INTEGER,
    
    is_active BOOLEAN,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)
```

---

## Usage Flow

### 1. Organization Onboarding

```go
// When org signs up, create default config
config := models.DefaultBillingConfig(orgID)
config.CompanyName = org.Name
billingConfigRepo.Create(ctx, config)
```

### 2. User Customizes Invoice

```bash
# User updates branding
curl -X PUT /api/v1/billing/billing-config \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Acme Pvt Ltd",
    "primary_color": "#ff6b6b",
    "bank_name": "ICICI Bank",
    ...
  }'
```

### 3. Invoice Generated with Custom Config

```go
// Get org's billing config
config, _ := billingConfigService.GetConfig(ctx, orgID)

// Generate PDF with custom branding
pdf, _ := invoiceService.GeneratePDFWithConfig(ctx, invoice, items, config)
```

---

## Features

### Customizable Elements

| Element | Default | Customizable |
|---------|---------|--------------|
| Company Name | ✅ | ✅ |
| Company Logo | ✅ | ✅ |
| Company Address | ✅ | ✅ |
| GSTIN/PAN | ✅ | ✅ |
| Primary Color | #3b82f6 | ✅ |
| Theme | modern | ✅ (modern/classic/minimal) |
| Bank Details | ✅ | ✅ |
| UPI ID | ✅ | ✅ |
| Invoice Prefix | INV | ✅ |
| Terms & Conditions | Default 4 | ✅ |
| Invoice Notes | ✅ | ✅ |
| Show HSN | true | ✅ |
| Show QR Code | true | ✅ |
| Show Payment Info | true | ✅ |
| Currency | INR | ✅ |
| Payment Terms | 30 days | ✅ |

---

## Validation Rules

| Field | Validation |
|-------|------------|
| `company_gstin` | 15 characters (Indian GST) |
| `company_pan` | 10 characters |
| `company_email` | Valid email format |
| `primary_color` | Hex color (#RRGGBB) |
| `default_currency` | 3 characters (ISO 4217) |
| `bank_account_number` | 9-18 digits |
| `bank_ifsc` | 11 characters |

---

## Example Configurations

### Modern Tech Company

```json
{
  "company_name": "TechCorp Solutions",
  "primary_color": "#6366f1",
  "theme": "modern",
  "show_qr_code": true,
  "bank_upi": "techcorp@oksbi"
}
```

### Traditional Manufacturing

```json
{
  "company_name": "ABC Manufacturing",
  "primary_color": "#dc2626",
  "theme": "classic",
  "show_hsn": true,
  "invoice_terms": [
    "Payment by cheque only",
    "Subject to local jurisdiction"
  ]
}
```

### Minimalist Startup

```json
{
  "company_name": "Startup Inc",
  "primary_color": "#000000",
  "theme": "minimal",
  "show_payment_info": false,
  "show_qr_code": true
}
```

---

## Migration

Run migration to add billing config table:

```bash
psql $DATABASE_URL -f migrations/048_custom_billing_config.sql
```

This will:
1. Create `billing_configurations` table
2. Add indexes for performance
3. Create default configs for existing organizations

---

## Testing

### Test Get Config

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8084/api/v1/billing/billing-config | jq
```

### Test Update Config

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"company_name": "Test Corp", "primary_color": "#ff0000"}' \
  http://localhost:8084/api/v1/billing/billing-config | jq
```

### Test Invoice with Custom Config

```bash
# 1. Update config
# 2. Create invoice
# 3. Download PDF
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8084/api/v1/billing/invoices/{id}/pdf \
  -o custom-invoice.pdf

# 4. Verify branding matches config
```

---

## Benefits

| Benefit | Description |
|---------|-------------|
| **White-label** | Each org looks professional |
| **Compliance** | GSTIN, PAN on every invoice |
| **Branding** | Colors, logo match company |
| **Flexibility** | Toggle features per org |
| **Multi-tenant** | One system, many brands |

---

## Security

- ✅ Config scoped to organization (`org_id`)
- ✅ Only org admins can modify
- ✅ Audit trail (created_by, updated_by)
- ✅ Soft delete (is_active flag)
- ✅ Input validation (GSTIN, PAN, email)

---

## Future Enhancements

- [ ] Logo upload via S3
- [ ] Multiple bank accounts
- [ ] Custom invoice templates
- [ ] Multi-language support
- [ ] Email template customization
- [ ] Automated logo extraction from URL

---

## License

Part of FlyFleet - Enterprise Fleet Management Platform
