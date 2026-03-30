-- FlyFleet Analytics Schema for ClickHouse
-- Run this once to set up analytics tables

CREATE DATABASE IF NOT EXISTS flyfleet_analytics;
USE flyfleet_analytics;

-- Trip Analytics (aggregated by hour)
CREATE TABLE IF NOT EXISTS trip_analytics_hourly
(
    org_id UUID,
    vehicle_id UUID,
    driver_id UUID,
    timestamp DateTime,
    year UInt16,
    month UInt8,
    day UInt8,
    hour UInt8,
    total_distance_km Float64,
    total_fuel_consumed Float64,
    trip_count UInt32,
    avg_speed_kmh Float64,
    max_speed_kmh Float64,
    idle_time_seconds UInt32,
    events_count UInt32
) ENGINE = MergeTree()
ORDER BY (org_id, vehicle_id, timestamp)
PARTITION BY toYYYYMMDD(timestamp)
SETTINGS index_granularity=8192;

-- Real-time GPS telemetry (last 30 days)
CREATE TABLE IF NOT EXISTS telemetry_realtime
(
    org_id UUID,
    vehicle_id UUID,
    driver_id UUID,
    timestamp DateTime,
    latitude Float64,
    longitude Float64,
    speed_kmh Float64,
    heading Float64,
    altitude Float64,
    fuel_level_percent Float32,
    engine_status Enum8('idle'=1, 'running'=2, 'off'=3),
    event_type String
) ENGINE = MergeTree()
ORDER BY (org_id, vehicle_id, timestamp)
PARTITION BY toYYYYMMDD(timestamp)
TTL timestamp + INTERVAL 30 DAY;

-- Invoice Analytics
CREATE TABLE IF NOT EXISTS invoice_analytics
(
    org_id UUID,
    invoice_id UUID,
    invoice_number String,
    customer_id UUID,
    invoice_date DateTime,
    due_date DateTime,
    status Enum8('draft'=1, 'pending'=2, 'paid'=3, 'overdue'=4, 'partial'=5, 'cancelled'=6),
    currency String,
    subtotal Float64,
    tax_amount Float64,
    total_amount Float64,
    amount_paid Float64,
    balance_due Float64,
    created_at DateTime,
    paid_at DateTime,
    days_to_pay Int16
) ENGINE = MergeTree()
ORDER BY (org_id, invoice_date)
PARTITION BY toYYYYMM(invoice_date);

-- Payment Analytics
CREATE TABLE IF NOT EXISTS payment_analytics
(
    org_id UUID,
    payment_id UUID,
    invoice_id UUID,
    amount Float64,
    payment_method String,
    status Enum8('pending'=1, 'succeeded'=2, 'failed'=3, 'refunded'=4),
    gateway String,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (org_id, created_at)
PARTITION BY toYYYYMM(created_at);

-- Daily Revenue Summary (Materialized View)
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_revenue_summary
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (org_id, date)
AS
SELECT
    org_id,
    toDate(invoice_date) AS date,
    countState() AS invoice_count,
    sumState(total_amount) AS total_revenue,
    sumState(tax_amount) AS total_tax,
    avgState(balance_due) AS avg_balance_due,
    countIfState(status = 'paid') AS paid_count,
    countIfState(status = 'overdue') AS overdue_count
FROM invoice_analytics
GROUP BY org_id, toDate(invoice_date);

-- Driver Performance Analytics
CREATE TABLE IF NOT EXISTS driver_performance_daily
(
    org_id UUID,
    driver_id UUID,
    date Date,
    total_distance_km Float64,
    total_fuel_consumed Float64,
    avg_speed_kmh Float64,
    max_speed_kmh Float64,
    idle_time_seconds UInt32,
    events_count UInt32,
    trips_completed UInt32
) ENGINE = SummingMergeTree()
ORDER BY (org_id, driver_id, date)
PARTITION BY toYYYYMM(date);
