--
-- PostgreSQL database dump
--

\restrict o0YlzcLDfgh03pZOAHhtgxuxTz3zJnggoOCgnQ25Y1mGT9KBfKQdaKzExtW48Ee

-- Dumped from database version 18.2
-- Dumped by pg_dump version 18.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: drop_old_telemetry_partitions(integer); Type: FUNCTION; Schema: public; Owner: fleet_user
--

CREATE FUNCTION public.drop_old_telemetry_partitions(retention_days integer DEFAULT 90) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    partition_record RECORD;
    partition_date DATE;
    cutoff_date DATE;
BEGIN
    cutoff_date := CURRENT_DATE - retention_days;

    FOR partition_record IN
        SELECT
            c.relname AS partition_name
        FROM
            pg_class c
        JOIN
            pg_inherits i ON c.oid = i.inhrelid
        JOIN
            pg_class p ON i.inhparent = p.oid
        WHERE
            p.relname = 'telemetry_events'
    LOOP
        -- Extract the date part from the partition name (assuming format: telemetry_events_pYYYY_MM)
        -- Example: telemetry_events_p2023_01
        
        BEGIN
            -- Using regex to extract YYYY_MM
            partition_date := to_date(substring(partition_record.partition_name from 'p(\d{4}_\d{2})'), 'YYYY_MM');
            
            -- Because the partition name represents the START of the month, 
            -- we want to drop it if the ENTIRE month is safely past the cutoff date.
            -- Adding 1 month to the start date gives us the upper bound of the partition.
            IF (partition_date + interval '1 month') < cutoff_date THEN
                EXECUTE format('DROP TABLE IF EXISTS %I;', partition_record.partition_name);
                RAISE NOTICE 'Dropped old partition: %', partition_record.partition_name;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            -- Ignore parsing errors for malformed names, just log it if necessary
            RAISE NOTICE 'Skipping unrecognized partition name: %', partition_record.partition_name;
        END;
    END LOOP;
END;
$$;


ALTER FUNCTION public.drop_old_telemetry_partitions(retention_days integer) OWNER TO fleet_user;

--
-- Name: get_current_org_id(); Type: FUNCTION; Schema: public; Owner: fleet_user
--

CREATE FUNCTION public.get_current_org_id() RETURNS uuid
    LANGUAGE plpgsql STABLE
    AS $$
BEGIN
  RETURN current_setting('flyfleet.current_org_id')::uuid;
EXCEPTION
  WHEN OTHERS THEN RETURN NULL;
END;
$$;


ALTER FUNCTION public.get_current_org_id() OWNER TO fleet_user;

--
-- Name: refresh_vehicle_status(); Type: FUNCTION; Schema: public; Owner: fleet_user
--

CREATE FUNCTION public.refresh_vehicle_status() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY vehicle_status;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.refresh_vehicle_status() OWNER TO fleet_user;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: fleet_user
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO fleet_user;

--
-- Name: FUNCTION update_updated_at_column(); Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON FUNCTION public.update_updated_at_column() IS 'Automatically sets updated_at = NOW() on row updates. Applied to all tables with updated_at column via trigger.';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: access_reviews; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.access_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    due_date timestamp with time zone,
    total_users integer DEFAULT 0,
    reviewed_users integer DEFAULT 0,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT access_reviews_status_check CHECK (((status)::text = ANY ((ARRAY['PENDING'::character varying, 'IN_PROGRESS'::character varying, 'COMPLETED'::character varying])::text[])))
);


ALTER TABLE public.access_reviews OWNER TO fleet_user;

--
-- Name: alert_rules; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.alert_rules (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    org_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(50) NOT NULL,
    description text,
    severity character varying(20) DEFAULT 'warning'::character varying,
    conditions jsonb NOT NULL,
    actions jsonb,
    vehicles jsonb,
    driver_ids jsonb,
    enabled boolean DEFAULT true,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.alert_rules OWNER TO fleet_user;

--
-- Name: alerts; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.alerts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    org_id uuid NOT NULL,
    rule_id uuid,
    type character varying(50) NOT NULL,
    severity character varying(20) DEFAULT 'warning'::character varying,
    vehicle_id uuid,
    driver_id uuid,
    "timestamp" timestamp with time zone DEFAULT now(),
    location jsonb,
    details jsonb,
    status character varying(20) DEFAULT 'new'::character varying,
    notes text,
    acknowledged_by uuid,
    acknowledged_at timestamp with time zone,
    resolved_by uuid,
    resolved_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    dedup_key text,
    escalation_level integer DEFAULT 0,
    last_escalated_at timestamp with time zone,
    escalated_to uuid[],
    sla_deadline timestamp with time zone,
    CONSTRAINT chk_alerts_status CHECK (((status)::text = ANY ((ARRAY['new'::character varying, 'acknowledged'::character varying, 'resolved'::character varying, 'dismissed'::character varying, 'escalated'::character varying])::text[])))
);


ALTER TABLE public.alerts OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_alerts_status ON alerts; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_alerts_status ON public.alerts IS 'Enforces allowed values for status column';


--
-- Name: analytics_jobs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.analytics_jobs (
    id character varying(36) NOT NULL,
    name character varying(255) NOT NULL,
    type character varying(100) NOT NULL,
    description text,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    job_type character varying(100) NOT NULL,
    parameters jsonb DEFAULT '{}'::jsonb NOT NULL,
    input_source character varying(255) NOT NULL,
    output_destination character varying(255) NOT NULL,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    progress numeric(5,2) DEFAULT 0,
    error_details text,
    org_id uuid CONSTRAINT analytics_jobs_organization_id_not_null NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT analytics_jobs_progress_check CHECK (((progress >= (0)::numeric) AND (progress <= (100)::numeric)))
);


ALTER TABLE public.analytics_jobs OWNER TO fleet_user;

--
-- Name: api_clients; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.api_clients (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    client_id character varying(100) NOT NULL,
    client_secret_hash text NOT NULL,
    name character varying(255) NOT NULL,
    scopes text[] DEFAULT '{}'::text[],
    expiry timestamp with time zone,
    status character varying(20) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.api_clients OWNER TO fleet_user;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    action character varying(100) NOT NULL,
    resource character varying(100) NOT NULL,
    details jsonb,
    ip_address character varying(45),
    user_agent text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    severity character varying(20) DEFAULT 'info'::character varying,
    status character varying(20) DEFAULT 'success'::character varying,
    org_id uuid,
    module character varying(100),
    entity_id character varying(100),
    before_state jsonb,
    after_state jsonb,
    metadata jsonb
);


ALTER TABLE public.audit_logs OWNER TO fleet_user;

--
-- Name: audit_policies; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.audit_policies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    retention_years integer DEFAULT 7,
    archive_policy character varying(100) DEFAULT 'MOVE_TO_COLD_STORAGE_1Y'::character varying,
    masking_rules jsonb DEFAULT '["credit_card", "ssn", "phone", "email"]'::jsonb,
    updated_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.audit_policies OWNER TO fleet_user;

--
-- Name: audit_settings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.audit_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    retention_days_online integer DEFAULT 90,
    retention_days_archival integer DEFAULT 2555,
    masking_rules jsonb DEFAULT '["password", "secret", "token", "phone", "email_local"]'::jsonb,
    tracked_event_types jsonb DEFAULT '["MFA_POLICY_UPDATE", "MFA_EXCEPTION_GRANT", "MFA_OVERRIDE_TOGGLE", "LOGIN_SUCCESS", "LOGIN_FAILED"]'::jsonb,
    alert_on_critical boolean DEFAULT true,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.audit_settings OWNER TO fleet_user;

--
-- Name: auth_tokens; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.auth_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    access_token text NOT NULL,
    refresh_token text NOT NULL,
    device_id character varying(255),
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    ip_address character varying(45),
    user_agent text,
    last_active timestamp with time zone DEFAULT now(),
    is_active boolean DEFAULT true
);


ALTER TABLE public.auth_tokens OWNER TO fleet_user;

--
-- Name: backup_verifications; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.backup_verifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    backup_id uuid NOT NULL,
    status character varying(50) DEFAULT 'PENDING'::character varying NOT NULL,
    checksum character varying(128),
    verified_at timestamp with time zone,
    verified_by uuid,
    error_message text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.backup_verifications OWNER TO fleet_user;

--
-- Name: backups; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.backups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    s3_key character varying(255) NOT NULL,
    size_bytes bigint DEFAULT 0 NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb
);


ALTER TABLE public.backups OWNER TO fleet_user;

--
-- Name: billing_accounting_exports; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_accounting_exports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    external_system character varying(50) NOT NULL,
    export_type character varying(50) NOT NULL,
    status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    file_key text,
    records_count integer DEFAULT 0 NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    completed_at timestamp with time zone,
    created_by uuid
);


ALTER TABLE public.billing_accounting_exports OWNER TO fleet_user;

--
-- Name: billing_coupons; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_coupons (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    discount_type character varying(20) NOT NULL,
    discount_value numeric(15,4) NOT NULL,
    currency character varying(3),
    max_redemptions integer,
    redeemed_count integer DEFAULT 0 NOT NULL,
    valid_from timestamp with time zone,
    valid_until timestamp with time zone,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid
);


ALTER TABLE public.billing_coupons OWNER TO fleet_user;

--
-- Name: billing_credit_transactions; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_credit_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    balance_id uuid NOT NULL,
    transaction_type character varying(50) NOT NULL,
    amount numeric(15,4) NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    reference_type character varying(50),
    reference_id uuid,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid
);


ALTER TABLE public.billing_credit_transactions OWNER TO fleet_user;

--
-- Name: billing_customer_balances; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_customer_balances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    balance numeric(15,4) DEFAULT 0.0000 NOT NULL,
    credit_limit numeric(15,4) DEFAULT 0.0000 NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_customer_balances OWNER TO fleet_user;

--
-- Name: billing_hardware_products; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_hardware_products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    model_code character varying(100) NOT NULL,
    category character varying(50) NOT NULL,
    one_time_price numeric(15,2) NOT NULL,
    lease_monthly_price numeric(15,2),
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.billing_hardware_products OWNER TO fleet_user;

--
-- Name: billing_invoice_documents; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_invoice_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_id uuid NOT NULL,
    template_id uuid NOT NULL,
    file_url text NOT NULL,
    storage_key text NOT NULL,
    generated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    template_version integer DEFAULT 1 NOT NULL,
    snapshot jsonb DEFAULT '{}'::jsonb NOT NULL,
    checksum text
);


ALTER TABLE public.billing_invoice_documents OWNER TO fleet_user;

--
-- Name: COLUMN billing_invoice_documents.snapshot; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.billing_invoice_documents.snapshot IS 'JSON snapshot of all data used for rendering this specific PDF version';


--
-- Name: COLUMN billing_invoice_documents.checksum; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.billing_invoice_documents.checksum IS 'SHA256 checksum of the generated PDF file for integrity verification';


--
-- Name: billing_invoice_sequences; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_invoice_sequences (
    org_id uuid NOT NULL,
    prefix character varying(20) DEFAULT 'INV-'::character varying NOT NULL,
    next_value bigint DEFAULT 1 NOT NULL,
    suffix character varying(20) DEFAULT ''::character varying,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_invoice_sequences OWNER TO fleet_user;

--
-- Name: billing_invoice_templates; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_invoice_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    version integer NOT NULL,
    html_content text NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_invoice_templates OWNER TO fleet_user;

--
-- Name: billing_plans; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_plans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_plans OWNER TO fleet_user;

--
-- Name: billing_pricing_components; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_pricing_components (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    plan_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(20) NOT NULL,
    unit_price numeric(15,4) DEFAULT 0.0000 NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    meter_event_name character varying(50),
    frequency character varying(20) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_pricing_components OWNER TO fleet_user;

--
-- Name: billing_recognized_revenue; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_recognized_revenue (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    schedule_id uuid NOT NULL,
    amount numeric(15,4) NOT NULL,
    recognized_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_recognized_revenue OWNER TO fleet_user;

--
-- Name: billing_revenue_schedules; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_revenue_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    invoice_line_item_id uuid NOT NULL,
    total_amount numeric(15,4) NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_revenue_schedules OWNER TO fleet_user;

--
-- Name: billing_settings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_settings (
    org_id uuid NOT NULL,
    default_tax_rate numeric(5,2) DEFAULT 0.00 NOT NULL,
    base_currency text DEFAULT 'USD'::text NOT NULL,
    grace_period_days integer DEFAULT 14 NOT NULL,
    auto_invoicing_enabled boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_settings OWNER TO fleet_user;

--
-- Name: billing_subscriptions; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    plan_id uuid NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    current_period_start timestamp with time zone NOT NULL,
    current_period_end timestamp with time zone NOT NULL,
    trial_end timestamp with time zone,
    cancel_at_period_end boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_billing_subscriptions_status CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'past_due'::character varying, 'canceled'::character varying, 'expired'::character varying])::text[])))
);


ALTER TABLE public.billing_subscriptions OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_billing_subscriptions_status ON billing_subscriptions; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_billing_subscriptions_status ON public.billing_subscriptions IS 'Enforces allowed values for status column';


--
-- Name: billing_tax_rates; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_tax_rates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    region_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    percentage numeric(5,2) NOT NULL,
    tax_type character varying(50) NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_tax_rates OWNER TO fleet_user;

--
-- Name: billing_tax_regions; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.billing_tax_regions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    country_code character varying(2) NOT NULL,
    state_code character varying(10),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.billing_tax_regions OWNER TO fleet_user;

--
-- Name: bulk_config; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.bulk_config (
    org_id uuid CONSTRAINT bulk_config_organization_id_not_null NOT NULL,
    max_records_per_file integer DEFAULT 5000,
    concurrent_threads integer DEFAULT 4,
    retention_months integer DEFAULT 6,
    mandatory_fields jsonb DEFAULT '["vin", "registration_no", "make", "model_year"]'::jsonb,
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.bulk_config OWNER TO fleet_user;

--
-- Name: bulk_job_records; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.bulk_job_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    job_id uuid NOT NULL,
    row_number integer NOT NULL,
    record_data jsonb,
    status character varying(50) NOT NULL,
    error_message text,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT bulk_job_records_status_check CHECK (((status)::text = ANY ((ARRAY['success'::character varying, 'failed'::character varying])::text[])))
);


ALTER TABLE public.bulk_job_records OWNER TO fleet_user;

--
-- Name: bulk_jobs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.bulk_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT bulk_jobs_organization_id_not_null NOT NULL,
    file_name character varying(255) NOT NULL,
    status character varying(50) NOT NULL,
    total_records integer DEFAULT 0,
    success_count integer DEFAULT 0,
    failed_count integer DEFAULT 0,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    retry_count integer DEFAULT 0,
    CONSTRAINT bulk_jobs_status_check CHECK (((status)::text = ANY ((ARRAY['processing'::character varying, 'completed'::character varying, 'failed'::character varying, 'partial_success'::character varying])::text[]))),
    CONSTRAINT chk_bulk_jobs_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'completed'::character varying, 'failed'::character varying, 'cancelled'::character varying])::text[])))
);


ALTER TABLE public.bulk_jobs OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_bulk_jobs_status ON bulk_jobs; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_bulk_jobs_status ON public.bulk_jobs IS 'Enforces allowed values for status column';


--
-- Name: consumption_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.consumption_events (
    "time" timestamp with time zone NOT NULL,
    org_id uuid NOT NULL,
    vehicle_id uuid,
    meter_code character varying(50) NOT NULL,
    amount numeric(15,4) NOT NULL,
    metadata jsonb,
    processed_invoice_id uuid
);


ALTER TABLE public.consumption_events OWNER TO fleet_user;

--
-- Name: documents; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vehicle_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    doc_number character varying(100) NOT NULL,
    issue_date date NOT NULL,
    expiry_date date NOT NULL,
    file_url text,
    status character varying(20) DEFAULT 'pending'::character varying,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    org_id uuid,
    title character varying(255),
    file_name character varying(255) DEFAULT ''::character varying,
    file_size_bytes integer DEFAULT 0,
    mime_type character varying(100) DEFAULT 'application/octet-stream'::character varying,
    storage_path character varying(500) DEFAULT ''::character varying,
    reminder_days integer DEFAULT 30,
    uploaded_by uuid,
    deleted_at timestamp with time zone
);


ALTER TABLE public.documents OWNER TO fleet_user;

--
-- Name: TABLE documents; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.documents IS 'Stores compliance documents associated with vehicles';


--
-- Name: COLUMN documents.metadata; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.documents.metadata IS 'Flexible JSON storage for document-specific attributes';


--
-- Name: dr_test_logs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.dr_test_logs (
    id uuid NOT NULL,
    test_id uuid NOT NULL,
    ts timestamp with time zone DEFAULT now() NOT NULL,
    level text NOT NULL,
    message text NOT NULL
);


ALTER TABLE public.dr_test_logs OWNER TO fleet_user;

--
-- Name: dr_test_settings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.dr_test_settings (
    id boolean DEFAULT true NOT NULL,
    schedule_enabled boolean DEFAULT false NOT NULL,
    frequency text DEFAULT 'WEEKLY'::text NOT NULL,
    schedule_day text DEFAULT 'SATURDAY'::text NOT NULL,
    schedule_time text DEFAULT '02:00'::text NOT NULL,
    intensity text DEFAULT 'NORMAL'::text NOT NULL,
    notification_emails jsonb DEFAULT '[]'::jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT dr_test_settings_id_check CHECK ((id = true))
);


ALTER TABLE public.dr_test_settings OWNER TO fleet_user;

--
-- Name: dr_test_steps; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.dr_test_steps (
    id uuid NOT NULL,
    test_id uuid NOT NULL,
    step_order integer NOT NULL,
    step_name text NOT NULL,
    status text NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone
);


ALTER TABLE public.dr_test_steps OWNER TO fleet_user;

--
-- Name: dr_tests; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.dr_tests (
    id uuid NOT NULL,
    test_id text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text,
    test_type text NOT NULL,
    target_env text NOT NULL,
    status text NOT NULL,
    progress_percent integer DEFAULT 0 NOT NULL,
    initiated_by text DEFAULT ''::text NOT NULL,
    success_criteria text DEFAULT ''::text,
    scheduled_at timestamp with time zone,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.dr_tests OWNER TO fleet_user;

--
-- Name: driver_attendance; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_attendance (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid NOT NULL,
    attendance_date date NOT NULL,
    status character varying(20) DEFAULT 'present'::character varying NOT NULL,
    check_in_time timestamp with time zone,
    check_out_time timestamp with time zone,
    remarks text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.driver_attendance OWNER TO fleet_user;

--
-- Name: TABLE driver_attendance; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.driver_attendance IS 'Normalized driver attendance records (daily attendance tracking). Replaces attendance_records JSONB in drivers table.';


--
-- Name: COLUMN driver_attendance.attendance_date; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_attendance.attendance_date IS 'Date of the attendance record (date only, no time)';


--
-- Name: COLUMN driver_attendance.status; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_attendance.status IS 'Attendance status: present, absent, leave, holiday';


--
-- Name: COLUMN driver_attendance.check_in_time; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_attendance.check_in_time IS 'Actual check-in timestamp (null if not checked in)';


--
-- Name: COLUMN driver_attendance.check_out_time; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_attendance.check_out_time IS 'Actual check-out timestamp (null if not checked out)';


--
-- Name: driver_behavior; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_behavior (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid NOT NULL,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    score double precision,
    safety_score double precision,
    efficiency_score double precision,
    compliance_score double precision,
    harsh_accelerations integer DEFAULT 0,
    harsh_brakings integer DEFAULT 0,
    harsh_cornerings integer DEFAULT 0,
    speeding_incidents integer DEFAULT 0,
    phone_usage integer DEFAULT 0,
    seatbelt_violations integer DEFAULT 0,
    fuel_consumption double precision DEFAULT 0,
    distance_driven double precision DEFAULT 0,
    idle_time integer DEFAULT 0,
    eco_driving_score double precision,
    hours_of_service double precision,
    rest_periods integer,
    document_compliance boolean DEFAULT true,
    maintenance_adherence boolean DEFAULT true,
    night_driving_hours double precision DEFAULT 0,
    highway_percentage double precision DEFAULT 0,
    city_driving_percentage double precision DEFAULT 0,
    risk_level character varying(20),
    risk_factors jsonb,
    improvement_areas jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.driver_behavior OWNER TO fleet_user;

--
-- Name: driver_behavior_analytics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_behavior_analytics (
    id character varying(36) NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid CONSTRAINT driver_behavior_analytics_organization_id_not_null NOT NULL,
    analysis_period_start date NOT NULL,
    analysis_period_end date NOT NULL,
    total_driving_hours numeric(8,2) DEFAULT 0 NOT NULL,
    aggressive_acceleration_count integer DEFAULT 0 CONSTRAINT driver_behavior_analytics_aggressive_acceleration_coun_not_null NOT NULL,
    hard_braking_count integer DEFAULT 0 NOT NULL,
    sharp_turn_count integer DEFAULT 0 NOT NULL,
    over_speeding_count integer DEFAULT 0 NOT NULL,
    idling_time numeric(8,2) DEFAULT 0 NOT NULL,
    average_speed numeric(8,2) DEFAULT 0 NOT NULL,
    max_speed numeric(8,2) DEFAULT 0 NOT NULL,
    fuel_efficiency numeric(8,2) DEFAULT 0 NOT NULL,
    safety_score numeric(5,2) DEFAULT 0 NOT NULL,
    efficiency_score numeric(5,2) DEFAULT 0 NOT NULL,
    comfort_score numeric(5,2) DEFAULT 0 NOT NULL,
    ranking integer,
    improvement_suggestions jsonb DEFAULT '[]'::jsonb NOT NULL,
    behavior_trend character varying(20) DEFAULT 'stable'::character varying NOT NULL,
    generated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT driver_behavior_analytics_comfort_score_check CHECK (((comfort_score >= (0)::numeric) AND (comfort_score <= (100)::numeric))),
    CONSTRAINT driver_behavior_analytics_efficiency_score_check CHECK (((efficiency_score >= (0)::numeric) AND (efficiency_score <= (100)::numeric))),
    CONSTRAINT driver_behavior_analytics_safety_score_check CHECK (((safety_score >= (0)::numeric) AND (safety_score <= (100)::numeric)))
);


ALTER TABLE public.driver_behavior_analytics OWNER TO fleet_user;

--
-- Name: driver_documents; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid NOT NULL,
    type character varying(100) NOT NULL,
    name character varying(255) NOT NULL,
    document_number character varying(100),
    issue_date date,
    expiry_date date,
    issuing_authority character varying(255),
    file_url text,
    file_name character varying(255),
    file_size_bytes bigint,
    mime_type character varying(100),
    status character varying(50) DEFAULT 'pending'::character varying,
    verified_by uuid,
    verified_at timestamp with time zone,
    verification_notes text,
    is_mandatory boolean DEFAULT true,
    reminder_days_before_expiry integer DEFAULT 30,
    last_reminder_sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid,
    updated_by uuid
);


ALTER TABLE public.driver_documents OWNER TO fleet_user;

--
-- Name: TABLE driver_documents; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.driver_documents IS 'Normalized driver documents (license, PAN, Aadhaar, etc.). Replaces drivers.documents JSONB array.';


--
-- Name: COLUMN driver_documents.type; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_documents.type IS 'Document type: license, pan_card, aadhaar, insurance, pollution_certificate, etc.';


--
-- Name: COLUMN driver_documents.status; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_documents.status IS 'Verification status: pending, verified, rejected, expired';


--
-- Name: COLUMN driver_documents.is_mandatory; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_documents.is_mandatory IS 'Whether this document is required for the driver to be active';


--
-- Name: COLUMN driver_documents.reminder_days_before_expiry; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_documents.reminder_days_before_expiry IS 'Days before expiry to send reminder notification';


--
-- Name: driver_incidents; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_incidents (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    severity character varying(20) NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    location jsonb,
    speed double precision,
    details jsonb,
    resolved boolean DEFAULT false,
    resolution_notes text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.driver_incidents OWNER TO fleet_user;

--
-- Name: driver_schedules; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_schedules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid NOT NULL,
    day_of_week integer NOT NULL,
    start_time character varying(10) NOT NULL,
    end_time character varying(10) NOT NULL,
    is_working boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT driver_schedules_day_of_week_check CHECK (((day_of_week >= 0) AND (day_of_week <= 6)))
);


ALTER TABLE public.driver_schedules OWNER TO fleet_user;

--
-- Name: TABLE driver_schedules; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.driver_schedules IS 'Normalized driver work schedule templates (weekly patterns). Replaces schedules JSONB in drivers table.';


--
-- Name: COLUMN driver_schedules.day_of_week; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_schedules.day_of_week IS 'Day of week: 0=Sunday, 1=Monday, ..., 6=Saturday';


--
-- Name: COLUMN driver_schedules.start_time; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_schedules.start_time IS 'Scheduled shift start time in HH:MM format (24h)';


--
-- Name: COLUMN driver_schedules.end_time; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_schedules.end_time IS 'Scheduled shift end time in HH:MM format (24h)';


--
-- Name: COLUMN driver_schedules.is_working; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.driver_schedules.is_working IS 'Whether the driver is scheduled to work on this day';


--
-- Name: driver_verifications; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.driver_verifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid NOT NULL,
    license_number character varying(50) NOT NULL,
    status character varying(20) NOT NULL,
    verification_date timestamp with time zone,
    name character varying(255) NOT NULL,
    father_name character varying(255),
    date_of_birth timestamp with time zone NOT NULL,
    gender character varying(10),
    address text,
    city character varying(100),
    district character varying(100),
    state character varying(100),
    pin character varying(10),
    license_class text[],
    issue_date timestamp with time zone NOT NULL,
    validity_from timestamp with time zone NOT NULL,
    validity_to timestamp with time zone NOT NULL,
    last_transaction character varying(255),
    enrollment_number character varying(100),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_by uuid
);


ALTER TABLE public.driver_verifications OWNER TO fleet_user;

--
-- Name: drivers; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.drivers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    org_id uuid NOT NULL,
    first_name character varying(50) NOT NULL,
    last_name character varying(50) NOT NULL,
    full_name character varying(100) GENERATED ALWAYS AS ((((first_name)::text || ' '::text) || (last_name)::text)) STORED,
    email character varying(100) NOT NULL,
    phone character varying(20) NOT NULL,
    license_number character varying(50) NOT NULL,
    license_type character varying(20),
    license_expiry timestamp with time zone,
    license_authority character varying(100),
    date_of_birth timestamp with time zone,
    address jsonb,
    emergency_contact jsonb,
    status character varying(20) DEFAULT 'active'::character varying,
    current_vehicle_id uuid,
    current_trip_id uuid,
    joining_date timestamp with time zone DEFAULT now(),
    resignation_date timestamp with time zone,
    experience_years integer DEFAULT 0,
    rating double precision DEFAULT 5.0,
    total_trips integer DEFAULT 0,
    total_distance_km double precision DEFAULT 0,
    total_hours_driven double precision DEFAULT 0,
    behavior_score integer DEFAULT 100,
    documents jsonb,
    attendance_records jsonb,
    schedules jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    created_by uuid,
    updated_by uuid,
    CONSTRAINT chk_drivers_license_type CHECK (((license_type)::text = ANY ((ARRAY['LMV'::character varying, 'HMV'::character varying, 'Transport'::character varying, 'Heavy'::character varying, 'Medium'::character varying, 'Light'::character varying])::text[]))),
    CONSTRAINT chk_drivers_status CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'suspended'::character varying, 'on_leave'::character varying, 'terminated'::character varying])::text[])))
);


ALTER TABLE public.drivers OWNER TO fleet_user;

--
-- Name: COLUMN drivers.documents; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.drivers.documents IS 'DEPRECATED: Use driver_documents table instead. Kept for backward compatibility during migration.';


--
-- Name: COLUMN drivers.attendance_records; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.drivers.attendance_records IS 'DEPRECATED: Use driver_attendance table. Kept for backward compatibility during migration.';


--
-- Name: COLUMN drivers.schedules; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.drivers.schedules IS 'DEPRECATED: Use driver_schedules table. Kept for backward compatibility during migration.';


--
-- Name: CONSTRAINT chk_drivers_status ON drivers; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_drivers_status ON public.drivers IS 'Enforces allowed values for status column';


--
-- Name: entity_audit_history; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.entity_audit_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    entity_type character varying(50) NOT NULL,
    entity_id character varying(100) NOT NULL,
    change_data jsonb NOT NULL,
    actor uuid,
    "timestamp" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.entity_audit_history OWNER TO fleet_user;

--
-- Name: evidence_exports; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.evidence_exports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    module character varying(50) NOT NULL,
    date_range_start timestamp with time zone,
    date_range_end timestamp with time zone,
    format character varying(20) DEFAULT 'pdf'::character varying,
    status character varying(20) DEFAULT 'processing'::character varying,
    requested_by uuid,
    file_url text,
    error_message text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.evidence_exports OWNER TO fleet_user;

--
-- Name: firmware_packages; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.firmware_packages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    version character varying(50) NOT NULL,
    hardware_target character varying(50) NOT NULL,
    file_url text NOT NULL,
    file_size integer,
    checksum character varying(64),
    release_notes text,
    status character varying(20) DEFAULT 'draft'::character varying,
    is_critical boolean DEFAULT false,
    released_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.firmware_packages OWNER TO fleet_user;

--
-- Name: firmware_rollout_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.firmware_rollout_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rollout_id uuid NOT NULL,
    org_id uuid NOT NULL,
    event_type character varying(64) NOT NULL,
    from_status character varying(32),
    to_status character varying(32),
    actor_user_id uuid,
    message text NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.firmware_rollout_events OWNER TO fleet_user;

--
-- Name: firmware_rollouts; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.firmware_rollouts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    firmware_id uuid,
    org_id uuid,
    name character varying(100),
    device_group character varying(100),
    target_filter jsonb,
    total_devices integer DEFAULT 0,
    completed_devices integer DEFAULT 0,
    failed_devices integer DEFAULT 0,
    status character varying(20) DEFAULT 'scheduled'::character varying,
    rollout_strategy character varying(20) DEFAULT 'gradual'::character varying,
    max_concurrent integer DEFAULT 10,
    created_by uuid,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_firmware_rollouts_status CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'scheduled'::character varying, 'in_progress'::character varying, 'paused'::character varying, 'completed'::character varying, 'cancelled'::character varying])::text[])))
);


ALTER TABLE public.firmware_rollouts OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_firmware_rollouts_status ON firmware_rollouts; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_firmware_rollouts_status ON public.firmware_rollouts IS 'Enforces allowed values for status column';


