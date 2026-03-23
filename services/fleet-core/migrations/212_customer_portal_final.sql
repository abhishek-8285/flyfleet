-- 212_customer_portal_final.sql
-- Migration for B2B Customer Portal Final Enterprise Features (Phase 3)

-- 1. Booking Approval Workflow
CREATE TABLE IF NOT EXISTS customer_v2_booking_status_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES customer_v2_booking_requests(id) ON DELETE CASCADE,
    previous_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    changed_by UUID, -- References customer_users or system UUID
    changed_at TIMESTAMPTZ DEFAULT NOW(),
    reason TEXT
);

CREATE TABLE IF NOT EXISTS customer_v2_booking_approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES customer_v2_booking_requests(id) ON DELETE CASCADE,
    approver_user_id UUID NOT NULL REFERENCES customer_v2_customer_users(id),
    approval_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    approved_at TIMESTAMPTZ,
    comments TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Multi-Shipment Bookings
CREATE TABLE IF NOT EXISTS customer_v2_booking_shipments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES customer_v2_booking_requests(id) ON DELETE CASCADE,
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    sequence_number INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(booking_id, shipment_id)
);

-- 3. Dispute Management
CREATE TABLE IF NOT EXISTS customer_v2_disputes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customer_v2_customers(id) ON DELETE CASCADE,
    shipment_id UUID REFERENCES customer_v2_shipments(id),
    invoice_id UUID, -- Virtual reference to billing_v2
    status VARCHAR(50) DEFAULT 'OPEN', -- OPEN, INVESTIGATING, RESOLVED, CLOSED
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS customer_v2_dispute_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dispute_id UUID NOT NULL REFERENCES customer_v2_disputes(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL, -- Either customer_user_id or system admin_id
    sender_type VARCHAR(50) NOT NULL, -- CUSTOMER, ADMIN
    message TEXT NOT NULL,
    sent_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_dispute_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dispute_id UUID NOT NULL REFERENCES customer_v2_disputes(id) ON DELETE CASCADE,
    message_id UUID REFERENCES customer_v2_dispute_messages(id),
    file_url VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

-- Note: Contract Pricing, SLA logic, Credit Limits, Delivery Verification and API Tokens
-- map cleanly to customer_v2_pricing_contracts, customer_v2_shipment_sla_rules, 
-- customer_v2_customer_wallets, customer_v2_delivery_proofs, customer_v2_pod_images,
-- and customer_v2_customer_api_keys.
