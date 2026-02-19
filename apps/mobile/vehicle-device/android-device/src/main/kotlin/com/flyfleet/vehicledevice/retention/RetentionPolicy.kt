package com.flyfleet.vehicledevice.retention

import kotlin.time.Duration.Companion.days

data class RetentionConfig(
    val criticalRetentionDays: Int = 30,
    val evidenceReserveMb: Long = 5000,
    val legalHold: Boolean = false,
)

data class StorageState(
    val usedMb: Long,
    val availableMb: Long,
)

data class RetentionDecision(
    val allowOverwrite: Boolean,
    val reason: String,
)

class RetentionPolicy(
    private val config: RetentionConfig,
    private val storageStateProvider: () -> StorageState,
    private val clockMsProvider: () -> Long,
) {
    fun shouldOverwriteNormalData(ageMs: Long): RetentionDecision {
        val now = clockMsProvider()
        if (config.legalHold) {
            return RetentionDecision(false, "legal_hold_active")
        }
        if (ageMs < config.criticalRetentionDays.days.inWholeMilliseconds) {
            return RetentionDecision(false, "within_critical_retention")
        }
        val storage = storageStateProvider()
        if (storage.availableMb < config.evidenceReserveMb) {
            return RetentionDecision(false, "reserve_threshold")
        }
        return RetentionDecision(true, "old_data_overwrite_allowed")
    }
}
