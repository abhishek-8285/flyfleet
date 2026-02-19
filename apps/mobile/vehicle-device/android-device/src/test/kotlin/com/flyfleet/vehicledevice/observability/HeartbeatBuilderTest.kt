package com.flyfleet.vehicledevice.observability

import com.flyfleet.vehicledevice.runtime.RuntimeState
import kotlin.test.Test
import kotlin.test.assertEquals

class HeartbeatBuilderTest {
    private val builder = HeartbeatBuilder()

    @Test
    fun `overall status is healthy when all modules are healthy`() {
        val snapshot = builder.build(
            deviceId = "dev-1",
            runtimeState = RuntimeState.ACTIVE_DRIVE,
            timestampMs = 1_000,
            queueDepth = 2,
            deadLetterCount = 0,
            healthMatrix = HealthMatrix(
                camera = ModuleHealth.HEALTHY,
                gps = ModuleHealth.HEALTHY,
                canBus = ModuleHealth.HEALTHY,
                network = ModuleHealth.HEALTHY,
                ai = ModuleHealth.HEALTHY,
                storage = ModuleHealth.HEALTHY,
            ),
        )

        assertEquals(ModuleHealth.HEALTHY, snapshot.overallStatus)
    }

    @Test
    fun `overall status is degraded when any module is degraded and none offline`() {
        val snapshot = builder.build(
            deviceId = "dev-1",
            runtimeState = RuntimeState.DEGRADED,
            timestampMs = 1_000,
            queueDepth = 3,
            deadLetterCount = 0,
            healthMatrix = HealthMatrix(
                camera = ModuleHealth.HEALTHY,
                gps = ModuleHealth.DEGRADED,
                canBus = ModuleHealth.HEALTHY,
                network = ModuleHealth.HEALTHY,
                ai = ModuleHealth.HEALTHY,
                storage = ModuleHealth.HEALTHY,
            ),
        )

        assertEquals(ModuleHealth.DEGRADED, snapshot.overallStatus)
    }

    @Test
    fun `overall status is offline when any module is offline`() {
        val snapshot = builder.build(
            deviceId = "dev-1",
            runtimeState = RuntimeState.DEGRADED,
            timestampMs = 1_000,
            queueDepth = 7,
            deadLetterCount = 1,
            healthMatrix = HealthMatrix(
                camera = ModuleHealth.HEALTHY,
                gps = ModuleHealth.HEALTHY,
                canBus = ModuleHealth.OFFLINE,
                network = ModuleHealth.HEALTHY,
                ai = ModuleHealth.DEGRADED,
                storage = ModuleHealth.HEALTHY,
            ),
        )

        assertEquals(ModuleHealth.OFFLINE, snapshot.overallStatus)
    }
}
