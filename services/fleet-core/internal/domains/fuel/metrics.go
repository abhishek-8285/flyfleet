package fuel

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Fuel transaction metrics
	FuelTransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_transactions_total",
		Help: "Total fuel card transactions",
	}, []string{"tenant_id", "vehicle_id", "card_id", "station_id"})

	FuelTransactionsAmountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_transactions_amount_total",
		Help: "Total fuel transaction amount in INR",
	}, []string{"tenant_id", "currency"})

	FuelTransactionsVolumeTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_transactions_volume_total",
		Help: "Total fuel volume in liters",
	}, []string{"tenant_id", "fuel_type"})

	FuelTransactionsFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_transactions_failed_total",
		Help: "Total failed fuel transactions",
	}, []string{"tenant_id", "failure_reason"})

	FuelTransactionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "fuel_transaction_duration_seconds",
		Help:    "Fuel transaction authorization and completion time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 8),
	}, []string{"tenant_id"})

	// OTP verification metrics
	FuelOTPSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_otp_sent_total",
		Help: "Total OTPs sent for fuel transactions",
	}, []string{"tenant_id", "channel"})

	FuelOTPVerifiedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_otp_verified_total",
		Help: "Total OTPs successfully verified",
	}, []string{"tenant_id"})

	FuelOTPFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_otp_failed_total",
		Help: "Total OTP verification failures",
	}, []string{"tenant_id", "failure_reason"})

	FuelOTPExpiredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_otp_expired_total",
		Help: "Total OTPs that expired before verification",
	}, []string{"tenant_id"})

	FuelOTPBypassAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_otp_bypass_attempts_total",
		Help: "Total attempts to bypass OTP verification",
	}, []string{"tenant_id", "attempt_type"})

	// Reconciliation metrics
	FuelReconciliationRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_reconciliation_runs_total",
		Help: "Total fuel reconciliation runs",
	}, []string{"tenant_id", "reconciliation_type"})

	FuelTransactionsReconciledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_transactions_reconciled_total",
		Help: "Total fuel transactions successfully reconciled",
	}, []string{"tenant_id"})

	FuelReconciliationGapsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_reconciliation_gaps_total",
		Help: "Total missing transactions identified during reconciliation",
	}, []string{"tenant_id", "gap_type"})

	FuelReconciliationMismatchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_reconciliation_mismatches_total",
		Help: "Total transactions with mismatched amounts",
	}, []string{"tenant_id", "mismatch_type"})

	// Anomaly detection metrics
	FuelAnomaliesDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_anomalies_detected_total",
		Help: "Total fuel anomalies detected by ML/signals",
	}, []string{"tenant_id", "anomaly_type", "severity"})

	FuelAnomaliesResolvedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_anomalies_resolved_total",
		Help: "Total fuel anomalies investigated and resolved",
	}, []string{"tenant_id", "resolution_type"})

	FuelAnomalyDetectionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "fuel_anomaly_detection_latency_seconds",
		Help:    "Time from transaction to anomaly detection",
		Buckets: prometheus.ExponentialBuckets(60, 2, 6),
	}, []string{"tenant_id", "anomaly_type"})

	// Card management metrics
	FuelCardsActiveTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_cards_active_total",
		Help: "Total active fuel cards",
	}, []string{"tenant_id", "card_type"})

	FuelCardsBlockedTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_cards_blocked_total",
		Help: "Total blocked fuel cards",
	}, []string{"tenant_id", "block_reason"})

	FuelCardSpendLimitExceeded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_card_spend_limit_exceeded_total",
		Help: "Total transactions blocked due to spend limit",
	}, []string{"tenant_id", "card_id"})

	// Station metrics
	FuelStationsSupportedTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fuel_stations_supported_total",
		Help: "Total supported fuel stations",
	})

	FuelStationAvailability = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_station_availability",
		Help: "Fuel station API/network availability (0-1)",
	}, []string{"station_id"})

	FuelStationTransactionSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_station_transaction_success_rate",
		Help: "Transaction success rate per station (0-1)",
	}, []string{"station_id"})

	// Limit management
	DailySpendLimitUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_daily_spend_limit_utilization_percentage",
		Help: "Percentage of daily spend limit utilized",
	}, []string{"tenant_id", "card_id"})

	FuelCardVolumeLimitUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_card_volume_limit_utilization_percentage",
		Help: "Percentage of volume limit (liters) utilized",
	}, []string{"tenant_id", "card_id"})

	// Fraud metrics
	FraudSuspicionScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fuel_fraud_suspicion_score",
		Help: "Fraud suspicion score (0-100) per transaction",
	}, []string{"transaction_id", "tenant_id"})

	PotentialFraudTransactions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_potential_fraud_transactions_total",
		Help: "Transactions flagged for potential fraud",
	}, []string{"tenant_id", "fraud_type"})

	// Location validation
	LocationValidationFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_location_validation_failures_total",
		Help: "Transactions blocked due to location mismatch",
	}, []string{"tenant_id", "failure_type"})

	GeofenceViolationFuelTransactions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_geofence_violation_transactions_total",
		Help: "Fuel transactions outside approved geofences",
	}, []string{"tenant_id", "vehicle_id"})

	// Time validation
	OffHourFuelTransactions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_off_hour_transactions_total",
		Help: "Fuel transactions outside approved hours",
	}, []string{"tenant_id"})

	FuelTransactionDuplicates = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fuel_transaction_duplicates_total",
		Help: "Duplicate fuel transactions detected",
	}, []string{"tenant_id"})
)

