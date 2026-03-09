-- migration: 051_billing_advanced_subscriptions.sql
SET search_path TO public;

-- 1. Billing Cycles (Tracks usage and invoicing periods)
CREATE TABLE IF NOT EXISTS billing_cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    subscription_id UUID NOT NULL REFERENCES billing_subscriptions(id),
    cycle_start TIMESTAMP WITH TIME ZONE NOT NULL,
    cycle_end TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, CLOSED, CANCELLED
    closed_at TIMESTAMP WITH TIME ZONE,
    invoice_id UUID REFERENCES billing_invoices(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Coupons (Promotional discount codes)
CREATE TABLE IF NOT EXISTS billing_coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    discount_type VARCHAR(20) NOT NULL, -- PERCENTAGE, FIXED_AMOUNT
    discount_value DECIMAL(15, 4) NOT NULL,
    currency VARCHAR(3), -- Optional, only used for FIXED_AMOUNT
    max_redemptions INT,
    redeemed_count INT NOT NULL DEFAULT 0,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_until TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, EXPIRED, DISABLED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

-- 3. Coupon Redemptions
CREATE TABLE IF NOT EXISTS billing_coupon_redemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_id UUID NOT NULL REFERENCES billing_coupons(id),
    org_id UUID NOT NULL REFERENCES organizations(id),
    subscription_id UUID REFERENCES billing_subscriptions(id),
    invoice_id UUID REFERENCES billing_invoices(id),
    redeemed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RLS Policies For Multi-Tenancy
ALTER TABLE billing_cycles ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS cycles_tenant_policy ON billing_cycles;
CREATE POLICY cycles_tenant_policy ON billing_cycles
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

ALTER TABLE billing_coupon_redemptions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS redemption_tenant_policy ON billing_coupon_redemptions;
CREATE POLICY redemption_tenant_policy ON billing_coupon_redemptions
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Note: billing_coupons does not have RLS as they are system-wide objects. User-facing apps will only query redemptions.

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cycles_sub_id ON billing_cycles(subscription_id);
CREATE INDEX IF NOT EXISTS idx_cycles_org_id ON billing_cycles(org_id);
CREATE INDEX IF NOT EXISTS idx_cycles_active ON billing_cycles(subscription_id, status);
CREATE INDEX IF NOT EXISTS idx_coupons_code ON billing_coupons(code);

-- Permissions
GRANT SELECT, INSERT, UPDATE ON billing_cycles TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE ON billing_coupons TO flyfleet_app;
GRANT SELECT, INSERT ON billing_coupon_redemptions TO flyfleet_app;
