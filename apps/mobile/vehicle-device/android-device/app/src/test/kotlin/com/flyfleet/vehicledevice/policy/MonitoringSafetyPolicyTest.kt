package com.flyfleet.vehicledevice.policy

import com.flyfleet.vehicledevice.config.FeatureFlags
import com.flyfleet.vehicledevice.runtime.RuntimeState
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class MonitoringSafetyPolicyTest {
    private val policy = MonitoringSafetyPolicy(lockoutSpeedMph = 5.0)

    @Test
    fun `camera allowed only in active and depot when camera flag enabled`() {
        val flags = FeatureFlags(cameraEnabled = true)

        assertTrue(policy.canCaptureCamera(flags, RuntimeState.ACTIVE_DRIVE))
        assertTrue(policy.canCaptureCamera(flags, RuntimeState.DEPOT_MODE))
        assertFalse(policy.canCaptureCamera(flags, RuntimeState.IDLE))
        assertFalse(policy.canCaptureCamera(flags, RuntimeState.DEGRADED))
    }

    @Test
    fun `camera blocked when camera feature disabled`() {
        val flags = FeatureFlags(cameraEnabled = false)
        assertFalse(policy.canCaptureCamera(flags, RuntimeState.ACTIVE_DRIVE))
    }

    @Test
    fun `driver lockout enabled only when speed is above threshold`() {
        val flags = FeatureFlags(driverLockoutEnabled = true)
        assertFalse(policy.isUiInteractionLocked(flags, speedMph = 5.0))
        assertTrue(policy.isUiInteractionLocked(flags, speedMph = 5.1))
    }

    @Test
    fun `driver lockout bypassed when flag disabled`() {
        val flags = FeatureFlags(driverLockoutEnabled = false)
        assertFalse(policy.isUiInteractionLocked(flags, speedMph = 120.0))
    }

    @Test
    fun `audio checks only allowed in active drive and when enabled`() {
        val enabled = FeatureFlags(audioEventChecksEnabled = true)
        val disabled = FeatureFlags(audioEventChecksEnabled = false)

        assertTrue(policy.isAudioEventCheckAllowed(enabled, RuntimeState.ACTIVE_DRIVE))
        assertFalse(policy.isAudioEventCheckAllowed(enabled, RuntimeState.DEPOT_MODE))
        assertFalse(policy.isAudioEventCheckAllowed(disabled, RuntimeState.ACTIVE_DRIVE))
    }
}
