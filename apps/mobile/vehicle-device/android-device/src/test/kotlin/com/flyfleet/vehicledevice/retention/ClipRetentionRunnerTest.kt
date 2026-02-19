package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals

class ClipRetentionRunnerTest {
    @Test
    fun `runs job against fetched clips`() {
        val policy = RetentionPolicy(RetentionConfig(criticalRetentionDays = 0, evidenceReserveMb = 0), { StorageState(0, 1000) }) { 1 }
        val job = ClipRetentionJob(ClipRetentionService(policy))
        val provider = ClipMetadataProvider {
            listOf(
                ClipMetadata(id = "clip-run-1", ageMs = 20.days.inWholeMilliseconds, isCritical = false),
            )
        }

        val runner = ClipRetentionRunner(job, provider)
        val results = runner.run()

        assertEquals(1, results.size)
        assertEquals("clip-run-1", results[0].clipId)
        assertEquals(true, results[0].allowed)
    }
}
