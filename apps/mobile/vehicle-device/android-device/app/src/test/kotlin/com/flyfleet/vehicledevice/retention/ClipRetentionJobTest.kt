package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.time.Duration.Companion.days
import kotlin.time.Duration.Companion.hours

class ClipRetentionJobTest {
    private val storageHealthy = { StorageState(usedMb = 2000, availableMb = 10000) }

    @Test
    fun `evaluates multiple clips`() {
        val policy = RetentionPolicy(RetentionConfig(criticalRetentionDays = 1, evidenceReserveMb = 0), storageHealthy) { 0L }
        val service = ClipRetentionService(policy)
        val job = ClipRetentionJob(service)

        val clips = listOf(
            ClipMetadata(id = "c1", ageMs = 10.days.inWholeMilliseconds, isCritical = false),
            ClipMetadata(id = "c2", ageMs = 1.hours.inWholeMilliseconds, isCritical = false),
        )

        val results = job.run(clips)
        assertEquals(2, results.size)
        assertEquals("c1", results[0].clipId)
        assertEquals(true, results[0].allowed)
        assertEquals("c2", results[1].clipId)
        assertEquals(false, results[1].allowed)
    }
}
