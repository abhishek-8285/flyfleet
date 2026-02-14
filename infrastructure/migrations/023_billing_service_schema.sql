-- Create invoices table
CREATE TABLE IF NOT EXISTS invoices (
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

-- Create invoice_items table
CREATE TABLE IF NOT EXISTS invoice_items (
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

-- Create billing_settings table
CREATE TABLE IF NOT EXISTS billing_settings (
    org_id UUID PRIMARY KEY,
    default_tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
    base_currency TEXT NOT NULL DEFAULT 'USD',
    grace_period_days INTEGER NOT NULL DEFAULT 14,
    auto_invoicing_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_invoices_org_id ON invoices(org_id);
CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice_id ON invoice_items(invoice_id);