--
-- Name: fleet_analytics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.fleet_analytics (
    id character varying(36) NOT NULL,
    org_id uuid CONSTRAINT fleet_analytics_organization_id_not_null NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    total_vehicles integer DEFAULT 0 NOT NULL,
    active_vehicles integer DEFAULT 0 NOT NULL,
    total_distance numeric(12,2) DEFAULT 0 NOT NULL,
    average_speed numeric(8,2) DEFAULT 0 NOT NULL,
    total_fuel_consumption numeric(12,2) DEFAULT 0 NOT NULL,
    average_fuel_efficiency numeric(8,2) DEFAULT 0 NOT NULL,
    total_trips integer DEFAULT 0 NOT NULL,
    completed_trips integer DEFAULT 0 NOT NULL,
    cancelled_trips integer DEFAULT 0 NOT NULL,
    average_trip_duration numeric(8,2) DEFAULT 0 NOT NULL,
    idle_time_percentage numeric(5,2) DEFAULT 0 NOT NULL,
    routing_efficiency numeric(5,2) DEFAULT 0 NOT NULL,
    fuel_cost numeric(12,2) DEFAULT 0 NOT NULL,
    maintenance_cost numeric(12,2) DEFAULT 0 NOT NULL,
    revenue numeric(15,2) DEFAULT 0 NOT NULL,
    profit_margin numeric(5,2) DEFAULT 0 NOT NULL,
    safety_score numeric(5,2) DEFAULT 0 NOT NULL,
    accidents_count integer DEFAULT 0 NOT NULL,
    incidents_count integer DEFAULT 0 NOT NULL,
    emergency_stops integer DEFAULT 0 NOT NULL,
    geofence_violations integer DEFAULT 0 NOT NULL,
    alerts_triggered integer DEFAULT 0 NOT NULL,
    data_points_collected bigint DEFAULT 0 NOT NULL,
    processing_time numeric(10,3) DEFAULT 0 NOT NULL,
    generated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fleet_analytics_idle_time_percentage_check CHECK (((idle_time_percentage >= (0)::numeric) AND (idle_time_percentage <= (100)::numeric))),
    CONSTRAINT fleet_analytics_routing_efficiency_check CHECK (((routing_efficiency >= (0)::numeric) AND (routing_efficiency <= (100)::numeric))),
    CONSTRAINT fleet_analytics_safety_score_check CHECK (((safety_score >= (0)::numeric) AND (safety_score <= (100)::numeric)))
);


ALTER TABLE public.fleet_analytics OWNER TO fleet_user;

--
-- Name: geofences; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.geofences (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    org_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(20) NOT NULL,
    center jsonb,
    radius_meters integer,
    coordinates jsonb,
    route_polyline text,
    buffer_meters integer,
    entry_alert boolean DEFAULT true,
    exit_alert boolean DEFAULT true,
    speed_limit integer,
    vehicles jsonb,
    enabled boolean DEFAULT true,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.geofences OWNER TO fleet_user;

--
-- Name: global_notification_defaults; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.global_notification_defaults (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email_enabled boolean DEFAULT true NOT NULL,
    sms_enabled boolean DEFAULT false NOT NULL,
    push_enabled boolean DEFAULT true NOT NULL,
    in_app_enabled boolean DEFAULT true NOT NULL,
    updated_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.global_notification_defaults OWNER TO fleet_user;

--
-- Name: governance_settings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.governance_settings (
    org_id uuid NOT NULL,
    retention_policies jsonb DEFAULT '{}'::jsonb,
    masking_rules jsonb DEFAULT '{}'::jsonb,
    auto_expiry_enabled boolean DEFAULT false,
    download_ttl_hours integer DEFAULT 24,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_by uuid
);


ALTER TABLE public.governance_settings OWNER TO fleet_user;

--
-- Name: gps_locations; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.gps_locations (
    "time" timestamp with time zone NOT NULL,
    vehicle_id uuid NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    altitude double precision,
    speed double precision,
    heading double precision,
    accuracy double precision,
    ignition boolean,
    odometer double precision,
    battery_level double precision,
    satellites integer,
    raw_data jsonb
);


ALTER TABLE public.gps_locations OWNER TO fleet_user;

--
-- Name: TABLE gps_locations; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.gps_locations IS 'Stores raw GPS telemetry from vehicles';


--
-- Name: gst_compliances; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.gst_compliances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    gstin character varying(15) NOT NULL,
    status character varying(20) NOT NULL,
    legal_name character varying(255) NOT NULL,
    trade_name character varying(255),
    address text NOT NULL,
    centre_jurisdiction character varying(100),
    state_jurisdiction character varying(100),
    date_of_registration timestamp with time zone NOT NULL,
    date_of_cancellation timestamp with time zone,
    last_updated timestamp with time zone NOT NULL,
    return_filing_frequency character varying(20),
    return_filing_status character varying(20),
    payment_status character varying(20),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_by uuid
);


ALTER TABLE public.gst_compliances OWNER TO fleet_user;

--
-- Name: incident_comments; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.incident_comments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    ticket_id uuid NOT NULL,
    author_id uuid,
    author_name character varying(100),
    comment_type character varying(20) DEFAULT 'INTERNAL'::character varying,
    content text NOT NULL,
    is_private boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.incident_comments OWNER TO fleet_user;

--
-- Name: TABLE incident_comments; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.incident_comments IS 'Comments and history for support tickets';


--
-- Name: inspection_templates; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.inspection_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT inspection_templates_organization_id_not_null NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    applicable_vehicle_types character varying(50)[],
    checklist_items jsonb DEFAULT '[]'::jsonb NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.inspection_templates OWNER TO fleet_user;

--
-- Name: TABLE inspection_templates; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.inspection_templates IS 'Structure for digital inspection checklists';


--
-- Name: integration_configs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.integration_configs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    service_type character varying(50) NOT NULL,
    api_endpoint text NOT NULL,
    api_key text,
    username text,
    password text,
    client_id text,
    client_secret text,
    environment character varying(20) NOT NULL,
    active boolean DEFAULT true,
    last_sync_at timestamp with time zone,
    next_sync_at timestamp with time zone,
    sync_interval integer DEFAULT 3600,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_by uuid
);


ALTER TABLE public.integration_configs OWNER TO fleet_user;

--
-- Name: integration_logs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.integration_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    integration_id uuid,
    method character varying(10) NOT NULL,
    endpoint text NOT NULL,
    request_body jsonb,
    response_body jsonb,
    status_code integer,
    status character varying(20),
    duration_ms integer,
    "timestamp" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.integration_logs OWNER TO fleet_user;

--
-- Name: integrations; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.integrations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    type character varying(50) NOT NULL,
    provider character varying(100) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying,
    config jsonb DEFAULT '{}'::jsonb,
    last_sync timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.integrations OWNER TO fleet_user;

--
-- Name: invites; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.invites (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(50) NOT NULL,
    org_id uuid NOT NULL,
    role character varying(50) NOT NULL,
    expiry timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    used boolean DEFAULT false
);


ALTER TABLE public.invites OWNER TO fleet_user;

--
-- Name: invoice_items; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.invoice_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_id uuid NOT NULL,
    description text NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    unit_price numeric(15,2) DEFAULT 0.00 NOT NULL,
    tax_percent numeric(5,2) DEFAULT 0.00 NOT NULL,
    total numeric(15,2) DEFAULT 0.00 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    org_id uuid,
    amount double precision DEFAULT 0 NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    deleted_at timestamp with time zone
);


ALTER TABLE public.invoice_items OWNER TO fleet_user;

--
-- Name: invoices; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.invoices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    customer_name text NOT NULL,
    invoice_date date NOT NULL,
    due_date date NOT NULL,
    status text DEFAULT 'PENDING'::text NOT NULL,
    currency text DEFAULT 'USD'::text NOT NULL,
    total_amount numeric(15,2) DEFAULT 0.00 NOT NULL,
    amount_paid numeric(15,2) DEFAULT 0.00 NOT NULL,
    balance_due numeric(15,2) DEFAULT 0.00 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    invoice_number text,
    customer_email text,
    billing_address text,
    subtotal double precision DEFAULT 0 NOT NULL,
    tax_amount double precision DEFAULT 0 NOT NULL,
    amount_remaining double precision DEFAULT 0 NOT NULL,
    notes text,
    is_deleted boolean DEFAULT false NOT NULL,
    deleted_at timestamp with time zone,
    custom_fields jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT chk_invoices_status CHECK ((status = ANY (ARRAY['draft'::text, 'sent'::text, 'viewed'::text, 'paid'::text, 'overdue'::text, 'void'::text, 'refunded'::text])))
);


ALTER TABLE public.invoices OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_invoices_status ON invoices; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_invoices_status ON public.invoices IS 'Enforces allowed values for status column';


--
-- Name: ip_rules; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.ip_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    cidr character varying(45) NOT NULL,
    action character varying(10) NOT NULL,
    description text,
    added_by uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT ip_rules_action_check CHECK (((action)::text = ANY ((ARRAY['ALLOW'::character varying, 'BLOCK'::character varying])::text[])))
);


ALTER TABLE public.ip_rules OWNER TO fleet_user;

--
-- Name: local_permits; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.local_permits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vehicle_id uuid NOT NULL,
    driver_id uuid,
    org_id uuid NOT NULL,
    permit_number character varying(50) NOT NULL,
    permit_type character varying(50) NOT NULL,
    permit_category character varying(50),
    permit_sub_category character varying(50),
    valid_from timestamp with time zone NOT NULL,
    valid_to timestamp with time zone NOT NULL,
    issuing_authority character varying(255) NOT NULL,
    jurisdiction character varying(255),
    route_details text,
    number_of_trips integer,
    max_load_capacity double precision,
    number_of_seats integer,
    fare double precision,
    status character varying(20) NOT NULL,
    remarks text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_by uuid
);


ALTER TABLE public.local_permits OWNER TO fleet_user;

--
-- Name: login_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.login_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    user_id uuid,
    email character varying(255),
    ip_address character varying(45),
    device text,
    browser text,
    location character varying(255),
    status character varying(20),
    risk_level character varying(20) DEFAULT 'low'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT login_events_status_check CHECK (((status)::text = ANY ((ARRAY['SUCCESS'::character varying, 'FAILED'::character varying, 'CHALLENGED'::character varying])::text[])))
);


ALTER TABLE public.login_events OWNER TO fleet_user;

--
-- Name: maintenance_vendors; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.maintenance_vendors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT maintenance_vendors_organization_id_not_null NOT NULL,
    name character varying(255) NOT NULL,
    specialization character varying(100),
    rating numeric(3,2) DEFAULT 0.00,
    contact_info jsonb DEFAULT '{}'::jsonb,
    payment_terms character varying(100),
    status character varying(20) DEFAULT 'ACTIVE'::character varying,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT maintenance_vendors_rating_check CHECK (((rating >= (0)::numeric) AND (rating <= (5)::numeric)))
);


ALTER TABLE public.maintenance_vendors OWNER TO fleet_user;

--
-- Name: TABLE maintenance_vendors; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.maintenance_vendors IS 'Service providers and part suppliers';


--
-- Name: mfa_codes; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    code character varying(10) NOT NULL,
    method character varying(50) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.mfa_codes OWNER TO fleet_user;

--
-- Name: mfa_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    org_id uuid,
    event_type character varying(50) NOT NULL,
    status character varying(20) NOT NULL,
    method character varying(20),
    ip_address character varying(45),
    user_agent text,
    location character varying(255),
    details jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.mfa_events OWNER TO fleet_user;

--
-- Name: mfa_exceptions; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_exceptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    org_id uuid,
    reason text NOT NULL,
    granted_by uuid,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    start_date timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.mfa_exceptions OWNER TO fleet_user;

--
-- Name: mfa_failures; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_failures (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    org_id uuid,
    method character varying(20),
    failure_reason text,
    ip_address character varying(45),
    location character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.mfa_failures OWNER TO fleet_user;

--
-- Name: mfa_lockouts; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_lockouts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    org_id uuid,
    reason character varying(255),
    locked_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone,
    unlocked_at timestamp with time zone,
    unlocked_by uuid
);


ALTER TABLE public.mfa_lockouts OWNER TO fleet_user;

--
-- Name: mfa_org_overrides; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_org_overrides (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    enabled boolean DEFAULT true,
    allowed_methods jsonb DEFAULT '["totp", "sms"]'::jsonb,
    max_attempts integer DEFAULT 5,
    lockout_minutes integer DEFAULT 30,
    remember_device_days integer DEFAULT 30,
    grace_period_days integer DEFAULT 7,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.mfa_org_overrides OWNER TO fleet_user;

--
-- Name: mfa_policies; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_policies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    require_mfa boolean DEFAULT false,
    allowed_methods jsonb DEFAULT '["totp", "sms"]'::jsonb,
    max_attempts integer DEFAULT 5,
    lockout_minutes integer DEFAULT 30,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    grace_period_days integer DEFAULT 7,
    remember_device_days integer DEFAULT 30,
    enforce_at_login boolean DEFAULT true,
    admin_only_mode boolean DEFAULT false,
    remember_device_allowed boolean DEFAULT true,
    mfa_challenge_timeout integer DEFAULT 60,
    version integer DEFAULT 1
);


ALTER TABLE public.mfa_policies OWNER TO fleet_user;

--
-- Name: mfa_policy_history; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_policy_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    policy_id uuid,
    org_id uuid,
    snapshot jsonb NOT NULL,
    changed_by uuid,
    change_reason text,
    version integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.mfa_policy_history OWNER TO fleet_user;

--
-- Name: mfa_users; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.mfa_users (
    user_id uuid NOT NULL,
    org_id uuid,
    status character varying(20) DEFAULT 'DISABLED'::character varying,
    method character varying(50),
    trusted_devices_count integer DEFAULT 0,
    last_reset timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.mfa_users OWNER TO fleet_user;

--
-- Name: notification_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.notification_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    notification_id uuid NOT NULL,
    event_type text NOT NULL,
    message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notification_events OWNER TO fleet_user;

--
-- Name: notification_logs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.notification_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    correlation_id text,
    user_id uuid,
    user_email text,
    channel text NOT NULL,
    type text NOT NULL,
    subject text NOT NULL,
    body text,
    status text NOT NULL,
    provider text,
    provider_response text,
    retry_count integer DEFAULT 0 NOT NULL,
    max_retries integer DEFAULT 3 NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notification_logs OWNER TO fleet_user;

--
-- Name: notification_preference_audit; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.notification_preference_audit (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    changed_by uuid,
    old_values jsonb,
    new_values jsonb,
    change_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notification_preference_audit OWNER TO fleet_user;

--
-- Name: notification_settings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.notification_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    max_retries integer DEFAULT 3 NOT NULL,
    priority integer DEFAULT 100 NOT NULL,
    retention_months integer DEFAULT 6 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notification_settings OWNER TO fleet_user;

--
-- Name: obd_readings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.obd_readings (
    "time" timestamp with time zone NOT NULL,
    vehicle_id uuid NOT NULL,
    rpm integer,
    coolant_temp integer,
    engine_load numeric(5,2),
    throttle_pos numeric(5,2),
    fuel_level numeric(5,2),
    dtcs jsonb,
    vin character varying(50),
    raw_data jsonb
);


ALTER TABLE public.obd_readings OWNER TO fleet_user;

--
-- Name: TABLE obd_readings; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.obd_readings IS 'Stores On-Board Diagnostics data from vehicles';


--
-- Name: org_support_metrics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.org_support_metrics (
    org_id uuid NOT NULL,
    open_tickets integer DEFAULT 0,
    resolved_tickets integer DEFAULT 0,
    avg_resolution_time interval,
    sla_breaches integer DEFAULT 0,
    critical_issues integer DEFAULT 0,
    last_calculated timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.org_support_metrics OWNER TO fleet_user;

--
-- Name: TABLE org_support_metrics; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.org_support_metrics IS 'Pre-calculated organization support metrics';


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.organizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    domain character varying(255),
    status character varying(50) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    default_theme character varying(20) DEFAULT 'light'::character varying,
    is_deleted boolean DEFAULT false,
    contact_email character varying(255),
    plan character varying(50) DEFAULT 'enterprise'::character varying,
    theme_color text,
    theme_logo_url text,
    CONSTRAINT chk_organizations_status CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'suspended'::character varying, 'pending'::character varying, 'archived'::character varying])::text[]))),
    CONSTRAINT organizations_default_theme_check CHECK (((default_theme)::text = ANY ((ARRAY['light'::character varying, 'dark'::character varying, 'system'::character varying])::text[])))
);


ALTER TABLE public.organizations OWNER TO fleet_user;

--
-- Name: COLUMN organizations.id; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.organizations.id IS 'UUID primary key - referenced by org_id throughout the schema';


--
-- Name: CONSTRAINT chk_organizations_status ON organizations; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_organizations_status ON public.organizations IS 'Enforces allowed values for status column';


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: u0_a412
--

CREATE TABLE public.permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    resource character varying(100) NOT NULL,
    action character varying(50) NOT NULL,
    description text,
    bit_position integer NOT NULL
);


ALTER TABLE public.permissions OWNER TO u0_a412;

--
-- Name: prediction_results; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.prediction_results (
    id character varying(36) NOT NULL,
    model_id character varying(36) NOT NULL,
    prediction_input jsonb NOT NULL,
    predicted_value jsonb,
    confidence_score numeric(5,4) NOT NULL,
    probability_scores jsonb,
    feature_importance jsonb,
    anomaly_detected boolean DEFAULT false NOT NULL,
    anomaly_score numeric(5,4),
    explanation text,
    org_id uuid CONSTRAINT prediction_results_organization_id_not_null NOT NULL,
    request_id character varying(36) NOT NULL,
    predicted_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT prediction_results_anomaly_score_check CHECK (((anomaly_score >= (0)::numeric) AND (anomaly_score <= (1)::numeric))),
    CONSTRAINT prediction_results_confidence_score_check CHECK (((confidence_score >= (0)::numeric) AND (confidence_score <= (1)::numeric)))
);


ALTER TABLE public.prediction_results OWNER TO fleet_user;

--
-- Name: predictive_maintenance_analytics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.predictive_maintenance_analytics (
    id character varying(36) NOT NULL,
    vehicle_id uuid NOT NULL,
    org_id uuid CONSTRAINT predictive_maintenance_analytics_organization_id_not_null NOT NULL,
    analysis_date date NOT NULL,
    next_service_due date NOT NULL,
    service_priority character varying(20) NOT NULL,
    component_risks jsonb DEFAULT '{}'::jsonb NOT NULL,
    failure_prediction jsonb,
    maintenance_schedule jsonb DEFAULT '[]'::jsonb NOT NULL,
    parts_needed jsonb DEFAULT '[]'::jsonb NOT NULL,
    cost_estimation numeric(12,2) DEFAULT 0 NOT NULL,
    uptime_prediction numeric(5,2) DEFAULT 0 NOT NULL,
    reliability_score numeric(5,2) DEFAULT 0 NOT NULL,
    recommendations jsonb DEFAULT '[]'::jsonb NOT NULL,
    generated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT predictive_maintenance_analytics_reliability_score_check CHECK (((reliability_score >= (0)::numeric) AND (reliability_score <= (100)::numeric))),
    CONSTRAINT predictive_maintenance_analytics_uptime_prediction_check CHECK (((uptime_prediction >= (0)::numeric) AND (uptime_prediction <= (100)::numeric)))
);


ALTER TABLE public.predictive_maintenance_analytics OWNER TO fleet_user;

--
-- Name: predictive_models; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.predictive_models (
    id character varying(36) NOT NULL,
    name character varying(255) NOT NULL,
    version character varying(50) NOT NULL,
    model_type character varying(100) NOT NULL,
    description text,
    algorithm character varying(100) NOT NULL,
    input_features jsonb DEFAULT '[]'::jsonb NOT NULL,
    output_variable character varying(255),
    training_data_start timestamp with time zone NOT NULL,
    training_data_end timestamp with time zone NOT NULL,
    accuracy numeric(5,4),
    "precision" numeric(5,4),
    recall numeric(5,4),
    f1_score numeric(5,4),
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    trained_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT predictive_models_accuracy_check CHECK (((accuracy >= (0)::numeric) AND (accuracy <= (1)::numeric))),
    CONSTRAINT predictive_models_f1_score_check CHECK (((f1_score >= (0)::numeric) AND (f1_score <= (1)::numeric))),
    CONSTRAINT predictive_models_precision_check CHECK ((("precision" >= (0)::numeric) AND ("precision" <= (1)::numeric))),
    CONSTRAINT predictive_models_recall_check CHECK (((recall >= (0)::numeric) AND (recall <= (1)::numeric)))
);


ALTER TABLE public.predictive_models OWNER TO fleet_user;

--
-- Name: restore_jobs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.restore_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    backup_id uuid NOT NULL,
    status character varying(50) NOT NULL,
    target_environment character varying(50) NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    performed_by uuid NOT NULL,
    logs text
);


ALTER TABLE public.restore_jobs OWNER TO fleet_user;

--
-- Name: review_items; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.review_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    review_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid,
    current_status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    reviewer_id uuid,
    reviewed_at timestamp with time zone,
    notes text,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT review_items_current_status_check CHECK (((current_status)::text = ANY ((ARRAY['PENDING'::character varying, 'APPROVED'::character varying, 'REVOKED'::character varying])::text[])))
);


ALTER TABLE public.review_items OWNER TO fleet_user;

--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: u0_a412
--

CREATE TABLE public.role_permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


ALTER TABLE public.role_permissions OWNER TO u0_a412;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(50) NOT NULL,
    description text,
    is_system boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    org_id uuid,
    permission_mask bigint DEFAULT 0,
    created_by uuid,
    deleted_at timestamp with time zone
);


ALTER TABLE public.roles OWNER TO fleet_user;

--
-- Name: route_optimization_jobs; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.route_optimization_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    request_id uuid NOT NULL,
    status character varying(50) DEFAULT 'queued'::character varying NOT NULL,
    worker_id character varying(100),
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_route_optimization_jobs_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'completed'::character varying, 'failed'::character varying, 'cancelled'::character varying])::text[])))
);


ALTER TABLE public.route_optimization_jobs OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_route_optimization_jobs_status ON route_optimization_jobs; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_route_optimization_jobs_status ON public.route_optimization_jobs IS 'Enforces allowed values for status column';


--
-- Name: route_optimization_requests; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.route_optimization_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT route_optimization_requests_organization_id_not_null NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    constraints jsonb DEFAULT '{}'::jsonb,
    stops jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.route_optimization_requests OWNER TO fleet_user;

--
-- Name: route_optimization_results; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.route_optimization_results (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    request_id uuid NOT NULL,
    routes jsonb DEFAULT '[]'::jsonb NOT NULL,
    total_distance numeric(10,2) DEFAULT 0,
    total_duration numeric(10,2) DEFAULT 0,
    total_cost numeric(10,2) DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.route_optimization_results OWNER TO fleet_user;

--
-- Name: security_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.security_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    event_type character varying(100) NOT NULL,
    ip_address character varying(45),
    user_agent text,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    actor_id uuid,
    on_behalf_of uuid
);


ALTER TABLE public.security_events OWNER TO fleet_user;

--
-- Name: sla_metrics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.sla_metrics (
    "time" timestamp with time zone NOT NULL,
    org_id uuid NOT NULL,
    metric_name character varying(50) NOT NULL,
    value numeric(15,4) NOT NULL,
    metadata jsonb
);


ALTER TABLE public.sla_metrics OWNER TO fleet_user;

--
-- Name: storage_pools; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.storage_pools (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    status character varying(50) DEFAULT 'ACTIVE'::character varying NOT NULL,
    total_bytes bigint DEFAULT 0 NOT NULL,
    used_bytes bigint DEFAULT 0 NOT NULL,
    available_bytes bigint DEFAULT 0 NOT NULL,
    location character varying(255),
    backup_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.storage_pools OWNER TO fleet_user;

--
-- Name: support_audit_log; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.support_audit_log (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    ticket_id uuid,
    user_id uuid,
    action character varying(100) NOT NULL,
    old_value jsonb,
    new_value jsonb,
    details text,
    created_at timestamp with time zone DEFAULT now(),
    ip_address character varying(45)
);


ALTER TABLE public.support_audit_log OWNER TO fleet_user;

--
-- Name: TABLE support_audit_log; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.support_audit_log IS 'Audit trail for all support operations';


--
-- Name: support_tickets; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.support_tickets (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    ticket_id character varying(50) NOT NULL,
    org_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    status character varying(50) DEFAULT 'OPEN'::character varying NOT NULL,
    priority character varying(20) DEFAULT 'MEDIUM'::character varying NOT NULL,
    category character varying(100),
    assigned_to uuid,
    reporter_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    resolved_at timestamp with time zone,
    sla_deadline timestamp with time zone,
    tags jsonb DEFAULT '[]'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT chk_support_tickets_status CHECK (((status)::text = ANY ((ARRAY['open'::character varying, 'in_progress'::character varying, 'waiting_customer'::character varying, 'resolved'::character varying, 'closed'::character varying, 'reopened'::character varying])::text[])))
);


ALTER TABLE public.support_tickets OWNER TO fleet_user;

--
-- Name: TABLE support_tickets; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.support_tickets IS 'Core support ticket records';


--
-- Name: CONSTRAINT chk_support_tickets_status ON support_tickets; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_support_tickets_status ON public.support_tickets IS 'Enforces allowed values for status column';


--
-- Name: telemetry; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    trip_id uuid,
    latitude numeric(10,8),
    longitude numeric(11,8),
    speed numeric(6,2),
    heading numeric(5,2),
    event_type character varying(50),
    event_time timestamp without time zone DEFAULT now()
);


ALTER TABLE public.telemetry OWNER TO fleet_user;

--
-- Name: telemetry_analytics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_analytics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vehicle_id uuid NOT NULL,
    analysis_type character varying(100) NOT NULL,
    data jsonb NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now(),
    confidence numeric(3,2) NOT NULL,
    CONSTRAINT telemetry_analytics_confidence_check CHECK (((confidence >= (0)::numeric) AND (confidence <= (1)::numeric)))
);


ALTER TABLE public.telemetry_analytics OWNER TO fleet_user;

--
-- Name: TABLE telemetry_analytics; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.telemetry_analytics IS 'Stores advanced analytics results for vehicle telemetry data';


--
-- Name: COLUMN telemetry_analytics.analysis_type; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.telemetry_analytics.analysis_type IS 'Type of analysis performed (e.g., speed_analysis, fuel_efficiency)';


--
-- Name: COLUMN telemetry_analytics.confidence; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.telemetry_analytics.confidence IS 'Confidence score for the analysis results (0-1)';


--
-- Name: telemetry_anomalies; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_anomalies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vehicle_id uuid NOT NULL,
    anomaly_type character varying(100) NOT NULL,
    data jsonb NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now(),
    confidence numeric(3,2) NOT NULL,
    anomaly_score numeric(5,2) NOT NULL,
    recommendation text,
    CONSTRAINT telemetry_anomalies_confidence_check CHECK (((confidence >= (0)::numeric) AND (confidence <= (1)::numeric)))
);


ALTER TABLE public.telemetry_anomalies OWNER TO fleet_user;

--
-- Name: TABLE telemetry_anomalies; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.telemetry_anomalies IS 'Stores detected anomalies in vehicle telemetry data';


--
-- Name: COLUMN telemetry_anomalies.anomaly_type; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.telemetry_anomalies.anomaly_type IS 'Type of anomaly detected (e.g., unusual_speed, fuel_theft)';


--
-- Name: COLUMN telemetry_anomalies.anomaly_score; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.telemetry_anomalies.anomaly_score IS 'Numerical score representing severity of the anomaly';


--
-- Name: COLUMN telemetry_anomalies.recommendation; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.telemetry_anomalies.recommendation IS 'Suggested action for addressing the anomaly';


--
-- Name: telemetry_events; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_events (
    id uuid NOT NULL,
    vehicle_id uuid NOT NULL,
    org_id uuid NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    speed real,
    heading real,
    fuel_level real,
    engine_rpm integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
)
PARTITION BY RANGE (created_at);


ALTER TABLE public.telemetry_events OWNER TO fleet_user;

--
-- Name: telemetry_events_p2026_04; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_events_p2026_04 (
    id uuid CONSTRAINT telemetry_events_id_not_null NOT NULL,
    vehicle_id uuid CONSTRAINT telemetry_events_vehicle_id_not_null NOT NULL,
    org_id uuid CONSTRAINT telemetry_events_org_id_not_null NOT NULL,
    latitude double precision CONSTRAINT telemetry_events_latitude_not_null NOT NULL,
    longitude double precision CONSTRAINT telemetry_events_longitude_not_null NOT NULL,
    speed real,
    heading real,
    fuel_level real,
    engine_rpm integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP CONSTRAINT telemetry_events_created_at_not_null NOT NULL
);


ALTER TABLE public.telemetry_events_p2026_04 OWNER TO fleet_user;

--
-- Name: telemetry_events_p2026_05; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_events_p2026_05 (
    id uuid CONSTRAINT telemetry_events_id_not_null NOT NULL,
    vehicle_id uuid CONSTRAINT telemetry_events_vehicle_id_not_null NOT NULL,
    org_id uuid CONSTRAINT telemetry_events_org_id_not_null NOT NULL,
    latitude double precision CONSTRAINT telemetry_events_latitude_not_null NOT NULL,
    longitude double precision CONSTRAINT telemetry_events_longitude_not_null NOT NULL,
    speed real,
    heading real,
    fuel_level real,
    engine_rpm integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP CONSTRAINT telemetry_events_created_at_not_null NOT NULL
);


ALTER TABLE public.telemetry_events_p2026_05 OWNER TO fleet_user;

--
-- Name: telemetry_events_p2026_06; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_events_p2026_06 (
    id uuid CONSTRAINT telemetry_events_id_not_null NOT NULL,
    vehicle_id uuid CONSTRAINT telemetry_events_vehicle_id_not_null NOT NULL,
    org_id uuid CONSTRAINT telemetry_events_org_id_not_null NOT NULL,
    latitude double precision CONSTRAINT telemetry_events_latitude_not_null NOT NULL,
    longitude double precision CONSTRAINT telemetry_events_longitude_not_null NOT NULL,
    speed real,
    heading real,
    fuel_level real,
    engine_rpm integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP CONSTRAINT telemetry_events_created_at_not_null NOT NULL
);


ALTER TABLE public.telemetry_events_p2026_06 OWNER TO fleet_user;

--
-- Name: telemetry_events_p2026_07; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.telemetry_events_p2026_07 (
    id uuid CONSTRAINT telemetry_events_id_not_null NOT NULL,
    vehicle_id uuid CONSTRAINT telemetry_events_vehicle_id_not_null NOT NULL,
    org_id uuid CONSTRAINT telemetry_events_org_id_not_null NOT NULL,
    latitude double precision CONSTRAINT telemetry_events_latitude_not_null NOT NULL,
    longitude double precision CONSTRAINT telemetry_events_longitude_not_null NOT NULL,
    speed real,
    heading real,
    fuel_level real,
    engine_rpm integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP CONSTRAINT telemetry_events_created_at_not_null NOT NULL
);


ALTER TABLE public.telemetry_events_p2026_07 OWNER TO fleet_user;

--
-- Name: threat_alerts; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.threat_alerts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid,
    type character varying(50) NOT NULL,
    severity character varying(20) NOT NULL,
    ip character varying(45),
    user_id uuid,
    description text,
    status character varying(20) DEFAULT 'open'::character varying,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.threat_alerts OWNER TO fleet_user;

