package com.flyfleet.vehicledevice.sync

import com.flyfleet.vehicledevice.events.DeviceEventType
import com.flyfleet.vehicledevice.events.EventPriorityResolver
import com.flyfleet.vehicledevice.events.UploadPriorityTier
import java.io.IOException
import java.nio.charset.StandardCharsets
import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.StandardCopyOption
import java.util.UUID

data class StoredEvent(
    val sequence: Long,
    val id: String,
    val eventType: DeviceEventType,
    val tier: UploadPriorityTier,
    val createdAtMs: Long,
    val payloadRef: String,
    val uploaded: Boolean,
)

interface EventStore {
    fun append(eventType: DeviceEventType, payloadRef: String): StoredEvent
    fun markUploaded(eventId: String): Boolean
    fun findById(eventId: String): StoredEvent?
    fun all(): List<StoredEvent>
    fun pending(): List<StoredEvent>
    fun replayFrom(sequenceInclusive: Long): List<StoredEvent>
}

class JournalEventStore(
    private val journalPath: Path,
    private val clock: Clock,
    private val priorityResolver: EventPriorityResolver = EventPriorityResolver(),
) : EventStore {
    private val events = mutableListOf<StoredEvent>()
    private var nextSequence: Long = 1

    init {
        load()
    }

    override fun append(eventType: DeviceEventType, payloadRef: String): StoredEvent {
        val entry = StoredEvent(
            sequence = nextSequence++,
            id = UUID.randomUUID().toString(),
            eventType = eventType,
            tier = priorityResolver.resolveTier(eventType),
            createdAtMs = clock.nowMs(),
            payloadRef = payloadRef,
            uploaded = false,
        )
        events += entry
        persist()
        return entry
    }

    override fun markUploaded(eventId: String): Boolean {
        val idx = events.indexOfFirst { it.id == eventId }
        if (idx < 0) {
            return false
        }
        val current = events[idx]
        if (current.uploaded) {
            return true
        }
        events[idx] = current.copy(uploaded = true)
        persist()
        return true
    }

    override fun findById(eventId: String): StoredEvent? = events.firstOrNull { it.id == eventId }

    override fun all(): List<StoredEvent> = events.sortedBy { it.sequence }

    override fun pending(): List<StoredEvent> =
        events
            .asSequence()
            .filter { !it.uploaded }
            .sortedWith(compareBy<StoredEvent>({ it.tier.ordinal }, { it.createdAtMs }, { it.sequence }))
            .toList()

    override fun replayFrom(sequenceInclusive: Long): List<StoredEvent> =
        events
            .asSequence()
            .filter { it.sequence >= sequenceInclusive }
            .sortedBy { it.sequence }
            .toList()

    private fun load() {
        events.clear()
        if (!Files.exists(journalPath)) {
            ensureParent()
            return
        }

        val lines = Files.readAllLines(journalPath, StandardCharsets.UTF_8)
        for (line in lines) {
            if (line.isBlank()) {
                continue
            }
            val parsed = decode(line)
            events += parsed
            nextSequence = maxOf(nextSequence, parsed.sequence + 1)
        }
    }

    private fun persist() {
        ensureParent()
        val temp = journalPath.resolveSibling(journalPath.fileName.toString() + ".tmp")
        val payload = buildString {
            for (event in events.sortedBy { it.sequence }) {
                append(encode(event))
                append('\n')
            }
        }
        Files.write(temp, payload.toByteArray(StandardCharsets.UTF_8))
        Files.move(temp, journalPath, StandardCopyOption.REPLACE_EXISTING, StandardCopyOption.ATOMIC_MOVE)
    }

    private fun ensureParent() {
        val parent = journalPath.parent
        if (parent != null) {
            Files.createDirectories(parent)
        }
    }

    private fun encode(event: StoredEvent): String {
        val payload = event.payloadRef.replace("%", "%25").replace("|", "%7C")
        return listOf(
            event.sequence.toString(),
            event.id,
            event.eventType.name,
            event.tier.name,
            event.createdAtMs.toString(),
            payload,
            event.uploaded.toString(),
        ).joinToString("|")
    }

    private fun decode(line: String): StoredEvent {
        val parts = line.split('|')
        if (parts.size != 7) {
            throw IOException("invalid_journal_record")
        }
        return StoredEvent(
            sequence = parts[0].toLong(),
            id = parts[1],
            eventType = DeviceEventType.valueOf(parts[2]),
            tier = UploadPriorityTier.valueOf(parts[3]),
            createdAtMs = parts[4].toLong(),
            payloadRef = parts[5].replace("%7C", "|").replace("%25", "%"),
            uploaded = parts[6].toBooleanStrict(),
        )
    }
}
