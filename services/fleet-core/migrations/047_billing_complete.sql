-- Migration: 047_billing_complete
-- Description: Completes the billing module schema by adding missing columns to invoices and creating supporting tables.

-- 1. Resiliently update invoices table
DO $$ 
BEGIN 
    -- Ensure table exists (it should from 036, but safety first)
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoices') THEN
        CREATE TABLE invoices (id UUID PRIMARY KEY DEFAULT gen_random_uuid());
    END IF;

    -- Standardize Identity
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    
    -- Sync columns with Billing models
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_name VARCHAR(255);
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_date DATE; -- alias for issue_date if needed
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS total_amount DECIMAL(12, 2) DEFAULT 0;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS amount_paid DECIMAL(12, 2) DEFAULT 0;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS balance_due DECIMAL(12, 2) DEFAULT 0;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS created_by UUID;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS updated_by UUID;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

    -- Handle issue_date / invoice_date duplication
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invoices' AND column_name = 'issue_date') AND
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invoices' AND column_name = 'invoice_date') THEN
        ALTER TABLE invoices RENAME COLUMN issue_date TO invoice_date;
    END IF;
END $$;

-- 2. Create invoice_items table
CREATE TABLE IF NOT EXISTS invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(12, 2) NOT NULL,
    tax_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create billing_settings table
CREATE TABLE IF NOT EXISTS billing_settings (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
    default_tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 18.0,
    base_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    grace_period_days INTEGER NOT NULL DEFAULT 14,
    auto_invoicing_enabled BOOLEAN NOT NULL DEFAULT false,
    invoice_prefix VARCHAR(20) NOT NULL DEFAULT 'INV',
    next_invoice_number INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 4. Create payment_records table
CREATE TABLE IF NOT EXISTS payment_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    payment_method VARCHAR(50),
    payment_reference VARCHAR(255),
    payment_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

-- 5. Create invoice_sequence table
CREATE TABLE IF NOT EXISTS invoice_sequence (
    id INTEGER PRIMARY KEY DEFAULT 1,
    current_value INTEGER NOT NULL DEFAULT 1,
    prefix VARCHAR(20) NOT NULL DEFAULT 'INV',
    year INTEGER NOT NULL DEFAULT EXTRACT(YEAR FROM CURRENT_DATE),
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT single_row CHECK (id = 1)
);

-- 6. Indexes
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);
CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice_id ON invoice_items(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_invoice_id ON payment_records(invoice_id);
