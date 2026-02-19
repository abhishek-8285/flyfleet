package com.flyfleet.vehicledevice.retention

class ClipRetentionService(
    private val policy: RetentionPolicy,
) {
    fun evaluateClip(metadata: ClipMetadata): RetentionAction {
        return RetentionManager(policy).evaluateClip(metadata)
    }
}
