package com.flyfleet.vehicledevice.telemetry

import android.Manifest
import android.content.Context
import android.content.pm.PackageManager
import android.location.Location
import android.os.Looper
import android.os.PowerManager
import com.google.android.gms.location.*
import com.google.android.gms.tasks.Task
import com.flyfleet.vehicledevice.api.ApiClient
import com.flyfleet.vehicledevice.models.*
import kotlinx.coroutines.*
import kotlinx.coroutines.CancellableContinuation
import kotlin.coroutines.resume
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import timber.log.Timber
import java.util.concurrent.ConcurrentLinkedQueue
import java.util.concurrent.atomic.AtomicLong

/**
 * Telemetry Service for FlyFleet Vehicle Device
 * 
 * Responsibilities:
 * - GPS location tracking at adaptive intervals
 * - CAN bus data integration (when available)
 * - Telemetry batch collection
 * - Periodic upload to backend
 * - Wake lock management
 */
class TelemetryService(
    private val context: Context,
    private val apiClient: ApiClient,
    private val config: TelemetryConfig = TelemetryConfig()
) {
    companion object {
        private const val TAG = "TelemetryService"
        private const val LOCATION_UPDATE_INTERVAL_MS = 1000L // 1 second
        private const val LOCATION_UPDATE_FASTEST_INTERVAL_MS = 500L // 0.5 seconds
    }

    private val fusedLocationClient: FusedLocationProviderClient =
        LocationServices.getFusedLocationProviderClient(context)

    private val powerManager: PowerManager =
        context.getSystemService(Context.POWER_SERVICE) as PowerManager

    private var wakeLock: PowerManager.WakeLock? = null

    private val telemetryBuffer = ConcurrentLinkedQueue<TelemetryPoint>()
    private val batchCounter = AtomicLong(0)
    
    private var locationCallback: LocationCallback? = null
    private var isTracking = false
    private var lastUploadTime = 0L

    // State flows for observability
    private val _trackingState = MutableStateFlow<TrackingState>(TrackingState.IDLE)
    val trackingState: StateFlow<TrackingState> = _trackingState.asStateFlow()

    private val _telemetryStats = MutableStateFlow(TelemetryStats())
    val telemetryStats: StateFlow<TelemetryStats> = _telemetryStats.asStateFlow()

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    /**
     * Start telemetry tracking
     */
    fun startTracking() {
        if (isTracking) {
            Timber.d("Already tracking, skipping")
            return
        }

        Timber.i("Starting telemetry tracking")

        // Check permissions
        if (!hasLocationPermission()) {
            Timber.e("Location permission not granted")
            _trackingState.value = TrackingState.ERROR("Location permission denied")
            return
        }

        // Acquire wake lock
        acquireWakeLock()

        // Start location updates
        startLocationUpdates()

        isTracking = true
        _trackingState.value = TrackingState.TRACKING

        // Start periodic upload job
        startPeriodicUpload()
    }

    /**
     * Stop telemetry tracking
     */
    fun stopTracking() {
        if (!isTracking) return

        Timber.i("Stopping telemetry tracking")

        // Release wake lock
        releaseWakeLock()

        // Stop location updates
        stopLocationUpdates()

        // Upload remaining telemetry
        uploadTelemetryBatch()

        isTracking = false
        _trackingState.value = TrackingState.IDLE
    }

    /**
     * Get current location
     */
    suspend fun getCurrentLocation(): Location? {
        return try {
            if (!hasLocationPermission()) {
                return null
            }

            fusedLocationClient.lastLocation.await()
        } catch (e: Exception) {
            Timber.e(e, "Failed to get current location")
            null
        }
    }

    /**
     * Get telemetry stats
     */
    fun getStats(): TelemetryStats = _telemetryStats.value

    /**
     * Cleanup resources
     */
    fun shutdown() {
        stopTracking()
        scope.cancel()
    }

    //region Private Implementation

    private fun hasLocationPermission(): Boolean {
        return context.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) ==
                PackageManager.PERMISSION_GRANTED
    }

    private fun acquireWakeLock() {
        try {
            wakeLock = powerManager.newWakeLock(
                PowerManager.PARTIAL_WAKE_LOCK,
                "flyfleet::TelemetryService"
            ).apply {
                acquire(10 * 60 * 1000L /*10 minutes*/)
            }
            Timber.d("Wake lock acquired")
        } catch (e: Exception) {
            Timber.e(e, "Failed to acquire wake lock")
        }
    }

    private fun releaseWakeLock() {
        try {
            wakeLock?.let {
                if (it.isHeld) {
                    it.release()
                    Timber.d("Wake lock released")
                }
            }
        } catch (e: Exception) {
            Timber.e(e, "Failed to release wake lock")
        }
        wakeLock = null
    }

    private fun startLocationUpdates() {
        val locationRequest = LocationRequest.Builder(
            Priority.PRIORITY_HIGH_ACCURACY,
            LOCATION_UPDATE_INTERVAL_MS
        ).apply {
            setMinUpdateIntervalMillis(LOCATION_UPDATE_FASTEST_INTERVAL_MS)
            setWaitForAccurateLocation(true)
        }.build()

        locationCallback = object : LocationCallback() {
            override fun onLocationResult(locationResult: LocationResult) {
                locationResult.lastLocation?.let { location ->
                    processLocation(location)
                }
            }
        }

        try {
            fusedLocationClient.requestLocationUpdates(
                locationRequest,
                locationCallback!!,
                Looper.getMainLooper()
            )
            Timber.d("Location updates started")
        } catch (e: Exception) {
            Timber.e(e, "Failed to start location updates")
            _trackingState.value = TrackingState.ERROR("Location updates failed: ${e.message}")
        }
    }

    private fun stopLocationUpdates() {
        locationCallback?.let {
            fusedLocationClient.removeLocationUpdates(it)
            Timber.d("Location updates stopped")
        }
        locationCallback = null
    }

    private fun processLocation(location: Location) {
        // Create telemetry point
        val telemetryPoint = TelemetryPoint(
            timestamp = System.currentTimeMillis(),
            gps = GPSData(
                latitude = location.latitude,
                longitude = location.longitude,
                accuracy = location.accuracy.toDouble(),
                satellites = 0 // Not available from FusedLocationProvider
            ),
            canBus = null, // Would come from CAN bus service
            speedKph = (location.speed * 3.6).toDouble(), // Convert m/s to km/h
            heading = location.bearing.toDouble(),
            altitude = location.altitude,
            accuracy = location.accuracy.toDouble()
        )

        // Add to buffer
        telemetryBuffer.add(telemetryPoint)

        // Update stats
        val currentStats = _telemetryStats.value
        _telemetryStats.value = currentStats.copy(
            pointsCollected = currentStats.pointsCollected + 1,
            lastLocationTime = System.currentTimeMillis()
        )

        // Check if we should upload
        if (telemetryBuffer.size >= config.batchSize ||
            System.currentTimeMillis() - lastUploadTime > config.uploadIntervalMs) {
            uploadTelemetryBatch()
        }
    }

    private fun startPeriodicUpload() {
        scope.launch {
            while (isTracking) {
                delay(config.uploadIntervalMs)
                uploadTelemetryBatch()
            }
        }
    }

    private fun uploadTelemetryBatch() {
        if (telemetryBuffer.isEmpty()) {
            return
        }

        // Collect batch
        val batch = mutableListOf<TelemetryPoint>()
        while (telemetryBuffer.isNotEmpty() && batch.size < config.batchSize) {
            telemetryBuffer.poll()?.let { batch.add(it) }
        }

        if (batch.isEmpty()) {
            return
        }

        // Create batch request
        val batchId = "batch_${System.currentTimeMillis()}_${batchCounter.incrementAndGet()}"
        val request = TelemetryBatchRequest(
            deviceId = config.deviceId,
            batchId = batchId,
            telemetry = batch,
            timestamp = System.currentTimeMillis()
        )

        // Upload
        scope.launch {
            try {
                val result = apiClient.uploadTelemetryBatch(request)
                result.onSuccess { response ->
                    Timber.d("Telemetry batch uploaded: ${response.receivedCount} points, next batch: ${response.nextBatchId}")
                    
                    lastUploadTime = System.currentTimeMillis()
                    
                    // Update stats
                    val currentStats = _telemetryStats.value
                    _telemetryStats.value = currentStats.copy(
                        batchesUploaded = currentStats.batchesUploaded + 1,
                        lastUploadTime = lastUploadTime
                    )
                }.onFailure { error ->
                    Timber.e(error, "Telemetry upload failed")
                    
                    // Re-add to buffer for retry
                    telemetryBuffer.addAll(batch)
                    
                    // Update stats
                    val currentStats = _telemetryStats.value
                    _telemetryStats.value = currentStats.copy(
                        uploadFailures = currentStats.uploadFailures + 1
                    )
                }
            } catch (e: Exception) {
                Timber.e(e, "Telemetry upload exception")
                telemetryBuffer.addAll(batch)
            }
        }
    }

    //endregion
}

