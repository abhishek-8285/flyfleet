package com.flyfleet.vehicledevice.retention

fun interface ClipMetadataProvider {
    fun fetchClips(): List<ClipMetadata>
}

class ClipRetentionRunner(
    private val job: ClipRetentionJob,
    private val provider: ClipMetadataProvider,
) {
    fun run(): List<ClipRetentionTaskResult> {
        val clips = provider.fetchClips()
        return job.run(clips)
    }
}
