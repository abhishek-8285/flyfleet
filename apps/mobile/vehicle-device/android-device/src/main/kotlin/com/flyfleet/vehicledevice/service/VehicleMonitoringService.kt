package com.flyfleet.vehicledevice.service

import com.flyfleet.vehicledevice.config.FeatureFlags
import com.flyfleet.vehicledevice.runtime.RuntimeSignal
import com.flyfleet.vehicledevice.runtime.RuntimeState
import com.flyfleet.vehicledevice.runtime.RuntimeStateMachine
import com.flyfleet.vehicledevice.policy.MonitoringSafetyPolicy
import kotlinx.coroutines.flow.StateFlow

/**
 * Service orchestration scaffold for vehicle monitoring lifecycle.
 * Android-specific foreground service wiring is intentionally separated from
 * this core coordinator so it can be exercised in local/unit tests.
 */
class VehicleMonitoringService(
    private val runtimeStateMachine: RuntimeStateMachine = RuntimeStateMachine(),
    private val safetyPolicy: MonitoringSafetyPolicy = MonitoringSafetyPolicy(),
) {
    private var featureFlags: FeatureFlags = FeatureFlags()

    fun runtimeState(): StateFlow<RuntimeState> = runtimeStateMachine.currentState()

    fun updateFeatureFlags(flags: FeatureFlags) {
        featureFlags = flags
    }

    fun canCaptureCamera(): Boolean = safetyPolicy.canCaptureCamera(featureFlags, runtimeStateMachine.currentState().value)

    fun canRunAi(): Boolean = safetyPolicy.canRunAi(featureFlags, runtimeStateMachine.currentState().value)

    fun canUseCanBus(): Boolean = safetyPolicy.canUseCanBus(featureFlags)

    fun isUiInteractionLocked(speedMph: Double): Boolean =
        safetyPolicy.isUiInteractionLocked(featureFlags, speedMph)

    fun isAudioEventCheckAllowed(): Boolean =
        safetyPolicy.isAudioEventCheckAllowed(featureFlags, runtimeStateMachine.currentState().value)

    fun onBootCompleted(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.STARTUP_COMPLETE)

    fun onIgnitionOn(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.IGNITION_ON)

    fun onIgnitionOff(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.IGNITION_OFF)

    fun onDepotEntered(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.DEPOT_ENTERED)

    fun onDepotExited(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.DEPOT_EXITED)

    fun onSensorFailure(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.SENSOR_FAILURE)

    fun onRecoveryReady(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.COMPONENT_RECOVERED)

    fun onShutdownRequested(): RuntimeState = runtimeStateMachine.transition(RuntimeSignal.SHUTDOWN_REQUESTED)
}
