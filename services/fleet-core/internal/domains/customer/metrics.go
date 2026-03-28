package customer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Customer metrics
	CustomersTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customers_total",
		Help: "Total number of customer organizations",
	}, []string{"tier", "region"})

	CustomersActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customers_active_total",
		Help: "Active customers with at least one trip in last 30 days",
	}, []string{"tier", "region"})

	CustomersTrial = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customers_trial_total",
		Help: "Customers on trial period",
	}, []string{"tier"})

	CustomerOnboardingStage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customer_onboarding_stage_total",
		Help: "Count of customers at each onboarding stage",
	}, []string{"stage"})

	// Support metrics
	SupportTicketsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "support_tickets_created_total",
		Help: "Total support tickets created",
	}, []string{"org_id", "priority", "category"})

	SupportTicketsResolvedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "support_tickets_resolved_total",
		Help: "Total support tickets resolved",
	}, []string{"org_id", "resolution_type"})

	SupportTicketResolutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "support_ticket_resolution_time_seconds",
		Help:    "Time from ticket creation to resolution",
		Buckets: prometheus.ExponentialBuckets(300, 2, 8), // 5min to ~21hrs
	}, []string{"org_id", "priority"})

	SupportCSATScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "support_csat_score",
		Help: "Customer satisfaction score (1-5)",
	}, []string{"org_id", "ticket_id"})

	SupportFirstResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "support_first_response_time_seconds",
		Help:    "Time from ticket creation to first responder assignment",
		Buckets: []float64{60, 300, 600, 1800, 3600, 7200},
	}, []string{"org_id", "priority"})

	// NPS metrics
	NPSResponsesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "customer_nps_responses_total",
		Help: "Total NPS survey responses",
	})

	NPSScore = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "customer_nps_score",
		Help: "Net Promoter Score (calculated -100 to 100)",
	})

	// Churn metrics
	ChurnRiskScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customer_churn_risk_score",
		Help: "Predicted churn risk score (0-100)",
	}, []string{"org_id"})

	CustomersChurnedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customers_churned_total",
		Help: "Total customers that churned",
	}, []string{"org_id", "churn_reason", "tier"})

	// Engagement metrics
	CustomerLoginFrequency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "customer_login_frequency_days",
		Help:    "Days between customer logins",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"org_id", "user_role"})

	FeatureAdoptionRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customer_feature_adoption_rate_percentage",
		Help: "Percentage of active customers using a specific feature",
	}, []string{"org_id", "feature_name"})

	// Usage metrics
	APICallsPerCustomer = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "customer_api_calls_total",
		Help:    "API calls per customer per day",
		Buckets: prometheus.ExponentialBuckets(10, 2, 12),
	}, []string{"org_id"})

	ConcurrentUsersPerCustomer = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customer_concurrent_users",
		Help: "Concurrent active users per customer",
	}, []string{"org_id"})

	// Billing interaction
	InvoiceViewsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_invoice_views_total",
		Help: "Total invoice views by customers",
	}, []string{"org_id"})

	PaymentRetriesByCustomer = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_payment_retries_total",
		Help: "Total payment retry attempts by customers",
	}, []string{"org_id", "reason"})

	// Communication metrics
	EmailsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_emails_sent_total",
		Help: "Total emails sent to customers",
	}, []string{"org_id", "email_type"})

	NotificationsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_notifications_sent_total",
		Help: "Total in-app notifications sent",
	}, []string{"org_id", "notification_type"})

	EmailOpenRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customer_email_open_rate_percentage",
		Help: "Email open rate percentage",
	}, []string{"org_id", "email_type"})

	// Contract metrics
	ContractRenewalsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_contract_renewals_total",
		Help: "Total contract renewals",
	}, []string{"org_id", "renewal_term"})

	ContractUpgradesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_contract_upgrades_total",
		Help: "Total contract upgrades (tier increases)",
	}, []string{"org_id", "from_tier", "to_tier"})

	ContractDowngradesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_contract_downgrades_total",
		Help: "Total contract downgrades",
	}, []string{"org_id", "from_tier", "to_tier"})

	// Risk metrics
	CustomerRiskScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "customer_risk_score",
		Help: "Customer financial/operational risk score (0-100)",
	}, []string{"org_id"})

	PaymentDelaysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "customer_payment_delays_total",
		Help: "Total delayed payments (>30 days)",
	}, []string{"org_id"})
)

