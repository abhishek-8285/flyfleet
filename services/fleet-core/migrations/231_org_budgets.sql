-- Migration: 231_org_budgets
-- Description: Create org_budgets table for budget tracking

CREATE TABLE IF NOT EXISTS org_budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    category VARCHAR(50) NOT NULL, -- FUEL, MAINTENANCE, GENERAL, TRIP_EXPENSES
    limit_amount DECIMAL(15, 2) NOT NULL,
    spent_amount DECIMAL(15, 2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'INR',
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    alert_threshold_pct INTEGER DEFAULT 80, -- Notify when 80% is spent
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, category, period_start)
);

-- Indexes
CREATE INDEX idx_org_budgets_org ON org_budgets(organization_id);
CREATE INDEX idx_org_budgets_period ON org_budgets(period_start, period_end);

-- Comments
COMMENT ON TABLE org_budgets IS 'Budgets for different expense categories at the organization level';
