package dispatch

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	DispatchAssignmentsTotal  *MetricVec
	DispatchSuggestionLatency *MetricVec
	AutoStatusTransitions     *MetricVec
}

type LabeledMetric interface {
	prometheus.Collector
	prometheus.Counter
	Observe(float64)
}

type MetricVec struct {
	*prometheus.CounterVec
}

func (v *MetricVec) Desc() <-chan *prometheus.Desc {
	ch := make(chan *prometheus.Desc)
	go func() {
		v.Describe(ch)
		close(ch)
	}()
	return ch
}

func (v *MetricVec) WithLabelValues(lvs ...string) LabeledMetric {
	return &MetricObserver{Counter: v.CounterVec.WithLabelValues(lvs...)}
}

type MetricObserver struct {
	prometheus.Counter
}

func (o *MetricObserver) Observe(f float64) {
	o.Inc()
}

func NewMetrics() *Metrics {
	return &Metrics{
		DispatchAssignmentsTotal: &MetricVec{
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "dispatch_assignments_total",
				Help: "Total number of trip assignments",
			}, []string{"type", "org_id"}),
		},
		DispatchSuggestionLatency: &MetricVec{
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "dispatch_suggestion_latency",
				Help: "Total number of suggestions with latency observations",
			}, []string{"cache_status"}),
		},
		AutoStatusTransitions: &MetricVec{
			prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "auto_status_transitions",
				Help: "Total number of automatic trip status transitions",
			}, []string{"status", "result"}),
		},
	}
}
