-- Migration: 034_maintenance_system_expansion.sql
-- Description: End-to-end maintenance lifecycle management expansion

-- 1. Create Maintenance Vendors Table
CREATE TABLE IF NOT EXISTS maintenance_vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    specialization VARCHAR(100), -- Engine, Body, Electronics, Tires, General
    rating DECIMAL(3, 2) DEFAULT 0.00 CHECK (rating >= 0 AND rating <= 5),
    contact_info JSONB DEFAULT '{}'::jsonb, -- Phone, Email, Address
    payment_terms VARCHAR(100), -- Net-30, Net-60, Prepaid
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, BLACKLISTED
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_vendor_per_org UNIQUE (organization_id, name)
);

-- 2. Create Inspection Templates Table
CREATE TABLE IF NOT EXISTS inspection_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    applicable_vehicle_types VARCHAR(50)[], -- ['TRUCK', 'VAN', 'CAR']
    checklist_items JSONB NOT NULL DEFAULT '[]'::jsonb, -- Array of {item_id, text, type, required}
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create Vehicle Inspections Table
CREATE TABLE IF NOT EXISTS vehicle_inspections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    template_id UUID REFERENCES inspection_templates(id),
    inspector_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'COMPLETED', -- DRAFT, COMPLETED, FAILED
    results JSONB NOT NULL DEFAULT '[]'::jsonb, -- Array of {item_id, status, value, photo_urls[], comments}
    notes TEXT,
    odometer_at_inspection BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Alter Work Orders
DO $$ 
BEGIN 
    -- Add vendor_id
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_orders' AND column_name='vendor_id') THEN
        ALTER TABLE work_orders ADD COLUMN vendor_id UUID REFERENCES maintenance_vendors(id);
    END IF;

    -- Add approval_status
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_orders' AND column_name='approval_status') THEN
        ALTER TABLE work_orders ADD COLUMN approval_status VARCHAR(50) DEFAULT 'NOT_REQUIRED';
        -- Possible values: NOT_REQUIRED, PENDING_MANAGER, PENDING_FLEET_HEAD, PENDING_FINANCE, APPROVED, REJECTED
    END IF;

    -- Add cost_breakdown
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_orders' AND column_name='cost_breakdown') THEN
        ALTER TABLE work_orders ADD COLUMN cost_breakdown JSONB DEFAULT '{"labor_cost": 0, "parts_cost": 0, "external_service_cost": 0}'::jsonb;
    END IF;

    -- Add inspection_id (link to inspection that triggered this WO if any)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_orders' AND column_name='inspection_id') THEN
        ALTER TABLE work_orders ADD COLUMN inspection_id UUID REFERENCES vehicle_inspections(id);
    END IF;
END $$;

-- 5. Alter Inventory Parts
DO $$ 
BEGIN 
    -- Add category
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='inventory_parts' AND column_name='category') THEN
        ALTER TABLE inventory_parts ADD COLUMN category VARCHAR(100);
    END IF;

    -- Add compatible_vehicles array
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='inventory_parts' AND column_name='compatible_vehicles') THEN
        ALTER TABLE inventory_parts ADD COLUMN compatible_vehicles VARCHAR(100)[];
    END IF;
END $$;

-- 6. Indexes
CREATE INDEX IF NOT EXISTS idx_maintenance_vendors_org ON maintenance_vendors(organization_id);
CREATE INDEX IF NOT EXISTS idx_inspection_templates_org ON inspection_templates(organization_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_inspections_vehicle ON vehicle_inspections(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_vendor ON work_orders(vendor_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_approval ON work_orders(approval_status);

-- 7. Triggers for updated_at
CREATE TRIGGER trigger_maintenance_vendors_updated_at
    BEFORE UPDATE ON maintenance_vendors
    FOR EACH ROW
    EXECUTE FUNCTION update_work_orders_updated_at();

CREATE TRIGGER trigger_inspection_templates_updated_at
    BEFORE UPDATE ON inspection_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_work_orders_updated_at();

CREATE TRIGGER trigger_vehicle_inspections_updated_at
    BEFORE UPDATE ON vehicle_inspections
    FOR EACH ROW
    EXECUTE FUNCTION update_work_orders_updated_at();

-- 8. Seed Default Inspection Template
INSERT INTO inspection_templates (organization_id, name, description, applicable_vehicle_types, checklist_items)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Standard Safety Inspection',
    'Routine safety check for all fleet vehicles',
    ARRAY['TRUCK', 'VAN', 'CAR'],
    '[
        {"item_id": "brakes", "text": "Brake pad thickness & fluid level", "type": "pass_fail", "required": true},
        {"item_id": "tires", "text": "Tire pressure & tread depth", "type": "pass_fail", "required": true},
        {"item_id": "lights", "text": "Headlights, tail lights & signals", "type": "pass_fail", "required": true},
        {"item_id": "fluids", "text": "Oil, coolant & washer fluid", "type": "pass_fail", "required": false},
        {"item_id": "engine", "text": "Engine noise & diagnostic codes", "type": "pass_fail", "required": true}
    ]'::jsonb
) ON CONFLICT DO NOTHING;

COMMENT ON TABLE maintenance_vendors IS 'Service providers and part suppliers';
COMMENT ON TABLE inspection_templates IS 'Structure for digital inspection checklists';
COMMENT ON TABLE vehicle_inspections IS 'Completed vehicle inspection reports';
