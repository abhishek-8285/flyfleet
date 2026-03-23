package driver

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the prometheus metrics for the Driver Module.
type Metrics struct {
	ActiveDrivers       prometheus.Gauge
	DriversOnTrip       prometheus.Gauge
	SOSAlertsTotal      prometheus.Counter
	BehaviorEventsTotal *prometheus.CounterVec
}

// NewMetrics creates a new set of metrics for the Driver Module.
func NewMetrics() *Metrics {
	return &Metrics{
		ActiveDrivers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "flyfleet_driver_active_total",
			Help: "Total number of active drivers",
		}),
		DriversOnTrip: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "flyfleet_driver_on_trip_total",
			Help: "Total number of drivers currently on a trip",
		}),
		SOSAlertsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "flyfleet_driver_sos_alerts_total",
			Help: "Total number of driver SOS alerts triggered",
		}),
		BehaviorEventsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "flyfleet_driver_behavior_events_total",
			Help: "Total number of driver behavior events by type",
		}, []string{"event_type"}),
	}
}

var (
	GlobalMetrics = NewMetrics()
)
