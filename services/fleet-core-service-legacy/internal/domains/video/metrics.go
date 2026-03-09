package video

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ─── Labeled Counter ──────────────────────────────────────────────────────────

// labelKey uniquely identifies a label combination.
type labelKey struct {
	Tenant     string
	CameraType string // camera_position value
	VideoType  string
}

// labeledCounter maps label combinations to int64 counts.
type labeledCounter struct {
	mu     sync.RWMutex
	counts map[labelKey]int64
}

func newLabeledCounter() *labeledCounter {
	return &labeledCounter{counts: make(map[labelKey]int64)}
}

func (c *labeledCounter) Inc(k labelKey) {
	c.mu.Lock()
	c.counts[k]++
	c.mu.Unlock()
}

func (c *labeledCounter) Add(k labelKey, n int64) {
	c.mu.Lock()
	c.counts[k] += n
	c.mu.Unlock()
}

// promLines serializes labeled counts into Prometheus text exposition format.
func (c *labeledCounter) promLines(metricName, help, metricType string, ts int64) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := fmt.Sprintf("# HELP %s %s\n# TYPE %s %s\n", metricName, help, metricName, metricType)
	for k, v := range c.counts {
		out += fmt.Sprintf(`%s{tenant="%s",camera_type="%s",video_type="%s"} %d %d`+"\n",
			metricName, k.Tenant, k.CameraType, k.VideoType, v, ts)
	}
	return out
}

// ─── VideoMetrics ─────────────────────────────────────────────────────────────

// VideoMetrics tracks key operational metrics for the video module,
// partitioned by tenant, camera_type, and video_type labels.
// In production, swap with github.com/prometheus/client_golang for richer histograms.
type VideoMetrics struct {
	logger *zap.Logger

	// Labeled counters — broken down by tenant / camera_type / video_type
	uploadCount   *labeledCounter // video_uploads_total
	uploadBytes   *labeledCounter // video_upload_bytes_total
	aiDetections  *labeledCounter // video_ai_detections_total
	cameraOffline *labeledCounter // video_camera_offline_total

	// Non-labeled scalar gauges / counters
	processingQueued atomic.Int64 // video_processing_queue_depth (gauge)
	dlqSize          atomic.Int64 // video_dlq_size (gauge)

	// Transcode duration histogram (shared across tenants)
	transcodeMu      sync.Mutex
	transcodeBuckets []float64
	transcodeCounts  []int64

	// Per-stage duration histogram with label: transcode|thumbnail|ai|finalize
	stageMu     sync.Mutex
	stageData   map[string]*stageHistogram // key = stage name
}

type stageHistogram struct {
	buckets []float64 // seconds upper bounds
	counts  []int64   // len(buckets)+1
}

func newStageHistogram() *stageHistogram {
	// Buckets: <1s, <5s, <30s, <60s, <120s, <600s, +Inf
	return &stageHistogram{
		buckets: []float64{1, 5, 30, 60, 120, 600},
		counts:  make([]int64, 7),
	}
}

func (h *stageHistogram) record(seconds float64) {
	for i, b := range h.buckets {
		if seconds <= b {
			h.counts[i]++
			return
		}
	}
	h.counts[len(h.buckets)]++
}

func NewVideoMetrics(logger *zap.Logger) *VideoMetrics {
	return &VideoMetrics{
		logger:           logger,
		uploadCount:      newLabeledCounter(),
		uploadBytes:      newLabeledCounter(),
		aiDetections:     newLabeledCounter(),
		cameraOffline:    newLabeledCounter(),
		transcodeBuckets: []float64{30, 60, 300, 600},
		transcodeCounts:  make([]int64, 5),
		stageData:        make(map[string]*stageHistogram),
	}
}

// ─── Increment Helpers ────────────────────────────────────────────────────────

func (m *VideoMetrics) IncUpload(tenant, cameraType, videoType string, bytes int64) {
	k := labelKey{tenant, cameraType, videoType}
	m.uploadCount.Inc(k)
	m.uploadBytes.Add(k, bytes)
}

func (m *VideoMetrics) IncAIDetection(tenant, cameraType, videoType string) {
	m.aiDetections.Inc(labelKey{tenant, cameraType, videoType})
}

