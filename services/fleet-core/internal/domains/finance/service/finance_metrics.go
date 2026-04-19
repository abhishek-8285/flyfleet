package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// FinanceMetrics holds all Prometheus metrics for the finance and fuel modules.
type FinanceMetrics struct {
	FuelCostTotal         *prometheus.CounterVec
	ExpenseTotal          *prometheus.CounterVec
	TripProfitMargin      *prometheus.GaugeVec
	AnomaliesDetected     *prometheus.CounterVec
	FraudIncidentsTotal   *prometheus.CounterVec
	TDSDeductedTotal      *prometheus.CounterVec
	VendorPaymentsTotal   *prometheus.CounterVec
	EInvoicesGenerated    prometheus.Counter
	ReconciliationMatched prometheus.Counter
	BudgetUtilization     *prometheus.GaugeVec
}

func NewFinanceMetrics() *FinanceMetrics {
	return &FinanceMetrics{
		FuelCostTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_fuel_cost_total",
			Help: "Total fuel cost ingested (INR)",
		}, []string{"tenant_id", "vehicle_type"}),

		ExpenseTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_expense_total",
			Help: "Total expense amount submitted (INR)",
		}, []string{"tenant_id", "category"}),

		TripProfitMargin: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "flyfleet_trip_profit_margin",
			Help: "Current trip profit margin percentage",
		}, []string{"tenant_id"}),

		AnomaliesDetected: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_anomalies_detected",
			Help: "Total fuel/finance anomalies detected",
		}, []string{"tenant_id", "anomaly_type"}),

		FraudIncidentsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_fraud_incidents_total",
			Help: "Total fraud incidents flagged",
		}, []string{"tenant_id", "risk_level"}),

		TDSDeductedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_tds_deducted_total",
			Help: "Total TDS deducted from vendor payments (INR)",
		}, []string{"tenant_id", "tds_section"}),

		VendorPaymentsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_vendor_payments_total",
			Help: "Total vendor payments processed (INR)",
		}, []string{"tenant_id", "payment_mode"}),

		EInvoicesGenerated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "flyfleet_einvoices_generated_total",
			Help: "Total GST e-invoices generated with IRN",
		}),

		ReconciliationMatched: promauto.NewCounter(prometheus.CounterOpts{
			Name: "flyfleet_reconciliation_matched_total",
			Help: "Total fuel card transactions successfully reconciled",
		}),

		BudgetUtilization: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "flyfleet_budget_utilization_pct",
			Help: "Budget utilization percentage per org/category",
		}, []string{"tenant_id", "category"}),
	}
}
