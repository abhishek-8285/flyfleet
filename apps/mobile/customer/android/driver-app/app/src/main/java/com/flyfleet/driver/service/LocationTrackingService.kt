package com.flyfleet.driver.service

import android.Manifest
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.content.pm.PackageManager
import android.location.Location
import android.os.Build
import android.os.IBinder
import android.os.Looper
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import com.flyfleet.driver.MainActivity
import com.flyfleet.driver.R
import com.flyfleet.driver.data.local.dao.GPSQueueDao
import com.flyfleet.driver.data.local.entity.GPSQueueItem
import com.flyfleet.driver.data.remote.ApiService
import com.google.android.gms.location.*
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import org.json.JSONArray
import org.json.JSONObject
import javax.inject.Inject

@AndroidEntryPoint
class LocationTrackingService : Service() {

    @Inject
    lateinit var apiService: ApiService

    @Inject
    lateinit var gpsQueueDao: GPSQueueDao

    private lateinit var fusedLocationClient: FusedLocationProviderClient
    private lateinit var locationCallback: LocationCallback
    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val uploadMutex = Mutex()
    private val retryStateByItemId = mutableMapOf<String, RetryState>()
    private val deadLetterIds = mutableSetOf<String>()
    private var periodicFlushJob: Job? = null

    companion object {
        const val CHANNEL_ID = "location_tracking_channel"
        const val NOTIFICATION_ID = 1001
        const val ACTION_START_TRACKING = "START_TRACKING"
        const val ACTION_STOP_TRACKING = "STOP_TRACKING"
        private const val RETRY_BASE_DELAY_MS = 5_000L
        private const val RETRY_MAX_DELAY_MS = 5 * 60 * 1000L
        private const val MAX_RETRY_ATTEMPTS = 5
        private const val MAX_DEAD_LETTER_EVENTS = 200
        private const val PREFS_NAME = "flyfleet"
        private const val PREF_DEAD_LETTER_IDS = "gps_dead_letter_ids"
        private const val PREF_DEAD_LETTER_EVENTS = "gps_dead_letter_events"
    }

    private data class RetryState(
        var attempts: Int = 0,
        var nextAttemptAtMs: Long = 0L,
        var lastError: String = ""
    )

    private data class DeadLetterEvent(
        val id: String,
        val deviceId: String,
        val timestamp: Long,
        val reason: String
    )

