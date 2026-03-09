-- 211_customer_portal_extensions.sql
-- Migration for B2B Customer Portal Enterprise Extensions (customer_v2)

-- 1. E-Way Bill Support on Shipments
ALTER TABLE customer_v2_shipments ADD COLUMN IF NOT EXISTS eway_bill_number VARCHAR(100);
ALTER TABLE customer_v2_shipments ADD COLUMN IF NOT EXISTS eway_bill_expiry TIMESTAMPTZ;
ALTER TABLE customer_v2_shipments ADD COLUMN IF NOT EXISTS eway_bill_document VARCHAR(255);

-- 2. POD Images and Signatures
CREATE TABLE IF NOT EXISTS customer_v2_pod_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delivery_proof_id UUID NOT NULL REFERENCES customer_v2_delivery_proofs(id) ON DELETE CASCADE,
    photo_url VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_v2_pod_signatures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delivery_proof_id UUID NOT NULL REFERENCES customer_v2_delivery_proofs(id) ON DELETE CASCADE,
    signature_url VARCHAR(255) NOT NULL,
    signed_by VARCHAR(255) NOT NULL,
    signed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Note: customer_v2_delivery_proofs already has signature_url and photo_url, but enterprise
-- requirements dictate multiple images/signatures mapped via these new explicit tables.
