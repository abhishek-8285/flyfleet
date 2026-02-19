package com.flyfleet.vehicledevice.retention

data class RetentionMetricsSnapshot(
    val totalClips: Int,
    val allowedDeletes: Int,
    val blockedDeletes: Int,
)

class RetentionMetrics {
    private var total = 0
    private var allowed = 0
    private var blocked = 0

    fun record(result: ClipRetentionTaskResult) {
        total += 1
        if (result.allowed) {
            allowed += 1
        } else {
            blocked += 1
        }
    }

    fun snapshot(): RetentionMetricsSnapshot = RetentionMetricsSnapshot(
        totalClips = total,
        allowedDeletes = allowed,
        blockedDeletes = blocked,
    )
}
