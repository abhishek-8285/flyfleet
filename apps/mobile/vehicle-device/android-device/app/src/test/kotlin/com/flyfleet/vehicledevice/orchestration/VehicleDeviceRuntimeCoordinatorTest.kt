package com.flyfleet.vehicledevice.orchestration

import com.flyfleet.vehicledevice.audit.HmacAuditSigner
import com.flyfleet.vehicledevice.audit.InMemoryAuditTrail
import com.flyfleet.vehicledevice.command.CommandEnvelopeGuard
import com.flyfleet.vehicledevice.command.DeviceCommandLifecycle
import com.flyfleet.vehicledevice.command.DeviceCommandType
import com.flyfleet.vehicledevice.command.HmacCommandEnvelopeSigner
import com.flyfleet.vehicledevice.command.InMemoryCommandIdRegistry
import com.flyfleet.vehicledevice.command.InMemoryNonceReplayCache
import com.flyfleet.vehicledevice.command.SignedCommandEnvelope
import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.forensics.EvidenceHashChain
import com.flyfleet.vehicledevice.forensics.EvidenceRecord
import com.flyfleet.vehicledevice.forensics.IncidentBundleBuilder
import com.flyfleet.vehicledevice.observability.HealthMatrix
import com.flyfleet.vehicledevice.observability.ModuleHealth
import com.flyfleet.vehicledevice.runtime.RuntimeState
import com.flyfleet.vehicledevice.service.VehicleMonitoringService
import com.flyfleet.vehicledevice.sync.Clock
import com.flyfleet.vehicledevice.sync.JournalEventStore
import com.flyfleet.vehicledevice.sync.OfflineUploadQueue
import com.flyfleet.vehicledevice.sync.StoredEvent
import com.flyfleet.vehicledevice.sync.SyncPipelineOrchestrator
import com.flyfleet.vehicledevice.sync.SyncStepOutcome
import com.flyfleet.vehicledevice.sync.UploadAttemptResult
import com.flyfleet.vehicledevice.sync.UploadTransport
import com.flyfleet.vehicledevice.twin.DeviceTwinDriftDetector
import com.flyfleet.vehicledevice.twin.DeviceTwinState
import com.flyfleet.vehicledevice.vehicle.FallbackInputs
import com.flyfleet.vehicledevice.vehicle.ObdTelemetryAdapter
import com.flyfleet.vehicledevice.vehicle.TelemetryConfidence
import com.flyfleet.vehicledevice.vehicle.VehicleIntegrationMode
import kotlin.io.path.createTempDirectory
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class VehicleDeviceRuntimeCoordinatorTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    private class SuccessTransport : UploadTransport {
        override fun upload(event: StoredEvent): UploadAttemptResult = UploadAttemptResult(success = true)
    }

    @Test
    fun `coordinator evaluates command safety using live runtime state`() {
        val fixture = createFixture()
        fixture.monitoringService.onBootCompleted()
        fixture.monitoringService.onIgnitionOn()

        val rebootDecision = fixture.coordinator.evaluateCommand(DeviceCommandType.REBOOT_DEVICE, speedMph = 30.0)
        val panicDecision = fixture.coordinator.evaluateCommand(DeviceCommandType.PANIC_ALERT_TEST, speedMph = 30.0)

        assertFalse(rebootDecision.allowed)
        assertTrue(panicDecision.allowed)
    }

    @Test
    fun `coordinator heartbeat includes queue depth and runtime state`() {
        val fixture = createFixture()
        fixture.monitoringService.onBootCompleted()
        fixture.monitoringService.onIgnitionOn()
        fixture.coordinator.ingestEvent(DeviceEventType.GPS_SIGNAL_LOSS, "blob://gps")

        val hb = fixture.coordinator.buildHeartbeat(
            HealthMatrix(
                camera = ModuleHealth.HEALTHY,
                gps = ModuleHealth.DEGRADED,
                canBus = ModuleHealth.HEALTHY,
                network = ModuleHealth.HEALTHY,
                ai = ModuleHealth.HEALTHY,
                storage = ModuleHealth.HEALTHY,
            ),
        )

        assertEquals(RuntimeState.ACTIVE_DRIVE, hb.runtimeState)
        assertEquals(1, hb.queueDepth)
        assertEquals(ModuleHealth.DEGRADED, hb.overallStatus)
    }

    @Test
    fun `coordinator heartbeat propagates telemetry and twin drift context`() {
        val fixture = createFixture()
        fixture.monitoringService.onBootCompleted()

        val telemetry = ObdTelemetryAdapter(
            inputsProvider = {
                FallbackInputs(
                    canAvailable = false,
                    gpsSignalQuality = 0.3,
                    imuAvailable = false,
                    speedConsistency = 0.2,
                )
            },
        ).readSnapshot()
        val drift = DeviceTwinDriftDetector().detect(
            desired = DeviceTwinState(true, true, true, true),
            reported = DeviceTwinState(false, true, false, true),
        )

        val hb = fixture.coordinator.buildHeartbeat(
            healthMatrix = HealthMatrix(
                camera = ModuleHealth.HEALTHY,
                gps = ModuleHealth.HEALTHY,
                canBus = ModuleHealth.DEGRADED,
                network = ModuleHealth.HEALTHY,
                ai = ModuleHealth.HEALTHY,
                storage = ModuleHealth.HEALTHY,
            ),
            vehicleTelemetry = telemetry,
            twinDrift = drift,
        )

        assertEquals(VehicleIntegrationMode.GPS_IMU_FALLBACK, hb.vehicleIntegrationMode)
        assertEquals(TelemetryConfidence.LOW.name, hb.telemetryConfidence)
        assertEquals("state_mismatch", hb.twinDriftReason)
        assertEquals(2, hb.twinMismatchedFieldsCount)
    }

    @Test
    fun `coordinator runs upload pipeline and reports outcome`() {
        val fixture = createFixture()
        fixture.monitoringService.onBootCompleted()
        fixture.monitoringService.onDepotEntered()
        fixture.coordinator.ingestEvent(DeviceEventType.PANIC_TRIGGERED, "blob://panic")

        val outcome = fixture.coordinator.processOneUpload()
        assertEquals(SyncStepOutcome.UPLOAD_SUCCEEDED, outcome)
    }

    @Test
    fun `coordinator emits can degraded and twin drift sync events`() {
        val fixture = createFixture()
        fixture.monitoringService.onBootCompleted()
        fixture.monitoringService.onIgnitionOn()

        val telemetry = ObdTelemetryAdapter(
            inputsProvider = {
                FallbackInputs(
                    canAvailable = false,
                    gpsSignalQuality = 0.2,
                    imuAvailable = false,
                    speedConsistency = 0.2,
                )
            },
        ).readSnapshot()
        val drift = DeviceTwinDriftDetector().detect(
            desired = DeviceTwinState(true, true, true, true),
            reported = DeviceTwinState(false, true, true, true),
        )

        val emitted = fixture.coordinator.emitOperationalSignals(telemetry, drift)
        assertEquals(2, emitted.size)
        assertTrue(emitted.any { it.eventType == DeviceEventType.CAN_BUS_DEGRADED })
        assertTrue(emitted.any { it.eventType == DeviceEventType.TWIN_DRIFT_DETECTED })
    }

    @Test
    fun `coordinator emits signed audit envelopes for command and operational signals`() {
        val fixture = createFixture(withAuditTrail = true)
        fixture.monitoringService.onBootCompleted()
        fixture.monitoringService.onIgnitionOn()

        fixture.coordinator.evaluateCommand(DeviceCommandType.FORCE_SYNC, speedMph = 12.0)
        val telemetry = ObdTelemetryAdapter(
            inputsProvider = {
                FallbackInputs(
                    canAvailable = false,
                    gpsSignalQuality = 0.2,
                    imuAvailable = false,
                    speedConsistency = 0.2,
                )
            },
        ).readSnapshot()
        fixture.coordinator.emitOperationalSignals(vehicleTelemetry = telemetry, twinDrift = null)

        val trail = fixture.auditTrail
        requireNotNull(trail)
        assertTrue(trail.snapshot().size >= 2)
        assertTrue(trail.verifyAll())
    }

    @Test
    fun `coordinator rejects replayed signed command envelope`() {
        val fixture = createFixture(withEnvelopeGuard = true)
        fixture.monitoringService.onBootCompleted()

        val signer = HmacCommandEnvelopeSigner("command-secret".toByteArray())
        val unsigned = SignedCommandEnvelope(
            commandId = "cmd-1",
            nonce = "nonce-1",
            issuedAtMs = 1_000,
            commandType = DeviceCommandType.FORCE_SYNC,
            payload = "{}",
            signature = "",
        )
        val canonical = listOf(
            unsigned.commandId,
            unsigned.nonce,
            unsigned.issuedAtMs.toString(),
            unsigned.commandType.name,
            unsigned.payload,
        ).joinToString("|")
        val signed = unsigned.copy(signature = signer.sign(canonical))

        val first = fixture.coordinator.evaluateSignedCommand(signed, speedMph = 0.0)
        val replay = fixture.coordinator.evaluateSignedCommand(signed, speedMph = 0.0)

        assertTrue(first.allowed)
        assertFalse(replay.allowed)
        assertEquals("replayed_nonce", replay.reason)
    }

    @Test
    fun `coordinator builds signed incident bundle`() {
        val fixture = createFixture(withAuditTrail = true, withIncidentBundle = true)
        fixture.monitoringService.onBootCompleted()
        val event = fixture.coordinator.ingestEvent(DeviceEventType.PANIC_TRIGGERED, "blob://panic")
        fixture.coordinator.evaluateCommand(DeviceCommandType.FORCE_SYNC, speedMph = 0.0)

        val chain = EvidenceHashChain()
        val chainRecord = chain.append(
            EvidenceRecord(
                evidenceId = event.id,
                timestampMs = event.createdAtMs,
                eventType = event.eventType.name,
                payloadDigestHex = "abc",
            ),
        )
        val bundle = fixture.coordinator.buildIncidentBundle(
            incidentId = "inc-100",
            events = listOf(event),
            chainRecords = listOf(chainRecord),
        )

        requireNotNull(bundle)
        assertEquals("inc-100", bundle.manifest.incidentId)
        assertTrue(bundle.manifest.auditEnvelopeCount >= 1)
    }

    private data class Fixture(
        val coordinator: VehicleDeviceRuntimeCoordinator,
        val monitoringService: VehicleMonitoringService,
        val auditTrail: InMemoryAuditTrail?,
    )

    private fun createFixture(
        withAuditTrail: Boolean = false,
        withEnvelopeGuard: Boolean = false,
        withIncidentBundle: Boolean = false,
    ): Fixture {
        val clock = FakeClock(now = 1_000)
        val dir = createTempDirectory(prefix = "coordinator-")
        val store = JournalEventStore(journalPath = dir.resolve("events.log"), clock = clock)
        val queue = OfflineUploadQueue(clock = clock)
        val monitoringService = VehicleMonitoringService()
        val auditTrail = if (withAuditTrail) {
            InMemoryAuditTrail(
                deviceId = "dev-coord-1",
                clock = clock,
                signer = HmacAuditSigner("audit-secret".toByteArray()),
            )
        } else {
            null
        }
        val orchestrator = SyncPipelineOrchestrator(
            eventStore = store,
            offlineQueue = queue,
            uploadTransport = SuccessTransport(),
            evidenceHashChain = EvidenceHashChain(),
        )
        val incidentBundleBuilder = if (withIncidentBundle) {
            IncidentBundleBuilder(HmacAuditSigner("bundle-secret".toByteArray()))
        } else {
            null
        }
        val envelopeGuard = if (withEnvelopeGuard) {
            CommandEnvelopeGuard(
                clock = clock,
                signer = HmacCommandEnvelopeSigner("command-secret".toByteArray()),
                replayCache = InMemoryNonceReplayCache(),
                commandIdRegistry = InMemoryCommandIdRegistry(),
            )
        } else {
            null
        }
        return Fixture(
            coordinator = VehicleDeviceRuntimeCoordinator(
                deviceId = "dev-coord-1",
                clock = clock,
                monitoringService = monitoringService,
                syncOrchestrator = orchestrator,
                uploadQueue = queue,
                commandLifecycle = DeviceCommandLifecycle(),
                auditEmitter = auditTrail ?: com.flyfleet.vehicledevice.audit.NoopAuditEmitter,
                commandEnvelopeGuard = envelopeGuard,
                incidentBundleBuilder = incidentBundleBuilder,
                auditSnapshotProvider = { auditTrail?.snapshot() ?: emptyList() },
            ),
            monitoringService = monitoringService,
            auditTrail = auditTrail,
        )
    }
}
