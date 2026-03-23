-- Create india_compliance_summary view for real-time compliance monitoring
CREATE OR REPLACE VIEW india_compliance_summary AS
WITH permit_counts AS (
    SELECT organization_id, COUNT(*) as expiring_permits
    FROM local_permits
    WHERE valid_to <= CURRENT_DATE + INTERVAL '30 days' AND status = 'active'
    GROUP BY organization_id
),
license_counts AS (
    SELECT organization_id, COUNT(*) as expiring_licenses
    FROM driver_verifications
    WHERE validity_to <= CURRENT_DATE + INTERVAL '30 days' AND status = 'verified'
    GROUP BY organization_id
),
insurance_puc_counts AS (
    SELECT organization_id, COUNT(*) as expiring_insurance
    FROM vehicle_verifications
    WHERE (insurance_expiry <= CURRENT_DATE + INTERVAL '30 days' OR puc_expiry <= CURRENT_DATE + INTERVAL '30 days')
      AND status = 'verified'
    GROUP BY organization_id
),
gst_non_compliant AS (
    SELECT organization_id, COUNT(*) as gst_non_compliant_orgs
    FROM gst_compliances
    WHERE status IN ('cancelled', 'suspended')
    GROUP BY organization_id
),
all_orgs AS (
    SELECT DISTINCT organization_id FROM integration_configs
)
SELECT 
    o.organization_id,
    COALESCE(p.expiring_permits, 0) as expiring_permits,
    COALESCE(l.expiring_licenses, 0) as expiring_licenses,
    COALESCE(i.expiring_insurance, 0) as expiring_insurance,
    COALESCE(g.gst_non_compliant_orgs, 0) as gst_non_compliant_orgs
FROM all_orgs o
LEFT JOIN permit_counts p ON o.organization_id = p.organization_id
LEFT JOIN license_counts l ON o.organization_id = l.organization_id
LEFT JOIN insurance_puc_counts i ON o.organization_id = i.organization_id
LEFT JOIN gst_non_compliant g ON o.organization_id = g.organization_id;
