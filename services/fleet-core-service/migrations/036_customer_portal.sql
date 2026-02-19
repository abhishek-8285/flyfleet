-- Create customers table
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    api_key VARCHAR(255) UNIQUE,
    api_secret_hash VARCHAR(255),
    address JSONB,
    billing_info JSONB,
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, suspended
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_customers_org_id ON customers(org_id);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_api_key ON customers(api_key);

-- Create shipments table
CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    trip_id UUID REFERENCES trips(id), -- Nullable if not yet assigned
    tracking_number VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'created', -- created, picked_up, in_transit, delivered, exception
    origin JSONB NOT NULL,
    destination JSONB NOT NULL,
    current_location JSONB,
    estimated_delivery TIMESTAMPTZ,
    actual_delivery TIMESTAMPTZ,
    cargo_details JSONB, -- weight, dimensions, type
    pod_url VARCHAR(255),
    pod_uploaded_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_shipments_customer_id ON shipments(customer_id);
CREATE INDEX idx_shipments_trip_id ON shipments(trip_id);
CREATE INDEX idx_shipments_tracking_number ON shipments(tracking_number);
CREATE INDEX idx_shipments_status ON shipments(status);

-- Create invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    shipment_id UUID REFERENCES shipments(id),
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'draft', -- draft, issued, paid, overdue, cancelled
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    paid_date DATE,
    pdf_url VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_shipment_id ON invoices(shipment_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);

-- Create booking_requests table
CREATE TABLE IF NOT EXISTS booking_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    request_type VARCHAR(50) DEFAULT 'standard', -- standard, express, hazardous
    origin JSONB NOT NULL,
    destination JSONB NOT NULL,
    requested_pickup_time TIMESTAMPTZ,
    cargo_details JSONB,
    status VARCHAR(50) DEFAULT 'pending', -- pending, approved, rejected, converted
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_booking_requests_customer_id ON booking_requests(customer_id);
CREATE INDEX idx_booking_requests_status ON booking_requests(status);