    override fun onCreate() {
        super.onCreate()
        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, createNotification())
        loadDeadLetterState()
        startPeriodicFlush()
    }

    private fun startPeriodicFlush() {
        periodicFlushJob?.cancel()
        periodicFlushJob = serviceScope.launch {
            while (isActive) {
                delay(30_000L)
                tryUploadGPSData()
            }
        }
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START_TRACKING -> startLocationUpdates()
            ACTION_STOP_TRACKING -> stopSelf()
        }
        return START_STICKY
    }

    private fun startLocationUpdates() {
        val locationRequest = LocationRequest.Builder(
            Priority.PRIORITY_HIGH_ACCURACY,
            10000L // 10 seconds
        ).apply {
            setMinUpdateIntervalMillis(5000L) // 5 seconds fastest
            setMaxUpdateDelayMillis(15000L) // 15 seconds max delay
        }.build()

        locationCallback = object : LocationCallback() {
            override fun onLocationResult(result: LocationResult) {
                result.lastLocation?.let { location ->
                    handleNewLocation(location)
                }
            }
        }

        if (ActivityCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) 
            == PackageManager.PERMISSION_GRANTED) {
            fusedLocationClient.requestLocationUpdates(
                locationRequest,
                locationCallback,
                Looper.getMainLooper()
            )
        }
    }

    private fun handleNewLocation(location: Location) {
        serviceScope.launch {
            val gpsItem = GPSQueueItem(
                deviceId = getDeviceId(),
                timestamp = System.currentTimeMillis() / 1000,
                latitude = location.latitude,
                longitude = location.longitude,
                altitude = if (location.hasAltitude()) location.altitude else null,
                speed = if (location.hasSpeed()) location.speed.toDouble() * 3.6 else 0.0, // m/s to km/h
                heading = if (location.hasBearing()) location.bearing.toDouble() else 0.0,
                accuracy = if (location.hasAccuracy()) location.accuracy.toDouble() else 0.0,
                satellites = null, // Not available from FusedLocation
                isSynced = false
            )

            // Save to local Room database
            gpsQueueDao.insert(gpsItem)

            // Try to upload immediately if online
            tryUploadGPSData()
        }
    }

    private suspend fun tryUploadGPSData() = uploadMutex.withLock {
        val pendingItems = gpsQueueDao.getUnsyncedItems(limit = 50)
        if (pendingItems.isEmpty()) return

        val now = System.currentTimeMillis()
        val eligibleItems = pendingItems.filter { item ->
            val itemKey = item.id.toString()
            if (deadLetterIds.contains(itemKey)) {
                return@filter false
            }
            val retryState = retryStateByItemId[itemKey]
            retryState == null || retryState.nextAttemptAtMs <= now
        }

        if (eligibleItems.isEmpty()) return

        try {
            val response = apiService.uploadGPSBatch(eligibleItems.map { item ->
                val ignitionState = getIgnitionState(item.speed)
                mapOf(
                    "device_id" to item.deviceId,
                    "timestamp" to item.timestamp,
                    "latitude" to item.latitude,
                    "longitude" to item.longitude,
                    "altitude" to item.altitude,
                    "speed" to item.speed,
                    "heading" to item.heading,
                    "accuracy" to item.accuracy,
                    "ignition" to ignitionState
                )
            })

            if (response.isSuccessful) {
                gpsQueueDao.markAsSynced(eligibleItems.map { it.id })
                eligibleItems.forEach { retryStateByItemId.remove(it.id.toString()) }

                // Delete old synced items (keep last 24 hours)
                val cutoffTime = System.currentTimeMillis() / 1000 - (24 * 3600)
                gpsQueueDao.deleteOldSyncedItems(cutoffTime)
                return
            }

            val failureCode = response.code()
            if (failureCode == 409 || failureCode == 422) {
                eligibleItems.forEach { markDeadLetter(it, "conflict_http_$failureCode") }
                return
            }
            registerBatchFailure(eligibleItems, "http_$failureCode")
        } catch (e: Exception) {
            registerBatchFailure(eligibleItems, e.message ?: e::class.java.simpleName)
        }
    }

    private fun registerBatchFailure(items: List<GPSQueueItem>, reason: String) {
        val now = System.currentTimeMillis()
        items.forEach { item ->
            val itemKey = item.id.toString()
            val current = retryStateByItemId[itemKey] ?: RetryState()
            current.attempts += 1
            current.lastError = reason

            if (current.attempts >= MAX_RETRY_ATTEMPTS) {
                retryStateByItemId[itemKey] = current.apply { nextAttemptAtMs = Long.MAX_VALUE }
                markDeadLetter(item, "max_retries:$reason")
                return@forEach
            }

            val exp = (1L shl (current.attempts - 1).coerceAtMost(10))
            val backoff = (RETRY_BASE_DELAY_MS * exp).coerceAtMost(RETRY_MAX_DELAY_MS)
            current.nextAttemptAtMs = now + backoff
            retryStateByItemId[itemKey] = current
        }
    }

    private fun markDeadLetter(item: GPSQueueItem, reason: String) {
        val itemKey = item.id.toString()
        deadLetterIds.add(itemKey)
        persistDeadLetterIds()
        appendDeadLetterEvent(
            DeadLetterEvent(
                id = itemKey,
                deviceId = item.deviceId,
                timestamp = item.timestamp,
                reason = reason
            )
        )
    }

    private fun loadDeadLetterState() {
        val prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
        val serialized = prefs.getString(PREF_DEAD_LETTER_IDS, "[]") ?: "[]"
        runCatching {
            val arr = JSONArray(serialized)
            for (i in 0 until arr.length()) {
                deadLetterIds.add(arr.getString(i))
            }
        }
    }

    private fun persistDeadLetterIds() {
        val arr = JSONArray()
        deadLetterIds.forEach { arr.put(it) }
        getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
            .edit()
            .putString(PREF_DEAD_LETTER_IDS, arr.toString())
            .apply()
    }

    private fun appendDeadLetterEvent(event: DeadLetterEvent) {
        val prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
        val raw = prefs.getString(PREF_DEAD_LETTER_EVENTS, "[]") ?: "[]"
        val arr = runCatching { JSONArray(raw) }.getOrElse { JSONArray() }
        val obj = JSONObject()
            .put("id", event.id)
            .put("device_id", event.deviceId)
            .put("timestamp", event.timestamp)
            .put("reason", event.reason)
        arr.put(obj)

        while (arr.length() > MAX_DEAD_LETTER_EVENTS) {
            arr.remove(0)
        }
        prefs.edit().putString(PREF_DEAD_LETTER_EVENTS, arr.toString()).apply()
    }

    private fun getDeviceId(): String {
        val prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
        return prefs.getString("device_id", "UNKNOWN") ?: "UNKNOWN"
    }

    private fun getIgnitionState(speedKmh: Double): Boolean {
        val prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
        if (prefs.contains("ignition_state")) {
            return prefs.getBoolean("ignition_state", false)
        }
        // Fallback heuristic when OBD state is unavailable.
        return speedKmh > 3.0
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Location Tracking",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "FlyFleet is tracking your location"
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun createNotification(): Notification {
        val intent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent, PendingIntent.FLAG_IMMUTABLE
        )

        val stopIntent = Intent(this, LocationTrackingService::class.java).apply {
            action = ACTION_STOP_TRACKING
        }
        val stopPendingIntent = PendingIntent.getService(
            this, 0, stopIntent, PendingIntent.FLAG_IMMUTABLE
        )

        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("FlyFleet Tracking Active")
            .setContentText("Your location is being tracked")
            .setSmallIcon(R.drawable.ic_location)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .addAction(R.drawable.ic_stop, "Stop", stopPendingIntent)
            .build()
    }

    override fun onDestroy() {
        super.onDestroy()
        fusedLocationClient.removeLocationUpdates(locationCallback)
        periodicFlushJob?.cancel()
        serviceScope.cancel()
    }

    override fun onBind(intent: Intent?): IBinder? = null
}
