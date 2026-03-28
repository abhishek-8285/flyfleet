package billing

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Invoice metrics
	InvoicesGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "billing_invoices_generated_total",
		Help: "Total number of invoices generated",
	})

	InvoicesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_invoices_sent_total",
		Help: "Total invoices sent to customers",
	}, []string{"org_id", "channel"})

	InvoicesPaidTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_invoices_paid_total",
		Help: "Total invoices marked as paid",
	}, []string{"org_id", "payment_method"})

	InvoiceAmountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_invoice_amount_total",
		Help: "Total invoice amount in INR",
	}, []string{"org_id", "currency"})

	OverdueInvoices = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "billing_overdue_invoices",
		Help: "Current count of overdue invoices",
	}, []string{"org_id", "aging_bucket"})

	// Payment metrics
	PaymentsAttemptedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_payments_attempted_total",
		Help: "Total payment attempts",
	}, []string{"org_id", "payment_method", "result"})

	PaymentsSucceededTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_payments_succeeded_total",
		Help: "Total successful payments",
	}, []string{"org_id", "payment_method"})

	PaymentAmountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_payment_amount_total",
		Help: "Total successful payment amount in INR",
	}, []string{"org_id", "payment_method"})

	PaymentFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_payment_failures_total",
		Help: "Total failed payment attempts",
	}, []string{"org_id", "gateway", "error_type"})

	PaymentRetryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_payment_retry_total",
		Help: "Total payment retry attempts",
	}, []string{"org_id", "retry_count"})

	// Collection metrics
	CollectionEfficiency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "billing_collection_efficiency_percentage",
		Help: "Collection efficiency percentage (collected / billed)",
	}, []string{"org_id", "period"})

	DSO = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "billing_dso_days",
		Help: "Days Sales Outstanding (average collection period)",
	}, []string{"org_id"})

	// Credit note metrics
	CreditNotesIssuedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_credit_notes_issued_total",
		Help: "Total credit notes issued",
	}, []string{"org_id", "reason"})

	CreditNoteAmountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_credit_note_amount_total",
		Help: "Total credit note amount in INR",
	}, []string{"org_id"})

	// Dispute metrics
	BillingDisputesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_disputes_total",
		Help: "Total billing disputes raised",
	}, []string{"org_id", "dispute_type"})

	DisputesResolvedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_disputes_resolved_total",
		Help: "Total disputes resolved",
	}, []string{"org_id", "resolution_type"})

	// Tax metrics
	TaxInvoicesGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_tax_invoices_generated_total",
		Help: "Total tax invoices generated",
	}, []string{"org_id", "tax_type"})

	TaxCollectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_tax_collected_total",
		Help: "Total tax collected in INR",
	}, []string{"org_id", "tax_type"})

	// PDF generation metrics
	PDFGenerationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "billing_pdf_generation_duration_seconds",
		Help:    "PDF invoice generation duration",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 6),
	}, []string{"org_id"})

	PDFGenerationFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_pdf_generation_failures_total",
		Help: "Total PDF generation failures",
	}, []string{"org_id", "reason"})

	// Email delivery metrics
	EmailDeliveryAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_email_delivery_attempts_total",
		Help: "Total email delivery attempts",
	}, []string{"org_id", "template_type"})

	EmailDeliverySuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_email_delivery_success_total",
		Help: "Total successful email deliveries",
	}, []string{"org_id", "template_type"})

	// Reconciliation metrics
	ReconciliationRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_reconciliation_runs_total",
		Help: "Total billing reconciliation runs",
	}, []string{"org_id", "reconciliation_type"})

	ReconciliationMismatchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_reconciliation_mismatches_total",
		Help: "Total reconciliation mismatches detected",
	}, []string{"org_id", "mismatch_type"})

	// Revenue recognition metrics
	RecognizedRevenueTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_recognized_revenue_total",
		Help: "Total recognized revenue (accrual basis)",
	}, []string{"org_id", "revenue_type", "period"})

	DeferredRevenue = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "billing_deferred_revenue_total",
		Help: "Current deferred revenue (unearned)",
	}, []string{"org_id"})
)

// Global metrics instance
var GlobalMetrics = struct {
	InvoicesGeneratedTotal        prometheus.Counter
	InvoicesSentTotal             *prometheus.CounterVec
	InvoicesPaidTotal             *prometheus.CounterVec
	InvoiceAmountTotal            *prometheus.CounterVec
	OverdueInvoices               *prometheus.GaugeVec
	PaymentsAttemptedTotal        *prometheus.CounterVec
	PaymentsSucceededTotal        *prometheus.CounterVec
	PaymentAmountTotal            *prometheus.CounterVec
	PaymentFailuresTotal          *prometheus.CounterVec
	PaymentRetryTotal             *prometheus.CounterVec
	CollectionEfficiency          *prometheus.GaugeVec
	DSO                           *prometheus.GaugeVec
	CreditNotesIssuedTotal        *prometheus.CounterVec
	CreditNoteAmountTotal         *prometheus.CounterVec
	BillingDisputesTotal          *prometheus.CounterVec
	DisputesResolvedTotal         *prometheus.CounterVec
	TaxInvoicesGeneratedTotal     *prometheus.CounterVec
	TaxCollectedTotal             *prometheus.CounterVec
	PDFGenerationDuration         *prometheus.HistogramVec
	PDFGenerationFailuresTotal    *prometheus.CounterVec
	EmailDeliveryAttemptsTotal    *prometheus.CounterVec
	EmailDeliverySuccessTotal     *prometheus.CounterVec
	ReconciliationRunsTotal       *prometheus.CounterVec
	ReconciliationMismatchesTotal *prometheus.CounterVec
	RecognizedRevenueTotal        *prometheus.CounterVec
	DeferredRevenue               *prometheus.GaugeVec
}{
	InvoicesGeneratedTotal:        InvoicesGeneratedTotal,
	InvoicesSentTotal:             InvoicesSentTotal,
	InvoicesPaidTotal:             InvoicesPaidTotal,
	InvoiceAmountTotal:            InvoiceAmountTotal,
	OverdueInvoices:               OverdueInvoices,
	PaymentsAttemptedTotal:        PaymentsAttemptedTotal,
	PaymentsSucceededTotal:        PaymentsSucceededTotal,
	PaymentAmountTotal:            PaymentAmountTotal,
	PaymentFailuresTotal:          PaymentFailuresTotal,
	PaymentRetryTotal:             PaymentRetryTotal,
	CollectionEfficiency:          CollectionEfficiency,
	DSO:                           DSO,
	CreditNotesIssuedTotal:        CreditNotesIssuedTotal,
	CreditNoteAmountTotal:         CreditNoteAmountTotal,
	BillingDisputesTotal:          BillingDisputesTotal,
	DisputesResolvedTotal:         DisputesResolvedTotal,
	TaxInvoicesGeneratedTotal:     TaxInvoicesGeneratedTotal,
	TaxCollectedTotal:             TaxCollectedTotal,
	PDFGenerationDuration:         PDFGenerationDuration,
	PDFGenerationFailuresTotal:    PDFGenerationFailuresTotal,
	EmailDeliveryAttemptsTotal:    EmailDeliveryAttemptsTotal,
	EmailDeliverySuccessTotal:     EmailDeliverySuccessTotal,
	ReconciliationRunsTotal:       ReconciliationRunsTotal,
	ReconciliationMismatchesTotal: ReconciliationMismatchesTotal,
	RecognizedRevenueTotal:        RecognizedRevenueTotal,
	DeferredRevenue:               DeferredRevenue,
}
