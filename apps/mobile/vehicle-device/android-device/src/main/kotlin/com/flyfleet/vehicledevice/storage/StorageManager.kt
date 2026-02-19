package com.flyfleet.vehicledevice.storage

import android.content.Context
import android.os.Environment
import android.os.StatFs
import com.flyfleet.vehicledevice.camera.EventClip
import timber.log.Timber
import java.io.File
import java.text.SimpleDateFormat
import java.util.*

/**
 * Storage Manager for FlyFleet Vehicle Device
 * 
 * Responsibilities:
 * - Manage local file storage for video clips
 * - Ring buffer storage management
 * - Storage quota enforcement
 * - Automatic cleanup of expired files
 * - Sync queue management
 */
class StorageManager(private val context: Context) {

    companion object {
        private const val TAG = "StorageManager"
        
        // Storage directories
        private const val BASE_DIR = "flyfleet"
        private const val CLIPS_DIR = "clips"
        private const val RING_BUFFER_DIR = "ring_buffer"
        private const val METADATA_DIR = "metadata"
        private const val SYNC_QUEUE_DIR = "sync_queue"
        
        // Storage quotas (in MB)
        private const val DEFAULT_TOTAL_QUOTA_MB = 15360L // 15GB default
        private const val CLIPS_QUOTA_PERCENT = 0.33 // 33% for clips (~5GB)
        private const val RING_BUFFER_QUOTA_PERCENT = 0.67 // 67% for ring buffer (~10GB)
        
        // Cleanup thresholds
        private const val CLEANUP_THRESHOLD_PERCENT = 0.90 // Start cleanup at 90% full
        private const val MIN_FREE_SPACE_MB = 512L // Keep 512MB free
    }

    // Base directories
    private val baseDir: File by lazy {
        val dir = File(context.getExternalFilesDir(null), BASE_DIR)
        if (!dir.exists()) {
            dir.mkdirs()
            Timber.i("Created base directory: ${dir.absolutePath}")
        }
        dir
    }

    val clipsDirectory: File by lazy {
        getDirectory(CLIPS_DIR)
    }

    val ringBufferDirectory: File by lazy {
        getDirectory(RING_BUFFER_DIR)
    }

    val metadataDirectory: File by lazy {
        getDirectory(METADATA_DIR)
    }

    val syncQueueDirectory: File by lazy {
        getDirectory(SYNC_QUEUE_DIR)
    }

    // Storage quotas
    private var totalQuotaBytes: Long = DEFAULT_TOTAL_QUOTA_MB * 1024 * 1024
    private var clipsQuotaBytes: Long = (totalQuotaBytes * CLIPS_QUOTA_PERCENT).toLong()
    private var ringBufferQuotaBytes: Long = (totalQuotaBytes * RING_BUFFER_QUOTA_PERCENT).toLong()

    /**
     * Initialize storage manager with custom quotas
     */
    fun initialize(
        totalQuotaMB: Long = DEFAULT_TOTAL_QUOTA_MB,
        clipsQuotaPercent: Double = CLIPS_QUOTA_PERCENT,
        ringBufferQuotaPercent: Double = RING_BUFFER_QUOTA_PERCENT
    ) {
        totalQuotaBytes = totalQuotaMB * 1024 * 1024
        clipsQuotaBytes = (totalQuotaBytes * clipsQuotaPercent).toLong()
        ringBufferQuotaBytes = (totalQuotaBytes * ringBufferQuotaPercent).toLong()

        Timber.i("Storage initialized: total=${totalQuotaMB}MB, clips=${clipsQuotaBytes / 1024 / 1024}MB, ring=${ringBufferQuotaBytes / 1024 / 1024}MB")
    }

    /**
     * Create a new clip file
     */
    fun createClipFile(eventType: String = "event"): File {
        val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
        val clipId = java.util.UUID.randomUUID().toString()
        return File(clipsDirectory, "${eventType}_${timestamp}_${clipId}.mp4")
    }

    /**
     * Create a new ring buffer segment file
     */
    fun createSegmentFile(): File {
        val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
        val segmentId = getSegmentCount()
        return File(ringBufferDirectory, "segment_${timestamp}_${segmentId}.mp4")
    }

    /**
     * Get storage statistics
     */
    fun getStorageStats(): StorageStats {
        val statFs = StatFs(baseDir.absolutePath)
        
        val totalBytes = statFs.totalBytes
        val freeBytes = statFs.freeBytes
        val availableBytes = statFs.availableBytes
        val usedBytes = totalBytes - freeBytes

        val clipsUsedBytes = clipsDirectory.walkTopDown()
            .filter { it.isFile }
            .sumOf { it.length() }

        val ringBufferUsedBytes = ringBufferDirectory.walkTopDown()
            .filter { it.isFile }
            .sumOf { it.length() }

        return StorageStats(
            totalBytes = totalBytes,
            freeBytes = freeBytes,
            availableBytes = availableBytes,
            usedBytes = usedBytes,
            clipsUsedBytes = clipsUsedBytes,
            ringBufferUsedBytes = ringBufferUsedBytes,
            totalQuotaBytes = totalQuotaBytes,
            clipsQuotaBytes = clipsQuotaBytes,
            ringBufferQuotaBytes = ringBufferQuotaBytes,
            clipsUsagePercent = (clipsUsedBytes.toDouble() / clipsQuotaBytes) * 100,
            ringBufferUsagePercent = (ringBufferUsedBytes.toDouble() / ringBufferQuotaBytes) * 100,
            isStoragePressure = (usedBytes.toDouble() / totalBytes) > CLEANUP_THRESHOLD_PERCENT
        )
    }

