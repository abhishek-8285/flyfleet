package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import kotlin.time.Duration.Companion.days
import kotlin.time.Duration.Companion.hours

class RetentionPolicyTest {
    private val now = 1_000_000L
    private val storageHealthy = { StorageState(usedMb = 2000, availableMb = 10000) }

    @Test
    fun `critical retention prevents overwrite for recent clips`() {
        val policy = RetentionPolicy(RetentionConfig(), storageHealthy) { now }
        val decision = policy.shouldOverwriteNormalData(ageMs = 1.hours.inWholeMilliseconds)
        assertFalse(decision.allowOverwrite)
        assertEquals("within_critical_retention", decision.reason)
    }

    @Test
    fun `legal hold blocks all overwrites`() {
        val policy = RetentionPolicy(RetentionConfig(legalHold = true), storageHealthy) { now }
        val decision = policy.shouldOverwriteNormalData(ageMs = 40.days.inWholeMilliseconds)
        assertFalse(decision.allowOverwrite)
        assertEquals("legal_hold_active", decision.reason)
    }

    @Test
    fun `reserve threshold prevents overwrite when free space low`() {
        val lowStorage = { StorageState(usedMb = 4000, availableMb = 400) }
        val policy = RetentionPolicy(RetentionConfig(evidenceReserveMb = 500), lowStorage) { now }
        val decision = policy.shouldOverwriteNormalData(ageMs = 100.days.inWholeMilliseconds)
        assertFalse(decision.allowOverwrite)
        assertEquals("reserve_threshold", decision.reason)
    }

    @Test
    fun `old data allowed when reserves satisfied`() {
        val policy = RetentionPolicy(RetentionConfig(criticalRetentionDays = 7, evidenceReserveMb = 200), storageHealthy) { now }
        val decision = policy.shouldOverwriteNormalData(ageMs = 20.days.inWholeMilliseconds)
        assertTrue(decision.allowOverwrite)
        assertEquals("old_data_overwrite_allowed", decision.reason)
    }
}
