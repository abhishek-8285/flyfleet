-- Migration: 210b_customer_v2_delivery_proofs
-- Description: Create the missing customer_v2_delivery_proofs table required by extensions.

CREATE TABLE IF NOT EXISTS customer_v2_delivery_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    shipment_id UUID NOT NULL REFERENCES customer_v2_shipments(id) ON DELETE CASCADE,
    receiver_name VARCHAR(255) NOT NULL,
    receiver_phone VARCHAR(50),
    signature_url VARCHAR(255),
    photo_url VARCHAR(255),
    delivered_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cust_v2_proof_tenant ON customer_v2_delivery_proofs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cust_v2_proof_shipment ON customer_v2_delivery_proofs(shipment_id);
