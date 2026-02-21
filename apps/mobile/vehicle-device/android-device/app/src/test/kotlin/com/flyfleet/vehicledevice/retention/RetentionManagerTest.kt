package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlin.time.Duration.Companion.days

class RetentionManagerTest {
    private val storageHealthy = { StorageState(usedMb = 2000, availableMb = 10000) }

    @Test
    fun `critical clip is never deleted even if old`() {
        val policy = RetentionPolicy(RetentionConfig(), storageHealthy) { 0L }
        val manager = RetentionManager(policy)

        val action = manager.evaluateClip(ClipMetadata(id = "clip-1", ageMs = 100.days.inWholeMilliseconds, isCritical = true))
        assertFalse(action.canDelete)
        assertEquals("critical_evidence", action.reason)
    }

    @Test
    fun `non critical old clip respects policy`() {
        val policy = RetentionPolicy(RetentionConfig(criticalRetentionDays = 1, evidenceReserveMb = 0), storageHealthy) { 0L }
        val manager = RetentionManager(policy)

        val action = manager.evaluateClip(ClipMetadata(id = "clip-2", ageMs = 5.days.inWholeMilliseconds, isCritical = false))
        assertTrue(action.canDelete)
        assertEquals("old_data_overwrite_allowed", action.reason)
    }
}
