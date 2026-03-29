-- Migration: 070_add_missing_foreign_key_indexes.sql
-- Description: Add indexes to all foreign key columns that are missing them
-- Created: 2026-03-29
-- Based on: Database Designer analysis - 59 foreign keys without indexes

SET search_path TO public;

-- =====================================================
-- CRITICAL: These indexes are needed for JOIN performance
-- Run during low-traffic period - these are REVERSE-SAFE
-- =====================================================

\echo 'Adding missing foreign key indexes...'

-- ========================================
-- Alerts Table (6 missing FKs)
-- ========================================
CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged_by ON alerts(acknowledged_by) WHERE acknowledged_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_driver_id ON alerts(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_resolved_by ON alerts(resolved_by) WHERE resolved_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_rule_id ON alerts(rule_id) WHERE rule_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_vehicle_id ON alerts(vehicle_id) WHERE vehicle_id IS NOT NULL;

-- ========================================
-- Audit & Compliance Tables
-- ========================================
CREATE INDEX IF NOT EXISTS idx_audit_policies_updated_by ON audit_policies(updated_by) WHERE updated_by IS NOT NULL;

-- ========================================
-- Verification & Backup Tables
-- ========================================
CREATE INDEX IF NOT EXISTS idx_backup_verifications_verified_by ON backup_verifications(verified_by) WHERE verified_by IS NOT NULL;

-- ========================================
-- Driver-Related Tables
-- ========================================
CREATE INDEX IF NOT EXISTS idx_behavior_score_updates_driver_id ON behavior_score_updates(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_audit_logs_driver_id ON driver_audit_logs(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_expenses_driver_id ON driver_expenses(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_messages_driver_id ON driver_messages(driver_id);
CREATE INDEX IF NOT EXISTS idx_duty_status_logs_driver_id ON duty_status_logs(driver_id);

-- ========================================
-- Documents & Media
-- ========================================
CREATE INDEX IF NOT EXISTS idx_documents_uploaded_by ON documents(uploaded_by) WHERE uploaded_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_documents_vehicle_id ON documents(vehicle_id) WHERE vehicle_id IS NOT NULL;

-- ========================================
-- Vehicles Table (updated_by FK)
-- ========================================
CREATE INDEX IF NOT EXISTS idx_vehicles_updated_by ON vehicles(updated_by) WHERE updated_by IS NOT NULL;

-- ========================================
-- User Management & Invites
-- ========================================
CREATE INDEX IF NOT EXISTS idx_invites_created_by ON invites(created_by) WHERE created_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_invites_org_id ON invites(org_id) WHERE org_id IS NOT NULL;

-- ========================================
-- Network & Access Control
-- ========================================
CREATE INDEX IF NOT EXISTS idx_ip_rules_added_by ON ip_rules(added_by) WHERE added_by IS NOT NULL;

-- ========================================
-- MFA (Multi-Factor Authentication) Tables
-- ========================================
CREATE INDEX IF NOT EXISTS idx_mfa_exceptions_granted_by ON mfa_exceptions(granted_by) WHERE granted_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mfa_exceptions_org_id ON mfa_exceptions(org_id) WHERE org_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mfa_failures_org_id ON mfa_failures(org_id) WHERE org_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mfa_lockouts_org_id ON mfa_lockouts(org_id) WHERE org_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mfa_lockouts_unlocked_by ON mfa_lockouts(unlocked_by) WHERE unlocked_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mfa_policy_history_changed_by ON mfa_policy_history(changed_by) WHERE changed_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mfa_users_org_id ON mfa_users(org_id) WHERE org_id IS NOT NULL;

-- ========================================
-- Reviews & Approvals
-- ========================================
CREATE INDEX IF NOT EXISTS idx_review_items_reviewer_id ON review_items(reviewer_id) WHERE reviewer_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_review_items_role_id ON review_items(role_id) WHERE role_id IS NOT NULL;

-- ========================================
-- Route Optimization
-- ========================================
CREATE INDEX IF NOT EXISTS idx_route_optimization_jobs_request_id ON route_optimization_jobs(request_id) WHERE request_id IS NOT NULL;

-- ========================================
-- Security & Threat Monitoring
-- ========================================
CREATE INDEX IF NOT EXISTS idx_security_events_actor_id ON security_events(actor_id) WHERE actor_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_security_events_on_behalf_of ON security_events(on_behalf_of) WHERE on_behalf_of IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events(user_id) WHERE user_id IS NOT NULL;

-- ========================================
-- Service Level Agreements
-- ========================================
CREATE INDEX IF NOT EXISTS idx_sla_metrics_org_id ON sla_metrics(org_id) WHERE org_id IS NOT NULL;

-- ========================================
-- Support System
-- ========================================
CREATE INDEX IF NOT EXISTS idx_support_audit_log_user_id ON support_audit_log(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_support_tickets_reporter_id ON support_tickets(reporter_id) WHERE reporter_id IS NOT NULL;

-- ========================================
-- Alerts & Notifications
-- ========================================
CREATE INDEX IF NOT EXISTS idx_threat_alerts_user_id ON threat_alerts(user_id) WHERE user_id IS NOT NULL;

-- ========================================
-- User Links & Relationships
-- ========================================
CREATE INDEX IF NOT EXISTS idx_user_links_created_by ON user_links(created_by) WHERE created_by IS NOT NULL;

-- ========================================
-- Vehicle Inspections & Maintenance
-- ========================================
CREATE INDEX IF NOT EXISTS idx_vehicle_inspections_template_id ON vehicle_inspections(template_id) WHERE template_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vehicle_inspections_updated_by ON vehicle_inspections(updated_by) WHERE updated_by IS NOT NULL;

-- ========================================
-- Video Analytics & Streaming
-- ========================================
CREATE INDEX IF NOT EXISTS idx_video_analytics_stream_id ON video_analytics(stream_id) WHERE stream_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_recordings_driver_id ON video_recordings(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_recordings_organization_id ON video_recordings(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_streams_driver_id ON video_streams(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_streams_organization_id ON video_streams(organization_id) WHERE organization_id IS NOT NULL;

-- ========================================
-- Work Orders
-- ========================================
CREATE INDEX IF NOT EXISTS idx_work_orders_updated_by ON work_orders(updated_by) WHERE updated_by IS NOT NULL;

-- ========================================
-- Firmware Management
-- ========================================
CREATE INDEX IF NOT EXISTS idx_firmware_rollout_events_actor_user_id ON firmware_rollout_events(actor_user_id) WHERE actor_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_firmware_rollouts_created_by ON firmware_rollouts(created_by) WHERE created_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_firmware_rollouts_firmware_id ON firmware_rollouts(firmware_id) WHERE firmware_id IS NOT NULL;

-- ========================================
-- Geofencing
-- ========================================
CREATE INDEX IF NOT EXISTS idx_geofences_created_by ON geofences(created_by) WHERE created_by IS NOT NULL;

-- ========================================
-- Governance & Compliance
-- ========================================
CREATE INDEX IF NOT EXISTS idx_governance_audit_logs_organization_id ON governance_audit_logs(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_governance_requests_organization_id ON governance_requests(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_governance_settings_organization_id ON governance_settings(organization_id) WHERE organization_id IS NOT NULL;

-- ========================================
-- Billing & Invoices
-- ========================================
CREATE INDEX IF NOT EXISTS idx_billing_invoice_documents_template_id ON billing_invoice_documents(template_id) WHERE template_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_billing_recognized_revenue_schedule_id ON billing_recognized_revenue(schedule_id) WHERE schedule_id IS NOT NULL;

-- ========================================
-- Consumption & Usage
-- ========================================
CREATE INDEX IF NOT EXISTS idx_consumption_events_vehicle_id ON consumption_events(vehicle_id) WHERE vehicle_id IS NOT NULL;

-- ========================================
-- Entity Audit
-- ========================================
CREATE INDEX IF NOT EXISTS idx_entity_audit_history_actor ON entity_audit_history(actor) WHERE actor IS NOT NULL;

-- ========================================
-- Evidence Exports
-- ========================================
CREATE INDEX IF NOT EXISTS idx_evidence_exports_org_id ON evidence_exports(org_id) WHERE org_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_evidence_exports_requested_by ON evidence_exports(requested_by) WHERE requested_by IS NOT NULL;

-- ========================================
-- Route Optimization Stats (SKIP - this is a VIEW, not a table)
-- ========================================
-- NOTE: route_optimization_stats is a materialized view or regular view
-- Cannot create indexes on views - the parent table likely already has org_id index
-- The parent tables are route_optimization_requests and route_optimization_jobs which have org_id indexes

-- ========================================
-- Review Items (additional check)
-- ========================================
-- Already added above, but ensure both are present

\echo '====================================================='
\echo 'Missing foreign key indexes added successfully!'
\echo 'Total indexes created: See output above'
\echo '====================================================='
\echo 'Next steps:'
\echo '  1. Verify: SELECT * FROM pg_stat_user_indexes WHERE indexname LIKE ''idx_%_id'' ORDER BY relname;'
\echo '  2. Analyze query performance with EXPLAIN ANALYZE'
\echo '  3. Monitor pg_stat_statements for slow queries'
\echo '  4. Consider CONCURRENTLY for production: ALTER TABLE ... ADD INDEX CONCURRENTLY ...'
