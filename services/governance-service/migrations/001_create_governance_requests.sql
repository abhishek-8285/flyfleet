-- Create governance request type enum
DO $$ BEGIN
    CREATE TYPE governance_request_type AS ENUM ('EXPORT', 'DELETION', 'ACCESS_REVIEW');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create governance request status enum
DO $$ BEGIN
    CREATE TYPE governance_request_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED', 'COMPLETED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create governance requests table
CREATE TABLE IF NOT EXISTS governance_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    requester_id UUID NOT NULL,
    request_type governance_request_type NOT NULL,
    data_scope TEXT NOT NULL,
    purpose TEXT NOT NULL,
    format TEXT NOT NULL DEFAULT 'CSV',
    delivery_method TEXT NOT NULL DEFAULT 'DIRECT',
    status governance_request_status NOT NULL DEFAULT 'PENDING',
    expiry_date TIMESTAMP WITH TIME ZONE,
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexing for performance
CREATE INDEX IF NOT EXISTS idx_gov_req_org ON governance_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_gov_req_status ON governance_requests(status);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_governance_requests_updated_at
    BEFORE UPDATE ON governance_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
