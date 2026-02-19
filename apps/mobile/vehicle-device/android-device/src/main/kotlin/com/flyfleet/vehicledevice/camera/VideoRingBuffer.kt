package com.flyfleet.vehicledevice.camera

import com.flyfleet.vehicledevice.security.SecurityModule
import com.flyfleet.vehicledevice.storage.StorageManager
import timber.log.Timber
import java.io.File
import java.util.concurrent.ConcurrentLinkedQueue

/**
 * Rolling video ring buffer for continuous evidence capture
 * 
 * Maintains configurable duration (5-15 minutes) of video segments.
 * Automatically deletes oldest segments when buffer is full.
 * Supports event-triggered clip extraction with pre/post windows.
 */
class VideoRingBuffer(
    private val targetDurationMinutes: Int = 10,
    private val segmentDurationSeconds: Int = 30,
    private val storageManager: StorageManager
) {
    private val segments = ConcurrentLinkedQueue<VideoSegment>()
    private val lock = Any()

    val segmentCount: Int
        get() = segments.size

    /**
     * Add a new video segment to the ring buffer
     */
    fun addSegment(file: File, durationSeconds: Int, encrypted: Boolean = false) {
        synchronized(lock) {
            val segment = VideoSegment(
                file = file,
                durationSeconds = durationSeconds,
                timestampMs = System.currentTimeMillis(),
                encrypted = encrypted
            )

            segments.add(segment)

            Timber.d("Added segment: ${file.name} (${durationSeconds}s)")

            // Remove oldest segments to maintain target duration
            pruneOldSegments()

            // Check storage and cleanup if needed
            enforceStorageReserve()
        }
    }

    /**
     * Extract clip with pre/post event windows from ring buffer
     */
    fun extractClip(
        preSeconds: Int,
        postSeconds: Int,
        outputDir: File
    ): EventClip? {
        synchronized(lock) {
            val totalNeededSeconds = preSeconds + postSeconds
            val currentDuration = currentDurationMinutes() * 60

            if (currentDuration < totalNeededSeconds) {
                Timber.e("Insufficient buffer: need ${totalNeededSeconds}s, have ${currentDuration}s")
                return null
            }

            // Find segments that contain the event window
            val eventTimestamp = System.currentTimeMillis()
            val preTimestamp = eventTimestamp - (preSeconds * 1000L)
            val postTimestamp = eventTimestamp + (postSeconds * 1000L)

            val relevantSegments = segments.filter { segment ->
                segment.timestampMs <= postTimestamp &&
                (segment.timestampMs + segment.durationSeconds * 1000L) >= preTimestamp
            }

            if (relevantSegments.isEmpty()) {
                Timber.e("No relevant segments found for event window")
                return null
            }

            Timber.d("Found ${relevantSegments.size} segments for event clip")

            // Concatenate segments
            val clipId = java.util.UUID.randomUUID().toString()
            val outputFile = File(outputDir, "event_$clipId.mp4")

            concatenateSegments(relevantSegments, outputFile)

            return EventClip(
                clipId = clipId,
                file = outputFile,
                durationSeconds = preSeconds + postSeconds,
                preEventSeconds = preSeconds,
                postEventSeconds = postSeconds,
                eventType = "event",
                timestamp = eventTimestamp,
                encrypted = true
            )
        }
    }

    /**
     * Return current duration of video in buffer (in minutes)
     */
    fun currentDurationMinutes(): Float {
        val totalSeconds = segments.sumOf { it.durationSeconds }
        return totalSeconds / 60f
    }

    /**
     * Return total storage used by buffer (in MB)
     */
    fun storageUsedMB(): Float {
        val totalBytes = segments.sumOf { it.file.length() }
        return totalBytes / (1024f * 1024f)
    }

    /**
     * Clear all segments from buffer
     */
    fun clear() {
        synchronized(lock) {
            segments.forEach { it.file.delete() }
            segments.clear()
            Timber.i("Ring buffer cleared")
        }
    }

    //region Private Implementation

    private fun pruneOldSegments() {
        val targetDurationSeconds = targetDurationMinutes * 60
        var currentDurationSeconds = segments.sumOf { it.durationSeconds }

        while (currentDurationSeconds > targetDurationSeconds && segments.isNotEmpty()) {
            val oldest = segments.poll()
            oldest?.file?.delete()
            currentDurationSeconds -= oldest?.durationSeconds ?: 0
            Timber.v("Pruned oldest segment: ${oldest?.file?.name}")
        }
    }

    private fun enforceStorageReserve() {
        val storageDir = segments.firstOrNull()?.file?.parentFile ?: return
        val freeSpaceMB = storageDir.freeSpace / (1024f * 1024f)
        val minReserveMB = 512 // Minimum 512MB free space

        while (freeSpaceMB < minReserveMB && segments.isNotEmpty()) {
            val oldest = segments.poll()
            oldest?.file?.delete()
            Timber.w("Storage pressure: deleted oldest segment to free space")
        }
    }

    private fun concatenateSegments(segments: List<VideoSegment>, outputFile: File) {
        // Simple file concatenation
        // In production, would use MediaMuxer for proper MP4 merging
        outputFile.outputStream().use { fos ->
            segments.forEach { segment ->
                segment.file.inputStream().use { input ->
                    val buffer = ByteArray(8192)
                    var bytesRead: Int
                    while (input.read(buffer).also { bytesRead = it } != -1) {
                        fos.write(buffer, 0, bytesRead)
                    }
                }
            }
        }

        Timber.i("Concatenated ${segments.size} segments to ${outputFile.name}")
    }

    //endregion
}

/**
 * Video segment metadata
 */
data class VideoSegment(
    val file: File,
    val durationSeconds: Int,
    val timestampMs: Long,
    val encrypted: Boolean
)
