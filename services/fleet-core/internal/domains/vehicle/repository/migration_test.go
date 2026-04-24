package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestApplyMigration315(t *testing.T) {
	connStr := "postgres://flyfleet_admin:flyfleet_secret@localhost:5432/flyfleet?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	migrationSQL := `
CREATE TABLE IF NOT EXISTS billing_hardware_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    model_code VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    one_time_price DECIMAL(15, 2) NOT NULL,
    lease_monthly_price DECIMAL(15, 2),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_device_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    product_id UUID NOT NULL REFERENCES billing_hardware_products(id),
    billing_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING_ACTIVATION',
    activated_at TIMESTAMPTZ,
    invoice_id UUID REFERENCES invoices(id), 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(device_id)
);

INSERT INTO billing_hardware_products (name, model_code, category, one_time_price, lease_monthly_price)
VALUES 
('FlyFleet Vision Pro Dashcam', 'FF-CAM-V2', 'DASHCAM', 399.00, 25.00),
('FlyFleet Standard GPS Tracker', 'FF-GPS-100', 'TRACKER', 99.00, 10.00),
('FlyFleet OBD-II Telematics Node', 'FF-OBD-V1', 'TRACKER', 149.00, 15.00)
ON CONFLICT (model_code) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_device_assignments_tenant_id ON billing_device_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_device_assignments_device_id ON billing_device_assignments(device_id);
`

	_, err = pool.Exec(context.Background(), migrationSQL)
	assert.NoError(t, err)
}
