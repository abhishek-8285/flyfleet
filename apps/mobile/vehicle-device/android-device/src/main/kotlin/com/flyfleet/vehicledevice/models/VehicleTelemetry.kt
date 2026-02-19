package com.flyfleet.vehicledevice.models

/**
 * Standardized vehicle telemetry data
 */
data class VehicleTelemetry(
    val timestamp: Long,
    val speedKmh: Double,
    val engineRpm: Double,
    val fuelLevelPercent: Double? = null,
    val odometerKm: Double? = null,
    val brakePedalPressed: Boolean = false,
    val engineCoolantTempC: Double? = null,
    val source: Source = Source.CAN_BUS
) {
    enum class Source {
        CAN_BUS,
        OBD_II,
        GPS_ESTIMATE,
        SIMULATION
    }
}
