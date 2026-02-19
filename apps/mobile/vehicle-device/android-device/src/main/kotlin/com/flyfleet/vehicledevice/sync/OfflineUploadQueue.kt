package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.events.EventPriorityResolver
import com.flyfleet.vehicledevice.events.UploadPriorityTier
import java.util.UUID

enum class QueueItemStatus {
    PENDING,
    IN_FLIGHT,
    DEAD_LETTER,
}

data class QueueItem(
    val id: String,
    val eventType: DeviceEventType,
    val createdAtMs: Long,
    val payloadRef: String,
    val tier: UploadPriorityTier,
    var status: QueueItemStatus = QueueItemStatus.PENDING,
    var attempts: Int = 0,
    var nextAttemptAtMs: Long = createdAtMs,
    var lastError: String? = null,
)

interface Clock {
    fun nowMs(): Long
}

class SystemClock : Clock {
    override fun nowMs(): Long = System.currentTimeMillis()
}

class OfflineUploadQueue(
    private val clock: Clock,
    private val priorityResolver: EventPriorityResolver = EventPriorityResolver(),
    private val maxRetries: Int = 8,
    private val baseBackoffMs: Long = 1_000,
    private val maxBackoffMs: Long = 300_000,
) {
    private val items = linkedMapOf<String, QueueItem>()

    fun enqueue(eventType: DeviceEventType, payloadRef: String): QueueItem {
        val now = clock.nowMs()
        val item = QueueItem(
            id = UUID.randomUUID().toString(),
            eventType = eventType,
            createdAtMs = now,
            payloadRef = payloadRef,
            tier = priorityResolver.resolveTier(eventType),
        )
        items[item.id] = item
        return item
    }

    fun acquireNextReady(): QueueItem? {
        val now = clock.nowMs()
        val candidate = items.values
            .asSequence()
            .filter { it.status == QueueItemStatus.PENDING }
            .filter { it.nextAttemptAtMs <= now }
            .sortedWith(compareBy<QueueItem>({ it.tier.ordinal }, { it.createdAtMs }))
            .firstOrNull() ?: return null

        candidate.status = QueueItemStatus.IN_FLIGHT
        return candidate
    }

    fun markSucceeded(id: String): Boolean {
        return items.remove(id) != null
    }

    fun markFailed(id: String, reason: String): Boolean {
        val item = items[id] ?: return false
        if (item.status != QueueItemStatus.IN_FLIGHT) {
            return false
        }

        item.attempts += 1
        item.lastError = reason

        if (item.attempts >= maxRetries) {
            item.status = QueueItemStatus.DEAD_LETTER
            item.nextAttemptAtMs = Long.MAX_VALUE
            return true
        }

        val nextDelay = computeBackoffMs(item.attempts)
        item.status = QueueItemStatus.PENDING
        item.nextAttemptAtMs = clock.nowMs() + nextDelay
        return true
    }

    fun snapshot(): List<QueueItem> = items.values.toList()

    private fun computeBackoffMs(attempts: Int): Long {
        val exponential = baseBackoffMs * (1L shl (attempts - 1).coerceAtMost(20))
        return minOf(exponential, maxBackoffMs)
    }
}
