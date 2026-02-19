package com.flyfleet.vehicledevice.runtime

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class RuntimeStateMachine(
    initialState: RuntimeState = RuntimeState.BOOTING
) {
    private val state = MutableStateFlow(initialState)

    fun currentState(): StateFlow<RuntimeState> = state.asStateFlow()

    fun transition(signal: RuntimeSignal): RuntimeState {
        val next = when (state.value) {
            RuntimeState.BOOTING -> when (signal) {
                RuntimeSignal.STARTUP_COMPLETE -> RuntimeState.IDLE
                RuntimeSignal.SENSOR_FAILURE -> RuntimeState.DEGRADED
                RuntimeSignal.SHUTDOWN_REQUESTED -> RuntimeState.SAFE_SHUTDOWN
                else -> state.value
            }

            RuntimeState.IDLE -> when (signal) {
                RuntimeSignal.IGNITION_ON -> RuntimeState.ACTIVE_DRIVE
                RuntimeSignal.DEPOT_ENTERED -> RuntimeState.DEPOT_MODE
                RuntimeSignal.SENSOR_FAILURE -> RuntimeState.DEGRADED
                RuntimeSignal.SHUTDOWN_REQUESTED -> RuntimeState.SAFE_SHUTDOWN
                else -> state.value
            }

            RuntimeState.ACTIVE_DRIVE -> when (signal) {
                RuntimeSignal.IGNITION_OFF -> RuntimeState.IDLE
                RuntimeSignal.DEPOT_ENTERED -> RuntimeState.DEPOT_MODE
                RuntimeSignal.SENSOR_FAILURE -> RuntimeState.DEGRADED
                RuntimeSignal.SHUTDOWN_REQUESTED -> RuntimeState.SAFE_SHUTDOWN
                else -> state.value
            }

            RuntimeState.DEPOT_MODE -> when (signal) {
                RuntimeSignal.DEPOT_EXITED -> RuntimeState.IDLE
                RuntimeSignal.IGNITION_ON -> RuntimeState.ACTIVE_DRIVE
                RuntimeSignal.SENSOR_FAILURE -> RuntimeState.DEGRADED
                RuntimeSignal.SHUTDOWN_REQUESTED -> RuntimeState.SAFE_SHUTDOWN
                else -> state.value
            }

            RuntimeState.DEGRADED -> when (signal) {
                RuntimeSignal.COMPONENT_RECOVERED -> RuntimeState.RECOVERY
                RuntimeSignal.SHUTDOWN_REQUESTED -> RuntimeState.SAFE_SHUTDOWN
                else -> state.value
            }

            RuntimeState.RECOVERY -> when (signal) {
                RuntimeSignal.STARTUP_COMPLETE -> RuntimeState.IDLE
                RuntimeSignal.RECOVERY_TIMEOUT -> RuntimeState.DEGRADED
                RuntimeSignal.SHUTDOWN_REQUESTED -> RuntimeState.SAFE_SHUTDOWN
                else -> state.value
            }

            RuntimeState.SAFE_SHUTDOWN -> RuntimeState.SAFE_SHUTDOWN
        }

        state.value = next
        return next
    }
}
