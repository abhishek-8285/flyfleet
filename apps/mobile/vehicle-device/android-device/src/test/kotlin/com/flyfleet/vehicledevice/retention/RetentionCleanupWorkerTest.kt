package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertSame

class RetentionCleanupWorkerTest {
    @Test
    fun `runs supervisor cycle and reports results`() {
        val configFile = java.nio.file.Files.createTempFile("retention", ".json")
        configFile.toFile().writeText("{\"retention\":{\"critical_clip_days\":0,\"min_storage_reserve_mb\":0,\"legal_hold\":false}}")
        val supervisor = RetentionSupervisor(
            configProvider = RetentionConfigProvider(configFile),
            metadataProvider = ClipMetadataProvider { listOf(ClipMetadata("c1", 100.days.inWholeMilliseconds, false)) },
            storageStateProvider = { StorageState(0, 1000) },
            clockProvider = { 0L },
        )
        val results = mutableListOf<List<ClipRetentionTaskResult>>()
        val worker = RetentionCleanupWorker(supervisor) { results += it }

        worker.runCycle()
        assertEquals(1, results.size)
        assertEquals("c1", results.first().first().clipId)
        assertEquals(1, worker.metricsSnapshot().totalClips)
        assertEquals(1, worker.metricsSnapshot().allowedDeletes)
    }
}
