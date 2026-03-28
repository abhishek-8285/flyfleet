package admin_support

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Admin action metrics
	AdminActionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_actions_total",
		Help: "Total admin actions performed",
	}, []string{"org_id", "action_type", "resource_type", "result"})

	AdminActionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "admin_action_duration_seconds",
		Help:    "Admin action execution time",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"org_id", "action_type"})

	AdminActionApprovalsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_action_approvals_total",
		Help: "Admin actions requiring approval",
	}, []string{"org_id", "approval_type", "approval_result"})

	AdminPrivilegeEscalation = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_privilege_escalation_total",
		Help: "Admin privilege elevation events",
	}, []string{"org_id", "escalation_reason"})

	// System override metrics
	SystemOverridesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "system_overrides_total",
		Help: "System override actions (block business rules)",
	}, []string{"org_id", "override_type", "override_reason"})

	OverrideJustificationRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "admin_override_justification_rate_percentage",
		Help: "Percentage of overrides with proper justification",
	}, []string{"org_id"})

	// Audit metrics
	AdminAuditEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_audit_events_total",
		Help: "Admin-generated audit events",
	}, []string{"org_id", "event_type", "resource_type"})

	AdminAuditLogIntegrity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "admin_audit_log_integrity_score",
		Help: "Audit log completeness and integrity (0-100)",
	}, []string{"org_id"})

	AdminActionTamperAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_action_tamper_attempts_total",
		Help: "Attempts to modify/delete admin audit logs",
	}, []string{"org_id"})

	// Bulk operations
	BulkOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_bulk_operations_total",
		Help: "Bulk administrative operations",
	}, []string{"org_id", "operation_type"})

	BulkOperationItemsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_bulk_operation_items_processed_total",
		Help: "Items processed in bulk operations",
	}, []string{"org_id", "operation_type"})

	BulkOperationFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_bulk_operation_failures_total",
		Help: "Bulk operation item failures",
	}, []string{"org_id", "operation_type", "failure_type"})

	// Data export/import
	AdminDataExportsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_data_exports_total",
		Help: "Admin-initiated data exports",
	}, []string{"org_id", "export_type", "format"})

	AdminDataImportsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_data_imports_total",
		Help: "Admin-initiated data imports",
	}, []string{"org_id", "import_type", "result"})

	// User management
	AdminUserSuspensions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_user_suspensions_total",
		Help: "User accounts suspended by admins",
	}, []string{"org_id", "suspension_reason"})

	AdminUserReactivations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_user_reactivations_total",
		Help: "Suspended users reactivated by admins",
	}, []string{"org_id", "reactivation_reason"})

	AdminPasswordResets = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_password_resets_total",
		Help: "Admin-initiated password resets",
	}, []string{"org_id", "user_type"})

	// Emergency operations
	EmergencyModeActivations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_emergency_mode_activations_total",
		Help: "Emergency mode/safety override activations",
	}, []string{"org_id", "emergency_type"})

	EmergencyModeDuration = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_emergency_mode_duration_seconds",
		Help: "Time spent in emergency mode",
	}, []string{"org_id", "emergency_type"})

	// Configuration management
	AdminConfigChangesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_config_changes_total",
		Help: "Admin configuration changes",
	}, []string{"org_id", "config_type", "change_type"})

	ConfigRollbackPerformedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_config_rollbacks_total",
		Help: "Admin-performed configuration rollbacks",
	}, []string{"org_id", "rollback_reason"})

	// Financial operations
	AdminManualCharges = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_manual_charges_total",
		Help: "Manual charges/adjustments by admins",
	}, []string{"org_id", "charge_type", "adjustment_type"})

	AdminRefundApprovals = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_refund_approvals_total",
		Help: "Refund approvals by admins",
	}, []string{"org_id", "refund_type"})

	AdminPaymentForces = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_payment_forces_total",
		Help: "Forced payment processing by admins",
	}, []string{"org_id", "payment_type"})

	// Data modification
	AdminDataCorrectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_data_corrections_total",
		Help: "Data corrections made by admins",
	}, []string{"org_id", "entity_type", "correction_type"})

	AdminDataDeletionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_data_deletions_total",
		Help: "Data deletion operations by admins",
	}, []string{"org_id", "entity_type", "deletion_type"})

	// Report generation
	AdminReportsGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_reports_generated_total",
		Help: "Reports generated by admins",
	}, []string{"org_id", "report_type"})

	// System diagnostics
	AdminDiagnosticRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_diagnostic_runs_total",
		Help: "Diagnostic/health check runs by admins",
	}, []string{"org_id", "diagnostic_type"})

	AdminLogInspections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_log_inspections_total",
		Help: "Log file inspections by admins",
	}, []string{"org_id", "log_type"})

	// Compliance monitoring
	AdminComplianceChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_compliance_checks_total",
		Help: "Manual compliance checks by admins",
	}, []string{"org_id", "check_type"})

	AdminSLAOverrides = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_sla_overrides_total",
		Help: "SLA threshold overrides by admins",
	}, []string{"org_id", "slo_type"})

	// API access
	AdminAPIKeyCreations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_api_key_creations_total",
		Help: "API keys created by admins",
	}, []string{"org_id", "key_type", "permission_level"})

	AdminImpersonations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_impersonations_total",
		Help: "Admin impersonation of regular users",
	}, []string{"org_id", "impersonated_user_type"})

	// Alert management
	AdminAlertSuppressions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_alert_suppressions_total",
		Help: "Alerts suppressed by admins",
	}, []string{"org_id", "alert_type", "suppression_duration"})

	AdminalertAcknowledgments = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_alert_acknowledgments_total",
		Help: "Alerts acknowledged by admins",
	}, []string{"org_id", "alert_severity"})
)

