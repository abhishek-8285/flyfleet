package india

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// FASTag metrics
	FASTagRechargeAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_fastag_recharge_attempts_total",
		Help: "Total FASTag recharge attempts",
	}, []string{"tenant_id", "bank", "amount_bucket"})

	FASTagRechargeSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_fastag_recharge_success_total",
		Help: "Total successful FASTag recharges",
	}, []string{"tenant_id", "bank"})

	FASTagRechargeAmountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_fastag_recharge_amount_total",
		Help: "Total FASTag recharge amount in INR",
	}, []string{"tenant_id"})

	FASTagBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_fastag_balance_current",
		Help: "Current FASTag balance per vehicle",
	}, []string{"tenant_id", "vehicle_id", "fastag_id"})

	FASTagLowBalanceAlerts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_fastag_low_balance_alerts_total",
		Help: "Total low balance alerts sent",
	}, []string{"tenant_id", "threshold"})

	FASTagTollTransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_fastag_toll_transactions_total",
		Help: "Total FASTag toll transactions",
	}, []string{"tenant_id", "toll_plaza", "state"})

	FASTagTollAmountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_fastag_toll_amount_total",
		Help: "Total toll amount paid via FASTag in INR",
	}, []string{"tenant_id"})

	// E-Way Bill metrics
	EWayBillGenerationAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_ewaybill_generation_attempts_total",
		Help: "Total E-Way Bill generation attempts",
	}, []string{"tenant_id", "transport_mode"})

	EWayBillGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_ewaybills_generated_total",
		Help: "Total E-Way Bills successfully generated",
	}, []string{"tenant_id", "generation_type"})

	EWayBillCancellationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_ewaybills_cancelled_total",
		Help: "Total E-Way Bills cancelled",
	}, []string{"tenant_id", "cancellation_reason"})

	EWayBillExtensionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_ewaybill_extensions_total",
		Help: "Total E-Way Bill validity extensions",
	}, []string{"tenant_id"})

	EWayBillGenerationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "india_ewaybill_generation_duration_seconds",
		Help:    "E-Way Bill generation API latency",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 6),
	}, []string{"tenant_id"})

	EWayBillValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_ewaybill_validation_errors_total",
		Help: "Total E-Way Bill validation errors from GSTN",
	}, []string{"tenant_id", "error_code"})

	// GST metrics
	GSTInvoicesGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_gst_invoices_generated_total",
		Help: "Total GST invoices generated",
	}, []string{"tenant_id", "invoice_type", "gst_regime"})

	GSTInvoicesWithIRNTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_gst_invoices_with_irn_total",
		Help: "Total invoices with IRN (Invoice Reference Number)",
	}, []string{"tenant_id", "irn_source"})

	GSTReturnsFiledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_gst_returns_filed_total",
		Help: "Total GST returns filed",
	}, []string{"tenant_id", "return_type", "period"})

	GSTR1FiledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_gstr1_filed_total",
		Help: "Total GSTR-1 (outward supplies) filings",
	}, []string{"tenant_id", "period"})

	GSTR3BFiledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_gstr3b_filed_total",
		Help: "Total GSTR-3B (summary) filings",
	}, []string{"tenant_id", "period"})

	GSTInputTaxCredit = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_gst_input_tax_credit_available",
		Help: "Available input tax credit in INR",
	}, []string{"tenant_id"})

	GSTLiabilityAccrued = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_gst_liability_accrued",
		Help: "Accrued GST liability in INR",
	}, []string{"tenant_id", "gst_type"})

	// HSN/SAC code metrics
	HSNCodesUsed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_hsn_codes_used_total",
		Help: "Unique HSN/SAC codes used in transactions",
	}, []string{"tenant_id"})

	HSNCodeValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_hsn_validation_errors_total",
		Help: "Invalid HSN/SAC code validation errors",
	}, []string{"tenant_id", "error_type"})

	// TDS/TCS metrics
	TDSDeductedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_tds_deducted_total",
		Help: "Total TDS deducted from vendor payments in INR",
	}, []string{"tenant_id", "tds_section", "vendor_type"})

	TDSDepositedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_tds_deposited_total",
		Help: "Total TDS deposited with government in INR",
	}, []string{"tenant_id", "tds_section", "period"})

	TCSCollectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_tcs_collected_total",
		Help: "Total TCS collected from customers in INR",
	}, []string{"tenant_id", "tc_section"})

	// e-Invoice metrics
	EInvoicesGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_einvoices_generated_total",
		Help: "Total e-Invoices generated with IRN",
	}, []string{"tenant_id", "invoice_type"})

	EInvoiceIRNFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_einvoice_irn_failures_total",
		Help: "Total e-Invoice IRN generation failures",
	}, []string{"tenant_id", "failure_reason"})

	EInvoiceCancellationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_einvoice_cancellations_total",
		Help: "Total e-Invoice cancellations",
	}, []string{"tenant_id"})

	// e-Way Bill validity
	EWayBillExpiringSoon = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_ewaybill_expiring_soon_total",
		Help: "E-Way Bills expiring within 24 hours",
	}, []string{"tenant_id"})

	EWayBillExpiredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_ewaybills_expired_total",
		Help: "Total E-Way Bills that expired before completion",
	}, []string{"tenant_id"})

	// GST compliance score
	GSTComplianceScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_gst_compliance_score_percentage",
		Help: "GST compliance score (timely filing, accuracy)",
	}, []string{"tenant_id"})

	// State-wise metrics
	InterStateTransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_interstate_transactions_total",
		Help: "Total inter-state transactions",
	}, []string{"tenant_id", "from_state", "to_state"})

	StateGSTBreakdown = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_state_gst_amount",
		Help: "GST amount by state (CGST+SGST/IGST) in INR",
	}, []string{"tenant_id", "state", "gst_type"})

	// Digital signature metrics
	DigitalSignatureUsage = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "india_digital_signature_usage_total",
		Help: "Digital signature usage count",
	}, []string{"tenant_id", "signature_type"})

	DigitalSignatureExpiringTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "india_digital_signatures_expiring_total",
		Help: "Digital signatures expiring within 30 days",
	}, []string{"tenant_id", "expiring_in_days"})
)

