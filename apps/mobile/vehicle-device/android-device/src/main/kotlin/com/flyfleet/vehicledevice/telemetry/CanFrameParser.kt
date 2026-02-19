package com.flyfleet.vehicledevice.telemetry

import timber.log.Timber

/**
 * Partial telemetry data extracted from a single CAN frame
 */
data class PartialTelemetry(
    val speedKmh: Double? = null,
    val engineRpm: Double? = null,
    val brakePedalPressed: Boolean? = null
)

/**
 * Parses raw CAN bus frames (J1939 Standard)
 */
object CanFrameParser {

    // J1939 PGNs
    private const val PGN_ENGINE_SPEED = 0xF004 // EEC1
    private const val PGN_VEHICLE_SPEED = 0xFEF1 // CCVS

    fun parse(canId: Int, data: ByteArray): PartialTelemetry? {
        try {
            // Extract PGN from 29-bit identifier (J1939)
            // (ID >> 8) & 0xFFFF is simplified; robust implementations check PDU format
            val pgn = (canId shr 8) and 0xFFFF

            return when (pgn) {
                PGN_ENGINE_SPEED -> parseEngineSpeed(data)
                PGN_VEHICLE_SPEED -> parseVehicleSpeed(data)
                else -> null
            }
        } catch (e: Exception) {
            Timber.e(e, "Failed to parse CAN frame")
            return null
        }
    }

    private fun parseEngineSpeed(data: ByteArray): PartialTelemetry {
        // Bytes 3,4 are Engine Speed (RPM/0.125)
        if (data.size < 5) return PartialTelemetry()
        val rpmRaw = ((data[4].toInt() and 0xFF) shl 8) or (data[3].toInt() and 0xFF)
        val rpm = rpmRaw * 0.125
        return PartialTelemetry(engineRpm = rpm)
    }

    private fun parseVehicleSpeed(data: ByteArray): PartialTelemetry {
       // Bytes 1,2 are Wheel Based Speed (km/h * 256)
       if (data.size < 3) return PartialTelemetry()
       val speedRaw = ((data[2].toInt() and 0xFF) shl 8) or (data[1].toInt() and 0xFF)
       val speed = speedRaw / 256.0
       
       // Byte 3 bit 4 is Brake Switch (0x10)
       val brakePressed = if (data.size > 3) (data[3].toInt() and 0x10) != 0 else null
       
       return PartialTelemetry(speedKmh = speed, brakePedalPressed = brakePressed)
    }
}
