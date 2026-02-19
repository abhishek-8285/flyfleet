package com.flyfleet.vehicledevice.retention

class RetentionSupervisor(
    private val configProvider: RetentionConfigProvider,
    private val metadataProvider: ClipMetadataProvider,
    private val storageStateProvider: () -> StorageState,
    private val clockProvider: () -> Long,
) {
    fun run(): List<ClipRetentionTaskResult> {
        val config = configProvider.load()
        val policy = RetentionPolicy(config, storageStateProvider, clockProvider)
        val runner = ClipRetentionRunner(ClipRetentionJob(ClipRetentionService(policy)), metadataProvider)
        return runner.run()
    }
}
