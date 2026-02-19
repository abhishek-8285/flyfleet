package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.forensics.EvidenceHashChain
import kotlin.io.path.createTempDirectory
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

class SyncPipelineOrchestratorTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    private class AlwaysSuccessTransport : UploadTransport {
        val uploaded = mutableListOf<StoredEvent>()

        override fun upload(event: StoredEvent): UploadAttemptResult {
            uploaded += event
            return UploadAttemptResult(success = true)
        }
    }

    private class AlwaysFailTransport : UploadTransport {
        override fun upload(event: StoredEvent): UploadAttemptResult {
            return UploadAttemptResult(success = false, error = "network_timeout")
        }
    }

    @Test
    fun `successful pipeline marks uploaded and appends forensic chain`() {
        val dir = createTempDirectory(prefix = "sync-pipeline-")
        val storePath = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)
        val store = JournalEventStore(storePath, clock)
        val queue = OfflineUploadQueue(clock)
        val transport = AlwaysSuccessTransport()
        val chain = EvidenceHashChain()
        val orchestrator = SyncPipelineOrchestrator(store, queue, transport, chain)

        val event = orchestrator.ingestEvent(DeviceEventType.PANIC_TRIGGERED, payloadRef = "blob://panic")
        val outcome = orchestrator.processNextUpload()

        assertEquals(SyncStepOutcome.UPLOAD_SUCCEEDED, outcome)
        assertTrue(store.findById(event.id)?.uploaded == true)
        assertEquals(1, transport.uploaded.size)
        assertEquals(event.id, transport.uploaded.first().id)
        assertTrue(chain.verify())
        assertEquals(1, chain.snapshot().size)
    }

    @Test
    fun `failed upload keeps item retriable with error metadata`() {
        val dir = createTempDirectory(prefix = "sync-pipeline-")
        val storePath = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)
        val store = JournalEventStore(storePath, clock)
        val queue = OfflineUploadQueue(clock, baseBackoffMs = 1_000)
        val chain = EvidenceHashChain()
        val orchestrator = SyncPipelineOrchestrator(store, queue, AlwaysFailTransport(), chain)

        orchestrator.ingestEvent(DeviceEventType.CAMERA_FRAME_STALL, payloadRef = "blob://cam")
        val outcome = orchestrator.processNextUpload()

        assertEquals(SyncStepOutcome.UPLOAD_FAILED, outcome)
        val snapshot = queue.snapshot()
        assertEquals(1, snapshot.size)
        assertEquals(QueueItemStatus.PENDING, snapshot[0].status)
        assertEquals("network_timeout", snapshot[0].lastError)
        assertEquals(0, chain.snapshot().size)
    }

    @Test
    fun `no ready item returns no ready outcome`() {
        val dir = createTempDirectory(prefix = "sync-pipeline-")
        val storePath = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)
        val orchestrator = SyncPipelineOrchestrator(
            JournalEventStore(storePath, clock),
            OfflineUploadQueue(clock),
            AlwaysSuccessTransport(),
            EvidenceHashChain(),
        )

        assertEquals(SyncStepOutcome.NO_READY_ITEM, orchestrator.processNextUpload())
    }

    @Test
    fun `missing event in store is treated as failure`() {
        val dir = createTempDirectory(prefix = "sync-pipeline-")
        val storePath = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)
        val store = JournalEventStore(storePath, clock)
        val queue = OfflineUploadQueue(clock)
        val chain = EvidenceHashChain()
        val orchestrator = SyncPipelineOrchestrator(store, queue, AlwaysSuccessTransport(), chain)

        // intentionally enqueue a queue item not backed by durable store
        queue.enqueue(DeviceEventType.GPS_SIGNAL_LOSS, payloadRef = "missing-event-id")

        assertEquals(SyncStepOutcome.UPLOAD_FAILED, orchestrator.processNextUpload())
        val item = queue.snapshot().firstOrNull()
        assertNotNull(item)
        assertEquals("event_not_found", item.lastError)
    }
}
