package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ViolationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "compliance_driver_violations_total",
		Help: "The total number of driver compliance violations detected",
	})

	PanicAlertsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "compliance_panic_alerts_total",
		Help: "The total number of panic alerts triggered",
	})

	DevicesOffline = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "compliance_devices_offline",
		Help: "The current number of offline AIS-140 devices",
	})

	DocumentExpiryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "compliance_document_expiry_today",
		Help: "The total number of documents expiring today",
	})
)

func RecordViolationMetric() {
	ViolationsTotal.Inc()
}

func RecordPanicMetric() {
	PanicAlertsTotal.Inc()
}

func SetOfflineDevicesMetric(count float64) {
	DevicesOffline.Set(count)
}
