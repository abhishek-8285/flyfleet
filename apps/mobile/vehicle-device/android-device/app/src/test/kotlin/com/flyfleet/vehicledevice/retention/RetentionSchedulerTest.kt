package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertTrue

class RetentionSchedulerTest {
    @Test
    fun `runs worker on schedule`() {
        var runCount = 0
        val supervisor = RetentionSupervisor(
            configProvider = object : RetentionConfigProvider(java.nio.file.Path.of("/tmp")) {
                override fun load() = RetentionConfig()
            },
            metadataProvider = { emptyList() },
            storageStateProvider = { StorageState(0, 1000) },
            clockProvider = { 0L },
        )
        val worker = object : RetentionCleanupWorker(supervisor, { }) {
            override fun runCycle() {
                runCount += 1
                super.runCycle()
            }
        }
        val scheduler = RetentionScheduler(worker, periodMs = 10)
        scheduler.start()
        Thread.sleep(35)
        scheduler.stop()
        assertTrue(runCount >= 2)
    }
}
