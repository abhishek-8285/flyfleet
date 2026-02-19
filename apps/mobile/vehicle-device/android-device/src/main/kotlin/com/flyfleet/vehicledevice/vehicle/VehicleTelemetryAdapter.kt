package com.flyfleet.vehicledevice.vehicle

enum class VehicleIntegrationMode {
    DIRECT_CAN,
    OBD_ADAPTER,
    GPS_IMU_FALLBACK,
}

enum class TelemetryConfidence {
    HIGH,
    MEDIUM,
    LOW,
}

data class VehicleTelemetrySnapshot(
    val mode: VehicleIntegrationMode,
    val engineRpm: Int? = null,
    val speedKph: Double? = null,
    val brakePressed: Boolean? = null,
    val turnSignalOn: Boolean? = null,
    val vin: String? = null,
    val dtcCount: Int? = null,
    val confidence: TelemetryConfidence,
    val confidenceReason: String,
)

data class FallbackInputs(
    val canAvailable: Boolean,
    val gpsSignalQuality: Double,
    val imuAvailable: Boolean,
    val speedConsistency: Double,
)

interface VehicleTelemetryAdapter {
    fun readSnapshot(): VehicleTelemetrySnapshot
}

class FallbackConfidenceScorer {
    fun score(inputs: FallbackInputs): Pair<TelemetryConfidence, String> {
        if (inputs.canAvailable) {
            return TelemetryConfidence.HIGH to "can_available"
        }

        var score = 0
        if (inputs.gpsSignalQuality >= 0.8) {
            score += 2
        } else if (inputs.gpsSignalQuality >= 0.5) {
            score += 1
        }

        if (inputs.imuAvailable) {
            score += 1
        }

        if (inputs.speedConsistency >= 0.8) {
            score += 2
        } else if (inputs.speedConsistency >= 0.5) {
            score += 1
        }

        return when {
            score >= 4 -> TelemetryConfidence.MEDIUM to "gps_imu_fallback_medium_confidence"
            else -> TelemetryConfidence.LOW to "gps_imu_fallback_low_confidence"
        }
    }
}

class DirectCanTelemetryAdapter : VehicleTelemetryAdapter {
    override fun readSnapshot(): VehicleTelemetrySnapshot {
        // Placeholder deterministic sample until real CAN interface wiring.
        return VehicleTelemetrySnapshot(
            mode = VehicleIntegrationMode.DIRECT_CAN,
            engineRpm = 1200,
            speedKph = 36.0,
            brakePressed = false,
            turnSignalOn = false,
            vin = "UNKNOWN",
            dtcCount = 0,
            confidence = TelemetryConfidence.HIGH,
            confidenceReason = "can_available",
        )
    }
}

class ObdTelemetryAdapter(
    private val scorer: FallbackConfidenceScorer = FallbackConfidenceScorer(),
    private val inputsProvider: () -> FallbackInputs,
) : VehicleTelemetryAdapter {
    override fun readSnapshot(): VehicleTelemetrySnapshot {
        val inputs = inputsProvider()
        val (confidence, reason) = scorer.score(inputs)
        return VehicleTelemetrySnapshot(
            mode = if (inputs.canAvailable) VehicleIntegrationMode.OBD_ADAPTER else VehicleIntegrationMode.GPS_IMU_FALLBACK,
            engineRpm = null,
            speedKph = null,
            brakePressed = null,
            turnSignalOn = null,
            vin = null,
            dtcCount = null,
            confidence = confidence,
            confidenceReason = reason,
        )
    }
}
