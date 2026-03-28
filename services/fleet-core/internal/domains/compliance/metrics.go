package compliance

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Compliance check metrics
	ComplianceChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_checks_total",
		Help: "Total compliance checks performed",
	}, []string{"org_id", "check_type", "regulation"})

	ComplianceChecksPassedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_checks_passed_total",
		Help: "Total compliance checks that passed",
	}, []string{"org_id", "check_type", "regulation"})

	ComplianceViolationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_violations_total",
		Help: "Total compliance violations detected",
	}, []string{"org_id", "violation_type", "severity", "regulation"})

	ComplianceViolationsResolvedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_violations_resolved_total",
		Help: "Total compliance violations resolved",
	}, []string{"org_id", "resolution_type"})

	// Policy enforcement metrics
	PoliciesEnforcedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_policies_enforced_total",
		Help: "Total policy enforcement actions taken",
	}, []string{"org_id", "policy_type", "action"})

	PolicyViolationsDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_policy_violations_detected_total",
		Help: "Total policy violations detected",
	}, []string{"org_id", "policy_id", "severity"})

	// Audit metrics
	AuditEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_audit_events_total",
		Help: "Total audit log events for compliance",
	}, []string{"org_id", "event_type", "resource_type"})

	AuditTrailIntegrity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_audit_trail_integrity_score",
		Help: "Audit trail completeness and integrity score (0-100)",
	}, []string{"org_id"})

	// Data retention metrics
	DataRetentionCompliance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_data_retention_compliance_percentage",
		Help: "Percentage of data compliant with retention policies",
	}, []string{"org_id", "data_type"})

	RecordsExpiredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_records_expired_total",
		Help: "Total records expired and purged per retention policy",
	}, []string{"org_id", "data_type"})

	// GDPR / Privacy metrics
	DataSubjectRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_data_subject_requests_total",
		Help: "Total data subject requests (GDPR, etc.)",
	}, []string{"org_id", "request_type"})

	DSRCompletedWithinSLATotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_dsr_completed_within_sla_total",
		Help: "Total DSARs completed within SLA (30 days)",
	}, []string{"org_id"})

	DSROverdueTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_dsar_overdue_total",
		Help: "Total DSARs past due date",
	}, []string{"org_id"})

	// India compliance metrics (FASTag, E-Way Bill, GST)
	FASTagComplianceScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_fastag_compliance_score",
		Help: "FASTag compliance score percentage",
	}, []string{"org_id"})

	EWayBillGenerationSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_ewaybill_generation_success_rate",
		Help: "E-Way Bill generation success rate percentage",
	}, []string{"org_id"})

	GSTReturnFilingTimeliness = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_gst_return_filing_timeliness_percentage",
		Help: "GST returns filed on time percentage",
	}, []string{"org_id", "return_type"})

	// Report generation metrics
	ComplianceReportsGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_reports_generated_total",
		Help: "Total compliance reports generated",
	}, []string{"org_id", "report_type"})

	ReportsDeliveredOnTime = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_reports_delivered_on_time_total",
		Help: "Total reports delivered before deadline",
	}, []string{"org_id", "report_type"})

	// Certification metrics
	CertificationsExpiringTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_certifications_expiring_total",
		Help: "Count of certifications expiring within 90 days",
	}, []string{"org_id", "certification_type", "expiring_in_days"})

	// Security scan metrics
	SecurityScansPerformedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_security_scans_performed_total",
		Help: "Total security compliance scans performed",
	}, []string{"org_id", "scan_type"})

	VulnerabilitiesFoundTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_vulnerabilities_found_total",
		Help: "Total vulnerabilities found in compliance scans",
	}, []string{"org_id", "severity", "cwe_id"})

	VulnerabilitiesRemediatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_vulnerabilities_remediated_total",
		Help: "Total vulnerabilities remediated",
	}, []string{"org_id", "severity"})

	// Access control metrics
	AccessReviewsPerformedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_access_reviews_performed_total",
		Help: "Total user access reviews performed",
	}, []string{"org_id"})

	OrphanedAccountsFoundTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_orphaned_accounts_found_total",
		Help: "Total orphaned user accounts identified",
	}, []string{"org_id"})

	ExcessivePermissionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_excessive_permissions_total",
		Help: "Total accounts with excessive permissions identified",
	}, []string{"org_id", "permission_level"})

	// Policy versioning metrics
	PolicyUpdatesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "compliance_policy_updates_total",
		Help: "Total compliance policy updates",
	}, []string{"org_id", "policy_type"})

	PolicyAcknowledgmentRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "compliance_policy_acknowledgment_rate_percentage",
		Help: "Percentage of users who acknowledged policy updates",
	}, []string{"org_id", "policy_id"})
)

