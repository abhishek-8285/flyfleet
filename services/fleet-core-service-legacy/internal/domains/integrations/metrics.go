package integrations

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// newJSONMarshal returns the standard json.Marshal for use in other files.
func newJSONMarshal() func(interface{}) ([]byte, error) {
	return json.Marshal
}

// ─── Integration Metrics ─────────────────────────────────────────────────────
//
// Prometheus-format metrics exposition for the integrations domain.
//
// Counters:
//   integration_events_total           — events routed (tenant, event_type labels)
//   integration_failures_total         — delivery failures
//
// Histograms:
//   webhook_delivery_latency_seconds   — per-webhook HTTP round-trip
//   connector_sync_duration_seconds    — per-job-type sync duration
//
// Gauges:
//   integration_dlq_size               — current DLQ entry count
//   integration_circuit_breakers_open  — open circuit breakers

type labeledIntCounter struct {
	mu   sync.RWMutex
	data map[string]int64 // key = labelset
}

func newLabeledIntCounter() *labeledIntCounter {
	return &labeledIntCounter{data: make(map[string]int64)}
}

func (c *labeledIntCounter) Inc(key string) {
	c.mu.Lock()
	c.data[key]++
	c.mu.Unlock()
}

func (c *labeledIntCounter) Snapshot() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int64, len(c.data))
	for k, v := range c.data {
		out[k] = v
	}
	return out
}

type IntegrationMetrics struct {
	events          *labeledIntCounter
	failures        *labeledIntCounter
	connectorErrors *labeledIntCounter
	webhookSuccess  *labeledIntCounter
	webhookTotal    *labeledIntCounter
	dlqSize         atomic.Int64
	openBreakers    atomic.Int64
	queueDepth      atomic.Int64

	// Histograms (unlabeled — high cardinality avoided at gauge level)
	webhookLatMu   sync.Mutex
	webhookLatBkts []float64 // seconds upper bounds
	webhookLatCnts []int64

	syncDurMu   sync.Mutex
	syncDurData map[string]*syncHist
}

type syncHist struct {
	buckets []float64
	counts  []int64
}

func newSyncHist() *syncHist {
	return &syncHist{
		buckets: []float64{1, 5, 30, 120, 600, 3600},
		counts:  make([]int64, 7),
	}
}

func (h *syncHist) record(secs float64) {
	for i, b := range h.buckets {
		if secs <= b {
			h.counts[i]++
			return
		}
	}
	h.counts[len(h.buckets)]++
}

func NewIntegrationMetrics() *IntegrationMetrics {
	return &IntegrationMetrics{
		events:          newLabeledIntCounter(),
		failures:        newLabeledIntCounter(),
		connectorErrors: newLabeledIntCounter(),
		webhookSuccess:  newLabeledIntCounter(),
		webhookTotal:    newLabeledIntCounter(),
		webhookLatBkts:  []float64{0.05, 0.1, 0.5, 1, 5, 15},
		webhookLatCnts:  make([]int64, 7),
		syncDurData:     make(map[string]*syncHist),
	}
}

func (m *IntegrationMetrics) IncEvents()                        { m.events.Inc("all") }
func (m *IntegrationMetrics) IncEventsLabeled(k string)         { m.events.Inc(k) }
func (m *IntegrationMetrics) IncFailures()                      { m.failures.Inc("all") }
func (m *IntegrationMetrics) IncFailuresLabeled(k string)       { m.failures.Inc(k) }
func (m *IntegrationMetrics) IncConnectorError(provider string) { m.connectorErrors.Inc(provider) }
func (m *IntegrationMetrics) IncWebhookDelivery(status string)  { 
	m.webhookTotal.Inc("all")
	if status == "success" {
		m.webhookSuccess.Inc("all")
	}
}
func (m *IntegrationMetrics) SetDLQSize(n int64)                { m.dlqSize.Store(n) }
func (m *IntegrationMetrics) SetOpenBreakers(n int64)           { m.openBreakers.Store(n) }
func (m *IntegrationMetrics) SetQueueDepth(n int64)             { m.queueDepth.Store(n) }

func (m *IntegrationMetrics) RecordWebhookLatency(secs float64) {
	m.webhookLatMu.Lock()
	defer m.webhookLatMu.Unlock()
	for i, b := range m.webhookLatBkts {
		if secs <= b {
			m.webhookLatCnts[i]++
			return
		}
	}
	m.webhookLatCnts[len(m.webhookLatBkts)]++
}

func (m *IntegrationMetrics) RecordSyncDuration(jobType string, secs float64) {
	m.syncDurMu.Lock()
	defer m.syncDurMu.Unlock()
	h, ok := m.syncDurData[jobType]
	if !ok {
		h = newSyncHist()
		m.syncDurData[jobType] = h
	}
	h.record(secs)
}

