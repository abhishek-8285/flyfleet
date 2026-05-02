-- Migration: 036_customer_portal
-- Description: Standardizes customer portal schema with full self-healing logic.

-- 1. Resilient Customers Table
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customers') THEN
        CREATE TABLE customers (id UUID PRIMARY KEY DEFAULT uuid_generate_v4());
    END IF;

    -- Standardize Columns
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS tenant_id UUID;
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS name VARCHAR(255);
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255);
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(50);
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS gst_number VARCHAR(50);
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS api_key VARCHAR(255);
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS api_secret_hash VARCHAR(255);
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS address JSONB;
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS billing_address JSONB;
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS metadata JSONB;
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE customers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

    -- Constraints
    IF NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'customers' AND column_name = 'contact_email') THEN
        ALTER TABLE customers ADD CONSTRAINT unique_customer_email UNIQUE (contact_email);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customers_contact_email ON customers(contact_email);
CREATE INDEX IF NOT EXISTS idx_customers_api_key ON customers(api_key);

-- 2. Resilient Shipments Table
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'shipments') THEN
        CREATE TABLE shipments (id UUID PRIMARY KEY DEFAULT uuid_generate_v4());
    END IF;

    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id);
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS trip_id UUID;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(100);
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'created';
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS origin JSONB;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS destination JSONB;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS current_location JSONB;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS estimated_delivery TIMESTAMPTZ;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS actual_delivery TIMESTAMPTZ;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS cargo_details JSONB;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pod_url VARCHAR(255);
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pod_uploaded_at TIMESTAMPTZ;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS metadata JSONB;
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE shipments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

    IF NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'shipments' AND column_name = 'tracking_number') THEN
        ALTER TABLE shipments ADD CONSTRAINT unique_tracking_number UNIQUE (tracking_number);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_shipments_customer_id ON shipments(customer_id);
CREATE INDEX IF NOT EXISTS idx_shipments_tracking_number ON shipments(tracking_number);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);

-- 3. Resilient Invoices Table
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoices') THEN
        CREATE TABLE invoices (id UUID PRIMARY KEY DEFAULT uuid_generate_v4());
    END IF;

    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id);
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS shipment_id UUID REFERENCES shipments(id);
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100);
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'draft';
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS amount DECIMAL(10, 2);
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'USD';
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS issue_date DATE;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS due_date DATE;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS paid_date DATE;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS pdf_url VARCHAR(255);
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS metadata JSONB;
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE invoices ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

    IF NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'invoices' AND column_name = 'invoice_number') THEN
        ALTER TABLE invoices ADD CONSTRAINT unique_invoice_number UNIQUE (invoice_number);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);

-- 4. Resilient Booking Requests Table
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'booking_requests') THEN
        CREATE TABLE booking_requests (id UUID PRIMARY KEY DEFAULT uuid_generate_v4());
    END IF;

    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id);
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS request_type VARCHAR(50) DEFAULT 'standard';
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS origin JSONB;
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS destination JSONB;
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS requested_pickup_time TIMESTAMPTZ;
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS cargo_details JSONB;
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'pending';
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS notes TEXT;
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS metadata JSONB;
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE booking_requests ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
END $$;

CREATE INDEX IF NOT EXISTS idx_booking_requests_customer_id ON booking_requests(customer_id);
CREATE INDEX IF NOT EXISTS idx_booking_requests_status ON booking_requests(status);
