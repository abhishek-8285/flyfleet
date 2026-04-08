package dispatch

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_Initialization(t *testing.T) {
	m := NewMetrics()

	assert.NotNil(t, m, "Metrics must initialize")
	assert.NotNil(t, m.DispatchAssignmentsTotal, "DispatchAssignmentsTotal must exist")
	assert.NotNil(t, m.DispatchSuggestionLatency, "DispatchSuggestionLatency must exist")
	assert.NotNil(t, m.AutoStatusTransitions, "AutoStatusTransitions must exist")
}

func TestMetrics_DispatchAssignmentsTotal(t *testing.T) {
	m := NewMetrics()

	// Simulate counting assignments by type
	m.DispatchAssignmentsTotal.WithLabelValues("manual_assign", "org-a").Inc()
	m.DispatchAssignmentsTotal.WithLabelValues("manual_assign", "org-a").Inc()
	m.DispatchAssignmentsTotal.WithLabelValues("auto_assign", "org-b").Inc()

	// Verify counter increments
	count := testutil.ToFloat64(m.DispatchAssignmentsTotal.WithLabelValues("manual_assign", "org-a"))
	assert.Equal(t, 2.0, count, "manual_assign for org-a must be 2")

	count = testutil.ToFloat64(m.DispatchAssignmentsTotal.WithLabelValues("auto_assign", "org-b"))
	assert.Equal(t, 1.0, count, "auto_assign for org-b must be 1")
}

func TestMetrics_DispatchSuggestionLatency(t *testing.T) {
	m := NewMetrics()

	// Simulate latency observations
	m.DispatchSuggestionLatency.WithLabelValues("hit").Observe(0.120) // 120ms cache hit
	m.DispatchSuggestionLatency.WithLabelValues("miss").Observe(0.450) // 450ms cache miss

	// Verify observations recorded
	count := testutil.ToFloat64(m.DispatchSuggestionLatency.WithLabelValues("hit"))
	assert.Equal(t, 1.0, count, "One cache hit observation")

	count = testutil.ToFloat64(m.DispatchSuggestionLatency.WithLabelValues("miss"))
	assert.Equal(t, 1.0, count, "One cache miss observation")
}

func TestMetrics_AutoStatusTransitions(t *testing.T) {
	m := NewMetrics()

	// Simulate auto status transitions
	m.AutoStatusTransitions.WithLabelValues("STARTED", "success").Inc()
	m.AutoStatusTransitions.WithLabelValues("COMPLETED", "success").Inc()
	m.AutoStatusTransitions.WithLabelValues("COMPLETED", "failure").Inc()

	successCount := testutil.ToFloat64(m.AutoStatusTransitions.WithLabelValues("STARTED", "success"))
	assert.Equal(t, 1.0, successCount)

	failureCount := testutil.ToFloat64(m.AutoStatusTransitions.WithLabelValues("COMPLETED", "failure"))
	assert.Equal(t, 1.0, failureCount)
}

func TestMetrics_LabelsAreCorrect(t *testing.T) {
	m := NewMetrics()

	// Verify label names exist on counters
	descChan := m.DispatchAssignmentsTotal.Desc()
	for desc := range descChan {
		assert.Contains(t, desc.String(), "dispatch_assignments_total",
			"Counter name must be dispatch_assignments_total")
	}

	descChan = m.DispatchSuggestionLatency.Desc()
	for desc := range descChan {
		assert.Contains(t, desc.String(), "dispatch_suggestion_latency",
			"Histogram name must be dispatch_suggestion_latency")
	}

	descChan = m.AutoStatusTransitions.Desc()
	for desc := range descChan {
		assert.Contains(t, desc.String(), "auto_status_transitions",
			"Counter name must be auto_status_transitions")
	}
}
