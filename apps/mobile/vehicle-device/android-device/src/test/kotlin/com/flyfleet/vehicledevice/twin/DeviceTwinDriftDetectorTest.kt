package com.flyfleet.vehicledevice.twin

import kotlin.test.Test
import kotlin.test.assertEquals

class DeviceTwinDriftDetectorTest {
    private val detector = DeviceTwinDriftDetector()

    @Test
    fun `returns unknown when desired or reported state is missing`() {
        val result = detector.detect(desired = null, reported = null)

        assertEquals(DriftStatus.UNKNOWN, result.status)
        assertEquals("desired_or_reported_missing", result.reason)
    }

    @Test
    fun `returns in sync when desired and reported match`() {
        val desired = DeviceTwinState(true, true, true, true)
        val reported = DeviceTwinState(true, true, true, true)

        val result = detector.detect(desired, reported)

        assertEquals(DriftStatus.IN_SYNC, result.status)
        assertEquals(0, result.mismatchedFields.size)
    }

    @Test
    fun `returns drifted with explicit field mismatches`() {
        val desired = DeviceTwinState(true, true, true, true)
        val reported = DeviceTwinState(false, true, false, true)

        val result = detector.detect(desired, reported)

        assertEquals(DriftStatus.DRIFTED, result.status)
        assertEquals("state_mismatch", result.reason)
        assertEquals(listOf("cameraEnabled", "canBusEnabled"), result.mismatchedFields)
    }
}