// Global metrics instance
var GlobalMetrics = struct {
	FuelTransactionsTotal             *prometheus.CounterVec
	FuelTransactionsAmountTotal       *prometheus.CounterVec
	FuelTransactionsVolumeTotal       *prometheus.CounterVec
	FuelTransactionsFailedTotal       *prometheus.CounterVec
	FuelTransactionDuration           *prometheus.HistogramVec
	FuelOTPSentTotal                  *prometheus.CounterVec
	FuelOTPVerifiedTotal              *prometheus.CounterVec
	FuelOTPFailedTotal                *prometheus.CounterVec
	FuelOTPExpiredTotal               *prometheus.CounterVec
	FuelOTPBypassAttempts             *prometheus.CounterVec
	FuelReconciliationRunsTotal       *prometheus.CounterVec
	FuelTransactionsReconciledTotal   *prometheus.CounterVec
	FuelReconciliationGapsTotal       *prometheus.CounterVec
	FuelReconciliationMismatchesTotal *prometheus.CounterVec
	FuelAnomaliesDetectedTotal        *prometheus.CounterVec
	FuelAnomaliesResolvedTotal        *prometheus.CounterVec
	FuelAnomalyDetectionLatency       *prometheus.HistogramVec
	FuelCardsActiveTotal              *prometheus.GaugeVec
	FuelCardsBlockedTotal             *prometheus.GaugeVec
	FuelCardSpendLimitExceeded        *prometheus.CounterVec
	FuelStationsSupportedTotal        prometheus.Gauge
	FuelStationAvailability           *prometheus.GaugeVec
	FuelStationTransactionSuccessRate *prometheus.GaugeVec
	DailySpendLimitUtilization        *prometheus.GaugeVec
	FuelCardVolumeLimitUtilization    *prometheus.GaugeVec
	FraudSuspicionScore               *prometheus.GaugeVec
	PotentialFraudTransactions        *prometheus.CounterVec
	LocationValidationFailures        *prometheus.CounterVec
	GeofenceViolationFuelTransactions *prometheus.CounterVec
	OffHourFuelTransactions           *prometheus.CounterVec
	FuelTransactionDuplicates         *prometheus.CounterVec
}{
	FuelTransactionsTotal:             FuelTransactionsTotal,
	FuelTransactionsAmountTotal:       FuelTransactionsAmountTotal,
	FuelTransactionsVolumeTotal:       FuelTransactionsVolumeTotal,
	FuelTransactionsFailedTotal:       FuelTransactionsFailedTotal,
	FuelTransactionDuration:           FuelTransactionDuration,
	FuelOTPSentTotal:                  FuelOTPSentTotal,
	FuelOTPVerifiedTotal:              FuelOTPVerifiedTotal,
	FuelOTPFailedTotal:                FuelOTPFailedTotal,
	FuelOTPExpiredTotal:               FuelOTPExpiredTotal,
	FuelOTPBypassAttempts:             FuelOTPBypassAttempts,
	FuelReconciliationRunsTotal:       FuelReconciliationRunsTotal,
	FuelTransactionsReconciledTotal:   FuelTransactionsReconciledTotal,
	FuelReconciliationGapsTotal:       FuelReconciliationGapsTotal,
	FuelReconciliationMismatchesTotal: FuelReconciliationMismatchesTotal,
	FuelAnomaliesDetectedTotal:        FuelAnomaliesDetectedTotal,
	FuelAnomaliesResolvedTotal:        FuelAnomaliesResolvedTotal,
	FuelAnomalyDetectionLatency:       FuelAnomalyDetectionLatency,
	FuelCardsActiveTotal:              FuelCardsActiveTotal,
	FuelCardsBlockedTotal:             FuelCardsBlockedTotal,
	FuelCardSpendLimitExceeded:        FuelCardSpendLimitExceeded,
	FuelStationsSupportedTotal:        FuelStationsSupportedTotal,
	FuelStationAvailability:           FuelStationAvailability,
	FuelStationTransactionSuccessRate: FuelStationTransactionSuccessRate,
	DailySpendLimitUtilization:        DailySpendLimitUtilization,
	FuelCardVolumeLimitUtilization:    FuelCardVolumeLimitUtilization,
	FraudSuspicionScore:               FraudSuspicionScore,
	PotentialFraudTransactions:        PotentialFraudTransactions,
	LocationValidationFailures:        LocationValidationFailures,
	GeofenceViolationFuelTransactions: GeofenceViolationFuelTransactions,
	OffHourFuelTransactions:           OffHourFuelTransactions,
	FuelTransactionDuplicates:         FuelTransactionDuplicates,
}
