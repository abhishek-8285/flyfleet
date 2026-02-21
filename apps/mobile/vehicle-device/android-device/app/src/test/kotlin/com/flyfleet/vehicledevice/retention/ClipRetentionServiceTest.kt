package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.time.Duration.Companion.days

class ClipRetentionServiceTest {
    private val storageHealthy = { StorageState(usedMb = 2000, availableMb = 10000) }

    @Test
    fun `delegates to retention manager decisions`() {
        val policy = RetentionPolicy(RetentionConfig(evidenceReserveMb = 0), storageHealthy) { 0L }
        val service = ClipRetentionService(policy)

        val action = service.evaluateClip(ClipMetadata(id = "clip-3", ageMs = 50.days.inWholeMilliseconds, isCritical = false))
        assertEquals("old_data_overwrite_allowed", action.reason)
        assertEquals(true, action.canDelete)
    }
}
