package com.flyfleet.vehicledevice

import android.app.Application
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.os.Build
import android.util.Log
import com.flyfleet.vehicledevice.analytics.PostHogAnalytics
import com.flyfleet.vehicledevice.api.ApiClient
import com.flyfleet.vehicledevice.config.DeviceConfig
import com.flyfleet.vehicledevice.security.SecurityModule
import com.flyfleet.vehicledevice.storage.StorageManager
import com.posthog.PostHog
import com.posthog.android.PostHogAndroid
import com.posthog.android.PostHogAndroidConfig
import timber.log.Timber

/**
 * Main Application class for FlyFleet Vehicle Device
 * 
 * Initializes core components:
 * - Logging (Timber)
 * - Security Module (Android Keystore)
 * - API Client (Retrofit)
 * - Device Configuration
 * - Storage Manager
 * - Notification Channels
 * - PostHog Analytics (Open-Source)
 */
class VehicleDeviceApplication : Application() {

    // Core components (lazy initialization)
    val securityModule: SecurityModule by lazy { SecurityModule(this) }
    
    val apiClient: ApiClient by lazy {
        ApiClient(
            baseUrl = DeviceConfig.getBackendUrl(this),
            jwtSecret = DeviceConfig.getJwtSecret(this),
            securityModule = securityModule
        )
    }
    
    val storageManager: StorageManager by lazy { StorageManager(this) }
    
    val deviceConfig: DeviceConfig by lazy { DeviceConfig(this) }
    
    val postHogConfig: PostHogAndroidConfig by lazy {
        PostHogAndroidConfig(
            apiKey = BuildConfig.POSTHOG_API_KEY,
            host = BuildConfig.POSTHOG_HOST
        ).apply {
            captureApplicationLifecycleEvents = true
            captureScreenViews = false // No UI screens in vehicle device
            debug = BuildConfig.DEBUG
        }
    }
    
    val analytics: PostHogAnalytics by lazy { PostHogAnalytics(this) }

    override fun onCreate() {
        super.onCreate()
        
        // Initialize logging
        initLogging()
        
        // Initialize security module
        initSecurity()
        
        // Initialize notification channels
        initNotificationChannels()
        
        // Initialize PostHog analytics
        initAnalytics()
        
        Timber.i("VehicleDeviceApplication initialized")
    }

    override fun onTerminate() {
        super.onTerminate()
        PostHog.close()
        Timber.i("VehicleDeviceApplication terminated")
    }

    override fun onLowMemory() {
        super.onLowMemory()
        Timber.w("Low memory warning")
        // Trigger cleanup
        storageManager.cleanup()
    }

    override fun onTrimMemory(level: Int) {
        super.onTrimMemory(level)
        Timber.d("Trimming memory: $level")
        when (level) {
            TRIM_MEMORY_RUNNING_CRITICAL,
            TRIM_MEMORY_COMPLETE -> {
                // Aggressive cleanup
                storageManager.cleanup()
            }
        }
    }

    /**
     * Initialize Timber logging
     */
    private fun initLogging() {
        if (BuildConfig.DEBUG) {
            Timber.plant(Timber.DebugTree())
        } else {
            // Plant release tree that logs to file/backend
            Timber.plant(FileLoggingTree(this))
        }
    }

    /**
     * Initialize security module with hardware-backed keys
     */
    private fun initSecurity() {
        try {
            val status = securityModule.initialize()
            Timber.i("Security module initialized: $status")
        } catch (e: Exception) {
            Timber.e(e, "Failed to initialize security module")
        }
    }

    /**
     * Create notification channels for foreground services
     */
    private fun initNotificationChannels() {
        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        // Monitoring service channel
        val monitoringChannel = NotificationChannel(
            CHANNEL_MONITORING,
            "Vehicle Monitoring",
            NotificationManager.IMPORTANCE_LOW
        ).apply {
            description = "Continuous vehicle monitoring and recording"
            setShowBadge(false)
            enableVibration(false)
            setSound(null, null)
        }

        // Alerts channel
        val alertsChannel = NotificationChannel(
            CHANNEL_ALERTS,
            "Safety Alerts",
            NotificationManager.IMPORTANCE_HIGH
        ).apply {
            description = "Critical safety alerts and events"
            enableVibration(true)
            enableLights(true)
            lightColor = android.graphics.Color.RED
        }

        // Sync channel
        val syncChannel = NotificationChannel(
            CHANNEL_SYNC,
            "Data Sync",
            NotificationManager.IMPORTANCE_LOW
        ).apply {
            description = "Cloud synchronization status"
            setShowBadge(false)
        }

        notificationManager.createNotificationChannels(
            listOf(monitoringChannel, alertsChannel, syncChannel)
        )
    }

    /**
     * Initialize PostHog analytics (open-source)
     */
    private fun initAnalytics() {
        try {
            // Initialize PostHog
            PostHogAndroid.setup(this, postHogConfig)
            analytics.initialize()
            
            Timber.i("PostHog analytics initialized")
        } catch (e: Exception) {
            Timber.e(e, "Failed to initialize PostHog analytics")
        }
    }

    companion object {
        // Notification channel IDs
        const val CHANNEL_MONITORING = "vehicle_monitoring"
        const val CHANNEL_ALERTS = "safety_alerts"
        const val CHANNEL_SYNC = "data_sync"

        // Static access (for use outside of Android context)
        @Volatile
        private var instance: VehicleDeviceApplication? = null

        fun getInstance(): VehicleDeviceApplication {
            return instance ?: throw IllegalStateException(
                "Application not initialized yet"
            )
        }
    }
}

/**
 * File-based logging tree for production builds
 */
class FileLoggingTree(private val context: Context) : Timber.Tree() {
    override fun log(priority: Int, tag: String?, message: String, t: Throwable?) {
        // In production, log to file for later upload
        // Implementation would use a rolling file appender
        // For now, just log to Logcat with reduced verbosity
        if (priority >= Log.WARN) {
            android.util.Log.println(priority, tag, message)
            t?.printStackTrace()
        }
    }
}
