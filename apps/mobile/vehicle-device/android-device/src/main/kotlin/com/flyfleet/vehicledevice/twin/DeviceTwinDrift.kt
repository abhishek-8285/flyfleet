package com.flyfleet.vehicledevice.twin

enum class DriftStatus {
    IN_SYNC,
    DRIFTED,
    UNKNOWN,
}

data class DeviceTwinState(
    val cameraEnabled: Boolean?,
    val aiEnabled: Boolean?,
    val canBusEnabled: Boolean?,
    val lockoutEnabled: Boolean?,
)

data class DeviceTwinDriftResult(
    val status: DriftStatus,
    val reason: String,
    val mismatchedFields: List<String>,
)

class DeviceTwinDriftDetector {
    fun detect(desired: DeviceTwinState?, reported: DeviceTwinState?): DeviceTwinDriftResult {
        if (desired == null || reported == null) {
            return DeviceTwinDriftResult(
                status = DriftStatus.UNKNOWN,
                reason = "desired_or_reported_missing",
                mismatchedFields = emptyList(),
            )
        }

        val mismatches = mutableListOf<String>()

        if (desired.cameraEnabled != reported.cameraEnabled) {
            mismatches += "cameraEnabled"
        }
        if (desired.aiEnabled != reported.aiEnabled) {
            mismatches += "aiEnabled"
        }
        if (desired.canBusEnabled != reported.canBusEnabled) {
            mismatches += "canBusEnabled"
        }
        if (desired.lockoutEnabled != reported.lockoutEnabled) {
            mismatches += "lockoutEnabled"
        }

        if (mismatches.isEmpty()) {
            return DeviceTwinDriftResult(
                status = DriftStatus.IN_SYNC,
                reason = "in_sync",
                mismatchedFields = emptyList(),
            )
        }

        return DeviceTwinDriftResult(
            status = DriftStatus.DRIFTED,
            reason = "state_mismatch",
            mismatchedFields = mismatches,
        )
    }
}
