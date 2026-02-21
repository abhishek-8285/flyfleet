package com.flyfleet.vehicledevice.retention

import kotlin.io.path.createTempFile
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.time.Duration.Companion.days

class RetentionSupervisorTest {
    @Test
    fun `reloads config and reuses dynamic policy`() {
        val sensor = createTempFile()
        sensor.toFile().writeText("{\"retention\":{\"criticalRetentionDays\":0,\"evidenceReserveMb\":0,\"legal_hold\":false}}")

        val provider = RetentionConfigProvider(sensor)
        val metadataProvider = ClipMetadataProvider {
            listOf(ClipMetadata(id = "clip-sup-1", ageMs = 10.days.inWholeMilliseconds, isCritical = false))
        }
        val supervisor = RetentionSupervisor(
            configProvider = provider,
            metadataProvider = metadataProvider,
            storageStateProvider = { StorageState(1000, 1000) },
            clockProvider = { 0L },
        )

        val results = supervisor.run()
        assertEquals(1, results.size)
        assertEquals("clip-sup-1", results[0].clipId)
        assertEquals(true, results[0].allowed)
    }
}
