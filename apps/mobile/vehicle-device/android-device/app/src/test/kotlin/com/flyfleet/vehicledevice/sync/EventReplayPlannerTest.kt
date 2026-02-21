package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import kotlin.io.path.createTempDirectory
import kotlin.test.Test
import kotlin.test.assertEquals

class EventReplayPlannerTest {
    private class FakeClock(var now: Long) : Clock {
        override fun nowMs(): Long = now
    }

    @Test
    fun `next pending batch respects deterministic pending order and max size`() {
        val dir = createTempDirectory(prefix = "flyfleet-replay-")
        val file = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)
        val store = JournalEventStore(journalPath = file, clock = clock)

        val e1 = store.append(DeviceEventType.GPS_SIGNAL_LOSS, "p://gps")
        clock.now += 10
        val e2 = store.append(DeviceEventType.PANIC_TRIGGERED, "p://panic")
        clock.now += 10
        val e3 = store.append(DeviceEventType.HARSH_BRAKE_DETECTED, "p://brake")

        val planner = EventReplayPlanner(store)
        val batch = planner.nextPendingBatch(maxItems = 2)

        assertEquals(listOf(e2.id, e3.id), batch.map { it.id })
        assertEquals(e1.id, planner.nextPendingBatch(maxItems = 3).last().id)
    }

    @Test
    fun `replay from sequence returns chronological records`() {
        val dir = createTempDirectory(prefix = "flyfleet-replay-")
        val file = dir.resolve("events.log")
        val clock = FakeClock(now = 1_000)
        val store = JournalEventStore(journalPath = file, clock = clock)

        store.append(DeviceEventType.PANIC_TRIGGERED, "p://1")
        val second = store.append(DeviceEventType.NETWORK_QUALITY_DROP, "p://2")
        val third = store.append(DeviceEventType.CAMERA_FRAME_STALL, "p://3")

        val planner = EventReplayPlanner(store)
        val replay = planner.replayFromSequence(sequenceInclusive = 2, maxItems = 10)

        assertEquals(listOf(second.id, third.id), replay.map { it.id })
    }
}
