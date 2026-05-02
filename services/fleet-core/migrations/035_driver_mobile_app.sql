-- Migration: 035_driver_mobile_app
-- Description: Driver Mobile App Schema with tenant isolation.

-- 1. Duty Status Logs
CREATE TABLE IF NOT EXISTS duty_status_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('OFF_DUTY', 'SLEEPER_BERTH', 'DRIVING', 'ON_DUTY')),
    location JSONB,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER,
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_duty_status_logs_tenant ON duty_status_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_duty_status_logs_driver ON duty_status_logs(driver_id);

-- 2. Driver Messages
CREATE TABLE IF NOT EXISTS driver_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    trip_id UUID,
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('INBOUND', 'OUTBOUND')),
    type VARCHAR(20) NOT NULL DEFAULT 'TEXT' CHECK (type IN ('TEXT', 'AUDIO', 'IMAGE')),
    content TEXT,
    attachment_url TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_driver_messages_tenant ON driver_messages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_driver_messages_driver ON driver_messages(driver_id);

-- 3. Driver Expenses
CREATE TABLE IF NOT EXISTS driver_expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    trip_id UUID,
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    receipt_url TEXT,
    notes TEXT,
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_driver_expenses_tenant ON driver_expenses(tenant_id);

-- 4. Proof of Delivery
CREATE TABLE IF NOT EXISTS proof_of_delivery (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    trip_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('SIGNATURE', 'PHOTO', 'DOCUMENT', 'QR_CODE')),
    url TEXT NOT NULL,
    metadata JSONB,
    uploaded_by UUID NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pod_tenant ON proof_of_delivery(tenant_id);

-- 5. Trip Offers
CREATE TABLE IF NOT EXISTS trip_offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    trip_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'OFFERED' CHECK (status IN ('OFFERED', 'ACCEPTED', 'REJECTED', 'EXPIRED')),
    rejection_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    responded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trip_offers_tenant ON trip_offers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_trip_offers_status ON trip_offers(status);

-- 6. Self-healing: Update drivers table if exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'drivers') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'tenant_id') THEN
            ALTER TABLE drivers ADD COLUMN tenant_id UUID REFERENCES tenants(id);
        END IF;
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'last_known_location') THEN
            ALTER TABLE drivers ADD COLUMN last_known_location JSONB;
            ALTER TABLE drivers ADD COLUMN last_location_update TIMESTAMP WITH TIME ZONE;
        END IF;
    END IF;
END $$;