func (m *VideoMetrics) IncCameraOffline(tenant, cameraType string) {
	m.cameraOffline.Inc(labelKey{Tenant: tenant, CameraType: cameraType})
}

func (m *VideoMetrics) SetQueueDepth(n int64) { m.processingQueued.Store(n) }
func (m *VideoMetrics) SetDLQSize(n int64)     { m.dlqSize.Store(n) }

func (m *VideoMetrics) RecordTranscodeDuration(seconds float64) {
	m.transcodeMu.Lock()
	defer m.transcodeMu.Unlock()
	for i, b := range m.transcodeBuckets {
		if seconds <= b {
			m.transcodeCounts[i]++
			return
		}
	}
	m.transcodeCounts[len(m.transcodeBuckets)]++
}

// RecordStageDuration records how long a pipeline stage took.
// stage must be one of: "transcode", "thumbnail", "ai", "finalize".
func (m *VideoMetrics) RecordStageDuration(stage string, seconds float64) {
	m.stageMu.Lock()
	defer m.stageMu.Unlock()
	h, ok := m.stageData[stage]
	if !ok {
		h = newStageHistogram()
		m.stageData[stage] = h
	}
	h.record(seconds)
}

// ─── Exposition ───────────────────────────────────────────────────────────────

// ServeMetrics returns all video metrics in Prometheus text exposition format.
// Labeled dimensions: tenant, camera_type, video_type.
func (m *VideoMetrics) ServeMetrics() string {
	ts := time.Now().UnixMilli()

	// Labeled sections
	out := m.uploadCount.promLines("video_uploads_total", "Total videos uploaded", "counter", ts)
	out += "\n"
	out += m.uploadBytes.promLines("video_upload_bytes_total", "Total bytes uploaded", "counter", ts)
	out += "\n"
	out += m.aiDetections.promLines("video_ai_detections_total", "Total AI safety detections", "counter", ts)
	out += "\n"
	out += m.cameraOffline.promLines("video_camera_offline_total", "Camera offline event count", "counter", ts)
	out += "\n"

	// Scalar gauges
	out += fmt.Sprintf("# HELP video_processing_queue_depth Current queued processing jobs\n"+
		"# TYPE video_processing_queue_depth gauge\n"+
		"video_processing_queue_depth %d %d\n\n"+
		"# HELP video_dlq_size DLQ entry count\n"+
		"# TYPE video_dlq_size gauge\n"+
		"video_dlq_size %d %d\n\n",
		m.processingQueued.Load(), ts,
		m.dlqSize.Load(), ts)

	// Histogram
	m.transcodeMu.Lock()
	out += "# HELP video_transcode_duration_seconds Transcode duration\n" +
		"# TYPE video_transcode_duration_seconds histogram\n"
	for i, b := range m.transcodeBuckets {
		out += fmt.Sprintf(`video_transcode_duration_seconds_bucket{le="%.0f"} %d %d`+"\n",
			b, m.transcodeCounts[i], ts)
	}
	out += fmt.Sprintf(`video_transcode_duration_seconds_bucket{le="+Inf"} %d %d`+"\n",
		m.transcodeCounts[len(m.transcodeBuckets)], ts)
	m.transcodeMu.Unlock()

	// Stage duration histograms (labeled by stage)
	m.stageMu.Lock()
	for stage, h := range m.stageData {
		out += fmt.Sprintf("# HELP processing_stage_duration_seconds Duration of each processing stage\n")
		out += fmt.Sprintf("# TYPE processing_stage_duration_seconds histogram\n")
		for i, b := range h.buckets {
			out += fmt.Sprintf(`processing_stage_duration_seconds_bucket{stage="%s",le="%.0f"} %d %d`+"\n",
				stage, b, h.counts[i], ts)
		}
		out += fmt.Sprintf(`processing_stage_duration_seconds_bucket{stage="%s",le="+Inf"} %d %d`+"\n",
			stage, h.counts[len(h.buckets)], ts)
	}
	m.stageMu.Unlock()

	return out
}

// LogSnapshot emits a periodic metrics snapshot to structured logs.
func (m *VideoMetrics) LogSnapshot(ctx context.Context) {
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.logger.Info("video_metrics_snapshot",
				zap.Int64("queue_depth", m.processingQueued.Load()),
				zap.Int64("dlq_size", m.dlqSize.Load()))
		}
	}
}
