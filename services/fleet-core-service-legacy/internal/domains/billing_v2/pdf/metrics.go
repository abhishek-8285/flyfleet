package pdf

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RenderLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "billing_invoice_render_latency_seconds",
		Help:    "Latency of invoice PDF rendering",
		Buckets: prometheus.DefBuckets,
	}, []string{"template_name"})

	RenderFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "billing_invoice_render_failures_total",
		Help: "Total number of failed invoice renderings",
	}, []string{"template_name", "reason"})

	QueueBacklog = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "billing_invoice_render_queue_backlog",
		Help: "Current number of pending invoice rendering tasks",
	})
)

type Metrics struct{}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) ObserveRender(templateName string, duration float64) {
	RenderLatency.WithLabelValues(templateName).Observe(duration)
}

func (m *Metrics) IncFailure(templateName string, reason string) {
	RenderFailures.WithLabelValues(templateName, reason).Inc()
}
