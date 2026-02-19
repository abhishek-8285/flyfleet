package com.flyfleet.vehicledevice.retention

import kotlin.concurrent.fixedRateTimer

class RetentionScheduler(
    private val worker: RetentionCleanupWorker,
    private val periodMs: Long,
) {
    private var timer: java.util.Timer? = null

    fun start() {
        if (timer != null) return
        timer = fixedRateTimer("retention-cleanup", daemon = true, period = periodMs) {
            worker.runCycle()
        }
    }

    fun stop() {
        timer?.cancel()
        timer = null
    }
}
