-- Migration: 077_india_compliance_view
-- Description: Real-time compliance monitoring view with tenant isolation.

CREATE OR REPLACE VIEW india_compliance_summary AS
WITH permit_counts AS (
    SELECT tenant_id, COUNT(*) as expiring_permits
    FROM local_permits
    WHERE valid_to <= CURRENT_DATE + INTERVAL '30 days' AND status = 'active'
    GROUP BY tenant_id
),
license_counts AS (
    SELECT tenant_id, COUNT(*) as expiring_licenses
    FROM driver_verifications
    WHERE validity_to <= CURRENT_DATE + INTERVAL '30 days' AND status = 'verified'
    GROUP BY tenant_id
),
insurance_puc_counts AS (
    SELECT tenant_id, COUNT(*) as expiring_insurance
    FROM vehicle_verifications
    WHERE (insurance_expiry <= CURRENT_DATE + INTERVAL '30 days' OR puc_expiry <= CURRENT_DATE + INTERVAL '30 days')
      AND status = 'verified'
    GROUP BY tenant_id
),
gst_non_compliant AS (
    SELECT tenant_id, COUNT(*) as gst_non_compliant_orgs
    FROM gst_compliances
    WHERE status IN ('cancelled', 'suspended')
    GROUP BY tenant_id
),
all_tenants AS (
    SELECT id as tenant_id FROM tenants
)
SELECT 
    o.tenant_id,
    COALESCE(p.expiring_permits, 0) as expiring_permits,
    COALESCE(l.expiring_licenses, 0) as expiring_licenses,
    COALESCE(i.expiring_insurance, 0) as expiring_insurance,
    COALESCE(g.gst_non_compliant_orgs, 0) as gst_non_compliant_orgs
FROM all_tenants o
LEFT JOIN permit_counts p ON o.tenant_id = p.tenant_id
LEFT JOIN license_counts l ON o.tenant_id = l.tenant_id
LEFT JOIN insurance_puc_counts i ON o.tenant_id = i.tenant_id
LEFT JOIN gst_non_compliant g ON o.tenant_id = g.tenant_id;
