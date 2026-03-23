-- Migration: 243_finance_fuel_stabilization
-- Description: Stabilize finance/fuel schema for enterprise production readiness

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -----------------------------------------------------------------------------
-- Fuel schema drift: canonical organization_id
-- -----------------------------------------------------------------------------

ALTER TABLE IF EXISTS fuel_logs
    ADD COLUMN IF NOT EXISTS organization_id UUID,
    ADD COLUMN IF NOT EXISTS driver_id UUID,
    ADD COLUMN IF NOT EXISTS station_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS fuel_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS quantity DECIMAL(10, 2),
    ADD COLUMN IF NOT EXISTS unit_price DECIMAL(10, 2),
    ADD COLUMN IF NOT EXISTS total_cost DECIMAL(10, 2),
    ADD COLUMN IF NOT EXISTS odometer DECIMAL(12, 2),
    ADD COLUMN IF NOT EXISTS transaction_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'PENDING',
    ADD COLUMN IF NOT EXISTS gst_number VARCHAR(15),
    ADD COLUMN IF NOT EXISTS gst_rate DECIMAL(5, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS gst_amount DECIMAL(10, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100),
    ADD COLUMN IF NOT EXISTS vendor_gstin VARCHAR(15),
    ADD COLUMN IF NOT EXISTS receipt_url TEXT,
    ADD COLUMN IF NOT EXISTS notes TEXT,
    ADD COLUMN IF NOT EXISTS anomaly_flag BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS anomaly_reason TEXT,
    ADD COLUMN IF NOT EXISTS fraud_score NUMERIC(5, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS risk_level VARCHAR(20) DEFAULT 'LOW',
    ADD COLUMN IF NOT EXISTS fraud_flags JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND column_name = 'org_id'
    ) THEN
        EXECUTE 'UPDATE fuel_logs SET organization_id = COALESCE(organization_id, org_id)';
    END IF;
END $$;

ALTER TABLE IF EXISTS fuel_transactions
    ADD COLUMN IF NOT EXISTS organization_id UUID,
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS gst_number VARCHAR(15),
    ADD COLUMN IF NOT EXISTS gst_rate DECIMAL(5, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS gst_amount DECIMAL(10, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100),
    ADD COLUMN IF NOT EXISTS vendor_gstin VARCHAR(15),
    ADD COLUMN IF NOT EXISTS fraud_score NUMERIC(5, 2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS risk_level VARCHAR(20) DEFAULT 'LOW',
    ADD COLUMN IF NOT EXISTS fraud_flags JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_transactions' AND column_name = 'org_id'
    ) THEN
        EXECUTE 'UPDATE fuel_transactions SET organization_id = COALESCE(organization_id, org_id)';
    END IF;
END $$;

ALTER TABLE IF EXISTS fuel_audit_logs
    ADD COLUMN IF NOT EXISTS organization_id UUID,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'fuel_logs'
    ) THEN
        EXECUTE '
            UPDATE fuel_audit_logs fa
            SET organization_id = fl.organization_id
            FROM fuel_logs fl
            WHERE fa.fuel_log_id = fl.id AND fa.organization_id IS NULL
        ';
    END IF;
END $$;

-- -----------------------------------------------------------------------------
-- Missing enterprise finance tables
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS expense_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    report_id UUID NOT NULL,
    category VARCHAR(30) NOT NULL CHECK (category IN ('FUEL','MEALS','LODGING','TOLL','MAINTENANCE','OTHER')),
    amount DECIMAL(15, 2) NOT NULL,
    gst_amount DECIMAL(15, 2) DEFAULT 0,
    gst_rate DECIMAL(5, 2) DEFAULT 0,
    invoice_number VARCHAR(100),
    receipt_url TEXT,
    description TEXT,
    vendor_name VARCHAR(255),
    vendor_gstin VARCHAR(15),
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    ocr_processed BOOLEAN DEFAULT FALSE,
    ocr_confidence DECIMAL(4, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE IF EXISTS expense_items
    ADD COLUMN IF NOT EXISTS org_id UUID,
    ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

CREATE TABLE IF NOT EXISTS gst_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    irn VARCHAR(64),
    ack_number VARCHAR(30),
    ack_date TIMESTAMP WITH TIME ZONE,
    qr_code TEXT,
    signed_invoice TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',
    entity_type VARCHAR(30),
    entity_id UUID,
    org_id UUID NOT NULL,
    gstin VARCHAR(15) NOT NULL,
    buyer_gstin VARCHAR(15),
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    taxable_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    cgst_amount DECIMAL(15, 2) DEFAULT 0,
    sgst_amount DECIMAL(15, 2) DEFAULT 0,
    igst_amount DECIMAL(15, 2) DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tds_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(30) NOT NULL,
    entity_id UUID NOT NULL,
    vendor_id UUID,
    org_id UUID NOT NULL,
    pan_number VARCHAR(10),
    tds_section VARCHAR(10) DEFAULT '194C',
    gross_amount DECIMAL(15, 2) NOT NULL,
    tds_rate DECIMAL(5, 2) NOT NULL,
    tds_amount DECIMAL(15, 2) NOT NULL,
    net_amount DECIMAL(15, 2) NOT NULL,
    deducted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deposited BOOLEAN DEFAULT FALSE,
    deposit_challan VARCHAR(50),
    financial_year VARCHAR(10),
    quarter VARCHAR(5),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vendor_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vendor_id UUID NOT NULL,
    org_id UUID NOT NULL,
    invoice_number VARCHAR(50) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    tds_deducted DECIMAL(15, 2) DEFAULT 0,
    net_payable DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    payment_status VARCHAR(20) DEFAULT 'PENDING',
    payment_mode VARCHAR(30),
    payment_reference VARCHAR(100),
    bank_account VARCHAR(30),
    ifsc_code VARCHAR(11),
    payment_date DATE,
    paid_at TIMESTAMP WITH TIME ZONE,
    gst_invoice_id UUID,
    tds_record_id UUID,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fuel_card_reconciliation (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    card_transaction_id UUID,
    internal_transaction_id UUID,
    status VARCHAR(20) DEFAULT 'PENDING',
    variance DECIMAL(15, 2) DEFAULT 0,
    card_amount DECIMAL(15, 2),
    internal_amount DECIMAL(15, 2),
    reconciled_by UUID,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE IF EXISTS toll_transactions
    ADD COLUMN IF NOT EXISTS org_id UUID,
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

ALTER TABLE IF EXISTS currency_rates
    ADD COLUMN IF NOT EXISTS currency_code VARCHAR(3),
    ADD COLUMN IF NOT EXISTS org_id UUID,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

UPDATE currency_rates
SET currency_code = COALESCE(currency_code, from_currency);

CREATE TABLE IF NOT EXISTS finance_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    entity_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id UUID,
    old_values JSONB,
    new_values JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS receipt_ocr_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    expense_item_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempts INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- -----------------------------------------------------------------------------
-- Indexes
-- -----------------------------------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_fuel_logs_org_id ON fuel_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_risk_level ON fuel_logs(risk_level);
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_org_id ON fuel_transactions(organization_id);
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_risk_level ON fuel_transactions(risk_level);
CREATE INDEX IF NOT EXISTS idx_fuel_audit_logs_org_id ON fuel_audit_logs(organization_id);

CREATE INDEX IF NOT EXISTS idx_expense_items_org_id ON expense_items(org_id);
CREATE INDEX IF NOT EXISTS idx_expense_items_report_id ON expense_items(report_id);
CREATE INDEX IF NOT EXISTS idx_expense_items_ocr_pending ON expense_items(ocr_processed, created_at);

CREATE INDEX IF NOT EXISTS idx_gst_invoices_org_id ON gst_invoices(org_id);
CREATE INDEX IF NOT EXISTS idx_gst_invoices_entity ON gst_invoices(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_gst_invoices_status ON gst_invoices(status);

CREATE INDEX IF NOT EXISTS idx_tds_records_org_id ON tds_records(org_id);
CREATE INDEX IF NOT EXISTS idx_tds_records_vendor_id ON tds_records(vendor_id);

CREATE INDEX IF NOT EXISTS idx_vendor_payments_org_id ON vendor_payments(org_id);
CREATE INDEX IF NOT EXISTS idx_vendor_payments_vendor_id ON vendor_payments(vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_payments_status ON vendor_payments(payment_status);

CREATE INDEX IF NOT EXISTS idx_toll_transactions_org_id ON toll_transactions(org_id);
CREATE INDEX IF NOT EXISTS idx_toll_transactions_vehicle ON toll_transactions(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_fuel_card_reconciliation_org_id ON fuel_card_reconciliation(org_id);
CREATE INDEX IF NOT EXISTS idx_fuel_card_reconciliation_status ON fuel_card_reconciliation(status);

CREATE INDEX IF NOT EXISTS idx_currency_rates_code ON currency_rates(currency_code);
CREATE INDEX IF NOT EXISTS idx_currency_rates_pair ON currency_rates(from_currency, to_currency);

CREATE INDEX IF NOT EXISTS idx_finance_audit_logs_org_id ON finance_audit_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_finance_audit_logs_entity ON finance_audit_logs(entity_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_finance_audit_logs_ts ON finance_audit_logs(timestamp);

CREATE INDEX IF NOT EXISTS idx_receipt_ocr_jobs_status ON receipt_ocr_jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_receipt_ocr_jobs_expense_item ON receipt_ocr_jobs(expense_item_id);

-- -----------------------------------------------------------------------------
-- Foreign keys (idempotent)
-- -----------------------------------------------------------------------------

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'expense_reports') THEN
        ALTER TABLE expense_items
            ADD CONSTRAINT expense_items_report_id_fkey
            FOREIGN KEY (report_id) REFERENCES expense_reports(id) ON DELETE CASCADE;
    END IF;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'fuel_vendors') THEN
        ALTER TABLE vendor_payments
            ADD CONSTRAINT vendor_payments_vendor_id_fkey
            FOREIGN KEY (vendor_id) REFERENCES fuel_vendors(id) ON DELETE RESTRICT;
    END IF;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
    ALTER TABLE vendor_payments
        ADD CONSTRAINT vendor_payments_gst_invoice_id_fkey
        FOREIGN KEY (gst_invoice_id) REFERENCES gst_invoices(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
    ALTER TABLE vendor_payments
        ADD CONSTRAINT vendor_payments_tds_record_id_fkey
        FOREIGN KEY (tds_record_id) REFERENCES tds_records(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
    ALTER TABLE receipt_ocr_jobs
        ADD CONSTRAINT receipt_ocr_jobs_expense_item_fkey
        FOREIGN KEY (expense_item_id) REFERENCES expense_items(id) ON DELETE CASCADE;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- -----------------------------------------------------------------------------
-- updated_at trigger
-- -----------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION set_finance_fuel_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_expense_items_updated_at') THEN
        CREATE TRIGGER trg_expense_items_updated_at BEFORE UPDATE ON expense_items
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_gst_invoices_updated_at') THEN
        CREATE TRIGGER trg_gst_invoices_updated_at BEFORE UPDATE ON gst_invoices
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_tds_records_updated_at') THEN
        CREATE TRIGGER trg_tds_records_updated_at BEFORE UPDATE ON tds_records
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_vendor_payments_updated_at') THEN
        CREATE TRIGGER trg_vendor_payments_updated_at BEFORE UPDATE ON vendor_payments
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fuel_card_reconciliation_updated_at') THEN
        CREATE TRIGGER trg_fuel_card_reconciliation_updated_at BEFORE UPDATE ON fuel_card_reconciliation
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_currency_rates_updated_at') THEN
        CREATE TRIGGER trg_currency_rates_updated_at BEFORE UPDATE ON currency_rates
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_finance_audit_logs_updated_at') THEN
        CREATE TRIGGER trg_finance_audit_logs_updated_at BEFORE UPDATE ON finance_audit_logs
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_receipt_ocr_jobs_updated_at') THEN
        CREATE TRIGGER trg_receipt_ocr_jobs_updated_at BEFORE UPDATE ON receipt_ocr_jobs
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
END $$;
