-- Migration: 223_fastag_toll_management
-- Description: Evolve toll_transactions table for FASTag integration with robustness for existing 201 schema.

DO $$ 
BEGIN 
    -- 1. Create or Evolve toll_transactions
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'toll_transactions') THEN
        CREATE TABLE toll_transactions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            tenant_id UUID NOT NULL REFERENCES tenants(id),
            vehicle_id UUID NOT NULL,
            fastag_id VARCHAR(100) NOT NULL,
            toll_plaza VARCHAR(255) NOT NULL,
            amount DECIMAL(10, 2) NOT NULL,
            currency VARCHAR(3) DEFAULT 'INR',
            transaction_time TIMESTAMP WITH TIME ZONE NOT NULL,
            external_transaction_id VARCHAR(255) UNIQUE,
            metadata JSONB DEFAULT '{}',
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
    ELSE
        -- Harmonize columns from 201 schema to Enterprise model
        
        -- Rename 'timestamp' to 'transaction_time'
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'timestamp') AND
           NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'transaction_time') THEN
            ALTER TABLE toll_transactions RENAME COLUMN timestamp TO transaction_time;
        END IF;

        -- Rename 'plaza_name' to 'toll_plaza'
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'plaza_name') AND
           NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'toll_plaza') THEN
            ALTER TABLE toll_transactions RENAME COLUMN plaza_name TO toll_plaza;
        END IF;

        -- Add missing columns
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'currency') THEN
            ALTER TABLE toll_transactions ADD COLUMN currency VARCHAR(3) DEFAULT 'INR';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'external_transaction_id') THEN
            ALTER TABLE toll_transactions ADD COLUMN external_transaction_id VARCHAR(255) UNIQUE;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'metadata') THEN
            ALTER TABLE toll_transactions ADD COLUMN metadata JSONB DEFAULT '{}';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'toll_transactions' AND column_name = 'tenant_id') THEN
            ALTER TABLE toll_transactions ADD COLUMN tenant_id UUID REFERENCES tenants(id);
        END IF;
    END IF;
END $$;

-- 2. Ensure Indexes
CREATE INDEX IF NOT EXISTS idx_toll_vehicle_id ON toll_transactions(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_toll_fastag_id ON toll_transactions(fastag_id);
CREATE INDEX IF NOT EXISTS idx_toll_time ON toll_transactions(transaction_time);
CREATE INDEX IF NOT EXISTS idx_toll_tenant_id ON toll_transactions(tenant_id);

-- 3. Comments
COMMENT ON TABLE toll_transactions IS 'Toll transactions ingested from FASTag providers';
