-- Migration 259: Standardize Customer Schema
-- Goal: Resolve "relation customers does not exist" and unify B2B/Admin customer tables.

DO $$ 
BEGIN 
    -- 1. If customer_v2_customers exists but customers does not, rename it
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customer_v2_customers') AND
       NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customers') THEN
        ALTER TABLE customer_v2_customers RENAME TO customers;
    END IF;

    -- 2. If customers does not exist at all, create it (base for new installs)
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customers') THEN
        CREATE TABLE customers (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            tenant_id UUID NOT NULL,
            name VARCHAR(255) NOT NULL,
            gst_number VARCHAR(50),
            billing_address JSONB,
            contact_email VARCHAR(255),
            contact_phone VARCHAR(50),
            status VARCHAR(50) DEFAULT 'ACTIVE',
            created_at TIMESTAMPTZ DEFAULT NOW(),
            updated_at TIMESTAMPTZ DEFAULT NOW()
        );
    END IF;

    -- 3. Standardize column names in the customers table
    
    -- Rename organization_id to tenant_id if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'organization_id') AND
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'tenant_id') THEN
        ALTER TABLE customers RENAME COLUMN organization_id TO tenant_id;
    END IF;

    -- Rename company_name to name if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'company_name') AND
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'name') THEN
        ALTER TABLE customers RENAME COLUMN company_name TO name;
    END IF;

    -- Add missing columns that might be needed by the repository
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'contact_email') THEN
        ALTER TABLE customers ADD COLUMN contact_email VARCHAR(255);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'contact_phone') THEN
        ALTER TABLE customers ADD COLUMN contact_phone VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'gst_number') THEN
        ALTER TABLE customers ADD COLUMN gst_number VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'billing_address') THEN
        ALTER TABLE customers ADD COLUMN billing_address JSONB;
    END IF;

END $$;

-- 4. Re-create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name);
