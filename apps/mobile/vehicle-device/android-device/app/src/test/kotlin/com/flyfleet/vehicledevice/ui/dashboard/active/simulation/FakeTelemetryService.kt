package com.flyfleet.vehicledevice.ui.dashboard.active.simulation

import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

/**
 * Fake Telemetry Service used for automated E2E device/hardware simulation. In a real environment,
 * this binds to the OBD-II dongle and vehicle CAN bus.
 */
class FakeTelemetryService {

    data class TelemetryEmission(
            val speedKmh: Double,
            val runTimeSeconds: Long,
            val rpm: Int,
            val obdConnected: Boolean,
            val ignitionOn: Boolean
    )

    /** Simulates a driving scenario: starts at 0, accelerates to 100km/h, and drops connection. */
    fun simulateDriveSession(): Flow<TelemetryEmission> = flow {
        var currentSpeed = 0.0
        var runtime = 0L

        // Idle state
        emit(TelemetryEmission(0.0, runtime, 800, true, true))
        delay(1000)

        // Acceleration
        while (currentSpeed < 60.0) {
            currentSpeed += 5.5
            runtime += 1
            emit(TelemetryEmission(currentSpeed, runtime, 1500, true, true))
            delay(1000)
        }

        // Sudden OBD disconnect logic
        emit(TelemetryEmission(currentSpeed, runtime + 1, 0, false, true))
        delay(1000)

        // Ignition off
        emit(TelemetryEmission(0.0, runtime + 2, 0, false, false))
    }
}
