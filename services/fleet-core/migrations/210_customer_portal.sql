-- Migration: 210_customer_portal
-- Description: B2B Customer Portal schema with mandatory tenant isolation.

-- 1. Customer Accounts (V2)
CREATE TABLE IF NOT EXISTS customer_v2_customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    company_name VARCHAR(255) NOT NULL,
    gst_number VARCHAR(50),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cust_v2_tenant ON customer_v2_customers(tenant_id);

-- 2. Customer Users
CREATE TABLE IF NOT EXISTS customer_v2_customer_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(50) DEFAULT 'VIEW_ONLY',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cust_v2_user_tenant ON customer_v2_customer_users(tenant_id);

-- 3. Booking Requests
CREATE TABLE IF NOT EXISTS customer_v2_booking_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    pickup_location JSONB NOT NULL,
    drop_location JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'REQUESTED',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cust_v2_booking_tenant ON customer_v2_booking_requests(tenant_id);

-- 4. Shipments
CREATE TABLE IF NOT EXISTS customer_v2_shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    trip_id UUID,
    status VARCHAR(50) DEFAULT 'BOOKED',
    origin JSONB NOT NULL,
    destination JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cust_v2_shipment_tenant ON customer_v2_shipments(tenant_id);