// Global metrics instance
var GlobalMetrics = struct {
	CustomersTotal              *prometheus.GaugeVec
	CustomersActive             *prometheus.GaugeVec
	CustomersTrial              *prometheus.GaugeVec
	CustomerOnboardingStage     *prometheus.GaugeVec
	SupportTicketsCreatedTotal  *prometheus.CounterVec
	SupportTicketsResolvedTotal *prometheus.CounterVec
	SupportTicketResolutionTime *prometheus.HistogramVec
	SupportCSATScore            *prometheus.GaugeVec
	SupportFirstResponseTime    *prometheus.HistogramVec
	NPSResponsesTotal           prometheus.Counter
	NPSScore                    prometheus.Gauge
	ChurnRiskScore              *prometheus.GaugeVec
	CustomersChurnedTotal       *prometheus.CounterVec
	CustomerLoginFrequency      *prometheus.HistogramVec
	FeatureAdoptionRate         *prometheus.GaugeVec
	APICallsPerCustomer         *prometheus.HistogramVec
	ConcurrentUsersPerCustomer  *prometheus.GaugeVec
	InvoiceViewsTotal           *prometheus.CounterVec
	PaymentRetriesByCustomer    *prometheus.CounterVec
	EmailsSentTotal             *prometheus.CounterVec
	NotificationsSentTotal      *prometheus.CounterVec
	EmailOpenRate               *prometheus.GaugeVec
	ContractRenewalsTotal       *prometheus.CounterVec
	ContractUpgradesTotal       *prometheus.CounterVec
	ContractDowngradesTotal     *prometheus.CounterVec
	CustomerRiskScore           *prometheus.GaugeVec
	PaymentDelaysTotal          *prometheus.CounterVec
}{
	CustomersTotal:              CustomersTotal,
	CustomersActive:             CustomersActive,
	CustomersTrial:              CustomersTrial,
	CustomerOnboardingStage:     CustomerOnboardingStage,
	SupportTicketsCreatedTotal:  SupportTicketsCreatedTotal,
	SupportTicketsResolvedTotal: SupportTicketsResolvedTotal,
	SupportTicketResolutionTime: SupportTicketResolutionTime,
	SupportCSATScore:            SupportCSATScore,
	SupportFirstResponseTime:    SupportFirstResponseTime,
	NPSResponsesTotal:           NPSResponsesTotal,
	NPSScore:                    NPSScore,
	ChurnRiskScore:              ChurnRiskScore,
	CustomersChurnedTotal:       CustomersChurnedTotal,
	CustomerLoginFrequency:      CustomerLoginFrequency,
	FeatureAdoptionRate:         FeatureAdoptionRate,
	APICallsPerCustomer:         APICallsPerCustomer,
	ConcurrentUsersPerCustomer:  ConcurrentUsersPerCustomer,
	InvoiceViewsTotal:           InvoiceViewsTotal,
	PaymentRetriesByCustomer:    PaymentRetriesByCustomer,
	EmailsSentTotal:             EmailsSentTotal,
	NotificationsSentTotal:      NotificationsSentTotal,
	EmailOpenRate:               EmailOpenRate,
	ContractRenewalsTotal:       ContractRenewalsTotal,
	ContractUpgradesTotal:       ContractUpgradesTotal,
	ContractDowngradesTotal:     ContractDowngradesTotal,
	CustomerRiskScore:           CustomerRiskScore,
	PaymentDelaysTotal:          PaymentDelaysTotal,
}
