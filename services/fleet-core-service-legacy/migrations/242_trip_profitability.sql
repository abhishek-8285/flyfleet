-- Migration: 242_trip_profitability
-- Description: Trip profitability tracking and finance dashboard aggregation

CREATE TABLE IF NOT EXISTS trip_profitability (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trip_id UUID NOT NULL UNIQUE,
    org_id UUID NOT NULL,
    vehicle_id UUID,
    driver_id UUID,
    trip_revenue DECIMAL(15, 2) DEFAULT 0,
    fuel_cost DECIMAL(15, 2) DEFAULT 0,
    toll_cost DECIMAL(15, 2) DEFAULT 0,
    driver_cost DECIMAL(15, 2) DEFAULT 0,
    maintenance_cost DECIMAL(15, 2) DEFAULT 0,
    other_cost DECIMAL(15, 2) DEFAULT 0,
    total_cost DECIMAL(15, 2) GENERATED ALWAYS AS (fuel_cost + toll_cost + driver_cost + maintenance_cost + other_cost) STORED,
    profit DECIMAL(15, 2) GENERATED ALWAYS AS (trip_revenue - (fuel_cost + toll_cost + driver_cost + maintenance_cost + other_cost)) STORED,
    profit_margin DECIMAL(5, 2),                    -- percentage, calculated by service
    distance_km DECIMAL(10, 2),
    cost_per_km DECIMAL(10, 4),
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_trip_profit_trip ON trip_profitability(trip_id);
CREATE INDEX idx_trip_profit_org ON trip_profitability(org_id);
CREATE INDEX idx_trip_profit_vehicle ON trip_profitability(vehicle_id) WHERE vehicle_id IS NOT NULL;

-- Finance dashboard metrics materialized view
CREATE MATERIALIZED VIEW IF NOT EXISTS finance_dashboard_metrics AS
SELECT
    o.org_id,
    DATE_TRUNC('month', NOW()) AS metric_month,
    COALESCE(SUM(tp.trip_revenue), 0) AS total_revenue,
    COALESCE(SUM(tp.fuel_cost), 0) AS total_fuel_cost,
    COALESCE(SUM(er.total_amount), 0) AS total_expenses,
    CASE WHEN SUM(tp.trip_revenue) > 0
         THEN ROUND(((SUM(tp.trip_revenue) - SUM(tp.total_cost)) / SUM(tp.trip_revenue)) * 100, 2)
         ELSE 0
    END AS profit_margin,
    CASE WHEN SUM(tp.distance_km) > 0
         THEN ROUND(SUM(tp.total_cost) / SUM(tp.distance_km), 4)
         ELSE 0
    END AS avg_cost_per_km
FROM (SELECT DISTINCT org_id FROM trip_profitability) o
LEFT JOIN trip_profitability tp ON tp.org_id = o.org_id
LEFT JOIN expense_reports er ON (
    er.metadata->>'org_id' = o.org_id::text
    OR er.org_id::text = o.org_id::text
)
GROUP BY o.org_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_finance_dashboard_metrics_org ON finance_dashboard_metrics(org_id);

COMMENT ON MATERIALIZED VIEW finance_dashboard_metrics IS 'Aggregated finance metrics per org, refreshed by FinanceMetricsWorker';
COMMENT ON TABLE trip_profitability IS 'Pre-calculated profitability metrics per trip';