--
-- Name: trip_settings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.trip_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT trip_settings_organization_id_not_null NOT NULL,
    retention_period integer DEFAULT 36,
    policies jsonb DEFAULT '{}'::jsonb,
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.trip_settings OWNER TO fleet_user;

--
-- Name: trips; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.trips (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    org_id uuid NOT NULL,
    vehicle_id uuid,
    driver_id uuid,
    status character varying(50) DEFAULT 'planned'::character varying NOT NULL,
    title character varying(255),
    description text,
    planned_route jsonb,
    actual_route jsonb,
    waypoints jsonb,
    start_location jsonb,
    end_location jsonb,
    estimated_distance_km numeric(10,2),
    actual_distance_km numeric(10,2),
    estimated_duration_min integer,
    actual_duration_min integer,
    fuel_consumed_liters numeric(10,2) DEFAULT 0,
    idling_time_minutes integer DEFAULT 0,
    cargo_details jsonb,
    pod_documents jsonb,
    scheduled_start timestamp with time zone,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    customer_name character varying(255),
    customer_phone character varying(50),
    customer_email character varying(255),
    cancellation_reason text,
    events jsonb,
    is_deleted boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT chk_trips_status CHECK (((status)::text = ANY ((ARRAY['planned'::character varying, 'scheduled'::character varying, 'in_progress'::character varying, 'completed'::character varying, 'cancelled'::character varying, 'delayed'::character varying, 'no_show'::character varying, 'emergency'::character varying])::text[])))
);


ALTER TABLE public.trips OWNER TO fleet_user;

--
-- Name: COLUMN trips.org_id; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.trips.org_id IS 'Organization UUID - foreign key to organizations(id)';


--
-- Name: CONSTRAINT chk_trips_status ON trips; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_trips_status ON public.trips IS 'Enforces allowed values for status column';


--
-- Name: usage_meters; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.usage_meters (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    unit character varying(20) NOT NULL,
    unit_price numeric(15,4) DEFAULT 0.0000 NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.usage_meters OWNER TO fleet_user;

--
-- Name: user_links; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.user_links (
    parent_user_id uuid NOT NULL,
    linked_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    CONSTRAINT no_self_link CHECK ((parent_user_id <> linked_user_id))
);


ALTER TABLE public.user_links OWNER TO fleet_user;

--
-- Name: user_notification_preferences; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.user_notification_preferences (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    email_enabled boolean DEFAULT true NOT NULL,
    sms_enabled boolean DEFAULT false NOT NULL,
    push_enabled boolean DEFAULT true NOT NULL,
    in_app_enabled boolean DEFAULT true NOT NULL,
    last_modified_by uuid,
    last_modified_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.user_notification_preferences OWNER TO fleet_user;

--
-- Name: user_organizations; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.user_organizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    org_id uuid,
    role character varying(50) DEFAULT 'member'::character varying,
    status character varying(50) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    role_id uuid,
    is_active boolean DEFAULT true
);


ALTER TABLE public.user_organizations OWNER TO fleet_user;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: u0_a412
--

CREATE TABLE public.user_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    org_id uuid NOT NULL,
    assigned_by uuid,
    assigned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.user_roles OWNER TO u0_a412;

--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.user_sessions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    session_id character varying(100) NOT NULL,
    ip_address character varying(45),
    user_agent text,
    device_info jsonb,
    login_time timestamp with time zone DEFAULT now(),
    last_activity timestamp with time zone DEFAULT now(),
    logout_time timestamp with time zone,
    is_active boolean DEFAULT true,
    location jsonb
);


ALTER TABLE public.user_sessions OWNER TO fleet_user;

--
-- Name: TABLE user_sessions; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.user_sessions IS 'User session tracking for support investigations';


--
-- Name: users; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    password_hash text NOT NULL,
    name character varying(255),
    phone character varying(50),
    mfa_enabled boolean DEFAULT false,
    mfa_secret text,
    status character varying(50) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    theme_updated_at timestamp with time zone,
    is_deleted boolean DEFAULT false,
    role_id uuid NOT NULL,
    permission_mask bigint DEFAULT 0,
    sidebar_collapsed boolean DEFAULT false NOT NULL,
    language character varying(10) DEFAULT 'en'::character varying NOT NULL,
    avatar_url text,
    mfa_backup_codes jsonb DEFAULT '[]'::jsonb,
    email_verified boolean DEFAULT false,
    failed_logins integer DEFAULT 0,
    locked_until timestamp without time zone,
    last_login_at timestamp without time zone,
    last_login_ip text,
    CONSTRAINT chk_users_status CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'suspended'::character varying, 'pending'::character varying, 'locked'::character varying])::text[])))
);


ALTER TABLE public.users OWNER TO fleet_user;

--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.users IS 'User accounts with multi-tenant org_id and RBAC role_id. org_id and role_id are both NOT NULL with proper foreign keys.';


--
-- Name: COLUMN users.phone; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.phone IS 'User phone number';


--
-- Name: COLUMN users.mfa_enabled; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.mfa_enabled IS 'Whether MFA is enabled';


--
-- Name: COLUMN users.mfa_secret; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.mfa_secret IS 'Encrypted MFA TOTP secret';


--
-- Name: COLUMN users.role_id; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.role_id IS 'Role UUID - NOT NULL, foreign key to roles(id) - for RBAC';


--
-- Name: COLUMN users.permission_mask; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.permission_mask IS 'Bitmask for user permissions';


--
-- Name: COLUMN users.sidebar_collapsed; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.sidebar_collapsed IS 'UI preference for sidebar state';


--
-- Name: COLUMN users.language; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.language IS 'User preferred language';


--
-- Name: COLUMN users.avatar_url; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.avatar_url IS 'URL or path to user avatar/profile image';


--
-- Name: COLUMN users.mfa_backup_codes; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.mfa_backup_codes IS 'Encrypted MFA backup codes';


--
-- Name: COLUMN users.email_verified; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.email_verified IS 'Whether email has been verified';


--
-- Name: COLUMN users.failed_logins; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.failed_logins IS 'Count of consecutive failed login attempts';


--
-- Name: COLUMN users.locked_until; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.locked_until IS 'Account lockout expiration time';


--
-- Name: COLUMN users.last_login_at; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.last_login_at IS 'Timestamp of last successful login';


--
-- Name: COLUMN users.last_login_ip; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.users.last_login_ip IS 'IP address of last successful login';


--
-- Name: CONSTRAINT chk_users_status ON users; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_users_status ON public.users IS 'Enforces allowed values for status column';


--
-- Name: vehicle_inspections; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.vehicle_inspections (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT vehicle_inspections_organization_id_not_null NOT NULL,
    vehicle_id uuid NOT NULL,
    template_id uuid,
    inspector_id uuid NOT NULL,
    status character varying(20) DEFAULT 'COMPLETED'::character varying,
    results jsonb DEFAULT '[]'::jsonb NOT NULL,
    notes text,
    odometer_at_inspection bigint,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    created_by uuid,
    updated_by uuid
);


ALTER TABLE public.vehicle_inspections OWNER TO fleet_user;

--
-- Name: TABLE vehicle_inspections; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.vehicle_inspections IS 'Completed vehicle inspection reports';


--
-- Name: vehicle_telemetry; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.vehicle_telemetry (
    "time" timestamp with time zone NOT NULL,
    vehicle_id uuid NOT NULL,
    latitude double precision,
    longitude double precision,
    speed_kph double precision,
    heading double precision,
    odometer_km double precision,
    fuel_level_percent double precision,
    engine_temp_celsius double precision,
    alerts jsonb DEFAULT '[]'::jsonb
);


ALTER TABLE public.vehicle_telemetry OWNER TO fleet_user;

--
-- Name: TABLE vehicle_telemetry; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.vehicle_telemetry IS 'Aggregated telemetry data for trip reporting';


--
-- Name: vehicle_verifications; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.vehicle_verifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vehicle_id uuid NOT NULL,
    org_id uuid NOT NULL,
    registration_number character varying(20) NOT NULL,
    status character varying(20) NOT NULL,
    verification_date timestamp with time zone,
    chassis_number character varying(50),
    engine_number character varying(50),
    model character varying(100),
    manufacturer character varying(100),
    fuel_type character varying(20),
    body_type character varying(50),
    cubic_capacity integer,
    seating_capacity integer,
    color character varying(30),
    manufacturing_year integer,
    owner_name character varying(255),
    owner_address text,
    owner_city character varying(100),
    owner_district character varying(100),
    owner_state character varying(100),
    owner_pin character varying(10),
    rc_status character varying(20),
    insurance_company character varying(100),
    insurance_policy character varying(100),
    insurance_expiry timestamp with time zone,
    puc_number character varying(100),
    puc_expiry timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_by uuid
);


ALTER TABLE public.vehicle_verifications OWNER TO fleet_user;

--
-- Name: vehicles; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.vehicles (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    org_id uuid NOT NULL,
    license_plate character varying(20) NOT NULL,
    model character varying(50) NOT NULL,
    make character varying(50) NOT NULL,
    year integer NOT NULL,
    color character varying(30),
    status character varying(20) DEFAULT 'active'::character varying,
    capacity integer DEFAULT 1,
    fuel_type character varying(20),
    vin character varying(50),
    registration character varying(50),
    insurance character varying(50),
    odometer bigint DEFAULT 0,
    last_service timestamp with time zone,
    current_driver_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    vehicle_no character varying(50),
    service_due timestamp with time zone,
    notes text,
    created_by uuid,
    updated_by uuid,
    tenant_id uuid,
    CONSTRAINT chk_vehicles_fuel_type CHECK (((fuel_type)::text = ANY ((ARRAY['petrol'::character varying, 'diesel'::character varying, 'electric'::character varying, 'hybrid'::character varying, 'cng'::character varying, 'lpg'::character varying])::text[]))),
    CONSTRAINT chk_vehicles_status CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'maintenance'::character varying, 'retired'::character varying, 'out_of_service'::character varying])::text[])))
);


ALTER TABLE public.vehicles OWNER TO fleet_user;

--
-- Name: CONSTRAINT chk_vehicles_status ON vehicles; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT chk_vehicles_status ON public.vehicles IS 'Enforces allowed values for status column';


--
-- Name: video_analytics; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.video_analytics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    recording_id uuid NOT NULL,
    stream_id uuid,
    objects_detected jsonb DEFAULT '[]'::jsonb NOT NULL,
    behavior_events jsonb DEFAULT '[]'::jsonb NOT NULL,
    traffic_signs jsonb DEFAULT '[]'::jsonb NOT NULL,
    incidents jsonb DEFAULT '[]'::jsonb NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    confidence numeric(3,2) NOT NULL,
    processed_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT video_analytics_confidence_check CHECK (((confidence >= (0)::numeric) AND (confidence <= (1)::numeric)))
);


ALTER TABLE public.video_analytics OWNER TO fleet_user;

--
-- Name: TABLE video_analytics; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.video_analytics IS 'Stores AI-powered insights extracted from video analysis';


--
-- Name: COLUMN video_analytics.objects_detected; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_analytics.objects_detected IS 'List of objects detected in the video';


--
-- Name: COLUMN video_analytics.behavior_events; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_analytics.behavior_events IS 'List of behavioral events detected in the video';


--
-- Name: COLUMN video_analytics.incidents; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_analytics.incidents IS 'List of incidents detected in the video';


--
-- Name: COLUMN video_analytics.confidence; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_analytics.confidence IS 'Confidence score for the analytics results (0-1)';


--
-- Name: video_recordings; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.video_recordings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    stream_id uuid NOT NULL,
    camera_id character varying(255) NOT NULL,
    vehicle_id uuid NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid CONSTRAINT video_recordings_organization_id_not_null NOT NULL,
    filename character varying(500) NOT NULL,
    filepath text NOT NULL,
    file_size bigint NOT NULL,
    duration integer NOT NULL,
    resolution character varying(20),
    framerate integer,
    bitrate integer,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone NOT NULL,
    location_start jsonb NOT NULL,
    location_end jsonb NOT NULL,
    status character varying(50) DEFAULT 'processing'::character varying NOT NULL,
    thumbnail_path text,
    tags jsonb DEFAULT '[]'::jsonb,
    event_type character varying(100),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.video_recordings OWNER TO fleet_user;

--
-- Name: TABLE video_recordings; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.video_recordings IS 'Stores information about recorded video clips';


--
-- Name: COLUMN video_recordings.status; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_recordings.status IS 'Status of the video recording (recorded, processing, available, archived, deleted)';


--
-- Name: COLUMN video_recordings.event_type; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_recordings.event_type IS 'Type of event that triggered the recording (sudden_stop, harsh_acceleration, etc.)';


--
-- Name: video_retention_policies; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.video_retention_policies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT video_retention_policies_organization_id_not_null NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    event_type character varying(100),
    retention_period integer NOT NULL,
    auto_archive boolean DEFAULT false,
    auto_delete boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.video_retention_policies OWNER TO fleet_user;

--
-- Name: TABLE video_retention_policies; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.video_retention_policies IS 'Stores retention policies for video data';


--
-- Name: COLUMN video_retention_policies.retention_period; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_retention_policies.retention_period IS 'Number of days to retain videos before archiving/deletion';


--
-- Name: COLUMN video_retention_policies.auto_archive; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_retention_policies.auto_archive IS 'Flag to automatically archive videos after retention period';


--
-- Name: COLUMN video_retention_policies.auto_delete; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_retention_policies.auto_delete IS 'Flag to automatically delete videos after retention period';


--
-- Name: video_streams; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.video_streams (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    camera_id character varying(255) NOT NULL,
    vehicle_id uuid NOT NULL,
    driver_id uuid NOT NULL,
    org_id uuid CONSTRAINT video_streams_organization_id_not_null NOT NULL,
    url text NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    resolution character varying(20),
    framerate integer,
    bitrate integer,
    protocol character varying(20),
    location jsonb NOT NULL,
    started_at timestamp with time zone NOT NULL,
    ended_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    analytics_enabled boolean DEFAULT false
);


ALTER TABLE public.video_streams OWNER TO fleet_user;

--
-- Name: TABLE video_streams; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.video_streams IS 'Stores information about live video streams from vehicles';


--
-- Name: COLUMN video_streams.status; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_streams.status IS 'Status of the video stream (active, inactive, error)';


--
-- Name: COLUMN video_streams.analytics_enabled; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.video_streams.analytics_enabled IS 'Flag to enable AI-powered analytics on this stream';


--
-- Name: video_thumbnails; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.video_thumbnails (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    recording_id uuid NOT NULL,
    image_path text NOT NULL,
    image_url text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.video_thumbnails OWNER TO fleet_user;

--
-- Name: TABLE video_thumbnails; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON TABLE public.video_thumbnails IS 'Stores thumbnail images for video recordings';


--
-- Name: webhooks; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.webhooks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid NOT NULL,
    url text NOT NULL,
    events text[] DEFAULT '{}'::text[],
    secret character varying(255),
    status character varying(20) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.webhooks OWNER TO fleet_user;

--
-- Name: work_orders; Type: TABLE; Schema: public; Owner: fleet_user
--

CREATE TABLE public.work_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    org_id uuid CONSTRAINT work_orders_organization_id_not_null NOT NULL,
    vehicle_id uuid,
    title text NOT NULL,
    description text,
    maintenance_type text DEFAULT 'PREVENTIVE'::text NOT NULL,
    priority text DEFAULT 'MEDIUM'::text NOT NULL,
    status text DEFAULT 'OPEN'::text NOT NULL,
    assigned_to uuid,
    scheduled_date timestamp with time zone,
    completed_date timestamp with time zone,
    estimated_cost double precision DEFAULT 0 NOT NULL,
    actual_cost double precision DEFAULT 0 NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    vendor_id uuid,
    approval_status text DEFAULT 'PENDING'::text NOT NULL,
    cost_breakdown jsonb DEFAULT '{}'::jsonb NOT NULL,
    inspection_id uuid,
    updated_by uuid
);


ALTER TABLE public.work_orders OWNER TO fleet_user;

--
-- Name: COLUMN work_orders.org_id; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON COLUMN public.work_orders.org_id IS 'Organization UUID - foreign key to organizations(id)';


--
-- Name: telemetry_events_p2026_04; Type: TABLE ATTACH; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events ATTACH PARTITION public.telemetry_events_p2026_04 FOR VALUES FROM ('2026-04-01 00:00:00+05:30') TO ('2026-05-01 00:00:00+05:30');


--
-- Name: telemetry_events_p2026_05; Type: TABLE ATTACH; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events ATTACH PARTITION public.telemetry_events_p2026_05 FOR VALUES FROM ('2026-05-01 00:00:00+05:30') TO ('2026-06-01 00:00:00+05:30');


--
-- Name: telemetry_events_p2026_06; Type: TABLE ATTACH; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events ATTACH PARTITION public.telemetry_events_p2026_06 FOR VALUES FROM ('2026-06-01 00:00:00+05:30') TO ('2026-07-01 00:00:00+05:30');


--
-- Name: telemetry_events_p2026_07; Type: TABLE ATTACH; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events ATTACH PARTITION public.telemetry_events_p2026_07 FOR VALUES FROM ('2026-07-01 00:00:00+05:30') TO ('2026-08-01 00:00:00+05:30');


