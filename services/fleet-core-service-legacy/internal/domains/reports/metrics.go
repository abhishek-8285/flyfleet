package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Metrics collects observability counters and gauges for the reports module.
type Metrics struct {
	logger        *zap.Logger
	reportsCreated   int64
	reportsCompleted int64
	reportsFailed    int64
	rowsGenerated    int64 // total rows across all completed reports
	queueDepth       int64
	workersBusy      int64
	totalWorkersN    int64

	// Simple histogram: count of reports completing within each bucket (seconds)
	// Buckets: <5s, <30s, <60s, <300s, <600s, >=600s
	durationBuckets [6]int64
}

func NewMetrics(logger *zap.Logger, totalWorkers int) *Metrics {
	return &Metrics{logger: logger, totalWorkersN: int64(totalWorkers)}
}

func (m *Metrics) IncCreated() {
	m.reportsCreated++
}

func (m *Metrics) IncCompleted(durationSeconds float64, rowCount int64) {
	m.reportsCompleted++
	m.rowsGenerated += rowCount
	// Record into histogram bucket
	switch {
	case durationSeconds < 5:
		m.durationBuckets[0]++
	case durationSeconds < 30:
		m.durationBuckets[1]++
	case durationSeconds < 60:
		m.durationBuckets[2]++
	case durationSeconds < 300:
		m.durationBuckets[3]++
	case durationSeconds < 600:
		m.durationBuckets[4]++
	default:
		m.durationBuckets[5]++
	}
	m.logger.Info("report_completed",
		zap.Float64("duration_s", durationSeconds),
		zap.Int64("row_count", rowCount),
		zap.Int64("rows_total", m.rowsGenerated))
}

func (m *Metrics) IncFailed() {
	m.reportsFailed++
}

func (m *Metrics) SetQueueDepth(depth int64) {
	m.queueDepth = depth
}

func (m *Metrics) WorkerBusy() {
	m.workersBusy++
}

func (m *Metrics) WorkerIdle() {
	if m.workersBusy > 0 {
		m.workersBusy--
	}
}

func (m *Metrics) WorkerUtilizationRatio() float64 {
	if m.totalWorkersN == 0 {
		return 0
	}
	return float64(m.workersBusy) / float64(m.totalWorkersN)
}

// LogSnapshot emits a periodic metrics snapshot to structured logs.
func (m *Metrics) LogSnapshot(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.logger.Info("reports_metrics_snapshot",
				zap.Int64("created", m.reportsCreated),
				zap.Int64("completed", m.reportsCompleted),
				zap.Int64("failed", m.reportsFailed),
				zap.Int64("queue_depth", m.queueDepth),
				zap.Float64("worker_utilization", m.WorkerUtilizationRatio()))
		}
	}
}

// ─── Prometheus Text Format ──────────────────────────────────────────────────

// PrometheusMetrics embeds Metrics and adds a Prometheus text/plain output method.
// In production, replace with github.com/prometheus/client_golang for full Prometheus integration.
type PrometheusMetrics struct {
	*Metrics
}

func NewPrometheusMetrics(logger *zap.Logger, totalWorkers int) *PrometheusMetrics {
	return &PrometheusMetrics{Metrics: NewMetrics(logger, totalWorkers)}
}

// ServeMetrics returns current metric values in Prometheus exposition format.
func (m *PrometheusMetrics) ServeMetrics() string {
	ts := time.Now().UnixMilli()
	bucketLabels := []string{"lt_5", "lt_30", "lt_60", "lt_300", "lt_600", "gte_600"}
	var histLines string
	for i, label := range bucketLabels {
		histLines += fmt.Sprintf(`report_generation_duration_bucket{le="%s"} %d %d`+"\n",
			label, m.durationBuckets[i], ts)
	}
	return fmt.Sprintf(
		"# HELP reports_created_total Total reports created\n"+
			"# TYPE reports_created_total counter\n"+
			"reports_created_total %d %d\n\n"+
			"# HELP reports_completed_total Total reports completed\n"+
			"# TYPE reports_completed_total counter\n"+
			"reports_completed_total %d %d\n\n"+
			"# HELP reports_failed_total Total reports failed\n"+
			"# TYPE reports_failed_total counter\n"+
			"reports_failed_total %d %d\n\n"+
			"# HELP report_rows_generated_total Total rows across all completed reports\n"+
			"# TYPE report_rows_generated_total counter\n"+
			"report_rows_generated_total %d %d\n\n"+
			"# HELP report_queue_depth Current pending job count\n"+
			"# TYPE report_queue_depth gauge\n"+
			"report_queue_depth %d %d\n\n"+
			"# HELP report_worker_utilization_ratio Fraction of busy workers\n"+
			"# TYPE report_worker_utilization_ratio gauge\n"+
			"report_worker_utilization_ratio %.4f %d\n\n"+
			"# HELP report_generation_duration_bucket Duration histogram (seconds)\n"+
			"# TYPE report_generation_duration_bucket histogram\n"+
			"%s",
		m.reportsCreated, ts,
		m.reportsCompleted, ts,
		m.reportsFailed, ts,
		m.rowsGenerated, ts,
		m.queueDepth, ts,
		m.WorkerUtilizationRatio(), ts,
		histLines,
	)
}

// ─── DLQ Requeue Helper ──────────────────────────────────────────────────────

// RequeueDLQJob moves a permanently-failed job back to QUEUED for re-execution.
func RequeueDLQJob(ctx context.Context, repo Repository, jobIDStr string, logger *zap.Logger) error {
	if repo == nil {
		return fmt.Errorf("repository not configured")
	}
	parsed, err := uuid.Parse(jobIDStr)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}
	if err := repo.RequeueDLQJob(ctx, parsed); err != nil {
		logger.Error("Failed to requeue DLQ job", zap.String("job_id", jobIDStr), zap.Error(err))
		return err
	}
	logger.Info("DLQ job requeued", zap.String("job_id", jobIDStr))
	return nil
}
