package com.flyfleet.vehicledevice.policy

import com.flyfleet.vehicledevice.config.FeatureFlags
import com.flyfleet.vehicledevice.runtime.RuntimeState

/**
 * Safety policy for runtime capabilities, built to be deterministic and testable.
 */
class MonitoringSafetyPolicy(
    private val lockoutSpeedMph: Double = 5.0,
) {
    fun canCaptureCamera(flags: FeatureFlags, state: RuntimeState): Boolean {
        if (!flags.cameraEnabled) {
            return false
        }
        return state == RuntimeState.ACTIVE_DRIVE || state == RuntimeState.DEPOT_MODE
    }

    fun canRunAi(flags: FeatureFlags, state: RuntimeState): Boolean {
        if (!flags.aiEnabled) {
            return false
        }
        return state == RuntimeState.ACTIVE_DRIVE || state == RuntimeState.DEPOT_MODE
    }

    fun canUseCanBus(flags: FeatureFlags): Boolean = flags.canBusEnabled

    fun isUiInteractionLocked(flags: FeatureFlags, speedMph: Double): Boolean {
        if (!flags.driverLockoutEnabled) {
            return false
        }
        return speedMph > lockoutSpeedMph
    }

    fun isAudioEventCheckAllowed(flags: FeatureFlags, state: RuntimeState): Boolean {
        if (!flags.audioEventChecksEnabled) {
            return false
        }
        return state == RuntimeState.ACTIVE_DRIVE
    }
}
