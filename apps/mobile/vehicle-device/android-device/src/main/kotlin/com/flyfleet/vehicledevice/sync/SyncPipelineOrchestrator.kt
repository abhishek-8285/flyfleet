package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.forensics.EvidenceHashChain
import com.flyfleet.vehicledevice.forensics.EvidenceRecord
import java.security.MessageDigest

data class UploadAttemptResult(
    val success: Boolean,
    val error: String? = null,
)

interface UploadTransport {
    fun upload(event: StoredEvent): UploadAttemptResult
}

enum class SyncStepOutcome {
    INGESTED,
    UPLOAD_SUCCEEDED,
    UPLOAD_FAILED,
    NO_READY_ITEM,
}

class SyncPipelineOrchestrator(
    private val eventStore: EventStore,
    private val offlineQueue: OfflineUploadQueue,
    private val uploadTransport: UploadTransport,
    private val evidenceHashChain: EvidenceHashChain,
) {
    fun ingestEvent(eventType: DeviceEventType, payloadRef: String): StoredEvent {
        val stored = eventStore.append(eventType, payloadRef)
        // Queue carries event-id as payload reference to keep upload queue and durable event store linked.
        offlineQueue.enqueue(eventType, payloadRef = stored.id)
        return stored
    }

    fun processNextUpload(): SyncStepOutcome {
        val item = offlineQueue.acquireNextReady() ?: return SyncStepOutcome.NO_READY_ITEM
        val event = eventStore.findById(item.payloadRef)
        if (event == null) {
            offlineQueue.markFailed(item.id, "event_not_found")
            return SyncStepOutcome.UPLOAD_FAILED
        }

        val result = uploadTransport.upload(event)
        if (!result.success) {
            offlineQueue.markFailed(item.id, result.error ?: "upload_failed")
            return SyncStepOutcome.UPLOAD_FAILED
        }

        eventStore.markUploaded(event.id)
        offlineQueue.markSucceeded(item.id)
        evidenceHashChain.append(
            EvidenceRecord(
                evidenceId = event.id,
                timestampMs = event.createdAtMs,
                eventType = event.eventType.name,
                payloadDigestHex = sha256Hex(event.payloadRef),
            ),
        )
        return SyncStepOutcome.UPLOAD_SUCCEEDED
    }

    private fun sha256Hex(input: String): String {
        val digest = MessageDigest.getInstance("SHA-256").digest(input.toByteArray())
        return digest.joinToString(separator = "") { "%02x".format(it) }
    }
}