/**
 * Telemetry configuration
 */
data class TelemetryConfig(
    val deviceId: String = "device-001",
    val batchSize: Int = 60, // Upload every 60 points
    val uploadIntervalMs: Long = 60000L, // Or every 60 seconds
    val adaptiveTracking: Boolean = true,
    val idleIntervalSeconds: Int = 30,
    val activeIntervalSeconds: Int = 1
)

/**
 * Tracking state
 */
sealed class TrackingState {
    object IDLE : TrackingState()
    object TRACKING : TrackingState()
    data class ERROR(val message: String) : TrackingState()
}

/**
 * Telemetry statistics
 */
data class TelemetryStats(
    val pointsCollected: Long = 0,
    val batchesUploaded: Long = 0,
    val uploadFailures: Long = 0,
    val lastLocationTime: Long = 0,
    val lastUploadTime: Long = 0
)

/**
 * Suspend function to await location
 */
suspend fun Task<Location>.await(): Location? {
    return try {
        kotlinx.coroutines.suspendCancellableCoroutine<Location?> { continuation: CancellableContinuation<Location?> ->
            addOnSuccessListener { location: Location? ->
                continuation.resume(location)
            }
            addOnFailureListener { exception: Exception ->
                continuation.resume(null)
            }
        }
    } catch (e: Exception) {
        null
    }
}
