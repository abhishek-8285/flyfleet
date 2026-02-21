package com.flyfleet.vehicledevice.service

import com.flyfleet.vehicledevice.config.FeatureFlags
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class VehicleMonitoringServicePolicyTest {
    @Test
    fun `service exposes policy decisions based on runtime state and flags`() {
        val service = VehicleMonitoringService()

        service.onBootCompleted()
        service.updateFeatureFlags(
            FeatureFlags(
                cameraEnabled = true,
                aiEnabled = true,
                canBusEnabled = false,
                driverLockoutEnabled = true,
                audioEventChecksEnabled = true,
            ),
        )

        assertFalse(service.canCaptureCamera())
        assertFalse(service.canRunAi())
        assertFalse(service.isAudioEventCheckAllowed())

        service.onIgnitionOn()

        assertTrue(service.canCaptureCamera())
        assertTrue(service.canRunAi())
        assertFalse(service.canUseCanBus())
        assertTrue(service.isUiInteractionLocked(12.0))
        assertTrue(service.isAudioEventCheckAllowed())
    }
}
