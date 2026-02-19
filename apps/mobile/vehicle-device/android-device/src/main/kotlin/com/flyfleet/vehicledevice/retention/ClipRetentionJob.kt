package com.flyfleet.vehicledevice.retention

data class ClipRetentionTaskResult(
    val clipId: String,
    val allowed: Boolean,
    val reason: String,
)

class ClipRetentionJob(private val service: ClipRetentionService) {
    fun run(metadata: List<ClipMetadata>): List<ClipRetentionTaskResult> {
        return metadata.map { clip ->
            val action = service.evaluateClip(clip)
            ClipRetentionTaskResult(
                clipId = clip.id,
                allowed = action.canDelete,
                reason = action.reason,
            )
        }
    }
}