--
-- Data for Name: access_reviews; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.access_reviews (id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: alert_rules; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.alert_rules (id, org_id, name, type, description, severity, conditions, actions, vehicles, driver_ids, enabled, created_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: alerts; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.alerts (id, org_id, rule_id, type, severity, vehicle_id, driver_id, "timestamp", location, details, status, notes, acknowledged_by, acknowledged_at, resolved_by, resolved_at, created_at, updated_at, dedup_key, escalation_level, last_escalated_at, escalated_to, sla_deadline) FROM stdin;
\.


--
-- Data for Name: analytics_jobs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.analytics_jobs (id, name, type, description, status, job_type, parameters, input_source, output_destination, start_time, end_time, progress, error_details, org_id, created_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: api_clients; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.api_clients (id, org_id, client_id, client_secret_hash, name, scopes, expiry, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.audit_logs (id, user_id, action, resource, details, ip_address, user_agent, created_at, severity, status, org_id, module, entity_id, before_state, after_state, metadata) FROM stdin;
\.


--
-- Data for Name: audit_policies; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.audit_policies (id, org_id, retention_years, archive_policy, masking_rules, updated_by, updated_at) FROM stdin;
\.


--
-- Data for Name: audit_settings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.audit_settings (id, org_id, retention_days_online, retention_days_archival, masking_rules, tracked_event_types, alert_on_critical, updated_at) FROM stdin;
\.


--
-- Data for Name: auth_tokens; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.auth_tokens (id, user_id, access_token, refresh_token, device_id, expires_at, created_at, ip_address, user_agent, last_active, is_active) FROM stdin;
\.


--
-- Data for Name: backup_verifications; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.backup_verifications (id, backup_id, status, checksum, verified_at, verified_by, error_message, created_at) FROM stdin;
\.


--
-- Data for Name: backups; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.backups (id, type, status, s3_key, size_bytes, created_by, created_at, completed_at, metadata) FROM stdin;
\.


--
-- Data for Name: billing_accounting_exports; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_accounting_exports (id, org_id, external_system, export_type, status, file_key, records_count, error_message, created_at, completed_at, created_by) FROM stdin;
\.


--
-- Data for Name: billing_coupons; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_coupons (id, code, name, discount_type, discount_value, currency, max_redemptions, redeemed_count, valid_from, valid_until, status, created_at, created_by) FROM stdin;
\.


--
-- Data for Name: billing_credit_transactions; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_credit_transactions (id, balance_id, transaction_type, amount, currency, reference_type, reference_id, description, created_at, created_by) FROM stdin;
\.


--
-- Data for Name: billing_customer_balances; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_customer_balances (id, org_id, currency, balance, credit_limit, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: billing_hardware_products; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_hardware_products (id, name, model_code, category, one_time_price, lease_monthly_price, currency, metadata, created_at, updated_at) FROM stdin;
7d6b766e-57e0-4ad2-947f-a95bd1ed1910	FlyFleet Vision Pro Dashcam	FF-CAM-V2	DASHCAM	399.00	25.00	USD	{}	2026-04-14 15:47:41.562924+05:30	2026-04-14 15:47:41.562924+05:30
1cb68d04-7258-4b58-b3c5-3200550ae5dd	FlyFleet Standard GPS Tracker	FF-GPS-100	TRACKER	99.00	10.00	USD	{}	2026-04-14 15:47:41.562924+05:30	2026-04-14 15:47:41.562924+05:30
415a03df-a97e-4fb5-a6d4-214d728865b6	FlyFleet OBD-II Telematics Node	FF-OBD-V1	TRACKER	149.00	15.00	USD	{}	2026-04-14 15:47:41.562924+05:30	2026-04-14 15:47:41.562924+05:30
\.


--
-- Data for Name: billing_invoice_documents; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_invoice_documents (id, invoice_id, template_id, file_url, storage_key, generated_at, template_version, snapshot, checksum) FROM stdin;
\.


--
-- Data for Name: billing_invoice_sequences; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_invoice_sequences (org_id, prefix, next_value, suffix, updated_at) FROM stdin;
\.


--
-- Data for Name: billing_invoice_templates; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_invoice_templates (id, name, version, html_content, is_default, created_at) FROM stdin;
\.


--
-- Data for Name: billing_plans; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_plans (id, name, description, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: billing_pricing_components; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_pricing_components (id, plan_id, name, type, unit_price, currency, meter_event_name, frequency, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: billing_recognized_revenue; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_recognized_revenue (id, schedule_id, amount, recognized_at, created_at) FROM stdin;
\.


--
-- Data for Name: billing_revenue_schedules; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_revenue_schedules (id, org_id, invoice_line_item_id, total_amount, currency, start_date, end_date, status, created_at) FROM stdin;
\.


--
-- Data for Name: billing_settings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_settings (org_id, default_tax_rate, base_currency, grace_period_days, auto_invoicing_enabled, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: billing_subscriptions; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_subscriptions (id, org_id, plan_id, status, current_period_start, current_period_end, trial_end, cancel_at_period_end, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: billing_tax_rates; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_tax_rates (id, region_id, name, percentage, tax_type, is_default, created_at) FROM stdin;
\.


--
-- Data for Name: billing_tax_regions; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.billing_tax_regions (id, name, country_code, state_code, is_active, created_at) FROM stdin;
\.


--
-- Data for Name: bulk_config; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.bulk_config (org_id, max_records_per_file, concurrent_threads, retention_months, mandatory_fields, updated_at) FROM stdin;
\.


--
-- Data for Name: bulk_job_records; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.bulk_job_records (id, job_id, row_number, record_data, status, error_message, created_at) FROM stdin;
\.


--
-- Data for Name: bulk_jobs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.bulk_jobs (id, org_id, file_name, status, total_records, success_count, failed_count, created_by, created_at, updated_at, retry_count) FROM stdin;
\.


--
-- Data for Name: consumption_events; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.consumption_events ("time", org_id, vehicle_id, meter_code, amount, metadata, processed_invoice_id) FROM stdin;
\.


--
-- Data for Name: documents; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.documents (id, vehicle_id, type, doc_number, issue_date, expiry_date, file_url, status, metadata, created_at, updated_at, org_id, title, file_name, file_size_bytes, mime_type, storage_path, reminder_days, uploaded_by, deleted_at) FROM stdin;
\.


--
-- Data for Name: dr_test_logs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.dr_test_logs (id, test_id, ts, level, message) FROM stdin;
\.


--
-- Data for Name: dr_test_settings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.dr_test_settings (id, schedule_enabled, frequency, schedule_day, schedule_time, intensity, notification_emails, updated_at) FROM stdin;
t	f	WEEKLY	SATURDAY	02:00	NORMAL	[]	2026-04-14 15:47:41.370608+05:30
\.


--
-- Data for Name: dr_test_steps; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.dr_test_steps (id, test_id, step_order, step_name, status, started_at, completed_at) FROM stdin;
\.


--
-- Data for Name: dr_tests; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.dr_tests (id, test_id, name, description, test_type, target_env, status, progress_percent, initiated_by, success_criteria, scheduled_at, started_at, completed_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: driver_attendance; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_attendance (id, driver_id, org_id, attendance_date, status, check_in_time, check_out_time, remarks, created_at) FROM stdin;
\.


--
-- Data for Name: driver_behavior; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_behavior (id, driver_id, org_id, period_start, period_end, score, safety_score, efficiency_score, compliance_score, harsh_accelerations, harsh_brakings, harsh_cornerings, speeding_incidents, phone_usage, seatbelt_violations, fuel_consumption, distance_driven, idle_time, eco_driving_score, hours_of_service, rest_periods, document_compliance, maintenance_adherence, night_driving_hours, highway_percentage, city_driving_percentage, risk_level, risk_factors, improvement_areas, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: driver_behavior_analytics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_behavior_analytics (id, driver_id, org_id, analysis_period_start, analysis_period_end, total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, behavior_trend, generated_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: driver_documents; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_documents (id, driver_id, org_id, type, name, document_number, issue_date, expiry_date, issuing_authority, file_url, file_name, file_size_bytes, mime_type, status, verified_by, verified_at, verification_notes, is_mandatory, reminder_days_before_expiry, last_reminder_sent_at, created_at, updated_at, created_by, updated_by) FROM stdin;
\.


--
-- Data for Name: driver_incidents; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_incidents (id, driver_id, org_id, type, severity, "timestamp", location, speed, details, resolved, resolution_notes, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: driver_schedules; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_schedules (id, driver_id, org_id, day_of_week, start_time, end_time, is_working, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: driver_verifications; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.driver_verifications (id, driver_id, org_id, license_number, status, verification_date, name, father_name, date_of_birth, gender, address, city, district, state, pin, license_class, issue_date, validity_from, validity_to, last_transaction, enrollment_number, created_at, created_by, updated_at, updated_by) FROM stdin;
\.


--
-- Data for Name: drivers; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.drivers (id, org_id, first_name, last_name, email, phone, license_number, license_type, license_expiry, license_authority, date_of_birth, address, emergency_contact, status, current_vehicle_id, current_trip_id, joining_date, resignation_date, experience_years, rating, total_trips, total_distance_km, total_hours_driven, behavior_score, documents, attendance_records, schedules, metadata, created_at, updated_at, deleted_at, created_by, updated_by) FROM stdin;
550e8400-e29b-41d4-a716-446655440010	550e8400-e29b-41d4-a716-446655440000	Rajesh	Kumar	rajesh@delta.io	+91 98765 43210	DL-12345	\N	\N	\N	\N	\N	\N	active	\N	\N	2026-04-15 20:04:27.487157+05:30	\N	0	5	0	0	0	100	\N	\N	\N	{}	2026-04-15 20:04:27.487157+05:30	2026-04-15 20:04:27.487157+05:30	\N	\N	\N
550e8400-e29b-41d4-a716-446655440011	550e8400-e29b-41d4-a716-446655440000	Amit	Sharma	amit@delta.io	+91 98765 43211	DL-23456	\N	\N	\N	\N	\N	\N	active	\N	\N	2026-04-15 20:04:27.487157+05:30	\N	0	5	0	0	0	100	\N	\N	\N	{}	2026-04-15 20:04:27.487157+05:30	2026-04-15 20:04:27.487157+05:30	\N	\N	\N
550e8400-e29b-41d4-a716-446655440012	550e8400-e29b-41d4-a716-446655440000	Suresh	Patel	suresh@delta.io	+91 98765 43212	DL-34567	\N	\N	\N	\N	\N	\N	active	\N	\N	2026-04-15 20:04:27.487157+05:30	\N	0	5	0	0	0	100	\N	\N	\N	{}	2026-04-15 20:04:27.487157+05:30	2026-04-15 20:04:27.487157+05:30	\N	\N	\N
\.


--
-- Data for Name: entity_audit_history; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.entity_audit_history (id, entity_type, entity_id, change_data, actor, "timestamp") FROM stdin;
\.


--
-- Data for Name: evidence_exports; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.evidence_exports (id, org_id, module, date_range_start, date_range_end, format, status, requested_by, file_url, error_message, created_at) FROM stdin;
\.


--
-- Data for Name: firmware_packages; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.firmware_packages (id, version, hardware_target, file_url, file_size, checksum, release_notes, status, is_critical, released_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: firmware_rollout_events; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.firmware_rollout_events (id, rollout_id, org_id, event_type, from_status, to_status, actor_user_id, message, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: firmware_rollouts; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.firmware_rollouts (id, firmware_id, org_id, name, device_group, target_filter, total_devices, completed_devices, failed_devices, status, rollout_strategy, max_concurrent, created_by, started_at, completed_at, created_at) FROM stdin;
\.


--
-- Data for Name: fleet_analytics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.fleet_analytics (id, org_id, start_date, end_date, total_vehicles, active_vehicles, total_distance, average_speed, total_fuel_consumption, average_fuel_efficiency, total_trips, completed_trips, cancelled_trips, average_trip_duration, idle_time_percentage, routing_efficiency, fuel_cost, maintenance_cost, revenue, profit_margin, safety_score, accidents_count, incidents_count, emergency_stops, geofence_violations, alerts_triggered, data_points_collected, processing_time, generated_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: geofences; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.geofences (id, org_id, name, type, center, radius_meters, coordinates, route_polyline, buffer_meters, entry_alert, exit_alert, speed_limit, vehicles, enabled, created_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: global_notification_defaults; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.global_notification_defaults (id, email_enabled, sms_enabled, push_enabled, in_app_enabled, updated_by, updated_at) FROM stdin;
418c29bc-0570-4d0d-9867-7174bb611f12	t	f	t	t	\N	2026-04-14 15:47:40.904464+05:30
\.


--
-- Data for Name: governance_settings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.governance_settings (org_id, retention_policies, masking_rules, auto_expiry_enabled, download_ttl_hours, created_at, created_by, updated_at, updated_by) FROM stdin;
\.


--
-- Data for Name: gps_locations; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.gps_locations ("time", vehicle_id, latitude, longitude, altitude, speed, heading, accuracy, ignition, odometer, battery_level, satellites, raw_data) FROM stdin;
2026-04-15 20:04:27.487157+05:30	550e8400-e29b-41d4-a716-446655440001	28.6139	77.209	\N	45	180	\N	t	\N	\N	\N	\N
2026-04-15 20:04:27.487157+05:30	550e8400-e29b-41d4-a716-446655440002	19.076	72.8777	\N	0	90	\N	t	\N	\N	\N	\N
2026-04-15 20:04:27.487157+05:30	550e8400-e29b-41d4-a716-446655440003	12.9716	77.5946	\N	12.5	45	\N	t	\N	\N	\N	\N
\.


--
-- Data for Name: gst_compliances; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.gst_compliances (id, org_id, gstin, status, legal_name, trade_name, address, centre_jurisdiction, state_jurisdiction, date_of_registration, date_of_cancellation, last_updated, return_filing_frequency, return_filing_status, payment_status, created_at, created_by, updated_at, updated_by) FROM stdin;
\.


--
-- Data for Name: incident_comments; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.incident_comments (id, ticket_id, author_id, author_name, comment_type, content, is_private, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: inspection_templates; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.inspection_templates (id, org_id, name, description, applicable_vehicle_types, checklist_items, is_active, created_at, updated_at) FROM stdin;
456516bb-d5b4-4203-8bf8-52dea0564335	00000000-0000-0000-0000-000000000001	Standard Safety Inspection	Routine safety check for all fleet vehicles	{TRUCK,VAN,CAR}	[{"text": "Brake pad thickness & fluid level", "type": "pass_fail", "item_id": "brakes", "required": true}, {"text": "Tire pressure & tread depth", "type": "pass_fail", "item_id": "tires", "required": true}, {"text": "Headlights, tail lights & signals", "type": "pass_fail", "item_id": "lights", "required": true}, {"text": "Oil, coolant & washer fluid", "type": "pass_fail", "item_id": "fluids", "required": false}, {"text": "Engine noise & diagnostic codes", "type": "pass_fail", "item_id": "engine", "required": true}]	t	2026-04-14 15:47:40.680947+05:30	2026-04-14 15:47:40.680947+05:30
\.


--
-- Data for Name: integration_configs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.integration_configs (id, org_id, service_type, api_endpoint, api_key, username, password, client_id, client_secret, environment, active, last_sync_at, next_sync_at, sync_interval, created_at, created_by, updated_at, updated_by) FROM stdin;
\.


--
-- Data for Name: integration_logs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.integration_logs (id, org_id, integration_id, method, endpoint, request_body, response_body, status_code, status, duration_ms, "timestamp") FROM stdin;
\.


--
-- Data for Name: integrations; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.integrations (id, org_id, name, type, provider, status, config, last_sync, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: invites; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.invites (id, code, org_id, role, expiry, created_at, created_by, used) FROM stdin;
\.


--
-- Data for Name: invoice_items; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.invoice_items (id, invoice_id, description, quantity, unit_price, tax_percent, total, created_at, updated_at, org_id, amount, is_deleted, deleted_at) FROM stdin;
\.


--
-- Data for Name: invoices; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.invoices (id, org_id, customer_name, invoice_date, due_date, status, currency, total_amount, amount_paid, balance_due, created_at, updated_at, invoice_number, customer_email, billing_address, subtotal, tax_amount, amount_remaining, notes, is_deleted, deleted_at, custom_fields) FROM stdin;
\.


--
-- Data for Name: ip_rules; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.ip_rules (id, org_id, cidr, action, description, added_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: local_permits; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.local_permits (id, vehicle_id, driver_id, org_id, permit_number, permit_type, permit_category, permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details, number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks, created_at, created_by, updated_at, updated_by) FROM stdin;
\.


--
-- Data for Name: login_events; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.login_events (id, org_id, user_id, email, ip_address, device, browser, location, status, risk_level, created_at) FROM stdin;
\.


--
-- Data for Name: maintenance_vendors; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.maintenance_vendors (id, org_id, name, specialization, rating, contact_info, payment_terms, status, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: mfa_codes; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_codes (id, user_id, code, method, expires_at, created_at) FROM stdin;
\.


--
-- Data for Name: mfa_events; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_events (id, user_id, org_id, event_type, status, method, ip_address, user_agent, location, details, created_at) FROM stdin;
\.


--
-- Data for Name: mfa_exceptions; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_exceptions (id, user_id, org_id, reason, granted_by, expires_at, created_at, start_date) FROM stdin;
\.


--
-- Data for Name: mfa_failures; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_failures (id, user_id, org_id, method, failure_reason, ip_address, location, created_at) FROM stdin;
\.


--
-- Data for Name: mfa_lockouts; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_lockouts (id, user_id, org_id, reason, locked_at, expires_at, unlocked_at, unlocked_by) FROM stdin;
\.


--
-- Data for Name: mfa_org_overrides; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_org_overrides (id, org_id, enabled, allowed_methods, max_attempts, lockout_minutes, remember_device_days, grace_period_days, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: mfa_policies; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_policies (id, org_id, require_mfa, allowed_methods, max_attempts, lockout_minutes, created_at, updated_at, grace_period_days, remember_device_days, enforce_at_login, admin_only_mode, remember_device_allowed, mfa_challenge_timeout, version) FROM stdin;
\.


--
-- Data for Name: mfa_policy_history; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_policy_history (id, policy_id, org_id, snapshot, changed_by, change_reason, version, created_at) FROM stdin;
\.


--
-- Data for Name: mfa_users; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.mfa_users (user_id, org_id, status, method, trusted_devices_count, last_reset, updated_at) FROM stdin;
\.


--
-- Data for Name: notification_events; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.notification_events (id, notification_id, event_type, message, created_at) FROM stdin;
\.


--
-- Data for Name: notification_logs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.notification_logs (id, correlation_id, user_id, user_email, channel, type, subject, body, status, provider, provider_response, retry_count, max_retries, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: notification_preference_audit; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.notification_preference_audit (id, user_id, changed_by, old_values, new_values, change_reason, created_at) FROM stdin;
\.


--
-- Data for Name: notification_settings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.notification_settings (id, channel, enabled, max_retries, priority, retention_months, updated_at) FROM stdin;
\.


--
-- Data for Name: obd_readings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.obd_readings ("time", vehicle_id, rpm, coolant_temp, engine_load, throttle_pos, fuel_level, dtcs, vin, raw_data) FROM stdin;
\.


--
-- Data for Name: org_support_metrics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.org_support_metrics (org_id, open_tickets, resolved_tickets, avg_resolution_time, sla_breaches, critical_issues, last_calculated, updated_at) FROM stdin;
\.


--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.organizations (id, name, domain, status, created_at, updated_at, default_theme, is_deleted, contact_email, plan, theme_color, theme_logo_url) FROM stdin;
550e8400-e29b-41d4-a716-446655440000	Default Fleet	\N	active	2026-04-15 00:25:52.439263+05:30	2026-04-15 00:25:52.439263+05:30	light	f	\N	enterprise	\N	\N
550e8400-e29b-41d4-a716-446655440002	Delta Logistics	\N	active	2026-04-15 09:38:44.666783+05:30	2026-04-15 09:38:44.666783+05:30	light	f	\N	enterprise	\N	\N
550e8400-e29b-41d4-a716-446655440003	Echo Transport	\N	active	2026-04-15 09:38:44.666783+05:30	2026-04-15 09:38:44.666783+05:30	light	f	\N	enterprise	\N	\N
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: u0_a412
--

COPY public.permissions (id, resource, action, description, bit_position) FROM stdin;
00823a5c-b066-4941-822f-0ace70f4e1b7	vehicles	create	Create new vehicles	0
97ffcef1-0637-409d-8842-3c59ec0a1ac7	vehicles	read	View vehicle details and list	1
55d5f03d-918b-45c0-8158-3c47f0292ac0	vehicles	update	Edit vehicle information	2
4348a7c2-a6e9-456b-8089-2806d71cc848	vehicles	delete	Delete vehicles	3
039e13a0-4681-4e3a-b484-46c036823bd0	vehicles	export	Export vehicle data	4
6c23c306-aa08-4ff0-96e7-0aa359cf277c	drivers	create	Add new drivers	5
4d97e613-78e7-476f-8707-44e69c5c6755	drivers	read	View driver profiles	6
244a187a-f19e-4bf3-a8ec-7d5ab21d9ef7	drivers	update	Edit driver information	7
012c158d-dd0f-414b-bbcd-795fc92554f3	drivers	delete	Remove drivers	8
52988e10-a804-403e-8771-1c5724333951	drivers	export	Export driver data	9
22dc45de-4d3b-4ecc-ad75-f63b8cf291be	trips	create	Create new trips	10
43a69522-53f5-411f-987e-71eb7e0d844f	trips	read	View trip details	11
6112ddf8-c08b-428e-a99b-1212c3a0512e	trips	update	Edit trip information	12
5072d322-ddd0-47d3-8e2d-15eade32d940	trips	delete	Delete trips	13
c4ef00e4-258b-4aaa-8fcb-4a875d1cd804	trips	export	Export trip data	14
ea100fb0-55f4-426f-8f10-46ec41729c07	users	create	Invite/create users	15
fb8751ee-de7c-4537-afa2-573ec67dcd78	users	read	View user profiles	16
2d6d8d18-c195-405c-ade8-21832a71c2ec	users	update	Edit user settings	17
4dbec463-907f-4740-b2c3-791b3f99969f	users	delete	Disable/delete users	18
d968b4a7-e8a9-4260-b0dc-ad47e23c10dc	roles	create	Create custom roles	19
b90710c8-a2a6-478a-95b6-0b466833f72f	roles	read	View role definitions	20
ddabcb9f-ac09-44be-bc5a-796fd8614d27	roles	update	Edit role permissions	21
0ffee3a9-07ef-4a9f-b5ea-25ac284cdb55	roles	delete	Delete custom roles	22
88713035-005c-45fb-a0b3-d79546a093ca	reports	read	View reports	23
55660c2b-7dea-4605-972f-6803e28bde9e	reports	export	Export reports	24
7b4ae0d2-d943-45c3-9d05-3f16f6803626	audit_logs	read	View audit logs	25
279a9bd3-8476-4563-84fd-bfff62ec7761	audit_logs	export	Export audit logs	26
dce2299e-3239-49e0-bce7-3260d7966c05	settings	read	View org settings	27
fb447a54-cad1-4bea-bb05-87cdc17eac3f	settings	update	Edit org settings	28
4469d324-f884-41d0-a282-8c48fd7745d6	alerts	read	View alerts	29
d9a9ccee-0a99-4bf5-b192-9e9cab43acdf	alerts	resolve	Resolve alerts	30
50ed340c-ebdc-4d82-b65c-9600cc1f4f1d	maintenance	create	Create maintenance tasks	31
feb9d808-5c38-4480-81f4-d3effc56c2ca	maintenance	read	View maintenance data	32
f78f0af4-35f6-4180-920e-38db9e570d90	maintenance	update	Update maintenance status	33
436c2ccd-b23e-4c62-84f7-934927bc1a1f	fuel	read	View fuel consumption	34
391c30a0-c217-4020-904f-c15efb1e3179	fuel	create	Log fuel entries	35
b9340ce9-e64e-41c9-baa6-9040ded2bbc8	billing	read	View billing info	36
6f6118e0-d5ef-4afd-8636-180902812a9e	billing	export	Export billing data	37
893ac6ce-9589-4a04-a138-6daa592dd449	analytics	read	View analytics dashboard	38
97a4b085-7447-43b8-a500-5e7fc69c49cc	analytics	export	Export analytics	39
90872cb5-fb79-4962-91aa-6a106ec634fa	integrations	read	View integrations	40
bf6e3ed2-f5f3-4d87-9db7-3ca256507436	integrations	manage	Configure integrations	41
0f618d77-8d99-4a31-880c-d8287536f959	governance	read	View governance dashboard	42
cc6d4a77-4ef4-44fd-ba90-3d50bfebe7cf	governance	manage	Manage governance policies	43
\.


--
-- Data for Name: prediction_results; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.prediction_results (id, model_id, prediction_input, predicted_value, confidence_score, probability_scores, feature_importance, anomaly_detected, anomaly_score, explanation, org_id, request_id, predicted_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: predictive_maintenance_analytics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.predictive_maintenance_analytics (id, vehicle_id, org_id, analysis_date, next_service_due, service_priority, component_risks, failure_prediction, maintenance_schedule, parts_needed, cost_estimation, uptime_prediction, reliability_score, recommendations, generated_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: predictive_models; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.predictive_models (id, name, version, model_type, description, algorithm, input_features, output_variable, training_data_start, training_data_end, accuracy, "precision", recall, f1_score, status, trained_at, expires_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: restore_jobs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.restore_jobs (id, backup_id, status, target_environment, started_at, completed_at, performed_by, logs) FROM stdin;
\.


--
-- Data for Name: review_items; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.review_items (id, review_id, user_id, role_id, current_status, reviewer_id, reviewed_at, notes, created_at) FROM stdin;
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: u0_a412
--

COPY public.role_permissions (id, role_id, permission_id) FROM stdin;
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.roles (id, name, description, is_system, created_at, updated_at, org_id, permission_mask, created_by, deleted_at) FROM stdin;
590f336e-277b-4832-920e-82d012578598	fleet_manager	Manages vehicles and drivers	t	2026-04-14 15:47:38.12678+05:30	2026-04-15 16:23:36.777614+05:30	550e8400-e29b-41d4-a716-446655440000	0	\N	\N
5a2e35d1-28f0-48a8-89e4-f38c346f7e4d	driver	Driver app access only	t	2026-04-14 15:47:38.12678+05:30	2026-04-15 16:23:36.777614+05:30	550e8400-e29b-41d4-a716-446655440000	0	\N	\N
819ac656-6648-41bf-9dec-c38906264adc	Fleet Manager	Manage vehicles, drivers, trips	t	2026-04-15 00:25:52.467914+05:30	2026-04-15 16:23:36.777614+05:30	550e8400-e29b-41d4-a716-446655440000	0	\N	\N
215779f2-b515-4bca-9c73-de2ac1e3f047	Dispatcher	Manage trips and routing	t	2026-04-15 00:25:52.467914+05:30	2026-04-15 16:23:36.777614+05:30	550e8400-e29b-41d4-a716-446655440000	0	\N	\N
857ad16f-1021-415b-b141-9bdc5b08b942	Driver	View assigned trips and vehicles	t	2026-04-15 00:25:52.467914+05:30	2026-04-15 16:23:36.777614+05:30	550e8400-e29b-41d4-a716-446655440000	0	\N	\N
d3a2b1c0-0000-0000-0000-000000000005	Viewer	Read-only access	t	2026-04-15 00:25:52.467914+05:30	2026-04-15 16:23:36.777614+05:30	550e8400-e29b-41d4-a716-446655440000	0	\N	\N
3629b4c1-d345-4e09-a5aa-06659b64259b	super_admin	Full platform access	t	2026-04-14 15:47:38.12678+05:30	2026-04-15 16:26:10.403844+05:30	550e8400-e29b-41d4-a716-446655440000	9223372036854775807	\N	\N
51116978-2cf0-4c79-af80-f803d6998d86	org_admin	Organization level administrator	t	2026-04-14 15:47:38.12678+05:30	2026-04-15 16:26:10.403844+05:30	550e8400-e29b-41d4-a716-446655440000	9223372036854775807	\N	\N
e349ce46-2c59-4c43-8338-afc37d01b88e	Super Admin	Full system access	t	2026-04-15 00:25:52.467914+05:30	2026-04-15 16:26:10.403844+05:30	550e8400-e29b-41d4-a716-446655440000	9223372036854775807	\N	\N
\.


--
-- Data for Name: route_optimization_jobs; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.route_optimization_jobs (id, request_id, status, worker_id, started_at, completed_at, error_message, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: route_optimization_requests; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.route_optimization_requests (id, org_id, status, constraints, stops, created_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: route_optimization_results; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.route_optimization_results (id, request_id, routes, total_distance, total_duration, total_cost, created_at) FROM stdin;
\.


--
-- Data for Name: security_events; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.security_events (id, user_id, event_type, ip_address, user_agent, metadata, created_at, actor_id, on_behalf_of) FROM stdin;
\.


--
-- Data for Name: sla_metrics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.sla_metrics ("time", org_id, metric_name, value, metadata) FROM stdin;
\.


--
-- Data for Name: storage_pools; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.storage_pools (id, name, status, total_bytes, used_bytes, available_bytes, location, backup_count, created_at, updated_at) FROM stdin;
a1b2c3d4-0001-0001-0001-000000000001	Primary Storage	ACTIVE	1073741824000	375809638400	697932185600	us-east-1	24	2026-01-14 15:47:40.281764+05:30	2026-04-14 15:47:40.281764+05:30
a1b2c3d4-0002-0002-0002-000000000002	DR Storage	ACTIVE	2147483648000	858993459200	1288490188800	us-west-2	48	2025-10-16 15:47:40.281764+05:30	2026-04-14 15:47:40.281764+05:30
a1b2c3d4-0003-0003-0003-000000000003	Archive Storage	WARNING	5368709120000	4509715660800	858993459200	eu-central-1	120	2025-04-14 15:47:40.281764+05:30	2026-04-13 15:47:40.281764+05:30
\.


--
-- Data for Name: support_audit_log; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.support_audit_log (id, ticket_id, user_id, action, old_value, new_value, details, created_at, ip_address) FROM stdin;
\.


--
-- Data for Name: support_tickets; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.support_tickets (id, ticket_id, org_id, title, description, status, priority, category, assigned_to, reporter_id, created_at, updated_at, resolved_at, sla_deadline, tags, metadata) FROM stdin;
\.


--
-- Data for Name: telemetry; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry (id, trip_id, latitude, longitude, speed, heading, event_type, event_time) FROM stdin;
\.


--
-- Data for Name: telemetry_analytics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry_analytics (id, vehicle_id, analysis_type, data, "timestamp", confidence) FROM stdin;
\.


--
-- Data for Name: telemetry_anomalies; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry_anomalies (id, vehicle_id, anomaly_type, data, "timestamp", confidence, anomaly_score, recommendation) FROM stdin;
\.


--
-- Data for Name: telemetry_events_p2026_04; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry_events_p2026_04 (id, vehicle_id, org_id, latitude, longitude, speed, heading, fuel_level, engine_rpm, created_at) FROM stdin;
\.


--
-- Data for Name: telemetry_events_p2026_05; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry_events_p2026_05 (id, vehicle_id, org_id, latitude, longitude, speed, heading, fuel_level, engine_rpm, created_at) FROM stdin;
\.


--
-- Data for Name: telemetry_events_p2026_06; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry_events_p2026_06 (id, vehicle_id, org_id, latitude, longitude, speed, heading, fuel_level, engine_rpm, created_at) FROM stdin;
\.


--
-- Data for Name: telemetry_events_p2026_07; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.telemetry_events_p2026_07 (id, vehicle_id, org_id, latitude, longitude, speed, heading, fuel_level, engine_rpm, created_at) FROM stdin;
\.


--
-- Data for Name: threat_alerts; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.threat_alerts (id, org_id, type, severity, ip, user_id, description, status, created_at) FROM stdin;
\.


--
-- Data for Name: trip_settings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.trip_settings (id, org_id, retention_period, policies, updated_at) FROM stdin;
\.


--
-- Data for Name: trips; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.trips (id, org_id, vehicle_id, driver_id, status, title, description, planned_route, actual_route, waypoints, start_location, end_location, estimated_distance_km, actual_distance_km, estimated_duration_min, actual_duration_min, fuel_consumed_liters, idling_time_minutes, cargo_details, pod_documents, scheduled_start, started_at, ended_at, customer_name, customer_phone, customer_email, cancellation_reason, events, is_deleted, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: usage_meters; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.usage_meters (id, code, name, description, unit, unit_price, currency, status, created_at) FROM stdin;
cdd762f6-1c14-4438-8dac-f34cb90db1e5	DISTANCE_KM	Mileage	\N	km	0.1500	USD	active	2026-04-14 15:47:41.114798+05:30
690a9909-63bb-40d4-bfca-9312d38b6866	HEARTBEAT_COUNT	Device Connectivity	\N	count	0.0100	USD	active	2026-04-14 15:47:41.114798+05:30
\.


--
-- Data for Name: user_links; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.user_links (parent_user_id, linked_user_id, created_at, created_by) FROM stdin;
\.


--
-- Data for Name: user_notification_preferences; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.user_notification_preferences (id, user_id, email_enabled, sms_enabled, push_enabled, in_app_enabled, last_modified_by, last_modified_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: user_organizations; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.user_organizations (id, user_id, org_id, role, status, created_at, role_id, is_active) FROM stdin;
00d0afd6-61b6-4620-b07d-5268dc3f50a2	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440000	admin	active	2026-04-15 00:25:52.47431+05:30	\N	t
9fddf4d4-536b-4f33-b290-e8c5eee1aa36	5bd5a78e-1db0-4fc4-a5bc-ffa6acdac381	550e8400-e29b-41d4-a716-446655440002	fleet_manager	active	2026-04-15 09:39:47.875438+05:30	\N	t
2fb5c2cc-a3a6-4438-beb6-d897770a9458	65f5713c-aee5-4b95-9d45-f77f82ec9ba1	550e8400-e29b-41d4-a716-446655440002	Dispatcher	active	2026-04-15 09:39:47.875438+05:30	\N	t
96460db8-0622-4fe2-ad43-88b06aa1e2da	4d577014-c7c1-4256-8762-cde342f27225	550e8400-e29b-41d4-a716-446655440002	fleet_manager	active	2026-04-15 09:39:47.875438+05:30	\N	t
1a582808-a88f-4fa4-99e5-e36388c05d79	8527a1c8-d448-41cc-aaea-0802f5d73f64	550e8400-e29b-41d4-a716-446655440002	fleet_manager	active	2026-04-15 09:39:47.875438+05:30	\N	t
c48f3154-6798-426a-98e6-c5ee42516997	7e7bde8c-c309-4014-8f0d-32c57930cc69	550e8400-e29b-41d4-a716-446655440002	Viewer	active	2026-04-15 09:39:47.875438+05:30	\N	t
c5ff5d55-2b36-413a-91ca-76b383abf7c4	da9ebc0e-ec26-4e91-a1c0-c1d572265733	550e8400-e29b-41d4-a716-446655440002	driver	active	2026-04-15 09:39:47.875438+05:30	\N	t
05f97c41-e428-47d8-8e7f-fbe7c720c010	f88abcea-35c0-42d9-a807-50778829b75d	550e8400-e29b-41d4-a716-446655440002	org_admin	active	2026-04-15 09:39:47.875438+05:30	51116978-2cf0-4c79-af80-f803d6998d86	t
2006bf22-f5fd-4149-a2ca-0c0073a27ca3	4fdce3b7-9cf0-478d-ab21-11341758ef89	550e8400-e29b-41d4-a716-446655440002	org_admin	active	2026-04-15 09:39:47.875438+05:30	51116978-2cf0-4c79-af80-f803d6998d86	t
70fa6637-caa6-445f-9eb6-243691d0a060	91fe43f5-8a2e-42da-a830-ecc689893190	550e8400-e29b-41d4-a716-446655440003	org_admin	active	2026-04-15 09:39:47.875438+05:30	51116978-2cf0-4c79-af80-f803d6998d86	t
f6382f3b-f3ba-4bab-8ad3-9317d3a2bc4a	5d6d3a46-29e4-4235-a40d-03e087ecd8b5	550e8400-e29b-41d4-a716-446655440002	org_admin	active	2026-04-15 15:49:04.269641+05:30	51116978-2cf0-4c79-af80-f803d6998d86	t
abc2ffe1-09c9-4c63-a891-68ad21a74021	5d6d3a46-29e4-4235-a40d-03e087ecd8b5	550e8400-e29b-41d4-a716-446655440000	org_admin	active	2026-04-15 15:55:50.979123+05:30	51116978-2cf0-4c79-af80-f803d6998d86	t
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: u0_a412
--

COPY public.user_roles (id, user_id, role_id, org_id, assigned_by, assigned_at) FROM stdin;
13796b2c-d61e-4a07-8983-cc0213a3c2c1	f88abcea-35c0-42d9-a807-50778829b75d	51116978-2cf0-4c79-af80-f803d6998d86	550e8400-e29b-41d4-a716-446655440002	\N	2026-04-15 17:26:39.882757+05:30
da36af3b-817e-40fa-b917-d7ffd40ad5fd	4fdce3b7-9cf0-478d-ab21-11341758ef89	51116978-2cf0-4c79-af80-f803d6998d86	550e8400-e29b-41d4-a716-446655440002	\N	2026-04-15 17:26:39.882757+05:30
3c40233e-7d42-4e70-b31e-6fc64cd67d49	91fe43f5-8a2e-42da-a830-ecc689893190	51116978-2cf0-4c79-af80-f803d6998d86	550e8400-e29b-41d4-a716-446655440003	\N	2026-04-15 17:26:39.882757+05:30
fe0b781d-697a-43ef-a06a-9573b6e345f0	5d6d3a46-29e4-4235-a40d-03e087ecd8b5	51116978-2cf0-4c79-af80-f803d6998d86	550e8400-e29b-41d4-a716-446655440002	\N	2026-04-15 17:26:39.882757+05:30
63cce0bd-f662-4864-baf3-44a7f28215c1	5d6d3a46-29e4-4235-a40d-03e087ecd8b5	51116978-2cf0-4c79-af80-f803d6998d86	550e8400-e29b-41d4-a716-446655440000	\N	2026-04-15 17:26:39.882757+05:30
\.


--
-- Data for Name: user_sessions; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.user_sessions (id, user_id, session_id, ip_address, user_agent, device_info, login_time, last_activity, logout_time, is_active, location) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.users (id, email, password_hash, name, phone, mfa_enabled, mfa_secret, status, created_at, updated_at, theme_updated_at, is_deleted, role_id, permission_mask, sidebar_collapsed, language, avatar_url, mfa_backup_codes, email_verified, failed_logins, locked_until, last_login_at, last_login_ip) FROM stdin;
63a20bb5-f3f4-48df-8a43-fc2d0a50c84c	testuser2@flyfleet.com	$2a$12$GfwlPOFH/Oz9VWpkIDDc4uswumZjJD8ldlmlyDPyM2DlxjfrlAyd.	Test User 2		f	\N	active	2026-04-15 09:33:39.796137+05:30	2026-04-15 09:33:39.796137+05:30	\N	f	3629b4c1-d345-4e09-a5aa-06659b64259b	0	f	en		[]	f	0	\N	\N	\N
5d6d3a46-29e4-4235-a40d-03e087ecd8b5	admin@delta.io	$2a$04$WOojoGDL.4DzIGNd3opDV.phzq0f8tFT8pEQXk2StJ7YvyGPDBahq	Admin User	\N	f	\N	active	2026-04-15 15:46:37.402001+05:30	2026-04-15 22:45:53.734708+05:30	\N	f	51116978-2cf0-4c79-af80-f803d6998d86	9223372036854775807	f	en	\N	[]	f	0	\N	2026-04-15 22:45:53.710365	127.0.0.1:38924
550e8400-e29b-41d4-a716-446655440001	admin@flyfleet.io	$2a$10$B2/XZniBnqEmecEhQGzUn.9ZU5sNpthotIAmxEJcWqGEkodD7Jp9i	Admin User		f		active	2026-04-15 00:25:52.469666+05:30	2026-04-15 09:35:28.028437+05:30	\N	f	e349ce46-2c59-4c43-8338-afc37d01b88e	9223372036854775807	f	en		[]	t	0	\N	2026-04-15 09:35:28.023467	127.0.0.1:41138
f88abcea-35c0-42d9-a807-50778829b75d	org_admin@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Org Admin	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	51116978-2cf0-4c79-af80-f803d6998d86	0	f	en	\N	[]	f	0	\N	\N	\N
5bd5a78e-1db0-4fc4-a5bc-ffa6acdac381	fleet_manager@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Fleet Manager	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	590f336e-277b-4832-920e-82d012578598	0	f	en	\N	[]	f	0	\N	\N	\N
65f5713c-aee5-4b95-9d45-f77f82ec9ba1	dispatcher@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Dispatcher	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	215779f2-b515-4bca-9c73-de2ac1e3f047	0	f	en	\N	[]	f	0	\N	\N	\N
4d577014-c7c1-4256-8762-cde342f27225	operator@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Fleet Operator	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	590f336e-277b-4832-920e-82d012578598	0	f	en	\N	[]	f	0	\N	\N	\N
8527a1c8-d448-41cc-aaea-0802f5d73f64	maintenance@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Maintenance Mgr	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	590f336e-277b-4832-920e-82d012578598	0	f	en	\N	[]	f	0	\N	\N	\N
7e7bde8c-c309-4014-8f0d-32c57930cc69	analyst@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Analyst	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	d3a2b1c0-0000-0000-0000-000000000005	0	f	en	\N	[]	f	0	\N	\N	\N
4fdce3b7-9cf0-478d-ab21-11341758ef89	billing@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Billing Admin	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	51116978-2cf0-4c79-af80-f803d6998d86	0	f	en	\N	[]	f	0	\N	\N	\N
da9ebc0e-ec26-4e91-a1c0-c1d572265733	driver@delta.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Driver	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	5a2e35d1-28f0-48a8-89e4-f38c346f7e4d	0	f	en	\N	[]	f	0	\N	\N	\N
91fe43f5-8a2e-42da-a830-ecc689893190	org_admin@echo.io	$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2	Echo Admin	\N	f	\N	active	2026-04-15 09:39:47.875438+05:30	2026-04-15 09:39:47.875438+05:30	\N	f	51116978-2cf0-4c79-af80-f803d6998d86	0	f	en	\N	[]	f	0	\N	\N	\N
\.


--
-- Data for Name: vehicle_inspections; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.vehicle_inspections (id, org_id, vehicle_id, template_id, inspector_id, status, results, notes, odometer_at_inspection, created_at, updated_at, created_by, updated_by) FROM stdin;
\.


--
-- Data for Name: vehicle_telemetry; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.vehicle_telemetry ("time", vehicle_id, latitude, longitude, speed_kph, heading, odometer_km, fuel_level_percent, engine_temp_celsius, alerts) FROM stdin;
2026-04-15 20:03:27.487157+05:30	550e8400-e29b-41d4-a716-446655440001	28.6139	77.209	45	180	\N	75.5	\N	[]
2026-04-15 20:03:27.487157+05:30	550e8400-e29b-41d4-a716-446655440002	19.076	72.8777	0	90	\N	85	\N	[]
2026-04-15 20:03:27.487157+05:30	550e8400-e29b-41d4-a716-446655440003	12.9716	77.5946	12.5	45	\N	60	\N	[]
\.


--
-- Data for Name: vehicle_verifications; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.vehicle_verifications (id, vehicle_id, org_id, registration_number, status, verification_date, chassis_number, engine_number, model, manufacturer, fuel_type, body_type, cubic_capacity, seating_capacity, color, manufacturing_year, owner_name, owner_address, owner_city, owner_district, owner_state, owner_pin, rc_status, insurance_company, insurance_policy, insurance_expiry, puc_number, puc_expiry, created_at, created_by, updated_at, updated_by) FROM stdin;
\.


--
-- Data for Name: vehicles; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.vehicles (id, org_id, license_plate, model, make, year, color, status, capacity, fuel_type, vin, registration, insurance, odometer, last_service, current_driver_id, created_at, updated_at, deleted_at, vehicle_no, service_due, notes, created_by, updated_by, tenant_id) FROM stdin;
550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440000	DL01AB1234	Ace Gold	Tata	2023	\N	active	1	\N	\N	\N	\N	0	\N	550e8400-e29b-41d4-a716-446655440010	2026-04-15 20:04:27.487157+05:30	2026-04-15 20:04:27.487157+05:30	\N	\N	\N	\N	\N	\N	\N
550e8400-e29b-41d4-a716-446655440002	550e8400-e29b-41d4-a716-446655440000	MH02CD5678	Partner	Ashok Leyland	2022	\N	active	1	\N	\N	\N	\N	0	\N	550e8400-e29b-41d4-a716-446655440011	2026-04-15 20:04:27.487157+05:30	2026-04-15 20:04:27.487157+05:30	\N	\N	\N	\N	\N	\N	\N
550e8400-e29b-41d4-a716-446655440003	550e8400-e29b-41d4-a716-446655440000	KA03EF9012	Pro 1000	Eicher	2023	\N	active	1	\N	\N	\N	\N	0	\N	550e8400-e29b-41d4-a716-446655440012	2026-04-15 20:04:27.487157+05:30	2026-04-15 20:04:27.487157+05:30	\N	\N	\N	\N	\N	\N	\N
\.


--
-- Data for Name: video_analytics; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.video_analytics (id, recording_id, stream_id, objects_detected, behavior_events, traffic_signs, incidents, "timestamp", confidence, processed_at, created_at) FROM stdin;
\.


--
-- Data for Name: video_recordings; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.video_recordings (id, stream_id, camera_id, vehicle_id, driver_id, org_id, filename, filepath, file_size, duration, resolution, framerate, bitrate, start_time, end_time, location_start, location_end, status, thumbnail_path, tags, event_type, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: video_retention_policies; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.video_retention_policies (id, org_id, name, description, event_type, retention_period, auto_archive, auto_delete, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: video_streams; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.video_streams (id, camera_id, vehicle_id, driver_id, org_id, url, status, resolution, framerate, bitrate, protocol, location, started_at, ended_at, created_at, updated_at, analytics_enabled) FROM stdin;
\.


--
-- Data for Name: video_thumbnails; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.video_thumbnails (id, recording_id, image_path, image_url, "timestamp", created_at) FROM stdin;
\.


--
-- Data for Name: webhooks; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.webhooks (id, org_id, url, events, secret, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: work_orders; Type: TABLE DATA; Schema: public; Owner: fleet_user
--

COPY public.work_orders (id, org_id, vehicle_id, title, description, maintenance_type, priority, status, assigned_to, scheduled_date, completed_date, estimated_cost, actual_cost, created_by, created_at, updated_at, vendor_id, approval_status, cost_breakdown, inspection_id, updated_by) FROM stdin;
\.


--
-- Name: access_reviews access_reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.access_reviews
    ADD CONSTRAINT access_reviews_pkey PRIMARY KEY (id);


--
-- Name: alert_rules alert_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_pkey PRIMARY KEY (id);


--
-- Name: alerts alerts_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_pkey PRIMARY KEY (id);


--
-- Name: analytics_jobs analytics_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.analytics_jobs
    ADD CONSTRAINT analytics_jobs_pkey PRIMARY KEY (id);


--
-- Name: api_clients api_clients_client_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.api_clients
    ADD CONSTRAINT api_clients_client_id_key UNIQUE (client_id);


--
-- Name: api_clients api_clients_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.api_clients
    ADD CONSTRAINT api_clients_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: audit_policies audit_policies_org_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_policies
    ADD CONSTRAINT audit_policies_org_id_key UNIQUE (org_id);


--
-- Name: audit_policies audit_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_policies
    ADD CONSTRAINT audit_policies_pkey PRIMARY KEY (id);


--
-- Name: audit_settings audit_settings_org_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_settings
    ADD CONSTRAINT audit_settings_org_id_key UNIQUE (org_id);


--
-- Name: audit_settings audit_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_settings
    ADD CONSTRAINT audit_settings_pkey PRIMARY KEY (id);


--
-- Name: auth_tokens auth_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);


--
-- Name: backup_verifications backup_verifications_backup_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.backup_verifications
    ADD CONSTRAINT backup_verifications_backup_id_key UNIQUE (backup_id);


--
-- Name: backup_verifications backup_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.backup_verifications
    ADD CONSTRAINT backup_verifications_pkey PRIMARY KEY (id);


--
-- Name: backups backups_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.backups
    ADD CONSTRAINT backups_pkey PRIMARY KEY (id);


--
-- Name: billing_accounting_exports billing_accounting_exports_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_accounting_exports
    ADD CONSTRAINT billing_accounting_exports_pkey PRIMARY KEY (id);


--
-- Name: billing_coupons billing_coupons_code_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_coupons
    ADD CONSTRAINT billing_coupons_code_key UNIQUE (code);


--
-- Name: billing_coupons billing_coupons_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_coupons
    ADD CONSTRAINT billing_coupons_pkey PRIMARY KEY (id);


--
-- Name: billing_credit_transactions billing_credit_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_credit_transactions
    ADD CONSTRAINT billing_credit_transactions_pkey PRIMARY KEY (id);


--
-- Name: billing_customer_balances billing_customer_balances_org_id_currency_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_customer_balances
    ADD CONSTRAINT billing_customer_balances_org_id_currency_key UNIQUE (org_id, currency);


--
-- Name: billing_customer_balances billing_customer_balances_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_customer_balances
    ADD CONSTRAINT billing_customer_balances_pkey PRIMARY KEY (id);


--
-- Name: billing_hardware_products billing_hardware_products_model_code_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_hardware_products
    ADD CONSTRAINT billing_hardware_products_model_code_key UNIQUE (model_code);


--
-- Name: billing_hardware_products billing_hardware_products_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_hardware_products
    ADD CONSTRAINT billing_hardware_products_pkey PRIMARY KEY (id);


--
-- Name: billing_invoice_documents billing_invoice_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_documents
    ADD CONSTRAINT billing_invoice_documents_pkey PRIMARY KEY (id);


--
-- Name: billing_invoice_sequences billing_invoice_sequences_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_sequences
    ADD CONSTRAINT billing_invoice_sequences_pkey PRIMARY KEY (org_id);


--
-- Name: billing_invoice_templates billing_invoice_templates_name_version_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_templates
    ADD CONSTRAINT billing_invoice_templates_name_version_key UNIQUE (name, version);


--
-- Name: billing_invoice_templates billing_invoice_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_templates
    ADD CONSTRAINT billing_invoice_templates_pkey PRIMARY KEY (id);


--
-- Name: billing_plans billing_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_plans
    ADD CONSTRAINT billing_plans_pkey PRIMARY KEY (id);


--
-- Name: billing_pricing_components billing_pricing_components_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_pricing_components
    ADD CONSTRAINT billing_pricing_components_pkey PRIMARY KEY (id);


--
-- Name: billing_recognized_revenue billing_recognized_revenue_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_recognized_revenue
    ADD CONSTRAINT billing_recognized_revenue_pkey PRIMARY KEY (id);


--
-- Name: billing_revenue_schedules billing_revenue_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_revenue_schedules
    ADD CONSTRAINT billing_revenue_schedules_pkey PRIMARY KEY (id);


--
-- Name: billing_settings billing_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_settings
    ADD CONSTRAINT billing_settings_pkey PRIMARY KEY (org_id);


--
-- Name: billing_subscriptions billing_subscriptions_org_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_subscriptions
    ADD CONSTRAINT billing_subscriptions_org_id_key UNIQUE (org_id);


--
-- Name: billing_subscriptions billing_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_subscriptions
    ADD CONSTRAINT billing_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: billing_tax_rates billing_tax_rates_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_tax_rates
    ADD CONSTRAINT billing_tax_rates_pkey PRIMARY KEY (id);


--
-- Name: billing_tax_regions billing_tax_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_tax_regions
    ADD CONSTRAINT billing_tax_regions_pkey PRIMARY KEY (id);


--
-- Name: bulk_config bulk_config_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.bulk_config
    ADD CONSTRAINT bulk_config_pkey PRIMARY KEY (org_id);


--
-- Name: bulk_job_records bulk_job_records_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.bulk_job_records
    ADD CONSTRAINT bulk_job_records_pkey PRIMARY KEY (id);


--
-- Name: bulk_jobs bulk_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.bulk_jobs
    ADD CONSTRAINT bulk_jobs_pkey PRIMARY KEY (id);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: dr_test_logs dr_test_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_test_logs
    ADD CONSTRAINT dr_test_logs_pkey PRIMARY KEY (id);


--
-- Name: dr_test_settings dr_test_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_test_settings
    ADD CONSTRAINT dr_test_settings_pkey PRIMARY KEY (id);


--
-- Name: dr_test_steps dr_test_steps_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_test_steps
    ADD CONSTRAINT dr_test_steps_pkey PRIMARY KEY (id);


--
-- Name: dr_tests dr_tests_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_tests
    ADD CONSTRAINT dr_tests_pkey PRIMARY KEY (id);


--
-- Name: dr_tests dr_tests_test_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_tests
    ADD CONSTRAINT dr_tests_test_id_key UNIQUE (test_id);


--
-- Name: driver_attendance driver_attendance_driver_id_attendance_date_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_attendance
    ADD CONSTRAINT driver_attendance_driver_id_attendance_date_key UNIQUE (driver_id, attendance_date);


--
-- Name: driver_attendance driver_attendance_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_attendance
    ADD CONSTRAINT driver_attendance_pkey PRIMARY KEY (id);


--
-- Name: driver_behavior_analytics driver_behavior_analytics_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_behavior_analytics
    ADD CONSTRAINT driver_behavior_analytics_pkey PRIMARY KEY (id);


--
-- Name: driver_behavior driver_behavior_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_behavior
    ADD CONSTRAINT driver_behavior_pkey PRIMARY KEY (id);


--
-- Name: driver_documents driver_documents_driver_id_type_document_number_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_driver_id_type_document_number_key UNIQUE (driver_id, type, document_number);


--
-- Name: driver_documents driver_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_pkey PRIMARY KEY (id);


--
-- Name: driver_incidents driver_incidents_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_incidents
    ADD CONSTRAINT driver_incidents_pkey PRIMARY KEY (id);


--
-- Name: driver_schedules driver_schedules_driver_id_day_of_week_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_schedules
    ADD CONSTRAINT driver_schedules_driver_id_day_of_week_key UNIQUE (driver_id, day_of_week);


--
-- Name: driver_schedules driver_schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_schedules
    ADD CONSTRAINT driver_schedules_pkey PRIMARY KEY (id);


--
-- Name: driver_verifications driver_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_verifications
    ADD CONSTRAINT driver_verifications_pkey PRIMARY KEY (id);


--
-- Name: drivers drivers_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_pkey PRIMARY KEY (id);


--
-- Name: entity_audit_history entity_audit_history_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.entity_audit_history
    ADD CONSTRAINT entity_audit_history_pkey PRIMARY KEY (id);


--
-- Name: evidence_exports evidence_exports_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.evidence_exports
    ADD CONSTRAINT evidence_exports_pkey PRIMARY KEY (id);


--
-- Name: firmware_packages firmware_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_packages
    ADD CONSTRAINT firmware_packages_pkey PRIMARY KEY (id);


--
-- Name: firmware_rollout_events firmware_rollout_events_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollout_events
    ADD CONSTRAINT firmware_rollout_events_pkey PRIMARY KEY (id);


--
-- Name: firmware_rollouts firmware_rollouts_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollouts
    ADD CONSTRAINT firmware_rollouts_pkey PRIMARY KEY (id);


--
-- Name: fleet_analytics fleet_analytics_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.fleet_analytics
    ADD CONSTRAINT fleet_analytics_pkey PRIMARY KEY (id);


--
-- Name: geofences geofences_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.geofences
    ADD CONSTRAINT geofences_pkey PRIMARY KEY (id);


--
-- Name: global_notification_defaults global_notification_defaults_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.global_notification_defaults
    ADD CONSTRAINT global_notification_defaults_pkey PRIMARY KEY (id);


--
-- Name: governance_settings governance_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.governance_settings
    ADD CONSTRAINT governance_settings_pkey PRIMARY KEY (org_id);


--
-- Name: gps_locations gps_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.gps_locations
    ADD CONSTRAINT gps_locations_pkey PRIMARY KEY ("time", vehicle_id);


--
-- Name: gst_compliances gst_compliances_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.gst_compliances
    ADD CONSTRAINT gst_compliances_pkey PRIMARY KEY (id);


--
-- Name: incident_comments incident_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.incident_comments
    ADD CONSTRAINT incident_comments_pkey PRIMARY KEY (id);


--
-- Name: inspection_templates inspection_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.inspection_templates
    ADD CONSTRAINT inspection_templates_pkey PRIMARY KEY (id);


--
-- Name: integration_configs integration_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integration_configs
    ADD CONSTRAINT integration_configs_pkey PRIMARY KEY (id);


--
-- Name: integration_logs integration_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integration_logs
    ADD CONSTRAINT integration_logs_pkey PRIMARY KEY (id);


--
-- Name: integrations integrations_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT integrations_pkey PRIMARY KEY (id);


--
-- Name: invites invites_code_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invites
    ADD CONSTRAINT invites_code_key UNIQUE (code);


--
-- Name: invites invites_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invites
    ADD CONSTRAINT invites_pkey PRIMARY KEY (id);


--
-- Name: invoice_items invoice_items_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invoice_items
    ADD CONSTRAINT invoice_items_pkey PRIMARY KEY (id);


--
-- Name: invoices invoices_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);


--
-- Name: ip_rules ip_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.ip_rules
    ADD CONSTRAINT ip_rules_pkey PRIMARY KEY (id);


--
-- Name: local_permits local_permits_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.local_permits
    ADD CONSTRAINT local_permits_pkey PRIMARY KEY (id);


--
-- Name: login_events login_events_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.login_events
    ADD CONSTRAINT login_events_pkey PRIMARY KEY (id);


--
-- Name: maintenance_vendors maintenance_vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.maintenance_vendors
    ADD CONSTRAINT maintenance_vendors_pkey PRIMARY KEY (id);


--
-- Name: mfa_codes mfa_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_codes
    ADD CONSTRAINT mfa_codes_pkey PRIMARY KEY (id);


--
-- Name: mfa_events mfa_events_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_events
    ADD CONSTRAINT mfa_events_pkey PRIMARY KEY (id);


--
-- Name: mfa_exceptions mfa_exceptions_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_exceptions
    ADD CONSTRAINT mfa_exceptions_pkey PRIMARY KEY (id);


--
-- Name: mfa_failures mfa_failures_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_failures
    ADD CONSTRAINT mfa_failures_pkey PRIMARY KEY (id);


--
-- Name: mfa_lockouts mfa_lockouts_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_lockouts
    ADD CONSTRAINT mfa_lockouts_pkey PRIMARY KEY (id);


--
-- Name: mfa_org_overrides mfa_org_overrides_org_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_org_overrides
    ADD CONSTRAINT mfa_org_overrides_org_id_key UNIQUE (org_id);


--
-- Name: mfa_org_overrides mfa_org_overrides_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_org_overrides
    ADD CONSTRAINT mfa_org_overrides_pkey PRIMARY KEY (id);


--
-- Name: mfa_policies mfa_policies_org_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policies
    ADD CONSTRAINT mfa_policies_org_id_key UNIQUE (org_id);


--
-- Name: mfa_policies mfa_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policies
    ADD CONSTRAINT mfa_policies_pkey PRIMARY KEY (id);


--
-- Name: mfa_policy_history mfa_policy_history_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policy_history
    ADD CONSTRAINT mfa_policy_history_pkey PRIMARY KEY (id);


--
-- Name: mfa_users mfa_users_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_users
    ADD CONSTRAINT mfa_users_pkey PRIMARY KEY (user_id);


--
-- Name: notification_events notification_events_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_events
    ADD CONSTRAINT notification_events_pkey PRIMARY KEY (id);


--
-- Name: notification_logs notification_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_logs
    ADD CONSTRAINT notification_logs_pkey PRIMARY KEY (id);


--
-- Name: notification_preference_audit notification_preference_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_preference_audit
    ADD CONSTRAINT notification_preference_audit_pkey PRIMARY KEY (id);


--
-- Name: notification_settings notification_settings_channel_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_channel_key UNIQUE (channel);


--
-- Name: notification_settings notification_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_pkey PRIMARY KEY (id);


--
-- Name: obd_readings obd_readings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.obd_readings
    ADD CONSTRAINT obd_readings_pkey PRIMARY KEY ("time", vehicle_id);


--
-- Name: org_support_metrics org_support_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.org_support_metrics
    ADD CONSTRAINT org_support_metrics_pkey PRIMARY KEY (org_id);


--
-- Name: organizations organizations_domain_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_domain_key UNIQUE (domain);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_bit_position_key; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_bit_position_key UNIQUE (bit_position);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_resource_action_key; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_resource_action_key UNIQUE (resource, action);


--
-- Name: prediction_results prediction_results_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.prediction_results
    ADD CONSTRAINT prediction_results_pkey PRIMARY KEY (id);


--
-- Name: predictive_maintenance_analytics predictive_maintenance_analytics_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.predictive_maintenance_analytics
    ADD CONSTRAINT predictive_maintenance_analytics_pkey PRIMARY KEY (id);


--
-- Name: predictive_models predictive_models_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.predictive_models
    ADD CONSTRAINT predictive_models_pkey PRIMARY KEY (id);


--
-- Name: restore_jobs restore_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.restore_jobs
    ADD CONSTRAINT restore_jobs_pkey PRIMARY KEY (id);


--
-- Name: review_items review_items_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.review_items
    ADD CONSTRAINT review_items_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_role_id_permission_id_key; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_permission_id_key UNIQUE (role_id, permission_id);


--
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: route_optimization_jobs route_optimization_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_jobs
    ADD CONSTRAINT route_optimization_jobs_pkey PRIMARY KEY (id);


--
-- Name: route_optimization_requests route_optimization_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_requests
    ADD CONSTRAINT route_optimization_requests_pkey PRIMARY KEY (id);


--
-- Name: route_optimization_results route_optimization_results_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_results
    ADD CONSTRAINT route_optimization_results_pkey PRIMARY KEY (id);


--
-- Name: security_events security_events_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.security_events
    ADD CONSTRAINT security_events_pkey PRIMARY KEY (id);


--
-- Name: storage_pools storage_pools_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.storage_pools
    ADD CONSTRAINT storage_pools_pkey PRIMARY KEY (id);


--
-- Name: support_audit_log support_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_audit_log
    ADD CONSTRAINT support_audit_log_pkey PRIMARY KEY (id);


--
-- Name: support_tickets support_tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_pkey PRIMARY KEY (id);


--
-- Name: support_tickets support_tickets_ticket_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_ticket_id_key UNIQUE (ticket_id);


--
-- Name: telemetry_analytics telemetry_analytics_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_analytics
    ADD CONSTRAINT telemetry_analytics_pkey PRIMARY KEY (id);


--
-- Name: telemetry_anomalies telemetry_anomalies_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_anomalies
    ADD CONSTRAINT telemetry_anomalies_pkey PRIMARY KEY (id);


--
-- Name: telemetry telemetry_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry
    ADD CONSTRAINT telemetry_pkey PRIMARY KEY (id);


--
-- Name: threat_alerts threat_alerts_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.threat_alerts
    ADD CONSTRAINT threat_alerts_pkey PRIMARY KEY (id);


--
-- Name: trip_settings trip_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trip_settings
    ADD CONSTRAINT trip_settings_pkey PRIMARY KEY (id);


--
-- Name: trips trips_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_pkey PRIMARY KEY (id);


--
-- Name: maintenance_vendors unique_vendor_per_org; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.maintenance_vendors
    ADD CONSTRAINT unique_vendor_per_org UNIQUE (org_id, name);


--
-- Name: trip_settings uq_trip_settings_org_id; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trip_settings
    ADD CONSTRAINT uq_trip_settings_org_id UNIQUE (org_id);


--
-- Name: usage_meters usage_meters_code_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.usage_meters
    ADD CONSTRAINT usage_meters_code_key UNIQUE (code);


--
-- Name: usage_meters usage_meters_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.usage_meters
    ADD CONSTRAINT usage_meters_pkey PRIMARY KEY (id);


--
-- Name: user_links user_links_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_links
    ADD CONSTRAINT user_links_pkey PRIMARY KEY (parent_user_id, linked_user_id);


--
-- Name: user_notification_preferences user_notification_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_pkey PRIMARY KEY (id);


--
-- Name: user_notification_preferences user_notification_preferences_user_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_key UNIQUE (user_id);


--
-- Name: user_organizations user_organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_pkey PRIMARY KEY (id);


--
-- Name: user_organizations user_organizations_user_id_org_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_user_id_org_id_key UNIQUE (user_id, org_id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: user_roles user_roles_user_id_role_id_org_id_key; Type: CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_role_id_org_id_key UNIQUE (user_id, role_id, org_id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_session_id_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_session_id_key UNIQUE (session_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: vehicle_inspections vehicle_inspections_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_inspections
    ADD CONSTRAINT vehicle_inspections_pkey PRIMARY KEY (id);


--
-- Name: vehicle_telemetry vehicle_telemetry_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_telemetry
    ADD CONSTRAINT vehicle_telemetry_pkey PRIMARY KEY ("time", vehicle_id);


--
-- Name: vehicle_verifications vehicle_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_verifications
    ADD CONSTRAINT vehicle_verifications_pkey PRIMARY KEY (id);


--
-- Name: vehicles vehicles_license_plate_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT vehicles_license_plate_key UNIQUE (license_plate);


--
-- Name: vehicles vehicles_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT vehicles_pkey PRIMARY KEY (id);


--
-- Name: vehicles vehicles_vehicle_no_key; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT vehicles_vehicle_no_key UNIQUE (vehicle_no);


--
-- Name: video_analytics video_analytics_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_analytics
    ADD CONSTRAINT video_analytics_pkey PRIMARY KEY (id);


--
-- Name: video_recordings video_recordings_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_recordings
    ADD CONSTRAINT video_recordings_pkey PRIMARY KEY (id);


--
-- Name: video_retention_policies video_retention_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_retention_policies
    ADD CONSTRAINT video_retention_policies_pkey PRIMARY KEY (id);


--
-- Name: video_streams video_streams_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_streams
    ADD CONSTRAINT video_streams_pkey PRIMARY KEY (id);


--
-- Name: video_thumbnails video_thumbnails_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_thumbnails
    ADD CONSTRAINT video_thumbnails_pkey PRIMARY KEY (id);


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);


--
-- Name: work_orders work_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.work_orders
    ADD CONSTRAINT work_orders_pkey PRIMARY KEY (id);


--
-- Name: idx_access_reviews_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_access_reviews_created_by ON public.access_reviews USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_access_reviews_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_access_reviews_org ON public.access_reviews USING btree (org_id);


--
-- Name: idx_access_reviews_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_access_reviews_org_id ON public.access_reviews USING btree (org_id);


--
-- Name: idx_access_reviews_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_access_reviews_status ON public.access_reviews USING btree (org_id, status);


--
-- Name: idx_accounting_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_accounting_org_id ON public.billing_accounting_exports USING btree (org_id);


--
-- Name: idx_accounting_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_accounting_status ON public.billing_accounting_exports USING btree (status);


--
-- Name: idx_alert_rules_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alert_rules_created_by ON public.alert_rules USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_alert_rules_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alert_rules_org_id ON public.alert_rules USING btree (org_id);


--
-- Name: idx_alerts_acknowledged_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_acknowledged_by ON public.alerts USING btree (acknowledged_by) WHERE (acknowledged_by IS NOT NULL);


--
-- Name: idx_alerts_dedup_active; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_dedup_active ON public.alerts USING btree (dedup_key) WHERE ((status)::text = 'new'::text);


--
-- Name: idx_alerts_driver_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_driver_id ON public.alerts USING btree (driver_id) WHERE (driver_id IS NOT NULL);


--
-- Name: idx_alerts_escalation_worker; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_escalation_worker ON public.alerts USING btree (status, sla_deadline) WHERE ((status)::text = 'new'::text);


--
-- Name: idx_alerts_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_org_id ON public.alerts USING btree (org_id);


--
-- Name: idx_alerts_resolved_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_resolved_by ON public.alerts USING btree (resolved_by) WHERE (resolved_by IS NOT NULL);


--
-- Name: idx_alerts_rule_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_rule_id ON public.alerts USING btree (rule_id) WHERE (rule_id IS NOT NULL);


--
-- Name: idx_alerts_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_status ON public.alerts USING btree (status);


--
-- Name: idx_alerts_timestamp; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_timestamp ON public.alerts USING btree ("timestamp");


--
-- Name: idx_alerts_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_alerts_vehicle_id ON public.alerts USING btree (vehicle_id);


--
-- Name: idx_analytics_jobs_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_analytics_jobs_org_id ON public.analytics_jobs USING btree (org_id);


--
-- Name: idx_analytics_jobs_org_status_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_analytics_jobs_org_status_type ON public.analytics_jobs USING btree (org_id, status, job_type);


--
-- Name: idx_api_clients_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_api_clients_org ON public.api_clients USING btree (org_id);


--
-- Name: idx_api_clients_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_api_clients_org_id ON public.api_clients USING btree (org_id);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_logs_org_id ON public.audit_logs USING btree (org_id);


--
-- Name: idx_audit_logs_severity; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_logs_severity ON public.audit_logs USING btree (severity);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_audit_policies_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_policies_org_id ON public.audit_policies USING btree (org_id);


--
-- Name: idx_audit_policies_updated_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_policies_updated_by ON public.audit_policies USING btree (updated_by) WHERE (updated_by IS NOT NULL);


--
-- Name: idx_audit_settings_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_audit_settings_org_id ON public.audit_settings USING btree (org_id);


--
-- Name: idx_auth_tokens_refresh_token; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_auth_tokens_refresh_token ON public.auth_tokens USING btree (refresh_token);


--
-- Name: idx_auth_tokens_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_auth_tokens_user_id ON public.auth_tokens USING btree (user_id);


--
-- Name: idx_backup_verifications_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_backup_verifications_status ON public.backup_verifications USING btree (status);


--
-- Name: idx_backup_verifications_verified_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_backup_verifications_verified_by ON public.backup_verifications USING btree (verified_by) WHERE (verified_by IS NOT NULL);


--
-- Name: idx_backups_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_backups_created_at ON public.backups USING btree (created_at);


--
-- Name: idx_backups_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_backups_status ON public.backups USING btree (status);


--
-- Name: idx_behavior_driver_period; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_behavior_driver_period ON public.driver_behavior USING btree (driver_id, period_start);


--
-- Name: idx_billing_accounting_exports_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_accounting_exports_org_id ON public.billing_accounting_exports USING btree (org_id);


--
-- Name: idx_billing_customer_balances_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_customer_balances_org_id ON public.billing_customer_balances USING btree (org_id);


--
-- Name: idx_billing_invoice_docs_snapshot; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_invoice_docs_snapshot ON public.billing_invoice_documents USING gin (snapshot);


--
-- Name: idx_billing_invoice_documents_template_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_invoice_documents_template_id ON public.billing_invoice_documents USING btree (template_id) WHERE (template_id IS NOT NULL);


--
-- Name: idx_billing_invoice_sequences_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_invoice_sequences_org_id ON public.billing_invoice_sequences USING btree (org_id);


--
-- Name: idx_billing_recognized_revenue_schedule_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_recognized_revenue_schedule_id ON public.billing_recognized_revenue USING btree (schedule_id) WHERE (schedule_id IS NOT NULL);


--
-- Name: idx_billing_revenue_schedules_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_revenue_schedules_org_id ON public.billing_revenue_schedules USING btree (org_id);


--
-- Name: idx_billing_settings_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_settings_org_id ON public.billing_settings USING btree (org_id);


--
-- Name: idx_billing_subscriptions_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_billing_subscriptions_org_id ON public.billing_subscriptions USING btree (org_id);


--
-- Name: idx_bulk_config_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_bulk_config_org_id ON public.bulk_config USING btree (org_id);


--
-- Name: idx_bulk_job_records_job_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_bulk_job_records_job_id ON public.bulk_job_records USING btree (job_id);


--
-- Name: idx_bulk_jobs_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_bulk_jobs_org ON public.bulk_jobs USING btree (org_id);


--
-- Name: idx_bulk_jobs_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_bulk_jobs_org_id ON public.bulk_jobs USING btree (org_id);


--
-- Name: idx_bulk_jobs_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_bulk_jobs_status ON public.bulk_jobs USING btree (status);


--
-- Name: idx_components_plan_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_components_plan_id ON public.billing_pricing_components USING btree (plan_id);


--
-- Name: idx_consumption_aggregation; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_consumption_aggregation ON public.consumption_events USING btree (org_id, meter_code, "time" DESC) WHERE (processed_invoice_id IS NULL);


--
-- Name: idx_consumption_events_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_consumption_events_org_id ON public.consumption_events USING btree (org_id);


--
-- Name: idx_consumption_events_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_consumption_events_vehicle_id ON public.consumption_events USING btree (vehicle_id) WHERE (vehicle_id IS NOT NULL);


--
-- Name: idx_credit_txn_balance_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_credit_txn_balance_created ON public.billing_credit_transactions USING btree (balance_id, created_at DESC);


--
-- Name: idx_credit_txn_balance_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_credit_txn_balance_id ON public.billing_credit_transactions USING btree (balance_id);


--
-- Name: idx_doc_invoice_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_doc_invoice_id ON public.billing_invoice_documents USING btree (invoice_id);


--
-- Name: idx_documents_expiry_date; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_expiry_date ON public.documents USING btree (expiry_date);


--
-- Name: idx_documents_org_expiry; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_org_expiry ON public.documents USING btree (org_id, expiry_date) WHERE ((deleted_at IS NULL) AND ((status)::text <> 'expired'::text));


--
-- Name: idx_documents_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_org_id ON public.documents USING btree (org_id);


--
-- Name: idx_documents_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_status ON public.documents USING btree (status);


--
-- Name: idx_documents_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_type ON public.documents USING btree (type);


--
-- Name: idx_documents_uploaded_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_uploaded_by ON public.documents USING btree (uploaded_by) WHERE (uploaded_by IS NOT NULL);


--
-- Name: idx_documents_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_vehicle_id ON public.documents USING btree (vehicle_id);


--
-- Name: idx_documents_vehicle_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_documents_vehicle_status ON public.documents USING btree (vehicle_id, status) WHERE (deleted_at IS NULL);


--
-- Name: idx_dr_test_logs_test_id_ts; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_dr_test_logs_test_id_ts ON public.dr_test_logs USING btree (test_id, ts DESC);


--
-- Name: idx_dr_test_steps_test_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_dr_test_steps_test_id ON public.dr_test_steps USING btree (test_id, step_order);


--
-- Name: idx_dr_tests_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_dr_tests_created_at ON public.dr_tests USING btree (created_at DESC);


--
-- Name: idx_dr_tests_scheduled_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_dr_tests_scheduled_at ON public.dr_tests USING btree (scheduled_at);


--
-- Name: idx_dr_tests_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_dr_tests_status ON public.dr_tests USING btree (status);


--
-- Name: idx_driver_attendance_date_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_attendance_date_status ON public.driver_attendance USING btree (attendance_date, status);


--
-- Name: idx_driver_attendance_driver; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_attendance_driver ON public.driver_attendance USING btree (driver_id);


--
-- Name: idx_driver_attendance_org_date; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_attendance_org_date ON public.driver_attendance USING btree (org_id, attendance_date DESC);


--
-- Name: idx_driver_attendance_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_attendance_org_id ON public.driver_attendance USING btree (org_id);


--
-- Name: idx_driver_behavior_analytics_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_behavior_analytics_org_id ON public.driver_behavior_analytics USING btree (org_id);


--
-- Name: idx_driver_behavior_driver; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_behavior_driver ON public.driver_behavior_analytics USING btree (driver_id, analysis_period_start, analysis_period_end);


--
-- Name: idx_driver_behavior_driver_recent; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_behavior_driver_recent ON public.driver_behavior USING btree (driver_id, period_start DESC);


--
-- Name: INDEX idx_driver_behavior_driver_recent; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON INDEX public.idx_driver_behavior_driver_recent IS 'Quick retrieval of latest driver behavior scores';


--
-- Name: idx_driver_behavior_org_dates; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_behavior_org_dates ON public.driver_behavior_analytics USING btree (org_id, analysis_period_start, analysis_period_end);


--
-- Name: idx_driver_behavior_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_behavior_org_id ON public.driver_behavior USING btree (org_id);


--
-- Name: idx_driver_behavior_org_period; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_behavior_org_period ON public.driver_behavior USING btree (org_id, period_start DESC);


--
-- Name: idx_driver_documents_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_created_by ON public.driver_documents USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_driver_documents_driver; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_driver ON public.driver_documents USING btree (driver_id);


--
-- Name: idx_driver_documents_driver_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_driver_type ON public.driver_documents USING btree (driver_id, type);


--
-- Name: idx_driver_documents_expiry; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_expiry ON public.driver_documents USING btree (expiry_date) WHERE (expiry_date IS NOT NULL);


--
-- Name: idx_driver_documents_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_org ON public.driver_documents USING btree (org_id);


--
-- Name: idx_driver_documents_org_expiry; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_org_expiry ON public.driver_documents USING btree (org_id, expiry_date) WHERE (expiry_date IS NOT NULL);


--
-- Name: idx_driver_documents_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_org_id ON public.driver_documents USING btree (org_id);


--
-- Name: idx_driver_documents_org_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_org_status ON public.driver_documents USING btree (org_id, status);


--
-- Name: idx_driver_documents_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_status ON public.driver_documents USING btree (status);


--
-- Name: idx_driver_documents_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_type ON public.driver_documents USING btree (type);


--
-- Name: idx_driver_documents_updated_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_updated_by ON public.driver_documents USING btree (updated_by) WHERE (updated_by IS NOT NULL);


--
-- Name: idx_driver_documents_verified_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_documents_verified_by ON public.driver_documents USING btree (verified_by) WHERE (verified_by IS NOT NULL);


--
-- Name: idx_driver_incidents_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_incidents_org_id ON public.driver_incidents USING btree (org_id);


--
-- Name: idx_driver_schedules_day_working; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_schedules_day_working ON public.driver_schedules USING btree (day_of_week, is_working);


--
-- Name: idx_driver_schedules_driver; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_schedules_driver ON public.driver_schedules USING btree (driver_id);


--
-- Name: idx_driver_schedules_org_day; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_schedules_org_day ON public.driver_schedules USING btree (org_id, day_of_week);


--
-- Name: idx_driver_schedules_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_driver_schedules_org_id ON public.driver_schedules USING btree (org_id);


--
-- Name: idx_drivers_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_drivers_created_by ON public.drivers USING btree (created_by);


--
-- Name: idx_drivers_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_drivers_org_id ON public.drivers USING btree (org_id);


--
-- Name: idx_drivers_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_drivers_status ON public.drivers USING btree (status);


--
-- Name: idx_drivers_updated_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_drivers_updated_by ON public.drivers USING btree (updated_by) WHERE (updated_by IS NOT NULL);


--
-- Name: idx_drivers_vehicle; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_drivers_vehicle ON public.drivers USING btree (current_vehicle_id);


--
-- Name: idx_entity_audit_history_actor; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_entity_audit_history_actor ON public.entity_audit_history USING btree (actor) WHERE (actor IS NOT NULL);


--
-- Name: idx_entity_history_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_entity_history_id ON public.entity_audit_history USING btree (entity_id);


--
-- Name: idx_entity_history_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_entity_history_type ON public.entity_audit_history USING btree (entity_type);


--
-- Name: idx_evidence_exports_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_evidence_exports_org_id ON public.evidence_exports USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_evidence_exports_requested_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_evidence_exports_requested_by ON public.evidence_exports USING btree (requested_by) WHERE (requested_by IS NOT NULL);


--
-- Name: idx_firmware_rollout_events_actor_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollout_events_actor_user_id ON public.firmware_rollout_events USING btree (actor_user_id) WHERE (actor_user_id IS NOT NULL);


--
-- Name: idx_firmware_rollout_events_org_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollout_events_org_created ON public.firmware_rollout_events USING btree (org_id, created_at DESC);


--
-- Name: idx_firmware_rollout_events_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollout_events_org_id ON public.firmware_rollout_events USING btree (org_id);


--
-- Name: idx_firmware_rollout_events_rollout_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollout_events_rollout_created ON public.firmware_rollout_events USING btree (rollout_id, created_at DESC);


--
-- Name: idx_firmware_rollouts_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollouts_created_by ON public.firmware_rollouts USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_firmware_rollouts_firmware_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollouts_firmware_id ON public.firmware_rollouts USING btree (firmware_id) WHERE (firmware_id IS NOT NULL);


--
-- Name: idx_firmware_rollouts_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollouts_org ON public.firmware_rollouts USING btree (org_id);


--
-- Name: idx_firmware_rollouts_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollouts_org_id ON public.firmware_rollouts USING btree (org_id);


--
-- Name: idx_firmware_rollouts_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_firmware_rollouts_status ON public.firmware_rollouts USING btree (status);


--
-- Name: idx_fleet_analytics_org_dates; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_fleet_analytics_org_dates ON public.fleet_analytics USING btree (org_id, start_date, end_date);


--
-- Name: idx_fleet_analytics_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_fleet_analytics_org_id ON public.fleet_analytics USING btree (org_id);


--
-- Name: idx_geofences_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_geofences_created_by ON public.geofences USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_geofences_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_geofences_org_id ON public.geofences USING btree (org_id);


--
-- Name: idx_gps_locations_vehicle_time_desc; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_gps_locations_vehicle_time_desc ON public.gps_locations USING btree (vehicle_id, "time" DESC);


--
-- Name: idx_gps_raw_data_gin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_gps_raw_data_gin ON public.gps_locations USING gin (raw_data);


--
-- Name: idx_gps_vehicle_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_gps_vehicle_time ON public.gps_locations USING btree (vehicle_id, "time" DESC);


--
-- Name: idx_gps_vehicle_time_brin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_gps_vehicle_time_brin ON public.gps_locations USING brin (vehicle_id, "time");


--
-- Name: idx_incident_comments_author_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incident_comments_author_id ON public.incident_comments USING btree (author_id) WHERE (author_id IS NOT NULL);


--
-- Name: idx_incident_comments_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incident_comments_created_at ON public.incident_comments USING btree (created_at);


--
-- Name: idx_incident_comments_ticket_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incident_comments_ticket_created ON public.incident_comments USING btree (ticket_id, created_at DESC);


--
-- Name: idx_incident_comments_ticket_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incident_comments_ticket_id ON public.incident_comments USING btree (ticket_id);


--
-- Name: idx_incidents_driver; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incidents_driver ON public.driver_incidents USING btree (driver_id);


--
-- Name: idx_incidents_driver_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incidents_driver_time ON public.driver_incidents USING btree (driver_id, "timestamp" DESC);


--
-- Name: idx_incidents_org_type_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_incidents_org_type_time ON public.driver_incidents USING btree (org_id, type, "timestamp" DESC);


--
-- Name: idx_inspection_templates_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_inspection_templates_org ON public.inspection_templates USING btree (org_id);


--
-- Name: idx_inspection_templates_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_inspection_templates_org_id ON public.inspection_templates USING btree (org_id);


--
-- Name: idx_inspections_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_inspections_created_by ON public.vehicle_inspections USING btree (created_by);


--
-- Name: idx_integration_logs_integration; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integration_logs_integration ON public.integration_logs USING btree (integration_id);


--
-- Name: idx_integration_logs_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integration_logs_org ON public.integration_logs USING btree (org_id);


--
-- Name: idx_integration_logs_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integration_logs_org_id ON public.integration_logs USING btree (org_id);


--
-- Name: idx_integration_logs_timestamp; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integration_logs_timestamp ON public.integration_logs USING btree ("timestamp" DESC);


--
-- Name: idx_integrations_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integrations_org ON public.integrations USING btree (org_id);


--
-- Name: idx_integrations_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integrations_org_id ON public.integrations USING btree (org_id);


--
-- Name: idx_integrations_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_integrations_status ON public.integrations USING btree (status);


--
-- Name: idx_invites_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invites_created_by ON public.invites USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_invites_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invites_org_id ON public.invites USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_invoice_items_invoice_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invoice_items_invoice_id ON public.invoice_items USING btree (invoice_id);


--
-- Name: idx_invoice_items_org_deleted; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invoice_items_org_deleted ON public.invoice_items USING btree (org_id, is_deleted);


--
-- Name: idx_invoice_items_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invoice_items_org_id ON public.invoice_items USING btree (org_id);


--
-- Name: idx_invoices_org_deleted; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invoices_org_deleted ON public.invoices USING btree (org_id, is_deleted);


--
-- Name: idx_invoices_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_invoices_org_id ON public.invoices USING btree (org_id);


--
-- Name: idx_ip_rules_action; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_ip_rules_action ON public.ip_rules USING btree (org_id, action);


--
-- Name: idx_ip_rules_added_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_ip_rules_added_by ON public.ip_rules USING btree (added_by) WHERE (added_by IS NOT NULL);


--
-- Name: idx_ip_rules_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_ip_rules_org ON public.ip_rules USING btree (org_id);


--
-- Name: idx_ip_rules_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_ip_rules_org_id ON public.ip_rules USING btree (org_id);


--
-- Name: idx_login_events_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_login_events_created_at ON public.login_events USING btree (created_at);


--
-- Name: idx_login_events_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_login_events_org_id ON public.login_events USING btree (org_id);


--
-- Name: idx_login_events_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_login_events_status ON public.login_events USING btree (status);


--
-- Name: idx_login_events_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_login_events_user_id ON public.login_events USING btree (user_id);


--
-- Name: idx_maintenance_vendors_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_maintenance_vendors_org ON public.maintenance_vendors USING btree (org_id);


--
-- Name: idx_maintenance_vendors_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_maintenance_vendors_org_id ON public.maintenance_vendors USING btree (org_id);


--
-- Name: idx_mfa_codes_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_codes_user_id ON public.mfa_codes USING btree (user_id);


--
-- Name: idx_mfa_events_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_events_created_at ON public.mfa_events USING btree (created_at);


--
-- Name: idx_mfa_events_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_events_org_id ON public.mfa_events USING btree (org_id);


--
-- Name: idx_mfa_events_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_events_user_id ON public.mfa_events USING btree (user_id);


--
-- Name: idx_mfa_exceptions_granted_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_exceptions_granted_by ON public.mfa_exceptions USING btree (granted_by) WHERE (granted_by IS NOT NULL);


--
-- Name: idx_mfa_exceptions_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_exceptions_org_id ON public.mfa_exceptions USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_mfa_exceptions_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_exceptions_user_id ON public.mfa_exceptions USING btree (user_id);


--
-- Name: idx_mfa_failures_location; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_failures_location ON public.mfa_failures USING btree (location);


--
-- Name: idx_mfa_failures_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_failures_org_id ON public.mfa_failures USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_mfa_failures_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_failures_user_id ON public.mfa_failures USING btree (user_id);


--
-- Name: idx_mfa_lockouts_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_lockouts_org_id ON public.mfa_lockouts USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_mfa_lockouts_unlocked_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_lockouts_unlocked_by ON public.mfa_lockouts USING btree (unlocked_by) WHERE (unlocked_by IS NOT NULL);


--
-- Name: idx_mfa_lockouts_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_lockouts_user_id ON public.mfa_lockouts USING btree (user_id);


--
-- Name: idx_mfa_org_overrides_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_org_overrides_org_id ON public.mfa_org_overrides USING btree (org_id);


--
-- Name: idx_mfa_policies_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_policies_org_id ON public.mfa_policies USING btree (org_id);


--
-- Name: idx_mfa_policy_history_changed_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_policy_history_changed_by ON public.mfa_policy_history USING btree (changed_by) WHERE (changed_by IS NOT NULL);


--
-- Name: idx_mfa_policy_history_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_policy_history_org_id ON public.mfa_policy_history USING btree (org_id);


--
-- Name: idx_mfa_users_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_mfa_users_org_id ON public.mfa_users USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_notification_events_notification_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_notification_events_notification_id ON public.notification_events USING btree (notification_id, created_at);


--
-- Name: idx_notification_logs_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_notification_logs_created_at ON public.notification_logs USING btree (created_at DESC);


--
-- Name: idx_notification_logs_status_channel; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_notification_logs_status_channel ON public.notification_logs USING btree (status, channel);


--
-- Name: idx_notification_pref_audit_user; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_notification_pref_audit_user ON public.notification_preference_audit USING btree (user_id, created_at DESC);


--
-- Name: idx_obd_dtcs_gin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_obd_dtcs_gin ON public.obd_readings USING gin (dtcs);


--
-- Name: idx_obd_raw_data_gin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_obd_raw_data_gin ON public.obd_readings USING gin (raw_data);


--
-- Name: idx_obd_vehicle_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_obd_vehicle_time ON public.obd_readings USING btree (vehicle_id, "time" DESC);


--
-- Name: idx_obd_vehicle_time_brin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_obd_vehicle_time_brin ON public.obd_readings USING brin (vehicle_id, "time");


--
-- Name: idx_org_support_metrics_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_org_support_metrics_org_id ON public.org_support_metrics USING btree (org_id);


--
-- Name: idx_organizations_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_organizations_status ON public.organizations USING btree (status, is_deleted) WHERE (is_deleted = false);


--
-- Name: idx_orgs_active_lookup; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_orgs_active_lookup ON public.organizations USING btree (is_deleted, status, created_at DESC) WHERE ((is_deleted = false) AND ((status)::text = 'active'::text));


--
-- Name: idx_pred_maint_analysis_date; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_pred_maint_analysis_date ON public.predictive_maintenance_analytics USING btree (analysis_date);


--
-- Name: idx_pred_maint_org_vehicle; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_pred_maint_org_vehicle ON public.predictive_maintenance_analytics USING btree (org_id, vehicle_id);


--
-- Name: idx_prediction_results_anomaly; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_prediction_results_anomaly ON public.prediction_results USING btree (anomaly_detected, created_at);


--
-- Name: idx_prediction_results_model_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_prediction_results_model_created ON public.prediction_results USING btree (model_id, created_at);


--
-- Name: idx_prediction_results_org_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_prediction_results_org_created ON public.prediction_results USING btree (org_id, created_at);


--
-- Name: idx_prediction_results_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_prediction_results_org_id ON public.prediction_results USING btree (org_id);


--
-- Name: idx_predictive_maintenance_analytics_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_predictive_maintenance_analytics_org_id ON public.predictive_maintenance_analytics USING btree (org_id);


--
-- Name: idx_restore_jobs_backup_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_restore_jobs_backup_id ON public.restore_jobs USING btree (backup_id);


--
-- Name: idx_restore_jobs_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_restore_jobs_status ON public.restore_jobs USING btree (status);


--
-- Name: idx_revenue_sch_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_revenue_sch_org ON public.billing_revenue_schedules USING btree (org_id);


--
-- Name: idx_review_items_review; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_review_items_review ON public.review_items USING btree (review_id);


--
-- Name: idx_review_items_reviewer_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_review_items_reviewer_id ON public.review_items USING btree (reviewer_id) WHERE (reviewer_id IS NOT NULL);


--
-- Name: idx_review_items_role_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_review_items_role_id ON public.review_items USING btree (role_id) WHERE (role_id IS NOT NULL);


--
-- Name: idx_review_items_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_review_items_status ON public.review_items USING btree (review_id, current_status);


--
-- Name: idx_review_items_user; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_review_items_user ON public.review_items USING btree (user_id);


--
-- Name: idx_route_opt_jobs_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_route_opt_jobs_status ON public.route_optimization_jobs USING btree (status);


--
-- Name: idx_route_opt_req_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_route_opt_req_org ON public.route_optimization_requests USING btree (org_id);


--
-- Name: idx_route_opt_req_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_route_opt_req_status ON public.route_optimization_requests USING btree (status);


--
-- Name: idx_route_opt_res_req; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_route_opt_res_req ON public.route_optimization_results USING btree (request_id);


--
-- Name: idx_route_optimization_jobs_request_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_route_optimization_jobs_request_id ON public.route_optimization_jobs USING btree (request_id) WHERE (request_id IS NOT NULL);


--
-- Name: idx_route_optimization_requests_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_route_optimization_requests_org_id ON public.route_optimization_requests USING btree (org_id);


--
-- Name: idx_security_events_actor_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_security_events_actor_id ON public.security_events USING btree (actor_id) WHERE (actor_id IS NOT NULL);


--
-- Name: idx_security_events_on_behalf_of; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_security_events_on_behalf_of ON public.security_events USING btree (on_behalf_of) WHERE (on_behalf_of IS NOT NULL);


--
-- Name: idx_security_events_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_security_events_user_id ON public.security_events USING btree (user_id) WHERE (user_id IS NOT NULL);


--
-- Name: idx_sla_metrics_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_sla_metrics_org_id ON public.sla_metrics USING btree (org_id) WHERE (org_id IS NOT NULL);


--
-- Name: idx_storage_pools_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_storage_pools_status ON public.storage_pools USING btree (status);


--
-- Name: idx_subs_plan_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_subs_plan_id ON public.billing_subscriptions USING btree (plan_id);


--
-- Name: idx_support_audit_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_audit_created_at ON public.support_audit_log USING btree (created_at);


--
-- Name: idx_support_audit_log_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_audit_log_user_id ON public.support_audit_log USING btree (user_id) WHERE (user_id IS NOT NULL);


--
-- Name: idx_support_audit_ticket_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_audit_ticket_id ON public.support_audit_log USING btree (ticket_id);


--
-- Name: idx_support_tickets_assigned_status_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_assigned_status_created ON public.support_tickets USING btree (assigned_to, status, created_at DESC) WHERE (assigned_to IS NOT NULL);


--
-- Name: idx_support_tickets_assigned_to; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_assigned_to ON public.support_tickets USING btree (assigned_to);


--
-- Name: idx_support_tickets_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_created_at ON public.support_tickets USING btree (created_at);


--
-- Name: idx_support_tickets_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_org_id ON public.support_tickets USING btree (org_id);


--
-- Name: idx_support_tickets_org_status_priority_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_org_status_priority_created ON public.support_tickets USING btree (org_id, status, priority, created_at DESC);


--
-- Name: INDEX idx_support_tickets_org_status_priority_created; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON INDEX public.idx_support_tickets_org_status_priority_created IS 'Optimizes support dashboard queries by org, status, priority';


--
-- Name: idx_support_tickets_priority; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_priority ON public.support_tickets USING btree (priority);


--
-- Name: idx_support_tickets_reporter_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_reporter_id ON public.support_tickets USING btree (reporter_id) WHERE (reporter_id IS NOT NULL);


--
-- Name: idx_support_tickets_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_support_tickets_status ON public.support_tickets USING btree (status);


--
-- Name: idx_tax_rates_region; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_tax_rates_region ON public.billing_tax_rates USING btree (region_id);


--
-- Name: idx_telemetry_analytics_composite; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_analytics_composite ON public.telemetry_analytics USING btree (vehicle_id, analysis_type, "timestamp" DESC);


--
-- Name: idx_telemetry_analytics_timestamp; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_analytics_timestamp ON public.telemetry_analytics USING btree ("timestamp");


--
-- Name: idx_telemetry_analytics_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_analytics_type ON public.telemetry_analytics USING btree (analysis_type);


--
-- Name: idx_telemetry_analytics_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_analytics_vehicle_id ON public.telemetry_analytics USING btree (vehicle_id);


--
-- Name: idx_telemetry_anomalies_composite; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_anomalies_composite ON public.telemetry_anomalies USING btree (vehicle_id, anomaly_type, "timestamp" DESC);


--
-- Name: idx_telemetry_anomalies_score; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_anomalies_score ON public.telemetry_anomalies USING btree (anomaly_score DESC);


--
-- Name: idx_telemetry_anomalies_timestamp; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_anomalies_timestamp ON public.telemetry_anomalies USING btree ("timestamp");


--
-- Name: idx_telemetry_anomalies_type; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_anomalies_type ON public.telemetry_anomalies USING btree (anomaly_type);


--
-- Name: idx_telemetry_anomalies_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_anomalies_vehicle_id ON public.telemetry_anomalies USING btree (vehicle_id);


--
-- Name: idx_telemetry_event_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_event_time ON public.telemetry USING btree (event_time);


--
-- Name: idx_telemetry_events_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_events_org_id ON ONLY public.telemetry_events USING btree (org_id);


--
-- Name: idx_telemetry_events_p2026_04_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_events_p2026_04_org_id ON public.telemetry_events_p2026_04 USING btree (org_id);


--
-- Name: idx_telemetry_events_p2026_05_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_events_p2026_05_org_id ON public.telemetry_events_p2026_05 USING btree (org_id);


--
-- Name: idx_telemetry_events_p2026_06_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_events_p2026_06_org_id ON public.telemetry_events_p2026_06 USING btree (org_id);


--
-- Name: idx_telemetry_events_p2026_07_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_events_p2026_07_org_id ON public.telemetry_events_p2026_07 USING btree (org_id);


--
-- Name: idx_telemetry_org_id_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_org_id_created_at ON ONLY public.telemetry_events USING btree (org_id, created_at DESC);


--
-- Name: idx_telemetry_trip_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_trip_id ON public.telemetry USING btree (trip_id);


--
-- Name: idx_telemetry_vehicle_id_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_telemetry_vehicle_id_created_at ON ONLY public.telemetry_events USING btree (vehicle_id, created_at DESC);


--
-- Name: idx_template_name; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_template_name ON public.billing_invoice_templates USING btree (name);


--
-- Name: idx_threat_alerts_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_threat_alerts_org_id ON public.threat_alerts USING btree (org_id);


--
-- Name: idx_threat_alerts_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_threat_alerts_status ON public.threat_alerts USING btree (status);


--
-- Name: idx_threat_alerts_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_threat_alerts_user_id ON public.threat_alerts USING btree (user_id) WHERE (user_id IS NOT NULL);


--
-- Name: idx_trip_settings_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trip_settings_org_id ON public.trip_settings USING btree (org_id);


--
-- Name: idx_trips_driver_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_driver_id ON public.trips USING btree (driver_id);


--
-- Name: idx_trips_org_driver_status_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_org_driver_status_time ON public.trips USING btree (org_id, driver_id, status, scheduled_start DESC);


--
-- Name: INDEX idx_trips_org_driver_status_time; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON INDEX public.idx_trips_org_driver_status_time IS 'Optimizes driver schedule and upcoming trip lookups';


--
-- Name: idx_trips_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_org_id ON public.trips USING btree (org_id);


--
-- Name: idx_trips_org_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_org_status ON public.trips USING btree (org_id, status);


--
-- Name: idx_trips_org_vehicle_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_org_vehicle_time ON public.trips USING btree (org_id, vehicle_id, scheduled_start DESC);


--
-- Name: INDEX idx_trips_org_vehicle_time; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON INDEX public.idx_trips_org_vehicle_time IS 'Speeds up vehicle trip history queries';


--
-- Name: idx_trips_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_status ON public.trips USING btree (status);


--
-- Name: idx_trips_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_trips_vehicle_id ON public.trips USING btree (vehicle_id);


--
-- Name: idx_user_links_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_links_created_by ON public.user_links USING btree (created_by) WHERE (created_by IS NOT NULL);


--
-- Name: idx_user_links_linked_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_links_linked_user_id ON public.user_links USING btree (linked_user_id);


--
-- Name: idx_user_org_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_org_org_id ON public.user_organizations USING btree (org_id);


--
-- Name: idx_user_org_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_org_user_id ON public.user_organizations USING btree (user_id);


--
-- Name: idx_user_organizations_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_organizations_org_id ON public.user_organizations USING btree (org_id);


--
-- Name: idx_user_sessions_is_active; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_sessions_is_active ON public.user_sessions USING btree (is_active);


--
-- Name: idx_user_sessions_user_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_user_sessions_user_id ON public.user_sessions USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_role_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_users_role_id ON public.users USING btree (role_id);


--
-- Name: idx_users_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_users_status ON public.users USING btree (status);


--
-- Name: idx_vehicle_inspections_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicle_inspections_org_id ON public.vehicle_inspections USING btree (org_id);


--
-- Name: idx_vehicle_inspections_template_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicle_inspections_template_id ON public.vehicle_inspections USING btree (template_id) WHERE (template_id IS NOT NULL);


--
-- Name: idx_vehicle_inspections_updated_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicle_inspections_updated_by ON public.vehicle_inspections USING btree (updated_by) WHERE (updated_by IS NOT NULL);


--
-- Name: idx_vehicle_inspections_vehicle; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicle_inspections_vehicle ON public.vehicle_inspections USING btree (vehicle_id);


--
-- Name: idx_vehicle_telemetry_alerts_gin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicle_telemetry_alerts_gin ON public.vehicle_telemetry USING gin (alerts);


--
-- Name: idx_vehicle_telemetry_vehicle_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicle_telemetry_vehicle_time ON public.vehicle_telemetry USING btree (vehicle_id, "time" DESC);


--
-- Name: idx_vehicles_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_created_by ON public.vehicles USING btree (created_by);


--
-- Name: idx_vehicles_org_created; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_org_created ON public.vehicles USING btree (org_id, created_at DESC);


--
-- Name: idx_vehicles_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_org_id ON public.vehicles USING btree (org_id);


--
-- Name: idx_vehicles_org_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_org_status ON public.vehicles USING btree (org_id, status);


--
-- Name: idx_vehicles_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_status ON public.vehicles USING btree (status);


--
-- Name: idx_vehicles_updated_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_updated_by ON public.vehicles USING btree (updated_by) WHERE (updated_by IS NOT NULL);


--
-- Name: idx_vehicles_vin; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_vehicles_vin ON public.vehicles USING btree (vin);


--
-- Name: idx_video_analytics_recording_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_analytics_recording_id ON public.video_analytics USING btree (recording_id);


--
-- Name: idx_video_analytics_stream_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_analytics_stream_id ON public.video_analytics USING btree (stream_id) WHERE (stream_id IS NOT NULL);


--
-- Name: idx_video_analytics_timestamp; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_analytics_timestamp ON public.video_analytics USING btree ("timestamp");


--
-- Name: idx_video_recordings_driver_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_driver_id ON public.video_recordings USING btree (driver_id) WHERE (driver_id IS NOT NULL);


--
-- Name: idx_video_recordings_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_org_id ON public.video_recordings USING btree (org_id);


--
-- Name: idx_video_recordings_start_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_start_time ON public.video_recordings USING btree (start_time);


--
-- Name: idx_video_recordings_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_status ON public.video_recordings USING btree (status);


--
-- Name: idx_video_recordings_stream_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_stream_id ON public.video_recordings USING btree (stream_id);


--
-- Name: idx_video_recordings_stream_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_stream_status ON public.video_recordings USING btree (stream_id, status);


--
-- Name: idx_video_recordings_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_vehicle_id ON public.video_recordings USING btree (vehicle_id);


--
-- Name: idx_video_recordings_vehicle_time; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_recordings_vehicle_time ON public.video_recordings USING btree (vehicle_id, start_time);


--
-- Name: idx_video_retention_policies_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_retention_policies_org_id ON public.video_retention_policies USING btree (org_id);


--
-- Name: idx_video_streams_driver_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_streams_driver_id ON public.video_streams USING btree (driver_id) WHERE (driver_id IS NOT NULL);


--
-- Name: idx_video_streams_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_streams_org_id ON public.video_streams USING btree (org_id);


--
-- Name: idx_video_streams_started_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_streams_started_at ON public.video_streams USING btree (started_at);


--
-- Name: idx_video_streams_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_streams_status ON public.video_streams USING btree (status);


--
-- Name: idx_video_streams_vehicle_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_streams_vehicle_id ON public.video_streams USING btree (vehicle_id);


--
-- Name: idx_video_streams_vehicle_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_streams_vehicle_status ON public.video_streams USING btree (vehicle_id, status);


--
-- Name: idx_video_thumbnails_recording_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_video_thumbnails_recording_id ON public.video_thumbnails USING btree (recording_id);


--
-- Name: idx_webhooks_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_webhooks_org ON public.webhooks USING btree (org_id);


--
-- Name: idx_webhooks_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_webhooks_org_id ON public.webhooks USING btree (org_id);


--
-- Name: idx_work_orders_created_at; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_created_at ON public.work_orders USING btree (created_at DESC);


--
-- Name: idx_work_orders_created_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_created_by ON public.work_orders USING btree (created_by);


--
-- Name: idx_work_orders_org; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_org ON public.work_orders USING btree (org_id);


--
-- Name: idx_work_orders_org_id; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_org_id ON public.work_orders USING btree (org_id);


--
-- Name: idx_work_orders_org_priority; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_org_priority ON public.work_orders USING btree (org_id, priority);


--
-- Name: idx_work_orders_org_status; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_org_status ON public.work_orders USING btree (org_id, status);


--
-- Name: idx_work_orders_updated_by; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX idx_work_orders_updated_by ON public.work_orders USING btree (updated_by) WHERE (updated_by IS NOT NULL);


--
-- Name: telemetry_events_p2026_04_org_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_04_org_id_created_at_idx ON public.telemetry_events_p2026_04 USING btree (org_id, created_at DESC);


--
-- Name: telemetry_events_p2026_04_org_id_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_04_org_id_idx ON public.telemetry_events_p2026_04 USING btree (org_id);


--
-- Name: telemetry_events_p2026_04_vehicle_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_04_vehicle_id_created_at_idx ON public.telemetry_events_p2026_04 USING btree (vehicle_id, created_at DESC);


--
-- Name: telemetry_events_p2026_05_org_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_05_org_id_created_at_idx ON public.telemetry_events_p2026_05 USING btree (org_id, created_at DESC);


--
-- Name: telemetry_events_p2026_05_org_id_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_05_org_id_idx ON public.telemetry_events_p2026_05 USING btree (org_id);


--
-- Name: telemetry_events_p2026_05_vehicle_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_05_vehicle_id_created_at_idx ON public.telemetry_events_p2026_05 USING btree (vehicle_id, created_at DESC);


--
-- Name: telemetry_events_p2026_06_org_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_06_org_id_created_at_idx ON public.telemetry_events_p2026_06 USING btree (org_id, created_at DESC);


--
-- Name: telemetry_events_p2026_06_org_id_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_06_org_id_idx ON public.telemetry_events_p2026_06 USING btree (org_id);


--
-- Name: telemetry_events_p2026_06_vehicle_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_06_vehicle_id_created_at_idx ON public.telemetry_events_p2026_06 USING btree (vehicle_id, created_at DESC);


--
-- Name: telemetry_events_p2026_07_org_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_07_org_id_created_at_idx ON public.telemetry_events_p2026_07 USING btree (org_id, created_at DESC);


--
-- Name: telemetry_events_p2026_07_org_id_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_07_org_id_idx ON public.telemetry_events_p2026_07 USING btree (org_id);


--
-- Name: telemetry_events_p2026_07_vehicle_id_created_at_idx; Type: INDEX; Schema: public; Owner: fleet_user
--

CREATE INDEX telemetry_events_p2026_07_vehicle_id_created_at_idx ON public.telemetry_events_p2026_07 USING btree (vehicle_id, created_at DESC);


--
-- Name: telemetry_events_p2026_04_org_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_org_id_created_at ATTACH PARTITION public.telemetry_events_p2026_04_org_id_created_at_idx;


--
-- Name: telemetry_events_p2026_04_org_id_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_events_org_id ATTACH PARTITION public.telemetry_events_p2026_04_org_id_idx;


--
-- Name: telemetry_events_p2026_04_vehicle_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_vehicle_id_created_at ATTACH PARTITION public.telemetry_events_p2026_04_vehicle_id_created_at_idx;


--
-- Name: telemetry_events_p2026_05_org_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_org_id_created_at ATTACH PARTITION public.telemetry_events_p2026_05_org_id_created_at_idx;


--
-- Name: telemetry_events_p2026_05_org_id_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_events_org_id ATTACH PARTITION public.telemetry_events_p2026_05_org_id_idx;


--
-- Name: telemetry_events_p2026_05_vehicle_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_vehicle_id_created_at ATTACH PARTITION public.telemetry_events_p2026_05_vehicle_id_created_at_idx;


--
-- Name: telemetry_events_p2026_06_org_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_org_id_created_at ATTACH PARTITION public.telemetry_events_p2026_06_org_id_created_at_idx;


--
-- Name: telemetry_events_p2026_06_org_id_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_events_org_id ATTACH PARTITION public.telemetry_events_p2026_06_org_id_idx;


--
-- Name: telemetry_events_p2026_06_vehicle_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_vehicle_id_created_at ATTACH PARTITION public.telemetry_events_p2026_06_vehicle_id_created_at_idx;


--
-- Name: telemetry_events_p2026_07_org_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_org_id_created_at ATTACH PARTITION public.telemetry_events_p2026_07_org_id_created_at_idx;


--
-- Name: telemetry_events_p2026_07_org_id_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_events_org_id ATTACH PARTITION public.telemetry_events_p2026_07_org_id_idx;


--
-- Name: telemetry_events_p2026_07_vehicle_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: fleet_user
--

ALTER INDEX public.idx_telemetry_vehicle_id_created_at ATTACH PARTITION public.telemetry_events_p2026_07_vehicle_id_created_at_idx;


--
-- Name: access_reviews update_access_reviews_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_access_reviews_updated_at BEFORE UPDATE ON public.access_reviews FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: alert_rules update_alert_rules_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON public.alert_rules FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: alerts update_alerts_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON public.alerts FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: analytics_jobs update_analytics_jobs_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_analytics_jobs_updated_at BEFORE UPDATE ON public.analytics_jobs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: api_clients update_api_clients_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_api_clients_updated_at BEFORE UPDATE ON public.api_clients FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: audit_policies update_audit_policies_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_audit_policies_updated_at BEFORE UPDATE ON public.audit_policies FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: audit_settings update_audit_settings_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_audit_settings_updated_at BEFORE UPDATE ON public.audit_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_customer_balances update_billing_customer_balances_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_customer_balances_updated_at BEFORE UPDATE ON public.billing_customer_balances FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_hardware_products update_billing_hardware_products_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_hardware_products_updated_at BEFORE UPDATE ON public.billing_hardware_products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_invoice_sequences update_billing_invoice_sequences_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_invoice_sequences_updated_at BEFORE UPDATE ON public.billing_invoice_sequences FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_plans update_billing_plans_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_plans_updated_at BEFORE UPDATE ON public.billing_plans FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_pricing_components update_billing_pricing_components_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_pricing_components_updated_at BEFORE UPDATE ON public.billing_pricing_components FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_settings update_billing_settings_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_settings_updated_at BEFORE UPDATE ON public.billing_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: billing_subscriptions update_billing_subscriptions_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_billing_subscriptions_updated_at BEFORE UPDATE ON public.billing_subscriptions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: bulk_config update_bulk_config_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_bulk_config_updated_at BEFORE UPDATE ON public.bulk_config FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: bulk_jobs update_bulk_jobs_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_bulk_jobs_updated_at BEFORE UPDATE ON public.bulk_jobs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: documents update_documents_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON public.documents FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: dr_test_settings update_dr_test_settings_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_dr_test_settings_updated_at BEFORE UPDATE ON public.dr_test_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: dr_tests update_dr_tests_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_dr_tests_updated_at BEFORE UPDATE ON public.dr_tests FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: driver_attendance update_driver_attendance_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_driver_attendance_updated_at BEFORE UPDATE ON public.driver_attendance FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: driver_behavior_analytics update_driver_behavior_analytics_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_driver_behavior_analytics_updated_at BEFORE UPDATE ON public.driver_behavior_analytics FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: driver_behavior update_driver_behavior_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_driver_behavior_updated_at BEFORE UPDATE ON public.driver_behavior FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: driver_documents update_driver_documents_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_driver_documents_updated_at BEFORE UPDATE ON public.driver_documents FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: driver_incidents update_driver_incidents_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_driver_incidents_updated_at BEFORE UPDATE ON public.driver_incidents FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: driver_schedules update_driver_schedules_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_driver_schedules_updated_at BEFORE UPDATE ON public.driver_schedules FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: drivers update_drivers_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON public.drivers FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: firmware_packages update_firmware_packages_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_firmware_packages_updated_at BEFORE UPDATE ON public.firmware_packages FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: fleet_analytics update_fleet_analytics_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_fleet_analytics_updated_at BEFORE UPDATE ON public.fleet_analytics FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: geofences update_geofences_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_geofences_updated_at BEFORE UPDATE ON public.geofences FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: global_notification_defaults update_global_notification_defaults_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_global_notification_defaults_updated_at BEFORE UPDATE ON public.global_notification_defaults FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: incident_comments update_incident_comments_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_incident_comments_updated_at BEFORE UPDATE ON public.incident_comments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: inspection_templates update_inspection_templates_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_inspection_templates_updated_at BEFORE UPDATE ON public.inspection_templates FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: integrations update_integrations_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_integrations_updated_at BEFORE UPDATE ON public.integrations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: invoice_items update_invoice_items_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_invoice_items_updated_at BEFORE UPDATE ON public.invoice_items FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: invoices update_invoices_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_invoices_updated_at BEFORE UPDATE ON public.invoices FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: ip_rules update_ip_rules_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_ip_rules_updated_at BEFORE UPDATE ON public.ip_rules FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: maintenance_vendors update_maintenance_vendors_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_maintenance_vendors_updated_at BEFORE UPDATE ON public.maintenance_vendors FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: mfa_org_overrides update_mfa_org_overrides_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_mfa_org_overrides_updated_at BEFORE UPDATE ON public.mfa_org_overrides FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: mfa_policies update_mfa_policies_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_mfa_policies_updated_at BEFORE UPDATE ON public.mfa_policies FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: mfa_users update_mfa_users_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_mfa_users_updated_at BEFORE UPDATE ON public.mfa_users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: notification_logs update_notification_logs_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_notification_logs_updated_at BEFORE UPDATE ON public.notification_logs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: notification_settings update_notification_settings_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_notification_settings_updated_at BEFORE UPDATE ON public.notification_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: org_support_metrics update_org_support_metrics_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_org_support_metrics_updated_at BEFORE UPDATE ON public.org_support_metrics FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: organizations update_organizations_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON public.organizations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: prediction_results update_prediction_results_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_prediction_results_updated_at BEFORE UPDATE ON public.prediction_results FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: predictive_maintenance_analytics update_predictive_maintenance_analytics_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_predictive_maintenance_analytics_updated_at BEFORE UPDATE ON public.predictive_maintenance_analytics FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: predictive_models update_predictive_models_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_predictive_models_updated_at BEFORE UPDATE ON public.predictive_models FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: roles update_roles_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON public.roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: route_optimization_jobs update_route_optimization_jobs_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_route_optimization_jobs_updated_at BEFORE UPDATE ON public.route_optimization_jobs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: route_optimization_requests update_route_optimization_requests_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_route_optimization_requests_updated_at BEFORE UPDATE ON public.route_optimization_requests FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: storage_pools update_storage_pools_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_storage_pools_updated_at BEFORE UPDATE ON public.storage_pools FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: support_tickets update_support_tickets_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_support_tickets_updated_at BEFORE UPDATE ON public.support_tickets FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: trip_settings update_trip_settings_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_trip_settings_updated_at BEFORE UPDATE ON public.trip_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: trips update_trips_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON public.trips FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: user_notification_preferences update_user_notification_preferences_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_user_notification_preferences_updated_at BEFORE UPDATE ON public.user_notification_preferences FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: vehicle_inspections update_vehicle_inspections_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_vehicle_inspections_updated_at BEFORE UPDATE ON public.vehicle_inspections FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: vehicles update_vehicles_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON public.vehicles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: video_recordings update_video_recordings_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_video_recordings_updated_at BEFORE UPDATE ON public.video_recordings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: video_retention_policies update_video_retention_policies_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_video_retention_policies_updated_at BEFORE UPDATE ON public.video_retention_policies FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: video_streams update_video_streams_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_video_streams_updated_at BEFORE UPDATE ON public.video_streams FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: webhooks update_webhooks_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON public.webhooks FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: work_orders update_work_orders_updated_at; Type: TRIGGER; Schema: public; Owner: fleet_user
--

CREATE TRIGGER update_work_orders_updated_at BEFORE UPDATE ON public.work_orders FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: access_reviews access_reviews_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.access_reviews
    ADD CONSTRAINT access_reviews_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: access_reviews access_reviews_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.access_reviews
    ADD CONSTRAINT access_reviews_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: alert_rules alert_rules_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: alert_rules alert_rules_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT alert_rules_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: alerts alerts_acknowledged_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_acknowledged_by_fkey FOREIGN KEY (acknowledged_by) REFERENCES public.users(id);


--
-- Name: alerts alerts_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: alerts alerts_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: alerts alerts_resolved_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_resolved_by_fkey FOREIGN KEY (resolved_by) REFERENCES public.users(id);


--
-- Name: alerts alerts_rule_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_rule_id_fkey FOREIGN KEY (rule_id) REFERENCES public.alert_rules(id);


--
-- Name: alerts alerts_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT alerts_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: api_clients api_clients_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.api_clients
    ADD CONSTRAINT api_clients_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: audit_logs audit_logs_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: audit_logs audit_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: audit_policies audit_policies_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_policies
    ADD CONSTRAINT audit_policies_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: audit_policies audit_policies_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_policies
    ADD CONSTRAINT audit_policies_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: audit_settings audit_settings_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_settings
    ADD CONSTRAINT audit_settings_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: auth_tokens auth_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: backup_verifications backup_verifications_backup_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.backup_verifications
    ADD CONSTRAINT backup_verifications_backup_id_fkey FOREIGN KEY (backup_id) REFERENCES public.backups(id) ON DELETE CASCADE;


--
-- Name: backup_verifications backup_verifications_verified_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.backup_verifications
    ADD CONSTRAINT backup_verifications_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES public.users(id);


--
-- Name: billing_accounting_exports billing_accounting_exports_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_accounting_exports
    ADD CONSTRAINT billing_accounting_exports_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: billing_credit_transactions billing_credit_transactions_balance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_credit_transactions
    ADD CONSTRAINT billing_credit_transactions_balance_id_fkey FOREIGN KEY (balance_id) REFERENCES public.billing_customer_balances(id);


--
-- Name: billing_customer_balances billing_customer_balances_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_customer_balances
    ADD CONSTRAINT billing_customer_balances_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: billing_invoice_documents billing_invoice_documents_invoice_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_documents
    ADD CONSTRAINT billing_invoice_documents_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES public.invoices(id) ON DELETE CASCADE;


--
-- Name: billing_invoice_documents billing_invoice_documents_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_documents
    ADD CONSTRAINT billing_invoice_documents_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.billing_invoice_templates(id);


--
-- Name: billing_invoice_sequences billing_invoice_sequences_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_sequences
    ADD CONSTRAINT billing_invoice_sequences_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: billing_pricing_components billing_pricing_components_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_pricing_components
    ADD CONSTRAINT billing_pricing_components_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.billing_plans(id) ON DELETE CASCADE;


--
-- Name: billing_recognized_revenue billing_recognized_revenue_schedule_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_recognized_revenue
    ADD CONSTRAINT billing_recognized_revenue_schedule_id_fkey FOREIGN KEY (schedule_id) REFERENCES public.billing_revenue_schedules(id);


--
-- Name: billing_revenue_schedules billing_revenue_schedules_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_revenue_schedules
    ADD CONSTRAINT billing_revenue_schedules_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: billing_subscriptions billing_subscriptions_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_subscriptions
    ADD CONSTRAINT billing_subscriptions_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: billing_subscriptions billing_subscriptions_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_subscriptions
    ADD CONSTRAINT billing_subscriptions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.billing_plans(id);


--
-- Name: billing_tax_rates billing_tax_rates_region_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_tax_rates
    ADD CONSTRAINT billing_tax_rates_region_id_fkey FOREIGN KEY (region_id) REFERENCES public.billing_tax_regions(id);


--
-- Name: bulk_job_records bulk_job_records_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.bulk_job_records
    ADD CONSTRAINT bulk_job_records_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.bulk_jobs(id) ON DELETE CASCADE;


--
-- Name: consumption_events consumption_events_meter_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.consumption_events
    ADD CONSTRAINT consumption_events_meter_code_fkey FOREIGN KEY (meter_code) REFERENCES public.usage_meters(code);


--
-- Name: consumption_events consumption_events_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.consumption_events
    ADD CONSTRAINT consumption_events_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: consumption_events consumption_events_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.consumption_events
    ADD CONSTRAINT consumption_events_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: documents documents_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: documents documents_uploaded_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_uploaded_by_fkey FOREIGN KEY (uploaded_by) REFERENCES public.users(id);


--
-- Name: documents documents_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id) ON DELETE CASCADE;


--
-- Name: dr_test_logs dr_test_logs_test_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_test_logs
    ADD CONSTRAINT dr_test_logs_test_id_fkey FOREIGN KEY (test_id) REFERENCES public.dr_tests(id) ON DELETE CASCADE;


--
-- Name: dr_test_steps dr_test_steps_test_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.dr_test_steps
    ADD CONSTRAINT dr_test_steps_test_id_fkey FOREIGN KEY (test_id) REFERENCES public.dr_tests(id) ON DELETE CASCADE;


--
-- Name: driver_attendance driver_attendance_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_attendance
    ADD CONSTRAINT driver_attendance_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id) ON DELETE CASCADE;


--
-- Name: driver_attendance driver_attendance_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_attendance
    ADD CONSTRAINT driver_attendance_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: driver_behavior driver_behavior_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_behavior
    ADD CONSTRAINT driver_behavior_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: driver_documents driver_documents_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: driver_documents driver_documents_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id) ON DELETE CASCADE;


--
-- Name: driver_documents driver_documents_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: driver_documents driver_documents_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: driver_documents driver_documents_verified_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT driver_documents_verified_by_fkey FOREIGN KEY (verified_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: driver_incidents driver_incidents_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_incidents
    ADD CONSTRAINT driver_incidents_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: driver_schedules driver_schedules_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_schedules
    ADD CONSTRAINT driver_schedules_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id) ON DELETE CASCADE;


--
-- Name: driver_schedules driver_schedules_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_schedules
    ADD CONSTRAINT driver_schedules_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: driver_verifications driver_verifications_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_verifications
    ADD CONSTRAINT driver_verifications_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: driver_verifications driver_verifications_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_verifications
    ADD CONSTRAINT driver_verifications_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: drivers drivers_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: drivers drivers_current_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_current_vehicle_id_fkey FOREIGN KEY (current_vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: drivers drivers_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: entity_audit_history entity_audit_history_actor_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.entity_audit_history
    ADD CONSTRAINT entity_audit_history_actor_fkey FOREIGN KEY (actor) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: evidence_exports evidence_exports_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.evidence_exports
    ADD CONSTRAINT evidence_exports_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: evidence_exports evidence_exports_requested_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.evidence_exports
    ADD CONSTRAINT evidence_exports_requested_by_fkey FOREIGN KEY (requested_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: firmware_rollout_events firmware_rollout_events_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollout_events
    ADD CONSTRAINT firmware_rollout_events_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);


--
-- Name: firmware_rollout_events firmware_rollout_events_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollout_events
    ADD CONSTRAINT firmware_rollout_events_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: firmware_rollout_events firmware_rollout_events_rollout_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollout_events
    ADD CONSTRAINT firmware_rollout_events_rollout_id_fkey FOREIGN KEY (rollout_id) REFERENCES public.firmware_rollouts(id) ON DELETE CASCADE;


--
-- Name: firmware_rollouts firmware_rollouts_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollouts
    ADD CONSTRAINT firmware_rollouts_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: firmware_rollouts firmware_rollouts_firmware_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollouts
    ADD CONSTRAINT firmware_rollouts_firmware_id_fkey FOREIGN KEY (firmware_id) REFERENCES public.firmware_packages(id) ON DELETE CASCADE;


--
-- Name: firmware_rollouts firmware_rollouts_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollouts
    ADD CONSTRAINT firmware_rollouts_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: access_reviews fk_access_reviews_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.access_reviews
    ADD CONSTRAINT fk_access_reviews_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_access_reviews_organization ON access_reviews; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_access_reviews_organization ON public.access_reviews IS 'Ensures access_reviews belongs to valid organization';


--
-- Name: alert_rules fk_alert_rules_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alert_rules
    ADD CONSTRAINT fk_alert_rules_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_alert_rules_organization ON alert_rules; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_alert_rules_organization ON public.alert_rules IS 'Ensures alert_rules belongs to valid organization';


--
-- Name: alerts fk_alerts_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.alerts
    ADD CONSTRAINT fk_alerts_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_alerts_organization ON alerts; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_alerts_organization ON public.alerts IS 'Ensures alerts belongs to valid organization';


--
-- Name: analytics_jobs fk_analytics_jobs_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.analytics_jobs
    ADD CONSTRAINT fk_analytics_jobs_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: api_clients fk_api_clients_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.api_clients
    ADD CONSTRAINT fk_api_clients_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_api_clients_organization ON api_clients; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_api_clients_organization ON public.api_clients IS 'Ensures api_clients belongs to valid organization';


--
-- Name: audit_logs fk_audit_logs_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_audit_logs_organization ON audit_logs; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_audit_logs_organization ON public.audit_logs IS 'Ensures audit_logs belongs to valid organization';


--
-- Name: audit_policies fk_audit_policies_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_policies
    ADD CONSTRAINT fk_audit_policies_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_audit_policies_organization ON audit_policies; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_audit_policies_organization ON public.audit_policies IS 'Ensures audit_policies belongs to valid organization';


--
-- Name: audit_settings fk_audit_settings_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.audit_settings
    ADD CONSTRAINT fk_audit_settings_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_audit_settings_organization ON audit_settings; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_audit_settings_organization ON public.audit_settings IS 'Ensures audit_settings belongs to valid organization';


--
-- Name: billing_accounting_exports fk_billing_accounting_exports_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_accounting_exports
    ADD CONSTRAINT fk_billing_accounting_exports_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_billing_accounting_exports_organization ON billing_accounting_exports; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_billing_accounting_exports_organization ON public.billing_accounting_exports IS 'Ensures billing_accounting_exports belongs to valid organization';


--
-- Name: billing_customer_balances fk_billing_customer_balances_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_customer_balances
    ADD CONSTRAINT fk_billing_customer_balances_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_billing_customer_balances_organization ON billing_customer_balances; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_billing_customer_balances_organization ON public.billing_customer_balances IS 'Ensures billing_customer_balances belongs to valid organization';


--
-- Name: billing_invoice_sequences fk_billing_invoice_sequences_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_invoice_sequences
    ADD CONSTRAINT fk_billing_invoice_sequences_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_billing_invoice_sequences_organization ON billing_invoice_sequences; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_billing_invoice_sequences_organization ON public.billing_invoice_sequences IS 'Ensures billing_invoice_sequences belongs to valid organization';


--
-- Name: billing_revenue_schedules fk_billing_revenue_schedules_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_revenue_schedules
    ADD CONSTRAINT fk_billing_revenue_schedules_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_billing_revenue_schedules_organization ON billing_revenue_schedules; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_billing_revenue_schedules_organization ON public.billing_revenue_schedules IS 'Ensures billing_revenue_schedules belongs to valid organization';


--
-- Name: billing_settings fk_billing_settings_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_settings
    ADD CONSTRAINT fk_billing_settings_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_billing_settings_organization ON billing_settings; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_billing_settings_organization ON public.billing_settings IS 'Ensures billing_settings belongs to valid organization';


--
-- Name: billing_subscriptions fk_billing_subscriptions_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.billing_subscriptions
    ADD CONSTRAINT fk_billing_subscriptions_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_billing_subscriptions_organization ON billing_subscriptions; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_billing_subscriptions_organization ON public.billing_subscriptions IS 'Ensures billing_subscriptions belongs to valid organization';


--
-- Name: bulk_config fk_bulk_config_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.bulk_config
    ADD CONSTRAINT fk_bulk_config_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: bulk_jobs fk_bulk_jobs_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.bulk_jobs
    ADD CONSTRAINT fk_bulk_jobs_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: consumption_events fk_consumption_events_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.consumption_events
    ADD CONSTRAINT fk_consumption_events_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_consumption_events_organization ON consumption_events; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_consumption_events_organization ON public.consumption_events IS 'Ensures consumption_events belongs to valid organization';


--
-- Name: documents fk_documents_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT fk_documents_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_documents_organization ON documents; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_documents_organization ON public.documents IS 'Ensures documents belongs to valid organization';


--
-- Name: driver_attendance fk_driver_attendance_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_attendance
    ADD CONSTRAINT fk_driver_attendance_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_driver_attendance_organization ON driver_attendance; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_driver_attendance_organization ON public.driver_attendance IS 'Ensures driver_attendance belongs to valid organization';


--
-- Name: driver_behavior_analytics fk_driver_behavior_analytics_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_behavior_analytics
    ADD CONSTRAINT fk_driver_behavior_analytics_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: driver_behavior fk_driver_behavior_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_behavior
    ADD CONSTRAINT fk_driver_behavior_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_driver_behavior_organization ON driver_behavior; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_driver_behavior_organization ON public.driver_behavior IS 'Ensures driver_behavior belongs to valid organization';


--
-- Name: driver_documents fk_driver_documents_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_documents
    ADD CONSTRAINT fk_driver_documents_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: driver_incidents fk_driver_incidents_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_incidents
    ADD CONSTRAINT fk_driver_incidents_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_driver_incidents_organization ON driver_incidents; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_driver_incidents_organization ON public.driver_incidents IS 'Ensures driver_incidents belongs to valid organization';


--
-- Name: driver_schedules fk_driver_schedules_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.driver_schedules
    ADD CONSTRAINT fk_driver_schedules_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_driver_schedules_organization ON driver_schedules; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_driver_schedules_organization ON public.driver_schedules IS 'Ensures driver_schedules belongs to valid organization';


--
-- Name: drivers fk_drivers_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT fk_drivers_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_drivers_organization ON drivers; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_drivers_organization ON public.drivers IS 'Ensures drivers belongs to valid organization';


--
-- Name: evidence_exports fk_evidence_exports_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.evidence_exports
    ADD CONSTRAINT fk_evidence_exports_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_evidence_exports_organization ON evidence_exports; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_evidence_exports_organization ON public.evidence_exports IS 'Ensures evidence_exports belongs to valid organization';


--
-- Name: firmware_rollout_events fk_firmware_rollout_events_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollout_events
    ADD CONSTRAINT fk_firmware_rollout_events_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_firmware_rollout_events_organization ON firmware_rollout_events; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_firmware_rollout_events_organization ON public.firmware_rollout_events IS 'Ensures firmware_rollout_events belongs to valid organization';


--
-- Name: firmware_rollouts fk_firmware_rollouts_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.firmware_rollouts
    ADD CONSTRAINT fk_firmware_rollouts_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_firmware_rollouts_organization ON firmware_rollouts; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_firmware_rollouts_organization ON public.firmware_rollouts IS 'Ensures firmware_rollouts belongs to valid organization';


--
-- Name: fleet_analytics fk_fleet_analytics_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.fleet_analytics
    ADD CONSTRAINT fk_fleet_analytics_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: geofences fk_geofences_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.geofences
    ADD CONSTRAINT fk_geofences_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_geofences_organization ON geofences; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_geofences_organization ON public.geofences IS 'Ensures geofences belongs to valid organization';


--
-- Name: integration_logs fk_integration_logs_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integration_logs
    ADD CONSTRAINT fk_integration_logs_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_integration_logs_organization ON integration_logs; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_integration_logs_organization ON public.integration_logs IS 'Ensures integration_logs belongs to valid organization';


--
-- Name: integrations fk_integrations_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT fk_integrations_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_integrations_organization ON integrations; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_integrations_organization ON public.integrations IS 'Ensures integrations belongs to valid organization';


--
-- Name: invites fk_invites_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invites
    ADD CONSTRAINT fk_invites_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_invites_organization ON invites; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_invites_organization ON public.invites IS 'Ensures invites belongs to valid organization';


--
-- Name: invoice_items fk_invoice_items_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invoice_items
    ADD CONSTRAINT fk_invoice_items_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_invoice_items_organization ON invoice_items; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_invoice_items_organization ON public.invoice_items IS 'Ensures invoice_items belongs to valid organization';


--
-- Name: invoices fk_invoices_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT fk_invoices_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_invoices_organization ON invoices; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_invoices_organization ON public.invoices IS 'Ensures invoices belongs to valid organization';


--
-- Name: ip_rules fk_ip_rules_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.ip_rules
    ADD CONSTRAINT fk_ip_rules_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_ip_rules_organization ON ip_rules; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_ip_rules_organization ON public.ip_rules IS 'Ensures ip_rules belongs to valid organization';


--
-- Name: login_events fk_login_events_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.login_events
    ADD CONSTRAINT fk_login_events_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_login_events_organization ON login_events; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_login_events_organization ON public.login_events IS 'Ensures login_events belongs to valid organization';


--
-- Name: maintenance_vendors fk_maintenance_vendors_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.maintenance_vendors
    ADD CONSTRAINT fk_maintenance_vendors_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_maintenance_vendors_organization ON maintenance_vendors; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_maintenance_vendors_organization ON public.maintenance_vendors IS 'Ensures maintenance_vendors belongs to valid organization';


--
-- Name: mfa_events fk_mfa_events_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_events
    ADD CONSTRAINT fk_mfa_events_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_events_organization ON mfa_events; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_events_organization ON public.mfa_events IS 'Ensures mfa_events belongs to valid organization';


--
-- Name: mfa_exceptions fk_mfa_exceptions_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_exceptions
    ADD CONSTRAINT fk_mfa_exceptions_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_exceptions_organization ON mfa_exceptions; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_exceptions_organization ON public.mfa_exceptions IS 'Ensures mfa_exceptions belongs to valid organization';


--
-- Name: mfa_failures fk_mfa_failures_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_failures
    ADD CONSTRAINT fk_mfa_failures_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_failures_organization ON mfa_failures; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_failures_organization ON public.mfa_failures IS 'Ensures mfa_failures belongs to valid organization';


--
-- Name: mfa_lockouts fk_mfa_lockouts_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_lockouts
    ADD CONSTRAINT fk_mfa_lockouts_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_lockouts_organization ON mfa_lockouts; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_lockouts_organization ON public.mfa_lockouts IS 'Ensures mfa_lockouts belongs to valid organization';


--
-- Name: mfa_org_overrides fk_mfa_org_overrides_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_org_overrides
    ADD CONSTRAINT fk_mfa_org_overrides_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_org_overrides_organization ON mfa_org_overrides; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_org_overrides_organization ON public.mfa_org_overrides IS 'Ensures mfa_org_overrides belongs to valid organization';


--
-- Name: mfa_policies fk_mfa_policies_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policies
    ADD CONSTRAINT fk_mfa_policies_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_policies_organization ON mfa_policies; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_policies_organization ON public.mfa_policies IS 'Ensures mfa_policies belongs to valid organization';


--
-- Name: mfa_policy_history fk_mfa_policy_history_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policy_history
    ADD CONSTRAINT fk_mfa_policy_history_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_policy_history_organization ON mfa_policy_history; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_policy_history_organization ON public.mfa_policy_history IS 'Ensures mfa_policy_history belongs to valid organization';


--
-- Name: mfa_users fk_mfa_users_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_users
    ADD CONSTRAINT fk_mfa_users_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_mfa_users_organization ON mfa_users; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_mfa_users_organization ON public.mfa_users IS 'Ensures mfa_users belongs to valid organization';


--
-- Name: org_support_metrics fk_org_support_metrics_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.org_support_metrics
    ADD CONSTRAINT fk_org_support_metrics_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_org_support_metrics_organization ON org_support_metrics; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_org_support_metrics_organization ON public.org_support_metrics IS 'Ensures org_support_metrics belongs to valid organization';


--
-- Name: prediction_results fk_prediction_results_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.prediction_results
    ADD CONSTRAINT fk_prediction_results_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: predictive_maintenance_analytics fk_predictive_maintenance_analytics_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.predictive_maintenance_analytics
    ADD CONSTRAINT fk_predictive_maintenance_analytics_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: route_optimization_requests fk_route_optimization_requests_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_requests
    ADD CONSTRAINT fk_route_optimization_requests_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: sla_metrics fk_sla_metrics_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.sla_metrics
    ADD CONSTRAINT fk_sla_metrics_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_sla_metrics_organization ON sla_metrics; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_sla_metrics_organization ON public.sla_metrics IS 'Ensures sla_metrics belongs to valid organization';


--
-- Name: support_tickets fk_support_tickets_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_support_tickets_organization ON support_tickets; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_support_tickets_organization ON public.support_tickets IS 'Ensures support_tickets belongs to valid organization';


--
-- Name: telemetry_analytics fk_telemetry_analytics_vehicle; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_analytics
    ADD CONSTRAINT fk_telemetry_analytics_vehicle FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: telemetry_anomalies fk_telemetry_anomalies_vehicle; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_anomalies
    ADD CONSTRAINT fk_telemetry_anomalies_vehicle FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: telemetry_events fk_telemetry_events_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events
    ADD CONSTRAINT fk_telemetry_events_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_telemetry_events_organization ON telemetry_events; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_telemetry_events_organization ON public.telemetry_events IS 'Ensures telemetry_events belongs to valid organization';


--
-- Name: telemetry_events_p2026_04 fk_telemetry_events_p2026_04_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events_p2026_04
    ADD CONSTRAINT fk_telemetry_events_p2026_04_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_telemetry_events_p2026_04_organization ON telemetry_events_p2026_04; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_telemetry_events_p2026_04_organization ON public.telemetry_events_p2026_04 IS 'Ensures telemetry_events_p2026_04 belongs to valid organization';


--
-- Name: telemetry_events_p2026_05 fk_telemetry_events_p2026_05_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events_p2026_05
    ADD CONSTRAINT fk_telemetry_events_p2026_05_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_telemetry_events_p2026_05_organization ON telemetry_events_p2026_05; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_telemetry_events_p2026_05_organization ON public.telemetry_events_p2026_05 IS 'Ensures telemetry_events_p2026_05 belongs to valid organization';


--
-- Name: telemetry_events_p2026_06 fk_telemetry_events_p2026_06_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events_p2026_06
    ADD CONSTRAINT fk_telemetry_events_p2026_06_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_telemetry_events_p2026_06_organization ON telemetry_events_p2026_06; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_telemetry_events_p2026_06_organization ON public.telemetry_events_p2026_06 IS 'Ensures telemetry_events_p2026_06 belongs to valid organization';


--
-- Name: telemetry_events_p2026_07 fk_telemetry_events_p2026_07_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry_events_p2026_07
    ADD CONSTRAINT fk_telemetry_events_p2026_07_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_telemetry_events_p2026_07_organization ON telemetry_events_p2026_07; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_telemetry_events_p2026_07_organization ON public.telemetry_events_p2026_07 IS 'Ensures telemetry_events_p2026_07 belongs to valid organization';


--
-- Name: threat_alerts fk_threat_alerts_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.threat_alerts
    ADD CONSTRAINT fk_threat_alerts_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_threat_alerts_organization ON threat_alerts; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_threat_alerts_organization ON public.threat_alerts IS 'Ensures threat_alerts belongs to valid organization';


--
-- Name: trip_settings fk_trip_settings_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trip_settings
    ADD CONSTRAINT fk_trip_settings_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: trips fk_trips_driver; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT fk_trips_driver FOREIGN KEY (driver_id) REFERENCES public.drivers(id) ON DELETE SET NULL;


--
-- Name: CONSTRAINT fk_trips_driver ON trips; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_trips_driver ON public.trips IS 'Ensures trip driver references valid driver in same organization';


--
-- Name: trips fk_trips_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT fk_trips_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_trips_organization ON trips; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_trips_organization ON public.trips IS 'Ensures trip belongs to valid organization';


--
-- Name: user_organizations fk_user_organizations_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT fk_user_organizations_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_user_organizations_organization ON user_organizations; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_user_organizations_organization ON public.user_organizations IS 'Ensures user_organizations belongs to valid organization';


--
-- Name: users fk_users_role; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE SET NULL;


--
-- Name: CONSTRAINT fk_users_role ON users; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_users_role ON public.users IS 'Links user to RBAC role';


--
-- Name: vehicle_inspections fk_vehicle_inspections_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_inspections
    ADD CONSTRAINT fk_vehicle_inspections_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_vehicle_inspections_organization ON vehicle_inspections; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_vehicle_inspections_organization ON public.vehicle_inspections IS 'Ensures vehicle_inspections belongs to valid organization';


--
-- Name: vehicles fk_vehicles_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT fk_vehicles_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_vehicles_organization ON vehicles; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_vehicles_organization ON public.vehicles IS 'Ensures vehicles belongs to valid organization';


--
-- Name: video_analytics fk_video_analytics_recording; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_analytics
    ADD CONSTRAINT fk_video_analytics_recording FOREIGN KEY (recording_id) REFERENCES public.video_recordings(id);


--
-- Name: video_analytics fk_video_analytics_stream; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_analytics
    ADD CONSTRAINT fk_video_analytics_stream FOREIGN KEY (stream_id) REFERENCES public.video_streams(id);


--
-- Name: video_recordings fk_video_recordings_driver; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_recordings
    ADD CONSTRAINT fk_video_recordings_driver FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: video_recordings fk_video_recordings_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_recordings
    ADD CONSTRAINT fk_video_recordings_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: video_recordings fk_video_recordings_stream; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_recordings
    ADD CONSTRAINT fk_video_recordings_stream FOREIGN KEY (stream_id) REFERENCES public.video_streams(id);


--
-- Name: video_recordings fk_video_recordings_vehicle; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_recordings
    ADD CONSTRAINT fk_video_recordings_vehicle FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: video_retention_policies fk_video_retention_policies_org; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_retention_policies
    ADD CONSTRAINT fk_video_retention_policies_org FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: video_retention_policies fk_video_retention_policies_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_retention_policies
    ADD CONSTRAINT fk_video_retention_policies_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: video_streams fk_video_streams_driver; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_streams
    ADD CONSTRAINT fk_video_streams_driver FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: video_streams fk_video_streams_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_streams
    ADD CONSTRAINT fk_video_streams_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: video_streams fk_video_streams_vehicle; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_streams
    ADD CONSTRAINT fk_video_streams_vehicle FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: video_thumbnails fk_video_thumbnails_recording; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.video_thumbnails
    ADD CONSTRAINT fk_video_thumbnails_recording FOREIGN KEY (recording_id) REFERENCES public.video_recordings(id);


--
-- Name: webhooks fk_webhooks_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT fk_webhooks_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_webhooks_organization ON webhooks; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_webhooks_organization ON public.webhooks IS 'Ensures webhooks belongs to valid organization';


--
-- Name: work_orders fk_work_orders_organization; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.work_orders
    ADD CONSTRAINT fk_work_orders_organization FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: CONSTRAINT fk_work_orders_organization ON work_orders; Type: COMMENT; Schema: public; Owner: fleet_user
--

COMMENT ON CONSTRAINT fk_work_orders_organization ON public.work_orders IS 'Ensures work order belongs to valid organization';


--
-- Name: geofences geofences_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.geofences
    ADD CONSTRAINT geofences_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: geofences geofences_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.geofences
    ADD CONSTRAINT geofences_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: governance_settings governance_settings_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.governance_settings
    ADD CONSTRAINT governance_settings_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: gst_compliances gst_compliances_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.gst_compliances
    ADD CONSTRAINT gst_compliances_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: incident_comments incident_comments_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.incident_comments
    ADD CONSTRAINT incident_comments_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: incident_comments incident_comments_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.incident_comments
    ADD CONSTRAINT incident_comments_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.support_tickets(id) ON DELETE CASCADE;


--
-- Name: integration_configs integration_configs_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integration_configs
    ADD CONSTRAINT integration_configs_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: integration_logs integration_logs_integration_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integration_logs
    ADD CONSTRAINT integration_logs_integration_id_fkey FOREIGN KEY (integration_id) REFERENCES public.integrations(id) ON DELETE SET NULL;


--
-- Name: integration_logs integration_logs_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integration_logs
    ADD CONSTRAINT integration_logs_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: integrations integrations_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT integrations_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: invites invites_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invites
    ADD CONSTRAINT invites_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: invites invites_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invites
    ADD CONSTRAINT invites_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: invoice_items invoice_items_invoice_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.invoice_items
    ADD CONSTRAINT invoice_items_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES public.invoices(id) ON DELETE CASCADE;


--
-- Name: ip_rules ip_rules_added_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.ip_rules
    ADD CONSTRAINT ip_rules_added_by_fkey FOREIGN KEY (added_by) REFERENCES public.users(id);


--
-- Name: ip_rules ip_rules_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.ip_rules
    ADD CONSTRAINT ip_rules_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: local_permits local_permits_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.local_permits
    ADD CONSTRAINT local_permits_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: local_permits local_permits_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.local_permits
    ADD CONSTRAINT local_permits_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: local_permits local_permits_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.local_permits
    ADD CONSTRAINT local_permits_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: login_events login_events_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.login_events
    ADD CONSTRAINT login_events_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: login_events login_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.login_events
    ADD CONSTRAINT login_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: mfa_codes mfa_codes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_codes
    ADD CONSTRAINT mfa_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mfa_events mfa_events_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_events
    ADD CONSTRAINT mfa_events_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_events mfa_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_events
    ADD CONSTRAINT mfa_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mfa_exceptions mfa_exceptions_granted_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_exceptions
    ADD CONSTRAINT mfa_exceptions_granted_by_fkey FOREIGN KEY (granted_by) REFERENCES public.users(id);


--
-- Name: mfa_exceptions mfa_exceptions_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_exceptions
    ADD CONSTRAINT mfa_exceptions_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_exceptions mfa_exceptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_exceptions
    ADD CONSTRAINT mfa_exceptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mfa_failures mfa_failures_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_failures
    ADD CONSTRAINT mfa_failures_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_failures mfa_failures_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_failures
    ADD CONSTRAINT mfa_failures_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mfa_lockouts mfa_lockouts_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_lockouts
    ADD CONSTRAINT mfa_lockouts_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_lockouts mfa_lockouts_unlocked_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_lockouts
    ADD CONSTRAINT mfa_lockouts_unlocked_by_fkey FOREIGN KEY (unlocked_by) REFERENCES public.users(id);


--
-- Name: mfa_lockouts mfa_lockouts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_lockouts
    ADD CONSTRAINT mfa_lockouts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mfa_org_overrides mfa_org_overrides_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_org_overrides
    ADD CONSTRAINT mfa_org_overrides_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_policies mfa_policies_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policies
    ADD CONSTRAINT mfa_policies_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_policy_history mfa_policy_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policy_history
    ADD CONSTRAINT mfa_policy_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: mfa_policy_history mfa_policy_history_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_policy_history
    ADD CONSTRAINT mfa_policy_history_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: mfa_users mfa_users_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_users
    ADD CONSTRAINT mfa_users_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: mfa_users mfa_users_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.mfa_users
    ADD CONSTRAINT mfa_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notification_events notification_events_notification_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_events
    ADD CONSTRAINT notification_events_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES public.notification_logs(id) ON DELETE CASCADE;


--
-- Name: notification_preference_audit notification_preference_audit_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.notification_preference_audit
    ADD CONSTRAINT notification_preference_audit_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: org_support_metrics org_support_metrics_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.org_support_metrics
    ADD CONSTRAINT org_support_metrics_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: prediction_results prediction_results_model_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.prediction_results
    ADD CONSTRAINT prediction_results_model_id_fkey FOREIGN KEY (model_id) REFERENCES public.predictive_models(id) ON DELETE CASCADE;


--
-- Name: restore_jobs restore_jobs_backup_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.restore_jobs
    ADD CONSTRAINT restore_jobs_backup_id_fkey FOREIGN KEY (backup_id) REFERENCES public.backups(id);


--
-- Name: review_items review_items_review_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.review_items
    ADD CONSTRAINT review_items_review_id_fkey FOREIGN KEY (review_id) REFERENCES public.access_reviews(id) ON DELETE CASCADE;


--
-- Name: review_items review_items_reviewer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.review_items
    ADD CONSTRAINT review_items_reviewer_id_fkey FOREIGN KEY (reviewer_id) REFERENCES public.users(id);


--
-- Name: review_items review_items_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.review_items
    ADD CONSTRAINT review_items_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: review_items review_items_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.review_items
    ADD CONSTRAINT review_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: roles roles_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: roles roles_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: route_optimization_jobs route_optimization_jobs_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_jobs
    ADD CONSTRAINT route_optimization_jobs_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.route_optimization_requests(id) ON DELETE CASCADE;


--
-- Name: route_optimization_requests route_optimization_requests_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_requests
    ADD CONSTRAINT route_optimization_requests_organization_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: route_optimization_results route_optimization_results_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.route_optimization_results
    ADD CONSTRAINT route_optimization_results_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.route_optimization_requests(id) ON DELETE CASCADE;


--
-- Name: security_events security_events_actor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.security_events
    ADD CONSTRAINT security_events_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: security_events security_events_on_behalf_of_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.security_events
    ADD CONSTRAINT security_events_on_behalf_of_fkey FOREIGN KEY (on_behalf_of) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: security_events security_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.security_events
    ADD CONSTRAINT security_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sla_metrics sla_metrics_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.sla_metrics
    ADD CONSTRAINT sla_metrics_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: support_audit_log support_audit_log_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_audit_log
    ADD CONSTRAINT support_audit_log_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.support_tickets(id);


--
-- Name: support_audit_log support_audit_log_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_audit_log
    ADD CONSTRAINT support_audit_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: support_tickets support_tickets_assigned_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_assigned_to_fkey FOREIGN KEY (assigned_to) REFERENCES public.users(id);


--
-- Name: support_tickets support_tickets_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: support_tickets support_tickets_reporter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES public.users(id);


--
-- Name: telemetry_events telemetry_events_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events
    ADD CONSTRAINT telemetry_events_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: telemetry_events telemetry_events_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events
    ADD CONSTRAINT telemetry_events_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id) ON DELETE CASCADE;


--
-- Name: telemetry telemetry_trip_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.telemetry
    ADD CONSTRAINT telemetry_trip_id_fkey FOREIGN KEY (trip_id) REFERENCES public.trips(id) ON DELETE CASCADE;


--
-- Name: threat_alerts threat_alerts_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.threat_alerts
    ADD CONSTRAINT threat_alerts_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: threat_alerts threat_alerts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.threat_alerts
    ADD CONSTRAINT threat_alerts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: trips trips_driver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_driver_id_fkey FOREIGN KEY (driver_id) REFERENCES public.drivers(id);


--
-- Name: trips trips_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: trips trips_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: user_links user_links_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_links
    ADD CONSTRAINT user_links_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_links user_links_linked_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_links
    ADD CONSTRAINT user_links_linked_user_id_fkey FOREIGN KEY (linked_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_links user_links_parent_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_links
    ADD CONSTRAINT user_links_parent_user_id_fkey FOREIGN KEY (parent_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_notification_preferences user_notification_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_assigned_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES public.users(id);


--
-- Name: user_roles user_roles_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: u0_a412
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: vehicle_inspections vehicle_inspections_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_inspections
    ADD CONSTRAINT vehicle_inspections_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: vehicle_inspections vehicle_inspections_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_inspections
    ADD CONSTRAINT vehicle_inspections_template_id_fkey FOREIGN KEY (template_id) REFERENCES public.inspection_templates(id);


--
-- Name: vehicle_inspections vehicle_inspections_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_inspections
    ADD CONSTRAINT vehicle_inspections_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: vehicle_verifications vehicle_verifications_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_verifications
    ADD CONSTRAINT vehicle_verifications_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: vehicle_verifications vehicle_verifications_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicle_verifications
    ADD CONSTRAINT vehicle_verifications_vehicle_id_fkey FOREIGN KEY (vehicle_id) REFERENCES public.vehicles(id);


--
-- Name: vehicles vehicles_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT vehicles_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: vehicles vehicles_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT vehicles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: vehicles vehicles_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.vehicles
    ADD CONSTRAINT vehicles_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: webhooks webhooks_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id);


--
-- Name: work_orders work_orders_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: fleet_user
--

ALTER TABLE ONLY public.work_orders
    ADD CONSTRAINT work_orders_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: access_reviews; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.access_reviews ENABLE ROW LEVEL SECURITY;

--
-- Name: access_reviews access_reviews_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY access_reviews_tenant_policy ON public.access_reviews USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: alert_rules; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.alert_rules ENABLE ROW LEVEL SECURITY;

--
-- Name: alert_rules alert_rules_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY alert_rules_tenant_policy ON public.alert_rules USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: alerts; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.alerts ENABLE ROW LEVEL SECURITY;

--
-- Name: alerts alerts_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY alerts_tenant_policy ON public.alerts USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: analytics_jobs; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.analytics_jobs ENABLE ROW LEVEL SECURITY;

--
-- Name: analytics_jobs analytics_jobs_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY analytics_jobs_tenant_policy ON public.analytics_jobs USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: api_clients; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.api_clients ENABLE ROW LEVEL SECURITY;

--
-- Name: api_clients api_clients_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY api_clients_tenant_policy ON public.api_clients USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: audit_logs; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.audit_logs ENABLE ROW LEVEL SECURITY;

--
-- Name: audit_logs audit_logs_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY audit_logs_tenant_policy ON public.audit_logs USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: audit_policies; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.audit_policies ENABLE ROW LEVEL SECURITY;

--
-- Name: audit_policies audit_policies_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY audit_policies_tenant_policy ON public.audit_policies USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: audit_settings; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.audit_settings ENABLE ROW LEVEL SECURITY;

--
-- Name: audit_settings audit_settings_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY audit_settings_tenant_policy ON public.audit_settings USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: billing_accounting_exports; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_accounting_exports ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_accounting_exports billing_accounting_exports_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY billing_accounting_exports_tenant_policy ON public.billing_accounting_exports USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: billing_customer_balances; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_customer_balances ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_customer_balances billing_customer_balances_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY billing_customer_balances_tenant_policy ON public.billing_customer_balances USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: billing_hardware_products; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_hardware_products ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_invoice_documents; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_invoice_documents ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_invoice_sequences; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_invoice_sequences ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_invoice_sequences billing_invoice_sequences_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY billing_invoice_sequences_tenant_policy ON public.billing_invoice_sequences USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: billing_invoice_templates; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_invoice_templates ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_revenue_schedules; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_revenue_schedules ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_revenue_schedules billing_revenue_schedules_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY billing_revenue_schedules_tenant_policy ON public.billing_revenue_schedules USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: billing_settings; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_settings ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_settings billing_settings_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY billing_settings_tenant_policy ON public.billing_settings USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: billing_subscriptions; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.billing_subscriptions ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_subscriptions billing_subscriptions_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY billing_subscriptions_tenant_policy ON public.billing_subscriptions USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: bulk_config; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.bulk_config ENABLE ROW LEVEL SECURITY;

--
-- Name: bulk_config bulk_config_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY bulk_config_tenant_policy ON public.bulk_config USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: bulk_jobs; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.bulk_jobs ENABLE ROW LEVEL SECURITY;

--
-- Name: bulk_jobs bulk_jobs_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY bulk_jobs_tenant_policy ON public.bulk_jobs USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: consumption_events; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.consumption_events ENABLE ROW LEVEL SECURITY;

--
-- Name: consumption_events consumption_events_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY consumption_events_tenant_policy ON public.consumption_events USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: documents; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.documents ENABLE ROW LEVEL SECURITY;

--
-- Name: documents documents_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY documents_tenant_policy ON public.documents USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: driver_attendance; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.driver_attendance ENABLE ROW LEVEL SECURITY;

--
-- Name: driver_attendance driver_attendance_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY driver_attendance_tenant_policy ON public.driver_attendance USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: driver_behavior; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.driver_behavior ENABLE ROW LEVEL SECURITY;

--
-- Name: driver_behavior_analytics; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.driver_behavior_analytics ENABLE ROW LEVEL SECURITY;

--
-- Name: driver_behavior_analytics driver_behavior_analytics_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY driver_behavior_analytics_tenant_policy ON public.driver_behavior_analytics USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: driver_behavior driver_behavior_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY driver_behavior_tenant_policy ON public.driver_behavior USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: driver_documents; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.driver_documents ENABLE ROW LEVEL SECURITY;

--
-- Name: driver_documents driver_documents_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY driver_documents_tenant_policy ON public.driver_documents USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: driver_incidents; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.driver_incidents ENABLE ROW LEVEL SECURITY;

--
-- Name: driver_incidents driver_incidents_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY driver_incidents_tenant_policy ON public.driver_incidents USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: driver_schedules; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.driver_schedules ENABLE ROW LEVEL SECURITY;

--
-- Name: driver_schedules driver_schedules_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY driver_schedules_tenant_policy ON public.driver_schedules USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: drivers; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.drivers ENABLE ROW LEVEL SECURITY;

--
-- Name: drivers drivers_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY drivers_tenant_policy ON public.drivers USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: evidence_exports; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.evidence_exports ENABLE ROW LEVEL SECURITY;

--
-- Name: evidence_exports evidence_exports_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY evidence_exports_tenant_policy ON public.evidence_exports USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: firmware_rollout_events; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.firmware_rollout_events ENABLE ROW LEVEL SECURITY;

--
-- Name: firmware_rollout_events firmware_rollout_events_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY firmware_rollout_events_tenant_policy ON public.firmware_rollout_events USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: firmware_rollouts; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.firmware_rollouts ENABLE ROW LEVEL SECURITY;

--
-- Name: firmware_rollouts firmware_rollouts_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY firmware_rollouts_tenant_policy ON public.firmware_rollouts USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: fleet_analytics; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.fleet_analytics ENABLE ROW LEVEL SECURITY;

--
-- Name: fleet_analytics fleet_analytics_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY fleet_analytics_tenant_policy ON public.fleet_analytics USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: geofences; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.geofences ENABLE ROW LEVEL SECURITY;

--
-- Name: geofences geofences_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY geofences_tenant_policy ON public.geofences USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: gps_locations; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.gps_locations ENABLE ROW LEVEL SECURITY;

--
-- Name: billing_hardware_products hardware_products_read_all; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY hardware_products_read_all ON public.billing_hardware_products FOR SELECT USING (true);


--
-- Name: inspection_templates; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.inspection_templates ENABLE ROW LEVEL SECURITY;

--
-- Name: inspection_templates inspection_templates_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY inspection_templates_tenant_policy ON public.inspection_templates USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: integration_logs; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.integration_logs ENABLE ROW LEVEL SECURITY;

--
-- Name: integration_logs integration_logs_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY integration_logs_tenant_policy ON public.integration_logs USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: integrations; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.integrations ENABLE ROW LEVEL SECURITY;

--
-- Name: integrations integrations_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY integrations_tenant_policy ON public.integrations USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: invites; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.invites ENABLE ROW LEVEL SECURITY;

--
-- Name: invites invites_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY invites_tenant_policy ON public.invites USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: invoice_items; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.invoice_items ENABLE ROW LEVEL SECURITY;

--
-- Name: invoice_items invoice_items_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY invoice_items_tenant_policy ON public.invoice_items USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: invoices; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.invoices ENABLE ROW LEVEL SECURITY;

--
-- Name: invoices invoices_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY invoices_tenant_policy ON public.invoices USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: ip_rules; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.ip_rules ENABLE ROW LEVEL SECURITY;

--
-- Name: ip_rules ip_rules_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY ip_rules_tenant_policy ON public.ip_rules USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: login_events; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.login_events ENABLE ROW LEVEL SECURITY;

--
-- Name: login_events login_events_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY login_events_tenant_policy ON public.login_events USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: maintenance_vendors; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.maintenance_vendors ENABLE ROW LEVEL SECURITY;

--
-- Name: maintenance_vendors maintenance_vendors_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY maintenance_vendors_tenant_policy ON public.maintenance_vendors USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_events; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_events ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_events mfa_events_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_events_tenant_policy ON public.mfa_events USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_exceptions; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_exceptions ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_exceptions mfa_exceptions_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_exceptions_tenant_policy ON public.mfa_exceptions USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_failures; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_failures ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_failures mfa_failures_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_failures_tenant_policy ON public.mfa_failures USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_lockouts; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_lockouts ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_lockouts mfa_lockouts_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_lockouts_tenant_policy ON public.mfa_lockouts USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_org_overrides; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_org_overrides ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_org_overrides mfa_org_overrides_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_org_overrides_tenant_policy ON public.mfa_org_overrides USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_policies; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_policies ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_policies mfa_policies_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_policies_tenant_policy ON public.mfa_policies USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_policy_history; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_policy_history ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_policy_history mfa_policy_history_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_policy_history_tenant_policy ON public.mfa_policy_history USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: mfa_users; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.mfa_users ENABLE ROW LEVEL SECURITY;

--
-- Name: mfa_users mfa_users_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY mfa_users_tenant_policy ON public.mfa_users USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: obd_readings; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.obd_readings ENABLE ROW LEVEL SECURITY;

--
-- Name: org_support_metrics; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.org_support_metrics ENABLE ROW LEVEL SECURITY;

--
-- Name: org_support_metrics org_support_metrics_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY org_support_metrics_tenant_policy ON public.org_support_metrics USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: organizations; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.organizations ENABLE ROW LEVEL SECURITY;

--
-- Name: prediction_results; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.prediction_results ENABLE ROW LEVEL SECURITY;

--
-- Name: prediction_results prediction_results_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY prediction_results_tenant_policy ON public.prediction_results USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: predictive_maintenance_analytics; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.predictive_maintenance_analytics ENABLE ROW LEVEL SECURITY;

--
-- Name: predictive_maintenance_analytics predictive_maintenance_analytics_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY predictive_maintenance_analytics_tenant_policy ON public.predictive_maintenance_analytics USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: route_optimization_requests; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.route_optimization_requests ENABLE ROW LEVEL SECURITY;

--
-- Name: route_optimization_requests route_optimization_requests_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY route_optimization_requests_tenant_policy ON public.route_optimization_requests USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: sla_metrics; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.sla_metrics ENABLE ROW LEVEL SECURITY;

--
-- Name: sla_metrics sla_metrics_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY sla_metrics_tenant_policy ON public.sla_metrics USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: support_tickets; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.support_tickets ENABLE ROW LEVEL SECURITY;

--
-- Name: support_tickets support_tickets_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY support_tickets_tenant_policy ON public.support_tickets USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: telemetry; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry ENABLE ROW LEVEL SECURITY;

--
-- Name: telemetry_events; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events ENABLE ROW LEVEL SECURITY;

--
-- Name: telemetry_events_p2026_04; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events_p2026_04 ENABLE ROW LEVEL SECURITY;

--
-- Name: telemetry_events_p2026_04 telemetry_events_p2026_04_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY telemetry_events_p2026_04_tenant_policy ON public.telemetry_events_p2026_04 USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: telemetry_events_p2026_05; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events_p2026_05 ENABLE ROW LEVEL SECURITY;

--
-- Name: telemetry_events_p2026_05 telemetry_events_p2026_05_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY telemetry_events_p2026_05_tenant_policy ON public.telemetry_events_p2026_05 USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: telemetry_events_p2026_06; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events_p2026_06 ENABLE ROW LEVEL SECURITY;

--
-- Name: telemetry_events_p2026_06 telemetry_events_p2026_06_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY telemetry_events_p2026_06_tenant_policy ON public.telemetry_events_p2026_06 USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: telemetry_events_p2026_07; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.telemetry_events_p2026_07 ENABLE ROW LEVEL SECURITY;

--
-- Name: telemetry_events_p2026_07 telemetry_events_p2026_07_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY telemetry_events_p2026_07_tenant_policy ON public.telemetry_events_p2026_07 USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: telemetry_events telemetry_events_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY telemetry_events_tenant_policy ON public.telemetry_events USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: telemetry_events tenant_isolation_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY tenant_isolation_policy ON public.telemetry_events USING ((org_id = public.get_current_org_id()));


--
-- Name: threat_alerts; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.threat_alerts ENABLE ROW LEVEL SECURITY;

--
-- Name: threat_alerts threat_alerts_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY threat_alerts_tenant_policy ON public.threat_alerts USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: trip_settings; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.trip_settings ENABLE ROW LEVEL SECURITY;

--
-- Name: trip_settings trip_settings_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY trip_settings_tenant_policy ON public.trip_settings USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: trips; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.trips ENABLE ROW LEVEL SECURITY;

--
-- Name: trips trips_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY trips_tenant_policy ON public.trips USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: user_organizations; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.user_organizations ENABLE ROW LEVEL SECURITY;

--
-- Name: user_organizations user_organizations_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY user_organizations_tenant_policy ON public.user_organizations USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

--
-- Name: vehicle_inspections; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.vehicle_inspections ENABLE ROW LEVEL SECURITY;

--
-- Name: vehicle_inspections vehicle_inspections_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY vehicle_inspections_tenant_policy ON public.vehicle_inspections USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: vehicles; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.vehicles ENABLE ROW LEVEL SECURITY;

--
-- Name: vehicles vehicles_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY vehicles_tenant_policy ON public.vehicles USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: video_recordings; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.video_recordings ENABLE ROW LEVEL SECURITY;

--
-- Name: video_recordings video_recordings_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY video_recordings_tenant_policy ON public.video_recordings USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: video_retention_policies; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.video_retention_policies ENABLE ROW LEVEL SECURITY;

--
-- Name: video_retention_policies video_retention_policies_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY video_retention_policies_tenant_policy ON public.video_retention_policies USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: video_streams; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.video_streams ENABLE ROW LEVEL SECURITY;

--
-- Name: video_streams video_streams_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY video_streams_tenant_policy ON public.video_streams USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: webhooks; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.webhooks ENABLE ROW LEVEL SECURITY;

--
-- Name: webhooks webhooks_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY webhooks_tenant_policy ON public.webhooks USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: work_orders; Type: ROW SECURITY; Schema: public; Owner: fleet_user
--

ALTER TABLE public.work_orders ENABLE ROW LEVEL SECURITY;

--
-- Name: work_orders work_orders_tenant_policy; Type: POLICY; Schema: public; Owner: fleet_user
--

CREATE POLICY work_orders_tenant_policy ON public.work_orders USING ((org_id = public.get_current_org_id())) WITH CHECK ((org_id = public.get_current_org_id()));


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT ALL ON SCHEMA public TO fleet_user;


--
-- Name: TABLE permissions; Type: ACL; Schema: public; Owner: u0_a412
--

GRANT ALL ON TABLE public.permissions TO fleet_user;


--
-- Name: TABLE role_permissions; Type: ACL; Schema: public; Owner: u0_a412
--

GRANT ALL ON TABLE public.role_permissions TO fleet_user;


--
-- Name: TABLE user_roles; Type: ACL; Schema: public; Owner: u0_a412
--

GRANT ALL ON TABLE public.user_roles TO fleet_user;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: public; Owner: u0_a412
--

ALTER DEFAULT PRIVILEGES FOR ROLE u0_a412 IN SCHEMA public GRANT ALL ON SEQUENCES TO fleet_user;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: u0_a412
--

ALTER DEFAULT PRIVILEGES FOR ROLE u0_a412 IN SCHEMA public GRANT ALL ON TABLES TO fleet_user;


--
-- PostgreSQL database dump complete
--

\unrestrict o0YlzcLDfgh03pZOAHhtgxuxTz3zJnggoOCgnQ25Y1mGT9KBfKQdaKzExtW48Ee

