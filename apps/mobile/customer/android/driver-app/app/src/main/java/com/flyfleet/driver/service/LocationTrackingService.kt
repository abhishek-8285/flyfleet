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

    companion object {
        const val CHANNEL_ID = "location_tracking_channel"
        const val NOTIFICATION_ID = 1001
        const val ACTION_START_TRACKING = "START_TRACKING"
        const val ACTION_STOP_TRACKING = "STOP_TRACKING"
    }

    override fun onCreate() {
        super.onCreate()
        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, createNotification())
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

    private suspend fun tryUploadGPSData() {
        val pendingItems = gpsQueueDao.getUnsyncedItems(limit = 50)
        
        if (pendingItems.isEmpty()) return

        try {
            // Upload batch
            val response = apiService.uploadGPSBatch(pendingItems.map { item ->
                mapOf(
                    "device_id" to item.deviceId,
                    "timestamp" to item.timestamp,
                    "latitude" to item.latitude,
                    "longitude" to item.longitude,
                    "altitude" to item.altitude,
                    "speed" to item.speed,
                    "heading" to item.heading,
                    "accuracy" to item.accuracy,
                    "ignition" to true // TODO: Get from OBD/vehicle state
                )
            })

            if (response.isSuccessful) {
                // Mark as synced
                gpsQueueDao.markAsSynced(pendingItems.map { it.id })
                
                // Delete old synced items (keep last 24 hours)
                val cutoffTime = System.currentTimeMillis() / 1000 - (24 * 3600)
                gpsQueueDao.deleteOldSyncedItems(cutoffTime)
            }
        } catch (e: Exception) {
            // Will retry on next location update
            e.printStackTrace()
        }
    }

    private fun getDeviceId(): String {
        val prefs = getSharedPreferences("flyfleet", MODE_PRIVATE)
        return prefs.getString("device_id", "UNKNOWN") ?: "UNKNOWN"
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
        serviceScope.cancel()
    }

    override fun onBind(intent: Intent?): IBinder? = null
}