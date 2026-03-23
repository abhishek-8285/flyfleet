-- Migration: 061_customer_schema_constraints
-- Description: Add data integrity constraints and indexes for customer domain

-- Add CHECK constraints to customers table
ALTER TABLE customers ADD CONSTRAINT chk_customers_email_format 
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

ALTER TABLE customers ADD CONSTRAINT chk_customers_status 
    CHECK (status IN ('active', 'inactive', 'suspended'));

ALTER TABLE customers ADD CONSTRAINT chk_customers_name_length
    CHECK (LENGTH(name) >= 2 AND LENGTH(name) <= 255);

-- Add CHECK constraints to shipments table
ALTER TABLE shipments ADD CONSTRAINT chk_shipments_status
    CHECK (status IN ('created', 'picked_up', 'in_transit', 'delivered', 'exception'));

ALTER TABLE shipments ADD CONSTRAINT chk_shipments_tracking_format
    CHECK (LENGTH(tracking_number) >= 5 AND LENGTH(tracking_number) <= 100);

ALTER TABLE shipments ADD CONSTRAINT chk_shipments_delivery_dates
    CHECK (actual_delivery IS NULL OR actual_delivery >= estimated_delivery);

-- Add CHECK constraints to invoices table
ALTER TABLE invoices ADD CONSTRAINT chk_invoices_status
    CHECK (status IN ('draft', 'issued', 'paid', 'overdue', 'cancelled'));

ALTER TABLE invoices ADD CONSTRAINT chk_invoices_amount
    CHECK (amount > 0);

ALTER TABLE invoices ADD CONSTRAINT chk_invoices_currency
    CHECK (LENGTH(currency) = 3);

ALTER TABLE invoices ADD CONSTRAINT chk_invoices_dates
    CHECK (due_date >= issue_date);

-- Add CHECK constraints to booking_requests table
ALTER TABLE booking_requests ADD CONSTRAINT chk_booking_requests_status
    CHECK (status IN ('pending', 'approved', 'rejected', 'converted'));

ALTER TABLE booking_requests ADD CONSTRAINT chk_booking_requests_type
    CHECK (request_type IN ('standard', 'express', 'hazardous'));

-- Add additional indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_org_status ON customers(org_id, status);
CREATE INDEX IF NOT EXISTS idx_customers_created ON customers(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_shipments_status_created ON shipments(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shipments_customer_status ON shipments(customer_id, status);
CREATE INDEX IF NOT EXISTS idx_shipments_estimated_delivery ON shipments(estimated_delivery) WHERE status = 'in_transit';

CREATE INDEX IF NOT EXISTS idx_invoices_customer_status ON invoices(customer_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date_status ON invoices(due_date, status) WHERE status IN ('issued', 'overdue');

CREATE INDEX IF NOT EXISTS idx_booking_requests_customer_status ON booking_requests(customer_id, status);
CREATE INDEX IF NOT EXISTS idx_booking_requests_created ON booking_requests(created_at DESC);

-- Add comments for documentation
COMMENT ON CONSTRAINT chk_customers_email_format ON customers IS 'Email must be valid format';
COMMENT ON CONSTRAINT chk_customers_status ON customers IS 'Status must be active, inactive, or suspended';
COMMENT ON CONSTRAINT chk_customers_name_length ON customers IS 'Name must be 2-255 characters';

COMMENT ON CONSTRAINT chk_shipments_status ON shipments IS 'Valid shipment status values';
COMMENT ON CONSTRAINT chk_shipments_tracking_format ON shipments IS 'Tracking number must be 5-100 characters';
COMMENT ON CONSTRAINT chk_shipments_delivery_dates ON shipments IS 'Actual delivery cannot be before estimated';

COMMENT ON CONSTRAINT chk_invoices_status ON invoices IS 'Valid invoice status values';
COMMENT ON CONSTRAINT chk_invoices_amount ON invoices IS 'Invoice amount must be positive';
COMMENT ON CONSTRAINT chk_invoices_currency ON invoices IS 'Currency must be 3-character ISO code';
COMMENT ON CONSTRAINT chk_invoices_dates ON invoices IS 'Due date cannot be before issue date';

COMMENT ON CONSTRAINT chk_booking_requests_status ON booking_requests IS 'Valid booking status values';
COMMENT ON CONSTRAINT chk_booking_requests_type ON booking_requests IS 'Valid booking request types';

-- Create view for overdue invoices
CREATE OR REPLACE VIEW customer_overdue_invoices AS
SELECT 
    i.id,
    i.customer_id,
    c.name as customer_name,
    c.email as customer_email,
    i.invoice_number,
    i.amount,
    i.currency,
    i.issue_date,
    i.due_date,
    CURRENT_DATE - i.due_date as days_overdue,
    i.status
FROM invoices i
JOIN customers c ON i.customer_id = c.id
WHERE i.status = 'overdue' OR (i.status = 'issued' AND i.due_date < CURRENT_DATE)
ORDER BY days_overdue DESC;

-- Create view for shipments in transit
CREATE OR REPLACE VIEW shipments_in_transit AS
SELECT 
    s.id,
    s.tracking_number,
    s.customer_id,
    c.name as customer_name,
    s.status,
    s.origin,
    s.destination,
    s.current_location,
    s.estimated_delivery,
    EXTRACT(EPOCH FROM (s.estimated_delivery - NOW())) / 3600 as hours_until_delivery
FROM shipments s
JOIN customers c ON s.customer_id = c.id
WHERE s.status = 'in_transit'
ORDER BY s.estimated_delivery ASC;

-- Create view for pending bookings
CREATE OR REPLACE VIEW pending_bookings AS
SELECT 
    br.id,
    br.customer_id,
    c.name as customer_name,
    c.email as customer_email,
    br.request_type,
    br.status,
    br.requested_pickup_time,
    br.created_at,
    EXTRACT(EPOCH FROM (NOW() - br.created_at)) / 3600 as hours_since_request
FROM booking_requests br
JOIN customers c ON br.customer_id = c.id
WHERE br.status = 'pending'
ORDER BY br.created_at ASC;
