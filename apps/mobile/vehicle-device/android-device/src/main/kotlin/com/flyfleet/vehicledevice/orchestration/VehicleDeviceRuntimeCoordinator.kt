package com.flyfleet.vehicledevice.orchestration

import com.flyfleet.vehicledevice.audit.AuditEmitter
import com.flyfleet.vehicledevice.audit.AuditEnvelope
import com.flyfleet.vehicledevice.audit.NoopAuditEmitter
import com.flyfleet.vehicledevice.command.CommandEnvelopeGuard
import com.flyfleet.vehicledevice.command.CommandContext
import com.flyfleet.vehicledevice.command.CommandDecision
import com.flyfleet.vehicledevice.command.DeviceCommandLifecycle
import com.flyfleet.vehicledevice.command.DeviceCommandType
import com.flyfleet.vehicledevice.command.SignedCommandEnvelope
import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.forensics.ChainedEvidenceRecord
import com.flyfleet.vehicledevice.forensics.IncidentBundle
import com.flyfleet.vehicledevice.forensics.IncidentBundleBuilder
import com.flyfleet.vehicledevice.observability.HealthMatrix
import com.flyfleet.vehicledevice.observability.HeartbeatBuilder
import com.flyfleet.vehicledevice.observability.HeartbeatSnapshot
import com.flyfleet.vehicledevice.service.VehicleMonitoringService
import com.flyfleet.vehicledevice.sync.Clock
import com.flyfleet.vehicledevice.sync.OfflineUploadQueue
import com.flyfleet.vehicledevice.sync.QueueItemStatus
import com.flyfleet.vehicledevice.sync.StoredEvent
import com.flyfleet.vehicledevice.sync.SyncPipelineOrchestrator
import com.flyfleet.vehicledevice.sync.SyncStepOutcome
import com.flyfleet.vehicledevice.twin.DeviceTwinDriftResult
import com.flyfleet.vehicledevice.twin.DriftStatus
import com.flyfleet.vehicledevice.vehicle.TelemetryConfidence
import com.flyfleet.vehicledevice.vehicle.VehicleTelemetrySnapshot
import com.flyfleet.vehicledevice.retention.RetentionMetricsSnapshot

class VehicleDeviceRuntimeCoordinator(
    private val deviceId: String,
    private val clock: Clock,
    private val monitoringService: VehicleMonitoringService,
    private val syncOrchestrator: SyncPipelineOrchestrator,
    private val uploadQueue: OfflineUploadQueue,
    private val commandLifecycle: DeviceCommandLifecycle,
    private val heartbeatBuilder: HeartbeatBuilder = HeartbeatBuilder(),
    private val auditEmitter: AuditEmitter = NoopAuditEmitter,
    private val commandEnvelopeGuard: CommandEnvelopeGuard? = null,
    private val incidentBundleBuilder: IncidentBundleBuilder? = null,
    private val auditSnapshotProvider: () -> List<AuditEnvelope> = { emptyList() },
) {
    fun ingestEvent(eventType: DeviceEventType, payloadRef: String): StoredEvent {
        return syncOrchestrator.ingestEvent(eventType, payloadRef)
    }

    fun processOneUpload(): SyncStepOutcome {
        return syncOrchestrator.processNextUpload()
    }

    fun evaluateCommand(commandType: DeviceCommandType, speedMph: Double): CommandDecision {
        val decision = commandLifecycle.evaluateApply(
            commandType,
            CommandContext(
                runtimeState = monitoringService.runtimeState().value,
                speedMph = speedMph,
            ),
        )
        auditEmitter.recordCommandDecision(commandType, speedMph, decision)
        return decision
    }

    fun evaluateSignedCommand(envelope: SignedCommandEnvelope, speedMph: Double): CommandDecision {
        val guard = commandEnvelopeGuard
        if (guard != null) {
            val envelopeDecision = guard.verify(envelope)
            if (!envelopeDecision.allowed) {
                val rejected = CommandDecision(allowed = false, reason = envelopeDecision.reason)
                auditEmitter.recordCommandDecision(envelope.commandType, speedMph, rejected)
                return rejected
            }
        }
        return evaluateCommand(envelope.commandType, speedMph)
    }

    fun buildHeartbeat(healthMatrix: HealthMatrix): HeartbeatSnapshot {
        return buildHeartbeat(healthMatrix, vehicleTelemetry = null, twinDrift = null, retentionMetrics = null)
    }

    fun buildHeartbeat(
        healthMatrix: HealthMatrix,
        vehicleTelemetry: VehicleTelemetrySnapshot?,
        twinDrift: DeviceTwinDriftResult?,
        retentionMetrics: RetentionMetricsSnapshot?,
    ): HeartbeatSnapshot {
        val snapshot = uploadQueue.snapshot()
        val deadLetters = snapshot.count { it.status == QueueItemStatus.DEAD_LETTER }
        return heartbeatBuilder.build(
            deviceId = deviceId,
            runtimeState = monitoringService.runtimeState().value,
            timestampMs = clock.nowMs(),
            queueDepth = snapshot.size,
            deadLetterCount = deadLetters,
            healthMatrix = healthMatrix,
            vehicleTelemetry = vehicleTelemetry,
            twinDrift = twinDrift,
            retentionMetrics = retentionMetrics,
        )
    }

    fun emitOperationalSignals(
        vehicleTelemetry: VehicleTelemetrySnapshot?,
        twinDrift: DeviceTwinDriftResult?,
    ): List<StoredEvent> {
        val emitted = mutableListOf<StoredEvent>()
        if (vehicleTelemetry != null && vehicleTelemetry.confidence == TelemetryConfidence.LOW) {
            val created = syncOrchestrator.ingestEvent(
                DeviceEventType.CAN_BUS_DEGRADED,
                payloadRef = "reason=${vehicleTelemetry.confidenceReason}",
            )
            emitted += created
            auditEmitter.recordOperationalSignal(created.eventType, created.payloadRef)
        }
        if (twinDrift != null && twinDrift.status == DriftStatus.DRIFTED) {
            val created = syncOrchestrator.ingestEvent(
                DeviceEventType.TWIN_DRIFT_DETECTED,
                payloadRef = "mismatches=${twinDrift.mismatchedFields.joinToString(",")}",
            )
            emitted += created
            auditEmitter.recordOperationalSignal(created.eventType, created.payloadRef)
        }
        return emitted
    }

    fun buildIncidentBundle(
        incidentId: String,
        events: List<StoredEvent>,
        chainRecords: List<ChainedEvidenceRecord>,
    ): IncidentBundle? {
        val builder = incidentBundleBuilder ?: return null
        return builder.build(
            deviceId = deviceId,
            incidentId = incidentId,
            events = events,
            chainRecords = chainRecords,
            auditEnvelopes = auditSnapshotProvider(),
        )
    }
}