// ServeMetrics returns Prometheus text exposition format.
func (m *IntegrationMetrics) ServeMetrics() string {
	ts := time.Now().UnixMilli()
	var out string

	// integration_events_total
	out += "# HELP integration_events_total Total integration events routed\n"
	out += "# TYPE integration_events_total counter\n"
	for k, v := range m.events.Snapshot() {
		out += fmt.Sprintf("integration_events_total{key=%q} %d %d\n", k, v, ts)
	}

	// integration_failures_total
	out += "# HELP integration_failures_total Total integration delivery failures\n"
	out += "# TYPE integration_failures_total counter\n"
	for k, v := range m.failures.Snapshot() {
		out += fmt.Sprintf("integration_failures_total{key=%q} %d %d\n", k, v, ts)
	}

	// connector_error_rate
	out += "# HELP connector_error_rate Total connector execution errors\n"
	out += "# TYPE connector_error_rate counter\n"
	for k, v := range m.connectorErrors.Snapshot() {
		out += fmt.Sprintf("connector_error_rate{provider=%q} %d %d\n", k, v, ts)
	}

	// webhook_delivery_success_rate (represented as ratio metric using counter data or gauges)
	// We expose the success and total counters separately so Prometheus can calculate the rate:
	// rate(webhook_deliveries_success[5m]) / rate(webhook_deliveries_total[5m])
	out += "# HELP webhook_deliveries_success Successful webhook deliveries\n"
	out += "# TYPE webhook_deliveries_success counter\n"
	for k, v := range m.webhookSuccess.Snapshot() {
		out += fmt.Sprintf("webhook_deliveries_success{key=%q} %d %d\n", k, v, ts)
	}
	out += "# HELP webhook_deliveries_total Total webhook delivery attempts\n"
	out += "# TYPE webhook_deliveries_total counter\n"
	for k, v := range m.webhookTotal.Snapshot() {
		out += fmt.Sprintf("webhook_deliveries_total{key=%q} %d %d\n", k, v, ts)
	}

	// Gauges
	out += fmt.Sprintf("# HELP integration_dlq_size Current DLQ entry count\n")
	out += fmt.Sprintf("# TYPE integration_dlq_size gauge\n")
	out += fmt.Sprintf("integration_dlq_size %d %d\n", m.dlqSize.Load(), ts)

	// integration_queue_depth
	out += fmt.Sprintf("# HELP integration_queue_depth Current pending integration event queue depth\n")
	out += fmt.Sprintf("# TYPE integration_queue_depth gauge\n")
	out += fmt.Sprintf("integration_queue_depth %d %d\n", m.queueDepth.Load(), ts)

	out += fmt.Sprintf("# HELP circuit_breaker_open_count Open circuit breakers\n")
	out += fmt.Sprintf("# TYPE circuit_breaker_open_count gauge\n")
	out += fmt.Sprintf("circuit_breaker_open_count %d %d\n", m.openBreakers.Load(), ts)

	// webhook_delivery_latency_seconds histogram
	m.webhookLatMu.Lock()
	out += "# HELP webhook_delivery_latency_seconds Webhook HTTP delivery latency\n"
	out += "# TYPE webhook_delivery_latency_seconds histogram\n"
	for i, b := range m.webhookLatBkts {
		out += fmt.Sprintf(`webhook_delivery_latency_seconds_bucket{le="%.2f"} %d %d`+"\n", b, m.webhookLatCnts[i], ts)
	}
	out += fmt.Sprintf(`webhook_delivery_latency_seconds_bucket{le="+Inf"} %d %d`+"\n", m.webhookLatCnts[len(m.webhookLatBkts)], ts)
	m.webhookLatMu.Unlock()

	// connector_sync_duration_seconds histogram (per job type)
	m.syncDurMu.Lock()
	for jobType, h := range m.syncDurData {
		out += fmt.Sprintf("# HELP connector_sync_duration_seconds Sync job duration\n")
		out += fmt.Sprintf("# TYPE connector_sync_duration_seconds histogram\n")
		for i, b := range h.buckets {
			out += fmt.Sprintf(`connector_sync_duration_seconds_bucket{job_type=%q,le="%.0f"} %d %d`+"\n", jobType, b, h.counts[i], ts)
		}
		out += fmt.Sprintf(`connector_sync_duration_seconds_bucket{job_type=%q,le="+Inf"} %d %d`+"\n", jobType, h.counts[len(h.buckets)], ts)
	}
	m.syncDurMu.Unlock()

	return out
}
