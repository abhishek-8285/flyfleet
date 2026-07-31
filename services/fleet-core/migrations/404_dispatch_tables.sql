-- Dispatch Domain: audit trail and suggestion cache tables
-- These tables support the dispatch orchestration layer (Assign, Redispatch, Dashboard).

-- Dispatch Actions: immutable audit trail of all dispatch decisions
CREATE TABLE dispatch_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    action_type     VARCHAR(30) NOT NULL CHECK (action_type IN (
        'auto_suggest', 'manual_assign', 'auto_assign',
        'bulk_assign', 'redispatch', 'override'
    )),
    shipment_id     UUID,
    trip_id         UUID,
    vehicle_id      UUID,
    driver_id       UUID,
    suggestion_data JSONB,
    decision_reason TEXT,
    performed_by    UUID,
    performed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for dashboard query performance
CREATE INDEX IF NOT EXISTS idx_dispatch_actions_tenant ON dispatch_actions (tenant_id, performed_at DESC);
CREATE INDEX IF NOT EXISTS idx_dispatch_actions_trip ON dispatch_actions (trip_id);
CREATE INDEX IF NOT EXISTS idx_dispatch_actions_shipment ON dispatch_actions (shipment_id);

-- Dispatch Suggestions: short-TTL cache for pre-computed vehicle/driver suggestions
CREATE TABLE dispatch_suggestions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    shipment_id   UUID NOT NULL,
    suggestion_rank INT NOT NULL,
    vehicle_id    UUID,
    driver_id     UUID,
    score         DOUBLE PRECISION,
    distance_km   DOUBLE PRECISION,
    eta_min       INT,
    reasons       TEXT[],
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '5 minutes'
);

CREATE INDEX IF NOT EXISTS idx_dispatch_suggestions_shipment ON dispatch_suggestions (tenant_id, shipment_id);
CREATE INDEX IF NOT EXISTS idx_dispatch_suggestions_expires ON dispatch_suggestions (expires_at);
