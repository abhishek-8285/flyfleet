package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.events.UploadPriorityTier
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class OfflineUploadQueueTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    @Test
    fun `acquires higher priority tier before lower priority`() {
        val clock = FakeClock(now = 1_000)
        val queue = OfflineUploadQueue(clock = clock)

        queue.enqueue(DeviceEventType.GPS_SIGNAL_LOSS, payloadRef = "p3")
        queue.enqueue(DeviceEventType.PANIC_TRIGGERED, payloadRef = "p1")

        val first = queue.acquireNextReady()
        assertNotNull(first)
        assertEquals(UploadPriorityTier.TIER1, first.tier)
    }

    @Test
    fun `within same tier older items are processed first`() {
        val clock = FakeClock(now = 1_000)
        val queue = OfflineUploadQueue(clock = clock)

        val older = queue.enqueue(DeviceEventType.PANIC_TRIGGERED, payloadRef = "old")
        clock.now += 10
        val newer = queue.enqueue(DeviceEventType.PANIC_TRIGGERED, payloadRef = "new")

        val first = queue.acquireNextReady()
        assertNotNull(first)
        assertEquals(older.id, first.id)

        queue.markSucceeded(first.id)
        val second = queue.acquireNextReady()
        assertNotNull(second)
        assertEquals(newer.id, second.id)
    }

    @Test
    fun `failed item is backoff delayed and not immediately reacquired`() {
        val clock = FakeClock(now = 1_000)
        val queue = OfflineUploadQueue(clock = clock, baseBackoffMs = 1_000, maxBackoffMs = 60_000)

        val item = queue.enqueue(DeviceEventType.CAMERA_FRAME_STALL, payloadRef = "cam")
        val acquired = queue.acquireNextReady()
        assertNotNull(acquired)
        assertEquals(item.id, acquired.id)

        assertTrue(queue.markFailed(item.id, "net-timeout"))
        assertNull(queue.acquireNextReady())

        clock.now += 1_000
        val retry = queue.acquireNextReady()
        assertNotNull(retry)
        assertEquals(item.id, retry.id)
    }

    @Test
    fun `moves to dead letter after max retries`() {
        val clock = FakeClock(now = 1_000)
        val queue = OfflineUploadQueue(clock = clock, maxRetries = 3, baseBackoffMs = 1)

        val item = queue.enqueue(DeviceEventType.NETWORK_QUALITY_DROP, payloadRef = "net")

        repeat(3) {
            val next = queue.acquireNextReady()
            assertNotNull(next)
            assertEquals(item.id, next.id)
            assertTrue(queue.markFailed(item.id, "failure-${it + 1}"))
            clock.now += 100
        }

        val stored = queue.snapshot().first { it.id == item.id }
        assertEquals(QueueItemStatus.DEAD_LETTER, stored.status)
    }
}
