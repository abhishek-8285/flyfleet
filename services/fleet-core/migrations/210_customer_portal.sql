-- 210_customer_portal.sql
-- Migration for B2B Customer Portal (customer_v2)

-- 1. Customer Accounts
CREATE TABLE IF NOT EXISTS customer_v2_customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    gst_number VARCHAR(50),
    billing_address JSONB,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED, BLOCKED
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    role VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    address_type VARCHAR(50), -- BILLING, SHIPPING, WAREHOUSE
    address JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    contract_number VARCHAR(100) UNIQUE,
    valid_from DATE NOT NULL,
    valid_to DATE,
    terms JSONB,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_credit_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE UNIQUE,
    credit_limit DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    available_credit DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) DEFAULT 'INR',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Customer Users
CREATE TABLE IF NOT EXISTS customer_v2_customer_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    role VARCHAR(50) DEFAULT 'VIEW_ONLY', -- ADMIN, OPERATIONS, FINANCE, VIEW_ONLY
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES customer_v2_customer_users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    key_name VARCHAR(100) NOT NULL,
    api_key_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Booking Requests
CREATE TABLE IF NOT EXISTS customer_v2_booking_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    pickup_location JSONB NOT NULL,
    drop_location JSONB NOT NULL,
    cargo_type VARCHAR(100),
    weight DECIMAL(10, 2),
    volume DECIMAL(10, 2),
    vehicle_type VARCHAR(100),
    requested_pickup_time TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'REQUESTED', -- REQUESTED, QUOTED, CONFIRMED, REJECTED
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_booking_quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES customer_v2_booking_requests(id) ON DELETE CASCADE,
    quote_amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    valid_until TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_booking_validations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES customer_v2_booking_requests(id) ON DELETE CASCADE,
    validation_type VARCHAR(100) NOT NULL, -- vehicle_availability, driver_availability, customer_credit, route_feasibility
    status VARCHAR(50) NOT NULL, -- PASSED, FAILED
    details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Shipment Domain
CREATE TABLE IF NOT EXISTS customer_v2_shipments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id),
    trip_id UUID,
    booking_id UUID REFERENCES customer_v2_booking_requests(id),
    status VARCHAR(50) DEFAULT 'BOOKED', -- BOOKED, CONFIRMED, ASSIGNED, IN_TRANSIT, AT_CHECKPOINT, DELIVERED, CANCELLED
    origin JSONB NOT NULL,
    destination JSONB NOT NULL,
    distance_km DECIMAL(10, 2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Multi-stop Shipments
CREATE TABLE IF NOT EXISTS customer_v2_shipment_stops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    stop_type VARCHAR(50) NOT NULL, -- PICKUP, DROPOFF, CHECKPOINT
    sequence INT NOT NULL,
    location JSONB NOT NULL,
    expected_arrival TIMESTAMPTZ,
    actual_arrival TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Shipment Events & Logs
CREATE TABLE IF NOT EXISTS customer_v2_shipment_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL, -- shipment_created, vehicle_assigned, trip_started, checkpoint_reached, delivery_completed
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    location JSONB,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS customer_v2_shipment_status_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    changed_by UUID,
    changed_at TIMESTAMPTZ DEFAULT NOW(),
    reason TEXT
);

-- 7. Proof of Delivery
CREATE TABLE IF NOT EXISTS customer_v2_delivery_proofs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE UNIQUE,
    receiver_name VARCHAR(255) NOT NULL,
    receiver_phone VARCHAR(50),
    signature_url VARCHAR(255),
    photo_url VARCHAR(255),
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 8. Shipment Documents
CREATE TABLE IF NOT EXISTS customer_v2_shipment_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL, -- LR_COPY, INVOICE, DELIVERY_NOTE, EWAY_BILL, POD
    file_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_shipment_document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES customer_v2_shipment_documents(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    file_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. Pricing Engine
CREATE TABLE IF NOT EXISTS customer_v2_pricing_contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    contract_name VARCHAR(255) NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_pricing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES customer_v2_pricing_contracts(id) ON DELETE CASCADE,
    pricing_model VARCHAR(50) NOT NULL, -- PER_KM, PER_TRIP, PER_TON, ZONE_BASED, CONTRACT_RATE
    base_rate DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    conditions JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_pricing_components (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID NOT NULL REFERENCES customer_v2_pricing_rules(id) ON DELETE CASCADE,
    component_type VARCHAR(50) NOT NULL, -- SURCHARGE, DISCOUNT, TAX
    value DECIMAL(12, 2) NOT NULL,
    is_percentage BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_pricing_quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES customer_v2_booking_requests(id) ON DELETE CASCADE,
    total_amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    breakdown JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 10. Customer Credit System
CREATE TABLE IF NOT EXISTS customer_v2_customer_wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE UNIQUE,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) DEFAULT 'INR',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_credit_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL REFERENCES customer_v2_customer_wallets(id) ON DELETE CASCADE,
    transaction_type VARCHAR(50) NOT NULL, -- CREDIT, DEBIT
    amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    reference_id UUID, -- shipment_id, booking_id
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_credit_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE UNIQUE,
    max_limit DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) DEFAULT 'INR',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 11. Invoice Access
CREATE TABLE IF NOT EXISTS customer_v2_customer_invoice_views (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID NOT NULL, -- References billing_v2 invoices
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    downloaded_at TIMESTAMPTZ DEFAULT NOW()
);

-- 12. SLA Monitoring
CREATE TABLE IF NOT EXISTS customer_v2_shipment_sla_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    rule_name VARCHAR(255) NOT NULL,
    metric VARCHAR(50) NOT NULL, -- delivery_time
    target_value DECIMAL(10, 2) NOT NULL, -- e.g. 24
    unit VARCHAR(20) NOT NULL, -- hours
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_shipment_sla_violations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES customer_v2_shipment_sla_rules(id),
    actual_value DECIMAL(10, 2) NOT NULL,
    violated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 13. Audit Logging & APIs
CREATE TABLE IF NOT EXISTS customer_v2_customer_api_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_id UUID REFERENCES customer_v2_customer_api_keys(id),
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INT,
    response_time_ms INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_customer_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    user_id UUID REFERENCES customer_v2_customer_users(id),
    event_type VARCHAR(100) NOT NULL, -- booking_created, tracking_viewed, invoice_downloaded, shipment_cancelled
    resource_id UUID,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cust_v2_cust_org_id ON customer_v2_customers(organization_id);
CREATE INDEX IF NOT EXISTS idx_cust_v2_users_cust_id ON customer_v2_customer_users(customer_id);
CREATE INDEX IF NOT EXISTS idx_cust_v2_bookings_cust_id ON customer_v2_booking_requests(customer_id);
CREATE INDEX IF NOT EXISTS idx_cust_v2_shipments_cust_id ON customer_v2_shipments(customer_id);
CREATE INDEX IF NOT EXISTS idx_cust_v2_shipments_trip_id ON customer_v2_shipments(trip_id);
CREATE INDEX IF NOT EXISTS idx_cust_v2_events_shipment_id ON customer_v2_shipment_events(shipment_id);
