package com.flyfleet.vehicledevice.observability

import com.flyfleet.vehicledevice.runtime.RuntimeState
import com.flyfleet.vehicledevice.twin.DeviceTwinDriftResult
import com.flyfleet.vehicledevice.twin.DriftStatus
import com.flyfleet.vehicledevice.vehicle.VehicleIntegrationMode
import com.flyfleet.vehicledevice.vehicle.VehicleTelemetrySnapshot
import com.flyfleet.vehicledevice.retention.RetentionMetricsSnapshot

enum class ModuleHealth {
    HEALTHY,
    DEGRADED,
    OFFLINE,
}

data class HealthMatrix(
    val camera: ModuleHealth,
    val gps: ModuleHealth,
    val canBus: ModuleHealth,
    val network: ModuleHealth,
    val ai: ModuleHealth,
    val storage: ModuleHealth,
)

data class HeartbeatSnapshot(
    val deviceId: String,
    val runtimeState: RuntimeState,
    val timestampMs: Long,
    val queueDepth: Int,
    val deadLetterCount: Int,
    val healthMatrix: HealthMatrix,
    val overallStatus: ModuleHealth,
    val vehicleIntegrationMode: VehicleIntegrationMode?,
    val telemetryConfidence: String?,
    val telemetryConfidenceReason: String?,
    val twinDriftStatus: DriftStatus?,
    val twinDriftReason: String?,
    val twinMismatchedFieldsCount: Int,
    val retentionAllowed: Int?,
    val retentionBlocked: Int?,
)

class HeartbeatBuilder {
    fun build(
        deviceId: String,
        runtimeState: RuntimeState,
        timestampMs: Long,
        queueDepth: Int,
        deadLetterCount: Int,
        healthMatrix: HealthMatrix,
        vehicleTelemetry: VehicleTelemetrySnapshot? = null,
        twinDrift: DeviceTwinDriftResult? = null,
        retentionMetrics: RetentionMetricsSnapshot? = null,
    ): HeartbeatSnapshot {
        val overall = deriveOverall(healthMatrix)
        return HeartbeatSnapshot(
            deviceId = deviceId,
            runtimeState = runtimeState,
            timestampMs = timestampMs,
            queueDepth = queueDepth,
            deadLetterCount = deadLetterCount,
            healthMatrix = healthMatrix,
            overallStatus = overall,
            vehicleIntegrationMode = vehicleTelemetry?.mode,
            telemetryConfidence = vehicleTelemetry?.confidence?.name,
            telemetryConfidenceReason = vehicleTelemetry?.confidenceReason,
            twinDriftStatus = twinDrift?.status,
            twinDriftReason = twinDrift?.reason,
            twinMismatchedFieldsCount = twinDrift?.mismatchedFields?.size ?: 0,
            retentionAllowed = retentionMetrics?.allowedDeletes,
            retentionBlocked = retentionMetrics?.blockedDeletes,
        )
    }

    private fun deriveOverall(matrix: HealthMatrix): ModuleHealth {
        val all = listOf(
            matrix.camera,
            matrix.gps,
            matrix.canBus,
            matrix.network,
            matrix.ai,
            matrix.storage,
        )
        return when {
            all.any { it == ModuleHealth.OFFLINE } -> ModuleHealth.OFFLINE
            all.any { it == ModuleHealth.DEGRADED } -> ModuleHealth.DEGRADED
            else -> ModuleHealth.HEALTHY
        }
    }
}
