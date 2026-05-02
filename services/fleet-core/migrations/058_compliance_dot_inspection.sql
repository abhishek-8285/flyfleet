-- Migration: 058_compliance_dot_inspection
-- Description: Add DOT inspection mode and compliance review features with tenant isolation.

-- 1. Create DOT inspection table
CREATE TABLE IF NOT EXISTS dot_inspections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    inspection_date TIMESTAMPTZ NOT NULL,
    inspection_level INTEGER NOT NULL,
    inspection_location TEXT,
    inspector_name TEXT,
    inspector_agency TEXT,
    inspector_id VARCHAR(50),
    overall_status VARCHAR(20) DEFAULT 'none',
    violations_found INTEGER DEFAULT 0,
    oos_violations INTEGER DEFAULT 0,
    violation_details JSONB,
    was_eligible_for_time_credit BOOLEAN DEFAULT false,
    time_credit_minutes INTEGER DEFAULT 0,
    roc_provided BOOLEAN DEFAULT false,
    medical_certificate_provided BOOLEAN DEFAULT false,
    cdl_provided BOOLEAN DEFAULT false,
    skill_performance_eval_provided BOOLEAN DEFAULT false,
    vehicle_oos BOOLEAN DEFAULT false,
    driver_oos BOOLEAN DEFAULT false,
    fine_amount DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    CONSTRAINT chk_inspection_level CHECK (inspection_level BETWEEN 1 AND 6)
);

-- 2. Create inspection document table
CREATE TABLE IF NOT EXISTS inspection_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inspection_id UUID NOT NULL REFERENCES dot_inspections(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    document_url TEXT,
    document_data BYTEA,
    provided_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create compliance review table
CREATE TABLE IF NOT EXISTS compliance_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    review_type VARCHAR(50) NOT NULL,
    review_date TIMESTAMPTZ NOT NULL,
    review_location TEXT,
    fmcsa_representative TEXT,
    overall_rating VARCHAR(20),
    acute_violations INTEGER DEFAULT 0,
    serious_violations INTEGER DEFAULT 0,
    other_violations INTEGER DEFAULT 0,
    unsafe_driving_score DECIMAL(5, 2),
    hos_compliance_score DECIMAL(5, 2),
    driver_fitness_score DECIMAL(5, 2),
    controlled_substances_score DECIMAL(5, 2),
    vehicle_maintenance_score DECIMAL(5, 2),
    hazmat_score DECIMAL(5, 2),
    crash_indicator_score DECIMAL(5, 2),
    corrective_action_required BOOLEAN DEFAULT false,
    corrective_action_plan TEXT,
    follow_up_date DATE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

-- 4. Create DOT inspection mode settings
CREATE TABLE IF NOT EXISTS dot_inspection_mode (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
    is_enabled BOOLEAN DEFAULT false,
    active_inspection_id UUID REFERENCES dot_inspections(id),
    default_roc_template UUID,
    default_violation_codes JSONB,
    safety_director_name TEXT,
    safety_director_phone TEXT,
    safety_director_email TEXT,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- 5. Indexes
CREATE INDEX IF NOT EXISTS idx_dot_inspections_tenant ON dot_inspections(tenant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_reviews_tenant ON compliance_reviews(tenant_id);

-- 6. Helper functions
CREATE OR REPLACE FUNCTION get_tenant_compliance_summary(
    p_tenant_id UUID
) RETURNS TABLE(
    total_inspections BIGINT,
    inspections_with_violations BIGINT,
    oos_inspections BIGINT,
    oos_rate DECIMAL,
    avg_violations_per_inspection DECIMAL,
    last_review_date TIMESTAMPTZ,
    last_review_rating VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(di.id) as total_inspections,
        COUNT(CASE WHEN di.violations_found > 0 THEN 1 END) as inspections_with_violations,
        COUNT(CASE WHEN di.overall_status = 'out_of_service' THEN 1 END) as oos_inspections,
        CASE 
            WHEN COUNT(di.id) > 0 THEN 
                ROUND(COUNT(CASE WHEN di.overall_status = 'out_of_service' THEN 1 END)::DECIMAL / COUNT(di.id)::DECIMAL * 100, 2)
            ELSE 0 
        END as oos_rate,
        CASE 
            WHEN COUNT(di.id) > 0 THEN 
                ROUND(AVG(di.violations_found)::DECIMAL, 2)
            ELSE 0 
        END as avg_violations_per_inspection,
        (SELECT MAX(review_date) FROM compliance_reviews WHERE tenant_id = p_tenant_id) as last_review_date,
        (SELECT overall_rating FROM compliance_reviews WHERE tenant_id = p_tenant_id ORDER BY review_date DESC LIMIT 1) as last_review_rating
    FROM dot_inspections di
    WHERE di.tenant_id = p_tenant_id;
END;
$$ LANGUAGE plpgsql;
