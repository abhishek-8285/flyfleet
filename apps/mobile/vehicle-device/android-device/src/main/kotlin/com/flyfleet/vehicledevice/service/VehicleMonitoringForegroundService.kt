package com.flyfleet.vehicledevice.service

import android.app.*
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder
import android.os.PowerManager
import androidx.core.app.ServiceCompat
import com.flyfleet.vehicledevice.VehicleDeviceApplication
import com.flyfleet.vehicledevice.R
import com.flyfleet.vehicledevice.camera.CameraService
import com.flyfleet.vehicledevice.config.DeviceConfig
import com.flyfleet.vehicledevice.runtime.RuntimeSignal
import com.flyfleet.vehicledevice.runtime.RuntimeState
import com.flyfleet.vehicledevice.telemetry.TelemetryService
import com.flyfleet.vehicledevice.telemetry.TrackingState
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.collectLatest
import timber.log.Timber

/**
 * Foreground Service for Continuous Vehicle Monitoring
 * 
 * This is the main Android service that:
 * - Runs as a foreground service with persistent notification
 * - Manages camera recording lifecycle
 * - Manages GPS telemetry tracking
 * - Handles state transitions (BOOTING, IDLE, ACTIVE_DRIVE, etc.)
 * - Survives app kills and device reboots
 */
class VehicleMonitoringForegroundService : Service() {

    companion object {
        private const val TAG = "VehicleMonitoringService"
        private const val NOTIFICATION_ID = 1001
        private const val CHANNEL_ID = "vehicle_monitoring"

        // Action constants
        const val ACTION_START = "com.flyfleet.vehicledevice.START_MONITORING"
        const val ACTION_STOP = "com.flyfleet.vehicledevice.STOP_MONITORING"
        const val ACTION_PAUSE = "com.flyfleet.vehicledevice.PAUSE_MONITORING"
        const val ACTION_RESUME = "com.flyfleet.vehicledevice.RESUME_MONITORING"

        // Start service helper
        fun start(context: Context) {
            val intent = Intent(context, VehicleMonitoringForegroundService::class.java).apply {
                action = ACTION_START
            }
            
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }

        // Stop service helper
        fun stop(context: Context) {
            val intent = Intent(context, VehicleMonitoringForegroundService::class.java).apply {
                action = ACTION_STOP
            }
            context.startService(intent)
        }
    }

    // Core services
    private lateinit var cameraService: CameraService
    private lateinit var telemetryService: TelemetryService
    private lateinit var deviceConfig: DeviceConfig

    // State management
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private var currentState: RuntimeState = RuntimeState.BOOTING

    // Wake lock for background operation
    private var wakeLock: PowerManager.WakeLock? = null

    override fun onCreate() {
        super.onCreate()

        Timber.i("VehicleMonitoringForegroundService created")

        // Get application components
        val app = application as VehicleDeviceApplication
        deviceConfig = app.deviceConfig

        // Initialize camera service
        cameraService = CameraService(
            context = this,
            lifecycleOwner = DummyLifecycleOwner(), // Service doesn't have lifecycle, use dummy
            securityModule = app.securityModule,
            storageManager = app.storageManager
        )

        // Initialize telemetry service
        telemetryService = TelemetryService(
            context = this,
            apiClient = app.apiClient
        )

        // Initialize state machine
        currentState = RuntimeState.BOOTING

        // Create notification channel
        createNotificationChannel()

        // Start as foreground service
        startForegroundWithNotification()

        // Acquire wake lock
        acquireWakeLock()

        // Start monitoring
        startMonitoring()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START -> {
                Timber.i("Start monitoring requested")
                startMonitoring()
            }
            ACTION_STOP -> {
                Timber.i("Stop monitoring requested")
                stopMonitoring()
                stopForeground(STOP_FOREGROUND_REMOVE)
                stopSelf()
            }
            ACTION_PAUSE -> {
                Timber.i("Pause monitoring requested")
                pauseMonitoring()
            }
            ACTION_RESUME -> {
                Timber.i("Resume monitoring requested")
                resumeMonitoring()
            }
        }

        // START_STICKY ensures service restarts if killed
        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder? {
        // Not a bound service
        return null
    }

    override fun onDestroy() {
        Timber.i("VehicleMonitoringForegroundService destroyed")

        // Stop monitoring
        stopMonitoring()

        // Release wake lock
        releaseWakeLock()

        // Cancel coroutines
        serviceScope.cancel()

        super.onDestroy()
    }