// Global metrics instance
var GlobalMetrics = struct {
	FASTagRechargeAttemptsTotal     *prometheus.CounterVec
	FASTagRechargeSuccessTotal      *prometheus.CounterVec
	FASTagRechargeAmountTotal       *prometheus.CounterVec
	FASTagBalance                   *prometheus.GaugeVec
	FASTagLowBalanceAlerts          *prometheus.CounterVec
	FASTagTollTransactionsTotal     *prometheus.CounterVec
	FASTagTollAmountTotal           *prometheus.CounterVec
	EWayBillGenerationAttemptsTotal *prometheus.CounterVec
	EWayBillGeneratedTotal          *prometheus.CounterVec
	EWayBillCancellationTotal       *prometheus.CounterVec
	EWayBillExtensionTotal          *prometheus.CounterVec
	EWayBillGenerationDuration      *prometheus.HistogramVec
	EWayBillValidationErrors        *prometheus.CounterVec
	GSTInvoicesGeneratedTotal       *prometheus.CounterVec
	GSTInvoicesWithIRNTotal         *prometheus.CounterVec
	GSTReturnsFiledTotal            *prometheus.CounterVec
	GSTR1FiledTotal                 *prometheus.CounterVec
	GSTR3BFiledTotal                *prometheus.CounterVec
	GSTInputTaxCredit               *prometheus.GaugeVec
	GSTLiabilityAccrued             *prometheus.GaugeVec
	HSNCodesUsed                    *prometheus.GaugeVec
	HSNCodeValidationErrors         *prometheus.CounterVec
	TDSDeductedTotal                *prometheus.CounterVec
	TDSDepositedTotal               *prometheus.CounterVec
	TCSCollectedTotal               *prometheus.CounterVec
	EInvoicesGeneratedTotal         *prometheus.CounterVec
	EInvoiceIRNFailuresTotal        *prometheus.CounterVec
	EInvoiceCancellationTotal       *prometheus.CounterVec
	EWayBillExpiringSoon            *prometheus.GaugeVec
	EWayBillExpiredTotal            *prometheus.CounterVec
	GSTComplianceScore              *prometheus.GaugeVec
	InterStateTransactionsTotal     *prometheus.CounterVec
	StateGSTBreakdown               *prometheus.GaugeVec
	DigitalSignatureUsage           *prometheus.CounterVec
	DigitalSignatureExpiringTotal   *prometheus.GaugeVec
}{
	FASTagRechargeAttemptsTotal:     FASTagRechargeAttemptsTotal,
	FASTagRechargeSuccessTotal:      FASTagRechargeSuccessTotal,
	FASTagRechargeAmountTotal:       FASTagRechargeAmountTotal,
	FASTagBalance:                   FASTagBalance,
	FASTagLowBalanceAlerts:          FASTagLowBalanceAlerts,
	FASTagTollTransactionsTotal:     FASTagTollTransactionsTotal,
	FASTagTollAmountTotal:           FASTagTollAmountTotal,
	EWayBillGenerationAttemptsTotal: EWayBillGenerationAttemptsTotal,
	EWayBillGeneratedTotal:          EWayBillGeneratedTotal,
	EWayBillCancellationTotal:       EWayBillCancellationTotal,
	EWayBillExtensionTotal:          EWayBillExtensionTotal,
	EWayBillGenerationDuration:      EWayBillGenerationDuration,
	EWayBillValidationErrors:        EWayBillValidationErrors,
	GSTInvoicesGeneratedTotal:       GSTInvoicesGeneratedTotal,
	GSTInvoicesWithIRNTotal:         GSTInvoicesWithIRNTotal,
	GSTReturnsFiledTotal:            GSTReturnsFiledTotal,
	GSTR1FiledTotal:                 GSTR1FiledTotal,
	GSTR3BFiledTotal:                GSTR3BFiledTotal,
	GSTInputTaxCredit:               GSTInputTaxCredit,
	GSTLiabilityAccrued:             GSTLiabilityAccrued,
	HSNCodesUsed:                    HSNCodesUsed,
	HSNCodeValidationErrors:         HSNCodeValidationErrors,
	TDSDeductedTotal:                TDSDeductedTotal,
	TDSDepositedTotal:               TDSDepositedTotal,
	TCSCollectedTotal:               TCSCollectedTotal,
	EInvoicesGeneratedTotal:         EInvoicesGeneratedTotal,
	EInvoiceIRNFailuresTotal:        EInvoiceIRNFailuresTotal,
	EInvoiceCancellationTotal:       EInvoiceCancellationTotal,
	EWayBillExpiringSoon:            EWayBillExpiringSoon,
	EWayBillExpiredTotal:            EWayBillExpiredTotal,
	GSTComplianceScore:              GSTComplianceScore,
	InterStateTransactionsTotal:     InterStateTransactionsTotal,
	StateGSTBreakdown:               StateGSTBreakdown,
	DigitalSignatureUsage:           DigitalSignatureUsage,
	DigitalSignatureExpiringTotal:   DigitalSignatureExpiringTotal,
}
