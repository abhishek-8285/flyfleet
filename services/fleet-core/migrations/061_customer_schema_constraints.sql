-- Migration: 061_customer_schema_constraints
-- Description: Add data integrity constraints and indexes with tenant isolation.

-- 1. Resilient constraints for customers
DO $$ 
BEGIN 
    ALTER TABLE customers ADD CONSTRAINT chk_customers_email_format 
        CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
    ALTER TABLE customers ADD CONSTRAINT chk_customers_status 
        CHECK (status IN ('active', 'inactive', 'suspended'));
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Resilient constraints for invoices
DO $$ 
BEGIN 
    ALTER TABLE invoices ADD CONSTRAINT chk_invoices_status
        CHECK (status IN ('draft', 'issued', 'paid', 'overdue', 'cancelled'));
    ALTER TABLE invoices ADD CONSTRAINT chk_invoices_amount
        CHECK (total_amount > 0);
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 3. Indexes
CREATE INDEX IF NOT EXISTS idx_customers_tenant_status ON customers(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_customer_status ON invoices(customer_id, status);

-- 4. Views
CREATE OR REPLACE VIEW customer_overdue_invoices AS
SELECT 
    i.id,
    i.tenant_id,
    i.customer_id,
    i.invoice_number,
    i.total_amount as amount,
    i.due_date,
    i.status
FROM invoices i
WHERE i.status = 'overdue' OR (i.status = 'issued' AND i.due_date < CURRENT_DATE);
