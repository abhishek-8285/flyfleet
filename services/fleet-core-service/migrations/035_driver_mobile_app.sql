-- Driver Mobile App Schema

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Duty Status Logs (HOOS)
CREATE TABLE IF NOT EXISTS duty_status_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL, -- references drivers(id) - loose coupling if drivers table in another service, but likely same DB
    status VARCHAR(50) NOT NULL CHECK (status IN ('OFF_DUTY', 'SLEEPER_BERTH', 'DRIVING', 'ON_DUTY')),
    location JSONB, -- {latitude, longitude, address}
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER,
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_duty_status_logs_driver_id ON duty_status_logs(driver_id);
CREATE INDEX idx_duty_status_logs_started_at ON duty_status_logs(started_at);

-- 2. Driver Messages
CREATE TABLE IF NOT EXISTS driver_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL,
    trip_id UUID, -- optional link to trip
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('INBOUND', 'OUTBOUND')), -- INBOUND = from driver to dispatch
    type VARCHAR(20) NOT NULL DEFAULT 'TEXT' CHECK (type IN ('TEXT', 'AUDIO', 'IMAGE')),
    content TEXT,
    attachment_url TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_driver_messages_driver_id ON driver_messages(driver_id);
CREATE INDEX idx_driver_messages_trip_id ON driver_messages(trip_id);

-- 3. Driver Expenses
CREATE TABLE IF NOT EXISTS driver_expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL,
    trip_id UUID,
    type VARCHAR(50) NOT NULL, -- FUEL, TOLL, MAINTENANCE, FOOD, LODGING, OTHER
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    receipt_url TEXT,
    notes TEXT,
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_driver_expenses_driver_id ON driver_expenses(driver_id);
CREATE INDEX idx_driver_expenses_trip_id ON driver_expenses(trip_id);

-- 4. Proof of Delivery (POD)
CREATE TABLE IF NOT EXISTS proof_of_delivery (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID NOT NULL REFERENCES trips(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('SIGNATURE', 'PHOTO', 'DOCUMENT', 'QR_CODE')),
    url TEXT NOT NULL,
    metadata JSONB, -- {recipient_name, location, timestamp}
    uploaded_by UUID NOT NULL, -- driver_id
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_pod_trip_id ON proof_of_delivery(trip_id);

-- 5. Trip Offers (Assignment)
CREATE TABLE IF NOT EXISTS trip_offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID NOT NULL REFERENCES trips(id),
    driver_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'OFFERED' CHECK (status IN ('OFFERED', 'ACCEPTED', 'REJECTED', 'EXPIRED')),
    rejection_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    responded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_trip_offers_driver_id ON trip_offers(driver_id);
CREATE INDEX idx_trip_offers_trip_id ON trip_offers(trip_id);
CREATE INDEX idx_trip_offers_status ON trip_offers(status);

-- Add last_known_location to drivers table if not exists (assuming drivers table exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'last_known_location') THEN
        ALTER TABLE drivers ADD COLUMN last_known_location JSONB;
        ALTER TABLE drivers ADD COLUMN last_location_update TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;
