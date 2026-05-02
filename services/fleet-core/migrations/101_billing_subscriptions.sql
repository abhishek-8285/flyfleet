-- Migration: 101_billing_subscriptions
-- Description: Seed pricing components for harmonized billing plans.

DO $$
DECLARE
    starter_id UUID := '11111111-1111-1111-1111-111111111111';
    growth_id UUID  := '22222222-2222-2222-2222-222222222222';
    enterprise_id UUID := '33333333-3333-3333-3333-333333333333';
BEGIN
    -- Starter Components
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (starter_id, 'Platform Fee', 'FLAT', 49.00, 'MONTHLY')
    ON CONFLICT DO NOTHING;

    -- Growth Components
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (growth_id, 'Base Platform Fee', 'FLAT', 199.00, 'MONTHLY')
    ON CONFLICT DO NOTHING;
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (growth_id, 'Vehicle Fee', 'PER_UNIT', 15.00, 'MONTHLY')
    ON CONFLICT DO NOTHING;

    -- Enterprise Components
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (enterprise_id, 'Custom Platform Fee', 'FLAT', 999.00, 'MONTHLY')
    ON CONFLICT DO NOTHING;
END $$;
