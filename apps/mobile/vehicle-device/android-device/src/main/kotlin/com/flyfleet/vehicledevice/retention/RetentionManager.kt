package com.flyfleet.vehicledevice.retention

data class ClipMetadata(
    val id: String,
    val ageMs: Long,
    val isCritical: Boolean,
)

data class RetentionAction(
    val canDelete: Boolean,
    val reason: String,
)

class RetentionManager(
    private val policy: RetentionPolicy,
) {
    fun evaluateClip(metadata: ClipMetadata): RetentionAction {
        if (metadata.isCritical) {
            return RetentionAction(false, "critical_evidence")
        }
        val decision = policy.shouldOverwriteNormalData(metadata.ageMs)
        return RetentionAction(decision.allowOverwrite, decision.reason)
    }
}
