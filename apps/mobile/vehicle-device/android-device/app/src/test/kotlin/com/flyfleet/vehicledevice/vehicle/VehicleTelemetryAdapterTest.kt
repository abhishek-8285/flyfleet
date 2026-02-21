package com.flyfleet.vehicledevice.vehicle

import kotlin.test.Test
import kotlin.test.assertEquals

class VehicleTelemetryAdapterTest {
    @Test
    fun `direct can adapter reports high confidence snapshot`() {
        val adapter = DirectCanTelemetryAdapter()

        val snapshot = adapter.readSnapshot()

        assertEquals(VehicleIntegrationMode.DIRECT_CAN, snapshot.mode)
        assertEquals(TelemetryConfidence.HIGH, snapshot.confidence)
    }

    @Test
    fun `fallback scorer returns medium when gps imu and speed consistency are strong`() {
        val scorer = FallbackConfidenceScorer()
        val (confidence, reason) = scorer.score(
            FallbackInputs(
                canAvailable = false,
                gpsSignalQuality = 0.9,
                imuAvailable = true,
                speedConsistency = 0.9,
            ),
        )

        assertEquals(TelemetryConfidence.MEDIUM, confidence)
        assertEquals("gps_imu_fallback_medium_confidence", reason)
    }

    @Test
    fun `fallback scorer returns low when fallback signals are weak`() {
        val scorer = FallbackConfidenceScorer()
        val (confidence, reason) = scorer.score(
            FallbackInputs(
                canAvailable = false,
                gpsSignalQuality = 0.2,
                imuAvailable = false,
                speedConsistency = 0.3,
            ),
        )

        assertEquals(TelemetryConfidence.LOW, confidence)
        assertEquals("gps_imu_fallback_low_confidence", reason)
    }

    @Test
    fun `obd adapter falls back to gps imu mode when can is unavailable`() {
        val adapter = ObdTelemetryAdapter(
            inputsProvider = {
                FallbackInputs(
                    canAvailable = false,
                    gpsSignalQuality = 0.8,
                    imuAvailable = true,
                    speedConsistency = 0.8,
                )
            },
        )

        val snapshot = adapter.readSnapshot()
        assertEquals(VehicleIntegrationMode.GPS_IMU_FALLBACK, snapshot.mode)
        assertEquals(TelemetryConfidence.MEDIUM, snapshot.confidence)
    }
}
