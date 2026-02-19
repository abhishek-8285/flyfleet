package com.flyfleet.vehicledevice.retention

class RetentionCleanupWorker(
    private val supervisor: RetentionSupervisor,
    private val cleanupCallback: (List<ClipRetentionTaskResult>) -> Unit,
    private val metrics: RetentionMetrics = RetentionMetrics(),
) {
    fun runCycle() {
        val results = supervisor.run()
        results.forEach(metrics::record)
        cleanupCallback(results)
    }

    fun metricsSnapshot(): RetentionMetricsSnapshot = metrics.snapshot()
}
