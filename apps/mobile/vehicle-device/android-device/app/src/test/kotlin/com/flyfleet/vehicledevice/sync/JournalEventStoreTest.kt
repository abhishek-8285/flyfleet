package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import kotlin.io.path.createTempDirectory
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class JournalEventStoreTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    @Test
    fun `appended events persist and are restored across restart`() {
        val dir = createTempDirectory(prefix = "flyfleet-event-store-")
        val file = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)

        val store1 = JournalEventStore(journalPath = file, clock = clock)
        val first = store1.append(DeviceEventType.PANIC_TRIGGERED, "p://1")
        clock.now += 10
        val second = store1.append(DeviceEventType.GPS_SIGNAL_LOSS, "p://2")

        val store2 = JournalEventStore(journalPath = file, clock = clock)
        val restored = store2.all()

        assertEquals(2, restored.size)
        assertEquals(first.id, restored[0].id)
        assertEquals(second.id, restored[1].id)
    }

    @Test
    fun `pending list is deterministic by tier then fifo`() {
        val dir = createTempDirectory(prefix = "flyfleet-event-store-")
        val file = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)

        val store = JournalEventStore(journalPath = file, clock = clock)
        val tier3 = store.append(DeviceEventType.GPS_SIGNAL_LOSS, "p://gps")
        clock.now += 10
        val tier1a = store.append(DeviceEventType.PANIC_TRIGGERED, "p://panic1")
        clock.now += 10
        val tier1b = store.append(DeviceEventType.HARSH_BRAKE_DETECTED, "p://panic2")

        val pending = store.pending()
        assertEquals(listOf(tier1a.id, tier1b.id, tier3.id), pending.map { it.id })
    }

    @Test
    fun `mark uploaded removes from pending but keeps replay history`() {
        val dir = createTempDirectory(prefix = "flyfleet-event-store-")
        val file = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)

        val store = JournalEventStore(journalPath = file, clock = clock)
        val first = store.append(DeviceEventType.PANIC_TRIGGERED, "p://1")
        val second = store.append(DeviceEventType.NETWORK_QUALITY_DROP, "p://2")

        assertTrue(store.markUploaded(first.id))
        assertFalse(store.pending().any { it.id == first.id })
        assertTrue(store.all().any { it.id == first.id && it.uploaded })

        val replay = store.replayFrom(1)
        assertEquals(listOf(first.id, second.id), replay.map { it.id })
    }
}
