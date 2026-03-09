-- Migration: 058_compliance_dot_inspection
-- Description: Add DOT inspection mode and compliance review features

-- Create DOT inspection table
CREATE TABLE IF NOT EXISTS dot_inspections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    
    -- Inspection details
    inspection_date TIMESTAMPTZ NOT NULL,
    inspection_level INTEGER NOT NULL, -- 1-6 (FMCSA inspection levels)
    inspection_location TEXT,
    inspector_name TEXT,
    inspector_agency TEXT,
    inspector_id VARCHAR(50),
    
    -- Results
    overall_status VARCHAR(20) DEFAULT 'none', -- none, warning, out_of_service
    violations_found INTEGER DEFAULT 0,
    oos_violations INTEGER DEFAULT 0,
    violation_details JSONB,
    
    -- FMCSA requirements
    was_eligible_for_time_credit BOOLEAN DEFAULT false,
    time_credit_minutes INTEGER DEFAULT 0,
    
    -- Documents provided
    roc_provided BOOLEAN DEFAULT false,
    medical_certificate_provided BOOLEAN DEFAULT false,
    cdl_provided BOOLEAN DEFAULT false,
    skill_performance_eval_provided BOOLEAN DEFAULT false,
    
    -- Outcome
    vehicle_oos BOOLEAN DEFAULT false,
    driver_oos BOOLEAN DEFAULT false,
    fine_amount DECIMAL(10, 2) DEFAULT 0,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    
    CONSTRAINT chk_inspection_level CHECK (inspection_level BETWEEN 1 AND 6),
    CONSTRAINT chk_overall_status CHECK (overall_status IN ('none', 'warning', 'out_of_service'))
);

-- Create inspection document table
CREATE TABLE IF NOT EXISTS inspection_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inspection_id UUID NOT NULL REFERENCES dot_inspections(id) ON DELETE CASCADE,
    
    document_type VARCHAR(50) NOT NULL, -- ROC, BILL_OF_LADING, MEDICAL_CERT, etc.
    document_url TEXT,
    document_data BYTEA, -- For storing PDF directly
    
    provided_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_document_type CHECK (document_type IN ('ROC', 'BILL_OF_LADING', 'MEDICAL_CERT', 'CDL', 'SPE', 'VEHICLE_REG', 'INSURANCE'))
);

-- Create compliance review table
CREATE TABLE IF NOT EXISTS compliance_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    
    -- Review details
    review_type VARCHAR(50) NOT NULL, -- compliance_review, focused_review, etc.
    review_date TIMESTAMPTZ NOT NULL,
    review_location TEXT,
    fmcsa_representative TEXT,
    
    -- Results
    overall_rating VARCHAR(20), -- satisfactory, conditional, unsatisfactory
    acute_violations INTEGER DEFAULT 0,
    serious_violations INTEGER DEFAULT 0,
    other_violations INTEGER DEFAULT 0,
    
    -- BASIC scores
    unsafe_driving_score DECIMAL(5, 2),
    hos_compliance_score DECIMAL(5, 2),
    driver_fitness_score DECIMAL(5, 2),
    controlled_substances_score DECIMAL(5, 2),
    vehicle_maintenance_score DECIMAL(5, 2),
    hazmat_score DECIMAL(5, 2),
    crash_indicator_score DECIMAL(5, 2),
    
    -- Outcome
    corrective_action_required BOOLEAN DEFAULT false,
    corrective_action_plan TEXT,
    follow_up_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    
    CONSTRAINT chk_review_type CHECK (review_type IN ('compliance_review', 'focused_review', 'offsite_review')),
    CONSTRAINT chk_overall_rating CHECK (overall_rating IN ('satisfactory', 'conditional', 'unsatisfactory'))
);

-- Create DOT inspection mode settings
CREATE TABLE IF NOT EXISTS dot_inspection_mode (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL UNIQUE,
    
    -- Mode settings
    is_enabled BOOLEAN DEFAULT false,
    active_inspection_id UUID REFERENCES dot_inspections(id),
    
    -- Quick access documents
    default_roc_template UUID,
    default_violation_codes JSONB,
    
    -- Contact info
    safety_director_name TEXT,
    safety_director_phone TEXT,
    safety_director_email TEXT,
    
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- Indexes
CREATE INDEX idx_dot_inspections_driver ON dot_inspections(driver_id);
CREATE INDEX idx_dot_inspections_vehicle ON dot_inspections(vehicle_id);
CREATE INDEX idx_dot_inspections_date ON dot_inspections(inspection_date);
CREATE INDEX idx_dot_inspections_status ON dot_inspections(overall_status);
CREATE INDEX idx_compliance_reviews_org ON compliance_reviews(org_id);
CREATE INDEX idx_compliance_reviews_date ON compliance_reviews(review_date);

-- Comments
COMMENT ON TABLE dot_inspections IS 'DOT/FMCSA roadside inspection records';
COMMENT ON TABLE inspection_documents IS 'Documents provided during DOT inspections';
COMMENT ON TABLE compliance_reviews IS 'FMCSA compliance review results';
COMMENT ON COLUMN dot_inspections.inspection_level IS 'FMCSA inspection level (1-6): 1=Full, 2=Walk-Around, 3=Driver-Only, 4=Special, 5=Crash, 6=Enhanced';
COMMENT ON COLUMN dot_inspections.overall_status IS 'Inspection result: none, warning, or out_of_service';

-- Function to get inspection history for driver
CREATE OR REPLACE FUNCTION get_driver_inspection_history(
    p_driver_id UUID,
    p_months INTEGER DEFAULT 12
) RETURNS TABLE(
    inspection_id UUID,
    inspection_date TIMESTAMPTZ,
    inspection_level INTEGER,
    location TEXT,
    overall_status VARCHAR,
    violations_found INTEGER,
    oos_violations INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        di.id,
        di.inspection_date,
        di.inspection_level,
        di.inspection_location,
        di.overall_status,
        di.violations_found,
        di.oos_violations
    FROM dot_inspections di
    WHERE di.driver_id = p_driver_id
      AND di.inspection_date >= CURRENT_TIMESTAMP - (p_months || ' months')::INTERVAL
    ORDER BY di.inspection_date DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to get organization compliance summary
CREATE OR REPLACE FUNCTION get_org_compliance_summary(
    p_org_id UUID
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
        (SELECT MAX(review_date) FROM compliance_reviews WHERE org_id = p_org_id) as last_review_date,
        (SELECT overall_rating FROM compliance_reviews WHERE org_id = p_org_id ORDER BY review_date DESC LIMIT 1) as last_review_rating
    FROM dot_inspections di
    JOIN drivers d ON di.driver_id = d.id
    WHERE d.org_id = p_org_id;
END;
$$ LANGUAGE plpgsql;