// Global metrics instance
var GlobalMetrics = struct {
	ComplianceChecksTotal             *prometheus.CounterVec
	ComplianceChecksPassedTotal       *prometheus.CounterVec
	ComplianceViolationsTotal         *prometheus.CounterVec
	ComplianceViolationsResolvedTotal *prometheus.CounterVec
	PoliciesEnforcedTotal             *prometheus.CounterVec
	PolicyViolationsDetected          *prometheus.CounterVec
	AuditEventsTotal                  *prometheus.CounterVec
	AuditTrailIntegrity               *prometheus.GaugeVec
	DataRetentionCompliance           *prometheus.GaugeVec
	RecordsExpiredTotal               *prometheus.CounterVec
	DataSubjectRequestsTotal          *prometheus.CounterVec
	DSRCompletedWithinSLATotal        *prometheus.CounterVec
	DSROverdueTotal                   *prometheus.CounterVec
	FASTagComplianceScore             *prometheus.GaugeVec
	EWayBillGenerationSuccessRate     *prometheus.GaugeVec
	GSTReturnFilingTimeliness         *prometheus.GaugeVec
	ComplianceReportsGeneratedTotal   *prometheus.CounterVec
	ReportsDeliveredOnTime            *prometheus.CounterVec
	CertificationsExpiringTotal       *prometheus.GaugeVec
	SecurityScansPerformedTotal       *prometheus.CounterVec
	VulnerabilitiesFoundTotal         *prometheus.CounterVec
	VulnerabilitiesRemediatedTotal    *prometheus.CounterVec
	AccessReviewsPerformedTotal       *prometheus.CounterVec
	OrphanedAccountsFoundTotal        *prometheus.CounterVec
	ExcessivePermissionsTotal         *prometheus.CounterVec
	PolicyUpdatesTotal                *prometheus.CounterVec
	PolicyAcknowledgmentRate          *prometheus.GaugeVec
}{
	ComplianceChecksTotal:             ComplianceChecksTotal,
	ComplianceChecksPassedTotal:       ComplianceChecksPassedTotal,
	ComplianceViolationsTotal:         ComplianceViolationsTotal,
	ComplianceViolationsResolvedTotal: ComplianceViolationsResolvedTotal,
	PoliciesEnforcedTotal:             PoliciesEnforcedTotal,
	PolicyViolationsDetected:          PolicyViolationsDetected,
	AuditEventsTotal:                  AuditEventsTotal,
	AuditTrailIntegrity:               AuditTrailIntegrity,
	DataRetentionCompliance:           DataRetentionCompliance,
	RecordsExpiredTotal:               RecordsExpiredTotal,
	DataSubjectRequestsTotal:          DataSubjectRequestsTotal,
	DSRCompletedWithinSLATotal:        DSRCompletedWithinSLATotal,
	DSROverdueTotal:                   DSROverdueTotal,
	FASTagComplianceScore:             FASTagComplianceScore,
	EWayBillGenerationSuccessRate:     EWayBillGenerationSuccessRate,
	GSTReturnFilingTimeliness:         GSTReturnFilingTimeliness,
	ComplianceReportsGeneratedTotal:   ComplianceReportsGeneratedTotal,
	ReportsDeliveredOnTime:            ReportsDeliveredOnTime,
	CertificationsExpiringTotal:       CertificationsExpiringTotal,
	SecurityScansPerformedTotal:       SecurityScansPerformedTotal,
	VulnerabilitiesFoundTotal:         VulnerabilitiesFoundTotal,
	VulnerabilitiesRemediatedTotal:    VulnerabilitiesRemediatedTotal,
	AccessReviewsPerformedTotal:       AccessReviewsPerformedTotal,
	OrphanedAccountsFoundTotal:        OrphanedAccountsFoundTotal,
	ExcessivePermissionsTotal:         ExcessivePermissionsTotal,
	PolicyUpdatesTotal:                PolicyUpdatesTotal,
	PolicyAcknowledgmentRate:          PolicyAcknowledgmentRate,
}