// Global metrics instance
var GlobalMetrics = struct {
	AdminActionsTotal            *prometheus.CounterVec
	AdminActionDuration          *prometheus.HistogramVec
	AdminActionApprovalsTotal    *prometheus.CounterVec
	AdminPrivilegeEscalation     *prometheus.CounterVec
	SystemOverridesTotal         *prometheus.CounterVec
	OverrideJustificationRate    *prometheus.GaugeVec
	AdminAuditEventsTotal        *prometheus.CounterVec
	AdminAuditLogIntegrity       *prometheus.GaugeVec
	AdminActionTamperAttempts    *prometheus.CounterVec
	BulkOperationsTotal          *prometheus.CounterVec
	BulkOperationItemsProcessed  *prometheus.CounterVec
	BulkOperationFailuresTotal   *prometheus.CounterVec
	AdminDataExportsTotal        *prometheus.CounterVec
	AdminDataImportsTotal        *prometheus.CounterVec
	AdminUserSuspensions         *prometheus.CounterVec
	AdminUserReactivations       *prometheus.CounterVec
	AdminPasswordResets          *prometheus.CounterVec
	EmergencyModeActivations     *prometheus.CounterVec
	EmergencyModeDuration        *prometheus.CounterVec
	AdminConfigChangesTotal      *prometheus.CounterVec
	ConfigRollbackPerformedTotal *prometheus.CounterVec
	AdminManualCharges           *prometheus.CounterVec
	AdminRefundApprovals         *prometheus.CounterVec
	AdminPaymentForces           *prometheus.CounterVec
	AdminDataCorrectionsTotal    *prometheus.CounterVec
	AdminDataDeletionsTotal      *prometheus.CounterVec
	AdminReportsGeneratedTotal   *prometheus.CounterVec
	AdminDiagnosticRunsTotal     *prometheus.CounterVec
	AdminLogInspections          *prometheus.CounterVec
	AdminComplianceChecksTotal   *prometheus.CounterVec
	AdminSLAOverrides            *prometheus.CounterVec
	AdminAPIKeyCreations         *prometheus.CounterVec
	AdminImpersonations          *prometheus.CounterVec
	AdminAlertSuppressions       *prometheus.CounterVec
	AdminalertAcknowledgments    *prometheus.CounterVec
}{
	AdminActionsTotal:            AdminActionsTotal,
	AdminActionDuration:          AdminActionDuration,
	AdminActionApprovalsTotal:    AdminActionApprovalsTotal,
	AdminPrivilegeEscalation:     AdminPrivilegeEscalation,
	SystemOverridesTotal:         SystemOverridesTotal,
	OverrideJustificationRate:    OverrideJustificationRate,
	AdminAuditEventsTotal:        AdminAuditEventsTotal,
	AdminAuditLogIntegrity:       AdminAuditLogIntegrity,
	AdminActionTamperAttempts:    AdminActionTamperAttempts,
	BulkOperationsTotal:          BulkOperationsTotal,
	BulkOperationItemsProcessed:  BulkOperationItemsProcessed,
	BulkOperationFailuresTotal:   BulkOperationFailuresTotal,
	AdminDataExportsTotal:        AdminDataExportsTotal,
	AdminDataImportsTotal:        AdminDataImportsTotal,
	AdminUserSuspensions:         AdminUserSuspensions,
	AdminUserReactivations:       AdminUserReactivations,
	AdminPasswordResets:          AdminPasswordResets,
	EmergencyModeActivations:     EmergencyModeActivations,
	EmergencyModeDuration:        EmergencyModeDuration,
	AdminConfigChangesTotal:      AdminConfigChangesTotal,
	ConfigRollbackPerformedTotal: ConfigRollbackPerformedTotal,
	AdminManualCharges:           AdminManualCharges,
	AdminRefundApprovals:         AdminRefundApprovals,
	AdminPaymentForces:           AdminPaymentForces,
	AdminDataCorrectionsTotal:    AdminDataCorrectionsTotal,
	AdminDataDeletionsTotal:      AdminDataDeletionsTotal,
	AdminReportsGeneratedTotal:   AdminReportsGeneratedTotal,
	AdminDiagnosticRunsTotal:     AdminDiagnosticRunsTotal,
	AdminLogInspections:          AdminLogInspections,
	AdminComplianceChecksTotal:   AdminComplianceChecksTotal,
	AdminSLAOverrides:            AdminSLAOverrides,
	AdminAPIKeyCreations:         AdminAPIKeyCreations,
	AdminImpersonations:          AdminImpersonations,
	AdminAlertSuppressions:       AdminAlertSuppressions,
	AdminalertAcknowledgments:    AdminalertAcknowledgments,
}
