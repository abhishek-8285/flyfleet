package com.flyfleet.vehicledevice.retention

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.io.path.createTempFile
import kotlinx.serialization.json.Json

class RetentionConfigProviderTest {
    @Test
    fun `loads retention config from json file`() {
        val temp = createTempFile().apply {
            writeText("""
                {"retention":{"criticalRetentionDays":10,"evidenceReserveMb":2048,"legalHold":true}}
            """)
        }

        val provider = RetentionConfigProvider(temp)
        val config = provider.load()

        assertEquals(10, config.criticalRetentionDays)
        assertEquals(2048, config.evidenceReserveMb)
        assertEquals(true, config.legalHold)
    }
}