    override fun onTaskRemoved(rootIntent: Intent?) {
        Timber.i("Task removed - service will continue running")
        // Don't call super - we want to keep running
    }

    //region Monitoring Lifecycle

    private fun startMonitoring() {
        Timber.i("Starting vehicle monitoring")

        // Transition to ACTIVE_DRIVE state
        currentState = RuntimeState.ACTIVE_DRIVE

        // Start camera recording
        val cameraStatus = cameraService.initialize()
        Timber.i("Camera initialized: $cameraStatus")

        cameraService.startRecording()
        cameraService.captureSampledFrames()

        // Start telemetry tracking
        telemetryService.startTracking()

        // Observe telemetry state
        serviceScope.launch {
            telemetryService.trackingState.collectLatest { state ->
                when (state) {
                    is TrackingState.ERROR -> {
                        Timber.e("Telemetry error: ${state.message}")
                        // Handle error
                    }
                    else -> {
                        // Normal operation
                    }
                }
            }
        }

        // Update notification
        updateNotification("Monitoring active")

        Timber.i("Vehicle monitoring started")
    }

    private fun stopMonitoring() {
        Timber.i("Stopping vehicle monitoring")

        // Transition to IDLE state
        currentState = RuntimeState.IDLE

        // Stop camera
        cameraService.stopRecording()
        cameraService.shutdown()

        // Stop telemetry
        telemetryService.stopTracking()
        telemetryService.shutdown()

        // Update notification
        updateNotification("Monitoring stopped")

        Timber.i("Vehicle monitoring stopped")
    }

    private fun pauseMonitoring() {
        Timber.i("Pausing vehicle monitoring")
        cameraService.stopRecording()
        telemetryService.stopTracking()
        updateNotification("Monitoring paused")
    }

    private fun resumeMonitoring() {
        Timber.i("Resuming vehicle monitoring")
        cameraService.startRecording()
        telemetryService.startTracking()
        updateNotification("Monitoring active")
    }

    //endregion

    //region Notification Management

    private fun createNotificationChannel() {
        val channel = NotificationChannel(
            CHANNEL_ID,
            "Vehicle Monitoring",
            NotificationManager.IMPORTANCE_LOW
        ).apply {
            description = "Continuous vehicle monitoring and recording"
            setShowBadge(false)
            enableVibration(false)
            setSound(null, null)
        }

        val notificationManager = getSystemService(NotificationManager::class.java)
        notificationManager.createNotificationChannel(channel)
    }

    private fun startForegroundWithNotification() {
        val notification = createNotification("Starting monitoring...")

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            ServiceCompat.startForeground(
                this,
                NOTIFICATION_ID,
                notification,
                ServiceInfo.FOREGROUND_SERVICE_TYPE_CAMERA or
                        ServiceInfo.FOREGROUND_SERVICE_TYPE_LOCATION
            )
        } else {
            startForeground(NOTIFICATION_ID, notification)
        }
    }

    private fun updateNotification(status: String) {
        val notification = createNotification(status)
        val notificationManager = getSystemService(NotificationManager::class.java)
        notificationManager.notify(NOTIFICATION_ID, notification)
    }

    private fun createNotification(status: String): Notification {
        // Create intent to open main activity
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            Intent(this, com.flyfleet.vehicledevice.ui.MainActivity::class.java),
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        return Notification.Builder(this, CHANNEL_ID)
            .setContentTitle("FlyFleet Vehicle Monitor")
            .setContentText(status)
            .setSmallIcon(android.R.drawable.ic_menu_camera)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setShowWhen(false)
            .build()
    }

    //endregion

    //region Power Management

    private fun acquireWakeLock() {
        try {
            val powerManager = getSystemService(Context.POWER_SERVICE) as PowerManager
            wakeLock = powerManager.newWakeLock(
                PowerManager.PARTIAL_WAKE_LOCK,
                "flyfleet::VehicleMonitoringService"
            ).apply {
                acquire(10 * 60 * 1000L) // 10 minutes timeout
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

    //endregion
}

/**
 * Dummy LifecycleOwner for use in service context
 * CameraX requires a LifecycleOwner, but service doesn't have one
 */
class DummyLifecycleOwner : androidx.lifecycle.LifecycleOwner {
    private val lifecycleRegistry = androidx.lifecycle.LifecycleRegistry(this)

    init {
        lifecycleRegistry.currentState = androidx.lifecycle.Lifecycle.State.RESUMED
    }

    override val lifecycle: androidx.lifecycle.Lifecycle
        get() = lifecycleRegistry
}
