package com.flyfleet.vehicledevice.sync

/**
 * Produces deterministic replay batches from durable event store.
 */
class EventReplayPlanner(
    private val eventStore: EventStore,
) {
    fun nextPendingBatch(maxItems: Int): List<StoredEvent> {
        if (maxItems <= 0) {
            return emptyList()
        }
        return eventStore.pending().take(maxItems)
    }

    fun replayFromSequence(sequenceInclusive: Long, maxItems: Int): List<StoredEvent> {
        if (maxItems <= 0) {
            return emptyList()
        }
        return eventStore.replayFrom(sequenceInclusive).take(maxItems)
    }
}
