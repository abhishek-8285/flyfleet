package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DiagnosticsDevicesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "diagnostics_devices_total",
		Help: "The total number of registered devices",
	})

	DiagnosticsDevicesOnline = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "diagnostics_devices_online",
		Help: "The total number of online devices",
	})

	DiagnosticsCommandsFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "diagnostics_commands_failed_total",
		Help: "Total count of failed device commands",
	})

	DiagnosticsRolloutsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "diagnostics_rollouts_active",
		Help: "The current number of active firmware rollouts",
	})

	DiagnosticsTelemetryRate = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "diagnostics_telemetry_rate",
		Help:    "Histogram of telemetry arrival rate tracking",
		Buckets: prometheus.LinearBuckets(1, 1, 10),
	})
)

func RecordDeviceOnlineStatus(online bool) {
	if online {
		DiagnosticsDevicesOnline.Inc()
	} else {
		DiagnosticsDevicesOnline.Dec()
	}
}

func RecordCommandFailure() {
	DiagnosticsCommandsFailedTotal.Inc()
}