    /**
     * Check if there's enough space for recording
     */
    fun hasSpaceForRecording(estimatedSizeBytes: Long): Boolean {
        val stats = getStorageStats()
        val availableForRingBuffer = ringBufferQuotaBytes - stats.ringBufferUsedBytes
        return availableForRingBuffer >= estimatedSizeBytes
    }

    /**
     * Cleanup expired and old files
     */
    fun cleanup(
        clipsRetentionDays: Int = 14,
        minFreeSpaceMB: Long = MIN_FREE_SPACE_MB
    ) {
        Timber.i("Starting storage cleanup")

        val stats = getStorageStats()
        
        // Delete clips older than retention period
        val clipsDeleted = deleteExpiredClips(clipsRetentionDays)
        Timber.i("Deleted $clipsDeleted expired clips")

        // If still under pressure, delete oldest non-critical clips
        if (stats.isStoragePressure) {
            val pressureDeleted = deleteOldestNonCriticalClips()
            Timber.i("Deleted $pressureDeleted clips due to storage pressure")
        }

        // Ensure minimum free space
        val currentStats = getStorageStats()
        if (currentStats.freeBytes < minFreeSpaceMB * 1024 * 1024) {
            val spaceFreed = freeUpSpace(minFreeSpaceMB * 1024 * 1024)
            Timber.i("Freed ${spaceFreed / 1024 / 1024}MB to maintain minimum free space")
        }

        Timber.i("Storage cleanup complete")
    }

    /**
     * Delete all files (for device wipe)
     */
    fun wipeAll() {
        Timber.w("Wiping all storage")
        
        baseDir.deleteRecursively()
        baseDir.mkdirs()
        
        Timber.i("Storage wiped")
    }

    //region Private Implementation

    private fun getDirectory(name: String): File {
        val dir = File(baseDir, name)
        if (!dir.exists()) {
            dir.mkdirs()
            Timber.d("Created directory: ${dir.absolutePath}")
        }
        return dir
    }

    private fun getSegmentCount(): Int {
        return ringBufferDirectory.listFiles()?.size ?: 0
    }

    private fun deleteExpiredClips(retentionDays: Int): Int {
        val cutoffTime = System.currentTimeMillis() - (retentionDays * 24 * 60 * 60 * 1000L)
        var deleted = 0

        clipsDirectory.listFiles()?.forEach { file ->
            if (file.lastModified() < cutoffTime) {
                if (file.delete()) {
                    deleted++
                    Timber.v("Deleted expired clip: ${file.name}")
                }
            }
        }

        return deleted
    }

    private fun deleteOldestNonCriticalClips(): Int {
        val clips = clipsDirectory.listFiles()
            ?.filter { it.isFile }
            ?.sortedBy { it.lastModified() }
            ?: return 0

        var deleted = 0
        val targetDeletions = (clips.size * 0.2).toInt() // Delete 20% of oldest clips

        for (i in 0 until minOf(targetDeletions, clips.size)) {
            if (clips[i].delete()) {
                deleted++
                Timber.v("Deleted oldest clip: ${clips[i].name}")
            }
        }

        return deleted
    }

    private fun freeUpSpace(targetFreeBytes: Long): Long {
        var freedBytes = 0L
        val currentStats = getStorageStats()

        if (currentStats.freeBytes >= targetFreeBytes) {
            return 0
        }

        val bytesNeeded = targetFreeBytes - currentStats.freeBytes
        
        // Delete oldest clips until we have enough space
        val clips = clipsDirectory.listFiles()
            ?.filter { it.isFile }
            ?.sortedBy { it.lastModified() }
            ?: return 0

        for (file in clips) {
            if (freedBytes >= bytesNeeded) break
            
            val fileSize = file.length()
            if (file.delete()) {
                freedBytes += fileSize
                Timber.v("Freed ${fileSize / 1024}KB: ${file.name}")
            }
        }

        return freedBytes
    }

    //endregion
}

/**
 * Storage statistics
 */
data class StorageStats(
    val totalBytes: Long,
    val freeBytes: Long,
    val availableBytes: Long,
    val usedBytes: Long,
    val clipsUsedBytes: Long,
    val ringBufferUsedBytes: Long,
    val totalQuotaBytes: Long,
    val clipsQuotaBytes: Long,
    val ringBufferQuotaBytes: Long,
    val clipsUsagePercent: Double,
    val ringBufferUsagePercent: Double,
    val isStoragePressure: Boolean
) {
    val totalUsagePercent: Double
        get() = (usedBytes.toDouble() / totalBytes) * 100

    val freeSpaceMB: Long
        get() = freeBytes / 1024 / 1024

    val clipsUsedMB: Long
        get() = clipsUsedBytes / 1024 / 1024

    val ringBufferUsedMB: Long
        get() = ringBufferUsedBytes / 1024 / 1024
}
