-- Create documents table for vehicle compliance
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'registration', 'insurance', 'permit', etc.
    doc_number VARCHAR(100) NOT NULL,
    issue_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    file_url TEXT, -- Path to file in storage or S3 URL
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'verified', 'expired', 'rejected'
    metadata JSONB DEFAULT '{}'::jsonb, -- Store additional attributes like issuing authority, coverage amount, etc.
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_documents_vehicle_id ON documents(vehicle_id);
CREATE INDEX idx_documents_expiry_date ON documents(expiry_date);
CREATE INDEX idx_documents_type ON documents(type);
CREATE INDEX idx_documents_status ON documents(status);

-- Comments for documentation
COMMENT ON TABLE documents IS 'Stores compliance documents associated with vehicles';
COMMENT ON COLUMN documents.metadata IS 'Flexible JSON storage for document-specific attributes';
